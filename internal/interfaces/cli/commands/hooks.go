package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
)

// hooksParams groups all parameters for the hooks command.
type hooksParams struct {
	preset  string
	locale  string
	output  string
	install string
}

// NewHooksCmd creates the hooks command.
func NewHooksCmd() *cobra.Command {
	var p hooksParams

	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Generate Claude Code hook bundles (deterministic guardrails)",
		Long: `Generate Claude Code hook bundles — shell scripts wired into events
(PreToolUse, PostToolUse, etc.) that run deterministically on every tool call.

Hooks complement skills (prompt-based) and workflows (orchestration):
  • skills    → tell Claude HOW to do something (prompts)
  • workflows → coordinate multi-step processes (lifecycle)
  • hooks     → enforce rules every single time (deterministic, exit codes)

Presets:
  linting                 - Auto-format and lint files on Edit/Write (PostToolUse)
  security-guardrails     - Block dangerous commands and protect sensitive files (PreToolUse)
  convention-enforcement  - Validate Conventional Commits and protect main branches
  all                     - All three presets merged into a single hooks.json

Output layout:
  {output}/hooks.json   - hook configuration block to merge into settings.json
  {output}/hooks/*.sh   - auxiliary scripts referenced by hooks.json

To activate the hooks:
  1. Move the scripts:    cp -r {output}/hooks/ ~/.claude/hooks/        (global)
                       or cp -r {output}/hooks/ .claude/hooks/          (project)
  2. Merge {output}/hooks.json into ~/.claude/settings.json (global)
                                  or .claude/settings.json (project)

Note: hooks are a Claude Code feature; this command does not target Antigravity
or Codex. Personalization is not supported — hooks are catalog-driven.

Requirements:
  - bash (Linux/macOS native; Windows requires Git Bash or WSL)
  - jq (used by all generated scripts to parse JSON input)
  - Claude Code v2.1.85+ for the convention-enforcement preset (uses 'if' field)

Examples:
  # Interactive mode (guided selection)
  codify hooks

  # Generate the linting bundle into ./codify-hooks/
  codify hooks --preset linting

  # Generate everything, install layout pre-built for the project scope
  codify hooks --preset all --install project

  # Custom output directory
  codify hooks --preset security-guardrails --output ./tmp/sec-hooks`,
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
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Output directory (default: ./codify-hooks)")
	cmd.Flags().StringVar(&p.install, "install", "", "Install scope: global or project (or omit for custom --output)")

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

	// 3. Resolve install scope and output path.
	install := p.install
	output := p.output

	if !explicit["install"] && !explicit["output"] && interactive {
		globalPath := globalHooksPath()
		projectPath := defaultHooksPath()

		var location string
		location, err = promptSelect("Output location", []selectOption{
			{fmt.Sprintf("Project (%s)", projectPath), "project"},
			{fmt.Sprintf("Global (%s)", globalPath), "global"},
			{"Custom output directory", "custom"},
		}, "project")
		if err != nil {
			return err
		}

		switch location {
		case "global":
			install = dto.InstallScopeGlobal
			output = globalPath
		case "project":
			install = dto.InstallScopeProject
			output = projectPath
		default:
			output, err = promptInput("Output directory", defaultHooksPath())
			if err != nil {
				return err
			}
		}
	} else if explicit["install"] {
		output = resolveHookInstallPath(install)
	} else if output == "" {
		output = defaultHooksPath()
	}

	// 4. Build config and execute.
	config := &dto.HookConfig{
		Category:   "hooks",
		Preset:     preset,
		Locale:     locale,
		OutputPath: output,
		Install:    install,
	}

	return executeHooks(config)
}

func executeHooks(config *dto.HookConfig) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)

	fmt.Println()
	fmt.Printf("Generating Claude Code hook bundle (static)\n")
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Install scope: %s\n", config.Install)
	}
	fmt.Println()

	result, err := cmd.Execute(config)
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
	fmt.Println("To activate:")
	fmt.Printf("  1. Copy scripts:    cp -r %s/hooks/ <YOUR_CLAUDE_DIR>/hooks/\n", result.OutputPath)
	fmt.Printf("  2. Merge hook block from %s/hooks.json into <YOUR_CLAUDE_DIR>/settings.json\n", result.OutputPath)
	fmt.Println()
	fmt.Println("Where <YOUR_CLAUDE_DIR> is:")
	fmt.Println("  ~/.claude    (global, applies to all projects)")
	fmt.Println("  .claude      (project-local, can be committed to the repo)")
	fmt.Println()
	fmt.Println("Verify activation: run /hooks inside Claude Code.")

	return nil
}

// --- Path resolution helpers ---

// defaultHooksPath returns the default project-local output path.
//
// We deliberately use ./codify-hooks/ instead of .claude/hooks/ so the user
// immediately sees the directory and understands these files require manual
// merging into settings.json. Codify never auto-activates hooks.
func defaultHooksPath() string {
	return filepath.Join(".", "codify-hooks")
}

// globalHooksPath returns the default global (per-user) output path.
func globalHooksPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, "codify-hooks")
}

// resolveHookInstallPath maps an install scope to its concrete output path.
func resolveHookInstallPath(install string) string {
	switch install {
	case dto.InstallScopeGlobal:
		return globalHooksPath()
	case dto.InstallScopeProject:
		return defaultHooksPath()
	default:
		return defaultHooksPath()
	}
}
