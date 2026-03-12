package llm

import (
	"strings"
	"testing"

	"github.com/jorelcb/ai-context-generator/internal/domain/service"
)

func TestPromptBuilder_BuildSystemPromptForFile(t *testing.T) {
	builder := NewPromptBuilder()

	tests := []struct {
		guideName    string
		wantFileName string
	}{
		{"agents", "AGENTS.md"},
		{"context", "CONTEXT.md"},
		{"interactions", "INTERACTIONS_LOG.md"},
	}

	for _, tt := range tests {
		t.Run(tt.guideName, func(t *testing.T) {
			prompt := builder.BuildSystemPromptForFile(tt.guideName, "en")

			if prompt == "" {
				t.Error("BuildSystemPromptForFile() returned empty string")
			}
			if !strings.Contains(prompt, tt.wantFileName) {
				t.Errorf("BuildSystemPromptForFile() should mention %s", tt.wantFileName)
			}
			// Verify XML tag structure
			if !strings.Contains(prompt, "<role>") {
				t.Error("BuildSystemPromptForFile() should contain <role> XML tag")
			}
			if !strings.Contains(prompt, "<workflow>") {
				t.Error("BuildSystemPromptForFile() should contain <workflow> XML tag")
			}
			if !strings.Contains(prompt, "<output_quality>") {
				t.Error("BuildSystemPromptForFile() should contain <output_quality> XML tag")
			}
		})
	}
}

func TestPromptBuilder_BuildUserMessageForFile(t *testing.T) {
	builder := NewPromptBuilder()

	req := service.GenerationRequest{
		ProjectDescription: "API REST de gestion de inventarios en Go",
		Language:           "go",
		ProjectType:        "api",
		Architecture:       "clean",
	}

	guide := service.TemplateGuide{
		Name:    "agents",
		Content: "# Agents template content here",
	}

	msg := builder.BuildUserMessageForFile(req, guide)

	if msg == "" {
		t.Error("BuildUserMessageForFile() returned empty string")
	}

	// Verify description is included in XML tags
	if !strings.Contains(msg, "API REST de gestion de inventarios en Go") {
		t.Error("BuildUserMessageForFile() missing project description")
	}
	if !strings.Contains(msg, "<project_description>") {
		t.Error("BuildUserMessageForFile() should use <project_description> XML tag")
	}

	// Verify optional fields in metadata
	if !strings.Contains(msg, "go") {
		t.Error("BuildUserMessageForFile() missing language")
	}
	if !strings.Contains(msg, "<project_metadata>") {
		t.Error("BuildUserMessageForFile() should use <project_metadata> XML tag")
	}

	// Verify template content in XML tag
	if !strings.Contains(msg, "# Agents template content here") {
		t.Error("BuildUserMessageForFile() missing template content")
	}
	if !strings.Contains(msg, "<template_guide") {
		t.Error("BuildUserMessageForFile() should use <template_guide> XML tag")
	}
}

func TestPromptBuilder_BuildUserMessageForFile_WithoutOptionalFields(t *testing.T) {
	builder := NewPromptBuilder()

	req := service.GenerationRequest{
		ProjectDescription: "A simple project",
	}

	guide := service.TemplateGuide{
		Name:    "context",
		Content: "# Template",
	}

	msg := builder.BuildUserMessageForFile(req, guide)

	if strings.Contains(msg, "<project_metadata>") {
		t.Error("should not include project_metadata when all fields are empty")
	}
}

func TestFileOutputName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"agents", "AGENTS.md"},
		{"context", "CONTEXT.md"},
		{"interactions", "INTERACTIONS_LOG.md"},
		{"development_guide", "DEVELOPMENT_GUIDE.md"},
		{"idioms", "IDIOMS.md"},
		{"constitution", "CONSTITUTION.md"},
		{"spec", "SPEC.md"},
		{"plan", "PLAN.md"},
		{"tasks", "TASKS.md"},
		// Skills output files
		{"ddd_entity", "SKILL.md"},
		{"clean_arch_layer", "SKILL.md"},
		{"bdd_scenario", "SKILL.md"},
		{"cqrs_command", "SKILL.md"},
		{"hexagonal_port", "SKILL.md"},
		{"code_review", "SKILL.md"},
		{"test_strategy", "SKILL.md"},
		{"refactor_safely", "SKILL.md"},
		{"api_design", "SKILL.md"},
		{"unknown", "unknown.md"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FileOutputName(tt.input)
			if got != tt.want {
				t.Errorf("FileOutputName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPromptBuilder_BuildSkillsSystemPrompt(t *testing.T) {
	builder := NewPromptBuilder()

	tests := []struct {
		skillName string
		target    string
		wantTag   string
	}{
		{"ddd_entity", "claude", "Claude Code"},
		{"code_review", "codex", "Codex CLI"},
		{"bdd_scenario", "antigravity", "Antigravity IDE"},
	}

	for _, tt := range tests {
		t.Run(tt.skillName+"_"+tt.target, func(t *testing.T) {
			prompt := builder.BuildSkillsSystemPrompt(tt.skillName, tt.target, "en")

			if prompt == "" {
				t.Error("BuildSkillsSystemPrompt() returned empty string")
			}
			if !strings.Contains(prompt, tt.skillName) {
				t.Errorf("BuildSkillsSystemPrompt() should mention skill name %s", tt.skillName)
			}
			if !strings.Contains(prompt, tt.wantTag) {
				t.Errorf("BuildSkillsSystemPrompt() should mention target %s", tt.wantTag)
			}
			if !strings.Contains(prompt, "<role>") {
				t.Error("BuildSkillsSystemPrompt() should contain <role> XML tag")
			}
			if !strings.Contains(prompt, "<skill_format>") {
				t.Error("BuildSkillsSystemPrompt() should contain <skill_format> XML tag")
			}
			if !strings.Contains(prompt, "SKILL.md") {
				t.Error("BuildSkillsSystemPrompt() should mention SKILL.md")
			}
		})
	}
}

func TestPromptBuilder_BuildSkillsUserMessage(t *testing.T) {
	builder := NewPromptBuilder()

	guide := service.TemplateGuide{
		Name:    "ddd_entity",
		Content: "# DDD Entity Creation Skill\n\n## Purpose\nGuide an AI agent...",
	}

	msg := builder.BuildSkillsUserMessage(guide, "claude")

	if msg == "" {
		t.Error("BuildSkillsUserMessage() returned empty string")
	}
	if !strings.Contains(msg, "<skill_name>ddd_entity</skill_name>") {
		t.Error("BuildSkillsUserMessage() should contain skill_name XML tag")
	}
	if !strings.Contains(msg, "<target_ecosystem>claude</target_ecosystem>") {
		t.Error("BuildSkillsUserMessage() should contain target_ecosystem XML tag")
	}
	if !strings.Contains(msg, "<template_guide>") {
		t.Error("BuildSkillsUserMessage() should contain template_guide XML tag")
	}
	if !strings.Contains(msg, "DDD Entity Creation Skill") {
		t.Error("BuildSkillsUserMessage() should include guide content")
	}
}

func TestPromptBuilder_BuildSpecSystemPrompt(t *testing.T) {
	builder := NewPromptBuilder()

	existingContext := "# My Project Context\n\nThis is the architecture description."
	prompt := builder.BuildSpecSystemPrompt(existingContext, "en")

	if prompt == "" {
		t.Error("BuildSpecSystemPrompt() returned empty string")
	}
	if !strings.Contains(prompt, existingContext) {
		t.Error("BuildSpecSystemPrompt() should embed existing context")
	}
	if !strings.Contains(prompt, "<existing_context>") {
		t.Error("BuildSpecSystemPrompt() should use <existing_context> XML tag")
	}
	if !strings.Contains(prompt, "<role>") {
		t.Error("BuildSpecSystemPrompt() should contain <role> XML tag")
	}
	if !strings.Contains(prompt, "SDD") {
		t.Error("BuildSpecSystemPrompt() should mention SDD")
	}
}
