package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	domain "github.com/jorelcb/codify/internal/domain/config"
	statedomain "github.com/jorelcb/codify/internal/domain/state"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	infrastate "github.com/jorelcb/codify/internal/infrastructure/state"
)

// codifyVersion se inyecta desde cli.Version en runtime; aquí se usa solo
// para popular state.json. Se importa indirectamente — el binario está
// configurado con build flags. Si no está disponible, se usa "dev".
var codifyVersion = "dev"

// NewInitCmd construye `codify init`, el smart entry point a nivel proyecto.
//
// Flujo:
//   1. Pregunta si el proyecto es nuevo o existente
//   2. Recolecta preset/language/locale (con override de defaults globales)
//   3. Si "nuevo": invoca generate (con descripción inline o desde archivo)
//      Si "existente": invoca analyze
//   4. Persiste .codify/config.yml + .codify/state.json
//
// Skills/workflows/hooks de proyecto se delegan a sus comandos respectivos —
// init imprime sugerencias de comandos para correrlos después, en lugar de
// re-implementar la lógica completa. Esto mantiene init delgado y respeta
// la decisión de coexistencia documentada en ADR-007.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a project with Codify (interactive)",
		Long: `Bootstrap a project with Codify by creating .codify/config.yml and
.codify/state.json, then generating context from your description (new project)
or by scanning the existing repo (existing project).

Skills, workflows, and hooks are delivered via their dedicated commands —
'codify init' prints the recommended next steps for those instead of bundling
them into a single mega-command. This keeps each command focused and aligned
with ADR-007 (commands coexist; init is the smart entry point, not a replacement).`,
		RunE: runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	if !isInteractive() {
		return fmt.Errorf("codify init requires an interactive TTY; pass flags to 'codify generate' or 'codify analyze' for non-interactive use")
	}

	repo := infraconfig.NewRepository()
	stateRepo := infrastate.NewRepository()

	// Load effective config (builtin + user) para usarlo como defaults
	effective, err := repo.LoadEffective()
	if err != nil {
		return fmt.Errorf("load effective config: %w", err)
	}

	fmt.Println("Codify Project Bootstrap")
	fmt.Println("════════════════════════")

	kind, err := promptSelect("Is this a new project or an existing one?", []selectOption{
		{"new — describe the project, generate context", "new"},
		{"existing — scan the codebase, generate context from what's there", "existing"},
	}, "new")
	if err != nil {
		return err
	}

	var projectName string
	var description string
	var fromFile string

	switch kind {
	case "new":
		projectName, err = promptInput("Project name", "")
		if err != nil {
			return err
		}
		if projectName == "" {
			return fmt.Errorf("project name is required for a new project")
		}

		descSource, err := promptSelect("How do you want to provide the description?", []selectOption{
			{"inline (prompt now)", "inline"},
			{"file (path to a file with the description)", "file"},
		}, "inline")
		if err != nil {
			return err
		}
		switch descSource {
		case "inline":
			description, err = promptInput("Project description", "")
			if err != nil {
				return err
			}
			if description == "" {
				return fmt.Errorf("description is required for a new project")
			}
		case "file":
			fromFile, err = promptInput("Path to description file", "")
			if err != nil {
				return err
			}
			if fromFile == "" {
				return fmt.Errorf("file path is required when 'file' source selected")
			}
		}
	case "existing":
		// Para proyectos existentes, derivamos el nombre del cwd a menos que el usuario quiera override
		cwd, _ := os.Getwd()
		defaultName := filepath.Base(cwd)
		projectName, err = promptInput(fmt.Sprintf("Project name (auto-detected: %s)", defaultName), defaultName)
		if err != nil {
			return err
		}
		if projectName == "" {
			projectName = defaultName
		}
	}

	// Preset (override del default global)
	preset := effective.Preset
	overridePreset, err := promptConfirm(fmt.Sprintf("Architectural preset is '%s' (from global default). Override?", preset), false)
	if err != nil {
		return err
	}
	if overridePreset {
		preset, err = promptPreset()
		if err != nil {
			return err
		}
	}

	// Language opcional
	language, err := promptLanguage()
	if err != nil {
		return err
	}

	// Locale (default desde effective)
	locale := effective.Locale
	if locale == "" {
		locale = "en"
	}

	// Persistir project config ANTES de invocar generate/analyze para que
	// esos comandos lo lean (futuro v1.22+: por ahora generate/analyze no leen,
	// pero el archivo queda escrito para próximos releases — task #23).
	projectCfg := domain.Config{
		Preset:      preset,
		Locale:      locale,
		Language:    language,
		ProjectName: projectName,
	}
	projectCfgPath, err := infraconfig.ProjectConfigPath()
	if err != nil {
		return err
	}
	if err := repo.Save(projectCfgPath, projectCfg); err != nil {
		return fmt.Errorf("save project config: %w", err)
	}
	fmt.Printf("\n✓ Wrote %s\n", projectCfgPath)

	// Output dir para artefactos generados
	outputDir, err := promptInput("Output directory", ".")
	if err != nil {
		return err
	}
	if outputDir == "" {
		outputDir = "."
	}

	// Resolve model si hay API keys (skip silencioso si no)
	model := effective.Model
	if model == "" {
		if m, err := promptModel(); err == nil {
			model = m
		}
	}

	// Disparar generate/analyze según rama
	fmt.Println()
	fmt.Println("--- Bootstrapping context ---")
	fmt.Println()

	switch kind {
	case "new":
		// Si fromFile está seteado, leer su contenido como descripción
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("read description file: %w", err)
			}
			description = string(data)
		}
		if err := runGenerate(projectName, description, language, "", "", model, preset, locale, outputDir); err != nil {
			return fmt.Errorf("generate failed: %w", err)
		}
	case "existing":
		// Para "existing", scaneamos cwd
		if err := runAnalyzeFromInit(".", projectName, language, model, preset, locale, outputDir); err != nil {
			return fmt.Errorf("analyze failed: %w", err)
		}
	}

	// Persistir state.json (sin hashes ni signals — eso lo agrega v1.23)
	st := statedomain.New()
	st.CodifyVersion = codifyVersion
	st.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	st.GeneratedBy = "init"
	st.Project = statedomain.ProjectInfo{
		Name:     projectName,
		Preset:   preset,
		Language: language,
		Locale:   locale,
		Target:   effective.Target,
		Kind:     kind,
	}
	statePath, err := infraconfig.ProjectStatePath()
	if err != nil {
		return err
	}
	if err := stateRepo.Save(statePath, st); err != nil {
		return fmt.Errorf("save state.json: %w", err)
	}
	fmt.Printf("\n✓ Wrote %s\n", statePath)

	// Recomendaciones de comandos siguientes (composición en lugar de mega-comando)
	fmt.Println()
	fmt.Println("Project bootstrapped. Recommended next steps:")
	fmt.Println("  codify skills      Install architecture/testing/conventions skills (project-scoped)")
	fmt.Println("  codify workflows   Install spec-driven-change / bug-fix / release-cycle workflows")
	fmt.Println("  codify hooks       Install Claude Code hook bundles (linting / security / conventions)")
	fmt.Println()
	fmt.Println("Lifecycle commands (codify check / update / audit / watch) arrive starting v1.23.")

	return nil
}

// runAnalyzeFromInit es un thin shim para invocar analyze desde init sin
// duplicar su lógica. Marca todos los campos como "explicit" para que
// runAnalyzeInteractive no re-prompte por valores que init ya recolectó.
func runAnalyzeFromInit(projectPath, projectName, language, model, preset, locale, output string) error {
	p := analyzeParams{
		name:     projectName,
		language: language,
		model:    model,
		preset:   preset,
		locale:   locale,
		output:   output,
	}
	explicit := map[string]bool{
		"name":     projectName != "",
		"language": language != "",
		"model":    model != "",
		"preset":   true,
		"locale":   true,
		"output":   true,
	}
	return runAnalyzeInteractive(projectPath, p, explicit)
}
