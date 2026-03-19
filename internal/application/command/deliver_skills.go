package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
)

// DeliverStaticSkillsCommand entrega skills estáticos desde el catálogo embebido sin LLM.
type DeliverStaticSkillsCommand struct {
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
}

// NewDeliverStaticSkillsCommand crea un nuevo comando de entrega estática.
func NewDeliverStaticSkillsCommand(
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
) *DeliverStaticSkillsCommand {
	return &DeliverStaticSkillsCommand{
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
	}
}

// Execute entrega skills estáticos: lee templates embebidos, agrega frontmatter, escribe SKILL.md.
func (c *DeliverStaticSkillsCommand) Execute(
	config *dto.SkillsConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	var generatedFiles []string

	for _, guide := range templateGuides {
		skillDirName := strings.ReplaceAll(guide.Name, "_", "-")
		skillDir := filepath.Join(config.OutputPath, skillDirName)

		if err := c.directoryManager.CreateDir(skillDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create skill directory %s: %w", skillDirName, err)
		}

		// Generar frontmatter según ecosistema + contenido del template
		frontmatter := catalog.GenerateFrontmatter(guide.Name, config.Target)
		content := frontmatter + "\n" + guide.Content

		filePath := filepath.Join(skillDir, "SKILL.md")
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
