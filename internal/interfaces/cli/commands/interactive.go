package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
)

// selectOption representa una opción en un menú interactivo.
type selectOption struct {
	Label string
	Value string
}

// isInteractive verifica que stdin y stdout sean terminales TTY.
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// promptSelect muestra un menú de selección interactivo.
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

// promptConfirm muestra una confirmación booleana interactiva.
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

// promptModel muestra selección de modelo LLM basada en API keys disponibles.
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
		options = []selectOption{
			{"Claude Sonnet 4.6 (Anthropic)", "claude-sonnet-4-6"},
			{"Claude Opus 4.6 (Anthropic)", "claude-opus-4-6"},
			{"Gemini 3.1 Pro Preview (Google)", "gemini-3.1-pro-preview"},
		}
	}

	if len(options) == 1 {
		return options[0].Value, nil
	}

	return promptSelect("Select LLM model", options, options[0].Value)
}

// promptLocale muestra selección de idioma de salida.
func promptLocale() (string, error) {
	return promptSelect("Select output language", []selectOption{
		{"English", "en"},
		{"Spanish", "es"},
	}, "en")
}

// promptPreset muestra selección de preset de templates.
func promptPreset() (string, error) {
	return promptSelect("Select template preset", []selectOption{
		{"Default (DDD / Clean Architecture / BDD)", "default"},
		{"Neutral (no architectural opinions)", "neutral"},
	}, "default")
}

// promptLanguage muestra selección de lenguaje de programación.
func promptLanguage() (string, error) {
	return promptSelect("Select programming language", []selectOption{
		{"Go", "go"},
		{"JavaScript / TypeScript", "javascript"},
		{"Python", "python"},
		{"None (skip idiomatic guides)", ""},
	}, "")
}
