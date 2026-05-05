package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// HookConfig holds configuration for generating Claude Code hook bundles.
//
// Hooks are Claude Code-specific: there is no equivalent feature in
// Antigravity or Codex, so HookConfig has no Target field. Likewise,
// hooks are catalog-driven (no LLM personalization), so there is no
// Mode, Model, or ProjectContext field.
//
// As of v1.20.0 the default flow is auto-activation: when Install is
// "global" or "project", the command merges into ~/.claude/settings.json
// (or .claude/settings.json) and copies scripts into ~/.claude/hooks/
// (or .claude/hooks/). OutputPath is used only as an escape hatch:
// preview/dry mode that writes a bundle to a custom directory without
// touching settings.
type HookConfig struct {
	Category   string // "hooks"
	Preset     string // "linting" | "security-guardrails" | "convention-enforcement" | "all"
	Locale     string // "en" or "es"
	OutputPath string // optional: when set with empty Install, runs in preview mode
	Install    string // install scope: "global", "project", or "" (preview/custom)
	DryRun     bool   // when true, prints the proposed merge but writes nothing
}

// ValidHookPresets enumerates the preset names accepted by the hooks command.
var ValidHookPresets = map[string]bool{
	"linting":                true,
	"security-guardrails":    true,
	"convention-enforcement": true,
	"all":                    true,
}

// Validate validates the hook configuration.
//
// Either Install ("global"/"project") or OutputPath (preview mode) must be
// set. Both may be set simultaneously when the caller wants preview output
// alongside an install scope, but at least one is required.
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
	if hc.Install == "" && hc.OutputPath == "" {
		return shared.ErrInvalidInput("either --install (global|project) or --output is required")
	}
	if hc.Install != "" && hc.Install != InstallScopeGlobal && hc.Install != InstallScopeProject {
		return shared.ErrInvalidInput("invalid install scope: must be 'global' or 'project'")
	}
	if hc.Locale == "" {
		return shared.ErrInvalidInput("locale is required")
	}
	if hc.Locale != "en" && hc.Locale != "es" {
		return shared.ErrInvalidInput("locale must be 'en' or 'es'")
	}
	return nil
}
