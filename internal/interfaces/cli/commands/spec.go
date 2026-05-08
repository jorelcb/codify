package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	domainservice "github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/config"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	"github.com/jorelcb/codify/internal/infrastructure/sdd"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

// resolveSpecStandard aplica la precedencia documentada en ADR-0011 para
// elegir el SpecStandard activo:
//
//  1. flagValue — `--sdd-standard` en la línea de comandos
//  2. project config (.codify/config.yml > sdd_standard)
//  3. user config (~/.codify/config.yml > sdd_standard)
//  4. built-in default (openspec)
//
// El registry resuelve los IDs contra los adapters disponibles. Si el ID en
// cualquiera de los niveles no existe, falla con error explícito (preferimos
// fallar a fallback silencioso a un standard que el usuario no eligió).
//
// Las capas user+project ya vienen mergeadas en effective.SDDStandard porque
// el repository.LoadEffective hace ese merge — basta con leer ese campo y
// dejarlo competir contra el flag y el default.
func resolveSpecStandard(flagValue string) (domainservice.SpecStandard, error) {
	registry := sdd.NewDefaultRegistry()

	repo := config.NewRepository()
	effective, err := repo.LoadEffective()
	if err != nil {
		return nil, fmt.Errorf("load effective config: %w", err)
	}

	// effective.SDDStandard ya combina user y project — Resolve solo necesita
	// el flag explícito y el merged value, en ese orden de prioridad.
	return registry.Resolve(flagValue, effective.SDDStandard, "")
}

// specTemplateMapping construye el mapping {filename → guideName} a partir
// de los artifacts del estándar. Convención: cada artifact con GuideName X
// se carga desde el template "X.template".
func specTemplateMapping(std domainservice.SpecStandard) map[string]string {
	artifacts := std.BootstrapArtifacts()
	m := make(map[string]string, len(artifacts))
	for _, a := range artifacts {
		m[a.GuideName+".template"] = a.GuideName
	}
	return m
}

// applyStandardOutputNames anota cada TemplateGuide con el OutputFileName
// que el SpecStandard activo dicta para ese guide. Esto permite que el
// mismo guide name ("spec") emita "SPEC.md" en OpenSpec y "spec.md" en
// Spec-Kit sin tocar el global fileOutputNames de prompt_builder.
//
// Si un guide cargado no aparece en BootstrapArtifacts (escenario raro
// pero posible si los templates se renombran y los adapters no se
// actualizan), se deja OutputFileName vacío para que GuideOutputName caiga
// al fallback global — el mismo comportamiento que antes del refactor.
func applyStandardOutputNames(guides []domainservice.TemplateGuide, std domainservice.SpecStandard) []domainservice.TemplateGuide {
	byGuide := make(map[string]string, len(std.BootstrapArtifacts()))
	for _, a := range std.BootstrapArtifacts() {
		byGuide[a.GuideName] = a.FileName
	}
	out := make([]domainservice.TemplateGuide, 0, len(guides))
	for _, g := range guides {
		if name, ok := byGuide[g.Name]; ok {
			g.OutputFileName = name
		}
		out = append(out, g)
	}
	return out
}

// slugifyFeatureID convierte un projectName en un slug seguro para
// filesystem (lowercase, ASCII, guiones para espacios y caracteres
// especiales). Usado solo cuando el SpecStandard activo tiene
// LayoutFeatureGrouped — en LayoutFlat el feature-id no se usa.
//
// Heurística mínima: si el nombre ya es lowercase + alfanuméricos +
// guiones, se devuelve tal cual. Si tiene mayúsculas o caracteres no-safe,
// se transforma. No hace transliteración Unicode — proyectos con nombres
// que dependan de eso pueden pasar --feature-id explícito en el futuro.
func slugifyFeatureID(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			prevDash = false
		case r == '-' || r == '_':
			b.WriteRune('-')
			prevDash = true
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "feature"
	}
	return out
}

// specParams groups all parameters for the spec command.
type specParams struct {
	fromContext string
	output      string
	model       string
	locale      string
	sddStandard string // ADR-0011 — empty = use config/default precedence
}

// NewSpecCmd creates the spec command
func NewSpecCmd() *cobra.Command {
	var p specParams

	cmd := &cobra.Command{
		Use:   "spec <project-name>",
		Short: "Generate SDD specification files from existing context",
		Long: `Generate spec-driven development files from previously generated context.

The set of files and their layout depend on the active SDD standard. Default
is OpenSpec (4 files at the root of specs/):
  - CONSTITUTION.md - Project DNA: stack, conventions, constraints, principles
  - SPEC.md         - Feature specifications with acceptance criteria
  - PLAN.md         - Technical design and architecture decisions
  - TASKS.md        - Implementation task breakdown with dependencies

Spec-Kit is also supported and produces a different file set under
specs/<feature-id>/ (lowercase file names, per-feature directory). Pick the
active standard with --sdd-standard (precedence: flag > .codify/config.yml >
~/.codify/config.yml > built-in default 'openspec'). For Spec-Kit, the
feature-id defaults to a slugified projectName.

See ADR-0011 and 'docs/command-reference.md' for details.

Requires existing context generated by the 'generate' command.
Requires ANTHROPIC_API_KEY (for Claude) or GEMINI_API_KEY (for Gemini) environment variable.

When run in a terminal, interactive menus guide you through options not provided via flags.

Examples:
  # Generate specs from context in current directory
  codify spec my-api --from-context .

  # Generate specs to a custom directory
  codify spec my-api --from-context ./context-dir/ --output ./specs-dir/

  # With custom model
  codify spec my-api --from-context . --model claude-sonnet-4-6`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := make(map[string]bool)
			cmd.Flags().Visit(func(f *pflag.Flag) {
				explicit[f.Name] = true
			})

			return runSpecInteractive(args[0], p, explicit)
		},
	}

	cmd.Flags().StringVar(&p.fromContext, "from-context", "", "Path to existing context directory (required; prompted in interactive mode)")
	cmd.Flags().StringVarP(&p.output, "output", "o", "", "Output directory (default: same as --from-context)")
	cmd.Flags().StringVarP(&p.model, "model", "m", "", "LLM model (default: claude-sonnet-4-6, or gemini-3.1-pro-preview)")
	cmd.Flags().StringVar(&p.locale, "locale", defaultLocale, "Output language: en (English) or es (Spanish)")
	cmd.Flags().StringVar(&p.sddStandard, "sdd-standard", "", "SDD standard: openspec (default) or spec-kit. Overrides project/user config.")

	return cmd
}

func runSpecInteractive(projectName string, p specParams, explicit map[string]bool) error {
	interactive := isInteractive()
	var err error

	// 1. Resolve from-context (required, but prompt in interactive mode).
	if p.fromContext == "" && interactive {
		p.fromContext, err = promptInput("Path to existing context directory", ".")
		if err != nil {
			return err
		}
	}
	if p.fromContext == "" {
		return fmt.Errorf("--from-context is required (use the flag or run interactively)")
	}
	// Best-effort heuristic: warn if the directory does not contain AGENTS.md or
	// CONTEXT.md, since spec generation has nothing to anchor on otherwise.
	hasContext := false
	for _, name := range []string{"AGENTS.md", "CONTEXT.md"} {
		if _, statErr := os.Stat(filepath.Join(p.fromContext, name)); statErr == nil {
			hasContext = true
			break
		}
	}
	if !hasContext {
		fmt.Fprintf(os.Stderr, "warning: %s does not appear to contain AGENTS.md or CONTEXT.md — spec generation may produce [DEFINE] markers\n", p.fromContext)
	}

	// 2. Resolve model
	if !explicit["model"] && interactive {
		p.model, err = promptModel()
		if err != nil {
			return err
		}
	}

	// 3. Resolve locale
	if !explicit["locale"] && interactive {
		p.locale, err = promptLocale()
		if err != nil {
			return err
		}
	}

	// 4. Resolve output
	if !explicit["output"] && interactive {
		defaultOutput := p.fromContext
		if defaultOutput == "" {
			defaultOutput = "."
		}
		p.output, err = promptInput("Output directory", defaultOutput)
		if err != nil {
			return err
		}
	}

	return runSpec(projectName, p.fromContext, p.output, p.model, p.locale, p.sddStandard)
}

func runSpec(projectName, fromContext, output, model, locale, sddStandardFlag string) error {
	if output == "" {
		output = fromContext
	}
	ctx := context.Background()

	// 1. Resolve API key for the selected provider
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return err
	}

	// 2. Read existing context files
	contextReader := filesystem.NewContextReader()
	existingContext, err := contextReader.ReadExistingContext(fromContext)
	if err != nil {
		return fmt.Errorf("failed to read existing context: %w", err)
	}

	// 3. Resolve active SDD standard via precedence (flag > project > user > default)
	//    and load its spec templates from templates/{locale}/sdd/{TemplateDir}/spec/.
	standard, err := resolveSpecStandard(sddStandardFlag)
	if err != nil {
		return err
	}
	templatePath := filepath.Join("templates", locale, "sdd", standard.TemplateDir(), "spec")
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(root.TemplatesFS, templatePath, specTemplateMapping(standard))
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load spec templates for standard %q: %w", standard.ID(), err)
	}

	// 3b. Augment guides with the adapter's per-standard output file names.
	//     Required because the same guide name (e.g., "spec") maps to
	//     "SPEC.md" in OpenSpec but "spec.md" in Spec-Kit.
	guides = applyStandardOutputNames(guides, standard)

	// 4. Initialize LLM provider
	provider, err := llm.NewProvider(ctx, model, apiKey, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// 5. Initialize infrastructure
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	// 6. Create command
	specCmd := command.NewGenerateSpecCommand(provider, fileWriter, dirManager)

	// 7. Build config
	//    - For LayoutFlat (OpenSpec) FeatureID is unused — output goes
	//      to <output>/specs/<file>.md.
	//    - For LayoutFeatureGrouped (Spec-Kit) FeatureID is the subdir
	//      under specs/. Default is the project name, slugified to a
	//      filesystem-safe form.
	featureID := slugifyFeatureID(projectName)
	config := &dto.SpecConfig{
		ProjectName:     projectName,
		FromContextPath: fromContext,
		OutputPath:      output,
		Model:           model,
		Locale:          locale,
		FeatureID:       featureID,
		Layout:          standard.OutputLayout(),
		StandardID:      standard.ID(),
		StandardHints:   standard.SystemPromptHints(locale),
	}

	// 8. Show progress
	fmt.Printf("Generating specs for: %s\n", projectName)
	fmt.Printf("  From context:  %s\n", fromContext)
	fmt.Printf("  Model:         %s\n", llm.DefaultModel(model))
	fmt.Printf("  Locale:        %s\n", locale)
	fmt.Printf("  SDD standard:  %s\n", standard.DisplayName())
	if standard.OutputLayout() == domainservice.LayoutFeatureGrouped {
		fmt.Printf("  Feature ID:    %s\n", featureID)
	}
	fmt.Println()
	fmt.Println("Generating spec files via LLM API...")

	// 9. Execute
	result, err := specCmd.Execute(ctx, config, existingContext, guides)
	if err != nil {
		return fmt.Errorf("spec generation failed: %w", err)
	}

	// 10. Update AGENTS.md with specs references
	if err := updateAgentsWithSpecsRef(output, locale, standard, featureID); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update AGENTS.md with specs reference: %v\n", err)
	}

	// 11. Show results
	fmt.Println()
	fmt.Println("Spec files generated successfully!")
	fmt.Printf("  Output: %s\n", result.OutputPath)
	fmt.Printf("  Model: %s\n", result.Model)
	fmt.Printf("  Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut)
	fmt.Println()
	fmt.Println("Generated files:")
	for _, f := range result.GeneratedFiles {
		fmt.Printf("  - %s\n", f)
	}

	return nil
}

// updateAgentsWithSpecsRef appends specs file references to AGENTS.md if not
// already present. The list of files is derived from the active SpecStandard,
// so adding a new standard with different artifacts does not require editing
// this function.
func updateAgentsWithSpecsRef(fromContextPath string, locale string, standard domainservice.SpecStandard, featureID string) error {
	agentsPath := filepath.Join(fromContextPath, "AGENTS.md")

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		// AGENTS.md might not exist — skip silently
		return nil
	}

	// Check if specs reference already exists (idempotent)
	if strings.Contains(string(content), "specs/") {
		return nil
	}

	specsRef := buildSpecsReferenceSection(locale, standard, featureID)
	updated := string(content) + specsRef
	return os.WriteFile(agentsPath, []byte(updated), 0644)
}

// buildSpecsReferenceSection renders the markdown block that
// updateAgentsWithSpecsRef appends to AGENTS.md. Kept separate so it can
// be reused by MCP responses or other consumers that want to surface the
// same per-standard file list.
//
// El path al archivo respeta el layout del estándar:
//   - LayoutFlat: "specs/<file>"
//   - LayoutFeatureGrouped: "specs/<featureID>/<file>"
func buildSpecsReferenceSection(locale string, standard domainservice.SpecStandard, featureID string) string {
	header := "\n## Specifications\n\n"
	if locale == "es" {
		header = "\n## Especificaciones\n\n"
	}

	prefix := "specs/"
	if standard.OutputLayout() == domainservice.LayoutFeatureGrouped && featureID != "" {
		prefix = "specs/" + featureID + "/"
	}

	var sb strings.Builder
	sb.WriteString(header)
	for _, a := range standard.BootstrapArtifacts() {
		sb.WriteString("- `")
		sb.WriteString(prefix)
		sb.WriteString(a.FileName)
		sb.WriteString("`\n")
	}
	return sb.String()
}
