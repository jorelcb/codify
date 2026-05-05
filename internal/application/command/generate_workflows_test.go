package command

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
)

func TestGenerateWorkflows_AntigravityWritesFlatFiles(t *testing.T) {
	tmp := t.TempDir()
	mock := llm.NewMockProvider()
	cmd := NewGenerateWorkflowsCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())

	cfg := &dto.WorkflowConfig{
		Preset:         "all",
		Target:         "antigravity",
		Mode:           "personalized",
		Locale:         "en",
		OutputPath:     tmp,
		ProjectContext: "Go monorepo",
	}

	guides := []service.TemplateGuide{
		{Name: "bug_fix", Content: "guide bug"},
		{Name: "release_cycle", Content: "guide release"},
	}

	result, err := cmd.Execute(context.Background(), cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles: got %d, want 2", len(result.GeneratedFiles))
	}

	for _, expected := range []string{"bug-fix.md", "release-cycle.md"} {
		p := filepath.Join(tmp, expected)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	last := mock.LastCall()
	if last.Mode != "workflows" {
		t.Fatalf("Mode for antigravity: got %q, want workflows", last.Mode)
	}
}

func TestGenerateWorkflows_ClaudeWritesSkillSubdirs(t *testing.T) {
	tmp := t.TempDir()
	mock := llm.NewMockProvider()
	cmd := NewGenerateWorkflowsCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())

	cfg := &dto.WorkflowConfig{
		Preset:         "all",
		Target:         "claude",
		Mode:           "personalized",
		Locale:         "en",
		OutputPath:     tmp,
		ProjectContext: "Go cli",
	}

	guides := []service.TemplateGuide{{Name: "release_cycle", Content: "guide release"}}

	result, err := cmd.Execute(context.Background(), cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 1 {
		t.Fatalf("GeneratedFiles: got %d, want 1", len(result.GeneratedFiles))
	}
	if _, err := os.Stat(filepath.Join(tmp, "release-cycle", "SKILL.md")); err != nil {
		t.Errorf("missing release-cycle/SKILL.md: %v", err)
	}

	last := mock.LastCall()
	if last.Mode != "workflow-skills" {
		t.Fatalf("Mode for claude: got %q, want workflow-skills", last.Mode)
	}
}

func TestGenerateWorkflows_DefaultsToAntigravity(t *testing.T) {
	mock := llm.NewMockProvider()
	cmd := NewGenerateWorkflowsCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())

	cfg := &dto.WorkflowConfig{
		Preset:         "all",
		Target:         "", // unset
		Mode:           "personalized",
		Locale:         "en",
		OutputPath:     t.TempDir(),
		ProjectContext: "ctx",
	}
	guides := []service.TemplateGuide{{Name: "bug_fix", Content: "x"}}

	if _, err := cmd.Execute(context.Background(), cfg, guides); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if mock.LastCall().Target != "antigravity" {
		t.Fatalf("Target default: got %q, want antigravity", mock.LastCall().Target)
	}
}
