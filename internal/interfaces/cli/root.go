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
	Short: "Codify - Generate AI-optimized context files for your projects",
	Long: `Codify takes your project description and generates
context files using LLMs (Anthropic Claude or Google Gemini). These files give
your AI development agent the architectural context it needs to build coherently.

Requires ANTHROPIC_API_KEY (for Claude) or GEMINI_API_KEY (for Gemini) environment variable.`,
	Version: Version,
	// PersistentPreRunE corre antes de cada subcomando, lo cual habilita el
	// auto-launch SOFT del wizard de configuración global la primera vez.
	// La función decide internamente si dispara o no según TTY, comando, y
	// presencia del marker file de opt-out (ver ADR-007).
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return commands.MaybeAutoLaunchConfig(cmd)
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

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"Codify v%s\nCommit: %s\nBuilt: %s\n",
		Version, Commit, Date,
	))

	// Add subcommands
	rootCmd.AddCommand(commands.NewGenerateCmd())
	rootCmd.AddCommand(commands.NewAnalyzeCmd())
	rootCmd.AddCommand(commands.NewSpecCmd())
	rootCmd.AddCommand(commands.NewSkillsCmd())
	rootCmd.AddCommand(commands.NewWorkflowsCmd())
	rootCmd.AddCommand(commands.NewHooksCmd())
	rootCmd.AddCommand(commands.NewServeCmd())
	rootCmd.AddCommand(commands.NewListCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewInitCmd())
	rootCmd.AddCommand(commands.NewCheckCmd())
	rootCmd.AddCommand(commands.NewResetStateCmd())
	rootCmd.AddCommand(commands.NewUsageCmd())
	rootCmd.AddCommand(commands.NewUpdateCmd())
	rootCmd.AddCommand(commands.NewAuditCmd())
	rootCmd.AddCommand(commands.NewWatchCmd())

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.codify.yaml)")
}