// Package state define el modelo de snapshot persistido en .codify/state.json.
//
// El state.json captura el estado del proyecto al momento de generación
// (init/generate/analyze). En v1.22 solo se ESCRIBE; los lifecycle commands
// (`check`, `update`, `audit`, `watch`) lo CONSUMEN a partir de v1.23.
//
// Schema documentado en docs/adr/0004-state-json-schema.md.
package state

// SchemaVersion del state.json. Incrementar major en cambios breaking.
const SchemaVersion = "1.0"

// State es el snapshot completo persistido a .codify/state.json.
type State struct {
	SchemaVersion string `json:"schema_version"`
	CodifyVersion string `json:"codify_version"`
	GeneratedAt   string `json:"generated_at"`
	GeneratedBy   string `json:"generated_by"` // command que generó: "init", "generate", "analyze"

	Project ProjectInfo `json:"project"`
	Git     GitInfo     `json:"git,omitempty"`

	Artifacts    map[string]ArtifactInfo `json:"artifacts,omitempty"`
	InputSignals map[string]SignalInfo   `json:"input_signals,omitempty"`

	SkillsInstalled    []InstalledItem `json:"skills_installed,omitempty"`
	WorkflowsInstalled []InstalledItem `json:"workflows_installed,omitempty"`
	HooksInstalled     []InstalledItem `json:"hooks_installed,omitempty"`
}

// ProjectInfo captura los parámetros de configuración del proyecto.
type ProjectInfo struct {
	Name     string `json:"name"`
	Preset   string `json:"preset"`
	Language string `json:"language,omitempty"`
	Locale   string `json:"locale"`
	Target   string `json:"target,omitempty"`
	Kind     string `json:"kind,omitempty"` // "new" o "existing"
}

// GitInfo captura el contexto git al momento del snapshot.
type GitInfo struct {
	Commit  string `json:"commit,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Remote  string `json:"remote,omitempty"`
	IsDirty bool   `json:"is_dirty,omitempty"`
}

// ArtifactInfo describe un archivo generado por Codify.
type ArtifactInfo struct {
	SHA256        string   `json:"sha256"`
	GeneratedAt   string   `json:"generated_at"`
	SizeBytes     int64    `json:"size_bytes"`
	GeneratedFrom []string `json:"generated_from,omitempty"`
}

// SignalInfo describe un input signal observado al momento de generación.
// Los lifecycle commands los comparan contra el estado actual para detectar
// drift relevante.
type SignalInfo struct {
	SHA256 string `json:"sha256"`
	// Campos adicionales según el tipo de signal: deps_count, targets_count, lines.
	DepsCount    int `json:"deps_count,omitempty"`
	TargetsCount int `json:"targets_count,omitempty"`
	Lines        int `json:"lines,omitempty"`
}

// InstalledItem describe un skill/workflow/hook instalado por init.
type InstalledItem struct {
	Category string `json:"category,omitempty"`
	Preset   string `json:"preset,omitempty"`
	Scope    string `json:"scope,omitempty"` // "global" o "project"
	Path     string `json:"path,omitempty"`
}

// New devuelve un State zero con SchemaVersion ya seteado.
func New() State {
	return State{
		SchemaVersion: SchemaVersion,
	}
}
