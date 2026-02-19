package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemTemplateLoader_LoadAll(t *testing.T) {
	// Create temp dir with test templates
	tmpDir := t.TempDir()

	templates := map[string]string{
		"prompt.template":       "# Prompt template content",
		"context.template":      "# Context template content",
		"scaffolding.template":  "# Scaffolding template content",
		"interactions.template": "# Interactions template content",
	}

	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test template %s: %v", name, err)
		}
	}

	loader := NewFileSystemTemplateLoader(tmpDir)
	guides, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(guides) != 4 {
		t.Errorf("LoadAll() returned %d guides, want 4", len(guides))
	}

	// Verify all expected names are present
	nameSet := make(map[string]bool)
	for _, g := range guides {
		nameSet[g.Name] = true
		if g.Content == "" {
			t.Errorf("guide %s has empty content", g.Name)
		}
	}

	expectedNames := []string{"prompt", "context", "scaffolding", "interactions"}
	for _, name := range expectedNames {
		if !nameSet[name] {
			t.Errorf("missing expected guide: %s", name)
		}
	}
}

func TestFileSystemTemplateLoader_LoadAll_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create one template, not all four
	if err := os.WriteFile(filepath.Join(tmpDir, "prompt.template"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write test template: %v", err)
	}

	loader := NewFileSystemTemplateLoader(tmpDir)
	_, err := loader.LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail when templates are missing")
	}
}

func TestFileSystemTemplateLoader_LoadAll_EmptyPath(t *testing.T) {
	loader := NewFileSystemTemplateLoader("/nonexistent/path")
	_, err := loader.LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail with nonexistent path")
	}
}
