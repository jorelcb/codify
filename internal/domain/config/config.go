// Package config define el modelo de dominio para la configuración de Codify
// a nivel de usuario (~/.codify/config.yml) y proyecto (.codify/config.yml).
//
// La configuración persiste defaults intencionales del usuario o del equipo
// y se aplica con precedencia: flags > project > user > built-in defaults
// (ver ADR-007 y la sección 2.2 de research/IMPLEMENTATION_PLAN.md).
package config

import (
	"fmt"
)

// SchemaVersion es la versión del schema de config.yml. Se persiste para
// permitir migraciones futuras sin romper archivos viejos.
const SchemaVersion = "1.0"

// Config representa la configuración persistente de Codify, ya sea a nivel
// usuario o a nivel proyecto. Los campos vacíos significan "no override —
// usar la siguiente capa de la cadena de precedencia".
type Config struct {
	// Version del schema (no de Codify). Permite migraciones forward-compatible.
	Version string `yaml:"version,omitempty"`

	// Preset arquitectónico por default: neutral, clean-ddd, hexagonal, event-driven.
	Preset string `yaml:"preset,omitempty"`

	// Locale por default: "en" o "es".
	Locale string `yaml:"locale,omitempty"`

	// Language idiomático por default: go, javascript, python, etc.
	Language string `yaml:"language,omitempty"`

	// Model LLM por default: claude-sonnet-4-6, gemini-3.1-pro-preview, etc.
	Model string `yaml:"model,omitempty"`

	// Target ecosystem por default: claude, codex, antigravity.
	Target string `yaml:"target,omitempty"`

	// Provider LLM explícito (anthropic, gemini). Si vacío, se infiere del Model.
	Provider string `yaml:"provider,omitempty"`

	// ProjectName solo aplica a nivel proyecto (.codify/config.yml).
	ProjectName string `yaml:"project_name,omitempty"`

	// CreatedAt timestamp ISO 8601 de cuándo se creó este archivo.
	CreatedAt string `yaml:"created_at,omitempty"`

	// UpdatedAt timestamp ISO 8601 de la última modificación.
	UpdatedAt string `yaml:"updated_at,omitempty"`
}

// BuiltinDefaults devuelve la configuración con los valores por default del
// binario — el último escalón de la cadena de precedencia. NO incluye
// ProjectName ni timestamps; esos solo aplican a archivos persistidos.
func BuiltinDefaults() Config {
	return Config{
		Version:  SchemaVersion,
		Preset:   "clean-ddd", // ADR-001: cambia a "neutral" en v2.0
		Locale:   "en",
		Target:   "claude",
		Language: "",
		Model:    "",
		Provider: "",
	}
}

// Merge aplica overrides sobre el receiver. Solo campos no-vacíos del
// override sobreescriben campos del receiver. Esto implementa la regla de
// precedencia "el override tiene prioridad si está seteado".
//
// Uso típico (de menor a mayor prioridad):
//
//	cfg := config.BuiltinDefaults()
//	cfg.Merge(userConfig)    // ~/.codify/config.yml
//	cfg.Merge(projectConfig) // .codify/config.yml
//	cfg.Merge(flagsConfig)   // CLI flags
func (c *Config) Merge(override Config) {
	if override.Version != "" {
		c.Version = override.Version
	}
	if override.Preset != "" {
		c.Preset = override.Preset
	}
	if override.Locale != "" {
		c.Locale = override.Locale
	}
	if override.Language != "" {
		c.Language = override.Language
	}
	if override.Model != "" {
		c.Model = override.Model
	}
	if override.Target != "" {
		c.Target = override.Target
	}
	if override.Provider != "" {
		c.Provider = override.Provider
	}
	if override.ProjectName != "" {
		c.ProjectName = override.ProjectName
	}
}

// Get devuelve el valor de un campo por nombre (key). Soporta los keys
// expuestos públicamente; campos privados como CreatedAt/UpdatedAt no son
// accesibles desde la CLI.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "preset":
		return c.Preset, nil
	case "locale":
		return c.Locale, nil
	case "language":
		return c.Language, nil
	case "model":
		return c.Model, nil
	case "target":
		return c.Target, nil
	case "provider":
		return c.Provider, nil
	case "project_name":
		return c.ProjectName, nil
	case "version":
		return c.Version, nil
	default:
		return "", fmt.Errorf("unknown config key: %q (valid: preset, locale, language, model, target, provider, project_name)", key)
	}
}

// Set asigna un valor a un campo por nombre. Devuelve error si el key es
// desconocido. Validación semántica del valor (e.g., preset válido) sucede
// en una capa superior, no acá.
func (c *Config) Set(key, value string) error {
	switch key {
	case "preset":
		c.Preset = value
	case "locale":
		c.Locale = value
	case "language":
		c.Language = value
	case "model":
		c.Model = value
	case "target":
		c.Target = value
	case "provider":
		c.Provider = value
	case "project_name":
		c.ProjectName = value
	default:
		return fmt.Errorf("unknown config key: %q (valid: preset, locale, language, model, target, provider, project_name)", key)
	}
	return nil
}

// Unset limpia un campo (lo deja vacío, equivalente a "no hay override").
func (c *Config) Unset(key string) error {
	return c.Set(key, "")
}

// Keys devuelve la lista de keys públicos en orden estable.
func Keys() []string {
	return []string{
		"preset",
		"locale",
		"language",
		"model",
		"target",
		"provider",
		"project_name",
	}
}
