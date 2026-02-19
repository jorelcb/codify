package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jorelcb/ai-context-generator/internal/application/command"
	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/filesystem"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/ai-context-generator/internal/infrastructure/template"
)

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	var (
		projectName  string
		description  string
		language     string
		projectType  string
		architecture string
		model        string
		interactive  bool
	)

	cmd := &cobra.Command{
		Use:   "generate [project-name]",
		Short: "Generate AI-optimized context files for a project",
		Long: `Generate context files using AI models:
  - PROMPT.md - Role and mission for the development agent
  - CONTEXT.md - Architecture, patterns, domain
  - SCAFFOLDING.md - Recommended project structure
  - INTERACTIONS_LOG.md - Initial development log

Requires ANTHROPIC_API_KEY environment variable.

Examples:
  # Interactive mode (coming soon)
  ai-context-generator generate -i

  # Generate with description
  ai-context-generator generate my-api \
    --description "API REST de gestion de inventarios en Go con Clean Architecture y PostgreSQL"

  # With optional hints
  ai-context-generator generate my-api \
    --description "API REST de gestion de inventarios" \
    --language go --type api --architecture ddd`,
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
			if description == "" {
				return fmt.Errorf("description is required (use -d to describe your project)")
			}

			return runGenerate(projectName, description, language, projectType, architecture, model)
		},
	}

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Project description (required)")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Programming language hint (optional)")
	cmd.Flags().StringVarP(&projectType, "type", "t", "", "Project type hint (optional)")
	cmd.Flags().StringVarP(&architecture, "architecture", "a", "", "Architecture pattern hint (optional)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Claude model to use (default: claude-sonnet-4-6)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")

	return cmd
}

func runInteractiveGenerate() error {
	fmt.Println("Interactive mode - Coming soon")
	fmt.Println("Will use survey/bubbletea for interactive UI")
	return nil
}

func runGenerate(projectName, description, language, projectType, architecture, model string) error {
	ctx := context.Background()

	// 1. Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	// 2. Load templates
	templateLoader := infratemplate.NewFileSystemTemplateLoader("templates/base")
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// 3. Initialize LLM provider (os.Stdout for streaming progress)
	provider := llm.NewAnthropicProvider(apiKey, model, os.Stdout)

	// 4. Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// 5. Create command
	generateCmd := command.NewGenerateContextCommand(provider, fileWriter, dirManager)

	// 6. Build config
	outputPath := filepath.Join("output", projectName)
	config := &dto.ProjectConfig{
		Name:         projectName,
		Description:  description,
		Language:     language,
		Type:         projectType,
		Architecture: architecture,
		Model:        model,
		OutputPath:   outputPath,
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
	usedModel := model
	if usedModel == "" {
		usedModel = "claude-sonnet-4-6"
	}
	fmt.Printf("  Model: %s\n", usedModel)
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
