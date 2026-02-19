package llm

import (
	"fmt"
	"strings"

	"github.com/jorelcb/ai-context-generator/internal/domain/service"
)

// PromptBuilder constructs prompts for the LLM from templates and project description.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// fileOutputNames maps template guide names to output file names.
var fileOutputNames = map[string]string{
	"prompt":       "PROMPT.md",
	"context":      "CONTEXT.md",
	"scaffolding":  "SCAFFOLDING.md",
	"interactions": "INTERACTIONS_LOG.md",
}

// FileOutputName returns the output file name for a given template guide name.
func FileOutputName(guideName string) string {
	if name, ok := fileOutputNames[guideName]; ok {
		return name
	}
	return guideName + ".md"
}

// BuildSystemPromptForFile returns a system prompt for generating a single context file.
func (b *PromptBuilder) BuildSystemPromptForFile(guideName string) string {
	return fmt.Sprintf(`You are an expert software architect and technical writer.

Your task is to generate the content for the file **%s** — a context file optimized for AI-assisted software development.

You will receive:
1. A description of the software project
2. A structural template guide that shows the expected format and sections for this file

Your job is to:
- Analyze the project description thoroughly
- Use the template guide as a structural reference (NOT as a template to fill with variables)
- Generate rich, specific, and intelligent content tailored to the described project
- Adapt language-specific, architecture-specific, and type-specific sections based on the project description
- Where the template uses variables like {{VARIABLE}}, replace them with actual content inferred from the project description

Respond ONLY with the markdown content for this file. Do NOT wrap it in code blocks. Do NOT add explanations before or after.
The content should be in Spanish, following the style of the template guide.`, FileOutputName(guideName))
}

// BuildUserMessageForFile constructs the user message for generating a single file.
func (b *PromptBuilder) BuildUserMessageForFile(req service.GenerationRequest, guide service.TemplateGuide) string {
	var sb strings.Builder

	sb.WriteString("## Descripcion del Proyecto\n\n")
	sb.WriteString(req.ProjectDescription)
	sb.WriteString("\n\n")

	if req.Language != "" {
		sb.WriteString(fmt.Sprintf("**Lenguaje:** %s\n", req.Language))
	}
	if req.ProjectType != "" {
		sb.WriteString(fmt.Sprintf("**Tipo de proyecto:** %s\n", req.ProjectType))
	}
	if req.Architecture != "" {
		sb.WriteString(fmt.Sprintf("**Arquitectura:** %s\n", req.Architecture))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString(fmt.Sprintf("## Guia Estructural para %s\n\n", FileOutputName(guide.Name)))
	sb.WriteString("Usa esta guia como referencia de formato y secciones. NO reemplaces variables literalmente, genera contenido inteligente y especifico al proyecto descrito.\n\n")
	sb.WriteString("```\n")
	sb.WriteString(guide.Content)
	sb.WriteString("\n```\n")

	return sb.String()
}
