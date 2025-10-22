package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// TemplateProcessor provides domain operations for processing templates
type TemplateProcessor struct {
	repository template.Repository
}

// NewTemplateProcessor creates a new template processor service
func NewTemplateProcessor(repository template.Repository) *TemplateProcessor {
	return &TemplateProcessor{
		repository: repository,
	}
}

// Process processes a template with given variables
func (s *TemplateProcessor) Process(tmpl *template.Template, variables map[string]string) (string, error) {
	if tmpl == nil {
		return "", fmt.Errorf("template cannot be nil")
	}

	content := tmpl.Content()

	// Replace all variables in the format {{VARIABLE_NAME}}
	for _, v := range tmpl.Variables() {
		placeholder := fmt.Sprintf("{{%s}}", v.Name)
		value, exists := variables[v.Name]

		if !exists {
			if v.Required {
				return "", fmt.Errorf("required variable %s not provided", v.Name)
			}
			// Use default value if available
			value = v.DefaultValue
		}

		content = strings.ReplaceAll(content, placeholder, value)
	}

	// Check for any remaining unprocessed variables
	remainingVars := s.findUnprocessedVariables(content)
	if len(remainingVars) > 0 {
		return "", fmt.Errorf("unprocessed variables found: %v", remainingVars)
	}

	return content, nil
}

// ExtractVariables extracts all variables from template content
func (s *TemplateProcessor) ExtractVariables(content string) []string {
	re := regexp.MustCompile(`\{\{([A-Z_]+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	variables := make([]string, 0)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}
		}
	}

	return variables
}

// Validate validates a template and its variables
func (s *TemplateProcessor) Validate(tmpl *template.Template, variables map[string]string) error {
	if tmpl == nil {
		return fmt.Errorf("template cannot be nil")
	}

	// Validate template itself
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Check required variables
	for _, v := range tmpl.Variables() {
		if v.Required {
			if _, exists := variables[v.Name]; !exists && v.DefaultValue == "" {
				return fmt.Errorf("required variable %s is missing", v.Name)
			}
		}
	}

	return nil
}

// findUnprocessedVariables finds variables that were not replaced
func (s *TemplateProcessor) findUnprocessedVariables(content string) []string {
	re := regexp.MustCompile(`\{\{([A-Z_]+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	variables := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}

	return variables
}