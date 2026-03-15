# Codify - Project Constitution

## Identity

- **Name:** Codify
- **Purpose:** A Go CLI tool that generates intelligent context files for AI agents from a project description.
- **Target users:** [DEFINE: Describe the primary user personas, e.g., developers, architects, AI agents]

## Technology Stack

| Component | Technology | Version | Justification |
|-----------|-----------|---------|---------------|
| Language | Go | 1.21+ | Modern, performant, strong standard library for CLI tools. |
| Framework | Cobra | N/A | De-facto standard for building powerful CLI applications in Go. |
| Testing | Godog, testify/assert | N/A | BDD for high-level features, unit testing for components. |
| Persistence | Filesystem / In-memory | N/A | Output is file-based; no traditional database required. |

## Immutable Principles

Principles that apply to EVERY design and implementation decision:

1. **Strict DDD Layer Separation:** The Domain layer is pure and has no external dependencies. Application orchestrates use cases. Infrastructure provides concrete implementations. Interfaces adapt external inputs.
2. **Clean Architecture Dependency Rule:** Dependencies must always point inwards: `Interfaces` -> `Application` -> `Domain`. Abstractions are owned by the inner layers.
3. **Explicit Error Handling:** All recoverable errors must be handled and wrapped with context using `%w`. No silent error swallowing or use of `panic` for control flow.

## Conventions

### Code
- **Naming:** Idiomatic Go conventions (camelCase for internal, PascalCase for exported).
- **File structure:** Adherence to the layered architecture: `internal/{domain,application,infrastructure,interfaces}`.
- **Imports:** Grouped in order: 1. Standard library, 2. External packages, 3. Internal project packages.
- **Error handling:** Use `fmt.Errorf` with `%w` to wrap errors, providing a clear chain of responsibility.

### Process
- **Commits:** Conventional Commits standard (`feat:`, `fix:`, `docs:`, etc.).
- **Branching:** GitFlow model (`main`, `develop`, `feature/*`, `release/*`, `hotfix/*`).
- **Code review:** [DEFINE: Code review process, e.g., required approvals, automated checks, PR template].

## Constraints

### Mandatory
- Strictly adhere to DDD layer separation and the Clean Architecture dependency rule.
- Define interfaces (ports) in the `domain` layer and implement them (adapters) in `infrastructure`.
- Use constructor-based dependency injection exclusively.
- Write BDD tests for high-level features and unit tests for components.
- Maintain minimum test coverage: 90% for domain, 80% for application.

### Prohibited
- The `domain` layer importing from `application`, `infrastructure`, or `interfaces`.
- Mixing business logic with infrastructure concerns in the same component.
- Using `panic` for recoverable errors.
- Committing API keys, secrets, or environment-specific configuration.
- Creating circular dependencies between packages.

## Approved Dependencies

| Library | Purpose | Restrictions |
|---------|---------|-------------|
| `github.com/spf13/cobra` | CLI application framework | [DEFINE: Version constraints] |
| `github.com/anthropics/anthropic-sdk-go` | Anthropic Claude LLM client | [DEFINE: Version constraints] |
| `google.golang.org/genai` | Google Gemini LLM client | [DEFINE: Version constraints] |
| `github.com/cucumber/godog` | BDD testing framework | [DEFINE: Version constraints] |
| `github.com/stretchr/testify` | Unit test assertion library | [DEFINE: Version constraints] |
| `github.com/mark3labs/mcp-go` | Model Context Protocol server | [DEFINE: Version constraints] |