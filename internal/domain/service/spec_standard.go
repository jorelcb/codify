package service

// SpecStandard abstrae el formato de SDD (Spec-Driven Development) que codify
// produce. Diferentes estándares (OpenSpec, GitHub Spec-Kit, custom internos)
// difieren en file names, layout en disco, y workflows del lifecycle.
//
// Ver ADR-0011 para razones, alternatives, y consequences.
//
// Adapters concretos viven en internal/infrastructure/sdd/. La selección del
// estándar activo se resuelve por precedencia: flag CLI > project config >
// user config > built-in default (OpenSpec).
type SpecStandard interface {
	// ID returns a stable identifier used in flags and config (e.g.,
	// "openspec", "spec-kit"). Lowercase, no spaces.
	ID() string

	// DisplayName returns a human-readable name for prompts and CLI output.
	DisplayName() string

	// BootstrapArtifacts returns the artifacts that `codify spec` produces.
	// The slice order also drives the order of file generation and listing.
	BootstrapArtifacts() []SpecArtifact

	// OutputLayout describes how spec output is organized on disk.
	// LayoutFlat = specs/<file>.md (OpenSpec).
	// LayoutFeatureGrouped = specs/<feature>/<file>.md (Spec-Kit).
	OutputLayout() OutputLayout

	// TemplateDir returns the directory name under templates/{locale}/sdd/
	// where this standard's templates live (e.g., "openspec", "spec-kit").
	// The full path is templates/{locale}/sdd/{TemplateDir()}/{kind}/...
	// where kind is "spec" or "workflows".
	TemplateDir() string

	// SystemPromptHints returns standard-specific instructions appended to
	// the base spec system prompt. May vary by locale for consistent tone.
	// OpenSpec returns delta-format reminders; Spec-Kit returns per-feature
	// directory conventions; etc.
	SystemPromptHints(locale string) string

	// LifecycleWorkflowIDs returns the workflow guide IDs that implement
	// this standard's lifecycle. Consumed by the workflows command when the
	// user installs the spec-driven-change preset to know which workflow
	// templates to ship for the active standard.
	//
	// OpenSpec → ["spec_propose", "spec_apply", "spec_archive"].
	// Spec-Kit → ["specify", "plan", "tasks"].
	LifecycleWorkflowIDs() []string
}

// SpecArtifact describes one file that `codify spec` generates. The triplet
// (GuideName, FileName, Required) is sufficient for the loader, prompt
// builder, and output writer to do their jobs without standard-specific
// branching elsewhere.
type SpecArtifact struct {
	// GuideName matches the template guide identifier used by the prompt
	// builder (e.g., "constitution", "spec", "plan", "tasks", "research").
	GuideName string

	// FileName is the on-disk file name (e.g., "CONSTITUTION.md", "spec.md").
	// Case matters — OpenSpec uses uppercase, Spec-Kit uses lowercase.
	FileName string

	// Required indicates whether the spec command must generate this file.
	// Optional artifacts (Spec-Kit's research.md, etc.) may be skipped.
	Required bool
}

// OutputLayout enumerates the supported spec output organizations.
type OutputLayout int

const (
	// LayoutFlat puts all spec files directly under specs/. Used by OpenSpec
	// and the original v1.x layout.
	LayoutFlat OutputLayout = iota

	// LayoutFeatureGrouped puts spec files under specs/<feature-id>/. Used
	// by GitHub Spec-Kit.
	LayoutFeatureGrouped
)
