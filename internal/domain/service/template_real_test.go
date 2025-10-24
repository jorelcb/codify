package service

import (
	"strings"
	"testing"
)

func TestTemplateEngine_RealTemplateSnippet(t *testing.T) {
	engine := NewTemplateEngine()

	// Real snippet from context.template
	realTemplate := `# {{PROJECT_NAME}} - Documento de Contexto Definitivo para Agentes de IA

## 1. Visión General del Proyecto

### Propósito
{{DESCRIPTION}}

### Características Principales
{{#if (eq PROJECT_TYPE "api")}}
- **API RESTful**: Endpoints bien definidos siguiendo estándares REST
- **Documentación OpenAPI**: Swagger/OpenAPI 3.0 generado automáticamente
- **Versionado de API**: Soporte para múltiples versiones
{{/if}}
{{#if (eq PROJECT_TYPE "cli")}}
- **Comandos intuitivos**: Interfaz de línea de comandos clara
- **Configuración flexible**: Archivos de config, env vars y flags
- **Output formateado**: JSON, YAML, tabla o texto plano
{{/if}}
- **Arquitectura basada en interfaces**: Abstracciones claras según DDD`

	context := map[string]interface{}{
		"PROJECT_NAME": "MyAwesomeAPI",
		"DESCRIPTION":  "A powerful REST API for managing resources",
		"PROJECT_TYPE": "api",
	}

	result, err := engine.Render(realTemplate, context)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify key parts are present
	expectedParts := []string{
		"# MyAwesomeAPI - Documento de Contexto",
		"A powerful REST API for managing resources",
		"- **API RESTful**: Endpoints bien definidos",
		"- **Documentación OpenAPI**: Swagger/OpenAPI 3.0",
		"- **Arquitectura basada en interfaces**",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("result should contain %q\nGot:\n%s", part, result)
		}
	}

	// Verify CLI-specific parts are NOT present (since PROJECT_TYPE is "api")
	notExpectedParts := []string{
		"**Comandos intuitivos**",
		"**Configuración flexible**",
	}

	for _, part := range notExpectedParts {
		if strings.Contains(result, part) {
			t.Errorf("result should NOT contain %q (wrong PROJECT_TYPE)\nGot:\n%s", part, result)
		}
	}

	t.Logf("Successfully rendered real template:\n%s", result)
}

func TestTemplateEngine_MultipleProjectTypes(t *testing.T) {
	engine := NewTemplateEngine()

	template := `# {{PROJECT_NAME}}

{{#if (eq PROJECT_TYPE "api")}}
This is an API project
{{/if}}
{{#if (eq PROJECT_TYPE "cli")}}
This is a CLI project
{{/if}}
{{#if (eq PROJECT_TYPE "library")}}
This is a library project
{{/if}}`

	testCases := []struct {
		projectType string
		expected    string
		notExpected []string
	}{
		{
			projectType: "api",
			expected:    "This is an API project",
			notExpected: []string{"This is a CLI project", "This is a library project"},
		},
		{
			projectType: "cli",
			expected:    "This is a CLI project",
			notExpected: []string{"This is an API project", "This is a library project"},
		},
		{
			projectType: "library",
			expected:    "This is a library project",
			notExpected: []string{"This is an API project", "This is a CLI project"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.projectType, func(t *testing.T) {
			context := map[string]interface{}{
				"PROJECT_NAME": "TestProject",
				"PROJECT_TYPE": tc.projectType,
			}

			result, err := engine.Render(template, context)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(result, tc.expected) {
				t.Errorf("result should contain %q\nGot:\n%s", tc.expected, result)
			}

			for _, notExp := range tc.notExpected {
				if strings.Contains(result, notExp) {
					t.Errorf("result should NOT contain %q\nGot:\n%s", notExp, result)
				}
			}
		})
	}
}