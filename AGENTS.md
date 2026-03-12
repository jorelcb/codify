# AI Context Generator

A Go CLI tool that generates intelligent context files for AI agents from a project description. It uses LLMs (Claude, Gemini) and follows a Clean Architecture/DDD pattern to bridge the gap between a high-level idea and AI-ready development specifications.

## Tech Stack

- **Language:** Go 1.21+
- **Architecture:** Clean Architecture with DDD (Domain-Driven Design)
- **Framework:** Cobra (for CLI)
- **Testing:** Godog (BDD), testify/assert (Unit Testing)
- **Key dependencies:**
  - `github.com/spf13/cobra`
  - `github.com/anthropics/anthropic-sdk-go`
  - `google.golang.org/genai`
  - `github.com/cucumber/godog`
  - `github.com/stretchr/testify`
  - `github.com/mark3labs/mcp-go`
- **Build System:** Taskfile

## Commands

```bash
# Build the binary
task build

# Run all tests (unit and BDD)
task test

# Run the CLI tool (example: generate)
go run ./cmd/ai-context-generator/ generate my-api --description "A new API"

# Run linters
task lint
```

## Architectural Principles

### DDD - Layer Separation
- **Domain:** Pure business logic. Entities, value objects, repository interfaces (ports). No external dependencies.
- **Application:** Use cases (CQRS). Commands and queries orchestrate domain objects. Depends only on domain.
- **Infrastructure:** Concrete implementations (adapters). LLM clients, filesystem writers, database repositories. Implements domain interfaces.
- **Interfaces:** Entry points. CLI commands (Cobra), MCP server. Adapts external requests to application layer.

### Clean Architecture - Dependency Rule
Dependencies point inwards: `Interfaces` -> `Infrastructure` -> `Application` -> `Domain`. The `Domain` layer knows nothing about the outer layers. Abstractions are defined in `Domain`, implementations in `Infrastructure`.

## Key Conventions

- **Naming:** Follow idiomatic Go conventions (camelCase for internal variables, PascalCase for exported symbols).
- **Import structure:** Group imports: 1. Standard library, 2. External packages, 3. Internal project packages.
- **Error handling:** Use `fmt.Errorf` with `%w` to wrap errors with context. No silent error swallowing.
- **Commits:** Conventional Commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`).
- **Branching:** GitFlow (`main`, `develop`, `feature/*`, `release/*`, `hotfix/*`).
- **Testing:** Minimum coverage: 90% domain, 80% application.

## Project Structure

```
.
├── cmd/ai-context-generator/  # Main application entry point
├── internal/
│   ├── domain/                  # Core business logic, entities, and interfaces
│   │   ├── project/
│   │   ├── service/
│   │   └── shared/
│   ├── application/             # Use cases (Commands/Queries)
│   │   ├── command/
│   │   ├── dto/
│   │   └── query/
│   ├── infrastructure/          # Concrete implementations of domain interfaces
│   │   ├── filesystem/
│   │   ├── llm/
│   │   ├── persistence/
│   │   ├── scanner/
│   │   └── template/
│   └── interfaces/              # Adapters for external interaction (CLI, MCP)
│       ├── cli/
│       └── mcp/
├── templates/                   # Template guides for the LLM
└── tests/                       # BDD tests (Godog)
```

## Constraints

### MUST:
- Strictly adhere to the DDD layer separation and the Clean Architecture dependency rule.
- Define interfaces (ports) in the `domain` layer and implement them (adapters) in the `infrastructure` layer.
- Use constructor-based dependency injection. No global state for services.
- Handle all errors explicitly; wrap them for context.
- Write BDD tests for high-level features and unit tests for individual components.

### MUST NOT:
- Allow the `domain` layer to import any package from `application`, `infrastructure`, or `interfaces`.
- Mix business logic with infrastructure concerns.
- Use `panic` for recoverable errors.
- Commit API keys or other secrets. Use environment variables (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY`).
- Create circular dependencies between packages.

## Additional Context

- **Detailed Architecture & Data Flow:** `context/CONTEXT.md`
- **Session History & ADRs:** `context/INTERACTIONS_LOG.md`
## Specifications

- Project constitution: `specs/CONSTITUTION.md`
- Feature specifications: `specs/SPEC.md`
- Technical design and plan: `specs/PLAN.md`
- Task breakdown: `specs/TASKS.md`
