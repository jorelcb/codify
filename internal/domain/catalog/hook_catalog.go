package catalog

import "fmt"

// HookMeta contiene metadata de un preset de hooks.
type HookMeta struct {
	Description string // max 250 chars
}

// HookMetadata mapea nombres de preset a su metadata.
var HookMetadata = map[string]HookMeta{
	"linting":                {Description: "Auto-format and lint files on every Edit/Write (PostToolUse)"},
	"security-guardrails":    {Description: "Block dangerous Bash commands and protect sensitive files (PreToolUse)"},
	"convention-enforcement": {Description: "Validate Conventional Commits and protect main branches (PreToolUse with `if`)"},
}

// HookCategories es el registro global de categorias de hooks.
//
// A diferencia de skills/workflows, los hooks NO usan TemplateMapping para resolver
// archivos individualmente: cada preset apunta a un directorio completo
// (templates/{locale}/hooks/{preset}/) cuyo contenido se copia tal cual al output.
// El mapping es nil intencionalmente; la logica de delivery lee el directorio entero.
var HookCategories = []SkillCategory{
	{
		Name:      "hooks",
		Label:     "Claude Code Hooks",
		Exclusive: false,
		Options: []SkillOption{
			{
				Name:            "linting",
				Label:           "Linting (auto-format on Edit/Write)",
				TemplateDir:     "hooks/linting",
				TemplateMapping: nil,
			},
			{
				Name:            "security-guardrails",
				Label:           "Security Guardrails (block dangerous commands and sensitive files)",
				TemplateDir:     "hooks/security-guardrails",
				TemplateMapping: nil,
			},
			{
				Name:            "convention-enforcement",
				Label:           "Convention Enforcement (Conventional Commits, protected branches)",
				TemplateDir:     "hooks/convention-enforcement",
				TemplateMapping: nil,
			},
		},
	},
}

// HookCategoryNames retorna los nombres de las categorias registradas.
func HookCategoryNames() []string {
	names := make([]string, len(HookCategories))
	for i, c := range HookCategories {
		names[i] = c.Name
	}
	return names
}

// FindHookCategory busca una categoria de hooks por nombre.
func FindHookCategory(name string) (*SkillCategory, error) {
	for i := range HookCategories {
		if HookCategories[i].Name == name {
			return &HookCategories[i], nil
		}
	}
	return nil, fmt.Errorf("unknown hook category: %s", name)
}

// HookPresetNames retorna los nombres de los presets disponibles dentro de la categoria "hooks".
// Util para validacion en CLI/MCP sin necesidad de buscar la categoria primero.
func HookPresetNames() []string {
	if len(HookCategories) == 0 {
		return nil
	}
	opts := HookCategories[0].Options
	names := make([]string, len(opts))
	for i, o := range opts {
		names[i] = o.Name
	}
	return names
}
