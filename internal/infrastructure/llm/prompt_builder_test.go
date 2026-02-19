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
		{"prompt", "PROMPT.md"},
		{"context", "CONTEXT.md"},
		{"scaffolding", "SCAFFOLDING.md"},
		{"interactions", "INTERACTIONS_LOG.md"},
	}

	for _, tt := range tests {
		t.Run(tt.guideName, func(t *testing.T) {
			prompt := builder.BuildSystemPromptForFile(tt.guideName)

			if prompt == "" {
				t.Error("BuildSystemPromptForFile() returned empty string")
			}
			if !strings.Contains(prompt, tt.wantFileName) {
				t.Errorf("BuildSystemPromptForFile() should mention %s", tt.wantFileName)
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
		Name:    "prompt",
		Content: "# Prompt template content here",
	}

	msg := builder.BuildUserMessageForFile(req, guide)

	if msg == "" {
		t.Error("BuildUserMessageForFile() returned empty string")
	}

	// Verify description is included
	if !strings.Contains(msg, "API REST de gestion de inventarios en Go") {
		t.Error("BuildUserMessageForFile() missing project description")
	}

	// Verify optional fields
	if !strings.Contains(msg, "go") {
		t.Error("BuildUserMessageForFile() missing language")
	}

	// Verify template content
	if !strings.Contains(msg, "# Prompt template content here") {
		t.Error("BuildUserMessageForFile() missing template content")
	}

	// Verify file name reference
	if !strings.Contains(msg, "PROMPT.md") {
		t.Error("BuildUserMessageForFile() missing output file name")
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

	if strings.Contains(msg, "**Lenguaje:**") {
		t.Error("should not include language when empty")
	}
	if strings.Contains(msg, "**Tipo de proyecto:**") {
		t.Error("should not include type when empty")
	}
	if strings.Contains(msg, "**Arquitectura:**") {
		t.Error("should not include architecture when empty")
	}
}

func TestFileOutputName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"prompt", "PROMPT.md"},
		{"context", "CONTEXT.md"},
		{"scaffolding", "SCAFFOLDING.md"},
		{"interactions", "INTERACTIONS_LOG.md"},
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
