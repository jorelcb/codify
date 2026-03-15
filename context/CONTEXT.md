# Codify - Architectural Context

## Architecture

Main pattern: **Clean Architecture with Domain-Driven Design (DDD)**

### DDD Layers and Responsibilities

#### Domain Layer (`internal/domain`)
- **Responsibility:** Pure business logic, technology-agnostic.
- **Contains:** Entities (`Project`), Value Objects (`Language`), repository interfaces (`ProjectRepository`), domain service interfaces (`LLMProvider`).
- **Rule:** No external dependencies. Defines WHAT it needs, not HOW it's implemented.

#### Application Layer (`internal/application`)
- **Responsibility:** Orchestrates use cases using the domain.
- **Contains:** Commands (`GenerateContextCommand`) and Queries (`ListProjectsQuery`) following CQRS. DTOs for inter-layer communication.
- **Rule:** Depends only on domain. Coordinates, does not implement business logic.

#### Infrastructure Layer (`internal/infrastructure`)
- **Responsibility:** Concrete implementations and technology adapters.
- **Contains:** `llm/` (Anthropic, Gemini providers), `template/` (loader), `filesystem/` (writer), `scanner/` (code scanner), `persistence/memory/` (repository implementation).
- **Rule:** Implements interfaces defined in the domain layer.

#### Interfaces Layer (`internal/interfaces`)
- **Responsibility:** Application entry points.
- **Contains:** `cli/` (Cobra commands), `mcp/` (Model Context Protocol server).
- **Rule:** Adapts user input (CLI flags, MCP requests) to application layer DTOs.

### Dependency Rule (Clean Architecture)
Dependencies always point inward: `Interfaces -> Application -> Domain`. The `Infrastructure` layer implements interfaces defined in `Domain` or `Application` (Dependency Inversion).

## Main Components

| Component | Responsibility | Layer | Dependencies/Interfaces |
|---|---|---|---|
| **CLI Commands** | Expose functionality (`generate`, `spec`, `analyze`) via Cobra. | Interfaces | `application.GenerateContextCommand` |
| **MCP Server** | Exposes functionality via Model Context Protocol. | Interfaces | `application.*` commands/queries |
| **LLM Provider** | Abstracts communication with LLM APIs (Claude, Gemini). | Infrastructure | Implements `domain.service.LLMProvider` |
| **Template Loader** | Loads structural templates from the filesystem. | Infrastructure | Implements `domain.service.TemplateLoader` |
| **Context Reader** | Reads existing `AGENTS.md` and `CONTEXT.md` for `spec` command. | Infrastructure | Used by `application.GenerateSpecCommand` |
| **Project Scanner** | Analyzes an existing codebase to detect language/framework. | Infrastructure | Used by `application.AnalyzeProjectCommand` |
| **File Writer** | Persists generated files to the output directory. | Infrastructure | Implements `domain.service.FileWriter` |

## Data Flow (`generate` command)

1.  **CLI (`interfaces/cli`):** Cobra command receives user input (`--description`, etc.).
2.  **Application (`application/command`):** `GenerateContextHandler` receives DTO.
3.  **Infrastructure (`infrastructure/template`):** `TemplateLoader` fetches structural guides.
4.  **Domain (`domain/service`):** `ProjectGenerator` orchestrates the logic.
5.  **Infrastructure (`infrastructure/llm`):** `LLMProvider` factory selects Claude/Gemini provider and makes API calls using templates as guides.
6.  **Infrastructure (`infrastructure/filesystem`):** `FileWriter` writes the streamed LLM response to disk (`output/<project-name>/`).
7.  **Application/CLI:** Returns `GenerationResult` (paths, tokens) to the user.

## Design Decisions

| Decision | Justification | Discarded Alternatives |
|---|---|---|
| **Per-file LLM generation** | Avoids token limits and JSON parsing failures. Allows for richer content per file and granular progress feedback. | Single LLM call returning a large JSON object with all file contents. |
| **AGENTS.md as root file** | Adheres to the Linux Foundation standard for AI agent context, maximizing tool compatibility. | Custom root file (`CLAUDE.md`, `.projectrc`). |
| **XML tags in system prompts** | Improves LLM's semantic understanding and output structure, especially for Anthropic's Claude models. | Markdown-only prompts. |
| **Multi-provider LLM factory** | Allows user flexibility (Claude/Gemini) without changing core logic. The factory pattern abstracts provider selection. | A single, hardcoded LLM provider. |
| **`spec` command is context-dependent** | Ensures that generated specifications are consistent with the established architectural context. | Standalone `spec` command that re-prompts for project details. |

## External Integrations

| Service | Purpose | Protocol | Resilience Pattern |
|---|---|---|---|
| **Anthropic API** | LLM for context generation (Claude models). | HTTPS/REST | Retry with backoff for transient errors. `[DEFINE]` Circuit Breaker. |
| **Google Gemini API** | LLM for context generation (Gemini models). | HTTPS/REST | Retry with backoff for transient errors. `[DEFINE]` Circuit Breaker. |

## Observability

Strategy requires implementation:
- **Tracing:** `[DEFINE]` Instrument key operations (LLM calls, file I/O) with OpenTelemetry spans.
- **Metrics:** `[DEFINE]` Track generation duration, token usage, error rates per model.
- **Logging:** Use a structured logger (e.g., slog). Correlate logs with trace IDs.

## Production Requirements (MCP Server)

### Graceful Shutdown
- **Required:** Capture `SIGINT`/`SIGTERM`.
- **Action:** Stop accepting new MCP requests, finish in-progress jobs, and close connections.
- **Timeout:** `[DEFINE]` A configurable timeout (e.g., 30s) for shutdown.

### Health Checks
- **Required:** Implement `/health/live` and `/health/ready` endpoints for the `serve` command.
- **Readiness:** Should check connectivity to required LLM APIs if API keys are present.

### Resilience
- **Required:** Implement retry with exponential backoff for all external API calls.
- **Required:** Implement configurable timeouts for LLM API requests.

## Implementation Checklist

### Phase 1: Foundation
- [x] Project structure according to DDD
- [x] Domain: entities, value objects, interfaces
- [x] Application: services for `generate` and `spec`
- [x] Infrastructure: Anthropic/Gemini providers, template loader, file writer
- [x] Interfaces: Cobra CLI commands
- [x] BDD tests for domain logic

### Phase 2: Robustness
- [ ] Graceful shutdown for `serve` command
- [ ] Health check endpoints for `serve` command
- [ ] Circuit breaker and comprehensive retry policies for LLM clients
- [ ] Integration tests for `generate` and `spec` flows
- [ ] Implement `analyze` command with Project Scanner

### Phase 3: Production & Observability
- [ ] `[DEFINE]` Authentication for MCP server
- [ ] OpenTelemetry instrumentation (tracing, metrics)
- [ ] Structured logging with correlation
- [ ] CI/CD pipeline for automated releases
- [ ] Complete user documentation and guides