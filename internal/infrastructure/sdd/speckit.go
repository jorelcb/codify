package sdd

import (
	"github.com/jorelcb/codify/internal/domain/service"
)

// SpecKitAdapter implementa service.SpecStandard para el formato GitHub
// Spec-Kit (https://github.com/github/spec-kit).
//
// Diferencias clave frente a OpenSpec:
//
//   - Layout: por feature (specs/<feature-id>/<file>.md) en vez de flat.
//   - File names lowercase con guiones (spec.md, plan.md, data-model.md).
//   - No existe "constitution" — la opinión del proyecto vive en otros files.
//   - Lifecycle linear: /specify → /plan → /tasks (sin deltas, sin archive).
//   - Files opcionales adicionales: research.md, data-model.md, quickstart.md.
//
// El adapter es stateless. La selección de feature-id (qué subdir bajo
// specs/) la hace el caller — para `codify spec`, el feature-id por
// default es el projectName slugificado.
type SpecKitAdapter struct{}

// NewSpecKitAdapter construye el adapter.
func NewSpecKitAdapter() *SpecKitAdapter {
	return &SpecKitAdapter{}
}

// ID returns "spec-kit".
func (SpecKitAdapter) ID() string { return "spec-kit" }

// DisplayName returns "GitHub Spec-Kit".
func (SpecKitAdapter) DisplayName() string { return "GitHub Spec-Kit" }

// BootstrapArtifacts returns the Spec-Kit canonical files.
//
// Required (3): spec, plan, tasks — el núcleo del workflow Spec-Kit.
// Optional (3): research, data_model, quickstart — útiles pero no obligatorios.
//
// El orden refleja la secuencia natural: el spec se escribe primero, después
// el plan basado en el spec, después las tasks que ejecutan el plan, y los
// optional alimentan los tres principales.
func (SpecKitAdapter) BootstrapArtifacts() []service.SpecArtifact {
	return []service.SpecArtifact{
		{GuideName: "speckit_spec", FileName: "spec.md", Required: true},
		{GuideName: "speckit_plan", FileName: "plan.md", Required: true},
		{GuideName: "speckit_tasks", FileName: "tasks.md", Required: true},
		{GuideName: "speckit_research", FileName: "research.md", Required: false},
		{GuideName: "speckit_data_model", FileName: "data-model.md", Required: false},
		{GuideName: "speckit_quickstart", FileName: "quickstart.md", Required: false},
	}
}

// OutputLayout returns LayoutFeatureGrouped — Spec-Kit escribe bajo
// specs/<feature-id>/, no a nivel raíz de specs/.
func (SpecKitAdapter) OutputLayout() service.OutputLayout {
	return service.LayoutFeatureGrouped
}

// TemplateDir returns "spec-kit". Templates viven en
// templates/{locale}/sdd/spec-kit/{spec,workflows}/.
func (SpecKitAdapter) TemplateDir() string { return "spec-kit" }

// SystemPromptHints returns Spec-Kit-specific guidance appended to the base
// spec system prompt. Refuerza el layout per-feature y los nombres lowercase
// — diferencias críticas frente a OpenSpec que el LLM debe respetar.
func (SpecKitAdapter) SystemPromptHints(locale string) string {
	if locale == "es" {
		return `<sdd_standard_hints>
Estandar activo: GitHub Spec-Kit.
Convenciones obligatorias:
- Cada archivo es lowercase con guiones (spec.md, plan.md, data-model.md, NO mayusculas).
- Los archivos viven bajo specs/<feature-id>/ — NO en la raiz de specs/.
- No emitas un archivo CONSTITUTION.md — Spec-Kit no usa ese concepto.
- El flujo es lineal: spec describe el "que", plan el "como", tasks el "que hacer concretamente".
- research.md, data-model.md y quickstart.md son opcionales y solo se generan si el contexto los justifica.
</sdd_standard_hints>
`
	}
	return `<sdd_standard_hints>
Active standard: GitHub Spec-Kit.
Required conventions:
- File names are lowercase with hyphens (spec.md, plan.md, data-model.md — NEVER uppercase).
- Files live under specs/<feature-id>/ — NEVER at the root of specs/.
- Do NOT emit a CONSTITUTION.md — Spec-Kit does not use that concept.
- Flow is linear: spec captures the "what", plan the "how", tasks the "concrete steps".
- research.md, data-model.md, and quickstart.md are optional — emit only if the context warrants them.
</sdd_standard_hints>
`
}

// LifecycleWorkflowIDs returns the Spec-Kit lifecycle workflows: /specify,
// /plan, /tasks. Estos slash commands son el equivalente de Spec-Kit al
// ciclo propose/apply/archive de OpenSpec, pero con semantics distintas:
// son linear, no hay deltas, y no hay archive (un feature termina cuando
// las tasks se completan, sin merge ceremony).
func (SpecKitAdapter) LifecycleWorkflowIDs() []string {
	return []string{"speckit_specify", "speckit_plan", "speckit_tasks"}
}
