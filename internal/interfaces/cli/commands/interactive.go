package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
)

// selectOption represents an option in an interactive menu.
type selectOption struct {
	Label string
	Value string
}

// isInteractive verifica que stdin y stdout sean terminales TTY.
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// promptSelect displays an interactive selection menu.
func promptSelect(title string, options []selectOption, defaultVal string) (string, error) {
	huhOpts := make([]huh.Option[string], len(options))
	for i, o := range options {
		huhOpts[i] = huh.NewOption(o.Label, o.Value)
	}

	selected := defaultVal
	err := huh.NewSelect[string]().
		Title(title).
		Options(huhOpts...).
		Value(&selected).
		Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled")
	}
	return selected, nil
}

// promptInput muestra un campo de entrada de texto interactivo.
func promptInput(title, defaultVal string) (string, error) {
	value := defaultVal
	err := huh.NewInput().
		Title(title).
		Value(&value).
		Run()
	if err != nil {
		return "", fmt.Errorf("input cancelled")
	}
	if value == "" {
		return defaultVal, nil
	}
	return value, nil
}

// promptConfirm displays an interactive boolean confirmation.
func promptConfirm(title string, defaultVal bool) (bool, error) {
	value := defaultVal
	err := huh.NewConfirm().
		Title(title).
		Value(&value).
		Run()
	if err != nil {
		return defaultVal, fmt.Errorf("confirmation cancelled")
	}
	return value, nil
}

// promptModel displays LLM model selection based on available API keys.
//
// Only models with their corresponding API key set in the environment are
// shown. If no key is set, returns a hard error so the user fixes their
// environment instead of seeing a "false affordance" — picking a model
// they cannot use and hitting an opaque API error later.
func promptModel() (string, error) {
	var options []selectOption
	hasAnthropic := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasGemini := os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GOOGLE_API_KEY") != ""

	if hasAnthropic {
		options = append(options, selectOption{"Claude Sonnet 4.6 (Anthropic)", "claude-sonnet-4-6"})
		options = append(options, selectOption{"Claude Opus 4.6 (Anthropic)", "claude-opus-4-6"})
	}
	if hasGemini {
		options = append(options, selectOption{"Gemini 3.1 Pro Preview (Google)", "gemini-3.1-pro-preview"})
	}

	if len(options) == 0 {
		return "", fmt.Errorf("no LLM API key found in environment; set ANTHROPIC_API_KEY or GEMINI_API_KEY (or GOOGLE_API_KEY) and re-run")
	}

	if len(options) == 1 {
		return options[0].Value, nil
	}

	return promptSelect("Select LLM model", options, options[0].Value)
}

// promptLocale displays output language selection.
func promptLocale() (string, error) {
	return promptSelect("Select output language", []selectOption{
		{"English", "en"},
		{"Spanish", "es"},
	}, "en")
}

// promptPreset displays template preset selection.
//
// Order: neutral first (recommended for greenfield), then opinionated presets
// in increasing complexity. Aligns with ADR-001: in v2.0 the default flips
// from clean-ddd to neutral; we already promote neutral as the recommended
// option in the menu.
func promptPreset() (string, error) {
	return promptSelect("Select template preset", []selectOption{
		{"Neutral (no architectural opinions — recommended)", "neutral"},
		{"Clean + DDD (DDD / Clean Architecture / BDD)", "clean-ddd"},
		{"Hexagonal (Ports & Adapters — lighter than clean-ddd)", "hexagonal"},
		{"Event-Driven (CQRS + Event Sourcing + Sagas)", "event-driven"},
	}, "clean-ddd")
}

// promptLanguage displays programming language selection.
func promptLanguage() (string, error) {
	return promptSelect("Select programming language", []selectOption{
		{"Go", "go"},
		{"JavaScript / TypeScript", "javascript"},
		{"Python", "python"},
		{"None (skip idiomatic guides)", ""},
	}, "")
}
