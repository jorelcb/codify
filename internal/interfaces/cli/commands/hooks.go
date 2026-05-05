package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/settings"
)

// hooksParams groups all parameters for the hooks command.
type hooksParams struct {
	preset  string
	locale  string
	output  string
	install string
	dryRun  bool
}

// NewHooksCmd creates the hooks command.
func NewHooksCmd() *cobra.Command {
	var p hooksParams

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Activate Claude Code hook bundles (deterministic guardrails)",
		Long: `Activate Claude Code hook bundles — shell scripts wired into events
(PreToolUse, PostToolUse, etc.) that run deterministically on every tool call.

Hooks complement skills (prompt-based) and workflows (orchestration):
  • skills    → tell Claude HOW to do something (prompts)
  • workflows → coordinate multi-step processes (lifecycle)
  • hooks     → enforce rules every single time (deterministic, exit codes)

Presets:
  linting                 - Auto-format and lint files on Edit/Write (PostToolUse)
  security-guardrails     - Block dangerous commands and protect sensitive files (PreToolUse)
  convention-enforcement  - Validate Conventional Commits and protect main branches
  all                     - All three presets merged into a single hook block

Activation modes:
  --install project   Auto-merge into .claude/settings.json + copy scripts to .claude/hooks/
  --install global    Auto-merge into ~/.claude/settings.json + copy to ~/.claude/hooks/
  --output PATH       Preview mode: write a standalone bundle (no settings change)
  --dry-run           Print the proposed merge to stdout, write nothing

Auto-install (default flow as of v1.20.0):
  - Backs up the existing settings.json before any modification
  - Idempotent: running twice with the same preset adds zero handlers the second time
  - Only writes scripts that do not already exist; conflicting scripts are reported

Note: hooks are a Claude Code feature; this command does not target Antigravity
or Codex. Personalization is not supported — hooks are catalog-driven.

Requirements:
  - bash (Linux/macOS native; Windows requires Git Bash or WSL)
  - jq (used by all generated scripts to parse JSON input)
  - Claude Code v2.1.85+ for the convention-enforcement preset (uses 'if' field)

Examples:
  # Interactive mode (guided selection — auto-installs by default)
  codify hooks

  # Activate the linting preset for this project
  codify hooks --preset linting --install project

  # Activate everything globally
  codify hooks --preset all --install global

  # Preview the bundle without touching settings.json
  codify hooks --preset security-guardrails --output ./tmp/preview

  # See the proposed merge without writing anything
  codify hooks --preset all --install project --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})
			return runHooks(p, explicit)
		},
	}

	cmd.Flags().StringVarP(&p.preset, "preset", "p", "", "Hook preset: linting, security-guardrails, convention-enforcement, or all")
	cmd.Flags().StringVar(&p.locale, "locale", defaultLocale, "Output language for stderr messages: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Preview mode: write standalone bundle to this directory (no settings change)")
	cmd.Flags().StringVar(&p.install, "install", "", "Install scope: global or project (auto-activates immediately)")
	cmd.Flags().BoolVar(&p.dryRun, "dry-run", false, "Print the proposed settings.json merge but write nothing")

	return cmd
}

func runHooks(p hooksParams, explicit map[string]bool) error {
	interactive := isInteractive()
	var err error

	// 1. Resolve preset.
	preset := p.preset
	if !explicit["preset"] && interactive {
		cat, cErr := catalog.FindHookCategory("hooks")
		if cErr != nil {
			return cErr
		}
		options := make([]selectOption, 0, len(cat.Options)+1)
		for _, o := range cat.Options {
			options = append(options, selectOption{o.Label, o.Name})
		}
		options = append(options, selectOption{"All hooks (merged bundle)", "all"})

		preset, err = promptSelect("Select hook preset", options, "")
		if err != nil {
			return err
		}
	}
	if preset == "" {
		return fmt.Errorf("hook preset is required; use --preset flag")
	}
	if !dto.ValidHookPresets[preset] {
		return fmt.Errorf("invalid preset: %s (valid: linting, security-guardrails, convention-enforcement, all)", preset)
	}

	// 2. Resolve locale.
	locale := p.locale
	if !explicit["locale"] && interactive {
		locale, err = promptLocale()
		if err != nil {
			return err
		}
	}
	if locale == "" {
		locale = defaultLocale
	}

	// 3. Resolve activation mode.
	//
	// Priority:
	//   --output → preview mode (no settings change)
	//   --install → auto-install
	//   neither + interactive → ask the user
	//   neither + non-interactive → default to project install
	install := p.install
	output := p.output
	dryRun := p.dryRun

	if !explicit["install"] && !explicit["output"] && interactive {
		var location string
		location, err = promptSelect("Activation mode", []selectOption{
			{"Project (.claude/settings.json + .claude/hooks/)", "project"},
			{"Global (~/.claude/settings.json + ~/.claude/hooks/)", "global"},
			{"Preview only (write a standalone bundle for inspection)", "preview"},
		}, "project")
		if err != nil {
			return err
		}
		switch location {
		case "global":
			install = dto.InstallScopeGlobal
		case "project":
			install = dto.InstallScopeProject
		default:
			output, err = promptInput("Preview output directory", "./codify-hooks")
			if err != nil {
				return err
			}
		}
	} else if !explicit["install"] && !explicit["output"] {
		install = dto.InstallScopeProject
	}

	config := &dto.HookConfig{
		Category:   "hooks",
		Preset:     preset,
		Locale:     locale,
		OutputPath: output,
		Install:    install,
		DryRun:     dryRun,
	}

	if install != "" {
		return executeInstall(config)
	}
	return executePreview(config)
}

// executeInstall auto-activates hooks via settings.json merge.
func executeInstall(config *dto.HookConfig) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliverer := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)
	installer := command.NewInstallHooksCommand(deliverer, fileWriter, dirManager)

	mode := "install"
	if config.DryRun {
		mode = "dry-run"
	}

	fmt.Println()
	fmt.Printf("Activating Claude Code hooks (%s)\n", mode)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Scope: %s\n", config.Install)
	fmt.Println()

	result, err := installer.Execute(config)
	if err != nil {
		return fmt.Errorf("hook activation failed: %w", err)
	}

	if result.DryRun {
		// For dry-run we already showed where it would go. Print the merged
		// settings.json preview by re-running the merge against the loaded
		// state — quickest path is to call PreviewMergedHooks again with a
		// fresh load.
		printDryRunPreview(config, result)
		return nil
	}

	fmt.Println("Hooks activated successfully")
	fmt.Printf("  Settings: %s\n", result.SettingsPath)
	if result.BackupPath != "" {
		fmt.Printf("  Backup:   %s\n", result.BackupPath)
	}
	fmt.Printf("  Hooks dir: %s\n", result.HooksDir)
	if total := sumMap(result.HandlersAdded); total > 0 {
		fmt.Printf("  Added:     %d handler(s) across %d event(s)\n", total, len(result.HandlersAdded))
	}
	if total := sumMap(result.HandlersSkipped); total > 0 {
		fmt.Printf("  Skipped:   %d handler(s) already present\n", total)
	}
	if len(result.ScriptsCopied) > 0 {
		fmt.Printf("  Scripts copied: %d\n", len(result.ScriptsCopied))
		for _, s := range result.ScriptsCopied {
			fmt.Printf("    + %s\n", s)
		}
	}
	if len(result.ScriptsSkipped) > 0 {
		fmt.Printf("  Scripts unchanged: %d (already on disk with identical content)\n", len(result.ScriptsSkipped))
	}
	if len(result.ScriptsConflict) > 0 {
		fmt.Printf("  Scripts in conflict: %d (existing differs — not overwritten)\n", len(result.ScriptsConflict))
		for _, s := range result.ScriptsConflict {
			fmt.Printf("    ! %s\n", s)
		}
	}
	fmt.Println()
	fmt.Println("Verify in Claude Code: run /hooks")
	return nil
}

// executePreview writes a standalone bundle — the v1.19.0 escape hatch.
func executePreview(config *dto.HookConfig) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliverer := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)

	fmt.Println()
	fmt.Printf("Generating Claude Code hook bundle (preview mode)\n")
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	fmt.Println()

	result, err := deliverer.Execute(config)
	if err != nil {
		return fmt.Errorf("hook bundle generation failed: %w", err)
	}

	fmt.Printf("Hook bundle written to %s\n", result.OutputPath)
	fmt.Println()
	fmt.Println("Generated files:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Println()
	fmt.Println("This is preview mode: settings.json was NOT modified.")
	fmt.Println("To activate, re-run with --install project|global, or merge manually.")
	return nil
}

// printDryRunPreview shows what settings.json would look like after the
// merge, without writing anything. The InstallHooksCommand already
// computed the merge, but the preview bytes were emitted internally —
// for the CLI we re-load the settings and run a second preview.
func printDryRunPreview(config *dto.HookConfig, result *command.InstallResult) {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliverer := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)

	bundle, err := deliverer.Build(config.Locale, config.Preset)
	if err != nil {
		fmt.Printf("dry-run preview failed: %v\n", err)
		return
	}
	s, err := settings.Load(result.SettingsPath)
	if err != nil {
		fmt.Printf("dry-run preview failed: %v\n", err)
		return
	}
	out, err := s.PreviewMergedHooks(bundle.HooksDoc)
	if err != nil {
		fmt.Printf("dry-run preview failed: %v\n", err)
		return
	}

	fmt.Println("Proposed settings.json merge:")
	fmt.Println()
	fmt.Println(string(out))
	fmt.Println("(dry-run: nothing was written)")
}

func sumMap(m map[string]int) int {
	t := 0
	for _, v := range m {
		t += v
	}
	return t
}
