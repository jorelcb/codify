package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	statedomain "github.com/jorelcb/codify/internal/domain/state"
)

func TestHashFile_Deterministic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	h1, ok, err := HashFile(path)
	if err != nil || !ok {
		t.Fatalf("hash: %v ok=%v", err, ok)
	}
	h2, _, _ := HashFile(path)
	if h1 != h2 {
		t.Errorf("hash should be deterministic: %s vs %s", h1, h2)
	}
	// Known SHA256("hello") = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	if h1 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("unexpected hash: %s", h1)
	}
}

func TestHashFile_Missing(t *testing.T) {
	_, ok, err := HashFile(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Errorf("missing should not error, got: %v", err)
	}
	if ok {
		t.Error("ok should be false for missing")
	}
}

func TestCountLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("a\nb\nc"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, ok, err := CountLines(path)
	if err != nil || !ok {
		t.Fatalf("count: %v ok=%v", err, ok)
	}
	if n != 3 {
		t.Errorf("got %d lines, want 3", n)
	}
}

func TestBuild_HashesArtifactsAndSignals(t *testing.T) {
	projectPath := t.TempDir()
	outputPath := t.TempDir()

	// Setup artifact (under outputPath, not nested)
	if err := os.WriteFile(filepath.Join(outputPath, "AGENTS.md"), []byte("agents content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outputPath, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputPath, "context", "CONTEXT.md"), []byte("ctx"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Setup input signal
	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := Build(BuildOptions{
		ProjectPath:   projectPath,
		OutputPath:    outputPath,
		Project:       statedomain.ProjectInfo{Name: "test", Preset: "neutral", Locale: "en"},
		GeneratedBy:   "test",
		CodifyVersion: "1.23.0",
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, ok := state.Artifacts["AGENTS.md"]; !ok {
		t.Errorf("AGENTS.md missing from artifacts; got %v", state.Artifacts)
	}
	if _, ok := state.Artifacts["context/CONTEXT.md"]; !ok {
		t.Errorf("context/CONTEXT.md missing")
	}
	if _, ok := state.InputSignals["go.mod"]; !ok {
		t.Errorf("go.mod signal missing")
	}
	if state.GeneratedBy != "test" {
		t.Errorf("generated_by: %q", state.GeneratedBy)
	}
	if state.CodifyVersion != "1.23.0" {
		t.Errorf("codify version: %q", state.CodifyVersion)
	}
}
