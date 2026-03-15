package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

const defaultLocale = "en"

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	var (
		projectName  string
		description  string
		fromFile     string
		language     string
		projectType  string
		architecture string
		model        string
		preset       string
		locale       string
		output       string
		withSpecs    bool
		interactive  bool
	)

	cmd := &cobra.Command{
		Use:   "generate [project-name]",
		Short: "Generate AI-optimized context files for a project",
		Long: `Generate context files using AI models:
  - AGENTS.md - Root file: tech stack, commands, conventions, structure (at project root)
  - CONTEXT.md - Architecture and technical design (in context/)
  - INTERACTIONS_LOG.md - Session history and ADR log (in context/)
  - DEVELOPMENT_GUIDE.md - Work methodology, testing, security, delivery (in context/)
  - IDIOMS.md - Language-specific patterns and conventions (in context/, requires --language)

Presets:
  default  - DDD/Clean Architecture/BDD recommended templates (default)
  neutral  - Generic templates without architectural opinions

Locales:
  en  - English (default)
  es  - Spanish

Requires ANTHROPIC_API_KEY (for Claude) or GEMINI_API_KEY (for Gemini) environment variable.

Examples:
  # Generate with description (English, default preset)
  codify generate my-api \
    --description "Inventory management REST API in Go with Clean Architecture and PostgreSQL"

  # Generate in Spanish
  codify generate my-api \
    --description "API REST de gestion de inventarios en Go" \
    --locale es

  # With language-specific guides
  codify generate my-api \
    --description "Inventory management REST API in Go" \
    --language go

  # From a detailed description file
  codify generate my-api \
    --from-file ./docs/project-description.md \
    --language go

  # With neutral preset
  codify generate my-api \
    --description "Inventory management REST API" \
    --preset neutral

  # Generate to a specific directory
  codify generate my-api \
    --description "Inventory management REST API in Go" \
    --output ./docs/

  # Generate context + specs in one command
  codify generate my-api \
    --description "Inventory management REST API in Go" \
    --with-specs`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				projectName = args[0]
			}

			if interactive {
				return runInteractiveGenerate()
			}

			if projectName == "" {
				return fmt.Errorf("project name is required (use -i for interactive mode)")
			}
			if fromFile != "" && description != "" {
				return fmt.Errorf("--description and --from-file are mutually exclusive")
			}
			if fromFile != "" {
				content, err := os.ReadFile(fromFile)
				if err != nil {
					return fmt.Errorf("failed to read description file: %w", err)
				}
				description = string(content)
			}
			if description == "" {
				return fmt.Errorf("description is required (use -d or --from-file)")
			}

			if output == "" {
				output = "."
			}

			if err := runGenerate(projectName, description, language, projectType, architecture, model, preset, locale, output); err != nil {
				return err
			}

			if withSpecs {
				fmt.Println()
				fmt.Println("--- Generating specs from context ---")
				fmt.Println()
				return runSpec(projectName, output, output, model, locale)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Project description (required unless --from-file)")
	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "Read project description from file (alternative to --description)")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (activates idiomatic guides)")
	cmd.Flags().StringVarP(&projectType, "type", "t", "", "Project type hint (api, cli, lib...)")
	cmd.Flags().StringVarP(&architecture, "architecture", "a", "", "Architecture pattern hint")
	cmd.Flags().StringVarP(&model, "model", "m", "", "LLM model (default: claude-sonnet-4-6, or gemini-3.1-pro-preview)")
	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Template preset: default (DDD/Clean Architecture) or neutral")
	cmd.Flags().StringVar(&locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&withSpecs, "with-specs", false, "Also generate SDD spec files after context generation")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")

	return cmd
}

func runInteractiveGenerate() error {
	fmt.Println("Interactive mode - Coming soon")
	fmt.Println("Will use survey/bubbletea for interactive UI")
	return nil
}

// validPresets maps preset names to their directory name.
var validPresets = map[string]bool{
	"default":  true,
	"neutral":  true,
	"workflow": true,
}

// resolveTemplatePath builds the full template path: templates/{locale}/{preset}
func resolveTemplatePath(locale, preset string) string {
	if !validPresets[preset] {
		preset = "default"
	}
	return filepath.Join("templates", locale, preset)
}

// resolveLocaleBase returns the locale base directory: templates/{locale}
func resolveLocaleBase(locale string) string {
	return filepath.Join("templates", locale)
}

func runGenerate(projectName, description, language, projectType, architecture, model, preset, locale, output string) error {
	ctx := context.Background()

	// 1. Resolve API key for the selected provider
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// 2. Load templates (base preset + language-specific if --language is provided)
	templatePath := resolveTemplatePath(locale, preset)
	var templateLoader service.TemplateLoader
	if language != "" {
		templateLoader = infratemplate.NewFileSystemTemplateLoaderWithLanguage(root.TemplatesFS, templatePath, resolveLocaleBase(locale), language)
	} else {
		templateLoader = infratemplate.NewFileSystemTemplateLoader(root.TemplatesFS, templatePath)
	}
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// 3. Initialize LLM provider (os.Stdout for streaming progress)
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// 4. Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// 5. Create command
	generateCmd := command.NewGenerateContextCommand(provider, fileWriter, dirManager)

	// 6. Build config
	config := &dto.ProjectConfig{
		Name:         projectName,
		Description:  description,
		Language:     language,
		Type:         projectType,
		Architecture: architecture,
		Model:        model,
		OutputPath:   output,
		Locale:       locale,
	}

	// 7. Show progress
	fmt.Printf("Generating context for: %s\n", projectName)
	fmt.Printf("  Description: %s\n", truncateStr(description, 80))
	if language != "" {
		fmt.Printf("  Language: %s\n", language)
	}
	if projectType != "" {
		fmt.Printf("  Type: %s\n", projectType)
	}
	if architecture != "" {
		fmt.Printf("  Architecture: %s\n", architecture)
	}
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Preset: %s\n", preset)
	fmt.Printf("  Locale: %s\n", locale)
	fmt.Println()
	fmt.Println("Generating context files via Claude API...")

	// 8. Execute
	result, err := generateCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// 9. Show results
	fmt.Println()
	fmt.Println("Context files generated successfully!")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Printf("  Model: %s\n", result.Model)
	fmt.Printf("  Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut)
	fmt.Println()
	fmt.Println("Generated files:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}