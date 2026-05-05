// Package drift implementa la detección de drift comparando un snapshot
// persistido contra el estado actual del FS.
package drift

import (
	"fmt"

	domain "github.com/jorelcb/codify/internal/domain/drift"
	statedomain "github.com/jorelcb/codify/internal/domain/state"
	"github.com/jorelcb/codify/internal/infrastructure/snapshot"
)

// Detector es la API pública para correr drift detection. No mantiene estado
// — todas las operaciones son sobre el snapshot que se le pasa.
type Detector struct{}

// NewDetector devuelve una instancia lista para usar.
func NewDetector() *Detector {
	return &Detector{}
}

// DetectOptions parametriza la detección.
type DetectOptions struct {
	Snapshot    statedomain.State
	ProjectPath string // raíz del proyecto (cwd típicamente)
	OutputPath  string // donde estaban los artefactos al momento del snapshot
}

// Detect compara el snapshot contra el estado actual del FS y devuelve un
// Report con todas las divergencias encontradas. Determinístico — sin red,
// sin LLM, sin side effects.
func (d *Detector) Detect(opts DetectOptions) (domain.Report, error) {
	report := domain.Report{}

	// 1. Reconstruir snapshot actual en memoria
	current, err := snapshot.Build(snapshot.BuildOptions{
		ProjectPath: opts.ProjectPath,
		OutputPath:  opts.OutputPath,
		Project:     opts.Snapshot.Project,
		GeneratedBy: "check",
	})
	if err != nil {
		return report, fmt.Errorf("rebuild snapshot: %w", err)
	}

	// 2. Comparar artefactos
	report.Entries = append(report.Entries, diffArtifacts(opts.Snapshot.Artifacts, current.Artifacts)...)

	// 3. Comparar input signals
	report.Entries = append(report.Entries, diffSignals(opts.Snapshot.InputSignals, current.InputSignals)...)

	return report, nil
}

func diffArtifacts(prev, curr map[string]statedomain.ArtifactInfo) []domain.Entry {
	entries := []domain.Entry{}

	// Modified or missing
	for path, p := range prev {
		c, ok := curr[path]
		if !ok {
			entries = append(entries, domain.Entry{
				Kind:     domain.ArtifactMissing,
				Severity: domain.SeverityOf(domain.ArtifactMissing),
				Path:     path,
				Detail:   "artifact present in snapshot but missing on disk",
			})
			continue
		}
		if p.SHA256 != c.SHA256 {
			entries = append(entries, domain.Entry{
				Kind:     domain.ArtifactModified,
				Severity: domain.SeverityOf(domain.ArtifactModified),
				Path:     path,
				Detail:   fmt.Sprintf("content hash changed (was %s, now %s)", short(p.SHA256), short(c.SHA256)),
			})
		}
	}
	// New artifacts (not in prev)
	for path := range curr {
		if _, ok := prev[path]; !ok {
			entries = append(entries, domain.Entry{
				Kind:     domain.ArtifactNew,
				Severity: domain.SeverityOf(domain.ArtifactNew),
				Path:     path,
				Detail:   "artifact appeared on disk but is not in the snapshot",
			})
		}
	}
	return entries
}

func diffSignals(prev, curr map[string]statedomain.SignalInfo) []domain.Entry {
	entries := []domain.Entry{}

	// Changed or removed
	for name, p := range prev {
		c, ok := curr[name]
		if !ok {
			entries = append(entries, domain.Entry{
				Kind:     domain.SignalRemoved,
				Severity: domain.SeverityOf(domain.SignalRemoved),
				Path:     name,
				Detail:   "input signal present in snapshot but missing on disk",
			})
			continue
		}
		if p.SHA256 != c.SHA256 {
			entries = append(entries, domain.Entry{
				Kind:     domain.SignalChanged,
				Severity: domain.SeverityOf(domain.SignalChanged),
				Path:     name,
				Detail:   fmt.Sprintf("content hash changed (was %s, now %s)", short(p.SHA256), short(c.SHA256)),
			})
		}
	}
	// Added (new in current)
	for name := range curr {
		if _, ok := prev[name]; !ok {
			entries = append(entries, domain.Entry{
				Kind:     domain.SignalAdded,
				Severity: domain.SeverityOf(domain.SignalAdded),
				Path:     name,
				Detail:   "input signal appeared on disk but is not in the snapshot",
			})
		}
	}
	return entries
}

// short trunca un hash hex a 7 chars para reportes legibles.
func short(hash string) string {
	if len(hash) <= 7 {
		return hash
	}
	return hash[:7]
}
