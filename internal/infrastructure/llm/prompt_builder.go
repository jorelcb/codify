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
	"agents":            "AGENTS.md",
	"context":           "CONTEXT.md",
	"interactions":      "INTERACTIONS_LOG.md",
	"development_guide": "DEVELOPMENT_GUIDE.md",
	"idioms":            "IDIOMS.md",
	// Spec command output files
	"constitution": "CONSTITUTION.md",
	"spec":         "SPEC.md",
	"plan":         "PLAN.md",
	"tasks":        "TASKS.md",
}

// localeLanguageNames maps locale codes to their language name for the LLM directive.
var localeLanguageNames = map[string]string{
	"en": "English",
	"es": "Spanish",
}

// FileOutputName returns the output file name for a given template guide name.
func FileOutputName(guideName string) string {
	if name, ok := fileOutputNames[guideName]; ok {
		return name
	}
	return guideName + ".md"
}

// outputLanguageName returns the language name for the given locale (defaults to English).
func outputLanguageName(locale string) string {
	if name, ok := localeLanguageNames[locale]; ok {
		return name
	}
	return "English"
}

// BuildSystemPromptForFile returns a system prompt for generating a single context file.
func (b *PromptBuilder) BuildSystemPromptForFile(guideName string, locale string) string {
	return fmt.Sprintf(`<role>
You are a senior software architect and expert technical writer.
Your task is to generate context files optimized for AI-assisted software development.
The files you generate will be consumed by AI agents as working context.
</role>

<task>
Generate the content for the file %s.
You will receive a project description and a structural template guide.
</task>

<grounding_rules>
CRITICAL RULE — Distinguish between two types of content:

1. TECHNICAL FRAMEWORK (you may opine freely): architectural patterns, project structure,
   code conventions, testing strategy, observability, DDD layers. This is the template's value.

2. DOMAIN LOGIC (only what the user stated): business rules, specific validations,
   default values, edge cases, data formats, concrete behaviors, error messages.

For domain logic:
- Only include what is EXPLICITLY in the project description
- DO NOT invent validation rules, default values, formats, or behaviors the user did not mention
- DO NOT generate speculative edge cases or error scenarios
- If a template section asks for domain details the description does not cover, indicate it must be defined by the team instead of inventing an answer
- Prefer marking "[DEFINE]" over inventing a concrete business rule
</grounding_rules>

<workflow>
1. Analyze the project description: identify language, architecture, type, key capabilities
2. Read the template guide provided in the user message
3. Mentally separate: what is technical framework (opine freely) vs domain logic (only what was stated)
4. For each template section, generate SPECIFIC and ACTIONABLE content for the described project
5. Where the template uses variables like {{VARIABLE}}, generate real content ONLY if the description supports it; otherwise mark as [DEFINE]
6. Verify that no business rule or specific behavior was invented
7. Place the most critical information at the BEGINNING and END of the file
</workflow>

<output_quality>
- Maximum 200 lines per generated file
- Zero filler sentences or generic boilerplate
- Structured formats (YAML, lists, tables) over prose for configuration and specs
- Every sentence must be actionable and useful for a consuming AI agent
- Critical information at the beginning and end of the file (attention-aware ordering)
- Commands must be exact and copy-pasteable, not generic placeholders
</output_quality>

<rules>
- Respond ONLY with the markdown content of the file
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
- Use the template guide as structural reference, NOT as a variable replacement template
</rules>`, FileOutputName(guideName), outputLanguageName(locale))
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
			sb.WriteString(fmt.Sprintf("- Language: %s\n", req.Language))
		}
		if req.ProjectType != "" {
			sb.WriteString(fmt.Sprintf("- Project type: %s\n", req.ProjectType))
		}
		if req.Architecture != "" {
			sb.WriteString(fmt.Sprintf("- Architecture: %s\n", req.Architecture))
		}
		sb.WriteString("</project_metadata>\n\n")
	}

	sb.WriteString(fmt.Sprintf("<template_guide file=\"%s\">\n", FileOutputName(guide.Name)))
	sb.WriteString(guide.Content)
	sb.WriteString("\n</template_guide>\n")

	return sb.String()
}

// BuildSpecSystemPrompt returns a system prompt for generating spec files from existing context.
func (b *PromptBuilder) BuildSpecSystemPrompt(existingContext string, locale string) string {
	return fmt.Sprintf(`<role>
You are a senior software architect specialized in technical specifications.
Your task is to generate SDD (Spec-Driven Development) specification documents from an existing project context.
The context you receive was previously generated and contains the project's architecture, patterns, and decisions.
</role>

<task>
Generate actionable specification documents based on the existing project context.
You will receive the complete project context and a template guide for the specific file to generate.
</task>

<existing_context>
%s
</existing_context>

<grounding_rules>
CRITICAL RULE — Distinguish between two types of content:

1. TECHNICAL FRAMEWORK (you may opine freely): milestone structure, testing strategy,
   implementation phases, task dependency graph, design patterns.

2. DOMAIN LOGIC (only what is stated in the context): business rules, validations,
   default values, data formats, edge cases, error messages, specific behaviors.

For domain logic:
- Only include rules, validations, formats, and behaviors EXPLICITLY mentioned in the existing context
- DO NOT invent edge cases, default values, validation rules, or behaviors not documented
- DO NOT speculate on how errors or edge cases should be handled if the context does not mention them
- If a section requires domain details not covered, use "[DEFINE: brief description of what is missing]"
- A precise but incomplete specification is better than a complete one with invented rules
</grounding_rules>

<workflow>
1. Deeply analyze the existing context: architecture, stack, patterns, constraints
2. Read the template guide provided in the user message
3. Identify which domain information is EXPLICITLY in the context and which you would have to invent
4. Generate CONCRETE specifications COHERENT with the existing context
5. For domain details not covered in the context, mark [DEFINE] instead of inventing
6. Each specification must be implementable and verifiable
7. Maintain total coherence with already documented architectural decisions
</workflow>

<output_quality>
- Actionable specifications, not generic
- Verifiable acceptance criteria based on real context information
- Total coherence with existing context
- Structured formats (lists, tables, YAML) over prose
- Maximum 200 lines per file
- Zero invented business rules — if not in the context, mark [DEFINE]
</output_quality>

<rules>
- Respond ONLY with the markdown content of the file
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
- Base ALL content on the existing context provided
- Mark with [DEFINE] any business rule, validation, or behavior not in the context
</rules>`, existingContext, outputLanguageName(locale))
}