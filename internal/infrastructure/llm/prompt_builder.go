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
	"agents":       "AGENTS.md",
	"context":      "CONTEXT.md",
	"interactions": "INTERACTIONS_LOG.md",
	// Spec command output files
	"constitution": "CONSTITUTION.md",
	"spec":         "SPEC.md",
	"plan":         "PLAN.md",
	"tasks":        "TASKS.md",
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
	return fmt.Sprintf(`<role>
Eres un arquitecto de software senior y escritor tecnico experto.
Tu tarea es generar archivos de contexto optimizados para desarrollo de software asistido por IA.
Los archivos que generas seran consumidos por agentes de IA como contexto de trabajo.
</role>

<task>
Genera el contenido para el archivo %s.
Recibiras una descripcion del proyecto y una guia de template estructural.
</task>

<workflow>
1. Analiza la descripcion del proyecto: identifica lenguaje, arquitectura, tipo, capacidades clave
2. Lee la guia de template proporcionada en el mensaje del usuario
3. Para cada seccion del template, genera contenido ESPECIFICO y ACCIONABLE para el proyecto descrito
4. Donde el template usa variables como {{VARIABLE}} o condicionales, genera contenido real inferido del proyecto
5. Verifica que toda la informacion sea internamente consistente
6. Coloca la informacion mas critica al INICIO y al FINAL del archivo
</workflow>

<output_quality>
- Maximo 200 lineas por archivo generado
- Cero oraciones de relleno o boilerplate generico
- Formatos estructurados (YAML, listas, tablas) sobre prosa para configuracion y specs
- Cada oracion debe ser accionable y util para un agente de IA consumidor
- Informacion critica al inicio y final del archivo (attention-aware ordering)
- Comandos deben ser exactos y copy-pasteables, no placeholders genericos
</output_quality>

<rules>
- Responde SOLO con el contenido markdown del archivo
- NO envuelvas la respuesta en bloques de codigo
- NO agregues explicaciones antes o despues del contenido
- El contenido debe estar en Espanol
- Usa la guia de template como referencia estructural, NO como template de reemplazo de variables
</rules>`, FileOutputName(guideName))
}

// BuildUserMessageForFile constructs the user message for generating a single file.
func (b *PromptBuilder) BuildUserMessageForFile(req service.GenerationRequest, guide service.TemplateGuide) string {
	var sb strings.Builder

	sb.WriteString("<project_description>\n")
	sb.WriteString(req.ProjectDescription)
	sb.WriteString("\n</project_description>\n\n")

	hasMetadata := req.Language != "" || req.ProjectType != "" || req.Architecture != ""
	if hasMetadata {
		sb.WriteString("<project_metadata>\n")
		if req.Language != "" {
			sb.WriteString(fmt.Sprintf("- Lenguaje: %s\n", req.Language))
		}
		if req.ProjectType != "" {
			sb.WriteString(fmt.Sprintf("- Tipo de proyecto: %s\n", req.ProjectType))
		}
		if req.Architecture != "" {
			sb.WriteString(fmt.Sprintf("- Arquitectura: %s\n", req.Architecture))
		}
		sb.WriteString("</project_metadata>\n\n")
	}

	sb.WriteString(fmt.Sprintf("<template_guide file=\"%s\">\n", FileOutputName(guide.Name)))
	sb.WriteString(guide.Content)
	sb.WriteString("\n</template_guide>\n")

	return sb.String()
}

// BuildSpecSystemPrompt returns a system prompt for generating spec files from existing context.
func (b *PromptBuilder) BuildSpecSystemPrompt(existingContext string) string {
	return fmt.Sprintf(`<role>
Eres un arquitecto de software senior especializado en especificaciones tecnicas.
Tu tarea es generar documentos de especificacion SDD (Spec-Driven Development) a partir de un contexto de proyecto existente.
El contexto que recibes fue generado previamente y contiene la arquitectura, patrones y decisiones del proyecto.
</role>

<task>
Genera documentos de especificacion accionables basados en el contexto existente del proyecto.
Recibiras el contexto completo del proyecto y una guia de template para el archivo especifico a generar.
</task>

<existing_context>
%s
</existing_context>

<workflow>
1. Analiza profundamente el contexto existente: arquitectura, stack, patrones, restricciones
2. Lee la guia de template proporcionada en el mensaje del usuario
3. Genera especificaciones CONCRETAS y COHERENTES con el contexto existente
4. Cada especificacion debe ser implementable y verificable
5. Mantén coherencia total con las decisiones arquitectonicas ya documentadas
6. Las tareas deben tener criterios de aceptacion claros
</workflow>

<output_quality>
- Especificaciones accionables, no genericas
- Criterios de aceptacion verificables
- Coherencia total con el contexto existente
- Formatos estructurados (listas, tablas, YAML) sobre prosa
- Maximo 200 lineas por archivo
</output_quality>

<rules>
- Responde SOLO con el contenido markdown del archivo
- NO envuelvas la respuesta en bloques de codigo
- NO agregues explicaciones antes o despues del contenido
- El contenido debe estar en Espanol
- Basa TODO el contenido en el contexto existente proporcionado
</rules>`, existingContext)
}
