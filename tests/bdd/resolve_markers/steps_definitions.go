package resolve_markers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"

	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/domain/service"
)

var featureContext = new(FeatureContext)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// Givens
	ctx.Step(`^a file "([^"]+)" with content "([^"]*)"$`, featureContext.aFileWithContent)
	ctx.Step(`^the user accepts the top-level prompt$`, featureContext.userAcceptsTopLevel)
	ctx.Step(`^the user declines the top-level prompt$`, featureContext.userDeclinesTopLevel)
	ctx.Step(`^the user answers "([^"]*)" for marker on line (\d+)$`, featureContext.userAnswers)
	ctx.Step(`^the user skips marker on line (\d+)$`, featureContext.userSkips)
	ctx.Step(`^the LLM provider returns rewritten content "([^"]*)"$`, featureContext.providerReturns)
	ctx.Step(`^the diff preview action is "([^"]+)"$`, featureContext.setPreviewAction)

	// Whens
	ctx.Step(`^the resolver runs over "([^"]+)"$`, featureContext.resolverRunsOverOne)
	ctx.Step(`^the resolver runs over files "([^"]+)" "([^"]+)"$`, featureContext.resolverRunsOverTwo)
	ctx.Step(`^the resolver runs in verbatim skip mode over "([^"]+)"$`, featureContext.resolverRunsVerbatim)

	// Thens
	ctx.Step(`^the file "([^"]+)" should equal "([^"]*)"$`, featureContext.fileShouldEqual)
	ctx.Step(`^the file "([^"]+)" should contain "([^"]*)"$`, featureContext.fileShouldContain)
	ctx.Step(`^the file "([^"]+)" should not contain "([^"]*)"$`, featureContext.fileShouldNotContain)
	ctx.Step(`^the resolve summary should report (\d+) markers? resolved$`, featureContext.summaryResolved)
	ctx.Step(`^the resolve summary should report (\d+) files? rewritten$`, featureContext.summaryRewritten)
	ctx.Step(`^the resolve summary should report (\d+) files? discarded$`, featureContext.summaryDiscarded)
	ctx.Step(`^the resolve summary should report decline$`, featureContext.summaryDecline)
}

// --- Givens ---

func (f *FeatureContext) aFileWithContent(path, content string) error {
	f.files[path] = content
	return nil
}

func (f *FeatureContext) userAcceptsTopLevel() error {
	f.confirm = true
	f.confirmExplicit = true
	return nil
}

func (f *FeatureContext) userDeclinesTopLevel() error {
	f.confirm = false
	f.confirmExplicit = true
	return nil
}

func (f *FeatureContext) userAnswers(answer string, _ int) error {
	f.answers = append(f.answers, service.PromptedAnswer{Answer: answer})
	return nil
}

func (f *FeatureContext) userSkips(_ int) error {
	f.answers = append(f.answers, service.PromptedAnswer{Skip: true})
	return nil
}

func (f *FeatureContext) providerReturns(content string) error {
	f.llmRewrite = content
	return nil
}

func (f *FeatureContext) setPreviewAction(action string) error {
	f.previewAction = action
	return nil
}

// --- Whens ---

func (f *FeatureContext) resolverRunsOverOne(path string) error {
	return f.runResolver([]string{path})
}

func (f *FeatureContext) resolverRunsOverTwo(a, b string) error {
	return f.runResolver([]string{a, b})
}

func (f *FeatureContext) resolverRunsVerbatim(path string) error {
	f.skipMode = service.SkipModeVerbatim
	return f.runResolver([]string{path})
}

func (f *FeatureContext) runResolver(files []string) error {
	prompter := &scriptedPrompter{f: f}

	cmd := command.NewResolveMarkersCommand(prompter, nil).
		WithFileIO(
			func(p string) ([]byte, error) {
				if c, ok := f.files[p]; ok {
					return []byte(c), nil
				}
				return nil, os.ErrNotExist
			},
			func(p string, data []byte, _ os.FileMode) error {
				f.written[p] = string(data)
				f.files[p] = string(data) // mirror so subsequent reads see writes
				return nil
			},
		).
		WithStderr(func(string, ...any) {}).
		WithToday(func() string { return "2026-05-06" })

	if f.llmRewrite != "" {
		// Re-create the command with provider since the constructor takes
		// it directly. Using NewResolveMarkersCommand with the provider gives
		// us the LLM rewrite path (with Phase 1 validator + fallback).
		cmd = command.NewResolveMarkersCommand(prompter, &scriptedProvider{f: f}).
			WithFileIO(
				func(p string) ([]byte, error) {
					if c, ok := f.files[p]; ok {
						return []byte(c), nil
					}
					return nil, os.ErrNotExist
				},
				func(p string, data []byte, _ os.FileMode) error {
					f.written[p] = string(data)
					f.files[p] = string(data)
					return nil
				},
			).
			WithStderr(func(string, ...any) {}).
			WithToday(func() string { return "2026-05-06" })
	}

	if f.previewAction != "" {
		cmd = cmd.WithPreviewer(&scriptedPreviewer{f: f})
	}

	res, err := cmd.Execute(context.Background(), command.ResolveRequest{
		Files:    files,
		Locale:   "en",
		SkipMode: f.skipMode,
	})
	f.result = res
	f.err = err
	return err
}

// --- Thens ---

func (f *FeatureContext) fileShouldEqual(path, want string) error {
	got, ok := f.files[path]
	if !ok {
		return fmt.Errorf("file %s not found in fixture", path)
	}
	if got != want {
		return fmt.Errorf("file %s mismatch:\n  got:  %q\n  want: %q", path, got, want)
	}
	return nil
}

func (f *FeatureContext) fileShouldContain(path, substr string) error {
	got, ok := f.files[path]
	if !ok {
		return fmt.Errorf("file %s not found", path)
	}
	if !strings.Contains(got, substr) {
		return fmt.Errorf("file %s does not contain %q; full content: %q", path, substr, got)
	}
	return nil
}

func (f *FeatureContext) fileShouldNotContain(path, substr string) error {
	got, ok := f.files[path]
	if !ok {
		return fmt.Errorf("file %s not found", path)
	}
	if strings.Contains(got, substr) {
		return fmt.Errorf("file %s should not contain %q; full content: %q", path, substr, got)
	}
	return nil
}

func (f *FeatureContext) summaryResolved(want int) error {
	if f.result == nil {
		return fmt.Errorf("no result captured")
	}
	if f.result.Resolved != want {
		return fmt.Errorf("Resolved counter: got %d want %d (full result: %+v)", f.result.Resolved, want, f.result)
	}
	return nil
}

func (f *FeatureContext) summaryRewritten(want int) error {
	if f.result == nil {
		return fmt.Errorf("no result captured")
	}
	if f.result.FilesRewritten != want {
		return fmt.Errorf("FilesRewritten counter: got %d want %d (full result: %+v)", f.result.FilesRewritten, want, f.result)
	}
	return nil
}

func (f *FeatureContext) summaryDiscarded(want int) error {
	if f.result == nil {
		return fmt.Errorf("no result captured")
	}
	if f.result.FilesDiscarded != want {
		return fmt.Errorf("FilesDiscarded counter: got %d want %d", f.result.FilesDiscarded, want)
	}
	return nil
}

func (f *FeatureContext) summaryDecline() error {
	if f.result == nil {
		return fmt.Errorf("no result captured")
	}
	if !f.result.Declined {
		return fmt.Errorf("expected Declined=true, got %+v", f.result)
	}
	return nil
}
