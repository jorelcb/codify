package commands

import (
	"fmt"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// HuhPrompter implements service.InteractivePrompter using charmbracelet/huh
// for the terminal UI. It is the default prompter wired by the CLI when
// invoking ResolveMarkersCommand.
//
// Phase 0 keeps the same UX as the legacy resolver: top-level confirm,
// per-file header, surrounding context display, plain-text input prompt.
// Phase 3 will replace AskMarker with the enriched UI (numbered suggestions,
// default, context-aware help).
type HuhPrompter struct{}

// NewHuhPrompter returns a prompter ready to be passed to
// command.NewResolveMarkersCommand.
func NewHuhPrompter() *HuhPrompter {
	return &HuhPrompter{}
}

// ConfirmTopLevel asks the global "resolve N markers across M files?" prompt.
func (p *HuhPrompter) ConfirmTopLevel(totalMarkers, totalFiles int) (bool, error) {
	fmt.Println()
	fmt.Printf("Found %d [DEFINE] marker(s) across %d file(s).\n", totalMarkers, totalFiles)
	proceed, err := promptConfirm("Resolve them interactively now?", true)
	if err != nil || !proceed {
		fmt.Println("Skipped. Markers remain in the files for manual editing.")
		return false, nil
	}
	return true, nil
}

// AnnounceFile prints the per-file header before its markers are walked.
func (p *HuhPrompter) AnnounceFile(path string, markerCount int) {
	fmt.Println()
	fmt.Printf("── %s (%d marker%s) ──\n", path, markerCount, pluralS(markerCount))
}

// AskMarker shows the surrounding context and prompts for the user's input.
// Empty input is treated as Skip=true.
func (p *HuhPrompter) AskMarker(fileContent string, marker service.EnrichedMarker) (service.PromptedAnswer, error) {
	showMarkerLineContext(fileContent, marker.Line)
	ans, err := promptInput(fmt.Sprintf("Your input for L%d (Enter to skip)", marker.Line), "")
	if err != nil {
		return service.PromptedAnswer{}, err
	}
	trimmed := strings.TrimSpace(ans)
	return service.PromptedAnswer{Answer: trimmed, Skip: trimmed == ""}, nil
}

// ReportFileResult prints the per-file outcome line.
func (p *HuhPrompter) ReportFileResult(path string, resolved int, mode string) {
	switch mode {
	case "unchanged":
		fmt.Printf("  (no answers — file unchanged)\n")
	case "llm":
		fmt.Printf("  ✓ %d marker(s) resolved via LLM rewrite\n", resolved)
	case "literal":
		fmt.Printf("  ✓ %d marker(s) resolved via literal substitution\n", resolved)
	default:
		fmt.Printf("  ✓ %d marker(s) resolved\n", resolved)
	}
}

// showMarkerLineContext prints the marker line plus a few lines of surrounding
// context so the user can decide what to fill in without alt-tabbing to an
// editor. Same UX as the legacy resolver.
func showMarkerLineContext(content string, line int) {
	const radius = 2
	lines := strings.Split(content, "\n")
	from := line - 1 - radius
	if from < 0 {
		from = 0
	}
	to := line - 1 + radius
	if to >= len(lines) {
		to = len(lines) - 1
	}
	fmt.Println()
	for i := from; i <= to; i++ {
		marker := "  "
		if i == line-1 {
			marker = "▸ "
		}
		fmt.Printf("    %s%4d  %s\n", marker, i+1, lines[i])
	}
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
