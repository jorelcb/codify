# Changelog - Codify

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.15.0] - 2026-04-13 - Claude workflow plugin generation

### Added
- Claude Code plugin generation: `--target claude` now produces complete plugin packages instead of SKILL.md files
- Plugin structure per workflow: `.claude-plugin/plugin.json`, `skills/`, `hooks/`, `agents/`, `scripts/`
- `AnnotationParser` — parses Antigravity execution annotations (`// turbo`, `// capture:`, `// if`) from workflow templates
- `PluginGenerator` — generates plugin manifest, hooks, skills, and agents from parsed annotations
- Annotation-to-hook mapping: `turbo` → `PreToolUse` auto-approve, `capture` → `PostToolUse` script, `if` → `PreToolUse` prompt
- `DeliverPluginCommand` — static mode plugin delivery (no API key needed)
- `GeneratePluginCommand` — personalized mode with LLM-generated SKILL.md (hooks/agents remain static)
- `BuildPluginSkillSystemPrompt()` — LLM prompt for plugin-aware skill generation
- `templates/scripts/capture-output.sh` — embedded hook script for PostToolUse output capture
- `workflow-runner.md` agent per plugin with locale-aware content (en/es)
- Plugin mode routing in both LLM providers (Anthropic + Gemini)
- MCP server routes Claude target to plugin commands
- BDD: 9 new scenarios for plugin generation (23 total workflow scenarios, 103 steps)
- Unit tests: 8 annotation parser tests + 13 plugin generator tests

### Changed
- Claude target output: plugin packages (`codify-wf-{preset}/`) replace `{workflow}/SKILL.md` subdirectories
- Claude install paths: `~/.claude/plugins/` (global), `.` (project) — previously `~/.claude/skills/`
- Interactive target prompt: "Claude Code (via plugin: skills + hooks + agents)" replaces "Claude Code (SKILL.md workflows)"
- `BuildClaudeWorkflowSystemPrompt` replaced by `BuildPluginSkillSystemPrompt`
- Workflows CLI help text updated for plugin structure
- READMEs rewritten: workflows section reflects plugin packages for Claude target

## [1.14.0] - 2026-03-27 - Multi-target workflows (Claude Code + Antigravity)

### Added
- `--target` flag on `workflows` command: `claude` (SKILL.md with prose instructions) or `antigravity` (native .md with execution annotations)
- `GenerateWorkflowFrontmatter(name, target)` — target-aware YAML frontmatter generation
- `BuildClaudeWorkflowSystemPrompt()` — LLM prompt with annotation-to-prose translation table
- Claude target output: `{workflow}/SKILL.md` in subdirectories with `user-invocable: true` frontmatter
- `Target` field in `WorkflowConfig` DTO with `ValidWorkflowTargets` validation
- `target` parameter on `generate_workflows` MCP tool
- Interactive target ecosystem prompt in workflows CLI
- Target-aware install paths: Claude → `.claude/skills/`, Antigravity → `.agent/workflows/`
- BDD: 3 new scenarios for Claude frontmatter (14 scenarios, 59 steps total)
- Unit tests: `TestGenerateWorkflowFrontmatter_Claude`, `TestGenerateWorkflowFrontmatter_UnknownClaude`

### Changed
- Workflows section in READMEs rewritten for multi-target support
- Hero/tagline updated to 4 pillars: Context, Specs, Skills, Workflows
- Go badge corrected from 1.21+ to 1.23+
- FAQ expanded with Skills vs Workflows, workflow API key, ecosystem questions

## [1.13.1] - 2026-03-27 - Rename skill category "workflow" to "conventions"

### Changed
- Renamed skill category `workflow` to `conventions` to eliminate naming ambiguity with the `workflows` command (Antigravity orchestration)
- Template directories renamed: `templates/{locale}/skills/workflow/` → `conventions/`
- Legacy mapping preserved: `--category workflow` still resolves to `conventions`

### Fixed
- Version references in READMEs and MCP server updated to 1.13.1

## [1.13.0] - 2026-03-25 - Antigravity Workflows command

### Added
- `workflows` command: generates multi-step Antigravity workflow files with execution annotations
- Workflow catalog (`internal/domain/catalog/workflow_catalog.go`) as separate bounded context from skills
- Three workflow presets: `feature-development`, `bug-fix`, `release-cycle`
- Workflow templates for both locales (en/es) with Antigravity annotations (`// turbo`, `// parallel`, `// capture`, `// if`)
- `WorkflowConfig` DTO with validation
- `DeliverStaticWorkflowsCommand` and `GenerateWorkflowsCommand` in application layer
- `BuildPersonalizedWorkflowsSystemPrompt()` and `BuildWorkflowsUserMessage()` in PromptBuilder
- `generate_workflows` MCP tool
- BDD test suite: `tests/bdd/workflow_catalog/` with 11 scenarios, 43 steps
- Interactive UX for workflows (preset, mode, locale, install scope)
- Install scopes: `global` (`~/.gemini/antigravity/global_workflows/`) and `project` (`.agent/workflows/`)

## [1.12.0] - 2026-03-25 - Testing skill category

### Added
- Testing skill category with 3 exclusive presets: `foundational`, `tdd`, `bdd`
- `foundational.template`: Kent Beck's Test Desiderata — 12 properties of good tests as trade-offs
- `tdd.template`: Part 1 (Desiderata) + Part 2 (Red-Green-Refactor, Three Laws, strategies)
- `bdd.template`: Part 1 (Desiderata) + Part 2 (Discovery/Formulation/Automation, Given/When/Then, Gherkin)
- TDD and BDD presets include foundational content as Part 1 (UX labels show "includes foundational")
- Templates for both locales (en/es)

## [1.11.0] - 2026-03-20 - Unified interactive UX and skills --install

### Added
- Unified interactive prompts: all commands (`generate`, `analyze`, `spec`, `skills`) prompt for missing flags when run in a terminal
- `--install` flag on `skills` command: `global` (agent home path) or `project` (current directory)
- Shared interactive helpers in `internal/interfaces/cli/commands/interactive.go` (charmbracelet/huh)
- `cmd.Flags().Visit()` pattern with explicit flag map to distinguish user-provided from defaults

### Fixed
- `DefaultModel()` bug: always returned Claude default even when another model was explicitly provided
- Translated all Spanish comments in CLI/DTO files to English

## [1.10.0] - 2026-03-18 - Dual-mode skills (static + personalized)

### Added
- Personalized skills mode: LLM adapts skill content to user's specific project context
- Static skills mode: delivers pre-built templates with ecosystem frontmatter (no API key needed)
- `GenerateSkillsCommand` for personalized mode, `DeliverStaticSkillsCommand` for static mode
- `BuildSkillsSystemPrompt()` and `BuildSkillsUserMessage()` in PromptBuilder
- `Mode` and `ProjectContext` fields in `SkillsConfig` DTO
- Skills generation mode (`"skills"`) in both AnthropicProvider and GeminiProvider

## [1.9.0] - 2026-03-16 - Extended interactive skill prompts

### Added
- All skills configuration options accessible via interactive menus (category, preset, mode, target, install, locale, model, project context)

## [1.8.0] - 2026-03-16 - Interactive skill categorization with catalog registry

### Added
- Declarative skill catalog (`internal/domain/catalog/skills_catalog.go`) with `SkillCategory`, `SkillOption`, `ResolvedSelection`, `SkillMetadata`
- Two categories: `architecture` (exclusive) and `workflow` (non-exclusive, renamed to `conventions` in v1.13.1)
- Interactive category → preset selection using charmbracelet/huh
- `SkillMetadata` registry for ecosystem-specific frontmatter (name, description, triggers)

## [1.7.1] - 2026-03-15 - Homebrew formula fix

### Fixed
- Switch from Homebrew cask to formula for proper macOS quarantine handling

## [1.7.0] - 2026-03-15 - Workflow skills preset and MCP knowledge tools

### Added
- Workflow skill preset: `conventional_commit.template` and `semantic_versioning.template` (both locales)
- MCP knowledge tools: `commit_guidance` and `version_guidance` — load embedded templates, no API key needed
- `loadKnowledgeTemplate()` helper for direct template content delivery

## [1.6.0] - 2026-03-15 - Output defaults and ecosystem paths

### Changed
- `generate` command now outputs to current directory by default (instead of `output/` subdirectory)
- Skills output paths: `.claude/skills/` for claude, `.agents/skills/` for codex/antigravity

## [1.5.0] - 2026-03-14 - Rebrand to Codify

### Changed
- Module renamed from `ai-context-generator` to `github.com/jorelcb/codify`
- Binary entry point: `cmd/codify/`
- All CLI references updated to `codify`

## [1.4.0] - 2026-03-14 - Embedded templates, auto-detect provider, Homebrew distribution

### Added
- GoReleaser v2 config with cross-compilation (macOS/Linux, arm64/amd64)
- GitHub Actions: CI (tests on push/PR) and release (on tag push)
- Homebrew tap distribution (`brew tap jorelcb/tap && brew install codify`)
- Embedded templates via `embed.FS` — binary works from any directory
- Auto-detect LLM provider from available API keys when `--model` is not specified

### Fixed
- Templates not found when running installed binary outside project root
- `ANTHROPIC_API_KEY environment variable is required` error when only Gemini key was set
- Version/ldflags not injected in CLI `--version` output

## [1.3.0] - 2026-03-12 - Agent Skills generation

### Added
- `skills` command: generates reusable Agent Skills (SKILL.md) based on architectural presets
- Multi-ecosystem support: `--target claude|codex|antigravity` with ecosystem-specific YAML frontmatter
- Default preset skills: DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port
- Neutral preset skills: code review, test strategy, safe refactoring, API design
- Skill templates for both locales (en/es)
- `GenerateSkillsCommand`, `SkillsConfig` DTO, `BuildSkillsSystemPrompt()`
- `generate_skills` MCP tool

## [1.2.0] - 2026-03-11 - Multi-provider LLM, MCP server, analyze command, HTTP transport

### Added
- **Gemini LLM provider**: Google Gemini API with streaming via `google.golang.org/genai` SDK v1.49.0
- **Provider factory**: `llm.NewProvider()` resolves by model prefix (`gemini-*` → Gemini, else → Anthropic)
- **MCP Server mode**: `serve` command with stdio + HTTP transport strategy
- MCP tools: `generate_context`, `generate_specs`, `analyze_project`
- `analyze` command: scans existing projects (language, framework, 20+ framework detection, config signals)
- `--with-specs` flag on `generate` and `analyze`
- `mcp-go` v0.45.0 dependency

## [1.1.0] - 2026-03-06 - Locale support, anti-hallucination, legacy cleanup

### Added
- Multi-locale support: `--locale en|es` flag
- Templates reorganized into `templates/{locale}/{preset}/`
- Language-specific `idioms.template` files (Go, JavaScript, Python)
- `<grounding_rules>` in system prompts with `[DEFINE]` markers
- `--from-file` / `-f` flag on generate command

### Changed
- System prompts rewritten in English with locale-controlled output

### Removed
- Legacy template directories, bash tests, unused domain/template layer

## [1.0.0] - 2026-02-19 - First stable release (AGENTS.md standard + spec command)

### Added
- `spec` command: generates SDD specifications from existing context
- `agents.template`: root file following AGENTS.md standard
- XML tags in system prompts

### Changed
- Output restructured: AGENTS.md at root, details in context/

## [0.2.0] - 2026-02-19 - DDD architecture with CLI

### Added
- Full DDD/Clean Architecture implementation
- CLI with Cobra (generate, list commands)
- Template system with configurable loader

## [0.1.0] - 2026-02-19 - Initial alpha

### Added
- Context file generation using Anthropic Claude API with streaming
- Per-file generation (independent API calls per output file)
- AnthropicProvider with official SDK
- Value objects with validation
- Unit tests for all components