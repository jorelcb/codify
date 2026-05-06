package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
)

// MaybeAutoLaunchConfig implementa el auto-launch SOFT del wizard `codify config`
// la primera vez que el binario corre en contexto interactivo. Algoritmo
// completo en docs/adr/0007-bootstrap-commands-naming.md.
//
// Returns nil si todo está OK (ya sea porque corrió el wizard, porque el
// usuario lo declinó, o porque las precondiciones no se cumplen). Returns
// error solo si el wizard mismo falla.
//
// La función es idempotente y rápida en el camino feliz (config existe):
// hace stat() del archivo y retorna sin más.
func MaybeAutoLaunchConfig(cmd *cobra.Command) error {
	// 1. Si la flag/env opta-out, salir sin hacer nada.
	if os.Getenv("CODIFY_NO_AUTO_CONFIG") == "1" {
		return nil
	}

	// 2. Si el comando NO es interactive-suitable, salir.
	//    Lista enumera explícitamente los comandos donde tiene sentido
	//    proponer el wizard; resto pasa silencioso (--help, --version,
	//    serve, etc).
	if !isInteractiveSuitable(cmd) {
		return nil
	}

	// 3. Si ya existe ~/.codify/config.yml, salir.
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return nil // best-effort: si no podemos resolver home, no rompemos el flujo principal
	}
	if infraconfig.FileExists(userPath) {
		return nil
	}

	// 4. Si existe el marker file de skip-permanently, salir.
	markerPath, err := infraconfig.UserNoAutoConfigMarker()
	if err == nil && infraconfig.FileExists(markerPath) {
		return nil
	}

	// 5. Si stdin no es TTY interactivo, salir.
	if !isInteractive() {
		return nil
	}

	// 6. Prompt: Y/n/skip-permanently
	answer, err := promptSelect("Codify isn't configured globally yet. Run interactive setup now?", []selectOption{
		{"Yes — launch wizard now (recommended)", "yes"},
		{"No — use built-in defaults this run (will ask again next time)", "no"},
		{"Skip permanently — use built-in defaults always (creates ~/.codify/.no-auto-config)", "skip"},
	}, "yes")
	if err != nil {
		// promptSelect cancelado (Ctrl+C, etc.): tratarlo como "no este vez"
		return nil
	}

	switch answer {
	case "yes":
		repo := infraconfig.NewRepository()
		if err := runConfigWizard(repo, userPath); err != nil {
			return err
		}
		fmt.Println()
	case "no":
		// no-op
	case "skip":
		if err := infraconfig.EnsureUserConfigDir(); err != nil {
			return nil
		}
		// touch marker file
		f, err := os.OpenFile(markerPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
			fmt.Fprintf(os.Stderr, "✓ Created %s. Codify will not auto-prompt for setup again.\nDelete the file or run 'codify config' to re-enable interactive setup.\n\n", markerPath)
		}
	}
	return nil
}

// isInteractiveSuitable determina si el comando que está corriendo amerita
// ofrecer el wizard. Cubre:
//   - El root `codify` sin subcomando (camino más natural post-instalación
//     — sin esto, el primer launch nunca dispara el wizard).
//   - Los flujos que producen output usando defaults globales (generate,
//     analyze, spec, skills, workflows, hooks, init).
//   - Los lifecycle/utility commands donde el usuario llega naturalmente
//     en su primer contacto (usage, check, audit, update, watch, list,
//     reset-state).
//
// Quedan fuera: --help, --version, serve, config (este último ya lanza
// el wizard por sí mismo). Cobra short-circuitea --help/--version antes
// de ejecutar RunE, por lo que PersistentPreRunE no fire para ellos.
func isInteractiveSuitable(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	switch cmd.Name() {
	case "codify",
		"generate", "analyze", "spec", "skills", "workflows", "hooks", "init",
		"usage", "check", "audit", "update", "watch", "list", "reset-state":
		return true
	default:
		return false
	}
}
