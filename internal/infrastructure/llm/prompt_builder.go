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
	// Testing skills
	"test_foundational": "SKILL.md",
	"test_tdd":          "SKILL.md",
	"test_bdd":          "SKILL.md",
	// Workflow command output files (flat .md files, not subdirectories)
	"feature_development": "feature-development.md",
	"bug_fix":             "bug-fix.md",
	"release_cycle":       "release-cycle.md",
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

// BuildAnalyzeSystemPromptForFile returns a system prompt optimized for analyze mode.
// Unlike the generate prompt, this treats scan data as factual ground truth from real code.
func (b *PromptBuilder) BuildAnalyzeSystemPromptForFile(guideName string, locale string) string {
	return fmt.Sprintf(`<role>
You are a senior software architect and expert technical writer.
Your task is to generate context files optimized for AI-assisted software development.
The files you generate will be consumed by AI agents as working context.
</role>

<task>
Generate the content for the file %s.
You will receive a project analysis AUTO-SCANNED from an existing codebase and a structural template guide.
</task>

<scan_trust>
The project description was extracted by scanning a real codebase. The following signals are FACTUAL:
- Language and framework: detected from manifest files (go.mod, package.json, etc.)
- Dependencies: parsed from the actual dependency manifest
- Directory structure: read from the real filesystem
- README content: extracted from the project's README file
- Infrastructure signals: detected from real config files (Dockerfile, CI workflows, etc.)
- Build targets: parsed from actual Makefile/Taskfile if present
- Existing context files: read verbatim from the project
- Testing patterns: detected from real test files and framework dependencies
- CI/CD pipelines: summarized from actual workflow definitions

Trust these signals fully. Generate content that matches the REAL state of the codebase.
</scan_trust>

<grounding_rules>
CRITICAL RULE — Distinguish between two types of content:

1. TECHNICAL FRAMEWORK + SCANNED SIGNALS (generate with confidence): architectural patterns,
   project structure, code conventions, testing strategy, build commands, CI/CD pipeline,
   dependencies. These are backed by real scan data — use them directly.

2. DOMAIN LOGIC (only what is explicitly stated): business rules, specific validations,
   default values, edge cases, data formats, concrete behaviors, error messages.

For scanned signals:
- Use detected language, framework, and dependencies as ground truth
- Generate exact commands from build targets (make build, task test, etc.)
- Reference the actual directory structure when describing the project layout
- Incorporate existing context files to maintain continuity with prior decisions
- Describe the real CI/CD pipeline, not a hypothetical one

For domain logic:
- Only include what is EXPLICITLY present in the README or existing context files
- DO NOT invent business rules, validations, or behaviors not documented in the scan
- Mark "[DEFINE]" for business-domain concepts not inferable from the codebase
</grounding_rules>

<workflow>
1. Analyze the scanned project data: language, framework, dependencies, structure, infrastructure
2. Read existing context files if present — they represent prior architectural decisions
3. Read the template guide provided in the user message
4. For each template section, generate SPECIFIC content grounded in the scan data
5. For commands sections, use EXACT build targets detected (not generic placeholders)
6. Where the template asks for domain details not in the scan, mark as [DEFINE]
7. Place the most critical information at the BEGINNING and END of the file
</workflow>

<output_quality>
- Maximum 200 lines per generated file
- Zero filler sentences or generic boilerplate
- Structured formats (YAML, lists, tables) over prose for configuration and specs
- Every sentence must be actionable and useful for a consuming AI agent
- Critical information at the beginning and end of the file (attention-aware ordering)
- Commands must be exact and copy-pasteable, derived from actual build targets
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

// BuildPersonalizedWorkflowsSystemPrompt returns a system prompt for generating personalized Antigravity workflows.
func (b *PromptBuilder) BuildPersonalizedWorkflowsSystemPrompt(workflowName, locale, projectContext string) string {
	return fmt.Sprintf(`<role>
You are a senior DevOps engineer and workflow automation specialist.
Your task is to generate Antigravity workflow files — multi-step recipes that AI agents execute on demand.
Workflows are markdown files with numbered steps that teach agents how to perform complex, repeatable tasks.
</role>

<task>
Generate a complete, production-ready Antigravity workflow file for: %s.
This workflow must be PERSONALIZED to the user's project context provided below.
The output must include proper YAML frontmatter with a description field (max 250 characters).
</task>

<project_context>
%s
</project_context>

<workflow_format>
The workflow file MUST follow this structure:

1. YAML frontmatter with a "description" field (max 250 characters, between --- markers)
2. Numbered steps in markdown, each with a bold title and detailed instructions
3. Execution annotations where appropriate (see below)
4. Code blocks with exact commands the agent should run

Execution annotations (place on a line by itself before the instruction):
- // turbo — Auto-execute the next command without user confirmation (for safe, non-destructive operations)
- // turbo-all — Auto-execute ALL remaining commands (for fully automated, safe workflows)
- // parallel — Run this step concurrently with other parallel-marked steps
- // if [condition] — Conditional step: agent evaluates at runtime, skips if false
- // capture: VARIABLE_NAME — Capture command output into a variable, reference later as {{VARIABLE_NAME}}
- // run workflow: [name] — Invoke another workflow (composition)
- // retry: N — Retry a turbo step N times on failure
- // timeout: duration — Set a timeout for a turbo step (e.g., "30s", "5m")
</workflow_format>

<personalization_rules>
CRITICAL — Adapt this workflow to the project context:

1. Use the project's actual tools, frameworks, and commands (not generic placeholders)
2. Reference the project's branch naming conventions, CI/CD tools, and testing frameworks
3. Adapt file paths and directory structures to match the project's layout
4. Include project-specific considerations (monorepo vs single repo, deployment targets, etc.)
5. Use the project's package manager, build tools, and test runners in commands
6. If the project uses specific code review tools, issue trackers, or deployment platforms, reference them

DO NOT:
- Use generic commands when the project context provides specific tools
- Include steps irrelevant to the project's tech stack
- Assume tools or services not mentioned in the project context
</personalization_rules>

<output_quality>
- YAML frontmatter with description (max 250 chars)
- 5-15 numbered steps (enough detail without bloat)
- Exact, copy-pasteable commands in code blocks
- Appropriate use of execution annotations (turbo for safe commands, capture for outputs needed later)
- Each step should be independently understandable
- Steps should flow logically from start to finish
</output_quality>

<rules>
- Respond ONLY with the complete workflow file (frontmatter + numbered steps)
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
- Start with the --- YAML frontmatter delimiter
</rules>`, workflowName, projectContext, outputLanguageName(locale))
}

// BuildWorkflowsUserMessage constructs the user message for generating a single workflow.
func (b *PromptBuilder) BuildWorkflowsUserMessage(guide service.TemplateGuide, target string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<workflow_name>%s</workflow_name>\n\n", guide.Name))
	sb.WriteString(fmt.Sprintf("<target_ecosystem>%s</target_ecosystem>\n\n", target))
	sb.WriteString("<template_guide>\n")
	sb.WriteString(guide.Content)
	sb.WriteString("\n</template_guide>\n")

	return sb.String()
}

// BuildPluginSkillSystemPrompt returns a system prompt for generating a Claude Code plugin SKILL.md.
// The LLM generates the skill content; hooks, agents, and scripts are generated statically.
func (b *PromptBuilder) BuildPluginSkillSystemPrompt(workflowName, locale, projectContext, pluginName string) string {
	return fmt.Sprintf(`<role>
You are a senior DevOps engineer and workflow automation specialist for Claude Code plugins.
Your task is to generate the SKILL.md component of a Claude Code plugin that orchestrates
a multi-step workflow using the plugin's hooks, agents, and scripts.
</role>

<task>
Generate a complete, production-ready SKILL.md for the plugin: %s.
Workflow: %s.
This skill must be PERSONALIZED to the user's project context provided below.
The output must include proper YAML frontmatter (name, description).
</task>

<project_context>
%s
</project_context>

<plugin_context>
This SKILL.md is part of a Claude Code plugin that includes:
- hooks/hooks.json: Auto-approval hooks for safe commands (PreToolUse), output capture (PostToolUse),
  and conditional step evaluation — these are already configured, do NOT duplicate their logic in prose
- agents/workflow-runner.md: A subagent with tool access (Bash, Read, Edit, Write, Grep, Glob) that
  executes workflow steps — reference it when delegation is appropriate
- scripts/capture-output.sh: Captures command output for use in later steps (already wired via hooks)

The plugin root is available as ${CLAUDE_PLUGIN_ROOT} in hook commands.
Persistent state can be stored in ${CLAUDE_PLUGIN_DATA}.
</plugin_context>

<skill_format>
The SKILL.md file MUST follow this structure:

1. YAML frontmatter (between --- markers) with:
   - name: workflow-name (kebab-case)
   - description: Short description of what this workflow does

2. Numbered steps in markdown with clear, actionable instructions
3. Code blocks with exact commands the agent should run
4. References to plugin components where appropriate

CRITICAL:
- DO NOT include Antigravity execution annotations (// turbo, // capture, // if, etc.)
- DO NOT write prose equivalents of hooks that are already configured (e.g., "auto-approve this")
- DO reference the workflow-runner agent for tool-heavy steps
- DO include conditional logic as natural prose ("If X, then Y; otherwise skip")
- Annotations are handled by the plugin's hooks — the skill focuses on WHAT to do, not HOW to approve
</skill_format>

<personalization_rules>
CRITICAL — Adapt this workflow to the project context:

1. Use the project's actual tools, frameworks, and commands (not generic placeholders)
2. Reference the project's branch naming conventions, CI/CD tools, and testing frameworks
3. Adapt file paths and directory structures to match the project's layout
4. Include project-specific considerations (monorepo vs single repo, deployment targets, etc.)
5. Use the project's package manager, build tools, and test runners in commands

DO NOT:
- Use generic commands when the project context provides specific tools
- Include steps irrelevant to the project's tech stack
- Assume tools or services not mentioned in the project context
</personalization_rules>

<output_quality>
- YAML frontmatter with name and description
- 5-15 numbered steps (enough detail without bloat)
- Exact, copy-pasteable commands in code blocks
- Each step independently understandable
- Steps flow logically from start to finish
</output_quality>

<rules>
- Respond ONLY with the complete SKILL.md content (frontmatter + numbered steps)
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
- Start with the --- YAML frontmatter delimiter
</rules>`, pluginName, workflowName, projectContext, outputLanguageName(locale))
}

// BuildPluginSkillUserMessage constructs the user message for generating a plugin SKILL.md.
func (b *PromptBuilder) BuildPluginSkillUserMessage(guide service.TemplateGuide, pluginName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<plugin_name>%s</plugin_name>\n\n", pluginName))
	sb.WriteString(fmt.Sprintf("<workflow_name>%s</workflow_name>\n\n", guide.Name))
	sb.WriteString("<template_guide>\n")
	sb.WriteString(guide.Content)
	sb.WriteString("\n</template_guide>\n")

	return sb.String()
}