package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
)

// rootFiles are written to the project root, not the context/ subdirectory.
var rootFiles = map[string]bool{
	"AGENTS.md": true,
}

// GenerateContextCommand orchestrates LLM-based context file generation.
type GenerateContextCommand struct {
	llmProvider      service.LLMProvider
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewGenerateContextCommand creates a new GenerateContextCommand.
func NewGenerateContextCommand(
	llmProvider service.LLMProvider,
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *GenerateContextCommand {
	return &GenerateContextCommand{
		llmProvider:      llmProvider,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute runs the full context generation pipeline:
// 1. Build generation request from config + templates
// 2. Call LLM provider
// 3. Create output directories
// 4. Write generated files to disk (AGENTS.md at root, rest in context/)
func (c *GenerateContextCommand) Execute(
	ctx context.Context,
	config *dto.ProjectConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	// 1. Build generation request
	req := service.GenerationRequest{
		ProjectDescription: config.Description,
		TemplateGuides:     templateGuides,
		Language:           config.Language,
		ProjectType:        config.Type,
		Architecture:       config.Architecture,
		Locale:             config.Locale,
		Mode:               config.Mode,
	}

	// 2. Call LLM provider
	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 3. Create output directories
	contextDir := filepath.Join(config.OutputPath, "context")
	if err := c.directoryManager.CreateDir(config.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := c.directoryManager.CreateDir(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// 4. Write each generated file
	var generatedFiles []string
	for _, file := range response.Files {
		// AGENTS.md goes to project root, rest goes to context/
		var filePath string
		if rootFiles[file.Name] {
			filePath = filepath.Join(config.OutputPath, file.Name)
		} else {
			filePath = filepath.Join(contextDir, file.Name)
		}

		if err := c.fileWriter.WriteFile(filePath, []byte(file.Content), os.FileMode(0644)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", file.Name, err)
		}
		generatedFiles = append(generatedFiles, filePath)
	}

	return &dto.GenerationResult{
		OutputPath:     config.OutputPath,
		GeneratedFiles: generatedFiles,
		Model:          response.Model,
		TokensIn:       response.TokensIn,
		TokensOut:      response.TokensOut,
	}, nil
}
