package resolve_markers

import (
	"context"
	"errors"

	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/domain/service"
)

// FeatureContext is the per-scenario state for resolve_markers BDD. It
// composes scripted prompter, in-memory FS, optional canned LLM provider,
// and an optional canned diff previewer.
type FeatureContext struct {
	files          map[string]string
	written        map[string]string
	confirm        bool
	confirmExplicit bool
	answers        []service.PromptedAnswer
	answerIdx      int
	skipMode       service.SkipMode
	previewAction  string // "", "apply", "discard"
	llmRewrite     string

	result *command.ResolveResult
	err    error
}

func (f *FeatureContext) SetupTest() {}

func (f *FeatureContext) reset() {
	f.files = map[string]string{}
	f.written = map[string]string{}
	f.confirm = false
	f.confirmExplicit = false
	f.answers = nil
	f.answerIdx = 0
	f.skipMode = service.SkipModeTODO
	f.previewAction = ""
	f.llmRewrite = ""
	f.result = nil
	f.err = nil
}

// scriptedPrompter implements service.InteractivePrompter using the
// FeatureContext canned answers/confirm state.
type scriptedPrompter struct{ f *FeatureContext }

func (p *scriptedPrompter) ConfirmTopLevel(int, int) (bool, error) {
	if !p.f.confirmExplicit {
		return true, nil // sensible default for scenarios that omit the step
	}
	return p.f.confirm, nil
}
func (p *scriptedPrompter) AnnounceFile(string, int) {}
func (p *scriptedPrompter) AskMarker(_ string, _ service.EnrichedMarker) (service.PromptedAnswer, error) {
	if p.f.answerIdx >= len(p.f.answers) {
		return service.PromptedAnswer{Skip: true}, nil
	}
	a := p.f.answers[p.f.answerIdx]
	p.f.answerIdx++
	return a, nil
}
func (p *scriptedPrompter) ReportFileResult(string, int, string) {}

// scriptedProvider returns the canned LLM rewrite text or an error.
type scriptedProvider struct{ f *FeatureContext }

func (p *scriptedProvider) GenerateContext(context.Context, service.GenerationRequest) (*service.GenerationResponse, error) {
	return nil, errors.New("not used")
}

func (p *scriptedProvider) EvaluatePrompt(context.Context, service.EvaluationRequest) (*service.EvaluationResponse, error) {
	if p.f.llmRewrite == "" {
		return nil, errors.New("no canned response")
	}
	return &service.EvaluationResponse{Text: p.f.llmRewrite}, nil
}

// scriptedPreviewer mirrors the user's diff-preview decision.
type scriptedPreviewer struct{ f *FeatureContext }

func (p *scriptedPreviewer) Preview(_ string, _, after []byte) (bool, []byte, error) {
	switch p.f.previewAction {
	case "discard":
		return false, nil, nil
	default:
		return true, after, nil
	}
}
