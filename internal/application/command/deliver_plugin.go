package command

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
)

// DeliverPluginCommand delivers static Claude Code plugins from the embedded catalog.
// Each workflow preset is packaged as a complete plugin with skills, hooks, agents, and scripts.
type DeliverPluginCommand struct {
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
	scriptsFS        embed.FS
}

// NewDeliverPluginCommand creates a new static plugin delivery command.
func NewDeliverPluginCommand(
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
	scriptsFS embed.FS,
) *DeliverPluginCommand {
	return &DeliverPluginCommand{
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
		scriptsFS:        scriptsFS,
	}
}

// Execute generates a complete Claude Code plugin for each workflow template guide.
func (c *DeliverPluginCommand) Execute(
	config *dto.WorkflowConfig,
	templateGuides []service.TemplateGuide,
) (*dto.GenerationResult, error) {
	if err := c.directoryManager.CreateDir(config.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}

	var generatedFiles []string

	locale := config.Locale
	if locale == "" {
		locale = "en"
	}

	for _, guide := range templateGuides {
		pluginDir := filepath.Join(config.OutputPath, catalog.PluginName(guide.Name))
		annotations := catalog.ParseAnnotations(guide.Content)
		meta, ok := catalog.WorkflowMetadata[guide.Name]
		if !ok {
			meta = catalog.WorkflowMeta{Description: fmt.Sprintf("Workflow for %s", catalog.PresetDirName(guide.Name))}
		}

		// Crear estructura de directorios del plugin
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

		// plugin.json
		manifestPath := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
		manifestContent := catalog.GeneratePluginManifest(guide.Name, meta.Description)
		if err := c.fileWriter.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", manifestPath, err)
		}
		generatedFiles = append(generatedFiles, manifestPath)

		// skills/{preset}/SKILL.md
		skillPath := filepath.Join(pluginDir, "skills", catalog.PresetDirName(guide.Name), "SKILL.md")
		skillContent := catalog.TransformToPluginSkill(guide.Name, guide.Content)
		if err := c.fileWriter.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", skillPath, err)
		}
		generatedFiles = append(generatedFiles, skillPath)

		// hooks/hooks.json
		hooksPath := filepath.Join(pluginDir, "hooks", "hooks.json")
		hooksContent := catalog.GeneratePluginHooks(annotations)
		if err := c.fileWriter.WriteFile(hooksPath, []byte(hooksContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", hooksPath, err)
		}
		generatedFiles = append(generatedFiles, hooksPath)

		// agents/workflow-runner.md
		agentPath := filepath.Join(pluginDir, "agents", "workflow-runner.md")
		agentContent := catalog.GenerateWorkflowAgent(guide.Name, locale)
		if err := c.fileWriter.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", agentPath, err)
		}
		generatedFiles = append(generatedFiles, agentPath)

		// scripts/capture-output.sh (solo si hay anotaciones capture)
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
		Model:          "static",
	}, nil
}

// PluginOutputSummary returns a formatted summary of the generated plugin structure.
func PluginOutputSummary(pluginDir string, files []string) string {
	summary := fmt.Sprintf("Plugin generated: %s/\n", filepath.Base(pluginDir))
	for _, f := range files {
		rel, err := filepath.Rel(filepath.Dir(pluginDir), f)
		if err != nil {
			rel = f
		}
		summary += fmt.Sprintf("  %s\n", rel)
	}
	pluginName := filepath.Base(pluginDir)
	summary += fmt.Sprintf("\nTo use: claude --plugin-dir ./%s\n", pluginName)
	return summary
}

// PluginDirPermissions ensures script files in the plugin have executable permissions.
func PluginDirPermissions(pluginDir string) error {
	scriptsDir := filepath.Join(pluginDir, "scripts")
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(scriptsDir, entry.Name())
			if err := os.Chmod(path, 0755); err != nil {
				return fmt.Errorf("failed to chmod %s: %w", path, err)
			}
		}
	}
	return nil
}