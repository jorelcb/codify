package command

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
)

// scriptedPrompter is a service.InteractivePrompter driven by canned answers.
// Each AskMarker call consumes the next entry in answers; ConfirmTopLevel
// returns confirm. Used to exercise the full orchestrator without a TTY.
type scriptedPrompter struct {
	confirm    bool
	confirmErr error
	answers    []service.PromptedAnswer
	asked      []service.MarkerHit
	results    []scriptedResult
	idx        int
}

type scriptedResult struct {
	path     string
	resolved int
	mode     string
}

func (p *scriptedPrompter) ConfirmTopLevel(_, _ int) (bool, error) {
	return p.confirm, p.confirmErr
}

func (p *scriptedPrompter) AnnounceFile(string, int) {}

func (p *scriptedPrompter) AskMarker(_ string, m service.EnrichedMarker) (service.PromptedAnswer, error) {
	p.asked = append(p.asked, m.MarkerHit)
	if p.idx >= len(p.answers) {
		return service.PromptedAnswer{Skip: true}, nil
	}
	a := p.answers[p.idx]
	p.idx++
	return a, nil
}

func (p *scriptedPrompter) ReportFileResult(path string, resolved int, mode string) {
	p.results = append(p.results, scriptedResult{path: path, resolved: resolved, mode: mode})
}

// inMemoryFS captures reads and writes for the orchestrator under test.
type inMemoryFS struct {
	read  map[string]string
	write map[string]string
}

func newInMemoryFS(initial map[string]string) *inMemoryFS {
	cp := make(map[string]string, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &inMemoryFS{read: cp, write: map[string]string{}}
}

func (fs *inMemoryFS) readFile(path string) ([]byte, error) {
	v, ok := fs.read[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(v), nil
}

func (fs *inMemoryFS) writeFile(path string, data []byte, _ os.FileMode) error {
	fs.write[path] = string(data)
	return nil
}

func newCommandWithFS(prompter service.InteractivePrompter, provider service.LLMProvider, fs *inMemoryFS) *ResolveMarkersCommand {
	return NewResolveMarkersCommand(prompter, provider).
		WithFileIO(fs.readFile, fs.writeFile).
		WithStderr(func(string, ...any) {}).
		WithToday(func() string { return "2026-05-06" })
}

func TestExecute_NoMarkers_IsNoop(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "no markers here",
	})
	prompter := &scriptedPrompter{confirm: true}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.TotalMarkers != 0 || res.FilesScanned != 0 {
		t.Errorf("expected no scan results, got %+v", res)
	}
	if len(fs.write) != 0 {
		t.Errorf("expected no writes, got %v", fs.write)
	}
}

func TestExecute_DeclineTopLevel_LeavesFilesUntouched(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency is [DEFINE: code]",
	})
	prompter := &scriptedPrompter{confirm: false}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.Declined {
		t.Errorf("expected Declined=true, got %+v", res)
	}
	if len(fs.write) != 0 {
		t.Errorf("expected no writes after decline, got %v", fs.write)
	}
}

func TestExecute_TopLevelError_TreatedAsDecline(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency is [DEFINE: code]",
	})
	prompter := &scriptedPrompter{confirmErr: errors.New("user cancelled")}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !res.Declined {
		t.Errorf("expected Declined=true after prompter error, got %+v", res)
	}
}

func TestExecute_SkipAll_FileUnchanged(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "a [DEFINE: x] and b [DEFINE: y]",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Skip: true},
			{Skip: true},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Skipped != 2 || res.Resolved != 0 || res.FilesUnchanged != 1 {
		t.Errorf("unexpected counters: %+v", res)
	}
	if len(fs.write) != 0 {
		t.Errorf("expected no writes, got %v", fs.write)
	}
	if len(prompter.results) != 1 || prompter.results[0].mode != "unchanged" {
		t.Errorf("expected single 'unchanged' report, got %+v", prompter.results)
	}
}

func TestExecute_LiteralPath_ReplacesAnsweredMarkers(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code], tz [DEFINE: tz]",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Answer: "USD"},
			{Skip: true},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs) // provider == nil → literal path

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Resolved != 1 || res.Skipped != 1 {
		t.Errorf("counters: got %+v", res)
	}
	if res.UsedLiteral != 1 || res.UsedLLM != 0 {
		t.Errorf("path usage: got %+v", res)
	}
	got := fs.write["AGENTS.md"]
	want := "currency USD, tz <!-- TODO 2026-05-06: tz -->"
	if got != want {
		t.Errorf("written content:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExecute_VerbatimSkipMode_LeavesMarkerUntouched(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code], tz [DEFINE: tz]",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Answer: "USD"},
			{Skip: true},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	_, err := cmd.Execute(context.Background(), ResolveRequest{
		Files:    []string{"AGENTS.md"},
		SkipMode: service.SkipModeVerbatim,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := fs.write["AGENTS.md"]
	want := "currency USD, tz [DEFINE: tz]"
	if got != want {
		t.Errorf("verbatim skip mode should not write TODO:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExecute_TODOSkipMode_AppliesAnchorBasedOnExtension(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"app.go":   "var currency = [DEFINE: code] // line",
		"data.yml": "tz: [DEFINE: tz]\n",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Skip: true}, // app.go marker
			{Skip: true}, // data.yml marker
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	// Both files have all-skipped markers, so neither gets rewritten — the
	// orchestrator only applies skip-mode after at least one answer triggers
	// a write. Validate that contract is preserved (FilesUnchanged=2).
	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"app.go", "data.yml"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.FilesUnchanged != 2 || res.FilesRewritten != 0 {
		t.Errorf("expected both files unchanged, got %+v", res)
	}
	if len(fs.write) != 0 {
		t.Errorf("no writes should happen when all markers skipped: %v", fs.write)
	}
}

func TestExecute_TODOSkipMode_AnchorsAppliedWhenFileWritten(t *testing.T) {
	// Mixed file: one answer triggers the rewrite path, the skipped marker
	// in the same file gets the TODO anchor.
	fs := newInMemoryFS(map[string]string{
		"app.go": "var currency = [DEFINE: code]\nvar tz = [DEFINE: tz]\n",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Answer: "\"USD\""}, // applies
			{Skip: true},         // gets TODO anchor in .go syntax
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	_, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"app.go"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := fs.write["app.go"]
	want := "var currency = \"USD\"\nvar tz = // TODO 2026-05-06: tz\n"
	if got != want {
		t.Errorf("written content:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExecute_LLMPath_UsesProviderRewrite(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code]",
	})
	provider := &fakeProvider{
		rewriteWith: "Currency is USD throughout the system.",
	}
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "USD"}},
	}
	cmd := newCommandWithFS(prompter, provider, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}, Locale: "en"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.UsedLLM != 1 || res.UsedLiteral != 0 {
		t.Errorf("expected LLM path, got %+v", res)
	}
	if got := fs.write["AGENTS.md"]; got != "Currency is USD throughout the system." {
		t.Errorf("LLM rewrite not applied; got %q", got)
	}
	if !provider.evalCalled {
		t.Error("provider.EvaluatePrompt was not invoked")
	}
	if !strings.Contains(provider.lastUserPrompt, "USD") {
		t.Errorf("user prompt should contain the answer, got: %s", provider.lastUserPrompt)
	}
}

func TestExecute_LLMRewriteWithIssues_FallsBackToLiteral(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code], tz [DEFINE: tz]",
	})
	// Provider returns content that hallucinates a new marker — validator
	// must catch this and trigger literal fallback.
	provider := &fakeProvider{
		rewriteWith: "currency USD, tz [DEFINE: tz] [DEFINE: hallucinated]",
	}
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Answer: "USD"},
			{Skip: true},
		},
	}
	cmd := newCommandWithFS(prompter, provider, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.UsedLLM != 0 || res.UsedLiteral != 1 {
		t.Errorf("expected literal fallback, got %+v", res)
	}
	got := fs.write["AGENTS.md"]
	want := "currency USD, tz <!-- TODO 2026-05-06: tz -->"
	if got != want {
		t.Errorf("written content:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExecute_LLMFailure_FallsBackToLiteral(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code]",
	})
	provider := &fakeProvider{evalErr: errors.New("network down")}
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "USD"}},
	}
	cmd := newCommandWithFS(prompter, provider, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute should not fail on LLM error: %v", err)
	}
	if res.UsedLiteral != 1 {
		t.Errorf("expected literal fallback, got %+v", res)
	}
	if got := fs.write["AGENTS.md"]; got != "currency USD" {
		t.Errorf("literal fallback content: got %q", got)
	}
}

func TestExecute_MultipleFiles_TracksCounters(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"A.md": "a [DEFINE: x]",
		"B.md": "b [DEFINE: y] [DEFINE: z]",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{
			{Answer: "alpha"},  // A.md
			{Answer: "beta"},   // B.md, first
			{Skip: true},       // B.md, second
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"A.md", "B.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.FilesScanned != 2 || res.FilesRewritten != 2 {
		t.Errorf("file counters: got %+v", res)
	}
	if res.TotalMarkers != 3 || res.Resolved != 2 || res.Skipped != 1 {
		t.Errorf("marker counters: got %+v", res)
	}
}

func TestExecute_MissingFile_IsSkippedSilently(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"existing.md": "[DEFINE: x]",
	})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "v"}},
	}
	cmd := newCommandWithFS(prompter, nil, fs)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"existing.md", "missing.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.FilesScanned != 1 || res.FilesRewritten != 1 {
		t.Errorf("missing file should be silently skipped: got %+v", res)
	}
}

// scriptedPreviewer returns canned (apply, content) tuples per call.
type scriptedPreviewer struct {
	apply       bool
	editedAfter []byte
	err         error
	calls       int
}

func (p *scriptedPreviewer) Preview(_ string, _, after []byte) (bool, []byte, error) {
	p.calls++
	if p.err != nil {
		return false, nil, p.err
	}
	if !p.apply {
		return false, nil, nil
	}
	if p.editedAfter != nil {
		return true, p.editedAfter, nil
	}
	return true, after, nil
}

func TestExecute_PreviewerDiscards_FileLeftUntouched(t *testing.T) {
	original := "currency [DEFINE: code]"
	fs := newInMemoryFS(map[string]string{"AGENTS.md": original})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "USD"}},
	}
	previewer := &scriptedPreviewer{apply: false}
	cmd := newCommandWithFS(prompter, nil, fs).WithPreviewer(previewer)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.FilesDiscarded != 1 || res.FilesRewritten != 0 {
		t.Errorf("counters: %+v", res)
	}
	if _, written := fs.write["AGENTS.md"]; written {
		t.Error("file should not be written when previewer discards")
	}
}

func TestExecute_PreviewerEdits_AppliesEditedContent(t *testing.T) {
	fs := newInMemoryFS(map[string]string{"AGENTS.md": "currency [DEFINE: code]"})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "USD"}},
	}
	previewer := &scriptedPreviewer{
		apply:       true,
		editedAfter: []byte("currency hand-edited"),
	}
	cmd := newCommandWithFS(prompter, nil, fs).WithPreviewer(previewer)

	_, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := fs.write["AGENTS.md"]; got != "currency hand-edited" {
		t.Errorf("expected edited content to be written, got %q", got)
	}
}

func TestExecute_PreviewerError_AppliesWithoutPreview(t *testing.T) {
	fs := newInMemoryFS(map[string]string{"AGENTS.md": "currency [DEFINE: code]"})
	prompter := &scriptedPrompter{
		confirm: true,
		answers: []service.PromptedAnswer{{Answer: "USD"}},
	}
	previewer := &scriptedPreviewer{err: errors.New("preview crashed")}
	cmd := newCommandWithFS(prompter, nil, fs).WithPreviewer(previewer)

	res, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute should not fail when previewer errors: %v", err)
	}
	if res.FilesRewritten != 1 {
		t.Errorf("expected file written despite preview error, got %+v", res)
	}
	if got := fs.write["AGENTS.md"]; got != "currency USD" {
		t.Errorf("content: %q", got)
	}
}

// scriptedEnricher returns canned EnrichedMarker entries. Used to verify that
// the orchestrator surfaces enrichment data through to the prompter.
type scriptedEnricher struct {
	out []service.EnrichedMarker
	err error
}

func (e *scriptedEnricher) Enrich(_ context.Context, _, _, _ string, hits []service.MarkerHit) ([]service.EnrichedMarker, error) {
	if e.err != nil {
		// Mirror the real enricher contract: on error, still return entries
		// for every hit so the orchestrator can keep walking.
		fallback := make([]service.EnrichedMarker, len(hits))
		for i, h := range hits {
			fallback[i] = service.EnrichedMarker{MarkerHit: h}
		}
		return fallback, e.err
	}
	return e.out, nil
}

// recordingPrompter is a scriptedPrompter wrapper that captures the
// EnrichedMarker passed to AskMarker so tests can assert the orchestrator
// forwarded enrichment data correctly.
type recordingPrompter struct {
	scriptedPrompter
	receivedEnrichments []service.EnrichedMarker
}

func (p *recordingPrompter) AskMarker(content string, m service.EnrichedMarker) (service.PromptedAnswer, error) {
	p.receivedEnrichments = append(p.receivedEnrichments, m)
	return p.scriptedPrompter.AskMarker(content, m)
}

func TestExecute_WithEnricher_ForwardsEnrichmentToPrompter(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code]",
	})
	enricher := &scriptedEnricher{
		out: []service.EnrichedMarker{
			{
				MarkerHit:   service.MarkerHit{Text: "[DEFINE: code]", Line: 1},
				Question:    "¿Qué moneda usa la aplicación?",
				Suggestions: []string{"USD", "EUR"},
				Default:     "USD",
				Rationale:   "fintech context",
			},
		},
	}
	prompter := &recordingPrompter{
		scriptedPrompter: scriptedPrompter{
			confirm: true,
			answers: []service.PromptedAnswer{{Answer: "USD"}},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs).WithEnricher(enricher)

	_, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(prompter.receivedEnrichments) != 1 {
		t.Fatalf("expected one AskMarker call, got %d", len(prompter.receivedEnrichments))
	}
	got := prompter.receivedEnrichments[0]
	if got.Question != "¿Qué moneda usa la aplicación?" {
		t.Errorf("question not forwarded: %q", got.Question)
	}
	if len(got.Suggestions) != 2 || got.Default != "USD" {
		t.Errorf("suggestions/default not forwarded: %+v / %q", got.Suggestions, got.Default)
	}
}

func TestExecute_EnricherError_FallsBackToZeroValueEnrichment(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code]",
	})
	enricher := &scriptedEnricher{err: errors.New("provider down")}
	prompter := &recordingPrompter{
		scriptedPrompter: scriptedPrompter{
			confirm: true,
			answers: []service.PromptedAnswer{{Answer: "USD"}},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs).WithEnricher(enricher)

	_, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute should not fail when enricher errors: %v", err)
	}
	if len(prompter.receivedEnrichments) != 1 {
		t.Fatalf("expected one AskMarker call, got %d", len(prompter.receivedEnrichments))
	}
	if prompter.receivedEnrichments[0].Question != "" {
		t.Errorf("expected zero-value enrichment after enricher error, got %+v", prompter.receivedEnrichments[0])
	}
}

func TestExecute_NoEnricher_StillWorksAsLegacy(t *testing.T) {
	fs := newInMemoryFS(map[string]string{
		"AGENTS.md": "currency [DEFINE: code]",
	})
	prompter := &recordingPrompter{
		scriptedPrompter: scriptedPrompter{
			confirm: true,
			answers: []service.PromptedAnswer{{Answer: "USD"}},
		},
	}
	cmd := newCommandWithFS(prompter, nil, fs) // no enricher

	_, err := cmd.Execute(context.Background(), ResolveRequest{Files: []string{"AGENTS.md"}})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if prompter.receivedEnrichments[0].Question != "" {
		t.Errorf("nil enricher should produce zero-value enrichment, got %+v", prompter.receivedEnrichments[0])
	}
}

// fakeProvider satisfies service.LLMProvider for testing the LLM path.
// GenerateContext is not used by ResolveMarkersCommand and is left as a no-op.
type fakeProvider struct {
	rewriteWith    string
	evalErr        error
	evalCalled     bool
	lastUserPrompt string
}

func (f *fakeProvider) GenerateContext(_ context.Context, _ service.GenerationRequest) (*service.GenerationResponse, error) {
	return nil, errors.New("not used")
}

func (f *fakeProvider) EvaluatePrompt(_ context.Context, req service.EvaluationRequest) (*service.EvaluationResponse, error) {
	f.evalCalled = true
	f.lastUserPrompt = req.UserPrompt
	if f.evalErr != nil {
		return nil, f.evalErr
	}
	return &service.EvaluationResponse{Text: f.rewriteWith}, nil
}
