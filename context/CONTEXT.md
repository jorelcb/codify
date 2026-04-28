# Codify - Architectural Context

## Architecture

Main pattern: **Clean Architecture with Domain-Driven Design (DDD)**

### DDD Layers and Responsibilities

#### Domain Layer (`internal/domain`)
- **Responsibility:** Pure business logic, technology-agnostic.
- **Contains:**
  - `catalog/` — Declarative skill catalog (`SkillCategory`, `SkillOption`, `SkillMetadata`) and workflow catalog (`WorkflowCategories`, `WorkflowMetadata`). Both catalogs share structural types (`SkillCategory`, `SkillOption`, `ResolvedSelection`) but are separate bounded contexts with independent metadata registries.
  - `project/` — `Project` entity (aggregate root), `ProjectRepository` interface.
  - `service/` — Domain service interfaces: `LLMProvider`, `TemplateLoader`, `FileWriter`, `DirectoryManager`. Also contains `GenerationRequest`/`GenerationResponse` types.
  - `shared/` — Value objects (`Language`, `ProjectDescription`), domain errors.
- **Rule:** No external dependencies. Defines WHAT it needs, not HOW it's implemented.

#### Application Layer (`internal/application`)
- **Responsibility:** Orchestrates use cases using the domain. Follows CQRS pattern.
- **Commands:**
  - `GenerateContextCommand` — generates context files (AGENTS.md, CONTEXT.md, etc.)
  - `GenerateSpecCommand` — generates SDD specs from existing context
  - `GenerateSkillsCommand` — generates LLM-personalized skills
  - `DeliverStaticSkillsCommand` — delivers pre-built skills from catalog
  - `GenerateWorkflowsCommand` — generates LLM-personalized workflows
  - `DeliverStaticWorkflowsCommand` — delivers pre-built workflows from catalog
- **Queries:** `ListProjectsQuery`
- **DTOs:** `ProjectConfig`, `SkillsConfig`, `WorkflowConfig`, `SpecConfig`, `GenerationResult`, `ProjectInfo`
- **Rule:** Depends only on domain. Coordinates, does not implement business logic.

#### Infrastructure Layer (`internal/infrastructure`)
- **Responsibility:** Concrete implementations and technology adapters.
- **Contains:**
  - `llm/` — `AnthropicProvider` (Claude SDK v1.25.0), `GeminiProvider` (GenAI SDK v1.49.0), `ProviderFactory` (resolves by model prefix), `PromptBuilder` (XML-tagged system prompts for context/spec/skills/workflows).
  - `template/` — `FileSystemTemplateLoader` with locale/preset/language-aware resolution. Supports template mappings for skills and workflows.
  - `filesystem/` — `FileWriter`, `DirectoryManager`, `ContextReader` (reads existing AGENTS.md/CONTEXT.md for spec generation).
  - `scanner/` — `ProjectScanner` detects language, framework (20+), dependencies, directory structure, config signals (CI/CD, Docker, K8s, Terraform).
  - `persistence/memory/` — In-memory `ProjectRepository` implementation.
- **Rule:** Implements interfaces defined in the domain layer.

#### Interfaces Layer (`internal/interfaces`)
- **Responsibility:** Application entry points.
- **Contains:**
  - `cli/commands/` — Cobra commands: `generate`, `analyze`, `spec`, `skills`, `workflows`, `serve`, `list`. Interactive prompts via charmbracelet/huh for missing flags.
  - `cli/commands/interactive.go` — Shared interactive menu helpers (TTY detection, form builders).
  - `mcp/server.go` — MCP server with 7 tools: `generate_context`, `generate_specs`, `analyze_project`, `generate_skills`, `generate_workflows`, `commit_guidance`, `version_guidance`.
  - `mcp/transport.go` — Transport strategy pattern: `StdioTransport`, `HTTPTransport`.
- **Rule:** Adapts user input (CLI flags, MCP requests) to application layer DTOs.

### Dependency Rule (Clean Architecture)
Dependencies always point inward: `Interfaces → Application → Domain`. The `Infrastructure` layer implements interfaces defined in `Domain` (Dependency Inversion). Nothing in `Domain` imports from outer layers.

## Main Components

| Component | Responsibility | Layer | Key Interfaces |
|---|---|---|---|
| **CLI Commands** | Expose functionality via Cobra with interactive menus | Interfaces | Application commands/queries |
| **MCP Server** | Expose functionality via Model Context Protocol (7 tools) | Interfaces | Application commands/queries |
| **Skill Catalog** | Declarative registry of skill categories, presets, metadata | Domain | `SkillCategory`, `SkillOption`, `SkillMetadata` |
| **Workflow Catalog** | Declarative registry of workflow presets and metadata | Domain | `WorkflowCategories`, `WorkflowMeta` |
| **LLM Provider** | Abstract communication with LLM APIs (Claude, Gemini) | Infrastructure | `domain.service.LLMProvider` |
| **Provider Factory** | Select provider by model prefix (`claude-*` → Anthropic, `gemini-*` → Gemini) | Infrastructure | `llm.NewProvider()` |
| **Prompt Builder** | Build XML-structured system prompts for each generation mode | Infrastructure | Context/Spec/Skills/Workflows prompt builders |
| **Template Loader** | Load structural templates with locale/preset/language awareness | Infrastructure | `domain.service.TemplateLoader` |
| **Context Reader** | Read existing AGENTS.md/CONTEXT.md for spec generation | Infrastructure | Used by `GenerateSpecCommand` |
| **Project Scanner** | Analyze existing codebase (language, framework, deps, structure) | Infrastructure | Used by `analyze` command |
| **File Writer** | Persist generated files to output directory | Infrastructure | `domain.service.FileWriter` |
| **Directory Manager** | Create/manage output directories | Infrastructure | `domain.service.DirectoryManager` |

## Data Flows

### `generate` command
1. **CLI** — Cobra command receives `--description`, `--name`, `--preset`, `--language`, `--locale`, `--model`. Missing flags prompt interactively.
2. **Application** — `GenerateContextCommand` receives `ProjectConfig` DTO.
3. **Infrastructure** — `TemplateLoader` fetches structural guides from `templates/{locale}/{preset}/`.
4. **Infrastructure** — `ProviderFactory` selects Claude/Gemini. `PromptBuilder` constructs XML system prompt. Provider makes per-file streaming API calls.
5. **Infrastructure** — `FileWriter` writes streamed response to `output/{project-name}/`.
6. **Application/CLI** — Returns `GenerationResult` (paths, tokens, model).

### `skills` command
1. **CLI** — Interactive or flag-based selection of category, preset, mode, target, install scope, locale.
2. **Domain** — `FindCategory()` + `Resolve()` from declarative catalog. Returns `ResolvedSelection` with template mappings.
3. **Infrastructure** — `TemplateLoader` loads skill templates from `templates/{locale}/skills/{dir}/`.
4. **Application** — Static mode: `DeliverStaticSkillsCommand` writes templates with ecosystem frontmatter. Personalized mode: `GenerateSkillsCommand` sends templates as guides to LLM, writes adapted output.
5. **Output** — SKILL.md files in target-specific paths (`.claude/skills/`, `.agents/skills/`, `~/.claude/skills/`).

### `workflows` command
1. **CLI** — Interactive or flag-based selection of preset, mode, install scope, locale. Target is always Antigravity.
2. **Domain** — `FindWorkflowCategory()` + `Resolve()`. Returns `ResolvedSelection` with workflow template mappings.
3. **Infrastructure** — `TemplateLoader` loads workflow templates from `templates/{locale}/workflows/`.
4. **Application** — Static: `DeliverStaticWorkflowsCommand` writes flat .md files with Antigravity frontmatter. Personalized: `GenerateWorkflowsCommand` adapts via LLM.
5. **Output** — `.agent/workflows/*.md` files with execution annotations (`// turbo`, `// capture`, `// if`, etc.).

## Design Decisions

| Decision | Justification | Discarded Alternatives |
|---|---|---|
| **Per-file LLM generation** | Avoids token limits and JSON parsing failures. Granular progress feedback. | Single LLM call returning large JSON. |
| **AGENTS.md as root file** | Linux Foundation standard for AI agent context. Maximum tool compatibility. | Custom root files (CLAUDE.md, .projectrc). |
| **XML tags in system prompts** | Improves LLM semantic understanding, especially for Claude. | Markdown-only prompts. |
| **Multi-provider LLM factory** | User flexibility (Claude/Gemini) without changing core logic. | Single hardcoded provider. |
| **Spec is context-dependent** | Ensures specifications are consistent with architectural context. | Standalone spec command. |
| **Declarative catalogs** | Skills/workflows defined as in-code data structures with metadata. No config files. | YAML/JSON config files, database. |
| **Separate skill and workflow catalogs** | Different bounded contexts — skills = expertise, workflows = orchestration. | Single unified catalog. |
| **Antigravity-first for workflows** | Only ecosystem with native workflow primitive. Validates concept before expanding. | Simultaneous multi-ecosystem support. |
| **Embedded templates (embed.FS)** | Binary works from any directory. No external file dependencies. | Filesystem templates alongside binary. |
| **Interactive prompts via huh** | Better UX — guides users through all configuration options. | Flag-only CLI requiring full command memorization. |

## Catalog Architecture

### Skills Catalog (`internal/domain/catalog/skills_catalog.go`)

Three categories, each with exclusive or non-exclusive preset resolution:

| Category | Exclusive | Presets | Template Dir |
|---|---|---|---|
| `architecture` | Yes (pick one) | `clean`, `neutral` | `default`, `neutral` |
| `testing` | Yes (pick one) | `foundational`, `tdd`, `bdd` | `testing` |
| `conventions` | No (supports `all`) | `conventional-commit`, `semantic-versioning`, `all` | `conventions` |

Each preset maps to template files via `TemplateMapping` (template filename → output name). `SkillMetadata` provides ecosystem frontmatter (name, description, triggers).

Legacy mapping: `"workflow"` → `{"conventions", "all"}` for backward compatibility.

### Workflow Catalog (`internal/domain/catalog/workflow_catalog.go`)

Single non-exclusive category with 3 presets:

| Preset | Description (max 250 chars, Antigravity constraint) |
|---|---|
| `spec-driven-change` | Spec-driven feature lifecycle (propose → apply → archive) — generates three skills (`/spec-propose`, `/spec-apply`, `/spec-archive`) implementing the OpenSpec-compatible SDD methodology |
| `bug-fix` | Structured bug fix: reproduce, diagnose, fix, test, and PR |
| `release-cycle` | Release process: version bump, changelog, tag, and deploy |

For Antigravity, workflows generate flat `.md` files. For Claude Code, workflows generate native skill directories (`{preset}/SKILL.md`). `GenerateWorkflowFrontmatter()` produces target-specific YAML frontmatter (Antigravity uses `description` only; Claude uses `name`, `description`, `disable-model-invocation`, `allowed-tools`).

#### Spec-driven Change: design rationale

`spec-driven-change` is the canonical workflow for feature work and non-trivial changes. It is the only preset that maps to **multiple templates** (3 skills from one preset selection) — exercising `TemplateMapping`'s built-in multi-mapping capability. This design choice reflects the cognitive separation between SDD phases:

- **Propose** (`/spec-propose`) — planning mode: read existing specs, identify capabilities, produce `proposal.md`, `design.md`, `tasks.md`, and capability-organized spec deltas (`ADDED` / `MODIFIED` / `REMOVED` requirements with G/W/T scenarios). No code is written. A feature branch is created and the proposal is committed.

- **Apply** (`/spec-apply`) — implementation mode: read approved proposal artifacts, execute `tasks.md` sequentially with atomic commits, run tests, open PR. Spec ambiguity is already resolved at this point.

- **Archive** (`/spec-archive`) — consolidation mode: merge spec deltas into source-of-truth (`openspec/specs/<capability>/spec.md`), move change to `openspec/changes/archive/YYYY-MM-DD-<id>/` for audit, merge feature branch, clean up.

The output structure follows the [OpenSpec](https://openspec.dev/) convention. Codify generates the skills; OpenSpec workspaces can consume them directly. Codify adds LLM personalization, multi-target support (Claude/Antigravity), and locale support on top of the OpenSpec-compatible format.

**Why this isn't a single workflow**: keeping the three phases separate is intentional. Each phase has a distinct cognitive mode — mixing planning, implementation, and consolidation in one prompt produces vague plans and sloppy code. The split also enables phase-specific review (intent review vs code review vs spec review) and parallel work (one developer drafting a propose while another applies a different change).

## External Integrations

| Service | Purpose | Protocol | SDK |
|---|---|---|---|
| **Anthropic API** | Claude LLM for generation | HTTPS/REST | `anthropic-sdk-go` v1.25.0 |
| **Google Gemini API** | Gemini LLM for generation | HTTPS/REST | `google.golang.org/genai` v1.49.0 |

API keys resolved via environment: `ANTHROPIC_API_KEY` for Claude, `GEMINI_API_KEY`/`GOOGLE_API_KEY` for Gemini. Provider auto-detected from `--model` prefix, or from available API key when no model specified.

Default models: `claude-sonnet-4-6` (Anthropic), `gemini-3.1-pro-preview` (Gemini).

## MCP Server

The `serve` command exposes Codify as an MCP server for AI coding agents:

- **Transport:** `StdioTransport` (local/pipes) or `HTTPTransport` (remote/Streamable HTTP).
- **7 Tools:** `generate_context`, `generate_specs`, `analyze_project`, `generate_skills`, `generate_workflows`, `commit_guidance`, `version_guidance`.
- **Knowledge tools** (`commit_guidance`, `version_guidance`) load embedded templates and return behavioral context — no API key needed.
- **Server version:** Tracks project version for compatibility.

## Testing Architecture

- **BDD Suites (Godog):**
  - `tests/bdd/project_repository/` — 14 scenarios, 72 steps. Tests Project entity lifecycle, repository CRUD, value object validation.
  - `tests/bdd/workflow_catalog/` — 11 scenarios, 43 steps. Tests workflow catalog operations: find category, resolve presets, resolve "all", unknown preset error, frontmatter generation, description length validation, category names.
  - Each suite: `*.feature` (Gherkin) + `context.go` (FeatureContext struct) + `steps_definitions.go` + `*_test.go` (Godog runner).
- **Unit tests:** Standard `*_test.go` files alongside source. Table-driven tests with testify/assert.
- **Total:** 25 BDD scenarios, 115 steps, all green.