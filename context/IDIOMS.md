# Codify - Go Idiomatic Guide

## Concurrency

### Fundamental Patterns
-   Manage concurrent file generation and API calls with `sync.WaitGroup`. A central goroutine should `wait` for all worker goroutines to complete.
-   Use channels for streaming data from LLM providers to file writers. This decouples the infrastructure layer from the application logic.
-   Use `select` with a `case <-ctx.Done()` to ensure all long-running operations (like LLM API calls) respect cancellation.
-   A worker pool pattern is ideal for the `generate` command when processing multiple files. Use a buffered channel to manage the job queue.
-   Use `sync.Mutex` for any in-memory state that might be accessed concurrently, such as a shared cache for templates.

### Context Propagation
-   **CRITICAL:** `context.Context` MUST be the first parameter in any function signature that involves I/O (filesystem, LLM APIs), long-running computations, or crosses API boundaries (CLI to Application, Application to Infrastructure).
-   Use `context.WithTimeout` for all external API calls to LLMs to prevent indefinite hangs.
-   Never store a `context.Context` inside a struct. Pass it explicitly to each function that needs it.

```go
// Correct: Pass context as the first argument
func (p *AnthropicProvider) GenerateContext(ctx context.Context, req GenerationRequest) (*Response, error) {
    // ...
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    // ...
    }
}
```

## Error Handling

-   **CRITICAL:** All errors originating from the `infrastructure` or `domain` layers must be wrapped. This preserves the call stack for debugging.
    -   Use `fmt.Errorf("application: failed to generate context: %w", err)`
-   Define specific, exported error variables in the `domain` layer for business rule violations (e.g., `var ErrProjectDescriptionEmpty = errors.New("project description cannot be empty")`).
-   Use `errors.Is()` to check for specific error values (e.g., `errors.Is(err, domain.ErrProjectDescriptionEmpty)`).
-   Use `errors.As()` to check if an error is of a specific custom type.
-   The application layer is responsible for handling errors and deciding what to present to the CLI layer. Infrastructure should only return errors, not log them.
-   `panic` is only acceptable during initialization if a critical configuration (like an API key) is missing. It should never be used for control flow.

## Language Conventions

### Naming
-   `PascalCase` for exported types, functions, and variables.
-   `camelCase` for unexported (internal) identifiers.
-   Interfaces should describe behavior: `LLMProvider`, `FileWriter`, `TemplateLoader`. Avoid `ILLMProvider`.
-   Acronyms are consistently cased: `LLM`, `API`, `HTTP`, `XML`.

### Import Organization
```go
// Standard library
import (
    "context"
    "fmt"
)

// External dependencies
import (
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

// Internal project modules
import (
    "codify/internal/domain/project"
    "codify/internal/infrastructure/llm"
)
```

### Structs and Methods
-   Use `New...` constructor functions that return a pointer to a struct and an error if initialization can fail.
-   Receivers should be short (e.g., `p` for `Project`), consistent, and not `this` or `self`.
-   Use a pointer receiver (`func (p *Project) ...`) for methods that modify the struct's state or for large structs to avoid copying.
-   Use a value receiver for small, immutable structs like Value Objects.

## Version and Tooling Constraints

-   **Go Version:** 1.23 or higher (uses `iter.Seq2` for Gemini streaming).
-   **Dependencies:** Use `go mod tidy` to keep `go.mod` and `go.sum` clean.
-   **Linting:** Use `golangci-lint` with a project-level configuration file (`.golangci.yml`). This should be part of the CI pipeline.
-   **Build System:** Use the existing `Taskfile.yml` for all common tasks like building, testing, and linting.
    -   `task build`
    -   `task test`
    -   `task lint`

## Testing in Go

-   **BDD Tests:** Use `godog` for Behavior-Driven Development tests that cover domain behavior. Each suite in `tests/bdd/{suite}/` with: `.feature` (Gherkin), `context.go` (FeatureContext), `steps_definitions.go`, `*_test.go` (runner).
-   **Unit Tests:** Use the standard library `testing` package with table-driven tests for focused testing of individual functions and methods.
-   **Assertions:** Use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` for readable and expressive test assertions.
-   **Mocks/Stubs:** For infrastructure dependencies (like `LLMProvider` or `FileWriter`), use interfaces and create test doubles (mocks or stubs) to isolate the code under test.
-   **Parallelism:** Use `t.Parallel()` in unit tests that do not share or modify global state to speed up test execution.
-   **Coverage:** Aim for high test coverage in the `domain` and `application` layers. Use `go test -coverprofile=coverage.out` and `go tool cover -html=coverage.out` to analyze coverage.