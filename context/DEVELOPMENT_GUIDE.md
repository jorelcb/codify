# Codify - Development Guide

## 1. Critical Information & Methodology

### Production Mindset
This is a production-grade tool. Every component must be reliable, observable, and maintainable. It is not a prototype. Prioritize long-term operability in all technical decisions.

### Build & Test Commands (Taskfile)
All common tasks are managed via `Taskfile.yml`.
```bash
# Build the binary
task build

# Run all tests (unit + BDD)
task test

# Run only BDD tests with Godog
task test-bdd

# Check for code issues
task vet
```

### Development Flow
1.  **Plan First**: Propose a clear implementation plan before coding.
2.  **Test-Driven**: Write tests alongside or before the code. Do not defer testing.
3.  **Incremental Commits**: Commit small, logical units of work.
4.  **Validate Continuously**: Ensure each component works before integration.
5.  **Clarify Ambiguity**: Ask for clarification on any unclear requirement.

### Testing Strategy
-   **Unit Tests**: Use the standard `testing` package and `testify/assert` for all domain logic and application services. Each public function must have a corresponding test file (e.g., `service.go`, `service_test.go`).
-   **BDD Tests**: Use `godog` for behavior-driven development of key features. Feature files are located in `tests/features/`. Step definitions are in `tests/bdd/`.
-   **Minimum Coverage**:
    -   Domain Layer: 90%
    -   Application Layer: 70%
    -   Infrastructure Layer: 60%

## 2. Architectural & Technical Standards

### Architectural Autonomy
The existing context files (`CONTEXT.md`, `ARCHITECTURE.md`) provide the blueprint. However, you have the autonomy to propose improvements to:
-   Domain modeling (entities, value objects)
-   Interfaces and abstractions
-   Architectural patterns
-   Code organization

To propose a change, document the problem, the proposed solution, and the trade-offs.

### Self-Validation Checklist
Before marking a component as complete, verify:
-   [ ] **DDD Layers**: Does it respect the separation between Domain, Application, and Infrastructure?
-   [ ] **Testability**: Can it be tested in isolation?
-   [ ] **Error Handling**: Are all errors handled gracefully and wrapped with context?
-   [ ] **Observability**: Are there sufficient logs and traces?
-   [ ] **Maintainability**: Is the code clear, idiomatic, and self-documenting?
-   [ ] **No Panics**: The application must never panic. Use `error` values for all recoverable errors.

### Security
-   **Credentials**: Never commit secrets. Use environment variables for API keys (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY`).
-   **Input Sanitization**: Treat all user input as untrusted.
-   **Log Sanitization**: Do not log sensitive data or API keys.

### Configuration
Configuration is loaded with the following precedence, using Cobra/Viper patterns:
1.  Command-line flags (e.g., `--model claude-3-opus-20240229`)
2.  Environment variables (e.g., `ANTHROPIC_API_KEY`)
3.  Default values defined in code.

## 3. Project-Specific Guidelines

### Project Type: CLI
This is primarily a CLI application built with Cobra.
-   **Commands**: Keep command structure intuitive (`generate`, `spec`, `analyze`, `serve`).
-   **Flags**: Use descriptive flags for all configuration options.
-   **Output**: Provide structured output to `stdout`. Use `stderr` for logs and errors.
-   **Exit Codes**: Use meaningful exit codes (0 for success, non-zero for errors).

### Project Type: MCP Server
The `serve` command exposes the tool's capabilities via the Model-Context-Protocol (MCP).
-   **Transport**: The server must support both `stdio` and `http` transports as defined in `internal/interfaces/mcp/`.
-   **Tools**: Expose `generate`, `analyze`, and `spec` as distinct tools available to the MCP client.

### Key DDD Concepts in this Project
-   **Ubiquitous Language**: Use terms from the domain (Project, Context, Spec, LLMProvider) consistently.
-   **Entities**: `Project` (`internal/domain/project/entity.go`) is the core entity with a distinct identity and lifecycle.
-   **Value Objects**: `ProjectDescription`, `Language` (`internal/domain/shared/value_objects.go`) are immutable and defined by their attributes.
-   **Repositories**: The `ProjectRepository` interface is defined in the domain layer (`internal/domain/project/repository.go`) and implemented in infrastructure.
-   **Services**: `ProjectGenerator` (`internal/domain/service/project_generator.go`) orchestrates domain logic that doesn't naturally fit within an entity.

## 4. Observability and Final Checks

### Detailed Observability with OpenTelemetry
-   **Dependency Injection**: The tracer provider must be injected via constructors. DO NOT use global tracer variables.
-   **Context Propagation**: Propagate `context.Context` through all function calls to ensure trace continuity.
-   **Spans**: Create detailed spans for key operations:
    -   LLM API calls (`llm.generate`)
    -   Template loading (`template.load`)
    -   File I/O (`filesystem.write`)
-   **Attributes**: Add semantic attributes to spans (e.g., `llm.model`, `file.path`).
-   **Logging**: Use a structured logger (e.g., `slog`) and include `TraceID` and `SpanID` in all logs.

### Communication Style
-   **Code**: Write self-documenting Go code. Comments should explain *why*, not *what*.
-   **Commits**: Follow Conventional Commits standard (`feat:`, `fix:`, `refactor:`, `docs:`).
-   **Decisions**: Justify significant decisions in `INTERACTIONS_LOG.md` as an Architectural Decision Record (ADR).