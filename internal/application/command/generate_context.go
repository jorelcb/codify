package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
)

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
// 3. Create output directory
// 4. Write generated files to disk
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
	}

	// 2. Call LLM provider
	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 3. Create output directory
	outputDir := filepath.Join(config.OutputPath, "context")
	if err := c.directoryManager.CreateDir(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// 4. Write each generated file
	var generatedFiles []string
	for _, file := range response.Files {
		filePath := filepath.Join(outputDir, file.Name)
		if err := c.fileWriter.WriteFile(filePath, []byte(file.Content), os.FileMode(0644)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", file.Name, err)
		}
		generatedFiles = append(generatedFiles, filePath)
	}

	return &dto.GenerationResult{
		OutputPath:     outputDir,
		GeneratedFiles: generatedFiles,
		Model:          response.Model,
		TokensIn:       response.TokensIn,
		TokensOut:      response.TokensOut,
	}, nil
}
