package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	root "github.com/jorelcb/ai-context-generator"
	"github.com/jorelcb/ai-context-generator/internal/application/command"
	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/filesystem"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/llm"
	infratemplate "github.com/jorelcb/ai-context-generator/internal/infrastructure/template"
)

// skillsDefaultTemplateMapping maps default preset skill template files to guide names.
var skillsDefaultTemplateMapping = map[string]string{
	"ddd_entity.template":       "ddd_entity",
	"clean_arch_layer.template": "clean_arch_layer",
	"bdd_scenario.template":     "bdd_scenario",
	"cqrs_command.template":     "cqrs_command",
	"hexagonal_port.template":   "hexagonal_port",
}

// skillsNeutralTemplateMapping maps neutral preset skill template files to guide names.
var skillsNeutralTemplateMapping = map[string]string{
	"code_review.template":     "code_review",
	"test_strategy.template":   "test_strategy",
	"refactor_safely.template": "refactor_safely",
	"api_design.template":      "api_design",
}

// NewSkillsCmd creates the skills command
func NewSkillsCmd() *cobra.Command {
	var (
		preset string
		locale string
		target string
		model  string
		output string
	)

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Generate reusable AI agent skills (SKILL.md)",
		Long: `Generate reusable Agent Skills based on architectural presets.
Skills are SKILL.md files that teach AI coding agents how to approach
specific architectural and engineering tasks. They are cross-project
and can be installed globally for any AI agent ecosystem.

Presets:
  default  - DDD, Clean Architecture, BDD, CQRS, Hexagonal (default)
  neutral  - Code review, test strategy, refactoring, API design

Target ecosystems:
  claude       - Claude Code (default)
  codex        - Codex CLI (OpenAI)
  antigravity  - Antigravity IDE (Google)

Locales:
  en  - English (default)
  es  - Spanish

Requires ANTHROPIC_API_KEY (for Claude) or GEMINI_API_KEY (for Gemini) environment variable.

Examples:
  # Generate default preset skills for Claude Code
  ai-context-generator skills

  # Generate neutral skills for Codex
  ai-context-generator skills --preset neutral --target codex

  # Generate in Spanish with custom output
  ai-context-generator skills --locale es --output ./my-skills/

  # Generate with Gemini
  ai-context-generator skills --model gemini-3.1-pro-preview`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkills(preset, locale, target, model, output)
		},
	}

	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Template preset: default (DDD/Clean Architecture) or neutral")
	cmd.Flags().StringVar(&locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVar(&target, "target", "claude", "Target ecosystem: claude, codex, or antigravity")
	cmd.Flags().StringVarP(&model, "model", "m", "", "LLM model (default: claude-sonnet-4-6, or gemini-3.1-pro-preview)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory (default: output/skills/{preset}/)")

	return cmd
}

func runSkills(preset, locale, target, model, output string) error {
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

	// 3. Validate and resolve preset
	if !validPresets[preset] {
		preset = "default"
	}

	// 4. Select template mapping and load templates
	templateMapping := skillsDefaultTemplateMapping
	if preset == "neutral" {
		templateMapping = skillsNeutralTemplateMapping
	}

	templatePath := filepath.Join("templates", locale, "skills", preset)
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, templateMapping)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load skill templates: %w", err)
	}

	// 5. Initialize LLM provider
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// 6. Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// 7. Create command
	skillsCmd := command.NewGenerateSkillsCommand(provider, fileWriter, dirManager)

	// 8. Build config
	if output == "" {
		output = filepath.Join("output", "skills", preset)
	}
	config := &dto.SkillsConfig{
		Preset:     preset,
		Locale:     locale,
		Target:     target,
		Model:      model,
		OutputPath: output,
	}

	// 9. Show progress
	fmt.Printf("Generating agent skills\n")
	fmt.Printf("  Preset: %s\n", preset)
	fmt.Printf("  Target: %s\n", target)
	fmt.Printf("  Model: %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale: %s\n", locale)
	fmt.Printf("  Output: %s\n", output)
	fmt.Printf("  Skills: %d\n", len(guides))
	fmt.Println()
	fmt.Println("Generating skills via LLM API...")

	// 10. Execute
	result, err := skillsCmd.Execute(ctx, config, guides)
	if err != nil {
		return fmt.Errorf("skills generation failed: %w", err)
	}

	// 11. Show results
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
