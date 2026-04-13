 package command

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
)

// GeneratePluginCommand orchestrates LLM-based Claude Code plugin generation.
// The LLM personalizes the SKILL.md; hooks, agents, and scripts remain static.
type GeneratePluginCommand struct {
	llmProvider      service.LLMProvider
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
	scriptsFS        embed.FS
}

// NewGeneratePluginCommand creates a new personalized plugin generation command.
func NewGeneratePluginCommand(
	llmProvider service.LLMProvider,
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
	scriptsFS embed.FS,
) *GeneratePluginCommand {
	return &GeneratePluginCommand{
		llmProvider:      llmProvider,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
		scriptsFS:        scriptsFS,
	}
}

// Execute generates Claude Code plugins with LLM-personalized SKILL.md files.
// Static components (plugin.json, hooks.json, agents, scripts) are generated from the catalog.
// The LLM is called once per workflow to generate a personalized SKILL.md.
func (c *GeneratePluginCommand) Execute(
	ctx context.Context,
	config *dto.WorkflowConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	if err := c.directoryManager.CreateDir(config.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}

	locale := config.Locale
	if locale == "" {
		locale = "en"
	}

	// LLM genera los SKILL.md personalizados (mode "plugin")
	req := service.GenerationRequest{
		TemplateGuides: templateGuides,
		Mode:           "plugin",
		Target:         "claude",
		Locale:         locale,
		ProjectContext: config.ProjectContext,
	}

	response, err := c.llmProvider.GenerateContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM plugin skill generation failed: %w", err)
	}

	var generatedFiles []string

	for i, guide := range templateGuides {
		pluginDir := filepath.Join(config.OutputPath, catalog.PluginName(guide.Name))
		annotations := catalog.ParseAnnotations(guide.Content)
		meta, ok := catalog.WorkflowMetadata[guide.Name]
		if !ok {
			meta = catalog.WorkflowMeta{Description: fmt.Sprintf("Workflow for %s", catalog.PresetDirName(guide.Name))}
		}

		// Crear estructura de directorios
		dirs := []string{
			filepath.Join(pluginDir, ".claude-plugin"),
			filepath.Join(pluginDir, "skills", catalog.PresetDirName(guide.Name)),
			filepath.Join(pluginDir, "agents"),
			filepath.Join(pluginDir, "hooks"),
		}
		if catalog.HasAnnotationType(annotations, "capture") {
			dirs = append(dirs, filepath.Join(pluginDir, "scripts"))
		}
		for _, dir := range dirs {
			if err := c.directoryManager.CreateDir(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// plugin.json (estático)
		manifestPath := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
		manifestContent := catalog.GeneratePluginManifest(guide.Name, meta.Description)
		if err := c.fileWriter.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", manifestPath, err)
		}
		generatedFiles = append(generatedFiles, manifestPath)

		// skills/{preset}/SKILL.md (personalizado por LLM)
		skillPath := filepath.Join(pluginDir, "skills", catalog.PresetDirName(guide.Name), "SKILL.md")
		skillContent := response.Files[i].Content
		if err := c.fileWriter.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", skillPath, err)
		}
		generatedFiles = append(generatedFiles, skillPath)

		// hooks/hooks.json (estático desde anotaciones)
		hooksPath := filepath.Join(pluginDir, "hooks", "hooks.json")
		hooksContent := catalog.GeneratePluginHooks(annotations)
		if err := c.fileWriter.WriteFile(hooksPath, []byte(hooksContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", hooksPath, err)
		}
		generatedFiles = append(generatedFiles, hooksPath)

		// agents/workflow-runner.md (estático)
		agentPath := filepath.Join(pluginDir, "agents", "workflow-runner.md")
		agentContent := catalog.GenerateWorkflowAgent(guide.Name, locale)
		if err := c.fileWriter.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", agentPath, err)
		}
		generatedFiles = append(generatedFiles, agentPath)

		// scripts/capture-output.sh (estático, solo si hay capture)
		if catalog.HasAnnotationType(annotations, "capture") {
			scriptData, err := c.scriptsFS.ReadFile("templates/scripts/capture-output.sh")
			if err != nil {
				return nil, fmt.Errorf("failed to read embedded capture-output.sh: %w", err)
			}
			scriptPath := filepath.Join(pluginDir, "scripts", "capture-output.sh")
			if err := c.fileWriter.WriteFile(scriptPath, scriptData, 0755); err != nil {
				return nil, fmt.Errorf("failed to write %s: %w", scriptPath, err)
			}
			generatedFiles = append(generatedFiles, scriptPath)
		}
	}

	return &dto.GenerationResult{
		OutputPath:     config.OutputPath,
		GeneratedFiles: generatedFiles,
		Model:          response.Model,
		TokensIn:       response.TokensIn,
		TokensOut:      response.TokensOut,
	}, nil
}