package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/jorelcb/ai-context-generator/internal/application/command"
	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/persistence/memory"
)

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	var (
		projectName  string
		language     string
		projectType  string
		architecture string
		interactive  bool
	)

	cmd := &cobra.Command{
		Use:   "generate [project-name]",
		Short: "Generate a new project with AI context documentation",
		Long: `Generate a new project including:
  - AI context documentation (PROMPT.md, CONTEXT.md, etc.)
  - Project scaffolding based on language and architecture
  - Taskfile for automation
  - Basic project structure

Examples:
  # Interactive mode (recommended)
  ai-context-generator generate -i

  # Direct mode with flags
  ai-context-generator generate my-api --language go --type api --architecture ddd

  # Generate with positional argument
  ai-context-generator generate my-service -l go -t microservice`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get project name from args or flag
			if len(args) > 0 {
				projectName = args[0]
			}

			// Interactive mode
			if interactive {
				return runInteractiveGenerate()
			}

			// Validate required flags in non-interactive mode
			if projectName == "" {
				return fmt.Errorf("project name is required (use -i for interactive mode)")
			}
			if language == "" {
				return fmt.Errorf("language is required (use -i for interactive mode)")
			}
			if projectType == "" {
				return fmt.Errorf("project type is required (use -i for interactive mode)")
			}

			// Run generation
			return runGenerate(projectName, language, projectType, architecture)
		},
	}

	// Flags
	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (go, javascript, python, etc.)")
	cmd.Flags().StringVarP(&projectType, "type", "t", "", "Project type (api, cli, library, etc.)")
	cmd.Flags().StringVarP(&architecture, "architecture", "a", "clean", "Architecture pattern (ddd, clean, hexagonal, etc.)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")

	return cmd
}

func runInteractiveGenerate() error {
	// TODO: Implement interactive wizard
	fmt.Println("Interactive mode - Coming soon in Phase 1")
	fmt.Println("Will use survey/bubbletea for interactive UI")
	return nil
}

func runGenerate(projectName, language, projectType, architecture string) error {
	ctx := context.Background()

	// Initialize repositories (in-memory for now)
	projectRepo := memory.NewProjectRepository()
	templateRepo := memory.NewTemplateRepository()

	// Initialize domain services
	projectGen := service.NewProjectGenerator(projectRepo)
	templateProc := service.NewTemplateProcessor(templateRepo)

	// Initialize command
	generateCmd := command.NewGenerateProjectCommand(
		projectRepo,
		templateRepo,
		projectGen,
		templateProc,
	)

	// Build config
	config := &dto.ProjectConfig{
		Name:         projectName,
		Language:     language,
		Type:         projectType,
		Architecture: architecture,
		OutputPath:   filepath.Join("output", projectName),
		Capabilities: []string{}, // TODO: Get from flags/interactive
		Metadata:     make(map[string]string),
	}

	// Execute command
	fmt.Printf("Generating project: %s\n", projectName)
	fmt.Printf("  Language: %s\n", language)
	fmt.Printf("  Type: %s\n", projectType)
	fmt.Printf("  Architecture: %s\n", architecture)
	fmt.Println()

	projectInfo, err := generateCmd.Execute(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Success output
	fmt.Println("✅ Project generated successfully!")
	fmt.Printf("  ID: %s\n", projectInfo.ID)
	fmt.Printf("  Output: %s\n", projectInfo.OutputPath)
	fmt.Printf("  Created: %s\n", projectInfo.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. cd", projectInfo.OutputPath)
	fmt.Println("  2. Review the generated files")
	fmt.Println("  3. Customize PROMPT.md and CONTEXT.md for your AI agent")

	return nil
}