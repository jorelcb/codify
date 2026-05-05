package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	driftdomain "github.com/jorelcb/codify/internal/domain/drift"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	infradrift "github.com/jorelcb/codify/internal/infrastructure/drift"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
)

// NewUpdateCmd construye `codify update` — refresh selectivo de artefactos
// cuando la realidad del proyecto cambió.
//
// Flujo:
//   1. Run check internamente para identificar drift.
//   2. Si no hay drift significativo, salir sin invocar LLM.
//   3. Si hay drift en signals (e.g. go.mod cambió), correr analyze para
//      regenerar context con la realidad nueva.
//   4. Si solo hay drift de artifact_modified (usuario editó a mano), recomendar
//      reset-state en lugar de regenerar (la edición probablemente era intencional).
//
// El costo LLM se registra en usage.json. Use --dry-run para previsualizar
// sin disparar la regeneración.
func NewUpdateCmd() *cobra.Command {
	var (
		outputPath  string
		dryRun      bool
		force       bool
		noTracking  bool
		acceptCurr  bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh artifacts when input signals (go.mod, Makefile, etc.) have drifted",
		Long: `Selective regeneration of generated artifacts based on detected drift.

Logic:
  - If no drift  → no-op, no LLM call, exit 0
  - Signal drift → run analyze internally to refresh artifacts (LLM cost)
  - Only artifact_modified (user hand-edits)  → suggest 'codify reset-state'
    instead, since regenerating would lose the edits

The command records LLM usage in .codify/usage.json and ~/.codify/usage.json
unless tracking is disabled (--no-tracking, CODIFY_NO_USAGE_TRACKING=1, or
~/.codify/.no-usage-tracking marker).

Flags:
  --dry-run         Show what would change without running the LLM
  --force           Regenerate even if drift is only minor (artifact_new, etc.)
  --accept-current  Treat hand-edits as the new baseline (calls reset-state internally)
  --no-tracking     Skip recording this invocation to usage.json (does not affect prior entries)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(outputPath, dryRun, force, noTracking, acceptCurr)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Directory where artifacts live")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would change without running the LLM")
	cmd.Flags().BoolVar(&force, "force", false, "Regenerate even on minor drift (otherwise minor drift is ignored)")
	cmd.Flags().BoolVar(&noTracking, "no-tracking", false, "Skip usage tracking for this invocation")
	cmd.Flags().BoolVar(&acceptCurr, "accept-current", false, "Accept the current FS as the new baseline (alias for 'codify reset-state')")
	return cmd
}

func runUpdate(outputPath string, dryRun, force, noTracking, acceptCurrent bool) error {
	if noTracking {
		_ = os.Setenv("CODIFY_NO_USAGE_TRACKING", "1")
	}

	if acceptCurrent {
		fmt.Println("Accepting current FS as baseline (delegating to reset-state)...")
		return runResetState(outputPath, dryRun)
	}

	statePath, err := infraconfig.ProjectStatePath()
	if err != nil {
		return err
	}
	state, exists, err := infrastate.NewRepository().Load(statePath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "No snapshot at %s. Run 'codify init', 'codify generate', or 'codify analyze' first.\n", statePath)
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

	if report.IsEmpty() {
		fmt.Println("✓ No drift detected. Nothing to update.")
		return nil
	}

	// Si solo hay artifact_modified (sin signal_changed), eso es hand-edit;
	// la respuesta correcta es accept-current, no regenerar.
	if !force && onlyArtifactEdits(report) {
		fmt.Println("Only hand-edits to generated artifacts detected (no upstream signal changes).")
		fmt.Println("Regenerating would overwrite your edits. Run one of:")
		fmt.Println("  codify update --accept-current   # treat current FS as the new baseline")
		fmt.Println("  codify reset-state               # equivalent — recompute snapshot only")
		fmt.Println("  codify update --force            # regenerate anyway (loses edits)")
		os.Exit(1)
	}

	if !report.HasSignificant() && !force {
		fmt.Println("Only minor drift detected. Use --force to regenerate anyway.")
		return nil
	}

	if dryRun {
		fmt.Println("Dry run — would run 'codify analyze' to refresh artifacts based on:")
		printDriftSummary(report)
		return nil
	}

	fmt.Println("Drift detected. Refreshing artifacts via analyze...")
	printDriftSummary(report)
	fmt.Println()

	// Delegar a analyze: scaneamos cwd, generamos los artefactos con la
	// realidad actual, sobreescribimos. analyze ya escribe state.json al
	// terminar (vía writeProjectSnapshot), por lo que el snapshot queda
	// consistente.
	return runAnalyzeFromInit(".", state.Project.Name, state.Project.Language, "", state.Project.Preset, state.Project.Locale, outputPath)
}

// onlyArtifactEdits retorna true si todas las entries son artifact_modified
// (sin signals cambiando). Esto identifica el caso "el usuario editó a mano".
func onlyArtifactEdits(report driftdomain.Report) bool {
	if report.IsEmpty() {
		return false
	}
	for _, e := range report.Entries {
		if e.Kind != driftdomain.ArtifactModified {
			return false
		}
	}
	return true
}

// printDriftSummary imprime las entries de un report en formato una-línea-por-entry,
// para que update muestre qué disparó la regeneración.
func printDriftSummary(report driftdomain.Report) {
	for _, e := range report.Entries {
		fmt.Printf("  [%s] %s — %s\n", e.Kind, e.Path, e.Detail)
	}
}
