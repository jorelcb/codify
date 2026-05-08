package commands

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	domain "github.com/jorelcb/codify/internal/domain/config"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
)

// NewConfigCmd construye `codify config` y sus subcomandos (get, set, edit,
// list, unset). Sin args dispara el wizard interactivo (o muestra el config
// actual si ya existe).
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage user-level Codify configuration (~/.codify/config.yml)",
		Long: `Manage user-level Codify configuration persisted at ~/.codify/config.yml.

Without subcommand: launches an interactive wizard for first-time setup, or
prints the current effective configuration if it already exists.

Subcommands:
  codify config get <key>      Print a single value
  codify config set <key> <v>  Set a single value
  codify config edit           Open ~/.codify/config.yml in $EDITOR
  codify config list           Print full config
  codify config unset <key>    Clear a single value

Valid keys: preset, locale, language, model, target, provider, project_name`,
		RunE: runConfigDefault,
	}

	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigUnsetCmd())
	cmd.AddCommand(newConfigEditCmd())
	cmd.AddCommand(newConfigListCmd())

	return cmd
}

// runConfigDefault implementa el comportamiento sin subcomando.
//   - Si no existe ~/.codify/config.yml: dispara el wizard interactivo (si TTY).
//   - Si existe: imprime el config actual.
//   - Si no es TTY y no existe: imprime instrucciones y sale ok.
func runConfigDefault(cmd *cobra.Command, args []string) error {
	repo := infraconfig.NewRepository()
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}

	cfg, exists, err := repo.Load(userPath)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Codify user config: %s\n\n", userPath)
		printConfig(cfg)
		fmt.Println("\nSubcommands: get, set, edit, list, unset. See 'codify config --help'.")
		return nil
	}

	if !isInteractive() {
		fmt.Fprintf(os.Stderr, "No user config found at %s.\n", userPath)
		fmt.Fprintln(os.Stderr, "Run 'codify config' from a TTY to launch the interactive wizard, or set values explicitly via 'codify config set <key> <value>'.")
		return nil
	}

	return runConfigWizard(repo, userPath)
}

// runConfigWizard ejecuta el wizard interactivo de primera vez.
//
// El wizard tiene dos fases:
//   - Defaults globales: target, model, locale, preset (default architectural
//     posture para comandos que no pasan --preset).
//   - Install opcional a nivel agente: skills globales por categoría + hooks
//     globales (solo Claude). Reutiliza la lógica de los comandos `skills` y
//     `hooks` para no duplicar pipelines.
//
// El orden es deliberado: target primero porque condiciona los paths de
// skills/hooks; preset al final porque es el más específico y solo aplica
// a comandos de proyecto futuros, no a la configuración del agente.
func runConfigWizard(repo *infraconfig.Repository, path string) error {
	fmt.Println("Codify · Bootstrap (workstation)")
	fmt.Println("════════════════════════════════")

	cfg := domain.BuiltinDefaults()

	target, err := promptSelect("Default target ecosystem", []selectOption{
		{"Claude Code (recommended — full support: skills, workflows, hooks)", "claude"},
		{"Codex (skills only)", "codex"},
		{"Antigravity (skills + workflows)", "antigravity"},
	}, "claude")
	if err != nil {
		return err
	}
	cfg.Target = target

	model, err := promptModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Skipping model selection: %v\n", err)
	} else {
		cfg.Model = model
	}

	locale, err := promptLocale()
	if err != nil {
		return err
	}
	cfg.Locale = locale

	preset, err := promptPreset()
	if err != nil {
		return err
	}
	cfg.Preset = preset

	if err := repo.Save(path, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("\n✓ Saved %s\n", path)

	if err := promptInstallSkills(target, locale, "global"); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: global skills install step failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "You can retry anytime with 'codify skills --install global'.")
	}

	if err := promptInstallWorkflows(target, locale, "global"); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: global workflows install step failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "You can retry anytime with 'codify workflows --install global'.")
	}

	if target == "claude" {
		if err := promptInstallHooks(locale, "global"); err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: global hooks install step failed: %v\n", err)
			fmt.Fprintln(os.Stderr, "You can retry anytime with 'codify hooks --install global'.")
		}
	}

	printConfigNextSteps()
	return nil
}

// printConfigNextSteps imprime el siguiente paso natural tras el bootstrap a
// nivel workstation: bootstrappear un proyecto. Mantiene paridad estructural
// con printInitNextSteps (init.go) para que el usuario vea la misma forma
// "Next steps" en ambos puntos de entrada (config y init).
func printConfigNextSteps() {
	fmt.Println()
	fmt.Println("✓ Workstation defaults saved.")
	fmt.Println()
	fmt.Println("Next steps")
	fmt.Println("──────────")
	fmt.Println()
	fmt.Println("Bootstrap (per project):")
	fmt.Println("  codify init       Bootstrap a project (new or existing) using these defaults")
	fmt.Println()
	fmt.Println("Update workstation defaults later:")
	fmt.Println("  codify config     Re-run this wizard")
	fmt.Println("  codify config set <key> <value>")
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Print a single config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := infraconfig.NewRepository()
			userPath, err := infraconfig.UserConfigPath()
			if err != nil {
				return err
			}
			cfg, _, err := repo.Load(userPath)
			if err != nil {
				return err
			}
			val, err := cfg.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Println(val)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a single config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mutateConfig(func(cfg *domain.Config) error {
				return cfg.Set(args[0], args[1])
			})
		},
	}
}

func newConfigUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Clear a single config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mutateConfig(func(cfg *domain.Config) error {
				return cfg.Unset(args[0])
			})
		},
	}
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open ~/.codify/config.yml in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			userPath, err := infraconfig.UserConfigPath()
			if err != nil {
				return err
			}
			// Asegurar que exista (vacío) para que el editor abra algo válido
			if !infraconfig.FileExists(userPath) {
				if err := infraconfig.NewRepository().Save(userPath, domain.BuiltinDefaults()); err != nil {
					return err
				}
			}
			ed := exec.Command(editor, userPath)
			ed.Stdin = os.Stdin
			ed.Stdout = os.Stdout
			ed.Stderr = os.Stderr
			return ed.Run()
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Print the full effective configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := infraconfig.NewRepository()
			cfg, err := repo.LoadEffective()
			if err != nil {
				return err
			}
			printConfig(cfg)
			return nil
		},
	}
}

// mutateConfig ejecuta una mutación sobre el config persistido a nivel
// usuario, manejando load → mutate → save de forma transaccional.
func mutateConfig(mut func(*domain.Config) error) error {
	repo := infraconfig.NewRepository()
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	cfg, _, err := repo.Load(userPath)
	if err != nil {
		return err
	}
	if err := mut(&cfg); err != nil {
		return err
	}
	return repo.Save(userPath, cfg)
}

// printConfig imprime el config en formato key: value, con keys en orden
// estable. Vacíos se muestran como "(unset)".
func printConfig(cfg domain.Config) {
	keys := domain.Keys()
	sort.Strings(keys)
	for _, k := range keys {
		v, _ := cfg.Get(k)
		if v == "" {
			fmt.Printf("  %-15s (unset)\n", k+":")
		} else {
			fmt.Printf("  %-15s %s\n", k+":", v)
		}
	}
	if cfg.UpdatedAt != "" {
		fmt.Printf("\n  %-15s %s\n", "updated_at:", cfg.UpdatedAt)
	}
	_ = strings.TrimSpace
}
