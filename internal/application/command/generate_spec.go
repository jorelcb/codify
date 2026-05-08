package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
)

// GenerateSpecCommand orchestrates LLM-based spec file generation from existing context.
type GenerateSpecCommand struct {
	llmProvider      service.LLMProvider
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewGenerateSpecCommand creates a new GenerateSpecCommand.
func NewGenerateSpecCommand(
	llmProvider service.LLMProvider,
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *GenerateSpecCommand {
	return &GenerateSpecCommand{
		llmProvider:      llmProvider,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute runs the spec generation pipeline:
// 1. Build generation request with existing context and spec mode
// 2. Call LLM provider
// 3. Create specs output directory
// 4. Write generated spec files to disk
func (c *GenerateSpecCommand) Execute(
	ctx context.Context,
	config *dto.SpecConfig,
	existingContext string,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	// 1. Build generation request in spec mode. SDDStandardHints carries
	//    the active standard's prompt addendum so the LLM respects per-standard
	//    conventions (file naming, layout, etc.).
	req := service.GenerationRequest{
		ProjectDescription: existingContext,
		TemplateGuides:     templateGuides,
		ExistingContext:    existingContext,
		Mode:               "spec",
		Locale:             config.Locale,
		SDDStandardHints:   config.StandardHints,
	}

	// 2. Call LLM provider
	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM spec generation failed: %w", err)
	}

	// 3. Create specs output directory honoring the active SpecStandard's layout.
	//    LayoutFlat            → <output>/specs/
	//    LayoutFeatureGrouped  → <output>/specs/<FeatureID>/
	specsDir := filepath.Join(config.OutputPath, "specs")
	if config.Layout == service.LayoutFeatureGrouped && config.FeatureID != "" {
		specsDir = filepath.Join(specsDir, config.FeatureID)
	}
	if err := c.directoryManager.CreateDir(specsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create specs directory: %w", err)
	}

	// 4. Write each generated spec file
	var generatedFiles []string
	for _, file := range response.Files {
		filePath := filepath.Join(specsDir, file.Name)
		if err := c.fileWriter.WriteFile(filePath, []byte(file.Content), os.FileMode(0644)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", file.Name, err)
		}
		generatedFiles = append(generatedFiles, filePath)
	}

	return &dto.GenerationResult{
		OutputPath:     specsDir,
		GeneratedFiles: generatedFiles,
		Model:          response.Model,
		TokensIn:       response.TokensIn,
		TokensOut:      response.TokensOut,
	}, nil
}
