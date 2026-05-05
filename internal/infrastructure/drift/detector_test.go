package drift

import (
	"os"
	"path/filepath"
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/drift"
	"github.com/jorelcb/codify/internal/infrastructure/snapshot"
)

func setup(t *testing.T) (projectPath, outputPath string) {
	t.Helper()
	projectPath = t.TempDir()
	outputPath = t.TempDir()
	if err := os.WriteFile(filepath.Join(outputPath, "AGENTS.md"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return projectPath, outputPath
}

// baseline construye el snapshot inicial usando la misma lógica que produce
// init/generate/analyze. Esto evita duplicar logic en los tests.
func baseline(t *testing.T, projectPath, outputPath string) []byte {
	t.Helper()
	state, err := snapshot.Build(snapshot.BuildOptions{
		ProjectPath:   projectPath,
		OutputPath:    outputPath,
		GeneratedBy:   "test",
		CodifyVersion: "test",
	})
	if err != nil {
		t.Fatalf("baseline build: %v", err)
	}
	return mustMarshal(t, state)
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	return nil // placeholder unused
}

func TestDetect_NoDrift(t *testing.T) {
	projectPath, outputPath := setup(t)
	prev, err := snapshot.Build(snapshot.BuildOptions{
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	if err != nil {
		t.Fatalf("build prev: %v", err)
	}
	report, err := NewDetector().Detect(DetectOptions{
		Snapshot:    prev,
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !report.IsEmpty() {
		t.Errorf("expected no drift, got: %+v", report.Entries)
	}
}

func TestDetect_ArtifactModified(t *testing.T) {
	projectPath, outputPath := setup(t)
	prev, _ := snapshot.Build(snapshot.BuildOptions{ProjectPath: projectPath, OutputPath: outputPath})
	// Modify AGENTS.md
	if err := os.WriteFile(filepath.Join(outputPath, "AGENTS.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := NewDetector().Detect(DetectOptions{
		Snapshot:    prev,
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !report.HasSignificant() {
		t.Errorf("expected significant drift, got: %+v", report.Entries)
	}
	found := false
	for _, e := range report.Entries {
		if e.Kind == domain.ArtifactModified && e.Path == "AGENTS.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ArtifactModified for AGENTS.md, got: %+v", report.Entries)
	}
}

func TestDetect_ArtifactMissing(t *testing.T) {
	projectPath, outputPath := setup(t)
	prev, _ := snapshot.Build(snapshot.BuildOptions{ProjectPath: projectPath, OutputPath: outputPath})
	if err := os.Remove(filepath.Join(outputPath, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}

	report, _ := NewDetector().Detect(DetectOptions{
		Snapshot:    prev,
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	found := false
	for _, e := range report.Entries {
		if e.Kind == domain.ArtifactMissing && e.Path == "AGENTS.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ArtifactMissing, got: %+v", report.Entries)
	}
}

func TestDetect_SignalChanged(t *testing.T) {
	projectPath, outputPath := setup(t)
	prev, _ := snapshot.Build(snapshot.BuildOptions{ProjectPath: projectPath, OutputPath: outputPath})
	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module y\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, _ := NewDetector().Detect(DetectOptions{
		Snapshot:    prev,
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	found := false
	for _, e := range report.Entries {
		if e.Kind == domain.SignalChanged && e.Path == "go.mod" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected SignalChanged for go.mod, got: %+v", report.Entries)
	}
	if !report.HasSignificant() {
		t.Errorf("signal change should be significant")
	}
}

func TestDetect_NewArtifact(t *testing.T) {
	projectPath, outputPath := setup(t)
	prev, _ := snapshot.Build(snapshot.BuildOptions{ProjectPath: projectPath, OutputPath: outputPath})
	// Add a new artifact (CONTEXT.md)
	if err := os.MkdirAll(filepath.Join(outputPath, "context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputPath, "context", "CONTEXT.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, _ := NewDetector().Detect(DetectOptions{
		Snapshot:    prev,
		ProjectPath: projectPath,
		OutputPath:  outputPath,
	})
	found := false
	for _, e := range report.Entries {
		if e.Kind == domain.ArtifactNew && e.Path == "context/CONTEXT.md" {
			if e.Severity != domain.Minor {
				t.Errorf("new artifact should be minor severity, got %s", e.Severity)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected ArtifactNew, got: %+v", report.Entries)
	}
}

// silence unused "baseline" warning — keep the helper for future tests.
var _ = baseline
