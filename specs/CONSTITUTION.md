# Codify - Project Constitution

## Identity

- **Name:** Codify
- **Purpose:** A Go CLI tool that generates intelligent context files, SDD specifications, agent skills, and orchestration workflows for AI coding agents from project descriptions.
- **Target users:** Software developers and architects who use AI coding agents (Claude Code, Codex CLI, Antigravity/Gemini CLI) and want to equip them with project-specific knowledge and reusable abilities.
- **Value proposition:** Bridges the gap between a high-level project idea and AI-ready development artefacts. Generates 4 types of output: Context (identity), Specs (plan), Skills (expertise), Workflows (orchestration).

## Technology Stack

| Component | Technology | Version | Justification |
|-----------|-----------|---------|---------------|
| Language | Go | 1.23+ | Modern, performant, strong standard library for CLI tools and streaming. |
| CLI Framework | Cobra | latest | De-facto standard for building CLI applications in Go. |
| Interactive UX | charmbracelet/huh | latest | Terminal forms for interactive parameter prompting. |
| Testing (BDD) | Godog | latest | Gherkin-based BDD for domain behavior validation. |
| Testing (Unit) | testify/assert | latest | Readable assertions for table-driven unit tests. |
| LLM (Anthropic) | anthropic-sdk-go | v1.25.0 | Official Go SDK for Claude models with streaming. |
| LLM (Google) | google.golang.org/genai | v1.49.0 | Official Go SDK for Gemini models with streaming. |
| MCP Server | mcp-go | v0.45.0 | Model Context Protocol for AI agent tool exposure. |
| Templates | embed.FS | stdlib | Binary-embedded templates — no external file dependencies. |
| Distribution | GoReleaser v2 | latest | Cross-compilation (macOS/Linux, arm64/amd64) + Homebrew tap. |

## Immutable Principles

Principles that apply to EVERY design and implementation decision:

1. **Strict DDD Layer Separation:** The Domain layer is pure and has no external dependencies. Application orchestrates use cases. Infrastructure provides concrete implementations. Interfaces adapt external inputs.
2. **Clean Architecture Dependency Rule:** Dependencies must always point inwards: `Interfaces` → `Application` → `Domain`. Abstractions are owned by the inner layers.
3. **Explicit Error Handling:** All recoverable errors must be handled and wrapped with context using `%w`. No silent error swallowing or use of `panic` for control flow.
4. **Templates are Guides, Not Deterministic:** Templates are structural guides that inform the LLM's generation — they are NOT rendered with variable replacement. The LLM interprets and expands them.
5. **Test-Driven Domain:** All domain behavior must be validated with BDD scenarios (Godog) before being considered complete.

## Conventions

### Code
- **Naming:** Idiomatic Go conventions (camelCase for internal, PascalCase for exported).
- **File structure:** `internal/{domain,application,infrastructure,interfaces}` — strict layered architecture.
- **Imports:** Grouped: 1. Standard library, 2. External packages, 3. Internal project packages.
- **Error handling:** `fmt.Errorf` with `%w` to wrap errors. Clear chain of responsibility.

### Process
- **Commits:** Conventional Commits standard (`feat:`, `fix:`, `refactor:`, `docs:`, `chore:`). Never include AI credits.
- **Versioning:** Semantic Versioning. feat → minor, fix → patch, breaking → major. Tags on main branch.
- **Version references:** When bumping version, update: README badges, project status, MCP server version, .version file.
- **Testing:** BDD first for domain behavior. Unit tests for components. `go build ./... && go test ./...` before every commit.

## Constraints

### Mandatory
- Strictly adhere to DDD layer separation and the Clean Architecture dependency rule.
- Define interfaces (ports) in `domain` and implement them (adapters) in `infrastructure`.
- Use constructor-based dependency injection exclusively. No global state.
- Write BDD tests for domain behavior and unit tests for components.
- Maintain minimum test coverage: 90% domain, 70% application.
- Templates must be treated as LLM guides — never as deterministic renderable templates.

### Prohibited
- The `domain` layer importing from `application`, `infrastructure`, or `interfaces`.
- Mixing business logic with infrastructure concerns.
- Using `panic` for recoverable errors.
- Committing API keys, secrets, or environment-specific configuration.
- Creating circular dependencies between packages.
- Including AI credits (Co-Authored-By, model mentions) in git commits.

## Approved Dependencies

| Library | Purpose | Version |
|---------|---------|---------|
| `github.com/spf13/cobra` | CLI framework | latest |
| `github.com/anthropics/anthropic-sdk-go` | Anthropic Claude SDK | v1.25.0 |
| `google.golang.org/genai` | Google Gemini SDK | v1.49.0 |
| `github.com/cucumber/godog` | BDD testing framework | latest |
| `github.com/stretchr/testify` | Unit test assertions | latest |
| `github.com/mark3labs/mcp-go` | Model Context Protocol server | v0.45.0 |
| `github.com/charmbracelet/huh` | Interactive terminal forms | latest |