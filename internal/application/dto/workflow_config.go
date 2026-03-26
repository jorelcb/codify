package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// WorkflowConfig holds configuration for generating Antigravity workflows.
type WorkflowConfig struct {
	Category       string // "workflows"
	Preset         string // "feature-development", "bug-fix", "release-cycle", "all"
	Mode           string // "static" or "personalized"
	Locale         string // "en" or "es"
	Model          string // LLM model (personalized mode only)
	OutputPath     string
	ProjectContext string // project description (personalized mode only)
	Install        string // "global", "project", or "" (custom)
}

// Validate validates the workflow configuration.
func (wc *WorkflowConfig) Validate() error {
	if wc.Category == "" {
		return shared.ErrInvalidInput("workflow category is required")
	}
	if wc.Preset == "" {
		return shared.ErrInvalidInput("workflow preset is required")
	}
	if wc.Mode == "" {
		return shared.ErrInvalidInput("workflow mode is required")
	}
	if wc.Mode != SkillModeStatic && wc.Mode != SkillModePersonalized {
		return shared.ErrInvalidInput("workflow mode must be 'static' or 'personalized'")
	}
	if wc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	if wc.Mode == SkillModePersonalized && wc.ProjectContext == "" {
		return shared.ErrInvalidInput("project context is required for personalized mode")
	}
	return nil
}
