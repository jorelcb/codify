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

// scopeLabel humanizes a dto.InstallScope* constant for prompt copy.
func scopeLabel(scope string) string {
	if scope == dto.InstallScopeGlobal {
		return "globally"
	}
	return "for this project"
}

// promptInstallSkills offers the user a chance to install skills at the given
// scope (global or project), one catalog category at a time. Each prompt has
// "skip" as the default — running through with all skips installs nothing.
//
// Skills are installed in static mode (no LLM, no API key needed). Power users
// who want personalized skills can run `codify skills` later.
func promptInstallSkills(target, locale, scope string) error {
	if !isInteractive() {
		return nil
	}

	skillsPath := skillsPathForScope(target, scope)

	fmt.Println()
	fmt.Printf("Skills (%s, optional)\n", scopeLabel(scope))
	fmt.Println("─────────────────────────────")
	fmt.Printf("Skills are installed to %s. Each preset is a curated bundle of related SKILL.md files.\n", skillsPath)
	fmt.Println("Pick one preset per category, or skip. You can revisit later with 'codify skills'.")

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
		if err := installSkill(target, locale, cat.Name, preset, scope); err != nil {
			fmt.Printf("  ✗ %s/%s install failed: %v\n", cat.Name, preset, err)
			continue
		}
	}
	return nil
}

// buildCategoryPresetOptions builds a select-list of presets for a skill
// category, with "Skip" as the first (and default) option. Each label is
// annotated with the file count so the user knows how many SKILL.md files
// the bundle contains.
func buildCategoryPresetOptions(cat catalog.SkillCategory) []selectOption {
	options := []selectOption{
		{"Skip — don't install this category now", "skip"},
	}
	for _, opt := range cat.Options {
		count := len(opt.TemplateMapping)
		label := fmt.Sprintf("%s — %d skill(s)", opt.Label, count)
		options = append(options, selectOption{label, opt.Name})
	}
	if !cat.Exclusive {
		options = append(options, selectOption{"All presets in this category", "all"})
	}
	return options
}

// installSkill executes a static-mode skills install at the given scope,
// reusing the same pipeline as `codify skills --install <scope>`.
func installSkill(target, locale, categoryName, preset, scope string) error {
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

	output := skillsPathForScope(target, scope)
	config := &dto.SkillsConfig{
		Category:   cat.Name,
		Preset:     preset,
		Mode:       dto.SkillModeStatic,
		Locale:     locale,
		Target:     target,
		OutputPath: output,
		Install:    scope,
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

func skillsPathForScope(target, scope string) string {
	if scope == dto.InstallScopeGlobal {
		return globalSkillsPath(target)
	}
	return defaultSkillsPath(target)
}

// promptInstallHooks offers the user a chance to install Claude Code hook
// bundles at the given scope (global or project). Skipping installs nothing.
//
// Hooks are Claude-only (Codex/Antigravity have no equivalent), so callers
// should gate this on target == "claude" before invoking.
func promptInstallHooks(locale, scope string) error {
	if !isInteractive() {
		return nil
	}

	settingsPath := "~/.claude/settings.json"
	hooksDir := "~/.claude/hooks/"
	if scope == dto.InstallScopeProject {
		settingsPath = ".claude/settings.json"
		hooksDir = ".claude/hooks/"
	}

	fmt.Println()
	fmt.Printf("Hooks (%s, Claude Code only, optional)\n", scopeLabel(scope))
	fmt.Println("──────────────────────────────────────────")
	fmt.Println("Hooks are deterministic guardrails (linting, security checks, commit conventions).")
	fmt.Printf("They merge into %s + copy scripts to %s.\n", settingsPath, hooksDir)

	preset, err := promptSelect("Hook bundle to install", []selectOption{
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
		Install:  scope,
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

// promptInstallWorkflows offers the user a chance to install workflow bundles
// at the given scope. Skipping installs nothing. Workflows currently target
// Claude Code (`.claude/skills/`, `~/.claude/skills/`) and Antigravity
// (`.agent/workflows/`, `~/.gemini/antigravity/global_workflows/`); Codex is
// not supported, so callers should gate on target before invoking.
func promptInstallWorkflows(target, locale, scope string) error {
	if !isInteractive() {
		return nil
	}
	if target != "claude" && target != "antigravity" {
		return nil
	}

	output := workflowsPathForScope(target, scope)

	fmt.Println()
	fmt.Printf("Workflows (%s, optional)\n", scopeLabel(scope))
	fmt.Println("────────────────────────────────")
	fmt.Printf("Workflows are multi-step lifecycle skills (bug-fix, release-cycle, spec-driven-change).\n")
	fmt.Printf("They install to %s.\n", output)

	cat := &catalog.WorkflowCategories[0] // single category: "workflows"
	options := []selectOption{
		{"Skip — don't install workflows now", "skip"},
	}
	for _, opt := range cat.Options {
		count := len(opt.TemplateMapping)
		label := fmt.Sprintf("%s — %d workflow(s)", opt.Label, count)
		options = append(options, selectOption{label, opt.Name})
	}
	options = append(options, selectOption{"All workflows (bug-fix + release-cycle + spec-driven-change)", "all"})

	preset, err := promptSelect("Workflow bundle to install", options, "skip")
	if err != nil {
		return err
	}
	if preset == "skip" {
		return nil
	}

	if err := installWorkflow(target, locale, preset, scope); err != nil {
		fmt.Printf("  ✗ workflows/%s install failed: %v\n", preset, err)
	}
	return nil
}

// installWorkflow executes a static-mode workflows install at the given scope.
func installWorkflow(target, locale, preset, scope string) error {
	cat, err := catalog.FindWorkflowCategory("workflows")
	if err != nil {
		return err
	}

	var selection *catalog.ResolvedSelection
	if preset == "all" {
		selection = catalog.ResolveAllWorkflows()
	} else {
		selection, err = cat.Resolve(preset)
		if err != nil {
			return err
		}
	}

	// Workflow templates live at templates/{locale}/workflows/ — selection.TemplateDir
	// is the literal "workflows" directory. Same path for claude and antigravity;
	// the deliver command handles target-specific frontmatter rendering.
	templatePath := filepath.Join("templates", locale, selection.TemplateDir)
	loader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, selection.TemplateMapping)
	guides, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("load workflow templates: %w", err)
	}

	output := workflowsPathForScope(target, scope)
	config := &dto.WorkflowConfig{
		Category:   "workflows",
		Preset:     preset,
		Mode:       dto.SkillModeStatic,
		Target:     target,
		Locale:     locale,
		OutputPath: output,
		Install:    scope,
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliver := command.NewDeliverStaticWorkflowsCommand(fileWriter, dirManager)

	result, err := deliver.Execute(config, guides)
	if err != nil {
		return err
	}

	fmt.Printf("  ✓ workflows/%s installed (%d file(s)) → %s\n", preset, len(result.GeneratedFiles), result.OutputPath)
	return nil
}

func workflowsPathForScope(target, scope string) string {
	if scope == dto.InstallScopeGlobal {
		return globalWorkflowsPath(target)
	}
	return defaultWorkflowsPath(target)
}
