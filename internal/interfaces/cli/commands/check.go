package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	driftdomain "github.com/jorelcb/codify/internal/domain/drift"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	infradrift "github.com/jorelcb/codify/internal/infrastructure/drift"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
)

// NewCheckCmd construye `codify check` — drift detection determinista.
//
// Salida:
//   - exit 0 si no hay drift, o si solo hay drifts minor sin --strict
//   - exit 1 si hay drifts significativos (default) o cualquier drift (con --strict)
//   - exit 2 si no hay state.json (proyecto sin bootstrap previo)
//
// El comando es **read-only y sin LLM**. Costo: cero.
func NewCheckCmd() *cobra.Command {
	var (
		strict     bool
		outputPath string
		jsonOut    bool
	)

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Detect drift between .codify/state.json and the current project state",
		Long: `Compare the project's persisted snapshot (.codify/state.json) against the
current filesystem state. Reports any divergence:

  - artifact_modified  AGENTS.md or context/*.md was edited after generation
  - artifact_missing   a generated file no longer exists on disk
  - artifact_new       a new artifact appeared since the snapshot
  - signal_changed     an input signal (go.mod, Makefile, README.md, etc.) changed
  - signal_added       a new signal appeared
  - signal_removed     an input signal disappeared

The command is fully deterministic — no LLM calls, no network, no costs.
Suitable for CI: returns exit 1 on significant drift (or any drift with --strict).

Run 'codify reset-state' to recompute the snapshot from the current FS without
modifying any artifact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(strict, outputPath, jsonOut)
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Treat any drift (including minor) as a failure (exit 1)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Directory where artifacts were generated (default: current dir)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit the drift report as JSON instead of human-readable text")
	return cmd
}

func runCheck(strict bool, outputPath string, jsonOut bool) error {
	statePath, err := infraconfig.ProjectStatePath()
	if err != nil {
		return err
	}

	state, exists, err := infrastate.NewRepository().Load(statePath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "No snapshot found at %s. Run 'codify init', 'codify generate', or 'codify analyze' first to bootstrap the project.\n", statePath)
		os.Exit(2)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	report, err := infradrift.NewDetector().Detect(infradrift.DetectOptions{
		Snapshot:    state,
		ProjectPath: cwd,
		OutputPath:  outputPath,
	})
	if err != nil {
		return fmt.Errorf("drift detection: %w", err)
	}

	if jsonOut {
		emitJSON(report)
	} else {
		emitHuman(report, statePath)
	}

	if shouldFail(report, strict) {
		os.Exit(1)
	}
	return nil
}

// shouldFail determina el exit code según el modo y el contenido del reporte.
//
//   - strict=true  → fail si hay CUALQUIER entry (significant o minor)
//   - strict=false → fail solo si hay alguna entry significant
//   - sin entries  → no fail (exit 0)
func shouldFail(report driftdomain.Report, strict bool) bool {
	if report.IsEmpty() {
		return false
	}
	if strict {
		return true
	}
	return report.HasSignificant()
}

// emitHuman imprime el reporte de drift en formato legible. Las entradas se
// agrupan por kind para facilitar la lectura, y se sortean por path dentro
// de cada grupo para reproducibilidad entre runs.
func emitHuman(report driftdomain.Report, statePath string) {
	if report.IsEmpty() {
		fmt.Println("✓ No drift detected. Snapshot is up to date.")
		return
	}

	groups := map[driftdomain.Kind][]driftdomain.Entry{}
	for _, e := range report.Entries {
		groups[e.Kind] = append(groups[e.Kind], e)
	}

	// Imprimir significant primero, después minor — más útil que orden alfabético
	order := []driftdomain.Kind{
		driftdomain.SignalRemoved,
		driftdomain.SignalChanged,
		driftdomain.ArtifactMissing,
		driftdomain.ArtifactModified,
		driftdomain.SignalAdded,
		driftdomain.ArtifactNew,
	}
	fmt.Printf("Drift detected (snapshot: %s)\n\n", statePath)
	for _, kind := range order {
		entries := groups[kind]
		if len(entries) == 0 {
			continue
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
		fmt.Printf("  [%s] (%s)\n", kind, entries[0].Severity)
		for _, e := range entries {
			fmt.Printf("    - %s — %s\n", e.Path, e.Detail)
		}
		fmt.Println()
	}

	if report.HasSignificant() {
		fmt.Println("Recommended next steps:")
		fmt.Println("  - Review the drift above to confirm intent")
		fmt.Println("  - For artifact_modified: revert hand edits, or run 'codify reset-state' to accept the new content as the new baseline")
		fmt.Println("  - For signal_changed: regenerate context (e.g. 'codify generate' or 'codify analyze') so artifacts match the new project state")
	} else {
		fmt.Println("All drift is minor — review at your discretion.")
	}
}

// emitJSON imprime el reporte como JSON para consumo programático en CI.
// El formato es deliberadamente delgado: array de objetos con kind/severity/
// path/detail. No incluye metadata del state.json — usar `cat .codify/state.json`
// para eso.
func emitJSON(report driftdomain.Report) {
	type out struct {
		Kind     string `json:"kind"`
		Severity string `json:"severity"`
		Path     string `json:"path"`
		Detail   string `json:"detail"`
	}
	items := make([]out, 0, len(report.Entries))
	for _, e := range report.Entries {
		items = append(items, out{
			Kind:     string(e.Kind),
			Severity: string(e.Severity),
			Path:     e.Path,
			Detail:   e.Detail,
		})
	}
	encodeJSON(items)
}
