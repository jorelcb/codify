# Changelog - AI Context Generator

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-03-14 - Embedded templates, auto-detect provider, Homebrew distribution

### Added
- GoReleaser v2 config with cross-compilation (macOS/Linux, arm64/amd64)
- GitHub Actions: CI (tests on push/PR) and release (on tag push)
- Homebrew tap distribution (`brew tap jorelcb/tap && brew install ai-context-generator`)
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
- Default preset skills: DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port/adapter
- Neutral preset skills: code review, test strategy, safe refactoring, API design
- Skill templates for both locales (en/es) in `templates/{locale}/skills/{preset}/`
- `GenerateSkillsCommand` in application layer
- `SkillsConfig` DTO with target ecosystem validation
- `BuildSkillsSystemPrompt()` and `BuildSkillsUserMessage()` in PromptBuilder
- `generate_skills` MCP tool
- `Target` field in `GenerationRequest` for skills mode
- Skills mode (`"skills"`) in both AnthropicProvider and GeminiProvider

## [1.2.0] - 2026-03-11 - Multi-provider LLM, MCP server, analyze command, HTTP transport

### Added
- **Gemini LLM provider** (`gemini_provider.go`): Google Gemini API with streaming via `google.golang.org/genai` SDK
- **Provider factory** (`provider_factory.go`): resolves provider by model prefix (`gemini-*` → Gemini, else → Anthropic)
- Independent API key resolution: `ANTHROPIC_API_KEY` for Claude, `GEMINI_API_KEY`/`GOOGLE_API_KEY` for Gemini
- Default Gemini model: `gemini-3.1-pro-preview`
- MCP Server mode (`serve` command): exposes tools via Model Context Protocol
- MCP tools: `generate_context`, `generate_specs`, `analyze_project`
- Transport strategy pattern: `StdioTransport` (local) and `HTTPTransport` (remote)
- `--transport` flag (stdio, http) and `--addr` flag for HTTP transport
- `analyze` command: scans existing projects and generates context from structure
- `ProjectScanner` infrastructure: detects language, framework, dependencies, directory tree, README, config signals
- Framework detection for 20+ frameworks (Go, JS, Python, Rust, etc.)
- Config signal detection (GitHub Actions, Docker, Makefile, Terraform, K8s, Helm)
- `--with-specs` flag on `generate` and `analyze`: chains context + spec generation in one command
- `google.golang.org/genai` v1.49.0 dependency
- `mcp-go` v0.45.0 dependency

### Changed
- `generate.go`, `spec.go`, `server.go` refactored to use `llm.NewProvider()` factory instead of direct `NewAnthropicProvider()`
- `serve.go` refactored from switch/case to transport strategy pattern
- `--model` flag now accepts both `claude-*` and `gemini-*` models

## [1.1.0] - 2026-03-06 - Locale support, anti-hallucination, legacy cleanup

### Added
- Multi-locale support: `--locale en|es` flag for both `generate` and `spec` commands
- Templates reorganized into `templates/{locale}/{preset}/` hierarchy
- `development_guide.template`: methodology, testing, security, delivery expectations
- Language-specific `idioms.template` files for Go, JavaScript, Python
- `<grounding_rules>` in system prompts: distinguishes TECHNICAL FRAMEWORK vs DOMAIN LOGIC
- `[DEFINE]` markers for domain details not covered by user input
- `--from-file` / `-f` flag on generate command: read description from file

### Changed
- System prompts rewritten in English (LLM's native language) with locale-controlled output
- Template loader now locale-aware and language-aware

### Removed
- Legacy template directories, bash tests, unused domain/template layer

## [1.0.0] - 2026-02-19 - First stable release (AGENTS.md standard + spec command)

### Added
- `spec` command: generates SDD specifications (CONSTITUTION.md, SPEC.md, PLAN.md, TASKS.md) from existing context
- `agents.template`: root file following AGENTS.md standard
- XML tags in system prompts (`<role>`, `<task>`, `<workflow>`, `<output_quality>`)
- `GenerateSpecCommand`, `SpecConfig`, `ContextReader`
- `BuildSpecSystemPrompt()` with `<existing_context>`

### Changed
- Output restructured: AGENTS.md at root, CONTEXT.md and INTERACTIONS_LOG.md in context/

## [0.2.0] - 2026-02-19 - DDD architecture with CLI

### Added
- Full DDD/Clean Architecture implementation
- CLI with Cobra (generate, list commands)
- Template system with configurable loader

## [0.1.0] - 2026-02-19 - Initial alpha

### Added
- Context file generation using Anthropic Claude API with streaming
- Per-file generation (independent API calls per output file)
- CLI: `ai-context-generator generate <name> --description "..." [--language] [--type] [--architecture] [--model]`
- PromptBuilder, FileSystemTemplateLoader, GenerateContextCommand
- AnthropicProvider with official SDK (`anthropic-sdk-go v1.25.0`)
- Value objects with validation
- Unit tests for all components
