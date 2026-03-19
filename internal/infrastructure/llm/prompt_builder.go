package llm

import (
	"fmt"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
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
	// Skills command output files (all produce SKILL.md in separate directories)
	"ddd_entity":       "SKILL.md",
	"clean_arch_layer": "SKILL.md",
	"bdd_scenario":     "SKILL.md",
	"cqrs_command":     "SKILL.md",
	"hexagonal_port":   "SKILL.md",
	"code_review":      "SKILL.md",
	"test_strategy":    "SKILL.md",
	"refactor_safely":      "SKILL.md",
	"api_design":           "SKILL.md",
	"conventional_commit":  "SKILL.md",
	"semantic_versioning":  "SKILL.md",
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

// targetEcosystemDescriptions provides context about each target ecosystem's SKILL.md format.
var targetEcosystemDescriptions = map[string]string{
	"claude": `Target ecosystem: Claude Code (Anthropic)
Skills are installed in ~/.claude/skills/ (global) or .claude/skills/ (project).
YAML frontmatter fields: name, description, allowed-tools, context, agent, user-invocable.
The skill is invoked via /skill-name or auto-invoked when Claude detects relevance.
Substitutions available: $ARGUMENTS, ${CLAUDE_SKILL_DIR}.`,

	"codex": `Target ecosystem: Codex CLI (OpenAI)
Skills are installed in ~/.codex/skills/ (global) or .agents/skills/ (project).
YAML frontmatter fields: name, description.
The skill is invoked via $skill-name or implicitly when Codex detects relevance.
Optional: agents/openai.yaml for UI metadata.`,

	"antigravity": `Target ecosystem: Antigravity IDE (Google)
Skills are installed in ~/.gemini/antigravity/skills/ (global) or .agent/skills/ (project).
YAML frontmatter fields: name, description, triggers.
The skill is auto-invoked when the agent determines relevance to the current request.
Skills can bundle scripts in scripts/ subdirectory.`,
}

// BuildPersonalizedSkillsSystemPrompt returns a system prompt for generating personalized Agent Skills.
// Unlike the generic version, this prompt instructs the LLM to adapt the skill to the user's project.
func (b *PromptBuilder) BuildPersonalizedSkillsSystemPrompt(skillName, target, locale, projectContext string) string {
	ecosystemDesc := targetEcosystemDescriptions[target]
	if ecosystemDesc == "" {
		ecosystemDesc = targetEcosystemDescriptions["claude"]
	}

	return fmt.Sprintf(`<role>
You are a senior software architect specialized in creating personalized Agent Skills.
Agent Skills are markdown-based instruction packages (SKILL.md) that teach AI coding agents
how to approach specific architectural and engineering tasks.
</role>

<task>
Generate a complete, production-ready SKILL.md file for the skill: %s.
This skill must be PERSONALIZED to the user's project context provided below.
The output must include proper YAML frontmatter for the target ecosystem.
</task>

<project_context>
%s
</project_context>

<target_ecosystem>
%s
</target_ecosystem>

<skill_format>
The SKILL.md file MUST follow this structure:

1. YAML frontmatter (between --- markers) with at minimum: name, description
2. Clear description of WHEN to use this skill (triggers/scenarios)
3. Step-by-step PROCESS the agent should follow
4. Concrete CODE EXAMPLES adapted to the project's domain, language, and patterns
5. ANTI-PATTERNS to avoid with explanations of why
6. VERIFICATION checklist to confirm correct application
</skill_format>

<personalization_rules>
CRITICAL — Adapt this skill to the project context:

1. Use the project's programming language for ALL code examples
2. Use the project's actual domain concepts (entities, services, modules) in examples
3. Reference the project's architecture patterns and conventions
4. Adapt naming conventions to match the project's style
5. Include project-specific considerations (frameworks, libraries, tools in use)
6. If the project uses specific testing frameworks, patterns, or CI/CD tools, reference them
7. The skill should feel tailor-made for this specific project, not generic

DO NOT:
- Use generic examples (Order, User, Product) when the project context provides real domain concepts
- Ignore the project's language or framework in favor of generic patterns
- Add patterns or tools not relevant to the project's stack
</personalization_rules>

<output_quality>
- Complete YAML frontmatter appropriate for the target ecosystem
- Maximum 200 lines of content (personalized skills may need more detail)
- Code examples in fenced blocks with the project's language tag
- Structured with clear markdown headers
- Every instruction must be directly actionable within this project's codebase
</output_quality>

<rules>
- Respond ONLY with the complete SKILL.md content (frontmatter + body)
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
- Start with the --- YAML frontmatter delimiter
</rules>`, skillName, projectContext, ecosystemDesc, outputLanguageName(locale))
}

// BuildSkillsUserMessage constructs the user message for generating a single skill.
func (b *PromptBuilder) BuildSkillsUserMessage(guide service.TemplateGuide, target string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<skill_name>%s</skill_name>\n\n", guide.Name))
	sb.WriteString(fmt.Sprintf("<target_ecosystem>%s</target_ecosystem>\n\n", target))
	sb.WriteString("<template_guide>\n")
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