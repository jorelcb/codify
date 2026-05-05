package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
)

func TestDeliverStaticSkills_WritesPerSkillDir(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewDeliverStaticSkillsCommand(filesystem.NewFileWriter(), filesystem.NewDirectoryManager())

	cfg := &dto.SkillsConfig{
		Category:   "architecture",
		Preset:     "clean",
		Mode:       dto.SkillModeStatic,
		Locale:     "en",
		Target:     "claude",
		OutputPath: tmp,
	}
	guides := []service.TemplateGuide{
		{Name: "ddd_entity", Content: "# DDD entity guide"},
		{Name: "clean_arch_layer", Content: "# Clean arch layer guide"},
	}

	result, err := cmd.Execute(cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles: got %d, want 2", len(result.GeneratedFiles))
	}

	// Underscores in guide names map to hyphens in directory names.
	for _, sub := range []string{"ddd-entity", "clean-arch-layer"} {
		p := filepath.Join(tmp, sub, "SKILL.md")
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("missing %s: %v", p, err)
			continue
		}
		if !strings.HasPrefix(string(data), "---") {
			t.Errorf("%s missing frontmatter delimiter", p)
		}
	}
}

func TestDeliverStaticWorkflows_AntigravityFlatLayout(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewDeliverStaticWorkflowsCommand(filesystem.NewFileWriter(), filesystem.NewDirectoryManager())
	cfg := &dto.WorkflowConfig{
		Preset:     "all",
		Target:     "antigravity",
		Mode:       "static",
		Locale:     "en",
		OutputPath: tmp,
	}
	guides := []service.TemplateGuide{
		{Name: "bug_fix", Content: "# Bug fix guide"},
		{Name: "release_cycle", Content: "# Release cycle guide"},
	}

	result, err := cmd.Execute(cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles: got %d, want 2", len(result.GeneratedFiles))
	}
	for _, name := range []string{"bug-fix.md", "release-cycle.md"} {
		if _, err := os.Stat(filepath.Join(tmp, name)); err != nil {
			t.Errorf("missing flat workflow %s: %v", name, err)
		}
	}
}

func TestDeliverStaticWorkflows_ClaudeStripsAnnotations(t *testing.T) {
	tmp := t.TempDir()
	cmd := NewDeliverStaticWorkflowsCommand(filesystem.NewFileWriter(), filesystem.NewDirectoryManager())
	cfg := &dto.WorkflowConfig{
		Preset:     "all",
		Target:     "claude",
		Mode:       "static",
		Locale:     "en",
		OutputPath: tmp,
	}
	guides := []service.TemplateGuide{
		{Name: "release_cycle", Content: "step 1\n// turbo\nstep 2"},
	}

	result, err := cmd.Execute(cfg, guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 1 {
		t.Fatalf("GeneratedFiles: got %d, want 1", len(result.GeneratedFiles))
	}

	body, err := os.ReadFile(filepath.Join(tmp, "release-cycle", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if strings.Contains(string(body), "// turbo") {
		t.Fatalf("expected // turbo annotation to be stripped for Claude target, got:\n%s", body)
	}
}
