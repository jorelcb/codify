package config

import "testing"

func TestBuiltinDefaults(t *testing.T) {
	cfg := BuiltinDefaults()
	if cfg.Preset != "clean-ddd" {
		t.Errorf("default preset: got %q, want %q", cfg.Preset, "clean-ddd")
	}
	if cfg.Locale != "en" {
		t.Errorf("default locale: got %q, want %q", cfg.Locale, "en")
	}
	if cfg.Target != "claude" {
		t.Errorf("default target: got %q, want %q", cfg.Target, "claude")
	}
	if cfg.Version != SchemaVersion {
		t.Errorf("schema version: got %q, want %q", cfg.Version, SchemaVersion)
	}
}

func TestMerge_OverrideNonEmptyFields(t *testing.T) {
	base := Config{Preset: "clean-ddd", Locale: "en", Target: "claude"}
	override := Config{Preset: "hexagonal", Language: "go"}
	base.Merge(override)
	if base.Preset != "hexagonal" {
		t.Errorf("preset should be overridden: got %q", base.Preset)
	}
	if base.Locale != "en" {
		t.Errorf("locale should not be overridden when override is empty: got %q", base.Locale)
	}
	if base.Target != "claude" {
		t.Errorf("target should not be overridden: got %q", base.Target)
	}
	if base.Language != "go" {
		t.Errorf("language should be set from override: got %q", base.Language)
	}
}

func TestMerge_PrecedenceChain(t *testing.T) {
	cfg := BuiltinDefaults()
	user := Config{Preset: "neutral", Locale: "es"}
	project := Config{Preset: "hexagonal"}
	flags := Config{Model: "claude-opus-4-7"}

	cfg.Merge(user)
	cfg.Merge(project)
	cfg.Merge(flags)

	if cfg.Preset != "hexagonal" {
		t.Errorf("project should win over user: got %q, want hexagonal", cfg.Preset)
	}
	if cfg.Locale != "es" {
		t.Errorf("user should win over builtin: got %q, want es", cfg.Locale)
	}
	if cfg.Model != "claude-opus-4-7" {
		t.Errorf("flags should win: got %q", cfg.Model)
	}
	if cfg.Target != "claude" {
		t.Errorf("builtin should fill the gap: got %q", cfg.Target)
	}
}

func TestGetSet(t *testing.T) {
	c := &Config{}
	if err := c.Set("preset", "hexagonal"); err != nil {
		t.Fatalf("set preset: %v", err)
	}
	v, err := c.Get("preset")
	if err != nil {
		t.Fatalf("get preset: %v", err)
	}
	if v != "hexagonal" {
		t.Errorf("got %q, want hexagonal", v)
	}
	if err := c.Set("nonexistent", "x"); err == nil {
		t.Error("expected error setting unknown key")
	}
	if _, err := c.Get("nonexistent"); err == nil {
		t.Error("expected error getting unknown key")
	}
}

func TestUnset(t *testing.T) {
	c := &Config{Preset: "hexagonal", Locale: "es"}
	if err := c.Unset("preset"); err != nil {
		t.Fatalf("unset preset: %v", err)
	}
	if c.Preset != "" {
		t.Errorf("preset should be empty after unset: got %q", c.Preset)
	}
	if c.Locale != "es" {
		t.Errorf("locale should be untouched: got %q", c.Locale)
	}
}
