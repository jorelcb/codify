# Codify - Development Guide

## 1. Critical Information & Methodology

### Production Mindset
This is a production-grade tool distributed via Homebrew. Every component must be reliable, observable, and maintainable. Prioritize long-term operability in all technical decisions.

### Build & Test Commands (Taskfile)
All common tasks are managed via `Taskfile.yml`.
```bash
# Build the binary
task build

# Run all tests (unit + BDD)
task test

# Run only BDD tests
go test ./tests/...

# Run specific BDD suite
go test ./tests/bdd/workflow_catalog/...

# Check for code issues
task vet

# Run linters
task lint

# Generate test coverage
go test -coverprofile=coverage.out ./...
```

### Development Flow
1. **Plan First**: Propose a clear implementation plan before coding.
2. **Test-Driven**: Write BDD scenarios for domain behavior BEFORE or alongside code. Unit tests for components.
3. **Incremental Commits**: Commit small, logical units using Conventional Commits.
4. **Validate Continuously**: `go build ./... && go test ./...` must pass before every commit.
5. **Version Properly**: feat → minor bump, fix → patch bump. Always update version references (badges, server version, .version file).
6. **Never include AI credits** in commits (no Co-Authored-By, no model mentions).

### Testing Strategy

#### BDD Tests (Godog)
BDD tests validate domain behavior through Gherkin scenarios. Each suite lives in its own directory under `tests/bdd/`:

```
tests/bdd/{suite_name}/
├── {suite_name}.feature      # Gherkin scenarios
├── {suite_name}_test.go      # Godog test runner (TestMain + godog.TestSuite)
├── context.go                # FeatureContext struct (test state)
└── steps_definitions.go      # Step implementations
```

Shared utilities in `tests/bdd/commons/` (assertions, godog options).

Current suites:
- `project_repository` — 14 scenarios, 72 steps (Project entity, repository, value objects)
- `workflow_catalog` — 11 scenarios, 43 steps (Workflow catalog operations)

#### Unit Tests
Standard `*_test.go` files alongside source. Table-driven tests with `testify/assert`. Key packages with unit tests:
- `internal/domain/catalog/` — Skill catalog resolution, category names, legacy mapping
- `internal/application/command/` — Command construction and validation
- `internal/application/dto/` — DTO validation
- `internal/infrastructure/llm/` — Prompt builder output verification
- `internal/infrastructure/template/` — Template loader resolution
- `internal/infrastructure/scanner/` — Project scanner detection
- `internal/infrastructure/filesystem/` — File writer operations

#### Coverage Targets
- Domain Layer: 90%+
- Application Layer: 70%+
- Infrastructure Layer: 60%+

## 2. Architectural & Technical Standards

### Architectural Autonomy
The context files (`context/CONTEXT.md`) provide the blueprint. Improvements to domain modeling, interfaces, and patterns are welcome — document the problem, proposed solution, and trade-offs as an ADR in `context/INTERACTIONS_LOG.md`.

### Self-Validation Checklist
Before marking a component as complete:
- [ ] **DDD Layers**: Respects Domain/Application/Infrastructure/Interfaces separation?
- [ ] **Dependency Rule**: No inward imports from outer layers?
- [ ] **Testability**: Can be tested in isolation with interfaces/mocks?
- [ ] **Error Handling**: All errors handled and wrapped with context?
- [ ] **BDD Coverage**: Domain behavior has Gherkin scenarios?
- [ ] **No Panics**: Uses `error` values for all recoverable errors?
- [ ] **Build Green**: `go build ./... && go test ./...` passes?

### Security
- **Credentials**: Never commit secrets. Use `ANTHROPIC_API_KEY`, `GEMINI_API_KEY` environment variables.
- **Input Sanitization**: Treat all user input as untrusted.
- **Log Sanitization**: Never log API keys or sensitive data.

### Configuration Precedence
1. Command-line flags (highest priority)
2. Environment variables
3. Default values in code

## 3. Project-Specific Guidelines

### CLI (Cobra + charmbracelet/huh)
- **Interactive prompts**: All commands detect TTY via `isInteractive()` and prompt for missing flags using `charmbracelet/huh` forms. `interactive.go` contains shared helpers.
- **Explicit flag map**: Commands use `cmd.Flags().Visit()` with an `explicit` map to distinguish user-provided flags from defaults.
- **Output**: Structured output to `stdout`. Logs and errors to `stderr`.
- **Exit codes**: 0 for success, non-zero for errors.

### MCP Server
- **Transport strategy**: `StdioTransport` (local) and `HTTPTransport` (remote/Streamable HTTP) in `internal/interfaces/mcp/transport.go`.
- **7 Tools**: `generate_context`, `generate_specs`, `analyze_project`, `generate_skills`, `generate_workflows`, `commit_guidance`, `version_guidance`.
- **Knowledge tools**: `commit_guidance` and `version_guidance` load embedded templates directly — no LLM API call needed.
- **Handler pattern**: Each tool has a `*Tool()` definition function + `handle*()` handler function in `server.go`.

### Skills System
- **Declarative catalog**: `internal/domain/catalog/skills_catalog.go` defines categories, presets, template mappings, and metadata as Go data structures.
- **Three categories**: `architecture` (exclusive), `testing` (exclusive), `conventions` (non-exclusive, supports "all").
- **Two modes**: `static` (instant, from catalog templates) and `personalized` (LLM-adapted to project context).
- **Three ecosystems**: `claude`, `codex`, `antigravity` — each gets specific YAML frontmatter.
- **Install scopes**: `global` (agent's home path) or `project` (current directory).

### Workflows System
- **Target**: Antigravity IDE exclusively (only ecosystem with native workflow primitive).
- **Catalog**: `internal/domain/catalog/workflow_catalog.go` — separate bounded context from skills, reuses same structural types.
- **Three presets**: `feature-development`, `bug-fix`, `release-cycle`.
- **Output**: Flat `.md` files in `.agent/workflows/` with Antigravity-specific YAML frontmatter and execution annotations (`// turbo`, `// parallel`, `// capture: VAR`, `// if [condition]`).
- **Planned**: Claude Code support via composite plugin (SKILL.md + hooks + agents) — not yet implemented.

### Template System
- **Embedded via `embed.FS`** — templates ship inside the binary.
- **Locale-aware**: `templates/{locale}/{preset}/` for context, `templates/{locale}/skills/{dir}/` for skills, `templates/{locale}/workflows/` for workflows.
- **Language-specific idioms**: `templates/{locale}/languages/{lang}/idioms.template` — adds `IDIOMS.md` to context output when `--language` is specified.
- **CRITICAL**: Templates are structural guides for the LLM, NOT deterministic templates. They are NOT rendered with variable replacement.

### Key DDD Concepts
- **Ubiquitous Language**: Project, Context, Spec, Skill, Workflow, LLMProvider, TemplateLoader, Catalog, Preset, Category.
- **Entities**: `Project` (`internal/domain/project/entity.go`) — aggregate root with lifecycle.
- **Value Objects**: `ProjectDescription`, `Language` (`internal/domain/shared/value_objects.go`) — immutable, identity-less.
- **Repositories**: `ProjectRepository` (domain interface) → `InMemoryProjectRepository` (infrastructure).
- **Domain Services**: Interfaces in `internal/domain/service/interfaces.go` — `LLMProvider`, `FileWriter`, `TemplateLoader`, `DirectoryManager`.
- **Catalogs**: `SkillCategory`/`SkillOption` structural types shared between skill and workflow catalogs. Each catalog has its own metadata registry and finder functions.

## 4. Version Management

When releasing a new version:
1. Create feature commit(s) with Conventional Commits
2. Create `chore: bump version to vX.Y.Z` commit
3. Create annotated tag: `git tag -a vX.Y.Z -m "vX.Y.Z — description"`
4. Update ALL version references before pushing:
   - `README.md` badge (`version-X.Y.Z-blue`)
   - `README_ES.md` badge
   - `README.md` / `README_ES.md` project status sections
   - `internal/interfaces/mcp/server.go` (`serverVersion` constant)
   - `.version` file
5. Push: `git push origin main --tags`