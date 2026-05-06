package commands

import (
	"context"

	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/domain/service"
	infraresolver "github.com/jorelcb/codify/internal/infrastructure/resolver"
)

// resolveDefineMarkers is the CLI adapter for the post-generation resolve
// hook. It assembles the dependencies (huh-based prompter, the active LLM
// provider, the LLM-driven enricher) and delegates to
// command.ResolveMarkersCommand for the actual orchestration.
//
// Returns nil when the environment is non-interactive (no TTY) — the same
// behavior as the legacy resolver. Errors from Execute are surfaced to the
// caller, which treats them as a warning, not a hard failure of the parent
// `generate` flow.
func resolveDefineMarkers(ctx context.Context, files []string, locale string, provider service.LLMProvider) error {
	if !isInteractive() {
		return nil
	}
	cmd := command.NewResolveMarkersCommand(NewHuhPrompter(), provider).
		WithEnricher(infraresolver.NewLLMEnricher(provider)).
		WithPreviewer(NewHuhDiffPreviewer())
	_, err := cmd.Execute(ctx, command.ResolveRequest{
		Files:  files,
		Locale: locale,
	})
	return err
}
