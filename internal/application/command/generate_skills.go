package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
)

// GenerateSkillsCommand orchestrates LLM-based Agent Skills generation.
type GenerateSkillsCommand struct {
	llmProvider      service.LLMProvider
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewGenerateSkillsCommand creates a new GenerateSkillsCommand.
func NewGenerateSkillsCommand(
	llmProvider service.LLMProvider,
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *GenerateSkillsCommand {
	return &GenerateSkillsCommand{
		llmProvider:      llmProvider,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute runs the skills generation pipeline:
// 1. Build generation request in skills mode
// 2. Call LLM provider (one call per skill)
// 3. Create per-skill output directories
// 4. Write SKILL.md files to disk
func (c *GenerateSkillsCommand) Execute(
	ctx context.Context,
	config *dto.SkillsConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	// 1. Build generation request in skills mode (personalized with project context)
	req := service.GenerationRequest{
		TemplateGuides: templateGuides,
		Mode:           "skills",
		Target:         config.Target,
		Locale:         config.Locale,
		ProjectContext: config.ProjectContext,
	}

	// 2. Call LLM provider
	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM skills generation failed: %w", err)
	}

	// 3. Write each skill to its own directory
	var generatedFiles []string
	for i, file := range response.Files {
		// Map guide name to directory name (underscore → hyphen)
		guideName := templateGuides[i].Name
		skillDirName := strings.ReplaceAll(guideName, "_", "-")
		skillDir := filepath.Join(config.OutputPath, skillDirName)

		if err := c.directoryManager.CreateDir(skillDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create skill directory %s: %w", skillDirName, err)
		}

		filePath := filepath.Join(skillDir, "SKILL.md")
		if err := c.fileWriter.WriteFile(filePath, []byte(file.Content), os.FileMode(0644)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filePath, err)
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
