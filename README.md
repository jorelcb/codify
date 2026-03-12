# 🧠 AI Context Generator

<div align="center">

[![Version](https://img.shields.io/badge/version-2.3.0-blue?style=for-the-badge)](https://github.com/jorelcb/ai-context-generator/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.21)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Give your AI agent the master blueprint it needs before writing the first line of code** 🏗️

*Because an agent without context is an intern with root access.*

**[English]** | [Español](README_ES.md)

[Quick Start](#-quick-start) · [MCP Server](#-mcp-server) · [Features](#-features) · [Skills](#-agent-skills) · [Language Guides](#-language-specific-guides) · [Presets](#-presets) · [Architecture](#%EF%B8%8F-architecture)

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

**AI Context Generator** takes your project description and generates intelligent context files using LLMs (Anthropic Claude or Google Gemini). Files that give your agent the master blueprint, domain constraints, and architectural memory it needs.

It follows the [AGENTS.md standard](https://github.com/anthropics/AGENTS.md) — an open specification backed by the Linux Foundation for providing AI agents with project context. This means the files work out of the box with Claude Code, Cursor, Codex, and any agent that reads the standard.

## 🧭 AI Spec-Driven Development

This tool enables a methodology we call **AI Spec-Driven Development (AI SDD)**: instead of going straight from an idea to code, you first generate a rich specification layer that grounds your agent's work.

```
Your idea → generate (context) → spec (specifications) → Agent writes code with full context
```

The `generate` command creates the **architectural blueprint** — what the project is, how it's built, what patterns it follows. The `spec` command takes that blueprint and produces **implementation-ready specifications** — features, acceptance criteria, technical plans, and task breakdowns.

Your agent doesn't improvise. It implements a spec. That's the difference.

## ✨ Before and after

### 😱 Without context (the current reality)

```
You: "Create a payments API in Go"

Agent: *creates main.go with everything in one file*
You: "No, use Clean Architecture"
Agent: *creates structure but mixes domain with infra*
You: "Repositories go in infrastructure"
Agent: *refactors for the third time*
You: "What about the BDD tests I asked for yesterday?"
Agent: "BDD tests? This is the first time you've mentioned that"

Result: 45 minutes correcting the agent 😤
```

### 🚀 With AI Context Generator

```
You: "Create a payments API in Go"

Agent: *reads AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md and IDIOMS.md*
Agent: "I see you use DDD with Clean Architecture, PostgreSQL,
        BDD testing with Godog, and idiomatic Go patterns.
        I'll create the payments endpoint in internal/domain/payment/
        following your patterns and concurrency conventions."

Result: Coherent code from the first line ✨
```

## ⚡ Quick Start

### Installation

```bash
# go install (recommended)
go install github.com/jorelcb/ai-context-generator/cmd/ai-context-generator@latest

# Or build from source
git clone https://github.com/jorelcb/ai-context-generator.git
cd ai-context-generator && go build -o bin/ai-context-generator ./cmd/ai-context-generator
```

### Your first context in 30 seconds

```bash
# 1. Set your API key (Claude or Gemini)
export ANTHROPIC_API_KEY="sk-ant-..."   # for Claude (default)
# or
export GEMINI_API_KEY="AI..."           # for Gemini

# 2. Describe your project (with language for idiomatic guides)
ai-context-generator generate payment-service \
  --description "Payment microservice in Go with gRPC, PostgreSQL and Kafka. \
  DDD with Clean Architecture. Stripe as payment processor." \
  --language go

# 3. Use Gemini instead of Claude
ai-context-generator generate payment-service \
  --description "Payment microservice in Go" \
  --model gemini-3.1-pro-preview

# 4. Done. Files generated.
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

## 🔌 MCP Server

Use AI Context Generator as an **MCP (Model Context Protocol) server** — no CLI needed. Your AI agent calls the tools directly.

### Setup for Claude Desktop

**1. Install the binary:**

```bash
go install github.com/jorelcb/ai-context-generator/cmd/ai-context-generator@latest
```

**2. Add to Claude Desktop config** (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "ai-context-generator": {
      "command": "ai-context-generator",
      "args": ["serve"],
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-...",
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

> Set the API key(s) for the provider(s) you want to use. The provider is selected automatically based on the `model` parameter.

> If `ai-context-generator` is not in your PATH, use the full path (e.g., `~/go/bin/ai-context-generator`).

**3. Restart Claude Desktop.** The tools appear automatically.

### Setup for Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "ai-context-generator": {
      "command": "ai-context-generator",
      "args": ["serve"],
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-...",
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

### Setup for Cursor

Add in **Settings > MCP Servers**:

| Field | Value |
|-------|-------|
| Name | `ai-context-generator` |
| Command | `ai-context-generator serve` |
| Environment | `ANTHROPIC_API_KEY=sk-ant-...` |

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `generate_context` | Generate context files from a project description |
| `generate_specs` | Generate SDD specs from existing context files |
| `analyze_project` | Scan an existing project and generate context from its structure |
| `generate_skills` | Generate reusable Agent Skills based on architectural presets |

All tools support `locale` (`en`/`es`), `model`, and `preset` parameters. `generate_context` and `analyze_project` also accept `with_specs` to chain spec generation automatically.

### Example prompts (Claude Desktop / Claude Code)

```
"Generate context for a payment microservice in Go with gRPC and PostgreSQL"
→ Agent calls generate_context

"Analyze my project at /path/to/my-app and generate specs"
→ Agent calls analyze_project with with_specs=true

"Generate specs from the context in ./output/my-api"
→ Agent calls generate_specs
```

---

## 🎨 Features

### 📋 `generate` command — Context for your agent

Generates files following the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard:

| File | What it does |
|------|-------------|
| `AGENTS.md` | Root file: tech stack, commands, conventions, structure |
| `CONTEXT.md` | Architecture, components, data flow, design decisions |
| `INTERACTIONS_LOG.md` | Session log and ADRs |
| `DEVELOPMENT_GUIDE.md` | Work methodology, testing practices, security, delivery expectations |
| `IDIOMS.md` | Language-specific concurrency, error handling, conventions *(requires `--language`)* |

Place these files at your project root. Compatible agents (Claude Code, Cursor, Codex, etc.) read them automatically.

### 📐 `spec` command — AI SDD specifications

From existing context, generates technical specifications ready for implementation:

```bash
ai-context-generator spec payment-service \
  --from-context ./output/payment-service/
```

| File | What it does |
|------|-------------|
| `CONSTITUTION.md` | Project DNA: stack, principles, constraints |
| `SPEC.md` | Feature specs with acceptance criteria |
| `PLAN.md` | Technical design and architecture decisions |
| `TASKS.md` | Task breakdown with dependencies and priority |

### 🔎 `analyze` command — Context from existing projects

Scans an existing codebase and generates context files automatically:

```bash
ai-context-generator analyze /path/to/my-project --with-specs
```

Auto-detects language, framework, dependencies, directory structure, README, existing context files, and infrastructure signals (Docker, CI/CD, Makefile, etc.). Everything feeds into the LLM for richer, project-aware generation.

### ⚡ `--with-specs` — Full pipeline in one command

Available on both `generate` and `analyze`. Chains context generation + spec generation + AGENTS.md update in a single run:

```bash
ai-context-generator generate my-api \
  --description "REST API in Go with PostgreSQL" \
  --language go \
  --with-specs
```

### 🧩 `skills` command — Agent Skills

Generates reusable [Agent Skills](https://agentskills.io) (SKILL.md) based on architectural presets. Skills are cross-project — install them globally and any AI agent will use them when relevant.

```bash
# Default preset: DDD, Clean Arch, BDD, CQRS, Hexagonal
ai-context-generator skills

# Neutral preset for Codex
ai-context-generator skills --preset neutral --target codex

# For Antigravity IDE in Spanish
ai-context-generator skills --target antigravity --locale es
```

| Preset | Skills generated |
|--------|-----------------|
| `default` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port/adapter |
| `neutral` | Code review, test strategy, safe refactoring, API design |

Target ecosystems: `claude` (default), `codex`, `antigravity` — each gets ecosystem-specific YAML frontmatter.

### 🔍 `list` command — Generated projects

```bash
ai-context-generator list
```

## 🌐 Language-Specific Guides

When you pass `--language`, the tool generates an additional `IDIOMS.md` file with patterns and conventions specific to that language. This is one of the most impactful features — it gives your agent deep knowledge of idiomatic patterns instead of generic advice.

| Language | What IDIOMS.md covers |
|----------|----------------------|
| `go` | Goroutines, channels, WaitGroups, `context.Context`, error wrapping with `%w`, table-driven tests |
| `javascript` | async/await, `Promise.all`, `AbortController`, worker threads, TypeScript, ESM, Jest patterns |
| `python` | asyncio, multiprocessing, type hints, pydantic, pytest fixtures, `ruff` |

```bash
# Go project with idiomatic guides
ai-context-generator generate my-api -d "REST API in Go" --language go

# TypeScript SDK with JS idioms
ai-context-generator generate my-sdk -d "SDK in TypeScript" --language javascript

# Python service with async patterns
ai-context-generator generate my-service -d "FastAPI service" --language python
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
ai-context-generator generate my-api \
  --description "Inventory management REST API in Go"
# Uses default preset automatically
```

### `--preset neutral`

No architectural stance. Lets the LLM adapt the structure to the project:

```bash
ai-context-generator generate my-api \
  --description "Inventory management REST API in Go" \
  --preset neutral
```

### `--from-file` — Rich descriptions from files

For detailed project descriptions (design docs, RFCs, 6-pagers), use `--from-file` instead of `--description`:

```bash
ai-context-generator generate my-api \
  --from-file ./docs/project-description.md \
  --language go
```

The file content becomes the project description. Supports any text format — markdown, plain text, etc. Mutually exclusive with `--description`.

## ⚙️ Options

```bash
ai-context-generator generate <name> [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Project description *(required unless `--from-file`)* | — |
| `--from-file` | `-f` | Read description from file *(alternative to `-d`)* | — |
| `--preset` | `-p` | Template preset | `default` |
| `--model` | `-m` | LLM model (`claude-*` or `gemini-*`) | `claude-sonnet-4-6` |
| `--language` | `-l` | Language (activates idiomatic guides) | — |
| `--locale` | | Output language (`en`, `es`) | `en` |
| `--with-specs` | | Also generate SDD specs after context | `false` |
| `--type` | `-t` | Project type hint (api, cli, lib...) | — |
| `--architecture` | `-a` | Architecture hint | — |

## 🏗️ Architecture

Built in Go with what it preaches — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Pure business logic
│   ├── project/         Project entity (aggregate root)
│   ├── shared/          Value objects, domain errors
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Use cases (CQRS)
│   ├── command/         GenerateContext, GenerateSpec
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
    └── mcp/             MCP server (stdio + HTTP transport, 4 tools)
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
│   ├── skills/                  Agent Skills templates
│   │   ├── default/             DDD, Clean Arch, BDD, CQRS, Hexagonal
│   │   └── neutral/             Code review, testing, refactoring, API design
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

**v2.3.0** 🎉

✅ **Working:**
- Multi-provider LLM support (Anthropic Claude + Google Gemini)
- Context generation with streaming
- SDD spec generation from existing context
- Agent Skills generation (SKILL.md) for Claude Code, Codex, Antigravity
- MCP Server mode (stdio + HTTP transport)
- `analyze` command — scan existing projects and generate context
- `--with-specs` flag — full pipeline in one command
- Preset system (default DDD/BDD, neutral)
- AGENTS.md standard as root file
- Language-specific idiomatic guides (Go, JavaScript, Python)
- Anti-hallucination grounding rules in prompts
- CLI with Cobra (generate, analyze, spec, skills, serve, list)

🚧 **Coming next:**
- End-to-end integration tests
- Retries and rate limit handling
- Interactive mode (wizard)
- MCP server authentication (OAuth/BYOK for remote deployments)
- Binary builds and distribution

## 💡 FAQ

**Which LLM providers are supported?**
Anthropic Claude (default) and Google Gemini. Set `ANTHROPIC_API_KEY` for Claude or `GEMINI_API_KEY` for Gemini. The provider is auto-detected from the `--model` flag: `claude-*` models use Anthropic, `gemini-*` models use Google.

**How much does each generation cost?**
4-5 API calls for `generate` (depending on `--language`), 4 for `spec`. Each generation costs pennies with either provider.

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

**Built to supercharge AI-assisted development** 🧠

*"An agent without context is an intern with root access"*

⭐ If this helped you, give it a star — it keeps us building

[🐛 Report bug](https://github.com/jorelcb/ai-context-generator/issues) · [💡 Request feature](https://github.com/jorelcb/ai-context-generator/issues)

</div>