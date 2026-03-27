package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
)

// DeliverStaticWorkflowsCommand delivers static workflows from the embedded catalog.
type DeliverStaticWorkflowsCommand struct {
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewDeliverStaticWorkflowsCommand creates a new static workflow delivery command.
func NewDeliverStaticWorkflowsCommand(
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *DeliverStaticWorkflowsCommand {
	return &DeliverStaticWorkflowsCommand{
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute delivers static workflows: reads embedded templates, adds frontmatter, writes flat .md files.
func (c *DeliverStaticWorkflowsCommand) Execute(
	config *dto.WorkflowConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	if err := c.directoryManager.CreateDir(config.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}

	var generatedFiles []string

	target := config.Target
	if target == "" {
		target = "antigravity"
	}

	for _, guide := range templateGuides {
		workflowDirName := strings.ReplaceAll(guide.Name, "_", "-")
		frontmatter := catalog.GenerateWorkflowFrontmatter(guide.Name, target)
		content := frontmatter + "\n" + guide.Content

		var filePath string
		if target == "claude" {
			// Claude: subdirectory per workflow with SKILL.md (mirrors skills pattern)
			workflowDir := fmt.Sprintf("%s/%s", config.OutputPath, workflowDirName)
			if err := c.directoryManager.CreateDir(workflowDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create workflow directory %s: %w", workflowDir, err)
			}
			filePath = fmt.Sprintf("%s/SKILL.md", workflowDir)
		} else {
			// Antigravity: flat .md files
			filePath = fmt.Sprintf("%s/%s.md", config.OutputPath, workflowDirName)
		}

		if err := c.fileWriter.WriteFile(filePath, []byte(content), os.FileMode(0644)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filePath, err)
		}
		generatedFiles = append(generatedFiles, filePath)
	}

	return &dto.GenerationResult{
		OutputPath:     config.OutputPath,
		GeneratedFiles: generatedFiles,
		Model:          "static",
	}, nil
}
