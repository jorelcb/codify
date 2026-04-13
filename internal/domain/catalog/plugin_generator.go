package catalog

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const pluginVersion = "1.0.0"
const pluginAuthor = "codify"

// PluginName returns the Claude Code plugin directory name for a workflow preset.
func PluginName(presetName string) string {
	return "codify-wf-" + strings.ReplaceAll(presetName, "_", "-")
}

// PresetDirName returns the kebab-case directory name for a preset.
func PresetDirName(presetName string) string {
	return strings.ReplaceAll(presetName, "_", "-")
}

// --- plugin.json ---

type pluginManifest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Author      pluginAuthorInfo `json:"author"`
}

type pluginAuthorInfo struct {
	Name string `json:"name"`
}

// GeneratePluginManifest produces the plugin.json content for a workflow plugin.
func GeneratePluginManifest(presetName, description string) string {
	manifest := pluginManifest{
		Name:        PluginName(presetName),
		Description: description,
		Version:     pluginVersion,
		Author:      pluginAuthorInfo{Name: pluginAuthor},
	}
	data, _ := json.MarshalIndent(manifest, "", "  ")
	return string(data) + "\n"
}

// --- hooks.json ---

type hooksConfig struct {
	Hooks map[string][]hookEntry `json:"hooks"`
}

type hookEntry struct {
	Matcher string     `json:"matcher,omitempty"`
	If      string     `json:"if,omitempty"`
	Hooks   []hookSpec `json:"hooks"`
}

type hookSpec struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// GeneratePluginHooks produces the hooks/hooks.json content from parsed annotations.
func GeneratePluginHooks(annotations []AnnotationMeta) string {
	config := hooksConfig{
		Hooks: make(map[string][]hookEntry),
	}

	// turbo annotations → PreToolUse auto-approve
	if HasAnnotationType(annotations, "turbo") {
		config.Hooks["PreToolUse"] = append(config.Hooks["PreToolUse"], hookEntry{
			Matcher: "Bash",
			Hooks: []hookSpec{
				{
					Type:    "command",
					Command: `echo '{"permissionDecision": "allow"}'`,
				},
			},
		})
	}

	// capture annotations → PostToolUse with capture script
	if HasAnnotationType(annotations, "capture") {
		config.Hooks["PostToolUse"] = append(config.Hooks["PostToolUse"], hookEntry{
			Matcher: "Bash",
			Hooks: []hookSpec{
				{
					Type:    "command",
					Command: "${CLAUDE_PLUGIN_ROOT}/scripts/capture-output.sh",
				},
			},
		})
	}

	// if annotations → PreToolUse prompt hooks for conditional evaluation
	ifAnnotations := FilterByType(annotations, "if")
	for _, a := range ifAnnotations {
		config.Hooks["PreToolUse"] = append(config.Hooks["PreToolUse"], hookEntry{
			Matcher: "Bash",
			Hooks: []hookSpec{
				{
					Type:    "prompt",
					Prompt:  fmt.Sprintf("Evaluate whether the following condition is true for the current project context: \"%s\". If true, allow the action. If false, block it with a brief explanation.", a.Value),
					Timeout: 30,
				},
			},
		})
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data) + "\n"
}

// --- SKILL.md transformation ---

var annotationLineRegex = regexp.MustCompile(`^\s*//\s*(turbo|capture:|if |parallel|retry:|timeout:)`)

// TransformToPluginSkill converts an Antigravity workflow template to a Claude plugin SKILL.md.
// It strips annotation lines and generates YAML frontmatter appropriate for a plugin skill.
func TransformToPluginSkill(guideName, templateContent string) string {
	presetDir := PresetDirName(guideName)
	meta, ok := WorkflowMetadata[guideName]
	if !ok {
		meta = WorkflowMeta{Description: fmt.Sprintf("Workflow for %s", presetDir)}
	}

	// Build frontmatter
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", presetDir))
	sb.WriteString(fmt.Sprintf("description: %s\n", meta.Description))
	sb.WriteString("---\n\n")

	// Process content: strip annotation lines
	lines := strings.Split(templateContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if annotationLineRegex.MatchString(trimmed) {
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// --- Agent generation ---

// GenerateWorkflowAgent produces the content for a workflow-runner agent definition.
func GenerateWorkflowAgent(presetName, locale string) string {
	presetDir := PresetDirName(presetName)
	meta, ok := WorkflowMetadata[presetName]
	if !ok {
		meta = WorkflowMeta{Description: fmt.Sprintf("Workflow runner for %s", presetDir)}
	}

	description := fmt.Sprintf("Executes the %s workflow steps with tool access. %s", presetDir, meta.Description)

	var instructions string
	if locale == "es" {
		instructions = fmt.Sprintf(`Eres un agente de ejecucion de workflow especializado en el proceso de %s.

Tu rol:
- Ejecutar los pasos del workflow de forma ordenada y disciplinada
- Usar las herramientas disponibles (Bash, Read, Edit, Write, Grep, Glob) para completar cada paso
- Reportar el resultado de cada paso antes de avanzar al siguiente
- Detenerte y solicitar intervencion si un paso falla o requiere decision humana

Sigue las instrucciones del skill orquestador exactamente como estan descritas.`, presetDir)
	} else {
		instructions = fmt.Sprintf(`You are a workflow execution agent specialized in the %s process.

Your role:
- Execute workflow steps in order and with discipline
- Use available tools (Bash, Read, Edit, Write, Grep, Glob) to complete each step
- Report the outcome of each step before moving to the next
- Stop and request intervention if a step fails or requires human decision

Follow the orchestrator skill instructions exactly as described.`, presetDir)
	}

	return fmt.Sprintf(`---
name: workflow-runner
description: %s
model: sonnet
tools: Bash, Read, Edit, Write, Grep, Glob
maxTurns: 50
---

%s
`, description, instructions)
}