package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// HookConfig holds configuration for generating Claude Code hook bundles.
//
// Hooks are Claude Code-specific: there is no equivalent feature in
// Antigravity or Codex, so HookConfig has no Target field. Likewise,
// hooks are catalog-driven (no LLM personalization), so there is no
// Mode, Model, or ProjectContext field.
type HookConfig struct {
	Category   string // "hooks"
	Preset     string // "linting" | "security-guardrails" | "convention-enforcement" | "all"
	Locale     string // "en" or "es"
	OutputPath string
	Install    string // install scope: "global", "project", or "" (custom output)
}

// ValidHookPresets enumerates the preset names accepted by the hooks command.
var ValidHookPresets = map[string]bool{
	"linting":                true,
	"security-guardrails":    true,
	"convention-enforcement": true,
	"all":                    true,
}

// Validate validates the hook configuration.
func (hc *HookConfig) Validate() error {
	if hc.Category == "" {
		return shared.ErrInvalidInput("hook category is required")
	}
	if hc.Preset == "" {
		return shared.ErrInvalidInput("hook preset is required")
	}
	if !ValidHookPresets[hc.Preset] {
		return shared.ErrInvalidInput("invalid hook preset: must be linting, security-guardrails, convention-enforcement, or all")
	}
	if hc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	if hc.Locale == "" {
		return shared.ErrInvalidInput("locale is required")
	}
	if hc.Locale != "en" && hc.Locale != "es" {
		return shared.ErrInvalidInput("locale must be 'en' or 'es'")
	}
	return nil
}
