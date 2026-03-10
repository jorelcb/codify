package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jorelcb/ai-context-generator/internal/infrastructure/scanner"
)

// NewAnalyzeCmd creates the analyze command
func NewAnalyzeCmd() *cobra.Command {
	var (
		name      string
		model     string
		preset    string
		locale    string
		language  string
		withSpecs bool
	)

	cmd := &cobra.Command{
		Use:   "analyze <project-path>",
		Short: "Analyze an existing project and generate AI context files",
		Long: `Scan an existing project directory, extract signals (language, dependencies,
structure, README, existing context), and generate optimized AI context files.

This is the recommended way to add AI context to an existing codebase.
If the project already has context files (AGENTS.md, CLAUDE.md), they will
be used to enrich the generated output.

Examples:
  # Analyze a Go project
  ai-context-generator analyze ./my-go-api

  # With explicit name and language
  ai-context-generator analyze ./my-go-api --name my-api --language go

  # Analyze and generate specs in one step
  ai-context-generator analyze ./my-go-api --with-specs

  # In Spanish
  ai-context-generator analyze ./my-go-api --locale es`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			// Resolve to absolute path
			absPath, err := filepath.Abs(projectPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Default name to directory name
			if name == "" {
				name = filepath.Base(absPath)
			}

			return runAnalyze(absPath, name, language, model, preset, locale, withSpecs)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name (defaults to directory name)")
	cmd.Flags().StringVarP(&language, "language", "l", "", "Override detected language (activates idiomatic guides)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Claude model to use (default: claude-sonnet-4-6)")
	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Template preset: default or neutral")
	cmd.Flags().StringVar(&locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().BoolVar(&withSpecs, "with-specs", false, "Also generate SDD spec files after context generation")

	return cmd
}

func runAnalyze(projectPath, name, language, model, preset, locale string, withSpecs bool) error {
	// 1. Scan project
	fmt.Printf("Scanning project: %s\n", projectPath)
	s := scanner.NewProjectScanner()
	result, err := s.Scan(projectPath)
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

	// 3. Use detected language if not overridden
	if language == "" && result.Language != "" {
		language = normalizeLanguageFlag(result.Language)
	}

	// 4. Format scan as description and delegate to generate pipeline
	description := result.FormatAsDescription()

	if err := runGenerate(name, description, language, "", "", model, preset, locale); err != nil {
		return err
	}

	// 5. Optionally generate specs
	if withSpecs {
		outputPath := filepath.Join("output", name)
		fmt.Println()
		fmt.Println("--- Generating specs from context ---")
		fmt.Println()
		return runSpec(name, outputPath, model, locale)
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
