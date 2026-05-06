package commands

import (
	"fmt"
	"strconv"
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
// Two UIs:
//   - Enriched (Phase 3): when marker.Question is non-empty, render the
//     LLM-derived question + numbered suggestions + default. Accepts a
//     numeric pick (1-N), free text, Enter for default-or-skip, or "s" /
//     "skip" for explicit skip.
//   - Legacy: when no enrichment is available (nil enricher, LLM failure,
//     sanitizer rejected everything), render the marker text and ask for
//     free-text input, same as pre-Phase-3.
func (p *HuhPrompter) AskMarker(fileContent string, marker service.EnrichedMarker) (service.PromptedAnswer, error) {
	showMarkerLineContext(fileContent, marker.Line)

	if marker.Question != "" {
		return p.askEnriched(marker)
	}
	return p.askLegacy(marker)
}

func (p *HuhPrompter) askLegacy(marker service.EnrichedMarker) (service.PromptedAnswer, error) {
	ans, err := promptInput(fmt.Sprintf("Your input for L%d (Enter to skip)", marker.Line), "")
	if err != nil {
		return service.PromptedAnswer{}, err
	}
	trimmed := strings.TrimSpace(ans)
	return service.PromptedAnswer{Answer: trimmed, Skip: trimmed == ""}, nil
}

func (p *HuhPrompter) askEnriched(marker service.EnrichedMarker) (service.PromptedAnswer, error) {
	fmt.Println()
	fmt.Printf("    %s\n", marker.Question)
	if marker.Rationale != "" {
		fmt.Printf("    (%s)\n", marker.Rationale)
	}
	if len(marker.Suggestions) > 0 {
		fmt.Println("    Suggestions:")
		for i, s := range marker.Suggestions {
			tag := ""
			if s == marker.Default {
				tag = " [default]"
			}
			fmt.Printf("      %d) %s%s\n", i+1, s, tag)
		}
	}

	hint := "1-N, text, Enter to skip"
	if marker.Default != "" {
		hint = "1-N, text, Enter for default, s to skip"
	}
	raw, err := promptInput(fmt.Sprintf("Your answer (%s)", hint), "")
	if err != nil {
		return service.PromptedAnswer{}, err
	}
	return ParseEnrichedInput(raw, marker.Suggestions, marker.Default), nil
}

// ParseEnrichedInput converts the user's raw input into a PromptedAnswer
// using the suggestions and default associated with the marker. Pure
// function — exported for direct unit testing.
//
// Accepted forms:
//   - empty input    -> default if non-empty, else skip
//   - "s" / "skip"   -> explicit skip (case-insensitive, trimmed)
//   - integer 1..N   -> picks the Nth suggestion (1-based)
//   - anything else  -> free text answer
func ParseEnrichedInput(raw string, suggestions []string, def string) service.PromptedAnswer {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		if def != "" {
			return service.PromptedAnswer{Answer: def}
		}
		return service.PromptedAnswer{Skip: true}
	}
	low := strings.ToLower(trimmed)
	if low == "s" || low == "skip" {
		return service.PromptedAnswer{Skip: true}
	}
	// Numeric pick within the suggestion range.
	if n, err := strconv.Atoi(trimmed); err == nil && n >= 1 && n <= len(suggestions) {
		return service.PromptedAnswer{Answer: suggestions[n-1]}
	}
	return service.PromptedAnswer{Answer: trimmed}
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
