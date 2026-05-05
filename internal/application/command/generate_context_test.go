package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
)

// mockLLMProvider implements service.LLMProvider for testing.
type mockLLMProvider struct {
	response *service.GenerationResponse
	err      error
}

func (m *mockLLMProvider) GenerateContext(_ context.Context, _ service.GenerationRequest) (*service.GenerationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// EvaluatePrompt is required by the LLMProvider interface but unused by
// generate_context tests. Returns a fixed empty response.
func (m *mockLLMProvider) EvaluatePrompt(_ context.Context, _ service.EvaluationRequest) (*service.EvaluationResponse, error) {
	return &service.EvaluationResponse{Text: "", Model: "mock"}, nil
}

func TestGenerateContextCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	mockProvider := &mockLLMProvider{
		response: &service.GenerationResponse{
			Files: []service.GeneratedFile{
				{Name: "AGENTS.md", Content: "# Agents content"},
				{Name: "CONTEXT.md", Content: "# Context content"},
				{Name: "INTERACTIONS_LOG.md", Content: "# Interactions content"},
			},
			Model:     "claude-sonnet-4-6",
			TokensIn:  1000,
			TokensOut: 5000,
		},
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := NewGenerateContextCommand(mockProvider, fileWriter, dirManager)

	config := &dto.ProjectConfig{
		Name:        "test-project",
		Description: "A test project for unit testing the generation pipeline",
		OutputPath:  tmpDir,
	}

	guides := []service.TemplateGuide{
		{Name: "agents", Content: "# Agents guide"},
	}

	result, err := cmd.Execute(context.Background(), config, guides)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Model != "claude-sonnet-4-6" {
		t.Errorf("Expected model claude-sonnet-4-6, got %s", result.Model)
	}
	if result.TokensIn != 1000 {
		t.Errorf("Expected 1000 tokens in, got %d", result.TokensIn)
	}
	if result.TokensOut != 5000 {
		t.Errorf("Expected 5000 tokens out, got %d", result.TokensOut)
	}
	if len(result.GeneratedFiles) != 3 {
		t.Errorf("Expected 3 generated files, got %d", len(result.GeneratedFiles))
	}

	// Verify AGENTS.md was written to project root (not context/)
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Errorf("AGENTS.md not found at root: %v", err)
	} else if len(content) == 0 {
		t.Error("AGENTS.md is empty")
	}

	// Verify other files were written to context/ subdirectory
	contextFiles := []string{"CONTEXT.md", "INTERACTIONS_LOG.md"}
	for _, fname := range contextFiles {
		fpath := filepath.Join(tmpDir, "context", fname)
		content, err := os.ReadFile(fpath)
		if err != nil {
			t.Errorf("File %s not found in context/: %v", fname, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("File %s is empty", fname)
		}
	}
}

func TestGenerateContextCommand_Execute_LLMError(t *testing.T) {
	mockProvider := &mockLLMProvider{
		err: fmt.Errorf("API rate limit exceeded"),
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := NewGenerateContextCommand(mockProvider, fileWriter, dirManager)

	config := &dto.ProjectConfig{
		Name:        "test-project",
		Description: "A test project",
		OutputPath:  t.TempDir(),
	}

	_, err := cmd.Execute(context.Background(), config, nil)
	if err == nil {
		t.Error("Execute() should fail when LLM returns error")
	}
}
