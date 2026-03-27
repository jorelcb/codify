# Codify - Feature Specification

## Feature 1: Context Generation from Project Description

### Description
Generates a set of intelligent context files (`AGENTS.md`, `CONTEXT.md`, `DEVELOPMENT_GUIDE.md`, `INTERACTIONS_LOG.md`, optionally `IDIOMS.md`) from a project name and description using an LLM. Core functionality exposed via the `generate` command.

### User Stories

**US-1:** As a developer, I want to provide a project name and description, so that a complete set of AI-ready context files is generated.

**Acceptance Criteria:**
- GIVEN a user provides a project name and description
- WHEN the `generate` command is executed
- THEN context files are created in the output directory, with `AGENTS.md` as root.
- AND when `--language` is specified, an additional `IDIOMS.md` file is generated.
- AND when `--with-specs` is specified, SDD specs are also generated.

**US-2:** As a developer, I want to read the description from a file for complex projects.

**Acceptance Criteria:**
- GIVEN a user specifies `--from-file path/to/description.md`
- WHEN the `generate` command is executed
- THEN the file content is used as the project description.

### Non-Functional Requirements
- **Performance:** Streaming API for all LLM calls to prevent timeouts.
- **Security:** API keys loaded from environment variables only.
- **Locale:** Output language controlled by `--locale en|es`.

---

## Feature 2: Specification Generation from Existing Context

### Description
Generates SDD specification files (`CONSTITUTION.md`, `SPEC.md`, `PLAN.md`, `TASKS.md`) based on existing context files. Exposed via the `spec` command.

### User Stories

**US-1:** As a developer, I want to generate implementation specs from my project context.

**Acceptance Criteria:**
- GIVEN a directory with valid `AGENTS.md` and `CONTEXT.md`
- WHEN the `spec` command is executed
- THEN spec files are created in `specs/` subdirectory.

---

## Feature 3: Project Analysis

### Description
Scans an existing codebase to detect language, framework, dependencies, and structure, then generates context files from the analysis. Exposed via `analyze` command.

### User Stories

**US-1:** As a developer with an existing project, I want to generate context without writing a description manually.

**Acceptance Criteria:**
- GIVEN an existing project directory with recognizable structure
- WHEN the `analyze` command is executed
- THEN the project scanner detects language, framework, and dependencies
- AND context files are generated from the analysis.

---

## Feature 4: Agent Skills Generation

### Description
Generates reusable Agent Skills (SKILL.md files) organized by category and preset. Supports static (instant, from catalog) and personalized (LLM-adapted) modes. Skills are installed to ecosystem-specific paths.

### User Stories

**US-1:** As a developer, I want pre-built skills for common patterns (DDD, testing, conventions).

**Acceptance Criteria:**
- GIVEN a user selects a category, preset, and static mode
- WHEN the `skills` command is executed
- THEN SKILL.md files are delivered with ecosystem-specific YAML frontmatter.

**US-2:** As a developer, I want skills tailored to my specific project.

**Acceptance Criteria:**
- GIVEN a user selects personalized mode and provides project context
- WHEN the `skills` command is executed
- THEN the LLM generates SKILL.md files adapted to the user's stack, domain, and architecture.

### Skill Catalog

| Category | Presets | Exclusive |
|---|---|---|
| `architecture` | `clean` (DDD, BDD, CQRS, Hexagonal), `neutral` (review, testing, API) | Yes |
| `testing` | `foundational` (Test Desiderata), `tdd` (includes foundational), `bdd` (includes foundational) | Yes |
| `conventions` | `conventional-commit`, `semantic-versioning`, `all` | No |

### Target Ecosystems

| Target | Frontmatter | Output Path |
|---|---|---|
| `claude` | `name`, `description`, `user-invocable: true` | `.claude/skills/` |
| `codex` | `name`, `description` | `.agents/skills/` |
| `antigravity` | `name`, `description`, `triggers` | `.agents/skills/` |

---

## Feature 5: Antigravity Workflows

### Description
Generates multi-step workflow files for Antigravity IDE with execution annotations (`// turbo`, `// parallel`, `// capture`, `// if`). Supports static and personalized modes.

### User Stories

**US-1:** As an Antigravity user, I want pre-built workflows for common development tasks.

**Acceptance Criteria:**
- GIVEN a user selects a workflow preset and static mode
- WHEN the `workflows` command is executed
- THEN workflow `.md` files are created in `.agent/workflows/` with Antigravity frontmatter and execution annotations.

### Workflow Catalog

| Preset | Description |
|---|---|
| `feature-development` | Branch → implement → test → PR → review lifecycle |
| `bug-fix` | Reproduce → diagnose → fix → test → PR |
| `release-cycle` | Version bump → changelog → tag → deploy |

### Planned: Claude Code Composite Workflows
Future phase: Generate composite plugin packages (SKILL.md + hooks.json + agents/*.md) that replicate workflow orchestration in Claude Code's compositional model. Three implementation options identified:
- **Option A (MVP):** Single SKILL.md with procedural multi-step instructions
- **Option B:** SKILL.md + hooks.json + agents/*.md package
- **Option C:** Full plugin directory structure

---

## Feature 6: MCP Server

### Description
Exposes all Codify capabilities as MCP tools for AI coding agents to call directly.

### Tools

| Tool | Type | API Key |
|---|---|---|
| `generate_context` | Generative | Required |
| `generate_specs` | Generative | Required |
| `analyze_project` | Generative | Required |
| `generate_skills` | Generative (personalized) / Static | Depends on mode |
| `generate_workflows` | Generative (personalized) / Static | Depends on mode |
| `commit_guidance` | Knowledge | Not needed |
| `version_guidance` | Knowledge | Not needed |

### Transport
- `stdio` — Local pipes, for Claude Code/Codex CLI integration
- `http` — Streamable HTTP, for remote deployments

---

## Priorities

| Feature | Priority | Status | Version |
|---------|----------|--------|---------|
| Context Generation (`generate`) | High | Complete | v0.1.0 |
| Spec Generation (`spec`) | High | Complete | v1.0.0 |
| Project Analysis (`analyze`) | Medium | Complete | v1.2.0 |
| Agent Skills (`skills`) | High | Complete | v1.3.0-v1.12.0 |
| Antigravity Workflows (`workflows`) | Medium | Complete | v1.13.0 |
| MCP Server (`serve`) | High | Complete (7 tools) | v1.2.0-v1.13.0 |
| Claude Code Composite Workflows | Medium | Planned | Next |