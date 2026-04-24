# Interaction Log - Codify

## Guiding Principles
This log documents the evolution of the project, serving as its institutional memory. It captures key decisions, problems solved, and lessons learned. New entries should follow the established format.

## Session Format

```markdown
## Session: YYYY-MM-DD - [Descriptive Title]

### Context
- **Version:** [x.y.z]
- **Goal:** [What to resolve or implement]

### Tasks Completed
1. **[Title]** - [Brief description, affected files, result]

### Architectural Decisions
- **ADR-XXX: [Title]**
  - **Decision:** [What was decided]
  - **Reason:** [Justification and trade-offs]
```

---

## Session: 2026-03-27 - v1.13.1: Rename Skill Category "workflow" to "conventions"

### Context
- **Version:** 1.13.1
- **Goal:** Eliminate naming ambiguity between the skill category "workflow" (Conventional Commits, Semantic Versioning) and the new `workflows` command (Antigravity orchestration).

### Tasks Completed
1. **Skill category rename** — Renamed `workflow` → `conventions` across catalog, tests, CLI, MCP, READMEs, and template directories (`templates/{locale}/skills/workflow/` → `conventions/`).
2. **Backward compatibility** — Added legacy mapping `"workflow" → {"conventions", "all"}` so old `--category workflow` still works.
3. **Version reference sync** — Updated badges, project status, and MCP server version to 1.13.1.

### Architectural Decisions

#### ADR-025: Rename Skill Category "workflow" to "conventions"
- **Decision:** The skill category previously named "workflow" is now "conventions". Legacy mapping preserved.
- **Reason:** Conventional Commits and Semantic Versioning are conventions/standards (declarative knowledge), not orchestration workflows (procedural steps). The new `workflows` command introduced ambiguity that needed resolution.

---

## Session: 2026-03-25 - v1.13.0: Antigravity Workflows Command

### Context
- **Version:** 1.13.0
- **Goal:** Add native workflow generation for Antigravity IDE — multi-step recipes with execution annotations.

### Tasks Completed
1. **Workflow catalog** — New bounded context in `internal/domain/catalog/workflow_catalog.go` with `WorkflowCategories`, `WorkflowMetadata`, `WorkflowMeta`, frontmatter generation, description validation (max 250 chars).
2. **BDD test suite** — `tests/bdd/workflow_catalog/` with 11 scenarios, 43 steps covering: find category, resolve presets, resolve "all", unknown preset error, frontmatter generation, description length, category names.
3. **Application commands** — `DeliverStaticWorkflowsCommand` and `GenerateWorkflowsCommand` with `WorkflowConfig` DTO.
4. **Workflow templates** — 3 presets (feature-development, bug-fix, release-cycle) for both en/es locales. Templates include Antigravity execution annotations.
5. **CLI command** — `workflows` command with interactive UX, install scopes (global/project), mode selection (static/personalized).
6. **MCP tool** — `generate_workflows` tool with handler in `server.go`.
7. **LLM integration** — `BuildPersonalizedWorkflowsSystemPrompt()` and `BuildWorkflowsUserMessage()` in prompt builder. `case "workflows"` in both providers.

### Architectural Decisions

#### ADR-023: Workflow Catalog as Separate Bounded Context
- **Decision:** The workflow catalog (`WorkflowCategories`) is a separate bounded context from the skills catalog, even though both share the same structural types (`SkillCategory`, `SkillOption`, `ResolvedSelection`).
- **Reason:** Skills = expertise (how to do X), Workflows = orchestration (do A then B then C). Different domain concepts, different metadata registries, different output formats (SKILL.md vs flat .md).

#### ADR-024: Multi-Target Workflow Strategy
- **Decision:** Workflows support both Claude Code (native skills with SKILL.md frontmatter) and Antigravity (native .md with execution annotations).
- **Reason:** Claude Code workflows are generated as native skills with frontmatter (`name`, `description`, `disable-model-invocation`, `allowed-tools`). Antigravity annotations are stripped and translated to prose instructions. Both targets share the same workflow catalog and template content.

---

## Session: 2026-03-25 - v1.12.0: Testing Skill Category

### Context
- **Version:** 1.12.0
- **Goal:** Add a third skill category "testing" with foundational, TDD, and BDD presets.

### Tasks Completed
1. **Testing templates** — `foundational.template` (Kent Beck's Test Desiderata — 12 properties as trade-offs), `tdd.template` (Part 1: Desiderata + Part 2: TDD discipline), `bdd.template` (Part 1: Desiderata + Part 2: BDD practice). Both en/es locales.
2. **Catalog update** — Added `testing` category with 3 exclusive presets + `SkillMetadata` entries.
3. **TDD/BDD include foundational** — Templates structured so TDD and BDD presets include the foundational Test Desiderata as Part 1, with their specific methodology as Part 2. UX labels show "(includes foundational)".

### Architectural Decisions

#### ADR-022: Inclusive Testing Presets
- **Decision:** TDD and BDD presets embed the foundational Test Desiderata content as Part 1, rather than requiring users to install foundational separately.
- **Reason:** Better UX — users selecting TDD or BDD get the complete testing philosophy without needing to know about the foundational preset. Labels clarify the inclusion.

---

## Session: 2026-03-20 - v1.11.0: Unified Interactive UX and Skills Install

### Context
- **Version:** 1.11.0
- **Goal:** (1) All commands prompt interactively for missing parameters. (2) Add `--install` flag with `global`/`project` scopes.

### Tasks Completed
1. **Unified interactive prompts** — All commands (`generate`, `analyze`, `spec`, `skills`) detect TTY and prompt for missing flags using `charmbracelet/huh`. Shared helpers in `interactive.go`.
2. **`--install` flag** — `global` installs to agent's home path (`~/.claude/skills/`, `~/.codex/skills/`), `project` installs to current directory. Ecosystem-aware path resolution.
3. **DefaultModel bug fix** — Fixed `DefaultModel()` always returning claude default even when another model was explicitly provided.
4. **CLI comments in English** — Translated all Spanish comments in CLI/DTO files.

### Architectural Decisions

#### ADR-021: Explicit Flag Detection Pattern
- **Decision:** Use `cmd.Flags().Visit()` with an `explicit` map to distinguish user-provided flags from defaults, enabling interactive prompting only for truly unspecified parameters.
- **Reason:** Cobra doesn't distinguish between "user passed --model claude" and "default value is claude". The Visit pattern tracks which flags were explicitly set.

---

## Session: 2026-03-18 - v1.10.0: Dual-Mode Skills (Static + Personalized)

### Context
- **Version:** 1.10.0
- **Goal:** Add personalized mode to skills — LLM adapts skill content to user's specific project.

### Tasks Completed
1. **Personalized mode** — `GenerateSkillsCommand` sends skill templates as guides to LLM along with project context. LLM generates adapted SKILL.md files.
2. **Static mode** — `DeliverStaticSkillsCommand` delivers pre-built templates with ecosystem frontmatter. No API key needed.
3. **`SkillsConfig` DTO** — Added `Mode`, `ProjectContext` fields with validation.
4. **Prompt builder** — Added `BuildSkillsSystemPrompt()` and `BuildSkillsUserMessage()`.
5. **Both providers** — Added `case "skills"` in `generateSingleFile` for AnthropicProvider and GeminiProvider.

---

## Session: 2026-03-16 - v1.8.0-v1.9.0: Interactive Skill Categorization

### Context
- **Version:** 1.8.0 → 1.9.0
- **Goal:** Replace flat skill presets with a categorized catalog and interactive selection.

### Tasks Completed
1. **Declarative catalog** — `skills_catalog.go` with `SkillCategory`, `SkillOption`, `ResolvedSelection`, `SkillMetadata` types. Categories: `architecture` (exclusive), `workflow` (non-exclusive). Metadata registry for ecosystem frontmatter.
2. **Interactive prompts** — Skills command prompts for category → preset → mode → target → install → locale using `charmbracelet/huh`.
3. **Extended interactivity** — All skills configuration options accessible interactively.

### Architectural Decisions

#### ADR-020: Declarative Catalog Pattern
- **Decision:** Skills are defined as in-code data structures (`SkillCategory` slices with `SkillOption` entries and `SkillMetadata` maps), not YAML/JSON config files.
- **Reason:** Compile-time safety, IDE autocompletion, easy to extend, no file parsing. The catalog is small enough that code-as-config is simpler than a config system.

---

## Session: 2026-03-15 - v1.7.0: Workflow Skills + MCP Knowledge Tools

### Context
- **Version:** 1.7.0
- **Goal:** Add workflow skill preset (Conventional Commits, Semantic Versioning) and MCP knowledge tools.

### Tasks Completed
1. **Workflow skill preset** — `conventional_commit.template` and `semantic_versioning.template` in both locales. Category: `workflow` (now renamed to `conventions`).
2. **MCP knowledge tools** — `commit_guidance` and `version_guidance` tools that load embedded templates and return behavioral context. No API key needed.
3. **Knowledge tool pattern** — `loadKnowledgeTemplate()` helper reads from `embed.FS` and returns content as MCP tool result.

---

## Session: 2026-03-15 - v1.6.0: Output Defaults and Ecosystem Paths

### Context
- **Version:** 1.6.0
- **Goal:** Output to current directory by default and set ecosystem-aware paths for skills.

### Tasks Completed
1. **Default output** — `generate` command now outputs to current directory instead of `output/` subdirectory.
2. **Ecosystem paths** — Skills output paths: `.claude/skills/` for claude, `.agents/skills/` for codex/antigravity.

---

## Session: 2026-03-14 - v1.5.0: Rebrand to Codify

### Context
- **Version:** 1.5.0
- **Goal:** Rebrand from `ai-context-generator` to `Codify`.

### Tasks Completed
1. **Module rename** — `github.com/jorelcb/codify` across all Go files.
2. **Binary rename** — `cmd/codify/` as entry point.
3. **CLI rename** — All references updated to `codify`.

---

## Session: 2026-03-14 - v1.4.0: Embedded Templates, GoReleaser, Homebrew

### Context
- **Version:** 1.4.0
- **Goal:** Embed templates in binary, add cross-compilation, Homebrew distribution.

### Tasks Completed
1. **Embedded templates** — `embed.FS` for templates directory. Binary works from any location.
2. **GoReleaser v2** — Cross-compilation (macOS/Linux, arm64/amd64).
3. **Homebrew tap** — `brew tap jorelcb/tap && brew install codify`.
4. **Auto-detect provider** — When `--model` not specified, selects provider based on available API key.
5. **CI/CD** — GitHub Actions for CI (tests) and release (on tag push).

---

## Session: 2026-03-12 - v1.3.0: Agent Skills Generation

### Context
- **Version:** 1.3.0
- **Goal:** Generate reusable Agent Skills (SKILL.md files) for AI coding agents.

### Tasks Completed
1. **Skills command** — `codify skills` with `--target claude|codex|antigravity`, ecosystem-specific YAML frontmatter.
2. **Default preset** — DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port.
3. **Neutral preset** — Code review, test strategy, safe refactoring, API design.
4. **MCP tool** — `generate_skills` MCP tool.

### Architectural Decisions

#### ADR-018: Skills as Separate Artefact Type
- **Decision:** Skills are a distinct artefact category alongside Context and Specs. They have their own command, templates, and output structure.
- **Reason:** Skills serve a different purpose (teaching agents HOW vs WHAT). They have their own lifecycle, ecosystem-specific formatting, and distribution model.

#### ADR-019: Multi-Ecosystem Frontmatter
- **Decision:** Each target ecosystem gets specific YAML frontmatter: Claude (`name`, `description`, `user-invocable`), Codex (`name`, `description`), Antigravity (`name`, `description`, `triggers`).
- **Reason:** Each AI agent platform has different discovery and activation mechanisms. Generic frontmatter would miss platform-specific features.

---

## Session: 2026-03-11 - v1.2.0: Multi-Provider LLM, MCP Server, Analyze Command

### Context
- **Version:** 1.2.0
- **Goal:** Add Gemini provider, MCP server, and project analysis command.

### Tasks Completed
1. **Gemini provider** — `GeminiProvider` with streaming via Google GenAI SDK.
2. **Provider factory** — `llm.NewProvider()` resolves by model prefix.
3. **MCP server** — `serve` command with stdio and HTTP transport strategy.
4. **Analyze command** — `ProjectScanner` detects language, framework (20+), dependencies.
5. **`--with-specs` flag** — Chains context + spec generation in one command.

---

## Session: 2026-03-06 - v1.1.0: Output Validation and Consistency Fixes

### Context
- **Version:** 1.1.0
- **Goal:** Validate generated output for a sample project and correct inconsistencies. Add locale and from-file support.

### Tasks Completed
1. **Spec emulation** — Generated specs for test project, identified 6 gaps.
2. **Consistency fixes** — Unified entity names, standardized timestamps, aligned JSON fields across 9 files.
3. **Locale support** — `--locale en|es` flag. Templates reorganized into `templates/{locale}/`.
4. **`--from-file` feature** — `-f` flag reads description from file (CLI-layer only).
5. **BDD test fixes** — Fixed repository save logic, unique names in test steps. 14 scenarios, 72 steps green.

### Architectural Decisions

#### ADR-015: `--from-file` as CLI-Only Feature
- **Decision:** File reading is confined to `interfaces/cli`. Application layer receives a simple string.
- **Reason:** Application is agnostic to input source, upholding separation of concerns.

#### ADR-016: Output Validation as Mandatory Workflow Step
- **Decision:** Always validate LLM-generated output against user's business specifications.
- **Reason:** LLMs can omit details or introduce subtle inconsistencies between files.

---

## Session: 2026-02-19 - v1.0.0: Restructure to AGENTS.md Standard & Spec Command

### Context
- **Version:** 1.0.0
- **Goal:** Align output with AGENTS.md standard and introduce spec command.

### Tasks Completed
1. **Output restructuring** — `AGENTS.md` as root file, details in `context/` directory.
2. **Spec command** — `spec <name> --from-context <path>` generates SDD specs.
3. **XML tags in prompts** — Switched from Markdown to XML for better LLM parsing.

### Architectural Decisions

#### ADR-011: Adopt AGENTS.md Standard
- **Decision:** Use `AGENTS.md` as main entry point (Linux Foundation standard).

#### ADR-013: Spec Generation is Context-Dependent
- **Decision:** `spec` command must operate on pre-existing context for coherence.

#### ADR-014: XML Tags in System Prompts
- **Decision:** Structure system prompts with XML tags for better Claude parsing.

---

## Session: 2026-02-19 - v0.1.0: Initial Release

### Context
- **Version:** 0.1.0
- **Goal:** Core functionality — generate context files via Anthropic Claude API.

### Architectural Decisions

#### Per-File Generation vs. Single JSON Blob
- **Decision:** Separate API calls per file, each returning pure Markdown.
- **Reason:** Avoids truncation and JSON parsing failures. Granular progress feedback.

#### Streaming API by Default
- **Decision:** Use streaming API for all LLM calls.
- **Reason:** Prevents gateway timeouts, better UX with real-time progress.

---
**CRITICAL:** This log serves as the project's memory. Review recent ADRs before making significant architectural changes to maintain consistency.