package cli

import (
	"fmt"

	"github.com/jorelcb/codify/internal/interfaces/cli/commands"
	"github.com/spf13/cobra"
)

// Version information (set from main.go)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "codify",
	Short: "Codify - Provision, equip, and maintain AI development environments",
	Long: `Codify provisions, equips, and maintains AI development environments —
generating project context with LLMs (Anthropic Claude or Google Gemini) and
managing its lifecycle as the project evolves.

Requires ANTHROPIC_API_KEY (for Claude) or GEMINI_API_KEY (for Gemini) environment variable.

Lifecycle phases:

    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Bootstrap  │ ──▶ │    Equip    │ ──▶ │  Maintain   │
    └─────────────┘     └─────────────┘     └─────────────┘
       config              generate            check
       init                analyze             update
                           spec                audit
                           skills              watch
                           workflows           usage
                           hooks               resolve

  • Bootstrap (one-time)  — set up the workstation (config) or a project (init)
  • Equip      (per need) — generate context, install skills/workflows/hooks, write specs
  • Maintain   (ongoing)  — detect drift, regenerate, audit commits, track usage`,
	Version: Version,
	// PersistentPreRunE corre antes de cada subcomando, lo cual habilita el
	// auto-launch SOFT del wizard de configuración global la primera vez.
	// La función decide internamente si dispara o no según TTY, comando, y
	// presencia del marker file de opt-out (ver ADR-007).
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return commands.MaybeAutoLaunchConfig(cmd)
	},
	// RunE explícito para `codify` sin subcomando. Sin esto, cobra cae al
	// help directamente y NUNCA dispara PersistentPreRunE — eso significaba
	// que la primera vez que un usuario invoca `codify` (camino más natural
	// post-instalación) el wizard nunca se ofrecía. Ahora: el auto-launch
	// fires (vía PersistentPreRunE), y luego imprimimos el help estándar.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"Codify v%s\nCommit: %s\nBuilt: %s\n",
		Version, Commit, Date,
	))
	return rootCmd.Execute()
}

// Command group IDs map to lifecycle phases. Cobra renders commands grouped
// under their group's title in `codify --help`. See ADR-0007 for the
// Bootstrap/Equip/Maintain phase model.
const (
	groupBootstrap = "bootstrap"
	groupEquip     = "equip"
	groupMaintain  = "maintain"
	groupSystem    = "system"
)

// withGroup attaches a GroupID to a cobra command and returns it. Keeps the
// init() block readable without a separate statement per command.
func withGroup(cmd *cobra.Command, groupID string) *cobra.Command {
	cmd.GroupID = groupID
	return cmd
}

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"Codify v%s\nCommit: %s\nBuilt: %s\n",
		Version, Commit, Date,
	))

	// Register lifecycle phase groups. Order here drives display order in --help.
	rootCmd.AddGroup(
		&cobra.Group{ID: groupBootstrap, Title: "Bootstrap (one-time setup):"},
		&cobra.Group{ID: groupEquip, Title: "Equip (install context, skills, workflows, hooks, specs):"},
		&cobra.Group{ID: groupMaintain, Title: "Maintain (ongoing lifecycle: drift, audit, usage):"},
		&cobra.Group{ID: groupSystem, Title: "System:"},
	)

	// Add subcommands, grouped by lifecycle phase.
	rootCmd.AddCommand(withGroup(commands.NewConfigCmd(), groupBootstrap))
	rootCmd.AddCommand(withGroup(commands.NewInitCmd(), groupBootstrap))

	rootCmd.AddCommand(withGroup(commands.NewGenerateCmd(), groupEquip))
	rootCmd.AddCommand(withGroup(commands.NewAnalyzeCmd(), groupEquip))
	rootCmd.AddCommand(withGroup(commands.NewSpecCmd(), groupEquip))
	rootCmd.AddCommand(withGroup(commands.NewSkillsCmd(), groupEquip))
	rootCmd.AddCommand(withGroup(commands.NewWorkflowsCmd(), groupEquip))
	rootCmd.AddCommand(withGroup(commands.NewHooksCmd(), groupEquip))

	rootCmd.AddCommand(withGroup(commands.NewCheckCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewUpdateCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewAuditCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewWatchCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewUsageCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewResolveCmd(), groupMaintain))
	rootCmd.AddCommand(withGroup(commands.NewResetStateCmd(), groupMaintain))

	rootCmd.AddCommand(withGroup(commands.NewServeCmd(), groupSystem))
	rootCmd.AddCommand(withGroup(commands.NewListCmd(), groupSystem))

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.codify.yaml)")
}