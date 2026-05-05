package state

import (
	"path/filepath"
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/state"
)

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".codify", "state.json")
	repo := NewRepository()

	st := domain.New()
	st.CodifyVersion = "1.22.0"
	st.GeneratedBy = "init"
	st.Project = domain.ProjectInfo{
		Name:   "test-proj",
		Preset: "clean-ddd",
		Locale: "en",
		Kind:   "new",
	}
	st.Artifacts = map[string]domain.ArtifactInfo{
		"AGENTS.md": {SHA256: "abc", SizeBytes: 1024},
	}

	if err := repo.Save(path, st); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, ok, err := repo.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !ok {
		t.Fatal("expected file to exist")
	}
	if loaded.Project.Name != "test-proj" {
		t.Errorf("project name: got %q", loaded.Project.Name)
	}
	if loaded.Artifacts["AGENTS.md"].SHA256 != "abc" {
		t.Errorf("artifact lost during roundtrip")
	}
	if loaded.SchemaVersion != domain.SchemaVersion {
		t.Errorf("schema version not set: %q", loaded.SchemaVersion)
	}
}

func TestLoad_Missing(t *testing.T) {
	_, ok, err := NewRepository().Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("missing should not error: %v", err)
	}
	if ok {
		t.Error("ok should be false for missing")
	}
}
