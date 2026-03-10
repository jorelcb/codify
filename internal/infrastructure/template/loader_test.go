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
		"agents.template":            "# Agents template content",
		"context.template":           "# Context template content",
		"interactions.template":      "# Interactions template content",
		"development_guide.template": "# Development guide content",
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

	expectedNames := []string{"agents", "context", "interactions", "development_guide"}
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

func TestFileSystemTemplateLoader_LoadAll_WithLanguage(t *testing.T) {
	// Create base template dir
	tmpDir := t.TempDir()

	baseTemplates := map[string]string{
		"agents.template":            "# Agents",
		"context.template":           "# Context",
		"interactions.template":      "# Interactions",
		"development_guide.template": "# Dev Guide",
	}
	for name, content := range baseTemplates {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write base template %s: %v", name, err)
		}
	}

	// Create language template dir (simulating templates/languages/go/)
	langDir := filepath.Join(tmpDir, "languages", "go")
	if err := os.MkdirAll(langDir, 0755); err != nil {
		t.Fatalf("failed to create lang dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(langDir, "idioms.template"), []byte("# Go idioms"), 0644); err != nil {
		t.Fatalf("failed to write lang template: %v", err)
	}

	// The loader resolves language templates relative to "templates/languages/{lang}/"
	// For testing, we override the language dir resolution by using the loader directly
	loader := &FileSystemTemplateLoader{
		basePath: tmpDir,
		mapping:  templateMapping,
		language: "go",
	}
	// Override language loading to use our temp dir
	guides, err := loader.loadFromMapping(tmpDir, templateMapping)
	if err != nil {
		t.Fatalf("loadFromMapping() error = %v", err)
	}

	langGuides, err := loader.loadFromMapping(langDir, languageTemplateMapping)
	if err != nil {
		t.Fatalf("loadFromMapping() for language error = %v", err)
	}
	guides = append(guides, langGuides...)

	if len(guides) != 5 {
		t.Errorf("Expected 5 guides (4 base + 1 language), got %d", len(guides))
	}

	nameSet := make(map[string]bool)
	for _, g := range guides {
		nameSet[g.Name] = true
	}

	if !nameSet["idioms"] {
		t.Error("missing language-specific guide: idioms")
	}
}

func TestFileSystemTemplateLoader_LoadAll_WithUnknownLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	baseTemplates := map[string]string{
		"agents.template":            "# Agents",
		"context.template":           "# Context",
		"interactions.template":      "# Interactions",
		"development_guide.template": "# Dev Guide",
	}
	for name, content := range baseTemplates {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write base template %s: %v", name, err)
		}
	}

	// Language "rust" has no template directory — should load base templates without error
	loader := &FileSystemTemplateLoader{
		basePath: tmpDir,
		mapping:  templateMapping,
		language: "rust",
	}
	guides, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() should not fail for unknown language, got: %v", err)
	}

	if len(guides) != 4 {
		t.Errorf("Expected 4 base guides for unknown language, got %d", len(guides))
	}
}

func TestFileSystemTemplateLoader_LoadAll_EmptyPath(t *testing.T) {
	loader := NewFileSystemTemplateLoader("/nonexistent/path")
	_, err := loader.LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail with nonexistent path")
	}
}
