package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// SpecConfig holds configuration for generating SDD specifications
type SpecConfig struct {
	ProjectName     string
	FromContextPath string // path to existing output directory (contains AGENTS.md and context/)
	OutputPath      string
	Model           string
	Locale          string
}

// Validate validates the spec configuration
func (sc *SpecConfig) Validate() error {
	if sc.ProjectName == "" {
		return shared.ErrInvalidInput("project name is required")
	}
	if sc.FromContextPath == "" {
		return shared.ErrInvalidInput("from-context path is required")
	}
	if sc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	return nil
}
