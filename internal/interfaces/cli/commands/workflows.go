package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// workflowsParams groups all parameters for the workflows command.
type workflowsParams struct {
	preset         string
	mode           string
	target         string
	locale         string
	model          string
	output         string
	install        string
	projectContext string
}

// NewWorkflowsCmd creates the workflows command.
func NewWorkflowsCmd() *cobra.Command {
	var p workflowsParams

	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Generate workflow files for AI agents",
		Long: `Generate workflow files — multi-step recipes that AI agents execute on demand.

Supports two target ecosystems:
  claude       - Claude Code plugins (skills + hooks + agents + scripts)
  antigravity  - Antigravity .md files with execution annotations (// turbo, etc.)

Modes:
  static        - Instant delivery from built-in catalog (no API key needed)
  personalized  - LLM-adapted workflows tailored to your project (requires API key)

Presets:
  feature-development  - Full feature lifecycle (branch → implement → test → PR)
  bug-fix              - Structured bug fix (reproduce → diagnose → fix → test)
  release-cycle        - Release process (version → changelog → tag → deploy)
  all                  - All workflow presets

Install:
  claude:
    global   - Install to ~/.claude/plugins/
    project  - Install to current directory (one plugin dir per workflow)
  antigravity:
    global   - Install to ~/.gemini/antigravity/global_workflows/
    project  - Install to .agent/workflows/

Each Claude plugin includes:
  .claude-plugin/plugin.json  - Plugin manifest
  skills/{preset}/SKILL.md    - Workflow skill (annotations stripped)
  hooks/hooks.json            - Auto-approve, capture, conditional hooks
  agents/workflow-runner.md   - Execution subagent
  scripts/capture-output.sh   - Output capture script (if needed)

When run without flags, an interactive menu is displayed.

Examples:
  # Interactive mode (guided selection)
  codify workflows

  # Claude Code: generate workflow plugins
  codify workflows --preset all --target claude --mode static

  # Claude Code: install globally
  codify workflows --preset all --target claude --mode static --install global

  # Antigravity: generate workflow files
  codify workflows --preset all --target antigravity --mode static

  # Install to current project (Antigravity)
  codify workflows --preset feature-development --target antigravity --mode static --install project

  # Personalized: LLM-adapted to your project
  codify workflows --preset all --target claude --mode personalized --context "Go microservice with CI/CD via GitHub Actions"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})
			return runWorkflows(p, explicit)
		},
	}

	cmd.Flags().StringVarP(&p.preset, "preset", "p", "", "Workflow preset: feature-development, bug-fix, release-cycle, or all")
	cmd.Flags().StringVar(&p.mode, "mode", "", "Generation mode: static (instant) or personalized (LLM)")
	cmd.Flags().StringVar(&p.target, "target", "antigravity", "Target ecosystem: claude or antigravity")
	cmd.Flags().StringVar(&p.locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&p.model, "model", "m", "", "LLM model (only for personalized mode)")
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Output directory")
	cmd.Flags().StringVar(&p.install, "install", "", "Install scope: global or project")
	cmd.Flags().StringVar(&p.projectContext, "context", "", "Project context for personalized mode")

	return cmd
}

func runWorkflows(p workflowsParams, explicit map[string]bool) error {
	ctx := context.Background()
	interactive := isInteractive()
	var err error

	// 1. Resolve preset
	preset := p.preset
	if !explicit["preset"] && interactive {
		cat, cErr := catalog.FindWorkflowCategory("workflows")
		if cErr != nil {
			return cErr
		}
		options := make([]selectOption, 0, len(cat.Options)+1)
		for _, o := range cat.Options {
			options = append(options, selectOption{o.Label, o.Name})
		}
		options = append(options, selectOption{"All workflows", "all"})

		preset, err = promptSelect("Select workflow preset", options, "")
		if err != nil {
			return err
		}
	}
	if preset == "" {
		return fmt.Errorf("workflow preset is required; use --preset flag")
	}

	// 2. Resolve target
	target := p.target
	if !explicit["target"] && interactive {
		target, err = promptSelect("Select target", []selectOption{
			{"Claude Code (via plugin: skills + hooks + agents)", "claude"},
			{"Antigravity (via native workflow file)", "antigravity"},
		}, "antigravity")
		if err != nil {
			return err
		}
	}
	if !dto.ValidWorkflowTargets[target] {
		return fmt.Errorf("invalid target: %s (available: claude, antigravity)", target)
	}

	// 3. Resolve mode
	mode := p.mode
	if !explicit["mode"] && interactive {
		mode, err = promptSelect("Select mode", []selectOption{
			{"Static (instant, no API key needed)", dto.SkillModeStatic},
			{"Personalized (LLM-adapted to your project)", dto.SkillModePersonalized},
		}, dto.SkillModeStatic)
		if err != nil {
			return err
		}
	}
	if mode == "" {
		mode = dto.SkillModeStatic
	}

	// 4. Resolve locale
	locale := p.locale
	if !explicit["locale"] && interactive {
		locale, err = promptLocale()
		if err != nil {
			return err
		}
	}

	// 5. Resolve install scope and output
	install := p.install
	output := p.output

	if !explicit["install"] && !explicit["output"] && interactive {
		globalPath := globalWorkflowsPath(target)
		projectPath := defaultWorkflowsPath(target)

		var location string
		location, err = promptSelect("Install location", []selectOption{
			{fmt.Sprintf("Global (%s)", globalPath), "global"},
			{fmt.Sprintf("Project (%s)", projectPath), "project"},
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
			output, err = promptInput("Output directory", defaultWorkflowsPath(target))
			if err != nil {
				return err
			}
		}
	} else if explicit["install"] {
		output = resolveWorkflowInstallPath(install, target)
	} else if output == "" {
		output = defaultWorkflowsPath(target)
	}

	// 6. Resolve templates from catalog
	cat, err := catalog.FindWorkflowCategory("workflows")
	if err != nil {
		return err
	}

	selection, err := cat.Resolve(preset)
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", locale, selection.TemplateDir)
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, selection.TemplateMapping)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load workflow templates: %w", err)
	}

	// 7. Build config
	config := &dto.WorkflowConfig{
		Category:   "workflows",
		Preset:     preset,
		Mode:       mode,
		Target:     target,
		Locale:     locale,
		OutputPath: output,
		Install:    install,
	}

	// 8. Execute based on mode
	if mode == dto.SkillModeStatic {
		return executeStaticWorkflows(config, guides)
	}
	return executePersonalizedWorkflows(ctx, p, config, guides, explicit, interactive)
}

func executeStaticWorkflows(config *dto.WorkflowConfig, guides []service.TemplateGuide) error {
	if config.Target == "claude" {
		return executeStaticPlugin(config, guides)
	}
	return executeStaticAntigravity(config, guides)
}

func executeStaticPlugin(config *dto.WorkflowConfig, guides []service.TemplateGuide) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	cmd := command.NewDeliverPluginCommand(fileWriter, dirManager, root.TemplatesFS)

	fmt.Println()
	fmt.Printf("Delivering Claude Code workflow plugins (static)\n")
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Install: %s\n", config.Install)
	}
	fmt.Printf("  Plugins: %d\n", len(guides))
	fmt.Println()

	result, err := cmd.Execute(config, guides)
	if err != nil {
		return fmt.Errorf("static plugin delivery failed: %w", err)
	}

	fmt.Printf("Claude Code plugins delivered successfully!\n")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Installed: %s scope\n", config.Install)
	}
	fmt.Println()
	fmt.Println("Generated plugin files:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Println()
	fmt.Println("To use a plugin:")
	for _, guide := range guides {
		pluginName := catalog.PluginName(guide.Name)
		presetDir := catalog.PresetDirName(guide.Name)
		fmt.Printf("  claude --plugin-dir ./%s\n", pluginName)
		fmt.Printf("  Then invoke: /%s:%s\n\n", pluginName, presetDir)
	}

	return nil
}

func executeStaticAntigravity(config *dto.WorkflowConfig, guides []service.TemplateGuide) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	cmd := command.NewDeliverStaticWorkflowsCommand(fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Delivering Antigravity workflows (static)\n")
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Install: %s\n", config.Install)
	}
	fmt.Printf("  Workflows: %d\n", len(guides))
	fmt.Println()

	result, err := cmd.Execute(config, guides)
	if err != nil {
		return fmt.Errorf("static workflow delivery failed: %w", err)
	}

	fmt.Printf("Antigravity workflows delivered successfully!\n")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Installed: %s scope\n", config.Install)
	}
	fmt.Println()
	fmt.Println("Delivered workflows:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

func executePersonalizedWorkflows(ctx context.Context, p workflowsParams, config *dto.WorkflowConfig, guides []service.TemplateGuide, explicit map[string]bool, interactive bool) error {
	var err error

	// Resolve project context
	projectContext := p.projectContext
	if !explicit["context"] && interactive {
		projectContext, err = promptInput("Describe your project (language, tools, CI/CD, deployment)", "")
		if err != nil {
			return err
		}
	}
	if projectContext == "" {
		return fmt.Errorf("personalized mode requires project context; use --context flag")
	}
	config.ProjectContext = projectContext

	// Resolve model
	model := p.model
	if !explicit["model"] && interactive {
		model, err = promptModel()
		if err != nil {
			return err
		}
	}
	config.Model = model

	// Resolve API key
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// Initialize LLM provider
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	if config.Target == "claude" {
		return executePersonalizedPlugin(ctx, config, guides, provider, fileWriter, dirManager, model)
	}
	return executePersonalizedAntigravity(ctx, config, guides, provider, fileWriter, dirManager, model)
}

func executePersonalizedPlugin(ctx context.Context, config *dto.WorkflowConfig, guides []service.TemplateGuide, provider service.LLMProvider, fileWriter service.FileWriter, dirManager service.DirectoryManager, model string) error {
	pluginCmd := command.NewGeneratePluginCommand(provider, fileWriter, dirManager, root.TemplatesFS)

	fmt.Println()
	fmt.Printf("Generating Claude Code workflow plugins (personalized)\n")
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Install: %s\n", config.Install)
	}
	fmt.Printf("  Plugins: %d\n", len(guides))
	fmt.Println()
	fmt.Println("Generating personalized plugin skills via LLM API...")

	result, err := pluginCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("personalized plugin generation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("Claude Code plugins generated successfully!\n")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Printf("  Model: %s\n", result.Model)
	fmt.Printf("  Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut)
	if config.Install != "" {
		fmt.Printf("  Installed: %s scope\n", config.Install)
	}
	fmt.Println()
	fmt.Println("Generated plugin files:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}
	fmt.Println()
	fmt.Println("To use a plugin:")
	for _, guide := range guides {
		pluginName := catalog.PluginName(guide.Name)
		presetDir := catalog.PresetDirName(guide.Name)
		fmt.Printf("  claude --plugin-dir ./%s\n", pluginName)
		fmt.Printf("  Then invoke: /%s:%s\n\n", pluginName, presetDir)
	}

	return nil
}

func executePersonalizedAntigravity(ctx context.Context, config *dto.WorkflowConfig, guides []service.TemplateGuide, provider service.LLMProvider, fileWriter service.FileWriter, dirManager service.DirectoryManager, model string) error {
	workflowsCmd := command.NewGenerateWorkflowsCommand(provider, fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Generating Antigravity workflows (personalized)\n")
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	if config.Install != "" {
		fmt.Printf("  Install: %s\n", config.Install)
	}
	fmt.Printf("  Workflows: %d\n", len(guides))
	fmt.Println()
	fmt.Println("Generating personalized workflows via LLM API...")

	result, err := workflowsCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("personalized workflow generation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("Antigravity workflows generated successfully!\n")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Printf("  Model: %s\n", result.Model)
	fmt.Printf("  Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut)
	if config.Install != "" {
		fmt.Printf("  Installed: %s scope\n", config.Install)
	}
	fmt.Println()
	fmt.Println("Generated workflows:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

// --- Workflow path resolution ---

// defaultWorkflowsPath returns the default project-local path based on target.
func defaultWorkflowsPath(target string) string {
	if target == "claude" {
		return "." // plugins go in project root as self-contained directories
	}
	return filepath.Join(".agent", "workflows")
}

// globalWorkflowsPath returns the global workflows path based on target.
func globalWorkflowsPath(target string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	if target == "claude" {
		return filepath.Join(home, ".claude", "plugins")
	}
	return filepath.Join(home, ".gemini", "antigravity", "global_workflows")
}

// resolveWorkflowInstallPath resolves the output path based on the install scope and target.
func resolveWorkflowInstallPath(install, target string) string {
	switch install {
	case dto.InstallScopeGlobal:
		return globalWorkflowsPath(target)
	case dto.InstallScopeProject:
		return defaultWorkflowsPath(target)
	default:
		return defaultWorkflowsPath(target)
	}
}

// workflowTargetLabel returns a display label for the target ecosystem.
func workflowTargetLabel(target string) string {
	if target == "claude" {
		return "Claude Code"
	}
	return "Antigravity"
}
