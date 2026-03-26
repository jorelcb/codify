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
		Short: "Generate Antigravity workflow files (.md)",
		Long: `Generate Antigravity workflow files — multi-step recipes that AI agents
execute on demand via /command.

Workflows are markdown files with numbered steps and execution annotations
(// turbo, // parallel, // capture, etc.) that teach agents how to perform
complex, repeatable development tasks.

Target: Antigravity IDE (Google) exclusively.

Modes:
  static        - Instant delivery from built-in catalog (no API key needed)
  personalized  - LLM-adapted workflows tailored to your project (requires API key)

Presets:
  feature-development  - Full feature lifecycle (branch → implement → test → PR)
  bug-fix              - Structured bug fix (reproduce → diagnose → fix → test)
  release-cycle        - Release process (version → changelog → tag → deploy)
  all                  - All workflow presets

Install:
  global   - Install to ~/.gemini/antigravity/global_workflows/
  project  - Install to .agent/workflows/

When run without flags, an interactive menu is displayed.

Examples:
  # Interactive mode (guided selection)
  codify workflows

  # Static: instant delivery, no API key
  codify workflows --preset all --mode static

  # Install globally
  codify workflows --preset all --mode static --install global

  # Install to current project
  codify workflows --preset feature-development --mode static --install project

  # Personalized: LLM-adapted to your project
  codify workflows --preset all --mode personalized --context "Go microservice with CI/CD via GitHub Actions"`,
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

	// 2. Resolve mode
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

	// 3. Resolve locale
	locale := p.locale
	if !explicit["locale"] && interactive {
		locale, err = promptLocale()
		if err != nil {
			return err
		}
	}

	// 4. Resolve install scope and output
	install := p.install
	output := p.output

	if !explicit["install"] && !explicit["output"] && interactive {
		globalPath := globalWorkflowsPath()
		projectPath := defaultWorkflowsPath()

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
			output, err = promptInput("Output directory", defaultWorkflowsPath())
			if err != nil {
				return err
			}
		}
	} else if explicit["install"] {
		output = resolveWorkflowInstallPath(install)
	} else if output == "" {
		output = defaultWorkflowsPath()
	}

	// 5. Resolve templates from catalog
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

	// 6. Build config
	config := &dto.WorkflowConfig{
		Category:   "workflows",
		Preset:     preset,
		Mode:       mode,
		Locale:     locale,
		OutputPath: output,
		Install:    install,
	}

	// 7. Execute based on mode
	if mode == dto.SkillModeStatic {
		return executeStaticWorkflows(config, guides)
	}
	return executePersonalizedWorkflows(ctx, p, config, guides, explicit, interactive)
}

func executeStaticWorkflows(config *dto.WorkflowConfig, guides []service.TemplateGuide) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	cmd := command.NewDeliverStaticWorkflowsCommand(fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Delivering Antigravity workflows (static)\n")
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

	fmt.Println("Antigravity workflows delivered successfully!")
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
	workflowsCmd := command.NewGenerateWorkflowsCommand(provider, fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Generating Antigravity workflows (personalized)\n")
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
	fmt.Println("Antigravity workflows generated successfully!")
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

// defaultWorkflowsPath returns the default project-local path for Antigravity workflows.
func defaultWorkflowsPath() string {
	return filepath.Join(".agent", "workflows")
}

// globalWorkflowsPath returns the global Antigravity workflows path.
func globalWorkflowsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, ".gemini", "antigravity", "global_workflows")
}

// resolveWorkflowInstallPath resolves the output path based on the install scope.
func resolveWorkflowInstallPath(install string) string {
	switch install {
	case dto.InstallScopeGlobal:
		return globalWorkflowsPath()
	case dto.InstallScopeProject:
		return defaultWorkflowsPath()
	default:
		return defaultWorkflowsPath()
	}
}
