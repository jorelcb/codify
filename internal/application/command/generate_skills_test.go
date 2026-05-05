package command

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
)

func TestGenerateSkills_WritesSkillPerGuide(t *testing.T) {
	tmp := t.TempDir()
	mock := llm.NewMockProvider()
	mock.Responses = map[string]string{
		"ddd_entity":       "---\nname: ddd-entity\n---\n# DDD entity body padded long enough so output validators do not flag length warnings during the test run.",
		"clean_arch_layer": "---\nname: clean-arch-layer\n---\n# Clean arch body padded long enough so output validators do not flag length warnings during the test run.",
	}

	cmd := NewGenerateSkillsCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())

	cfg := &dto.SkillsConfig{
		Category:       "architecture",
		Preset:         "clean",
		Mode:           dto.SkillModePersonalized,
		Locale:         "en",
		Target:         "claude",
		OutputPath:     tmp,
		ProjectContext: "Go DDD service",
	}

	guides := []service.TemplateGuide{
		{Name: "ddd_entity", Content: "guide a"},
		{Name: "clean_arch_layer", Content: "guide b"},
	}

	result, err := cmd.Execute(context.Background(), cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles: got %d, want 2", len(result.GeneratedFiles))
	}

	expected := []string{
		filepath.Join(tmp, "ddd-entity", "SKILL.md"),
		filepath.Join(tmp, "clean-arch-layer", "SKILL.md"),
	}
	for _, p := range expected {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing file %s: %v", p, err)
		}
	}

	// Verify the mode and project context propagated to the LLM call.
	last := mock.LastCall()
	if last.Mode != "skills" {
		t.Fatalf("Mode: got %q, want skills", last.Mode)
	}
	if last.ProjectContext != "Go DDD service" {
		t.Fatalf("ProjectContext: got %q", last.ProjectContext)
	}
	if last.Target != "claude" {
		t.Fatalf("Target: got %q", last.Target)
	}
}

func TestGenerateSkills_LLMErrorPropagates(t *testing.T) {
	mock := llm.NewMockProvider()
	mock.Err = errors.New("rate limit")

	cmd := NewGenerateSkillsCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())
	cfg := &dto.SkillsConfig{
		Category:       "architecture",
		Preset:         "clean",
		Mode:           dto.SkillModePersonalized,
		Locale:         "en",
		Target:         "claude",
		OutputPath:     t.TempDir(),
		ProjectContext: "ctx",
	}

	_, err := cmd.Execute(context.Background(), cfg, []service.TemplateGuide{{Name: "ddd_entity", Content: "x"}})
	if err == nil {
		t.Fatal("expected error from LLM provider")
	}
}
