// Package catalog define el registro declarativo de categorías y opciones de skills.
package catalog

import (
	"fmt"
	"maps"
	"strings"
)

// SkillCategory representa una categoría de nivel 1 en el menú de skills.
type SkillCategory struct {
	Name      string        // identificador: "architecture", "workflow"
	Label     string        // display: "Architecture", "Workflow"
	Exclusive bool          // true = sub-opciones mutuamente excluyentes (sin "all")
	Options   []SkillOption // sub-opciones disponibles
}

// SkillOption representa una sub-opción dentro de una categoría.
type SkillOption struct {
	Name            string            // identificador: "clean", "neutral", "conventional-commit"
	Label           string            // display: "Clean (DDD, BDD, CQRS, Hexagonal)"
	TemplateDir     string            // directorio en templates/{locale}/skills/
	TemplateMapping map[string]string // nil = todos los templates del dir; map = solo los indicados
}

// ResolvedSelection es el resultado de resolver una selección del catálogo.
type ResolvedSelection struct {
	TemplateDir     string
	TemplateMapping map[string]string // nil = cargar todos los templates del directorio
}

// SkillMeta contiene metadata de un skill para generación de frontmatter estático.
type SkillMeta struct {
	Description string
	Triggers    []string // usado por ecosistemas que lo soportan (e.g. antigravity)
}

// SkillMetadata mapea guide names a su metadata para frontmatter.
var SkillMetadata = map[string]SkillMeta{
	// Architecture: clean
	"ddd_entity":       {Description: "Create domain entities following DDD principles", Triggers: []string{"entity", "aggregate", "value object", "domain model"}},
	"clean_arch_layer": {Description: "Implement features respecting Clean Architecture layers", Triggers: []string{"layer", "architecture", "dependency rule"}},
	"bdd_scenario":     {Description: "Write BDD scenarios in Gherkin with proper step definitions", Triggers: []string{"test", "scenario", "gherkin", "bdd"}},
	"cqrs_command":     {Description: "Implement CQRS commands and queries with proper separation", Triggers: []string{"command", "query", "cqrs", "handler"}},
	"hexagonal_port":   {Description: "Design ports and adapters following Hexagonal Architecture", Triggers: []string{"port", "adapter", "hexagonal", "dependency inversion"}},
	// Architecture: neutral
	"code_review":     {Description: "Perform structured, actionable code reviews", Triggers: []string{"review", "pull request", "code quality"}},
	"test_strategy":   {Description: "Design test strategies with proper test pyramid coverage", Triggers: []string{"test", "testing", "coverage", "test plan"}},
	"refactor_safely": {Description: "Refactor code safely with incremental transformations", Triggers: []string{"refactor", "cleanup", "tech debt"}},
	"api_design":      {Description: "Design REST APIs with consistent contracts and versioning", Triggers: []string{"api", "endpoint", "rest", "contract"}},
	// Workflow
	"conventional_commit":  {Description: "Write commit messages following Conventional Commits 1.0.0", Triggers: []string{"commit", "git commit", "conventional"}},
	"semantic_versioning":  {Description: "Determine version bumps following Semantic Versioning 2.0.0", Triggers: []string{"version", "release", "tag", "semver"}},
}

// GenerateFrontmatter genera YAML frontmatter para un skill según el ecosistema target.
func GenerateFrontmatter(guideName, target string) string {
	name := strings.ReplaceAll(guideName, "_", "-")
	meta, ok := SkillMetadata[guideName]
	if !ok {
		meta = SkillMeta{Description: fmt.Sprintf("Agent skill for %s", name)}
	}

	switch target {
	case "codex":
		return fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n", name, meta.Description)
	case "antigravity":
		var triggers strings.Builder
		for _, t := range meta.Triggers {
			triggers.WriteString(fmt.Sprintf("  - %s\n", t))
		}
		return fmt.Sprintf("---\nname: %s\ndescription: %s\ntriggers:\n%s---\n", name, meta.Description, triggers.String())
	default: // claude
		return fmt.Sprintf("---\nname: %s\ndescription: %s\nuser-invocable: true\n---\n", name, meta.Description)
	}
}

// Categories es el registro global de categorías de skills.
var Categories = []SkillCategory{
	{
		Name:      "architecture",
		Label:     "Architecture",
		Exclusive: true,
		Options: []SkillOption{
			{
				Name:        "clean",
				Label:       "Clean (DDD, BDD, CQRS, Hexagonal)",
				TemplateDir: "default",
				TemplateMapping: map[string]string{
					"ddd_entity.template":       "ddd_entity",
					"clean_arch_layer.template": "clean_arch_layer",
					"bdd_scenario.template":     "bdd_scenario",
					"cqrs_command.template":     "cqrs_command",
					"hexagonal_port.template":   "hexagonal_port",
				},
			},
			{
				Name:        "neutral",
				Label:       "Neutral (Code review, testing, API design, refactoring)",
				TemplateDir: "neutral",
				TemplateMapping: map[string]string{
					"code_review.template":     "code_review",
					"test_strategy.template":   "test_strategy",
					"refactor_safely.template": "refactor_safely",
					"api_design.template":      "api_design",
				},
			},
		},
	},
	{
		Name:      "workflow",
		Label:     "Workflow",
		Exclusive: false,
		Options: []SkillOption{
			{
				Name:        "conventional-commit",
				Label:       "Conventional Commits",
				TemplateDir: "workflow",
				TemplateMapping: map[string]string{
					"conventional_commit.template": "conventional_commit",
				},
			},
			{
				Name:        "semantic-versioning",
				Label:       "Semantic Versioning",
				TemplateDir: "workflow",
				TemplateMapping: map[string]string{
					"semantic_versioning.template": "semantic_versioning",
				},
			},
		},
	},
}

// CategoryNames devuelve los nombres de todas las categorías registradas.
func CategoryNames() []string {
	names := make([]string, len(Categories))
	for i, c := range Categories {
		names[i] = c.Name
	}
	return names
}

// FindCategory busca una categoría por nombre.
func FindCategory(name string) (*SkillCategory, error) {
	for i := range Categories {
		if Categories[i].Name == name {
			return &Categories[i], nil
		}
	}
	return nil, fmt.Errorf("unknown category: %s", name)
}

// Resolve resuelve la selección de una sub-opción (o "all") dentro de la categoría.
func (c *SkillCategory) Resolve(preset string) (*ResolvedSelection, error) {
	if preset == "all" {
		if c.Exclusive {
			return nil, fmt.Errorf("category %q does not support 'all' (options are mutually exclusive)", c.Name)
		}
		return c.resolveAll(), nil
	}

	for _, opt := range c.Options {
		if opt.Name == preset {
			return &ResolvedSelection{
				TemplateDir:     opt.TemplateDir,
				TemplateMapping: opt.TemplateMapping,
			}, nil
		}
	}
	return nil, fmt.Errorf("unknown preset %q in category %q", preset, c.Name)
}

// resolveAll combina todas las opciones de la categoría en una sola selección.
func (c *SkillCategory) resolveAll() *ResolvedSelection {
	merged := make(map[string]string)
	var dir string
	for _, opt := range c.Options {
		dir = opt.TemplateDir
		maps.Copy(merged, opt.TemplateMapping)
	}
	return &ResolvedSelection{
		TemplateDir:     dir,
		TemplateMapping: merged,
	}
}

// OptionNames devuelve los nombres de las sub-opciones de la categoría.
func (c *SkillCategory) OptionNames() []string {
	names := make([]string, len(c.Options))
	for i, o := range c.Options {
		names[i] = o.Name
	}
	return names
}

// OptionLabels devuelve los labels de las sub-opciones (para el menú interactivo).
func (c *SkillCategory) OptionLabels() []string {
	labels := make([]string, len(c.Options))
	for i, o := range c.Options {
		labels[i] = o.Label
	}
	return labels
}

// LegacyPresetMapping mapea presets legados (--preset flag antiguo) al nuevo modelo.
var LegacyPresetMapping = map[string][2]string{
	"default":  {"architecture", "clean"},
	"neutral":  {"architecture", "neutral"},
	"workflow": {"workflow", "all"},
}
