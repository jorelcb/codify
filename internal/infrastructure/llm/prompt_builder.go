package llm

import (
	"fmt"
	"os"
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
	"bug_fix":       "bug-fix.md",
	"release_cycle": "release-cycle.md",
	"spec_propose":  "spec-propose.md",
	"spec_apply":    "spec-apply.md",
	"spec_archive":  "spec-archive.md",
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
//
// If the locale is non-empty but unknown, a warning is emitted to stderr so the
// caller can detect silent fallbacks during development. The function never
// errors — the LLM still receives a valid language directive.
func outputLanguageName(locale string) string {
	if name, ok := localeLanguageNames[locale]; ok {
		return name
	}
	if locale != "" {
		_, _ = fmt.Fprintf(os.Stderr, "warning: locale %q is not supported, defaulting to English\n", locale)
	}
	return "English"
}

// --- Shared prompt fragments (private helpers) ----------------------------------
//
// These helpers consolidate previously duplicated text across the six prompt
// builders. They produce identical XML blocks regardless of which prompt invokes
// them, so editing the rule body only requires touching one place.

// groundingRulesForGeneratedContent returns the anti-hallucination block used
// by file-generation modes (generate, analyze, spec). It distinguishes
// technical framework choices (LLM may opine) from domain logic (LLM must
// only echo what was stated).
func groundingRulesForGeneratedContent() string {
	return `<grounding_rules>
CRITICAL — Distinguish between two types of content:

1. TECHNICAL FRAMEWORK (you may opine freely): architectural patterns, project
   structure, code conventions, testing strategy, observability, layer
   responsibilities. This is the template's value.

2. DOMAIN LOGIC (only what the user/context stated): business rules, specific
   validations, default values, edge cases, data formats, concrete behaviors,
   error messages.

For domain logic:
- Only include what is EXPLICITLY in the project description or scanned context
- DO NOT invent validation rules, default values, formats, or behaviors
- DO NOT generate speculative edge cases or error scenarios
- If a template section asks for domain details the input does not cover,
  mark "[DEFINE: short hint of what is needed]" instead of inventing an answer
</grounding_rules>`
}

// personalizationGroundingRules returns the anti-hallucination block used by
// personalized modes (skills, workflows, workflow-skills). The domain string
// is interpolated into the rule headline (e.g. "skill", "workflow",
// "Claude Code skill") so the rule reads naturally for the target artifact.
func personalizationGroundingRules(domain string) string {
	return fmt.Sprintf(`<grounding_rules>
For domain-specific knowledge in this %s:
- Only reference tools, patterns, frameworks, and conventions present in the project context
- DO NOT invent stack components not mentioned (libraries, services, infra, providers)
- DO NOT assume API endpoints, environment variables, or configuration values
- DO NOT fabricate command names, file paths, or directory layouts
- If required knowledge is missing, mark "[DEFINE: short hint of what is needed]"
  rather than inserting a plausible-looking placeholder
</grounding_rules>`, domain)
}

// commonOutputRules returns the closing rules block, always last in every
// system prompt. It enforces the response shape (raw markdown, no wrappers,
// no commentary) and the output language.
func commonOutputRules(locale string) string {
	return fmt.Sprintf(`<rules>
- Respond ONLY with the requested content (markdown body or full file)
- DO NOT wrap the response in code blocks
- DO NOT add explanations before or after the content
- Content must be in %s
</rules>`, outputLanguageName(locale))
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

%s

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
- Use the template guide as structural reference, NOT as a variable replacement template
</output_quality>

%s`, FileOutputName(guideName), groundingRulesForGeneratedContent(), commonOutputRules(locale))
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
- Use the template guide as structural reference, NOT as a variable replacement template
</output_quality>

%s`, FileOutputName(guideName), commonOutputRules(locale))
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

%s

<output_example>
Reference shape (do not copy verbatim — adapt every section to the user's actual stack and domain):

---
name: ddd-entity
description: Design rich domain entities for the order module using Go and gorm
---

# DDD Entity — Order aggregate

## When to use
- Adding a new aggregate root inside ` + "`internal/domain/order/`" + `
- Splitting an existing entity that has grown too many responsibilities
- Encapsulating invariants currently scattered across services

## Process
1. Identify the invariants the aggregate must protect (e.g. ` + "`Total >= 0`" + `, status transitions)
2. Model identity as a value object (` + "`OrderID`" + `) — never expose primitives
3. Place behavior on the entity, not on a service: ` + "`o.Confirm()`" + `, ` + "`o.Cancel(reason)`" + `
4. Validate every state transition before mutating fields
5. Emit a domain event per state change (` + "`OrderConfirmed`" + `, ` + "`OrderCancelled`" + `)

## Example
` + "```go" + `
type Order struct {
    id        OrderID
    status    OrderStatus
    total     Money
    events    []DomainEvent
}

func (o *Order) Confirm() error {
    if o.status != StatusDraft {
        return ErrInvalidTransition
    }
    o.status = StatusConfirmed
    o.events = append(o.events, OrderConfirmed{ID: o.id})
    return nil
}
` + "```" + `

## Anti-patterns
- Anemic models: data-only structs with all logic in services
- Public setters that bypass invariants
- Direct mutation of related aggregates from this entity

## Verification
- [ ] Identity is a value object, not a primitive
- [ ] Every state transition validates and emits an event
- [ ] Tests cover both happy path and invalid transitions
</output_example>

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
</rules>`, skillName, projectContext, ecosystemDesc, personalizationGroundingRules("skill"), outputLanguageName(locale))
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

%s

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
- Base ALL content on the existing context provided
</output_quality>

%s`, existingContext, groundingRulesForGeneratedContent(), commonOutputRules(locale))
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

%s

<output_example>
Reference shape (do not copy verbatim — adapt commands and tooling to the user's actual stack):

---
description: Run unit + integration tests with coverage for the Go monorepo and surface failures fast
---

# Run full test suite

1. **Format and vet first**
   ` + "```bash" + `
   gofmt -l . && go vet ./...
   ` + "```" + `
   // turbo

2. **Unit tests with coverage**
   ` + "```bash" + `
   go test -count=1 -race -coverprofile=coverage.out ./...
   ` + "```" + `
   // turbo
   // capture: COVERAGE_OUT

3. **Integration tests (requires docker compose up)**
   // if docker compose ps --status=running | grep -q postgres
   ` + "```bash" + `
   go test -tags=integration ./tests/integration/...
   ` + "```" + `

4. **Coverage threshold gate**
   ` + "```bash" + `
   go tool cover -func={{COVERAGE_OUT}} | tail -1
   ` + "```" + `
   // turbo

5. **Report**
   Summarize: total packages tested, %% coverage, any failing tests with file:line references.
</output_example>

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
</rules>`, workflowName, projectContext, personalizationGroundingRules("workflow"), outputLanguageName(locale))
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

// BuildWorkflowSkillSystemPrompt returns a system prompt for generating a Claude Code native SKILL.md.
// The LLM generates a complete skill file with frontmatter and personalized workflow instructions.
func (b *PromptBuilder) BuildWorkflowSkillSystemPrompt(workflowName, locale, projectContext string) string {
	return fmt.Sprintf(`<role>
You are a senior DevOps engineer and workflow automation specialist for Claude Code.
Your task is to generate a native Claude Code skill (SKILL.md) that guides an AI agent
through a multi-step workflow using natural instructions.
</role>

<task>
Generate a complete, production-ready SKILL.md for the workflow: %s.
This skill must be PERSONALIZED to the user's project context provided below.
The output must include proper YAML frontmatter.
</task>

<project_context>
%s
</project_context>

<skill_format>
The SKILL.md file MUST follow this structure:

1. YAML frontmatter (between --- markers) with these REQUIRED fields:
   - name: workflow-name (kebab-case)
   - description: Short description of what this workflow does (max 250 chars)
   - disable-model-invocation: true
   - allowed-tools: space-separated list of tools Claude can use without permission prompts
     (e.g., "Bash(git *) Bash(npm *) Bash(go test *)")

   Optional frontmatter fields (include when applicable):
   - agent: auto | manual (default manual; use auto only for fully safe, idempotent flows)
   - user-invocable: true | false (default true; set false if intended only as a nested helper
     invoked by other skills/workflows)
   - context: list of variables this workflow expects from the conversation (e.g., [BRANCH_NAME, TICKET_ID])

2. Numbered steps in markdown with clear, actionable instructions
3. Code blocks with exact commands the agent should run
4. Conditional logic as natural prose ("If X, then Y; otherwise skip this step")

CRITICAL:
- DO NOT include Antigravity execution annotations (// turbo, // capture:, // if, // parallel, etc.)
- The allowed-tools frontmatter field handles tool auto-approval — do not reference it in step prose
- Write all instructions as natural language that Claude can follow directly
- Each step should be self-contained and actionable
</skill_format>

<personalization_rules>
CRITICAL — Adapt this workflow to the project context:

1. Use the project's actual tools, frameworks, and commands (not generic placeholders)
2. Reference the project's branch naming conventions, CI/CD tools, and testing frameworks
3. Adapt file paths and directory structures to match the project's layout
4. Include project-specific considerations (monorepo vs single repo, deployment targets, etc.)
5. Use the project's package manager, build tools, and test runners in commands
6. Set allowed-tools to match the project's actual toolchain (e.g., Bash(go *) for Go projects)

DO NOT:
- Use generic commands when the project context provides specific tools
- Include steps irrelevant to the project's tech stack
- Assume tools or services not mentioned in the project context
</personalization_rules>

%s

<output_example>
Reference shape (do not copy verbatim — adapt to the user's actual stack):

---
name: release-cycle
description: Cut a tagged release for the Go CLI, update changelog, and push to origin
disable-model-invocation: true
allowed-tools: Bash(git *) Bash(go *) Bash(gh *) Read Edit
agent: manual
user-invocable: true
context: [VERSION, RELEASE_NOTES]
---

# Release cycle

1. **Verify clean working tree**
   Run ` + "`git status --short`" + `. If any files are listed, stop and ask the user how to proceed.

2. **Run the full test suite**
   Execute ` + "`go test -count=1 -race ./...`" + `. If any test fails, stop and report the first failing package.

3. **Bump version**
   Edit ` + "`internal/version/version.go`" + ` and replace the constant with ` + "`{{VERSION}}`" + `.

4. **Update CHANGELOG.md**
   Prepend a new section dated today with ` + "`{{RELEASE_NOTES}}`" + `, grouped by Added / Changed / Fixed.

5. **Commit and tag**
   ` + "```bash" + `
   git add internal/version/version.go CHANGELOG.md
   git commit -m "chore(release): v{{VERSION}}"
   git tag -a v{{VERSION}} -m "v{{VERSION}}"
   ` + "```" + `

6. **Push branch and tag**
   ` + "```bash" + `
   git push origin HEAD && git push origin v{{VERSION}}
   ` + "```" + `

7. **Open GitHub release**
   ` + "```bash" + `
   gh release create v{{VERSION}} --notes-file CHANGELOG.md --title "v{{VERSION}}"
   ` + "```" + `
</output_example>

<output_quality>
- YAML frontmatter with name, description, disable-model-invocation, allowed-tools (plus optional agent / user-invocable / context when applicable)
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
</rules>`, workflowName, projectContext, personalizationGroundingRules("Claude Code skill"), outputLanguageName(locale))
}

// BuildWorkflowSkillUserMessage constructs the user message for generating a native Claude Code SKILL.md.
func (b *PromptBuilder) BuildWorkflowSkillUserMessage(guide service.TemplateGuide) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<workflow_name>%s</workflow_name>\n\n", guide.Name))
	sb.WriteString("<template_guide>\n")
	sb.WriteString(guide.Content)
	sb.WriteString("\n</template_guide>\n")

	return sb.String()
}