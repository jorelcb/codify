package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// defineMarkerRE matches the [DEFINE] / [DEFINE: hint] placeholders the LLM
// emits in generated context files when it lacks information. The dual form
// (with and without colon-hint) is intentional: bare [DEFINE] is legacy from
// pre-v2.0.4, kept matched for backward compatibility on previously-generated
// files. New generations always include a colon-hint per prompt rules.
var defineMarkerRE = regexp.MustCompile(`\[DEFINE(?::[^\]\n]+)?\]`)

// MarkerHit captures a single [DEFINE] occurrence inside a file: its verbatim
// text, 1-based line number, and the user's answer (empty when skipped).
//
// Lives in the domain layer because the marker concept is part of the
// product's vocabulary, not an implementation detail.
type MarkerHit struct {
	Text   string // verbatim "[DEFINE: hint]" or "[DEFINE]"
	Line   int    // 1-based
	Answer string // empty = user chose to skip
}

// EnrichedMarker decorates a MarkerHit with LLM-derived metadata used by the
// interactive prompter to render a friendlier question. Phase 3 populates the
// Question / Suggestions / Default fields. Until then, only the embedded hit
// matters and the rest stay zero-valued so the prompter falls back to the
// legacy "show hint, ask for input" UI.
type EnrichedMarker struct {
	MarkerHit
	Question    string   // natural-language question in the file's locale
	Suggestions []string // grounded suggestions inferred from file context
	Default     string   // optional default — must be one of Suggestions or empty
	Rationale   string   // brief reasoning (shown as helper text)
}

// PromptedAnswer is what the InteractivePrompter returns for each marker.
// Skip is preferred over checking Answer == "" because Phase 3 will accept
// inputs like "s" / "skip" as explicit skip signals — distinguishing those
// from a user who typed a literal "s" as their answer.
type PromptedAnswer struct {
	Answer string
	Skip   bool
}

// MarkerEnricher abstracts the LLM-driven step that turns raw [DEFINE: hint]
// markers into user-friendly EnrichedMarker entries (natural-language
// question + grounded suggestions + default).
//
// The orchestrator calls Enrich once per file before walking through
// AskMarker. Implementations may return the same hits with empty Question /
// Suggestions when they cannot infer safely — the prompter then degrades
// gracefully to the legacy "show hint, ask for input" UI.
//
// Errors short-circuit enrichment for the file but never abort the whole
// resolve flow: the orchestrator falls back to the unenriched UI for that
// file.
type MarkerEnricher interface {
	Enrich(ctx context.Context, fileName, fileContent, locale string, hits []MarkerHit) ([]EnrichedMarker, error)
}

// InteractivePrompter abstracts the user-facing question loop so the
// orchestrator can be exercised with a scripted prompter in tests. The CLI
// implementation wraps charmbracelet/huh; the test implementation replays a
// canned sequence.
type InteractivePrompter interface {
	// ConfirmTopLevel asks the global "Resolve N markers across M files?"
	// question. Returns true when the user accepts. Errors short-circuit the
	// flow upstream — the caller treats an error the same as a decline.
	ConfirmTopLevel(totalMarkers, totalFiles int) (bool, error)

	// AnnounceFile emits the per-file header. Implementations may render
	// formatting, log to stderr, or no-op (test prompters typically no-op).
	AnnounceFile(path string, markerCount int)

	// AskMarker prompts the user for one marker's answer. Implementations
	// receive the surrounding file content + the EnrichedMarker so they can
	// render context (lines around the marker) and any suggestions.
	AskMarker(fileContent string, marker EnrichedMarker) (PromptedAnswer, error)

	// ReportFileResult prints the per-file outcome. Mirrors AnnounceFile.
	ReportFileResult(path string, resolved int, mode string)
}

// ScanMarkers finds every [DEFINE]/[DEFINE: hint] occurrence in content and
// returns a slice of hits sorted by appearance. Pure function — no IO, no
// state. Used both by the orchestrator and by the post-rewrite validator
// (Phase 1).
func ScanMarkers(content string) []MarkerHit {
	var hits []MarkerHit
	for _, idx := range defineMarkerRE.FindAllStringIndex(content, -1) {
		hits = append(hits, MarkerHit{
			Text: content[idx[0]:idx[1]],
			Line: strings.Count(content[:idx[0]], "\n") + 1,
		})
	}
	return hits
}

// ResolveDelta classifies markers after a rewrite, distinguishing legitimate
// outcomes from LLM hallucinations that warrant a literal-substitution
// fallback. The classification is by marker text frequency before/after the
// rewrite, not by line — line numbers shift naturally as the LLM integrates
// answers into surrounding prose.
type ResolveDelta struct {
	Resolved   []MarkerHit // user answered AND marker disappeared: legitimate
	Skipped    []MarkerHit // user did not answer AND marker still present: legitimate
	NotApplied []MarkerHit // user answered BUT marker still present: LLM ignored answer
	Lost       []MarkerHit // user skipped BUT marker disappeared: LLM hallucinated a fix
	Spurious   []string    // marker text in output that did NOT exist in input: invented
}

// HasIssues reports whether the delta contains any class that indicates the
// LLM rewrite is unsafe to keep — and hence the orchestrator should fall back
// to literal substitution.
func (d ResolveDelta) HasIssues() bool {
	return len(d.NotApplied) > 0 || len(d.Lost) > 0 || len(d.Spurious) > 0
}

// ValidateRewrite compares the LLM-rewritten content against the original
// hits and reports the delta. The validator does not know what the LLM
// changed in the surrounding prose — it only checks marker presence/absence
// against the user's answer/skip choices.
func ValidateRewrite(after string, hits []MarkerHit) ResolveDelta {
	afterHits := ScanMarkers(after)
	afterCountByText := map[string]int{}
	for _, h := range afterHits {
		afterCountByText[h.Text]++
	}

	expectedRemainingByText := map[string]int{} // skipped markers that should still be present
	knownTexts := map[string]bool{}
	var skipped, resolved []MarkerHit
	for _, h := range hits {
		knownTexts[h.Text] = true
		if h.Answer == "" {
			expectedRemainingByText[h.Text]++
			skipped = append(skipped, h)
		} else {
			resolved = append(resolved, h)
		}
	}

	delta := ResolveDelta{Resolved: resolved, Skipped: skipped}

	// Texts that appear in the rewrite: classify excess vs deficit.
	for text, afterCount := range afterCountByText {
		if !knownTexts[text] {
			for i := 0; i < afterCount; i++ {
				delta.Spurious = append(delta.Spurious, text)
			}
			continue
		}
		expected := expectedRemainingByText[text]
		switch {
		case afterCount > expected:
			// More markers remain than the user wanted: LLM left some answered
			// markers unchanged. Tag those as NotApplied.
			excess := afterCount - expected
			for _, h := range resolved {
				if h.Text == text && excess > 0 {
					delta.NotApplied = append(delta.NotApplied, h)
					excess--
				}
			}
		case afterCount < expected:
			// Fewer markers remain than the user wanted: LLM altered some
			// skipped markers it was told to leave alone.
			deficit := expected - afterCount
			for _, h := range skipped {
				if h.Text == text && deficit > 0 {
					delta.Lost = append(delta.Lost, h)
					deficit--
				}
			}
		}
	}

	// Texts that should remain (skipped) but don't appear in the rewrite at
	// all: every occurrence is Lost.
	for text, expected := range expectedRemainingByText {
		if _, present := afterCountByText[text]; present {
			continue // handled by the deficit branch above
		}
		remaining := expected
		for _, h := range skipped {
			if h.Text == text && remaining > 0 {
				delta.Lost = append(delta.Lost, h)
				remaining--
			}
		}
	}

	return delta
}

// SkipMode controls what happens to markers the user chose to skip.
//
// SkipModeUnset is the zero value and resolves to SkipModeTODO at apply
// time — that way callers don't have to remember to set it explicitly to
// get the user-friendly default.
type SkipMode int

const (
	// SkipModeUnset means "use the default", which is SkipModeTODO. Lets
	// ResolveRequest{} stay valid without the caller spelling out the mode.
	SkipModeUnset SkipMode = iota

	// SkipModeTODO replaces each skipped marker with a date-stamped TODO
	// comment in the file's native syntax (// for go/js, # for python/yaml,
	// <!-- --> for markdown/html, etc). Files with unknown extensions fall
	// back to leaving the marker verbatim.
	SkipModeTODO

	// SkipModeVerbatim leaves every skipped marker as [DEFINE: ...]. Useful
	// for users who want to grep their codebase for [DEFINE] later instead
	// of TODO, or for files that should not be polluted with comments.
	SkipModeVerbatim
)

// skipCommentTemplateByExt maps a file extension to a fmt template with two
// %s placeholders: ISO date and the marker hint. Lowercase keys; lookup is
// case-insensitive.
var skipCommentTemplateByExt = map[string]string{
	".md":   "<!-- TODO %s: %s -->",
	".html": "<!-- TODO %s: %s -->",
	".htm":  "<!-- TODO %s: %s -->",
	".xml":  "<!-- TODO %s: %s -->",
	".go":   "// TODO %s: %s",
	".js":   "// TODO %s: %s",
	".jsx":  "// TODO %s: %s",
	".ts":   "// TODO %s: %s",
	".tsx":  "// TODO %s: %s",
	".java": "// TODO %s: %s",
	".kt":   "// TODO %s: %s",
	".rs":   "// TODO %s: %s",
	".c":    "// TODO %s: %s",
	".cpp":  "// TODO %s: %s",
	".cc":   "// TODO %s: %s",
	".h":    "// TODO %s: %s",
	".hpp":  "// TODO %s: %s",
	".swift": "// TODO %s: %s",
	".cs":   "// TODO %s: %s",
	".py":   "# TODO %s: %s",
	".rb":   "# TODO %s: %s",
	".sh":   "# TODO %s: %s",
	".bash": "# TODO %s: %s",
	".yml":  "# TODO %s: %s",
	".yaml": "# TODO %s: %s",
	".toml": "# TODO %s: %s",
	".ini":  "# TODO %s: %s",
}

// hintFromMarker extracts the human-readable hint inside a [DEFINE: ...]
// marker. Returns "value needed" for the bare [DEFINE] form so the resulting
// TODO comment is never empty.
func hintFromMarker(markerText string) string {
	inner := strings.TrimPrefix(markerText, "[DEFINE")
	inner = strings.TrimSuffix(inner, "]")
	inner = strings.TrimPrefix(inner, ":")
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return "value needed"
	}
	return inner
}

// SkipReplacement returns the text that should replace markerText when the
// user skipped it. Empty string means "leave the marker verbatim" — the
// caller should not perform a substitution. Verbatim is also returned for
// unknown extensions so we never write a comment in a syntax we don't
// recognize.
//
// today is injected (as ISO yyyy-mm-dd) for deterministic testing; production
// callers pass time.Now().Format("2006-01-02").
func SkipReplacement(markerText, fileExt string, mode SkipMode, today string) string {
	if mode == SkipModeUnset {
		mode = SkipModeTODO
	}
	if mode == SkipModeVerbatim {
		return ""
	}
	template, ok := skipCommentTemplateByExt[strings.ToLower(fileExt)]
	if !ok {
		return ""
	}
	return fmt.Sprintf(template, today, hintFromMarker(markerText))
}

// ApplySkipMode rewrites content so every still-present skipped marker is
// replaced with a TODO anchor in the file's native syntax. Markers the user
// answered are ignored — they were already integrated by the LLM rewrite or
// literal substitution upstream.
//
// Pure function. Idempotent: running it twice on the same input produces the
// same output (the second pass finds no markers to replace).
func ApplySkipMode(content string, hits []MarkerHit, mode SkipMode, fileExt, today string) string {
	if mode == SkipModeUnset {
		mode = SkipModeTODO
	}
	if mode == SkipModeVerbatim {
		return content
	}
	for _, h := range hits {
		if h.Answer != "" {
			continue
		}
		replacement := SkipReplacement(h.Text, fileExt, mode, today)
		if replacement == "" {
			continue
		}
		content = strings.Replace(content, h.Text, replacement, 1)
	}
	return content
}

// LiteralSubstitute replaces each answered marker with its answer text 1:1,
// preserving skipped markers verbatim. Pure function — used as fallback when
// no LLM provider is available, or as recovery path when the LLM rewrite
// fails or alters skipped markers (Phase 1 validator decision).
//
// Each marker text is replaced once per occurrence; this matters when two
// hits in the slice share the same Text but differ in Line (uncommon but
// possible in templated docs).
func LiteralSubstitute(content string, hits []MarkerHit) string {
	for _, h := range hits {
		if h.Answer == "" {
			continue
		}
		content = strings.Replace(content, h.Text, h.Answer, 1)
	}
	return content
}
