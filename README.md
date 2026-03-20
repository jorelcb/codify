# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-1.10.0-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.21)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Context. Specs. Skills. Everything your AI agent needs before writing the first line of code.** 🏗️

*Because an agent without context is an intern with root access.*

**[English]** | [Español](README_ES.md)

[Quick Start](#-quick-start) · [Context](#-context-generation) · [Specs](#-spec-driven-development) · [Skills](#-agent-skills) · [MCP Server](#-mcp-server) · [Language Guides](#-language-specific-guides) · [Architecture](#%EF%B8%8F-architecture)

</div>

---

## 🎯 The Problem

You tell your agent: *"Build a payments API in Go with microservices"*

And the agent, with all its capability, improvises:
- Folder structures nobody asked for
- Patterns that contradict your architecture
- Decisions you'll revert in the next session
- Zero continuity between sessions

**It's not the agent's fault. It starts from scratch. Every. Single. Time.** 🔄

## 💡 The Solution

**Codify** equips your AI agent with three things it needs to stop improvising:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Context    │     │    Specs     │     │   Skills     │
│              │     │              │     │              │
│  What the    │     │  What to     │     │  How to      │
│  project is  │────▶│  build next  │     │  do things   │
│              │     │              │     │  right       │
│  generate    │     │  spec        │     │  skills      │
│  analyze     │     │  --with-specs│     │              │
└──────────────┘     └──────────────┘     └──────────────┘
     Memory             Plan              Abilities
```

- **Context** gives the agent architectural memory — stack, patterns, conventions, domain knowledge
- **Specs** give the agent an implementation plan — features, acceptance criteria, task breakdowns
- **Skills** give the agent reusable abilities — how to commit, version, design entities, review code

It follows the [AGENTS.md standard](https://github.com/anthropics/AGENTS.md) — an open specification backed by the Linux Foundation for providing AI agents with project context. Files work out of the box with Claude Code, Cursor, Codex, and any agent that reads the standard.

## ✨ Before and after

### 😱 Without Codify

```
You: "Create a payments API in Go"

Agent: *creates main.go with everything in one file*
You: "No, use Clean Architecture"
Agent: *creates structure but mixes domain with infra*
You: "Repositories go in infrastructure"
Agent: *refactors for the third time*
You: "What about the BDD tests I asked for yesterday?"
Agent: "BDD tests? This is the first time you've mentioned that"
You: "At least commit this properly"
Agent: *writes "update code" as commit message*

Result: 45 minutes correcting the agent 😤
```

### 🚀 With Codify

```
You: "Create a payments API in Go"

Agent: *reads AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md*
Agent: "I see you use DDD with Clean Architecture, PostgreSQL,
        BDD testing with Godog, and idiomatic Go patterns.
        I'll create the payments endpoint in internal/domain/payment/
        following your patterns and concurrency conventions."

Agent: *reads SKILL.md for conventional commits*
Agent: "Done. Here's the commit following Conventional Commits:
        feat(payment): add payment domain entity with Stripe integration"

Result: Coherent code from the first line ✨
```

## ⚡ Quick Start

### Installation

```bash
# Homebrew (macOS/Linux — no Go required)
brew tap jorelcb/tap
brew install codify

# Or via go install
go install github.com/jorelcb/codify/cmd/codify@latest

# Or download pre-built binaries from GitHub Releases
# https://github.com/jorelcb/codify/releases
```

### Three ways to equip your agent

```bash
# 1. Set your API key (Claude or Gemini)
export ANTHROPIC_API_KEY="sk-ant-..."   # for Claude (default)
# or
export GEMINI_API_KEY="AI..."           # for Gemini

# ── Context: give your agent project memory ──
codify generate payment-service \
  --description "Payment microservice in Go with gRPC, PostgreSQL and Kafka. \
  DDD with Clean Architecture. Stripe as payment processor." \
  --language go

# ── Specs: give your agent an implementation plan ──
codify spec payment-service \
  --from-context ./output/payment-service/

# ── Skills: give your agent reusable abilities ──
codify skills
# Interactive mode guides you through category, preset, mode, and target.
# No API key needed for static mode.
```

### What you'll see

```
🚀 Generating context for: payment-service
  Model: claude-sonnet-4-6
  Preset: default
  Language: go

  [1/5] Generating AGENTS.md... ✓
  [2/5] Generating CONTEXT.md... ✓
  [3/5] Generating INTERACTIONS_LOG.md... ✓
  [4/5] Generating DEVELOPMENT_GUIDE.md... ✓
  [5/5] Generating IDIOMS.md... ✓

📁 Output: output/payment-service/
  ├── AGENTS.md                → Root file (tech stack, commands, conventions)
  └── context/
      ├── CONTEXT.md           → Architecture and technical design
      ├── INTERACTIONS_LOG.md  → Session log and ADRs
      ├── DEVELOPMENT_GUIDE.md → Work methodology, testing, security
      └── IDIOMS.md            → Language-specific patterns (Go)

✅ Done! 5 files generated
   Total tokens: ~18,200
```

---

## 📋 Context Generation

The foundation. Generates files following the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard that give your agent deep project memory.

### `generate` command — Context from a description

```bash
codify generate payment-service \
  --description "Payment microservice in Go with gRPC and PostgreSQL" \
  --language go
```

### `analyze` command — Context from an existing project

Scans an existing codebase — auto-detects language, framework, dependencies, directory structure, README, infrastructure signals (Docker, CI/CD, Makefile) — and generates context files from what it finds.

```bash
codify analyze /path/to/my-project
```

### Generated files

| File | What it does |
|------|-------------|
| `AGENTS.md` | Root file: tech stack, commands, conventions, structure |
| `CONTEXT.md` | Architecture, components, data flow, design decisions |
| `INTERACTIONS_LOG.md` | Session log and ADRs |
| `DEVELOPMENT_GUIDE.md` | Work methodology, testing practices, security, delivery expectations |
| `IDIOMS.md` | Language-specific concurrency, error handling, conventions *(requires `--language`)* |

Place these files at your project root. Compatible agents (Claude Code, Cursor, Codex, etc.) read them automatically.

### Options

```bash
codify generate <name> [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Project description *(required unless `--from-file`)* | — |
| `--from-file` | `-f` | Read description from file *(alternative to `-d`)* | — |
| `--preset` | `-p` | Template preset (`default`, `neutral`) | `default` |
| `--model` | `-m` | LLM model (`claude-*` or `gemini-*`) | `claude-sonnet-4-6` |
| `--language` | `-l` | Language (activates idiomatic guides) | — |
| `--locale` | | Output language (`en`, `es`) | `en` |
| `--with-specs` | | Also generate SDD specs after context | `false` |
| `--type` | `-t` | Project type hint (api, cli, lib...) | — |
| `--architecture` | `-a` | Architecture hint | — |

---

## 📐 Spec-Driven Development

From existing context, generates implementation-ready specifications. This enables **AI Spec-Driven Development (AI SDD)**: your agent implements a spec, not an improvisation.

```
Your idea → generate (context) → spec (specifications) → Agent writes code with full context
```

### `spec` command

```bash
codify spec payment-service \
  --from-context ./output/payment-service/
```

### `--with-specs` — Full pipeline in one command

Available on both `generate` and `analyze`. Chains context generation + spec generation + AGENTS.md update in a single run:

```bash
codify generate my-api \
  --description "REST API in Go with PostgreSQL" \
  --language go \
  --with-specs
```

### Generated spec files

| File | What it does |
|------|-------------|
| `CONSTITUTION.md` | Project DNA: stack, principles, constraints |
| `SPEC.md` | Feature specs with acceptance criteria |
| `PLAN.md` | Technical design and architecture decisions |
| `TASKS.md` | Task breakdown with dependencies and priority |

---

## 🧩 Agent Skills

Skills are reusable [Agent Skills](https://agentskills.io) (SKILL.md files) that teach your agent *how* to perform specific tasks — following Conventional Commits, applying DDD patterns, doing code reviews, versioning releases. They complement context files: context tells the agent *what* your project is, skills tell it *how* to do things right.

### Two modes

| Mode | What it does | API key | Cost | Speed |
|------|-------------|---------|------|-------|
| **Static** | Delivers pre-built skills from the embedded catalog. Production-ready, ecosystem-aware frontmatter. | Not needed | Free | Instant |
| **Personalized** | LLM adapts skills to your project — examples use your domain, language, and stack. | Required | ~pennies | ~10s |

### Interactive mode

Just run `codify skills` — the interactive menu guides you through every choice:

```bash
codify skills
# → Select category (architecture, workflow)
# → Select preset (clean, neutral, conventional-commit, ...)
# → Select mode (static or personalized)
# → Select target ecosystem (claude, codex, antigravity)
# → Select locale, output path
# → If personalized: describe your project, choose model
```

### CLI mode

```bash
# Static: instant delivery, no API key
codify skills --category workflow --preset all --mode static

# Personalized: LLM-adapted to your project
codify skills --category architecture --preset clean --mode personalized \
  --context "Go microservice with DDD, Godog BDD, PostgreSQL"

# Architecture skills for Codex ecosystem
codify skills --category architecture --preset neutral --target codex
```

### Skill catalog

| Category | Preset | Skills |
|----------|--------|--------|
| `architecture` | `clean` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `workflow` | `conventional-commit` | Conventional Commits |
| `workflow` | `semantic-versioning` | Semantic Versioning |
| `workflow` | `all` | All workflow skills combined |

### Target ecosystems

Each ecosystem gets specific YAML frontmatter and output paths:

| Target | Frontmatter | Output path |
|--------|-------------|-------------|
| `claude` *(default)* | `name`, `description`, `user-invocable: true` | `.claude/skills/` |
| `codex` | `name`, `description` | `.agents/skills/` |
| `antigravity` | `name`, `description`, `triggers` | `.agents/skills/` |

### Options

```bash
codify skills [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--category` | Skill category (`architecture`, `workflow`) | *(interactive)* |
| `--preset` | Preset within category | *(interactive)* |
| `--mode` | Generation mode: `static` or `personalized` | *(interactive)* |
| `--context` | Project description for personalized mode | — |
| `--target` | Target ecosystem (`claude`, `codex`, `antigravity`) | `claude` |
| `--model` | `-m` | LLM model (personalized mode only) | `claude-sonnet-4-6` |
| `--locale` | Output language (`en`, `es`) | `en` |
| `--output` | `-o` | Output directory | `.claude/skills/` |

---

## 🔌 MCP Server

Use Codify as an **MCP server** — your AI coding agent calls the tools directly, no manual CLI needed.

### Install

```bash
go install github.com/jorelcb/codify/cmd/codify@latest
```

### Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "codify": {
      "command": "codify",
      "args": ["serve"],
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-...",
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

### Codex CLI

```bash
# Register the MCP server
codex mcp add codify -- codify serve
```

### Gemini CLI

Add to `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "codify": {
      "command": "codify",
      "args": ["serve"],
      "env": {
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

> Set the API key(s) for the provider(s) you want to use. The provider is auto-detected from the `model` parameter. If the binary is not in your PATH, use the full path (e.g., `/Users/you/go/bin/codify`).

### Available MCP Tools

#### Generative tools (require LLM API key)

| Tool | Description |
|------|-------------|
| `generate_context` | Generate context files from a project description |
| `generate_specs` | Generate SDD specs from existing context files |
| `analyze_project` | Scan an existing project and generate context from its structure |
| `generate_skills` | Generate Agent Skills — supports `static` (instant) and `personalized` (LLM-adapted) modes |

All generative tools support `locale` (`en`/`es`) and `model` parameters. `generate_context` and `analyze_project` also accept `with_specs`. `generate_skills` accepts `mode`, `category`, `preset`, and `project_context`.

#### Knowledge tools (no API key needed)

| Tool | Description |
|------|-------------|
| `commit_guidance` | Conventional Commits spec and behavioral context for generating proper commit messages |
| `version_guidance` | Semantic Versioning spec and behavioral context for determining version bumps |

Knowledge tools inject behavioral context into the calling agent — the agent receives the spec and instructions, then applies them to the current task. Supports `locale` (`en`/`es`).

### Example prompts

```
"Generate context for a payment microservice in Go with gRPC and PostgreSQL"
→ Agent calls generate_context

"Analyze my project at /path/to/my-app and generate specs"
→ Agent calls analyze_project with with_specs=true

"Generate workflow skills for my project"
→ Agent calls generate_skills with mode=static, category=workflow, preset=all

"Create DDD skills adapted to my Go project with Clean Architecture"
→ Agent calls generate_skills with mode=personalized, project_context="Go with DDD..."

"Help me commit these changes following conventional commits"
→ Agent calls commit_guidance, receives the spec, crafts the message

"What version should I release based on recent changes?"
→ Agent calls version_guidance, receives semver rules, analyzes commits
```

---

## 🌐 Language-Specific Guides

When you pass `--language`, the tool generates an additional `IDIOMS.md` file with patterns and conventions specific to that language. This is one of the most impactful features — it gives your agent deep knowledge of idiomatic patterns instead of generic advice.

| Language | What IDIOMS.md covers |
|----------|----------------------|
| `go` | Goroutines, channels, WaitGroups, `context.Context`, error wrapping with `%w`, table-driven tests |
| `javascript` | async/await, `Promise.all`, `AbortController`, worker threads, TypeScript, ESM, Jest patterns |
| `python` | asyncio, multiprocessing, type hints, pydantic, pytest fixtures, `ruff` |

```bash
# Go project with idiomatic guides
codify generate my-api -d "REST API in Go" --language go

# TypeScript SDK with JS idioms
codify generate my-sdk -d "SDK in TypeScript" --language javascript

# Python service with async patterns
codify generate my-service -d "FastAPI service" --language python
```

Without `--language`, the tool generates 4 files. With it, you get 5 — and significantly richer output.

## 🎭 Presets

Choose the philosophy for your contexts:

### `--preset default` *(default)*

Recommended: **DDD + Clean Architecture + BDD**. Includes:
- Strict layer separation (Domain → Application → Infrastructure → Interfaces)
- BDD testing with coverage targets (80% domain, 70% application)
- OpenTelemetry observability
- Mandatory dependency injection
- MUST/MUST NOT constraints
- Development methodology and self-validation checklist

```bash
codify generate my-api \
  --description "Inventory management REST API in Go"
# Uses default preset automatically
```

### `--preset neutral`

No architectural stance. Lets the LLM adapt the structure to the project:

```bash
codify generate my-api \
  --description "Inventory management REST API in Go" \
  --preset neutral
```

### `--from-file` — Rich descriptions from files

For detailed project descriptions (design docs, RFCs, 6-pagers), use `--from-file` instead of `--description`:

```bash
codify generate my-api \
  --from-file ./docs/project-description.md \
  --language go
```

The file content becomes the project description. Supports any text format — markdown, plain text, etc. Mutually exclusive with `--description`.

## 🏗️ Architecture

Built in Go with what it preaches — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Pure business logic
│   ├── project/         Project entity (aggregate root)
│   ├── catalog/         Declarative skill catalog and metadata registry
│   ├── shared/          Value objects, domain errors
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Use cases (CQRS)
│   ├── command/         GenerateContext, GenerateSpec, GenerateSkills, DeliverStaticSkills
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementations
│   ├── llm/             LLM providers (Claude, Gemini) + prompt builder
│   ├── template/        Template loader (locale + preset + language-aware)
│   ├── scanner/         Project scanner (language, deps, framework detection)
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 Entry points
    ├── cli/commands/    generate, analyze, spec, skills, serve, list
    └── mcp/             MCP server (stdio + HTTP transport, 6 tools)
```

### Template system

```
templates/
├── en/                          English locale
│   ├── default/                 Recommended preset (DDD/Clean Architecture)
│   │   ├── agents.template
│   │   ├── context.template
│   │   ├── interactions.template
│   │   └── development_guide.template
│   ├── neutral/                 Generic preset (no architectural opinions)
│   │   └── (same files)
│   ├── spec/                    Specification templates (AI SDD)
│   │   ├── constitution.template
│   │   ├── spec.template
│   │   ├── plan.template
│   │   └── tasks.template
│   ├── skills/                  Agent Skills templates (static + LLM guides)
│   │   ├── default/             Architecture: Clean (DDD, BDD, CQRS, Hexagonal)
│   │   ├── neutral/             Architecture: Neutral (review, testing, API)
│   │   └── workflow/            Workflow (conventional commits, semver)
│   └── languages/               Language-specific idiomatic guides
│       ├── go/idioms.template
│       ├── javascript/idioms.template
│       └── python/idioms.template
└── es/                          Spanish locale (same structure)
```

The golden rule: `Infrastructure → Application → Domain`. Nothing in domain depends on anything external.

See [context/CONTEXT.md](context/CONTEXT.md) for full architectural details.

## 🧪 Tests

```bash
# All tests
go test ./...

# BDD with Godog
go test ./tests/...
```

## 📊 Project status

**v1.10.0** 🎉

✅ **Working:**
- Multi-provider LLM support (Anthropic Claude + Google Gemini)
- **Context generation** with streaming (`generate`, `analyze`)
- **SDD spec generation** from existing context (`spec`, `--with-specs`)
- **Agent Skills** with dual mode (static/personalized), interactive guided selection, and declarative catalog
- Skill categories (architecture, workflow) with ecosystem-aware frontmatter (Claude, Codex, Antigravity)
- MCP Server mode (stdio + HTTP transport) with 6 tools
- MCP knowledge tools (commit_guidance, version_guidance) — no API key needed
- Preset system (default: DDD/Clean, neutral: generic)
- AGENTS.md standard as root file
- Language-specific idiomatic guides (Go, JavaScript, Python)
- Anti-hallucination grounding rules in prompts
- CLI with Cobra + interactive menus (charmbracelet/huh)
- Homebrew formula distribution (macOS/Linux)

🚧 **Coming next:**
- Testing skill category (unit, integration, e2e)
- End-to-end integration tests
- Retries and rate limit handling
- MCP server authentication (OAuth/BYOK for remote deployments)

## 💡 FAQ

**Which LLM providers are supported?**
Anthropic Claude (default) and Google Gemini. Set `ANTHROPIC_API_KEY` for Claude or `GEMINI_API_KEY` for Gemini. The provider is auto-detected from the `--model` flag: `claude-*` models use Anthropic, `gemini-*` models use Google.

**How much does each generation cost?**
4-5 API calls for `generate` (depending on `--language`), 4 for `spec`. Skills in static mode are free (no API calls). Personalized skills use 1 API call per skill. Each generation costs pennies with either provider.

**Do I need an API key for skills?**
Only for personalized mode. Static mode delivers pre-built skills instantly from the embedded catalog — no LLM, no API key, no cost.

**What's the difference between static and personalized skills?**
Static skills are production-ready, generic best practices delivered instantly. Personalized skills use an LLM to adapt examples, naming, and patterns to your specific project context (language, domain, stack).

**Are the templates fixed?**
They're structural guides, not renderable output. The LLM generates intelligent, project-specific content following the template structure.

**Can I customize the templates?**
You can create your own presets in `templates/<locale>/`. Each preset needs 4 files: `agents.template`, `context.template`, `interactions.template`, and `development_guide.template`. Language-specific templates go in `templates/<locale>/languages/<lang>/idioms.template`.

**Which agents support the generated files?**
Any agent compatible with the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard: Claude Code, Cursor, GitHub Copilot Workspace, Codex, and more.

**What's AI Spec-Driven Development?**
A methodology where you generate context and specifications *before* writing code. Your agent implements a spec, not an improvisation. `generate` creates the blueprint, `spec` creates the implementation plan.

## 📚 Documentation

- [📋 AGENTS.md](AGENTS.md) — Project context for AI agents
- [🏛️ Architecture](context/CONTEXT.md) — DDD/Clean Architecture details
- [📝 Changelog](CHANGELOG.md) — Change history
- [📐 Specs](specs/) — Technical specifications (SDD)

## 📄 License

Apache License 2.0 — see [LICENSE](LICENSE).

---

<div align="center">

**Context. Specs. Skills. Your agent, fully equipped.** 🧠

*"An agent without context is an intern with root access"*

⭐ If this helped you, give it a star — it keeps us building

[🐛 Report bug](https://github.com/jorelcb/codify/issues) · [💡 Request feature](https://github.com/jorelcb/codify/issues)

</div>