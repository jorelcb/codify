package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/jorelcb/ai-context-generator/internal/application/command"
	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/filesystem"
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
		Short: "Generate AI-optimized context files for a new project",
		Long: `Generate context files using AI models:
  - PROMPT.md - Role and mission for the development agent
  - CONTEXT.md - Architecture, patterns, domain
  - SCAFFOLDING.md - Recommended project structure
  - INTERACTIONS_LOG.md - Initial development log

Examples:
  # Interactive mode (recommended)
  ai-context-generator generate -i

  # Direct mode with flags
  ai-context-generator generate my-api --language go --type api --architecture ddd`,
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
			if language == "" {
				return fmt.Errorf("language is required (use -i for interactive mode)")
			}
			if projectType == "" {
				return fmt.Errorf("project type is required (use -i for interactive mode)")
			}

			return runGenerate(projectName, language, projectType, architecture)
		},
	}

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Programming language (go, javascript, python, etc.)")
	cmd.Flags().StringVarP(&projectType, "type", "t", "", "Project type (api, cli, library, etc.)")
	cmd.Flags().StringVarP(&architecture, "architecture", "a", "clean", "Architecture pattern (ddd, clean, hexagonal, etc.)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")

	return cmd
}

func runInteractiveGenerate() error {
	fmt.Println("Interactive mode - Coming soon")
	fmt.Println("Will use survey/bubbletea for interactive UI")
	return nil
}

func runGenerate(projectName, language, projectType, architecture string) error {
	ctx := context.Background()

	// Initialize repositories
	projectRepo := memory.NewProjectRepository()

	// Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// Initialize domain service
	projectGen := service.NewProjectGenerator(projectRepo, fileWriter, dirManager)

	// Initialize command
	generateCmd := command.NewGenerateProjectCommand(projectRepo, projectGen)

	// Build config
	config := &dto.ProjectConfig{
		Name:         projectName,
		Language:     language,
		Type:         projectType,
		Architecture: architecture,
		OutputPath:   filepath.Join("output", projectName),
		Capabilities: []string{},
		Metadata:     make(map[string]string),
	}

	fmt.Printf("Generating project: %s\n", projectName)
	fmt.Printf("  Language: %s\n", language)
	fmt.Printf("  Type: %s\n", projectType)
	fmt.Printf("  Architecture: %s\n", architecture)
	fmt.Println()

	projectInfo, err := generateCmd.Execute(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Println("Project entity created successfully!")
	fmt.Printf("  ID: %s\n", projectInfo.ID)
	fmt.Printf("  Output: %s\n", projectInfo.OutputPath)
	fmt.Printf("  Created: %s\n", projectInfo.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("Note: LLM-powered context generation coming soon.")

	return nil
}