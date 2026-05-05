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

func TestGenerateSpec_WritesUnderSpecsDir(t *testing.T) {
	tmp := t.TempDir()
	mock := llm.NewMockProvider()
	mock.Responses = map[string]string{
		"constitution": "# Constitution body padded long enough so output validators do not flag length warnings during the test run.",
		"spec":         "# Spec body padded long enough so output validators do not flag length warnings during the test run.",
	}

	cmd := NewGenerateSpecCommand(mock, filesystem.NewFileWriter(), filesystem.NewDirectoryManager())
	cfg := &dto.SpecConfig{
		ProjectName:     "test",
		FromContextPath: tmp,
		OutputPath:      tmp,
		Locale:          "en",
	}

	guides := []service.TemplateGuide{
		{Name: "constitution", Content: "g"},
		{Name: "spec", Content: "g"},
	}

	result, err := cmd.Execute(context.Background(), cfg, "EXISTING CONTEXT", guides)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.GeneratedFiles) != 2 {
		t.Fatalf("GeneratedFiles: got %d, want 2", len(result.GeneratedFiles))
	}
	for _, name := range []string{"CONSTITUTION.md", "SPEC.md"} {
		p := filepath.Join(tmp, "specs", name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	last := mock.LastCall()
	if last.Mode != "spec" {
		t.Fatalf("Mode: got %q, want spec", last.Mode)
	}
	if last.ExistingContext != "EXISTING CONTEXT" {
		t.Fatalf("ExistingContext: got %q", last.ExistingContext)
	}
}
