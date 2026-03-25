package dto

import "github.com/jorelcb/codify/internal/domain/shared"

// ValidTargets maps valid target ecosystem names.
var ValidTargets = map[string]bool{
	"claude":      true,
	"codex":       true,
	"antigravity": true,
}

// Skills generation modes
const (
	SkillModeStatic       = "static"
	SkillModePersonalized = "personalized"
)

// Skills install scopes
const (
	InstallScopeGlobal  = "global"
	InstallScopeProject = "project"
)

// SkillsConfig holds configuration for generating reusable Agent Skills
type SkillsConfig struct {
	Category       string // "architecture" or "workflow"
	Preset         string // "clean", "neutral", "conventional-commit", "all", etc.
	Mode           string // "static" or "personalized"
	Locale         string // "en" or "es"
	Target         string // target ecosystem: "claude", "codex", "antigravity"
	Model          string
	OutputPath     string
	ProjectContext string // project context for personalized mode
	Install        string // install scope: "global", "project", or "" (custom output)
}

// Validate validates the skills configuration
func (sc *SkillsConfig) Validate() error {
	if sc.OutputPath == "" {
		return shared.ErrInvalidInput("output path is required")
	}
	if !ValidTargets[sc.Target] {
		return shared.ErrInvalidInput("invalid target: must be claude, codex, or antigravity")
	}
	return nil
}
