package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	statedomain "github.com/jorelcb/codify/internal/domain/state"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	infradrift "github.com/jorelcb/codify/internal/infrastructure/drift"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
	"github.com/jorelcb/codify/internal/infrastructure/watch"
)

// NewWatchCmd construye `codify watch` — foreground file watcher que dispara
// drift detection cuando los paths registrados en .codify/state.json cambian.
//
// Diseño documentado en docs/adr/0008-watch-model-decision.md.
func NewWatchCmd() *cobra.Command {
	var (
		debounce   time.Duration
		autoUpdate bool
		strict     bool
		noTracking bool
		outputPath string
	)

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Foreground watcher: detect drift while you edit",
		Long: `Foreground file watcher that monitors paths registered in .codify/state.json
(input_signals + artifacts) and re-runs drift detection on change.

Designed for active development sessions, NOT as a background daemon. Runs in
foreground and exits cleanly on Ctrl+C. For persistence wrap with tmux/nohup/
systemd — Codify intentionally does not handle daemonization (see ADR-008).

Behavior:
  - Loads .codify/state.json once at startup; exits 2 if missing
  - Subscribes via fsnotify to the parent dirs of registered paths
  - Debounces events (default 2s) before firing drift detection
  - Prints drift reports to stdout, keeps watching
  - --auto-update fires 'codify update' on detected drift (records LLM usage)
  - Without --auto-update, drift is informational; user runs check/update manually

Examples:
  codify watch                         # default 2s debounce, report-only
  codify watch --debounce 500ms        # tighter debounce for fast feedback
  codify watch --auto-update --strict  # aggressively keep artifacts in sync

Alternative — for git-hook validation use 'codify check' wired into a tool
like lefthook, pre-commit, or watchexec. See README "Lifecycle" section for
example configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(debounce, autoUpdate, strict, noTracking, outputPath)
		},
	}

	cmd.Flags().DurationVar(&debounce, "debounce", 2*time.Second, "Quiet time before firing drift detection (e.g. 500ms, 2s, 5s)")
	cmd.Flags().BoolVar(&autoUpdate, "auto-update", false, "Run 'codify update' on drift instead of just reporting (records LLM usage)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat any drift (including minor) as actionable")
	cmd.Flags().BoolVar(&noTracking, "no-tracking", false, "Skip usage tracking when --auto-update fires LLM calls")
	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Directory where artifacts live")
	return cmd
}

func runWatch(debounce time.Duration, autoUpdate, strict, noTracking bool, outputPath string) error {
	if noTracking {
		_ = os.Setenv("CODIFY_NO_USAGE_TRACKING", "1")
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

	paths := collectWatchPaths(state, cwd, outputPath)
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "No watchable paths in state.json. Snapshot may be empty or corrupted.")
		os.Exit(2)
	}
	// Watch state.json itself so user notices if it's removed/replaced.
	paths = append(paths, statePath)

	fmt.Printf("Codify watch — debounce=%s, auto-update=%v, strict=%v\n", debounce, autoUpdate, strict)
	fmt.Printf("Watching %d paths from %s\n", len(paths), statePath)
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	w, err := watch.New(watch.Options{
		Paths:    paths,
		Debounce: debounce,
		OnEvent: func(ev watch.Event) {
			handleWatchEvent(ev, state, cwd, outputPath, autoUpdate, strict)
		},
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		},
	})
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	if err := w.Start(ctx); err != nil {
		return fmt.Errorf("watcher: %w", err)
	}
	fmt.Println("\n✓ Watcher stopped cleanly.")
	return nil
}

// collectWatchPaths construye la lista de archivos a monitorear a partir del
// snapshot. Para artifacts: keys son relativas al output dir, prefijadas con
// outputPath. Para input_signals: keys son relativas al project root (cwd).
func collectWatchPaths(state statedomain.State, cwd, outputPath string) []string {
	paths := []string{}
	for name := range state.Artifacts {
		paths = append(paths, filepath.Join(outputPath, name))
	}
	for name := range state.InputSignals {
		paths = append(paths, filepath.Join(cwd, name))
	}
	return paths
}

// handleWatchEvent re-runs drift detection on the current FS state and prints
// the result. If autoUpdate is set and there's significant drift, fires update.
func handleWatchEvent(ev watch.Event, state statedomain.State, cwd, outputPath string, autoUpdate, strict bool) {
	timestamp := ev.Triggered.Format("15:04:05")
	fmt.Printf("[%s] change detected (%d files):\n", timestamp, len(ev.Paths))
	for _, p := range ev.Paths {
		// Print as relative path when possible for readability.
		if rel, err := filepath.Rel(cwd, p); err == nil {
			fmt.Printf("  - %s\n", rel)
		} else {
			fmt.Printf("  - %s\n", p)
		}
	}

	report, err := infradrift.NewDetector().Detect(infradrift.DetectOptions{
		Snapshot:    state,
		ProjectPath: cwd,
		OutputPath:  outputPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "  drift detection failed: %v\n", err)
		fmt.Println()
		return
	}

	if report.IsEmpty() {
		fmt.Println("  no drift")
		fmt.Println()
		return
	}

	for _, e := range report.Entries {
		fmt.Printf("  [%s] (%s) %s — %s\n", e.Kind, e.Severity, e.Path, e.Detail)
	}

	if autoUpdate && (report.HasSignificant() || strict) {
		fmt.Println()
		fmt.Println("→ Auto-updating artifacts (--auto-update)…")
		if err := runAnalyzeFromInit(".", state.Project.Name, state.Project.Language, "", state.Project.Preset, state.Project.Locale, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "  update failed: %v\n", err)
		}
	}
	fmt.Println()
}
