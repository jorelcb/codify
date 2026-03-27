# Codify

A Go CLI tool that generates intelligent context files, specifications, agent skills, and orchestration workflows for AI coding agents. It uses LLMs (Claude, Gemini) to transform project descriptions into production-ready artefacts that teach AI agents *what* your project is, *how* to build it, *what* conventions to follow, and *how* to orchestrate complex tasks.

## Tech Stack

- **Language:** Go 1.23+
- **Architecture:** Clean Architecture with DDD (Domain-Driven Design)
- **CLI Framework:** Cobra + charmbracelet/huh (interactive menus)
- **Testing:** Godog (BDD), testify/assert (Unit)
- **Key dependencies:**
  - `github.com/spf13/cobra` — CLI framework
  - `github.com/anthropics/anthropic-sdk-go` v1.25.0 — Anthropic Claude SDK
  - `google.golang.org/genai` v1.49.0 — Google Gemini SDK
  - `github.com/cucumber/godog` — BDD testing framework
  - `github.com/stretchr/testify` — Unit test assertions
  - `github.com/mark3labs/mcp-go` v0.45.0 — Model Context Protocol server
  - `github.com/charmbracelet/huh` — Interactive terminal forms
- **Build System:** Taskfile
- **Distribution:** Homebrew formula, GoReleaser, GitHub Actions CI/CD

## Commands

```bash
# Build
task build

# Run all tests (unit + BDD)
task test

# Run BDD tests only
go test ./tests/...

# Run the CLI
go run ./cmd/codify/ <command>

# Linters
task lint
```

### Available CLI Commands

| Command | Description |
|---------|-------------|
| `generate` | Generate context files (AGENTS.md, CONTEXT.md, etc.) from a project description |
| `analyze` | Scan an existing project and generate context from its structure |
| `spec` | Generate SDD specs (CONSTITUTION.md, SPEC.md, PLAN.md, TASKS.md) from existing context |
| `skills` | Generate reusable Agent Skills (SKILL.md) by category, preset, and mode |
| `workflows` | Generate Antigravity workflow files with execution annotations |
| `serve` | Start MCP server (stdio or HTTP transport) with 7 tools |
| `list` | List previously generated projects |

All commands support `--locale en|es` for bilingual output. Commands with missing flags prompt interactively when run in a terminal (charmbracelet/huh forms).

## Architectural Principles

### DDD - Layer Separation
- **Domain (`internal/domain/`):** Pure business logic. Entities, value objects, repository interfaces, service interfaces, declarative catalogs. No external dependencies.
- **Application (`internal/application/`):** Use cases (CQRS). Commands and queries orchestrate domain objects. Depends only on domain.
- **Infrastructure (`internal/infrastructure/`):** Concrete implementations. LLM providers, filesystem writers, template loaders, project scanner. Implements domain interfaces.
- **Interfaces (`internal/interfaces/`):** Entry points. CLI commands (Cobra), MCP server. Adapts external requests to application layer.

### Clean Architecture - Dependency Rule
Dependencies point inwards: `Interfaces` → `Infrastructure` → `Application` → `Domain`. The `Domain` layer knows nothing about outer layers. Abstractions are defined in `Domain`, implementations in `Infrastructure`.

### Key Design Patterns
- **CQRS:** Commands (generate, spec, skills, workflows) and Queries (list) in Application layer.
- **Factory Pattern:** `llm.NewProvider()` selects Claude/Gemini based on model prefix.
- **Strategy Pattern:** MCP transport (stdio/HTTP), template loading.
- **Declarative Catalog:** Skills and Workflows use in-code catalogs with metadata registries — no config files.
- **Dependency Injection:** Constructor-based DI everywhere. No global state.

## Key Conventions

- **Naming:** Idiomatic Go (camelCase internal, PascalCase exported). Acronyms consistently cased: `LLM`, `API`, `HTTP`.
- **Imports:** Grouped: 1. Standard library, 2. External packages, 3. Internal project packages.
- **Error handling:** `fmt.Errorf` with `%w` to wrap errors. No silent swallowing. No `panic` for control flow.
- **Commits:** Conventional Commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`).
- **Versioning:** Semantic Versioning. Tags on main branch.
- **Testing:** BDD with Godog for domain behavior (catalogs, entities). Unit tests with testify for components.
- **Templates are NOT deterministic.** They are structural guides for the LLM, not files to render with variable replacement.

## Project Structure

```
.
├── cmd/codify/                         # Main entry point
├── internal/
│   ├── domain/                         # 💎 Pure business logic
│   │   ├── catalog/                    # Declarative skill + workflow catalogs
│   │   │   ├── skills_catalog.go       # 3 categories: architecture, testing, conventions
│   │   │   ├── skills_catalog_test.go
│   │   │   ├── workflow_catalog.go     # 1 category: workflows (3 presets)
│   │   │   └── workflow_catalog_test.go
│   │   ├── project/                    # Project entity (aggregate root) + repository interface
│   │   ├── service/                    # Interfaces: LLMProvider, FileWriter, TemplateLoader
│   │   └── shared/                     # Value objects, domain errors
│   │
│   ├── application/                    # 🔄 Use cases (CQRS)
│   │   ├── command/                    # GenerateContext, GenerateSpec, GenerateSkills,
│   │   │                               # DeliverStaticSkills, GenerateWorkflows, DeliverStaticWorkflows
│   │   ├── dto/                        # ProjectConfig, SkillsConfig, WorkflowConfig, SpecConfig, GenerationResult
│   │   └── query/                      # ListProjects
│   │
│   ├── infrastructure/                 # 🔧 Implementations
│   │   ├── llm/                        # AnthropicProvider, GeminiProvider, ProviderFactory, PromptBuilder
│   │   ├── template/                   # FileSystemTemplateLoader (locale + preset + language-aware)
│   │   ├── scanner/                    # ProjectScanner (language, deps, framework detection)
│   │   ├── filesystem/                 # FileWriter, DirectoryManager, ContextReader
│   │   └── persistence/memory/         # In-memory ProjectRepository
│   │
│   └── interfaces/                     # 🎯 Entry points
│       ├── cli/
│       │   ├── commands/               # generate, analyze, spec, skills, workflows, serve, list
│       │   │   └── interactive.go      # Shared interactive menu helpers (charmbracelet/huh)
│       │   └── root.go                 # Root command registration
│       └── mcp/
│           ├── server.go               # 7 MCP tools registration + handlers
│           └── transport.go            # Transport strategy (stdio, HTTP)
│
├── templates/                          # Embedded template guides for the LLM
│   ├── {en,es}/
│   │   ├── default/                    # Context preset: DDD/Clean Architecture
│   │   ├── neutral/                    # Context preset: generic/no opinions
│   │   ├── spec/                       # Spec templates (constitution, spec, plan, tasks)
│   │   ├── skills/                     # Skill templates by category
│   │   │   ├── default/               # Architecture: Clean (DDD, BDD, CQRS, Hexagonal)
│   │   │   ├── neutral/               # Architecture: Neutral (review, testing, API)
│   │   │   ├── testing/               # Testing: Foundational, TDD, BDD
│   │   │   └── conventions/           # Conventions: Conventional Commits, Semantic Versioning
│   │   ├── workflows/                  # Antigravity workflow templates
│   │   │   ├── feature_development.template
│   │   │   ├── bug_fix.template
│   │   │   └── release_cycle.template
│   │   └── languages/                  # Language-specific idiomatic guides
│   │       ├── go/, javascript/, python/
│
├── tests/
│   ├── bdd/                            # BDD test suites (Godog)
│   │   ├── commons/                    # Shared assertions and options
│   │   ├── project_repository/         # 14 scenarios, 72 steps
│   │   └── workflow_catalog/           # 11 scenarios, 43 steps
│   └── features/domain/               # Feature files for domain entities
│
├── context/                            # Project's own context documentation
│   ├── CONTEXT.md                      # Architectural deep-dive
│   ├── DEVELOPMENT_GUIDE.md            # Development methodology
│   ├── IDIOMS.md                       # Go idiomatic patterns
│   └── INTERACTIONS_LOG.md             # Session history + ADRs
│
├── specs/                              # Project's own specifications
│   ├── CONSTITUTION.md, SPEC.md, PLAN.md, TASKS.md
│
└── input/                              # Reference/input documents
```

## Constraints

### MUST:
- Strictly adhere to DDD layer separation and Clean Architecture dependency rule.
- Define interfaces (ports) in `domain` and implement them (adapters) in `infrastructure`.
- Use constructor-based dependency injection. No global state for services.
- Handle all errors explicitly; wrap them for context.
- Write BDD tests for domain behavior (catalogs, entities) and unit tests for components.
- Templates are guides for LLM generation — never treat them as deterministic templates.

### MUST NOT:
- Allow `domain` to import from `application`, `infrastructure`, or `interfaces`.
- Mix business logic with infrastructure concerns.
- Use `panic` for recoverable errors.
- Commit API keys or secrets. Use environment variables (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY`).
- Create circular dependencies between packages.
- Include AI credits (Co-Authored-By, model mentions) in commits.

## Additional Context

- **Detailed Architecture & Data Flow:** `context/CONTEXT.md`
- **Development Methodology:** `context/DEVELOPMENT_GUIDE.md`
- **Go Idioms:** `context/IDIOMS.md`
- **Session History & ADRs:** `context/INTERACTIONS_LOG.md`

## Specifications

- Project constitution: `specs/CONSTITUTION.md`
- Feature specifications: `specs/SPEC.md`
- Technical design and plan: `specs/PLAN.md`
- Task breakdown and roadmap: `specs/TASKS.md`