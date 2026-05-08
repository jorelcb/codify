package dto

import (
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/domain/shared"
)

// SpecConfig holds configuration for generating SDD specifications.
//
// Layout/FeatureID/StandardID son opcionales — si Layout es la zero value
// (LayoutFlat) el comportamiento es el histórico de OpenSpec: archivos
// directamente bajo specs/. Cuando Layout es LayoutFeatureGrouped, los
// archivos se escriben bajo specs/<FeatureID>/. StandardID se persiste para
// que logs y consumidores aguas abajo sepan qué adapter generó los archivos.
type SpecConfig struct {
	ProjectName     string
	FromContextPath string // path to existing output directory (contains AGENTS.md and context/)
	OutputPath      string
	Model           string
	Locale          string

	// Layout describe la organización en disco (Flat vs FeatureGrouped).
	// Determinado por el SpecStandard activo.
	Layout service.OutputLayout

	// FeatureID es el subdir bajo specs/ cuando Layout=LayoutFeatureGrouped.
	// Vacío para Flat. El caller tipicamente lo deriva del projectName.
	FeatureID string

	// StandardID identifica el SpecStandard activo (e.g., "openspec",
	// "spec-kit"). Útil para logs y validaciones aguas abajo.
	StandardID string

	// StandardHints es el bloque que el SpecStandard activo agrega al
	// system prompt para reforzar convenciones. Vacío para OpenSpec
	// (su formato es la base implícita); no-vacío para Spec-Kit.
	StandardHints string
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
	if sc.Layout == service.LayoutFeatureGrouped && sc.FeatureID == "" {
		return shared.ErrInvalidInput("feature id is required for feature-grouped layout")
	}
	return nil
}
