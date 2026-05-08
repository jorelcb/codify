// Package sdd contiene los adapters concretos del port SpecStandard
// (ver internal/domain/service/spec_standard.go y ADR-0011).
//
// Cada adapter encapsula el conocimiento específico de un estándar de
// Spec-Driven Development: file names, layout en disco, hints para el LLM,
// y workflows del lifecycle. La selección del adapter activo se resuelve
// en el wiring de la CLI según la precedencia documentada en ADR-0011.
package sdd

import (
	"github.com/jorelcb/codify/internal/domain/service"
)

// OpenSpecAdapter implementa service.SpecStandard para el formato OpenSpec
// (https://github.com/open-rfc/openspec). Es el default histórico y preserva
// 100% el comportamiento de codify v1.x:
//
//   - Bootstrap artifacts: 4 archivos uppercase a nivel raíz de specs/.
//   - Layout: flat (specs/<FILE>.md, no subdirectorios por feature).
//   - Lifecycle workflows: spec_propose / spec_apply / spec_archive con
//     deltas ADDED/MODIFIED/REMOVED y archive YYYY-MM-DD-<id>/.
//
// La implementación es stateless — todas las operaciones son puras y no
// dependen del filesystem. El template path completo lo arma el consumidor
// vía service.SpecStandardTemplatePath.
type OpenSpecAdapter struct{}

// NewOpenSpecAdapter construye el adapter. No tiene dependencias externas.
func NewOpenSpecAdapter() *OpenSpecAdapter {
	return &OpenSpecAdapter{}
}

// ID returns the stable identifier "openspec".
func (OpenSpecAdapter) ID() string { return "openspec" }

// DisplayName returns "OpenSpec".
func (OpenSpecAdapter) DisplayName() string { return "OpenSpec" }

// BootstrapArtifacts returns the four canonical OpenSpec files in fixed
// order: CONSTITUTION → SPEC → PLAN → TASKS. All four are required.
func (OpenSpecAdapter) BootstrapArtifacts() []service.SpecArtifact {
	return []service.SpecArtifact{
		{GuideName: "constitution", FileName: "CONSTITUTION.md", Required: true},
		{GuideName: "spec", FileName: "SPEC.md", Required: true},
		{GuideName: "plan", FileName: "PLAN.md", Required: true},
		{GuideName: "tasks", FileName: "TASKS.md", Required: true},
	}
}

// OutputLayout returns LayoutFlat — OpenSpec writes files directly under
// specs/, no per-feature subdirectories.
func (OpenSpecAdapter) OutputLayout() service.OutputLayout {
	return service.LayoutFlat
}

// TemplateDir returns "openspec". Templates live at
// templates/{locale}/sdd/openspec/{spec,workflows}/.
func (OpenSpecAdapter) TemplateDir() string { return "openspec" }

// SystemPromptHints returns OpenSpec-specific guidance appended to the base
// spec system prompt. Currently empty — the base prompt already encodes the
// behavior expected for OpenSpec output. Reserved for future tightening.
func (OpenSpecAdapter) SystemPromptHints(locale string) string {
	// Intentionally empty for v1: the base BuildSpecSystemPrompt already
	// produces OpenSpec-compatible output. Spec-Kit (C.3) will provide its
	// own non-empty hints (per-feature directory, lower-case file names,
	// etc.).
	return ""
}

// LifecycleWorkflowIDs returns the three OpenSpec lifecycle workflows in
// canonical order: propose → apply → archive.
func (OpenSpecAdapter) LifecycleWorkflowIDs() []string {
	return []string{"spec_propose", "spec_apply", "spec_archive"}
}
