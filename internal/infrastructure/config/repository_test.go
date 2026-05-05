package config

import (
	"os"
	"path/filepath"
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/config"
)

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")

	repo := NewRepository()
	cfg := domain.Config{
		Preset:   "hexagonal",
		Locale:   "es",
		Language: "go",
	}
	if err := repo.Save(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, ok, err := repo.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !ok {
		t.Fatal("expected file to exist")
	}
	if loaded.Preset != "hexagonal" || loaded.Locale != "es" || loaded.Language != "go" {
		t.Errorf("roundtrip mismatch: %+v", loaded)
	}
	if loaded.Version != domain.SchemaVersion {
		t.Errorf("version not set on save: got %q", loaded.Version)
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt should be set on first save")
	}
	if loaded.UpdatedAt == "" {
		t.Error("UpdatedAt should be set on save")
	}
}

func TestLoad_Missing(t *testing.T) {
	repo := NewRepository()
	cfg, ok, err := repo.Load(filepath.Join(t.TempDir(), "missing.yml"))
	if err != nil {
		t.Fatalf("missing file should not error, got: %v", err)
	}
	if ok {
		t.Error("ok should be false for missing file")
	}
	if cfg.Preset != "" {
		t.Errorf("config should be zero, got %+v", cfg)
	}
}

func TestSave_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	repo := NewRepository()

	if err := repo.Save(path, domain.Config{Preset: "neutral"}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := repo.Save(path, domain.Config{Preset: "hexagonal"}); err != nil {
		t.Fatalf("second save: %v", err)
	}

	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Errorf("backup should exist: %v", err)
	}
}

func TestLoadEffective_PrecedenceFromPaths(t *testing.T) {
	// Override HOME y cwd a temp dirs para que Load resuelva contra ellos.
	homeDir := t.TempDir()
	cwdDir := t.TempDir()

	t.Setenv("HOME", homeDir)
	originalCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalCwd) }()
	if err := os.Chdir(cwdDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	repo := NewRepository()
	// User config: setea preset=neutral
	userDir := filepath.Join(homeDir, ".codify")
	_ = os.MkdirAll(userDir, 0o755)
	if err := repo.Save(filepath.Join(userDir, "config.yml"), domain.Config{Preset: "neutral", Locale: "es"}); err != nil {
		t.Fatalf("save user: %v", err)
	}
	// Project config: override preset=hexagonal
	projDir := filepath.Join(cwdDir, ".codify")
	_ = os.MkdirAll(projDir, 0o755)
	if err := repo.Save(filepath.Join(projDir, "config.yml"), domain.Config{Preset: "hexagonal"}); err != nil {
		t.Fatalf("save project: %v", err)
	}

	cfg, err := repo.LoadEffective()
	if err != nil {
		t.Fatalf("LoadEffective: %v", err)
	}
	if cfg.Preset != "hexagonal" {
		t.Errorf("project should win: got %q, want hexagonal", cfg.Preset)
	}
	if cfg.Locale != "es" {
		t.Errorf("user should fill: got %q, want es", cfg.Locale)
	}
	if cfg.Target != "claude" {
		t.Errorf("builtin default should fill: got %q, want claude", cfg.Target)
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if FileExists(path) {
		t.Error("missing file: should report false")
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(path) {
		t.Error("existing file: should report true")
	}
	if FileExists(dir) {
		t.Error("directory: should not report as file")
	}
}
