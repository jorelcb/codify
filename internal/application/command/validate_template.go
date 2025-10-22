package command

import (
	"context"
	"fmt"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// ValidateTemplateCommand representa el comando para validar un template
type ValidateTemplateCommand struct {
	templateRepo template.Repository
}

// NewValidateTemplateCommand crea una nueva instancia del comando
func NewValidateTemplateCommand(templateRepo template.Repository) *ValidateTemplateCommand {
	return &ValidateTemplateCommand{
		templateRepo: templateRepo,
	}
}

// ValidationResult representa el resultado de la validación
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

// Execute ejecuta el comando de validación
func (c *ValidateTemplateCommand) Execute(ctx context.Context, templateID string) (*ValidationResult, error) {
	// 1. Buscar el template
	tmpl, err := c.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// 2. Validar usando el método del domain
	if err := tmpl.Validate(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// 3. Validaciones adicionales
	if len(tmpl.Content()) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "template content is empty")
	}

	// TODO: Add more validations when Template entity has these methods
	// - ExtractVariables()
	// - Description()
	// - Tags()

	return result, nil
}