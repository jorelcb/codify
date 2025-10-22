package dto

import "github.com/jorelcb/ai-context-generator/internal/domain/shared"

// ProjectConfig representa la configuración para generar un proyecto
type ProjectConfig struct {
	Name         string
	Language     string
	Type         string
	Architecture string
	OutputPath   string
	Capabilities []string
	Metadata     map[string]string
}

// Validate valida la configuración del proyecto
func (pc *ProjectConfig) Validate() error {
	if pc.Name == "" {
		return shared.ErrInvalidInput("project name is required")
	}
	if pc.Language == "" {
		return shared.ErrInvalidInput("language is required")
	}
	if pc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	return nil
}