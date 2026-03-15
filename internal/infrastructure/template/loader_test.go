package template

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestFileSystemTemplateLoader_LoadAll(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/agents.template":            {Data: []byte("# Agents template content")},
		"templates/context.template":           {Data: []byte("# Context template content")},
		"templates/interactions.template":      {Data: []byte("# Interactions template content")},
		"templates/development_guide.template": {Data: []byte("# Development guide content")},
	}

	loader := NewFileSystemTemplateLoader(fsys, "templates")
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
	fsys := fstest.MapFS{
		"templates/spec/constitution.template": {Data: []byte("# Constitution content")},
		"templates/spec/spec.template":         {Data: []byte("# Spec content")},
		"templates/spec/plan.template":         {Data: []byte("# Plan content")},
		"templates/spec/tasks.template":        {Data: []byte("# Tasks content")},
	}

	customMapping := map[string]string{
		"constitution.template": "constitution",
		"spec.template":         "spec",
		"plan.template":         "plan",
		"tasks.template":        "tasks",
	}

	loader := NewFileSystemTemplateLoaderWithMapping(fsys, "templates/spec", customMapping)
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
	fsys := fstest.MapFS{
		"templates/agents.template": {Data: []byte("content")},
	}

	loader := NewFileSystemTemplateLoader(fsys, "templates")
	_, err := loader.LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail when templates are missing")
	}
}

func TestFileSystemTemplateLoader_LoadAll_WithLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	baseTemplates := map[string]string{
		"templates/default/agents.template":            "# Agents",
		"templates/default/context.template":           "# Context",
		"templates/default/interactions.template":      "# Interactions",
		"templates/default/development_guide.template": "# Dev Guide",
	}
	for name, content := range baseTemplates {
		fullPath := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write template %s: %v", name, err)
		}
	}

	langPath := filepath.Join(tmpDir, "templates", "languages", "go", "idioms.template")
	if err := os.MkdirAll(filepath.Dir(langPath), 0755); err != nil {
		t.Fatalf("failed to create lang dir: %v", err)
	}
	if err := os.WriteFile(langPath, []byte("# Go idioms"), 0644); err != nil {
		t.Fatalf("failed to write lang template: %v", err)
	}

	fsys := os.DirFS(tmpDir)
	loader := NewFileSystemTemplateLoaderWithLanguage(fsys, "templates/default", "templates", "go")
	guides, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

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
	fsys := fstest.MapFS{
		"templates/agents.template":            {Data: []byte("# Agents")},
		"templates/context.template":           {Data: []byte("# Context")},
		"templates/interactions.template":      {Data: []byte("# Interactions")},
		"templates/development_guide.template": {Data: []byte("# Dev Guide")},
	}

	loader := &FileSystemTemplateLoader{
		fsys:       fsys,
		basePath:   "templates",
		mapping:    templateMapping,
		language:   "rust",
		localeBase: "templates",
	}
	guides, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() should not fail for unknown language, got: %v", err)
	}

	if len(guides) != 4 {
		t.Errorf("Expected 4 base guides for unknown language, got %d", len(guides))
	}
}

func TestFileSystemTemplateLoader_LoadAll_NonexistentPath(t *testing.T) {
	fsys := fstest.MapFS{}

	loader := NewFileSystemTemplateLoader(fsys, "nonexistent")
	_, err := loader.LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail with nonexistent path")
	}
}
