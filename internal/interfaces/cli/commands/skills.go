package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// NewSkillsCmd creates the skills command
func NewSkillsCmd() *cobra.Command {
	var (
		category string
		preset   string
		locale   string
		target   string
		model    string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Generate reusable AI agent skills (SKILL.md)",
		Long: `Generate reusable Agent Skills based on skill categories and presets.
Skills are SKILL.md files that teach AI coding agents how to approach
specific architectural and engineering tasks. They are cross-project
and can be installed globally for any AI agent ecosystem.

Categories:
  architecture  - Architecture patterns and best practices
  workflow      - Development workflow automation

Presets:
  architecture:
    clean    - DDD, Clean Architecture, BDD, CQRS, Hexagonal
    neutral  - Code review, test strategy, refactoring, API design
  workflow:
    conventional-commit   - Conventional Commits spec
    semantic-versioning   - Semantic Versioning spec
    all                   - All workflow skills

When run without --category, an interactive menu is displayed.

Target ecosystems:
  claude       - Claude Code → .claude/skills/ (default)
  codex        - Codex CLI (OpenAI) → .agents/skills/
  antigravity  - Antigravity (Google) → .agents/skills/

Examples:
  # Interactive mode (select category and preset from menu)
  codify skills

  # Non-interactive: architecture skills
  codify skills --category architecture --preset clean

  # Non-interactive: all workflow skills
  codify skills --category workflow --preset all

  # Generate for Codex ecosystem
  codify skills --category architecture --preset neutral --target codex`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkills(category, preset, locale, target, model, output)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Skill category: architecture, workflow")
	cmd.Flags().StringVarP(&preset, "preset", "p", "", "Preset within category (or 'all' if supported)")
	cmd.Flags().StringVar(&locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVar(&target, "target", "claude", "Target ecosystem: claude, codex, or antigravity")
	cmd.Flags().StringVarP(&model, "model", "m", "", "LLM model (default: claude-sonnet-4-6, or gemini-3.1-pro-preview)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory (default: ecosystem-specific, e.g. .claude/skills/)")

	return cmd
}

func runSkills(categoryName, preset, locale, target, model, output string) error {
	ctx := context.Background()

	// 1. Resolve API key
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// 2. Validate target
	if !dto.ValidTargets[target] {
		return fmt.Errorf("invalid target: %s (available: claude, codex, antigravity)", target)
	}

	// 3. Resolve category and preset (interactive or flags)
	cat, selectedPreset, err := resolveSelection(categoryName, preset)
	if err != nil {
		return err
	}

	// 4. Resolve templates from catalog
	selection, err := cat.Resolve(selectedPreset)
	if err != nil {
		return err
	}

	// 5. Load templates
	templatePath := filepath.Join("templates", locale, "skills", selection.TemplateDir)
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, selection.TemplateMapping)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load skill templates: %w", err)
	}

	// 6. Initialize LLM provider
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// 7. Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// 8. Create command
	skillsCmd := command.NewGenerateSkillsCommand(provider, fileWriter, dirManager)

	// 9. Build config
	if output == "" {
		output = defaultSkillsPath(target)
	}
	config := &dto.SkillsConfig{
		Category:   cat.Name,
		Preset:     selectedPreset,
		Locale:     locale,
		Target:     target,
		Model:      model,
		OutputPath: output,
	}

	// 10. Show progress
	fmt.Printf("Generating agent skills\n")
	fmt.Printf("  Category: %s\n", cat.Name)
	fmt.Printf("  Preset: %s\n", selectedPreset)
	fmt.Printf("  Target: %s\n", target)
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale: %s\n", locale)
	fmt.Printf("  Output: %s\n", output)
	fmt.Printf("  Skills: %d\n", len(guides))
	fmt.Println()
	fmt.Println("Generating skills via LLM API...")

	// 11. Execute
	result, err := skillsCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("skills generation failed: %w", err)
	}

	// 12. Show results
	fmt.Println()
	fmt.Println("Agent skills generated successfully!")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Printf("  Model: %s\n", result.Model)
	fmt.Printf("  Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut)
	fmt.Println()
	fmt.Println("Generated skills:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

// resolveSelection determina categoría y preset, interactivamente si es necesario.
func resolveSelection(categoryName, preset string) (*catalog.SkillCategory, string, error) {
	// Si ambos flags están presentes, usar directo
	if categoryName != "" && preset != "" {
		cat, err := catalog.FindCategory(categoryName)
		if err != nil {
			return nil, "", err
		}
		return cat, preset, nil
	}

	// Si no hay TTY, requerir flags
	if !isInteractive() {
		return nil, "", fmt.Errorf("interactive mode requires a terminal; use --category and --preset flags")
	}

	// Menú interactivo nivel 1: seleccionar categoría
	if categoryName == "" {
		categoryName = promptCategory()
		if categoryName == "" {
			return nil, "", fmt.Errorf("no category selected")
		}
	}

	cat, err := catalog.FindCategory(categoryName)
	if err != nil {
		return nil, "", err
	}

	// Menú interactivo nivel 2: seleccionar preset
	if preset == "" {
		preset = promptPreset(cat)
		if preset == "" {
			return nil, "", fmt.Errorf("no preset selected")
		}
	}

	return cat, preset, nil
}

// promptCategory muestra el menú interactivo de categorías.
func promptCategory() string {
	options := make([]huh.Option[string], len(catalog.Categories))
	for i, c := range catalog.Categories {
		options[i] = huh.NewOption(c.Label, c.Name)
	}

	var selected string
	err := huh.NewSelect[string]().
		Title("Select skill category").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return ""
	}
	return selected
}

// promptPreset muestra el menú interactivo de sub-opciones dentro de una categoría.
func promptPreset(cat *catalog.SkillCategory) string {
	options := make([]huh.Option[string], 0, len(cat.Options)+1)
	for _, o := range cat.Options {
		options = append(options, huh.NewOption(o.Label, o.Name))
	}
	// Agregar "All" solo si la categoría lo permite
	if !cat.Exclusive {
		options = append(options, huh.NewOption("All", "all"))
	}

	var selected string
	err := huh.NewSelect[string]().
		Title(fmt.Sprintf("Select %s preset", cat.Label)).
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return ""
	}
	return selected
}

// isInteractive verifica si stdout es un terminal.
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// defaultSkillsPath returns the ecosystem-specific default skills directory.
func defaultSkillsPath(target string) string {
	switch target {
	case "codex", "antigravity":
		return filepath.Join(".agents", "skills")
	default: // claude
		return filepath.Join(".claude", "skills")
	}
}
