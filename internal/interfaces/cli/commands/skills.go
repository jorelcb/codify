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
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// skillsParams agrupa todos los parámetros del comando skills.
type skillsParams struct {
	category       string
	preset         string
	mode           string
	target         string
	locale         string
	model          string
	output         string
	projectContext string
}

// NewSkillsCmd creates the skills command
func NewSkillsCmd() *cobra.Command {
	var p skillsParams

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Generate reusable AI agent skills (SKILL.md)",
		Long: `Generate reusable Agent Skills based on skill categories and presets.
Skills are SKILL.md files that teach AI coding agents how to approach
specific architectural and engineering tasks.

Modes:
  static        - Instant delivery from built-in catalog (no API key needed)
  personalized  - LLM-adapted skills tailored to your project (requires API key)

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

Examples:
  # Interactive mode (guided selection)
  codify skills

  # Static: instant delivery, no API key
  codify skills --category workflow --preset all --mode static

  # Personalized: LLM-adapted to your project
  codify skills --category architecture --preset clean --mode personalized --context "Go microservice with DDD"

  # Generate for Codex ecosystem
  codify skills --category architecture --preset neutral --target codex`,
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})
			return runSkills(p, explicit)
		},
	}

	cmd.Flags().StringVarP(&p.category, "category", "c", "", "Skill category: architecture, workflow")
	cmd.Flags().StringVarP(&p.preset, "preset", "p", "", "Preset within category (or 'all' if supported)")
	cmd.Flags().StringVar(&p.mode, "mode", "", "Generation mode: static (instant) or personalized (LLM)")
	cmd.Flags().StringVar(&p.target, "target", "claude", "Target ecosystem: claude, codex, or antigravity")
	cmd.Flags().StringVar(&p.locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&p.model, "model", "m", "", "LLM model (only for personalized mode)")
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Output directory (default: ecosystem-specific)")
	cmd.Flags().StringVar(&p.projectContext, "context", "", "Project context for personalized mode")

	return cmd
}

func runSkills(p skillsParams, explicit map[string]bool) error {
	ctx := context.Background()
	interactive := isInteractive()

	// 1. Resolve category and preset
	cat, preset, err := resolveSelection(p.category, p.preset)
	if err != nil {
		return err
	}

	// 2. Resolve mode
	mode := p.mode
	if !explicit["mode"] && interactive {
		mode, err = promptSelect("Select mode", []selectOption{
			{"Static (instant, no API key needed)", dto.SkillModeStatic},
			{"Personalized (LLM-adapted to your project)", dto.SkillModePersonalized},
		}, dto.SkillModeStatic)
		if err != nil {
			return err
		}
	}
	if mode == "" {
		mode = dto.SkillModeStatic
	}

	// 3. Resolve target
	target := p.target
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

	// 4. Resolve locale
	locale := p.locale
	if !explicit["locale"] && interactive {
		locale, err = promptSelect("Select language", []selectOption{
			{"English", "en"},
			{"Spanish", "es"},
		}, "en")
		if err != nil {
			return err
		}
	}

	// 5. Resolve output
	output := p.output
	if output == "" {
		output = defaultSkillsPath(target)
	}
	if !explicit["output"] && interactive {
		output, err = promptInput("Output directory", output)
		if err != nil {
			return err
		}
	}

	// 6. Resolve templates from catalog
	selection, err := cat.Resolve(preset)
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", locale, "skills", selection.TemplateDir)
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, selection.TemplateMapping)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load skill templates: %w", err)
	}

	// 7. Build config
	config := &dto.SkillsConfig{
		Category:   cat.Name,
		Preset:     preset,
		Mode:       mode,
		Locale:     locale,
		Target:     target,
		OutputPath: output,
	}

	// 8. Execute based on mode
	if mode == dto.SkillModeStatic {
		return executeStaticSkills(config, guides)
	}
	return executePersonalizedSkills(ctx, p, config, guides, explicit, interactive)
}

func executeStaticSkills(config *dto.SkillsConfig, guides []service.TemplateGuide) error {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	cmd := command.NewDeliverStaticSkillsCommand(fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Delivering agent skills (static)\n")
	fmt.Printf("  Category: %s\n", config.Category)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	fmt.Printf("  Skills: %d\n", len(guides))
	fmt.Println()

	result, err := cmd.Execute(config, guides)
	if err != nil {
		return fmt.Errorf("static skills delivery failed: %w", err)
	}

	fmt.Println("Agent skills delivered successfully!")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Println()
	fmt.Println("Delivered skills:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

func executePersonalizedSkills(ctx context.Context, p skillsParams, config *dto.SkillsConfig, guides []service.TemplateGuide, explicit map[string]bool, interactive bool) error {
	var err error

	// Resolve project context
	projectContext := p.projectContext
	if !explicit["context"] && interactive {
		projectContext, err = promptInput("Describe your project (language, architecture, domain, stack)", "")
		if err != nil {
			return err
		}
	}
	if projectContext == "" {
		return fmt.Errorf("personalized mode requires project context; use --context flag")
	}
	config.ProjectContext = projectContext

	// Resolve model
	model := p.model
	if !explicit["model"] && interactive {
		model, err = promptModel()
		if err != nil {
			return err
		}
	}
	config.Model = model

	// Resolve API key
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// Initialize LLM provider
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	skillsCmd := command.NewGenerateSkillsCommand(provider, fileWriter, dirManager)

	fmt.Println()
	fmt.Printf("Generating agent skills (personalized)\n")
	fmt.Printf("  Category: %s\n", config.Category)
	fmt.Printf("  Preset: %s\n", config.Preset)
	fmt.Printf("  Target: %s\n", config.Target)
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale: %s\n", config.Locale)
	fmt.Printf("  Output: %s\n", config.OutputPath)
	fmt.Printf("  Skills: %d\n", len(guides))
	fmt.Println()
	fmt.Println("Generating personalized skills via LLM API...")

	result, err := skillsCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("personalized skills generation failed: %w", err)
	}

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

func resolveSelection(categoryName, preset string) (*catalog.SkillCategory, string, error) {
	if categoryName != "" && preset != "" {
		cat, err := catalog.FindCategory(categoryName)
		if err != nil {
			return nil, "", err
		}
		return cat, preset, nil
	}

	if !isInteractive() {
		return nil, "", fmt.Errorf("interactive mode requires a terminal; use --category and --preset flags")
	}

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
		options = []selectOption{
			{"Claude Sonnet 4.6 (Anthropic)", "claude-sonnet-4-6"},
			{"Claude Opus 4.6 (Anthropic)", "claude-opus-4-6"},
			{"Gemini 3.1 Pro Preview (Google)", "gemini-3.1-pro-preview"},
		}
	}

	if len(options) == 1 {
		return options[0].Value, nil
	}

	return promptSelect("Select LLM model", options, options[0].Value)
}

func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

func defaultSkillsPath(target string) string {
	switch target {
	case "codex", "antigravity":
		return filepath.Join(".agents", "skills")
	default:
		return filepath.Join(".claude", "skills")
	}
}
