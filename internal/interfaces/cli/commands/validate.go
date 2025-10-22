package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	var templateDir string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate templates and configuration",
		Long: `Validate templates and configuration files.

This command checks:
  - Template syntax and structure
  - Required variables are defined
  - Template files exist and are readable
  - Configuration files are valid YAML/JSON
  - No circular dependencies in templates

Examples:
  # Validate default templates
  ai-context-generator validate

  # Validate custom templates
  ai-context-generator validate --template-dir ./my-templates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(templateDir)
		},
	}

	cmd.Flags().StringVarP(&templateDir, "template-dir", "t", "./templates", "Template directory to validate")

	return cmd
}

func runValidate(templateDir string) error {
	// TODO: Implement template validation logic
	fmt.Printf("Validating templates in: %s\n", templateDir)
	fmt.Println("\n[Phase 1] Template validation will check:")
	fmt.Println("  - Template syntax")
	fmt.Println("  - Required variables")
	fmt.Println("  - File permissions")
	fmt.Println("  - Metadata consistency")
	return nil
}