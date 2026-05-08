package llm

import (
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
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

func TestPromptBuilder_BuildPersonalizedSkillsSystemPrompt(t *testing.T) {
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
			prompt := builder.BuildPersonalizedSkillsSystemPrompt(tt.skillName, tt.target, "en", "Go project with DDD architecture")

			if prompt == "" {
				t.Error("BuildPersonalizedSkillsSystemPrompt() returned empty string")
			}
			if !strings.Contains(prompt, tt.skillName) {
				t.Errorf("BuildPersonalizedSkillsSystemPrompt() should mention skill name %s", tt.skillName)
			}
			if !strings.Contains(prompt, tt.wantTag) {
				t.Errorf("BuildPersonalizedSkillsSystemPrompt() should mention target %s", tt.wantTag)
			}
			if !strings.Contains(prompt, "<role>") {
				t.Error("BuildPersonalizedSkillsSystemPrompt() should contain <role> XML tag")
			}
			if !strings.Contains(prompt, "<skill_format>") {
				t.Error("BuildPersonalizedSkillsSystemPrompt() should contain <skill_format> XML tag")
			}
			if !strings.Contains(prompt, "SKILL.md") {
				t.Error("BuildPersonalizedSkillsSystemPrompt() should mention SKILL.md")
			}
			if !strings.Contains(prompt, "<project_context>") {
				t.Error("BuildPersonalizedSkillsSystemPrompt() should contain project_context XML tag")
			}
			if !strings.Contains(prompt, "<personalization_rules>") {
				t.Error("BuildPersonalizedSkillsSystemPrompt() should contain personalization_rules XML tag")
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

func TestPromptBuilder_BuildAnalyzeSystemPromptForFile(t *testing.T) {
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
			prompt := builder.BuildAnalyzeSystemPromptForFile(tt.guideName, "en")

			if prompt == "" {
				t.Error("BuildAnalyzeSystemPromptForFile() returned empty string")
			}
			if !strings.Contains(prompt, tt.wantFileName) {
				t.Errorf("BuildAnalyzeSystemPromptForFile() should mention %s", tt.wantFileName)
			}

			// Verify analyze-specific XML tags
			if !strings.Contains(prompt, "<scan_trust>") {
				t.Error("BuildAnalyzeSystemPromptForFile() should contain <scan_trust> XML tag")
			}
			if !strings.Contains(prompt, "AUTO-SCANNED") {
				t.Error("BuildAnalyzeSystemPromptForFile() should mention AUTO-SCANNED")
			}
			if !strings.Contains(prompt, "FACTUAL") {
				t.Error("BuildAnalyzeSystemPromptForFile() should mention FACTUAL")
			}

			// Verify shared structure tags
			if !strings.Contains(prompt, "<role>") {
				t.Error("BuildAnalyzeSystemPromptForFile() should contain <role> XML tag")
			}
			if !strings.Contains(prompt, "<workflow>") {
				t.Error("BuildAnalyzeSystemPromptForFile() should contain <workflow> XML tag")
			}
			if !strings.Contains(prompt, "<output_quality>") {
				t.Error("BuildAnalyzeSystemPromptForFile() should contain <output_quality> XML tag")
			}

			// Verify it does NOT contain the aspirational grounding rules
			if strings.Contains(prompt, "only what the user stated") {
				t.Error("BuildAnalyzeSystemPromptForFile() should NOT contain aspirational grounding language")
			}
		})
	}
}

func TestPromptBuilder_BuildAnalyzeSystemPromptForFile_Locale(t *testing.T) {
	builder := NewPromptBuilder()

	promptEN := builder.BuildAnalyzeSystemPromptForFile("agents", "en")
	promptES := builder.BuildAnalyzeSystemPromptForFile("agents", "es")

	if !strings.Contains(promptEN, "English") {
		t.Error("English locale prompt should contain 'English'")
	}
	if !strings.Contains(promptES, "Spanish") {
		t.Error("Spanish locale prompt should contain 'Spanish'")
	}
}

func TestPromptBuilder_BuildSpecSystemPrompt(t *testing.T) {
	builder := NewPromptBuilder()

	existingContext := "# My Project Context\n\nThis is the architecture description."
	prompt := builder.BuildSpecSystemPrompt(existingContext, "en", "")

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
