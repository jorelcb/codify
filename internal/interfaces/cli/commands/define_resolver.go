package commands

import (
	"context"

	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/domain/service"
)

// resolveDefineMarkers is the CLI adapter for the post-generation resolve
// hook. It assembles the dependencies (huh-based prompter, the active LLM
// provider) and delegates to command.ResolveMarkersCommand for the actual
// orchestration.
//
// Kept as a free function (not a method on a type) because the call site in
// generate.go is a single line and the adapter has no state of its own.
//
// Returns nil when the environment is non-interactive (no TTY) — the same
// behavior as the legacy resolver. Errors from Execute are surfaced to the
// caller, which treats them as a warning, not a hard failure of the parent
// `generate` flow.
func resolveDefineMarkers(ctx context.Context, files []string, locale string, provider service.LLMProvider) error {
	if !isInteractive() {
		return nil
	}
	cmd := command.NewResolveMarkersCommand(NewHuhPrompter(), provider)
	_, err := cmd.Execute(ctx, command.ResolveRequest{
		Files:  files,
		Locale: locale,
	})
	return err
}
