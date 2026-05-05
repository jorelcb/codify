package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
)

// NewResetStateCmd construye `codify reset-state` — recompute .codify/state.json
// from the current FS without touching any artifact.
//
// Use case: el usuario editó intencionalmente AGENTS.md y quiere "aceptar"
// las ediciones como el nuevo baseline. En lugar de regenerar (caro, requiere
// LLM), reset-state solo recalcula los hashes y persiste — rapidísimo, cero
// costo.
func NewResetStateCmd() *cobra.Command {
	var (
		outputPath string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "reset-state",
		Short: "Recompute .codify/state.json from the current FS without touching artifacts",
		Long: `Recompute the project snapshot stored at .codify/state.json based on the
current state of generated artifacts and input signals. Useful when:

  - You intentionally edited AGENTS.md or context/*.md and want to accept those
    edits as the new baseline (instead of regenerating with an LLM)
  - state.json got corrupted or out of sync

The command is read-only over your artifacts: it never modifies AGENTS.md,
CONTEXT.md, etc. It only updates state.json. Existing state.json is backed up
to state.json.bak (atomic write).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResetState(outputPath, dryRun)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Directory where artifacts live (default: current dir)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the recomputed snapshot without writing it")
	return cmd
}

func runResetState(outputPath string, dryRun bool) error {
	statePath, err := infraconfig.ProjectStatePath()
	if err != nil {
		return err
	}

	stateRepo := infrastate.NewRepository()

	// Cargar el state existente para preservar Project info (preset, target, etc.)
	// — el reset solo refresca hashes y signals, no la metadata del proyecto.
	prev, exists, err := stateRepo.Load(statePath)
	if err != nil {
		return fmt.Errorf("load existing state: %w", err)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "No existing snapshot at %s.\n", statePath)
		fmt.Fprintln(os.Stderr, "Run 'codify init' or 'codify generate' to create one. reset-state only refreshes an existing snapshot.")
		os.Exit(2)
	}

	// Recomputar snapshot preservando ProjectInfo previa
	cfg, err := infraconfig.NewRepository().LoadEffective()
	if err != nil {
		return fmt.Errorf("load effective config: %w", err)
	}
	_ = cfg

	if dryRun {
		fmt.Printf("Would recompute state.json at: %s\n", statePath)
		fmt.Printf("  Project: %s\n", prev.Project.Name)
		fmt.Printf("  Preset:  %s\n", prev.Project.Preset)
		fmt.Println("  (dry-run — no changes made)")
		return nil
	}

	// Reusar el helper que ya hace todo bien
	writeProjectSnapshot("reset-state", prev.Project.Name, prev.Project.Preset, prev.Project.Language, prev.Project.Locale, prev.Project.Target, prev.Project.Kind, outputPath)

	fmt.Println()
	fmt.Println("Snapshot refreshed. Subsequent 'codify check' runs will compare against the new baseline.")
	return nil
}
