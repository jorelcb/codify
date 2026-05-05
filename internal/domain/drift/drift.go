// Package drift define el modelo de drift detection: cómo se nombra cada
// tipo de divergencia entre un snapshot persistido y el estado actual del FS.
//
// El package es intencionalmente delgado: tipos + categorización. La lógica
// de detección vive en internal/infrastructure/drift, donde puede leer el FS.
package drift

// Kind clasifica el tipo de drift detectado.
type Kind string

const (
	// ArtifactModified — un artefacto generado por Codify cambió desde el snapshot.
	// Causa típica: el usuario editó AGENTS.md a mano. Severity: significant.
	ArtifactModified Kind = "artifact_modified"

	// ArtifactMissing — un artefacto registrado en el snapshot ya no existe en disco.
	// Severity: significant.
	ArtifactMissing Kind = "artifact_missing"

	// ArtifactNew — un archivo en el output dir que no estaba en el snapshot.
	// Severity: minor (útil informarlo pero no necesariamente acciónable).
	ArtifactNew Kind = "artifact_new"

	// SignalChanged — un input signal (go.mod, Makefile, etc.) cambió desde
	// el snapshot. Severity: significant — el contexto generado puede haber
	// quedado desfasado.
	SignalChanged Kind = "signal_changed"

	// SignalAdded — apareció un input signal que no estaba al momento del snapshot
	// (e.g. proyecto Go que ahora también tiene package.json). Severity: minor.
	SignalAdded Kind = "signal_added"

	// SignalRemoved — un input signal del snapshot ya no existe (e.g. el
	// usuario borró Makefile). Severity: significant.
	SignalRemoved Kind = "signal_removed"
)

// Severity refleja qué tan accionable es un drift.
type Severity string

const (
	// Significant: el drift impacta la validez del contexto generado y debería
	// corregirse via `codify update` o regeneración manual.
	Significant Severity = "significant"

	// Minor: drift informativo, no requiere acción inmediata.
	Minor Severity = "minor"
)

// Entry describe una sola divergencia detectada.
type Entry struct {
	Kind     Kind
	Severity Severity
	Path     string // path relativo del archivo afectado
	Detail   string // mensaje legible para el usuario
}

// Report agrupa todos los drifts detectados en un solo run de check.
type Report struct {
	Entries []Entry
}

// HasSignificant reporta si al menos uno de los drifts es de severidad
// significativa. Lo usa el comando `check` para decidir el exit code en
// modo no-strict.
func (r Report) HasSignificant() bool {
	for _, e := range r.Entries {
		if e.Severity == Significant {
			return true
		}
	}
	return false
}

// IsEmpty reporta si no hay drift alguno (entradas vacías).
func (r Report) IsEmpty() bool {
	return len(r.Entries) == 0
}

// SeverityOf devuelve la severidad asociada a un Kind. Centralizado para
// evitar que cada caller tenga que recordarlo.
func SeverityOf(k Kind) Severity {
	switch k {
	case ArtifactModified, ArtifactMissing, SignalChanged, SignalRemoved:
		return Significant
	case ArtifactNew, SignalAdded:
		return Minor
	default:
		return Minor
	}
}
