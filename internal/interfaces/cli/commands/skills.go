package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// skillsConfig agrupa todos los parámetros del comando skills.
type skillsConfig struct {
	category string
	preset   string
	target   string
	locale   string
	model    string
	output   string
}

// NewSkillsCmd creates the skills command
func NewSkillsCmd() *cobra.Command {
	var cfg skillsConfig

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

When run without flags, an interactive menu is displayed.

Target ecosystems:
  claude       - Claude Code → .claude/skills/ (default)
  codex        - Codex CLI (OpenAI) → .agents/skills/
  antigravity  - Antigravity (Google) → .agents/skills/

Examples:
  # Interactive mode (guided selection)
  codify skills

  # Non-interactive: architecture skills
  codify skills --category architecture --preset clean

  # Non-interactive: all workflow skills
  codify skills --category workflow --preset all

  # Generate for Codex ecosystem
  codify skills --category architecture --preset neutral --target codex`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Detectar qué flags fueron proporcionados explícitamente
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})
			return runSkills(cfg, explicit)
		},
	}

	cmd.Flags().StringVarP(&cfg.category, "category", "c", "", "Skill category: architecture, workflow")
	cmd.Flags().StringVarP(&cfg.preset, "preset", "p", "", "Preset within category (or 'all' if supported)")
	cmd.Flags().StringVar(&cfg.target, "target", "claude", "Target ecosystem: claude, codex, or antigravity")
	cmd.Flags().StringVar(&cfg.locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&cfg.model, "model", "m", "", "LLM model (default: auto-detected from API key)")
	cmd.Flags().StringVarP(&cfg.output, "output", "o", "", "Output directory (default: ecosystem-specific)")

	return cmd
}

func runSkills(cfg skillsConfig, explicit map[string]bool) error {
	ctx := context.Background()
	interactive := isInteractive()

	// 1. Resolve category and preset (interactive or flags)
	cat, preset, err := resolveSelection(cfg.category, cfg.preset)
	if err != nil {
		return err
	}

	// 2. Resolve remaining config interactively if needed
	target := cfg.target
	if !explicit["target"] && interactive {
		target, err = promptSelect("Select target ecosystem", []selectOption{
			{"Claude Code → .claude/skills/", "claude"},
			{"Codex CLI → .agents/skills/", "codex"},
			{"Antigravity → .agents/skills/", "antigravity"},
		}, "claude")
		if err != nil {
			return err
		}
	}
	if !dto.ValidTargets[target] {
		return fmt.Errorf("invalid target: %s (available: claude, codex, antigravity)", target)
	}

	locale := cfg.locale
	if !explicit["locale"] && interactive {
		locale, err = promptSelect("Select language", []selectOption{
			{"English", "en"},
			{"Spanish", "es"},
		}, "en")
		if err != nil {
			return err
		}
	}

	model := cfg.model
	if !explicit["model"] && interactive {
		model, err = promptModel()
		if err != nil {
			return err
		}
	}

	output := cfg.output
	if output == "" {
		output = defaultSkillsPath(target)
	}
	if !explicit["output"] && interactive {
		output, err = promptInput("Output directory", output)
		if err != nil {
			return err
		}
	}

	// 3. Resolve API key (now that model is known)
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// 4. Resolve templates from catalog
	selection, err := cat.Resolve(preset)
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
	config := &dto.SkillsConfig{
		Category:   cat.Name,
		Preset:     preset,
		Locale:     locale,
		Target:     target,
		Model:      model,
		OutputPath: output,
	}

	// 10. Show progress
	fmt.Println()
	fmt.Printf("Generating agent skills\n")
	fmt.Printf("  Category: %s\n", cat.Name)
	fmt.Printf("  Preset: %s\n", preset)
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

// --- Resolución de selección (categoría + preset) ---

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
		options := make([]selectOption, len(catalog.Categories))
		for i, c := range catalog.Categories {
			options[i] = selectOption{c.Label, c.Name}
		}
		var err error
		categoryName, err = promptSelect("Select skill category", options, "")
		if err != nil {
			return nil, "", err
		}
	}

	cat, err := catalog.FindCategory(categoryName)
	if err != nil {
		return nil, "", err
	}

	// Menú interactivo nivel 2: seleccionar preset
	if preset == "" {
		options := make([]selectOption, 0, len(cat.Options)+1)
		for _, o := range cat.Options {
			options = append(options, selectOption{o.Label, o.Name})
		}
		if !cat.Exclusive {
			options = append(options, selectOption{"All", "all"})
		}

		preset, err = promptSelect(fmt.Sprintf("Select %s preset", cat.Label), options, "")
		if err != nil {
			return nil, "", err
		}
	}

	return cat, preset, nil
}

// --- Prompts interactivos genéricos ---

type selectOption struct {
	Label string
	Value string
}

// promptSelect muestra un selector interactivo con opciones y valor por defecto.
func promptSelect(title string, options []selectOption, defaultVal string) (string, error) {
	huhOpts := make([]huh.Option[string], len(options))
	for i, o := range options {
		huhOpts[i] = huh.NewOption(o.Label, o.Value)
	}

	selected := defaultVal
	err := huh.NewSelect[string]().
		Title(title).
		Options(huhOpts...).
		Value(&selected).
		Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled")
	}
	return selected, nil
}

// promptInput muestra un campo de texto interactivo con valor por defecto.
func promptInput(title, defaultVal string) (string, error) {
	value := defaultVal
	err := huh.NewInput().
		Title(title).
		Value(&value).
		Run()
	if err != nil {
		return "", fmt.Errorf("input cancelled")
	}
	if value == "" {
		return defaultVal, nil
	}
	return value, nil
}

// promptModel muestra un selector de modelo basado en las API keys disponibles.
func promptModel() (string, error) {
	var options []selectOption
	hasAnthropic := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasGemini := os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GOOGLE_API_KEY") != ""

	if hasAnthropic {
		options = append(options, selectOption{"Claude Sonnet 4.6 (Anthropic)", "claude-sonnet-4-6"})
		options = append(options, selectOption{"Claude Opus 4.6 (Anthropic)", "claude-opus-4-6"})
	}
	if hasGemini {
		options = append(options, selectOption{"Gemini 3.1 Pro Preview (Google)", "gemini-3.1-pro-preview"})
	}

	if len(options) == 0 {
		// Sin API keys detectadas, mostrar todos y dejar que falle después
		options = []selectOption{
			{"Claude Sonnet 4.6 (Anthropic)", "claude-sonnet-4-6"},
			{"Claude Opus 4.6 (Anthropic)", "claude-opus-4-6"},
			{"Gemini 3.1 Pro Preview (Google)", "gemini-3.1-pro-preview"},
		}
	}

	// Si solo hay un proveedor disponible con un modelo, usar directo
	if len(options) == 1 {
		return options[0].Value, nil
	}

	return promptSelect("Select LLM model", options, options[0].Value)
}

// isInteractive verifica si stdin/stdout son terminales.
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
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
