package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// ProjectConfig representa la configuración para generar un proyecto
type ProjectConfig struct {
	Name         string
	Description  string
	Language     string
	Type         string
	Architecture string
	OutputPath   string
	Model        string
	Locale       string
	Capabilities []string
	Metadata     map[string]string
}

// Validate valida la configuración del proyecto
func (pc *ProjectConfig) Validate() error {
	if pc.Name == "" {
		return shared.ErrInvalidInput("project name is required")
	}
	if pc.Description == "" {
		return shared.ErrInvalidInput("project description is required")
	}
	if pc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	return nil
}