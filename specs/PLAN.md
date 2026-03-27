# Codify - Technical Plan

## High-Level Architecture

The system uses **Clean Architecture with Domain-Driven Design (DDD)** divided into four layers with a strict inward-pointing dependency rule:

- **Interfaces (`interfaces/`):** Entry points — Cobra CLI (with charmbracelet/huh interactive menus) and MCP server (stdio + HTTP transport). Adapts external requests into application-layer DTOs.
- **Application (`application/`):** Use cases following CQRS — commands (generate, spec, skills, workflows) and queries (list). Depends only on Domain.
- **Infrastructure (`infrastructure/`):** Concrete implementations — LLM providers (Anthropic, Gemini), filesystem (writer, reader, directory manager), template loader (locale/preset/language-aware), project scanner, persistence.
- **Domain (`domain/`):** Pure business logic — Project entity, value objects, service interfaces, declarative catalogs (skills + workflows with separate bounded contexts).

## Technical Decisions

| Decision | Chosen Option | Justification |
|---|---|---|
| Generation granularity | Per-file LLM calls | Avoids token limits, prevents JSON parsing failures, granular progress. |
| Root file standard | AGENTS.md | Linux Foundation standard, maximum AI tool compatibility. |
| Prompting strategy | XML tags in system prompts | Improves Claude's semantic understanding and output consistency. |
| LLM integration | Multi-provider factory | User flexibility (Claude/Gemini) without changing core logic. |
| Spec generation | Context-dependent | Ensures specs are derived from established architectural context. |
| Catalog pattern | Declarative in-code catalogs | Compile-time safety, no config file parsing, easy to extend. |
| Workflow strategy | Antigravity-first | Only ecosystem with native workflow primitive. Validates concept first. |
| Template embedding | `embed.FS` | Binary works from any directory. No external file dependencies. |
| Interactive UX | charmbracelet/huh forms | Guides users through configuration; TTY detection for non-interactive safety. |
| Distribution | GoReleaser + Homebrew | Cross-platform binary distribution with simple install path. |

## Data Model

### Core Types

- **`Project` entity** (`domain/project/entity.go`): Aggregate root. Name, description, language, type, architecture, model.
- **`SkillCategory`** (`domain/catalog/skills_catalog.go`): Category name, label, exclusive flag, options list.
- **`SkillOption`**: Preset name, label, template directory, template mapping (file → output name).
- **`ResolvedSelection`**: Result of catalog resolution — template dir + merged template mappings.
- **`WorkflowMeta`** (`domain/catalog/workflow_catalog.go`): Description (max 250 chars).
- **DTOs**: `ProjectConfig`, `SkillsConfig`, `WorkflowConfig`, `SpecConfig`, `GenerationResult`.

### Domain Service Interfaces (`domain/service/interfaces.go`)

```go
type LLMProvider interface {
    GenerateContext(ctx, request) (*GenerationResponse, error)
}

type TemplateLoader interface {
    LoadAll() ([]TemplateGuide, error)
}

type FileWriter interface {
    WriteFile(path, content string) error
}

type DirectoryManager interface {
    EnsureDir(path string) error
}
```

## Data Flows

### Context Generation (`generate`)
CLI → `ProjectConfig` DTO → `GenerateContextCommand` → `TemplateLoader` (guides) → `ProviderFactory` → `LLMProvider` (per-file streaming) → `FileWriter` → `GenerationResult`

### Skills Generation (`skills`)
CLI → `SkillsConfig` DTO → `FindCategory()` + `Resolve()` → `TemplateLoader` → Static: `DeliverStaticSkillsCommand` / Personalized: `GenerateSkillsCommand` → `FileWriter` → `GenerationResult`

### Workflow Generation (`workflows`)
CLI → `WorkflowConfig` DTO → `FindWorkflowCategory()` + `Resolve()` → `TemplateLoader` → Static: `DeliverStaticWorkflowsCommand` / Personalized: `GenerateWorkflowsCommand` → `FileWriter` → `GenerationResult`

## Testing Strategy

| Level | Scope | Framework | Coverage |
|---|---|---|---|
| BDD | Domain behavior (catalogs, entities, value objects) | Godog | 25 scenarios, 115 steps |
| Unit | Components (commands, DTOs, prompt builder, scanner, loader) | testify/assert | Domain 90%+, App 70%+ |
| Integration | End-to-end flows | Planned | Not yet implemented |

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| LLM API unavailability | Medium | High | Retry with backoff (planned). Static mode as fallback for skills/workflows. |
| Inconsistent LLM output | Medium | Medium | XML-tagged prompts, grounding rules, per-file generation. |
| API key leakage | Low | Critical | Environment variables only. Never in logs or commits. |
| Token limits | Low | High | Per-file generation strategy stays within context windows. |
| Ecosystem API changes | Medium | Medium | Provider abstraction via factory pattern isolates changes. |

## Next Phase: Claude Code Composite Workflows

### Strategy
Replicate Antigravity workflow behavior in Claude Code using its compositional model:

| Option | Output | Complexity | Value |
|---|---|---|---|
| **A (MVP)** | SKILL.md with procedural multi-step instructions | Low | Immediate — `/workflow-name` command in Claude Code |
| **B** | SKILL.md + hooks.json + agents/*.md | Medium | Higher — deterministic validation via hooks |
| **C** | Full plugin directory | High | Highest — complete distributable package |

**Recommended path:** Start with Option A, evolve to B based on user feedback.

### What exists for reuse
- Workflow catalog (domain) — presets, metadata, resolution logic
- Template content — same workflow steps, different output format
- LLM pipeline — personalized mode infrastructure
- CLI/MCP patterns — command structure, tool registration

### What needs building
- Frontmatter adapter (Antigravity → Claude skill format)
- Annotation transformer (`// turbo` → prose instructions)
- Output path resolver (`.claude/skills/{workflow}/SKILL.md`)
- Prompt variant for Claude Code format
- `--target` flag on `workflows` command (currently hardcoded to antigravity)