package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// ValidTargets maps valid target ecosystem names.
var ValidTargets = map[string]bool{
	"claude":      true,
	"codex":       true,
	"antigravity": true,
}

// SkillsConfig representa la configuracion para generar Agent Skills reutilizables
type SkillsConfig struct {
	Category   string // "architecture" o "workflow"
	Preset     string // sub-opcion: "clean", "neutral", "conventional-commit", "all", etc.
	Locale     string // "en" o "es"
	Target     string // ecosistema destino: "claude", "codex", "antigravity"
	Model      string
	OutputPath string
}

// Validate valida la configuracion de skills
func (sc *SkillsConfig) Validate() error {
	if sc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	if !ValidTargets[sc.Target] {
		return shared.ErrInvalidInput("invalid target: must be claude, codex, or antigravity")
	}
	return nil
}
