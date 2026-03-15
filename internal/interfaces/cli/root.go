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
	rootCmd.AddCommand(commands.NewServeCmd())
	rootCmd.AddCommand(commands.NewListCmd())

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.codify.yaml)")
}