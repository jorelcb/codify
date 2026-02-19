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
		"agents.template":       "# Agents template content",
		"context.template":      "# Context template content",
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

	if len(guides) != 3 {
		t.Errorf("LoadAll() returned %d guides, want 3", len(guides))
	}

	// Verify all expected names are present
	nameSet := make(map[string]bool)
	for _, g := range guides {
		nameSet[g.Name] = true
		if g.Content == "" {
			t.Errorf("guide %s has empty content", g.Name)
		}
	}

	expectedNames := []string{"agents", "context", "interactions"}
	for _, name := range expectedNames {
		if !nameSet[name] {
			t.Errorf("missing expected guide: %s", name)
		}
	}
}

func TestFileSystemTemplateLoader_LoadAll_WithCustomMapping(t *testing.T) {
	tmpDir := t.TempDir()

	templates := map[string]string{
		"constitution.template": "# Constitution content",
		"spec.template":         "# Spec content",
		"plan.template":         "# Plan content",
		"tasks.template":        "# Tasks content",
	}

	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test template %s: %v", name, err)
		}
	}

	customMapping := map[string]string{
		"constitution.template": "constitution",
		"spec.template":         "spec",
		"plan.template":         "plan",
		"tasks.template":        "tasks",
	}

	loader := NewFileSystemTemplateLoaderWithMapping(tmpDir, customMapping)
	guides, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(guides) != 4 {
		t.Errorf("LoadAll() returned %d guides, want 4", len(guides))
	}

	nameSet := make(map[string]bool)
	for _, g := range guides {
		nameSet[g.Name] = true
	}

	expectedNames := []string{"constitution", "spec", "plan", "tasks"}
	for _, name := range expectedNames {
		if !nameSet[name] {
			t.Errorf("missing expected guide: %s", name)
		}
	}
}

func TestFileSystemTemplateLoader_LoadAll_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create one template, not all three
	if err := os.WriteFile(filepath.Join(tmpDir, "agents.template"), []byte("content"), 0644); err != nil {
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
