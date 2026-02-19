# 🧠 AI Context Generator

<div align="center">

[![Version](https://img.shields.io/badge/version-2.0.0-blue?style=for-the-badge)](https://github.com/jorelcb/ai-context-generator/releases)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.21)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Powered%20by-Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)

**Give your AI agent the master blueprint it needs before writing the first line of code** 🏗️

*Because an agent without context is an intern with root access.*

**[English]** | [Español](README_ES.md)

[Quick Start](#-quick-start) · [Features](#-features) · [Presets](#-presets) · [Architecture](#%EF%B8%8F-architecture) · [Docs](#-documentation)

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

**AI Context Generator** takes your project description and generates intelligent context files using Anthropic Claude. Files that give your agent the master blueprint, domain constraints, and architectural memory it needs.

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

Agent: *reads AGENTS.md and CONTEXT.md*
Agent: "I see you use DDD with Clean Architecture, PostgreSQL,
        and BDD testing with Godog. I'll create the payments
        endpoint in internal/domain/payment/ following your patterns."

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
# 1. Set your API key
export ANTHROPIC_API_KEY="sk-ant-..."

# 2. Describe your project
ai-context-generator generate payment-service \
  --description "Payment microservice in Go with gRPC, PostgreSQL and Kafka. \
  DDD with Clean Architecture. Stripe as payment processor."

# 3. Done. Files generated.
```

### What you'll see

```
🚀 Generating context for: payment-service
  Model: claude-sonnet-4-6
  Preset: default

  [1/3] Generating AGENTS.md... ✓
  [2/3] Generating CONTEXT.md... ✓
  [3/3] Generating INTERACTIONS_LOG.md... ✓

📁 Output: output/payment-service/
  ├── AGENTS.md              → Root file (tech stack, commands, conventions)
  └── context/
      ├── CONTEXT.md         → Architecture and technical design
      └── INTERACTIONS_LOG.md → Session log and ADRs

✅ Done! 3 files generated
   Total tokens: ~12,450
```

## 🎨 Features

### 📋 `generate` command — Context for your agent

Generates files following the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard:

| File | What it does |
|------|-------------|
| `AGENTS.md` | Root file: tech stack, commands, conventions, structure |
| `CONTEXT.md` | Architecture, components, data flow, design decisions |
| `INTERACTIONS_LOG.md` | Session log and ADRs |

Place these files at your project root. Compatible agents (Claude Code, Cursor, Codex, etc.) read them automatically.

### 📐 `spec` command — SDD specifications

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

### 🔍 `list` command — Generated projects

```bash
ai-context-generator list
```

## 🎭 Presets

Choose the philosophy for your contexts:

### `--preset default` *(default)*

Opinionated: **DDD + Clean Architecture + BDD**. Includes:
- Strict layer separation (Domain → Application → Infrastructure → Interfaces)
- BDD testing with coverage targets (80% domain, 70% application)
- OpenTelemetry observability
- Mandatory dependency injection
- MUST/MUST NOT constraints

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

## ⚙️ Options

```bash
ai-context-generator generate <name> [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Project description *(required)* | — |
| `--preset` | `-p` | Template preset | `default` |
| `--model` | `-m` | Claude model | `claude-sonnet-4-6` |
| `--language` | `-l` | Language hint | — |
| `--type` | `-t` | Project type hint (api, cli, lib...) | — |
| `--architecture` | `-a` | Architecture hint | — |

## 🏗️ Architecture

Built in Go with what it preaches — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Pure business logic
│   ├── project/         Project entity (aggregate root)
│   ├── template/        Template entity
│   ├── shared/          Value objects, domain errors
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Use cases (CQRS)
│   ├── command/         GenerateContext, GenerateSpec
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementations
│   ├── llm/             Anthropic Claude adapter + prompt builder
│   ├── template/        Template loader with configurable mapping
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 CLI with Cobra
    └── cli/commands/    generate, spec, list
```

The golden rule: `Infrastructure → Application → Domain`. Nothing in domain depends on anything external.

See [ARCHITECTURE.md](ARCHITECTURE.md) for full details.

## 🧪 Tests

```bash
# All tests
go test ./...

# BDD with Godog
go test ./tests/...
```

## 📊 Project status

**v2.0.0** 🎉

✅ **Working:**
- Context generation with Claude API (streaming)
- SDD spec generation from existing context
- Preset system (default DDD/BDD, neutral)
- AGENTS.md standard as root file
- CLI with Cobra (generate, spec, list)
- Templates as structural guides for the LLM

🚧 **Coming next:**
- End-to-end integration tests
- Retries and rate limit handling
- Interactive mode (wizard)
- Second LLM provider
- Binary builds and distribution

👉 [Full roadmap](ROADMAP.md)

## 💡 FAQ

**Do I need an Anthropic API key?**
Yes. Export it as `ANTHROPIC_API_KEY`. Get one at [console.anthropic.com](https://console.anthropic.com).

**How much does each generation cost?**
About 3 API calls for `generate`, 4 for `spec`. With claude-sonnet-4-6, each generation costs pennies.

**Does it work with other LLMs?**
Currently only Anthropic Claude. The `LLMProvider` interface is designed to add more providers without changing the core.

**Are the templates fixed?**
They're structural guides, not renderable output. The LLM generates intelligent, project-specific content following the template structure.

**Can I customize the templates?**
You can create your own presets in the `templates/` directory. Each preset needs 3 files: `agents.template`, `context.template`, `interactions.template`.

**Which agents support the generated files?**
Any agent compatible with the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard: Claude Code, Cursor, GitHub Copilot Workspace, Codex, and more.

## 📚 Documentation

- [🏛️ Architecture Guide](ARCHITECTURE.md) — DDD/Clean Architecture
- [🚀 Getting Started](GETTING_STARTED.md) — Step-by-step guide
- [🗺️ Roadmap](ROADMAP.md) — Development plan
- [📝 Changelog](context/CHANGELOG.md) — Change history

## 📄 License

MIT License — see [LICENSE](LICENSE).

---

<div align="center">

**Built to supercharge AI-assisted development** 🧠

*"An agent without context is an intern with root access"*

⭐ If this helped you, give it a star — it keeps us building

[🐛 Report bug](https://github.com/jorelcb/ai-context-generator/issues) · [💡 Request feature](https://github.com/jorelcb/ai-context-generator/issues) · [🗺️ See roadmap](ROADMAP.md)

</div>
