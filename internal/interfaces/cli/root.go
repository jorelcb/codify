package cli

import (
	"fmt"

	"github.com/jorelcb/ai-context-generator/internal/interfaces/cli/commands"
	"github.com/spf13/cobra"
)

// Version information (set from main.go)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ai-context-generator",
	Short: "AI Context Generator - Generate context-rich project documentation",
	Long: `AI Context Generator is a tool to generate comprehensive project documentation
and scaffolding for AI-assisted development.

It creates structured markdown files (prompts, context, interactions, changelog)
and project scaffolding based on templates.`,
	Version: Version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"AI Context Generator v%s\nCommit: %s\nBuilt: %s\n",
		Version, Commit, Date,
	))

	// Add subcommands
	rootCmd.AddCommand(commands.NewGenerateCmd())
	rootCmd.AddCommand(commands.NewListCmd())

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-context-generator.yaml)")
}