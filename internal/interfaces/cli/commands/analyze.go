package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jorelcb/codify/internal/infrastructure/scanner"
)

// analyzeParams groups all parameters for the analyze command.
type analyzeParams struct {
	name      string
	model     string
	preset    string
	locale    string
	language  string
	output    string
	withSpecs bool
}

// NewAnalyzeCmd creates the analyze command
func NewAnalyzeCmd() *cobra.Command {
	var p analyzeParams

	cmd := &cobra.Command{
		Use:   "analyze <project-path>",
		Short: "Analyze an existing project and generate AI context files",
		Long: `Scan an existing project directory, extract signals (language, dependencies,
structure, README, existing context), and generate optimized AI context files.

This is the recommended way to add AI context to an existing codebase.
If the project already has context files (AGENTS.md, CLAUDE.md), they will
be used to enrich the generated output.

When run in a terminal, interactive menus guide you through options not provided via flags.

Examples:
  # Analyze a Go project (interactive prompts for missing options)
  codify analyze ./my-go-api

  # With explicit name and language
  codify analyze ./my-go-api --name my-api --language go

  # Analyze and generate specs in one step
  codify analyze ./my-go-api --with-specs

  # Output to a specific directory
  codify analyze ./my-go-api --output ./docs/

  # In Spanish
  codify analyze ./my-go-api --locale es`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})

			return runAnalyzeInteractive(args[0], p, explicit)
		},
	}

	cmd.Flags().StringVarP(&p.name, "name", "n", "", "Project name (defaults to directory name)")
	cmd.Flags().StringVarP(&p.language, "language", "l", "", "Override detected language (activates idiomatic guides)")
	cmd.Flags().StringVarP(&p.model, "model", "m", "", "LLM model (default: claude-sonnet-4-6, or gemini-3.1-pro-preview)")
	cmd.Flags().StringVarP(&p.preset, "preset", "p", "neutral", "Template preset: neutral (default — no architectural opinion), clean-ddd (DDD + Clean Architecture), hexagonal (Ports & Adapters), event-driven (CQRS + Event Sourcing + Sagas)")
	cmd.Flags().StringVar(&p.locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&p.withSpecs, "with-specs", false, "Also generate SDD spec files after context generation")

	return cmd
}

func runAnalyzeInteractive(projectPath string, p analyzeParams, explicit map[string]bool) error {
	interactive := isInteractive()
	var err error

	// 0. Effective config fills in missing values that weren't explicit.
	//    See config_merge.go for the precedence rule.
	cfg := loadEffectiveConfig()
	applyConfigDefaults(&p.preset, cfg.Preset, explicit["preset"])
	applyConfigDefaults(&p.locale, cfg.Locale, explicit["locale"])
	applyConfigDefaults(&p.language, cfg.Language, explicit["language"])
	applyConfigDefaults(&p.model, cfg.Model, explicit["model"])

	// Resolve to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// 1. Scan project
	fmt.Printf("Scanning project: %s\n", absPath)
	s := scanner.NewProjectScanner()
	result, err := s.Scan(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// 2. Report scan results
	fmt.Println()
	fmt.Println("Scan results:")
	if result.Language != "" {
		fmt.Printf("  Language: %s\n", result.Language)
	}
	if result.Framework != "" {
		fmt.Printf("  Framework: %s\n", result.Framework)
	}
	if len(result.Dependencies) > 0 {
		fmt.Printf("  Dependencies: %d detected\n", len(result.Dependencies))
	}
	if len(result.ConfigSignals) > 0 {
		fmt.Printf("  Infrastructure: %s\n", joinMax(result.ConfigSignals, 3))
	}
	if len(result.ExistingContext) > 0 {
		fmt.Printf("  Existing context: %d files found\n", len(result.ExistingContext))
	}
	fmt.Println()

	// 3. Resolve name
	defaultName := filepath.Base(absPath)
	if p.name == "" {
		p.name = defaultName
	}
	if !explicit["name"] && interactive {
		p.name, err = promptInput("Project name", p.name)
		if err != nil {
			return err
		}
	}

	// 4. Resolve language — confirm auto-detected or prompt
	detectedLang := normalizeLanguageFlag(result.Language)
	if !explicit["language"] && interactive {
		if detectedLang != "" {
			useDetected, confirmErr := promptConfirm(fmt.Sprintf("Use detected language: %s?", result.Language), true)
			if confirmErr != nil {
				return confirmErr
			}
			if useDetected {
				p.language = detectedLang
			} else {
				p.language, err = promptLanguage()
				if err != nil {
					return err
				}
			}
		} else {
			p.language, err = promptLanguage()
			if err != nil {
				return err
			}
		}
	} else if p.language == "" && detectedLang != "" {
		p.language = detectedLang
	}

	// 5. Resolve preset
	if !explicit["preset"] && interactive {
		p.preset, err = promptPreset()
		if err != nil {
			return err
		}
	}

	// 6. Resolve locale
	if !explicit["locale"] && interactive {
		p.locale, err = promptLocale()
		if err != nil {
			return err
		}
	}

	// 7. Resolve model
	if !explicit["model"] && interactive {
		p.model, err = promptModel()
		if err != nil {
			return err
		}
	}

	// 8. Resolve output
	if p.output == "" {
		p.output = "."
	}
	if !explicit["output"] && interactive {
		p.output, err = promptInput("Output directory", p.output)
		if err != nil {
			return err
		}
	}

	// 9. Resolve with-specs
	if !explicit["with-specs"] && interactive {
		p.withSpecs, err = promptConfirm("Also generate SDD specs?", false)
		if err != nil {
			return err
		}
	}

	// 10. Generate — use scan description and delegate to generate pipeline with analyze mode
	description := result.FormatAsDescription()

	if err := runGenerateWithMode(p.name, description, p.language, "", "", p.model, p.preset, p.locale, p.output, "analyze"); err != nil {
		return err
	}

	// 11. Optionally generate specs
	if p.withSpecs {
		fmt.Println()
		fmt.Println("--- Generating specs from context ---")
		fmt.Println()
		// `analyze --with-specs` no expone --sdd-standard; respeta la
		// elección del config (project > user > default). Si se necesita
		// overridear desde analyze, se agrega flag dedicado y se pasa acá.
		return runSpec(p.name, p.output, p.output, p.model, p.locale, "")
	}

	return nil
}

// normalizeLanguageFlag maps detected language names to CLI flag values.
func normalizeLanguageFlag(detected string) string {
	mapping := map[string]string{
		"Go":                    "go",
		"JavaScript/TypeScript": "javascript",
		"Python":                "python",
		"Rust":                  "rust",
		"Java":                  "java",
		"Java/Kotlin":           "kotlin",
		"Ruby":                  "ruby",
		"Elixir":                "elixir",
		"PHP":                   "php",
		"Swift":                 "swift",
		"C#/.NET":               "csharp",
	}
	if flag, ok := mapping[detected]; ok {
		return flag
	}
	return ""
}

// joinMax joins up to max elements with ", " and adds "+N more" if truncated.
func joinMax(items []string, max int) string {
	if len(items) <= max {
		return joinStrings(items)
	}
	return joinStrings(items[:max]) + fmt.Sprintf(" (+%d more)", len(items)-max)
}

func joinStrings(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}
