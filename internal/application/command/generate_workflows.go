package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
)

// GenerateWorkflowsCommand orchestrates LLM-based workflow generation.
type GenerateWorkflowsCommand struct {
	llmProvider      service.LLMProvider
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewGenerateWorkflowsCommand creates a new personalized workflow generation command.
func NewGenerateWorkflowsCommand(
	llmProvider service.LLMProvider,
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *GenerateWorkflowsCommand {
	return &GenerateWorkflowsCommand{
		llmProvider:      llmProvider,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute runs the workflow generation pipeline:
// 1. Build generation request in workflows mode
// 2. Call LLM provider (one call per workflow)
// 3. Write flat .md files to output directory
func (c *GenerateWorkflowsCommand) Execute(
	ctx context.Context,
	config *dto.WorkflowConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	if err := c.directoryManager.CreateDir(config.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}

	target := config.Target
	if target == "" {
		target = "antigravity"
	}

	mode := "workflows"
	if target == "claude" {
		mode = "workflow-skills"
	}

	req := service.GenerationRequest{
		TemplateGuides: templateGuides,
		Mode:           mode,
		Target:         target,
		Locale:         config.Locale,
		ProjectContext: config.ProjectContext,
	}

	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM workflow generation failed: %w", err)
	}

	var generatedFiles []string
	for i, file := range response.Files {
		guideName := templateGuides[i].Name
		workflowDirName := strings.ReplaceAll(guideName, "_", "-")

		var filePath string
		if target == "claude" {
			// Claude: subdirectory per workflow with SKILL.md
			workflowDir := fmt.Sprintf("%s/%s", config.OutputPath, workflowDirName)
			if err := c.directoryManager.CreateDir(workflowDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create workflow directory %s: %w", workflowDir, err)
			}
			filePath = fmt.Sprintf("%s/SKILL.md", workflowDir)
		} else {
			// Antigravity: flat .md files
			filePath = fmt.Sprintf("%s/%s.md", config.OutputPath, workflowDirName)
		}

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
