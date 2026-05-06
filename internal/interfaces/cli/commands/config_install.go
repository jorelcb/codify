package commands

import (
	"fmt"
	"path/filepath"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// promptInstallGlobalSkills offers the user a chance to install global skills
// during first-time setup, one category at a time. Each category is asked
// independently with a "skip" option as the safe default — running the
// wizard with all skips installs nothing.
//
// Skills are installed in static mode (no LLM, no API key). Power users who
// want personalized skills can run `codify skills` later.
func promptInstallGlobalSkills(target, locale string) error {
	if !isInteractive() {
		return nil
	}

	fmt.Println()
	fmt.Println("Global skills (optional)")
	fmt.Println("────────────────────────")
	fmt.Printf("Skills are installed to %s and become available across all your projects.\n", globalSkillsPath(target))
	fmt.Println("Pick one preset per category, or skip. You can revisit with 'codify skills --install global'.")

	for _, cat := range catalog.Categories {
		preset, err := promptSelect(
			fmt.Sprintf("Skills — %s", cat.Label),
			buildCategoryPresetOptions(cat),
			"skip",
		)
		if err != nil {
			return err
		}
		if preset == "skip" {
			continue
		}
		if err := installGlobalSkill(target, locale, cat.Name, preset); err != nil {
			fmt.Printf("  ✗ %s/%s install failed: %v\n", cat.Name, preset, err)
			continue
		}
	}
	return nil
}

// buildCategoryPresetOptions builds a select-list of presets for a skill
// category, with "Skip" as the first (and default) option.
func buildCategoryPresetOptions(cat catalog.SkillCategory) []selectOption {
	options := []selectOption{
		{"Skip — don't install this category now", "skip"},
	}
	for _, opt := range cat.Options {
		options = append(options, selectOption{opt.Label, opt.Name})
	}
	if !cat.Exclusive {
		options = append(options, selectOption{"All presets in this category", "all"})
	}
	return options
}

// installGlobalSkill executes a static-mode skills install at global scope,
// reusing the same pipeline that `codify skills --install global` uses.
func installGlobalSkill(target, locale, categoryName, preset string) error {
	cat, err := catalog.FindCategory(categoryName)
	if err != nil {
		return err
	}
	selection, err := cat.Resolve(preset)
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", locale, "skills", selection.TemplateDir)
	loader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, selection.TemplateMapping)
	guides, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("load skill templates: %w", err)
	}

	output := globalSkillsPath(target)
	config := &dto.SkillsConfig{
		Category:   cat.Name,
		Preset:     preset,
		Mode:       dto.SkillModeStatic,
		Locale:     locale,
		Target:     target,
		OutputPath: output,
		Install:    dto.InstallScopeGlobal,
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliver := command.NewDeliverStaticSkillsCommand(fileWriter, dirManager)

	result, err := deliver.Execute(config, guides)
	if err != nil {
		return err
	}

	fmt.Printf("  ✓ %s/%s installed (%d file(s)) → %s\n", cat.Name, preset, len(result.GeneratedFiles), result.OutputPath)
	return nil
}

// promptInstallGlobalHooks offers the user a chance to install Claude Code
// hook bundles globally during first-time setup. Skipping installs nothing.
//
// Hooks are Claude-only (Codex/Antigravity have no equivalent), so callers
// should gate this on target == "claude" before invoking.
func promptInstallGlobalHooks(locale string) error {
	if !isInteractive() {
		return nil
	}

	fmt.Println()
	fmt.Println("Global hooks (optional, Claude Code only)")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Hooks are deterministic guardrails (linting, security checks, commit conventions).")
	fmt.Println("They merge into ~/.claude/settings.json + copy scripts to ~/.claude/hooks/.")

	preset, err := promptSelect("Hook bundle to install globally", []selectOption{
		{"Skip — don't install hooks now", "skip"},
		{"linting (auto-format on Edit/Write)", "linting"},
		{"security-guardrails (block dangerous commands)", "security-guardrails"},
		{"convention-enforcement (validate commits + protect main)", "convention-enforcement"},
		{"all (linting + security-guardrails + convention-enforcement)", "all"},
	}, "skip")
	if err != nil {
		return err
	}
	if preset == "skip" {
		return nil
	}

	config := &dto.HookConfig{
		Category: "hooks",
		Preset:   preset,
		Locale:   locale,
		Install:  dto.InstallScopeGlobal,
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliverer := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)
	installer := command.NewInstallHooksCommand(deliverer, fileWriter, dirManager)

	result, err := installer.Execute(config)
	if err != nil {
		return err
	}

	fmt.Printf("  ✓ hooks/%s installed → %s\n", preset, result.SettingsPath)
	if total := sumMap(result.HandlersAdded); total > 0 {
		fmt.Printf("  ✓ %d handler(s) added across %d event(s)\n", total, len(result.HandlersAdded))
	}
	if len(result.ScriptsCopied) > 0 {
		fmt.Printf("  ✓ %d script(s) copied to %s\n", len(result.ScriptsCopied), result.HooksDir)
	}
	return nil
}
