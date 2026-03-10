package dto

import "github.com/jorelcb/ai-context-generator/internal/domain/shared"

// SpecConfig representa la configuracion para generar especificaciones SDD
type SpecConfig struct {
	ProjectName     string
	FromContextPath string // ruta al directorio de output existente (contiene AGENTS.md y context/)
	OutputPath      string
	Model           string
	Locale          string
}

// Validate valida la configuracion de spec
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
