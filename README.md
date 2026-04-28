# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-1.18.0-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.23)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Context. Specs. Skills. Workflows. Everything your AI agent needs before writing the first line of code.** 🏗️

*Because an agent without context is an intern with root access.*

**[English]** | [Español](README_ES.md)

[Quick Start](#-quick-start) · [Context](#-context-generation) · [Specs](#-spec-driven-development) · [Skills](#-agent-skills) · [Workflows](#-workflows) · [MCP Server](#-mcp-server) · [Language Guides](#-language-specific-guides) · [Architecture](#%EF%B8%8F-architecture)

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

**Codify** equips your AI agent with four things it needs to stop improvising:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Context    │     │    Specs     │     │   Skills     │     │  Workflows   │
│              │     │              │     │              │     │              │
│  What the    │     │  What to     │     │  How to      │     │  Multi-step  │
│  project is  │────▶│  build next  │     │  do things   │     │  recipes     │
│              │     │              │     │  right       │     │  on demand   │
│  generate    │     │  spec        │     │  skills      │     │  workflows   │
│  analyze     │     │  --with-specs│     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
     Memory             Plan              Abilities          Orchestration
```

- **Context** gives the agent architectural memory — stack, patterns, conventions, domain knowledge
- **Specs** give the agent an implementation plan — features, acceptance criteria, task breakdowns
- **Skills** give the agent reusable abilities — how to commit, version, design entities, review code
- **Workflows** give the agent orchestration recipes — multi-step processes like feature development, bug fixing, releases

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

### Four ways to equip your agent

Every command supports **interactive mode** — run without flags and menus guide you through all options. Or pass flags explicitly for CI/scripting.

```bash
# 1. Set your API key (Claude or Gemini)
export ANTHROPIC_API_KEY="sk-ant-..."   # for Claude (default)
# or
export GEMINI_API_KEY="AI..."           # for Gemini

# ── Context: give your agent project memory ──
codify generate
# Interactive menus for: name, description, preset, language, model, locale, output, specs

# Or pass all flags explicitly (zero prompts):
codify generate payment-service \
  --description "Payment microservice in Go with gRPC, PostgreSQL and Kafka" \
  --language go

# ── Specs: give your agent an implementation plan ──
codify spec payment-service \
  --from-context ./output/payment-service/

# ── Skills: give your agent reusable abilities ──
codify skills
# Interactive menus for: category, preset, mode, target, install location
# No API key needed for static mode.

# ── Workflows: give your agent orchestration recipes ──
codify workflows
# Interactive menus for: preset, target, mode, locale, install location
# Supports Claude Code (native skills) and Antigravity (native .md) targets.
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

Scans an existing codebase and generates context files from what it finds. Uses a **differentiated prompt** that treats scan data as factual ground truth, producing more accurate output than a manual description.

**What the scanner detects:**
- Language, framework, and dependencies (Go, JS/TS, Python, Rust, Java, Ruby)
- Directory structure (3 levels deep)
- README content (filtered: badges, HTML comments, ToC removed)
- Existing context files (18+ patterns: AGENTS.md, .claude/CLAUDE.md, ADRs, OpenAPI specs, etc.)
- Build targets from Makefile/Taskfile (exact commands for AGENTS.md)
- Testing patterns (frameworks, BDD scenarios, coverage config)
- CI/CD pipelines (GitHub Actions triggers and jobs, GitLab CI)
- Infrastructure signals (Docker, Terraform, Kubernetes, Helm)

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
codify generate [project-name] [flags]
```

All flags are optional in a terminal — interactive menus prompt for missing values.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Project description *(required unless `--from-file`)* | *(interactive)* |
| `--from-file` | `-f` | Read description from file *(alternative to `-d`)* | — |
| `--preset` | `-p` | Template preset (`default`, `neutral`) | *(interactive)* |
| `--model` | `-m` | LLM model (`claude-*` or `gemini-*`) | auto-detected |
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
# → Select category (architecture, testing, conventions)
# → Select preset (clean, neutral, conventional-commit, ...)
# → Select mode (static or personalized)
# → Select target ecosystem (claude, codex, antigravity)
# → Select install location (global, project, or custom)
# → Select locale
# → If personalized: describe your project, choose model
```

### CLI mode

```bash
# Static: instant delivery, no API key
codify skills --category conventions --preset all --mode static

# Install globally — skills available from any project
codify skills --category conventions --preset all --mode static --install global

# Install to current project — shareable via git
codify skills --category architecture --preset clean --mode static --install project

# Personalized: LLM-adapted to your project
codify skills --category architecture --preset clean --mode personalized \
  --context "Go microservice with DDD, Godog BDD, PostgreSQL"

# Architecture skills for Codex ecosystem
codify skills --category architecture --preset neutral --target codex
```

### Install scopes

| Scope | Path (Claude) | Path (Codex) | Use case |
|-------|---------------|--------------|----------|
| `global` | `~/.claude/skills/` | `~/.codex/skills/` | Available from any project |
| `project` | `./.claude/skills/` | `./.agents/skills/` | Committed to git, shared with team |

### Skill catalog

| Category | Preset | Skills |
|----------|--------|--------|
| `architecture` | `clean` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `testing` | `foundational` | Test Desiderata — Kent Beck's 12 properties of good tests |
| `testing` | `tdd` | Test-Driven Development — Red-Green-Refactor *(includes foundational)* |
| `testing` | `bdd` | Behavior-Driven Development — Given/When/Then *(includes foundational)* |
| `conventions` | `conventional-commit` | Conventional Commits |
| `conventions` | `semantic-versioning` | Semantic Versioning |
| `conventions` | `all` | All convention skills combined |

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
| `--category` | Skill category (`architecture`, `testing`, `conventions`) | *(interactive)* |
| `--preset` | Preset within category | *(interactive)* |
| `--mode` | Generation mode: `static` or `personalized` | *(interactive)* |
| `--install` | Install scope: `global` (agent path) or `project` (current dir) | *(interactive)* |
| `--context` | Project description for personalized mode | — |
| `--target` | Target ecosystem (`claude`, `codex`, `antigravity`) | `claude` |
| `--model` `-m` | LLM model (personalized mode only) | auto-detected |
| `--locale` | Output language (`en`, `es`) | `en` |
| `--output` `-o` | Output directory (overrides `--install`) | ecosystem-specific |

---

## 🔄 Workflows

Workflows are multi-step orchestration recipes that AI agents execute on demand. Unlike skills (which teach *how* to do a specific task), workflows orchestrate *sequences of tasks* — from branch creation to PR merge, from bug report to fix deployment.

Codify generates workflows for two ecosystems:

| Target | Output format | Output path | Invocation |
|--------|--------------|-------------|------------|
| **Claude Code** | Native skill (`SKILL.md` with frontmatter) | `.claude/skills/{preset}/SKILL.md` | `/{preset}` (e.g., `/spec-propose`) |
| **Antigravity** | Native `.md` with execution annotations (`// turbo`, `// capture`, etc.) | `.agent/workflows/{workflow}.md` | `/workflow-name` |

Each Claude Code skill includes YAML frontmatter with:
- `name` — Skill name (kebab-case, used as `/slash-command`)
- `description` — What the workflow does
- `disable-model-invocation: true` — Only the user invokes it (workflows have side effects)
- `allowed-tools: Bash(*)` — Auto-approves shell commands for uninterrupted execution

### Two modes

| Mode | What it does | API key | Cost | Speed |
|------|-------------|---------|------|-------|
| **Static** | Delivers pre-built workflows from the embedded catalog. Ecosystem-aware frontmatter. | Not needed | Free | Instant |
| **Personalized** | LLM adapts workflows to your project — steps reference your tools, CI/CD, and deployment targets. | Required | ~pennies | ~10s |

### Interactive mode

```bash
codify workflows
# → Select preset (spec-driven-change, bug-fix, release-cycle, all)
# → Select target ecosystem (claude, antigravity)
# → Select mode (static or personalized)
# → Select locale
# → Select install location (global, project, or custom)
# → If personalized: describe your project, choose model
```

### CLI mode

```bash
# Claude Code: generate workflow skills
codify workflows --preset all --target claude --mode static

# Claude Code: install skills globally
codify workflows --preset all --target claude --mode static --install global

# Claude Code: generate spec-driven feature lifecycle (propose → apply → archive)
codify workflows --preset spec-driven-change --target claude --mode static

# Antigravity: generate native workflow files
codify workflows --preset all --target antigravity --mode static

# Antigravity: install globally
codify workflows --preset all --target antigravity --mode static --install global

# Personalized: LLM-adapted skills for your project
codify workflows --preset all --target claude --mode personalized \
  --context "Go microservice with CI/CD via GitHub Actions"
```

### Target ecosystems

| Target | Output | Structure | Key difference |
|--------|--------|-----------|----------------|
| `claude` | Native skill | `{preset}/SKILL.md` with YAML frontmatter | Annotations stripped, tool auto-approval via `allowed-tools` |
| `antigravity` *(default)* | Flat `.md` file | `{workflow}.md` with YAML frontmatter | Native annotations: `// turbo`, `// capture`, `// if`, `// parallel` |

### Install scopes

| Scope | Claude path | Antigravity path |
|-------|-------------|------------------|
| `global` | `~/.claude/skills/` | `~/.gemini/antigravity/global_workflows/` |
| `project` | `.claude/skills/` | `.agent/workflows/` |

### Workflow catalog

| Preset | Workflow | Description |
|--------|----------|-------------|
| `spec-driven-change` | Spec-driven Change | Propose → apply → archive — full SDD lifecycle with formal spec deltas, branch creation, and merge cleanup |
| `bug-fix` | Bug Fix | Reproduce → diagnose → fix → test → PR |
| `release-cycle` | Release Cycle | Version bump → changelog → tag → deploy |
| `all` | All workflows | All workflow presets combined |

### Spec-driven Change: the philosophy

`spec-driven-change` is the recommended workflow for adding features and making non-trivial changes. It implements **Spec-Driven Development (SDD)**: a methodology where formal planning artifacts precede code, and where every change to the system is a tracked, reviewable evolution of specifications — not just a code diff.

**The problem with chat-driven AI development:**
- Plans disappear when the chat session ends
- Code reviews see *what* changed but not *why* it changed
- AI agents lose context between sessions and re-litigate decisions
- Specs (when they exist) get out of sync with the code

**The SDD answer:**
- **Specs live in the repository**, organized by capability under `openspec/specs/<capability>/spec.md`
- **Each change is a self-contained workspace** under `openspec/changes/<change-id>/`
- **Deltas (ADDED / MODIFIED / REMOVED requirements)** describe how specs evolve, not just final state
- **Reviewers approve intent first** (proposal + deltas) before approving code
- **Archived changes preserve audit trail** indefinitely

#### The three phases

Each phase is a separate cognitive mode with a clear hand-off:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  /spec-      │     │  /spec-      │     │  /spec-      │
│  propose     │ ──▶ │  apply       │ ──▶ │  archive     │
│              │     │              │     │              │
│  Plan the    │     │  Execute the │     │  Consolidate │
│  change      │     │  plan        │     │  & cleanup   │
└──────────────┘     └──────────────┘     └──────────────┘
   Intent              Implementation        Truth
```

| Phase | What it produces | Cognitive mode |
|-------|------------------|----------------|
| **Propose** | `proposal.md` (motivation), `design.md` (technical decisions), `tasks.md` (atomic checklist), `specs/<capability>/spec.md` (deltas with ADDED/MODIFIED/REMOVED) — plus a feature branch with the proposal committed | "What should change and why" — no code yet |
| **Apply** | Sequential task execution, atomic commits per task, tests, self-review, pull request | "How to make it real" — focused on implementation, deltas already approved |
| **Archive** | Spec deltas merged into `openspec/specs/<capability>/spec.md`, change moved to `openspec/changes/archive/YYYY-MM-DD-<id>/`, feature branch merged and deleted | "Make the truth durable" — close the loop |

#### Concrete example

```
$ /spec-propose Add two-factor authentication via TOTP

  ✓ Read openspec/specs/auth-login/spec.md
  ✓ Created change-id: add-2fa
  ✓ Created openspec/changes/add-2fa/
      ├── proposal.md       (motivation, scope, impact)
      ├── design.md         (TOTP library choice, schema changes)
      ├── tasks.md          (8 atomic tasks across 3 phases)
      └── specs/auth-login/spec.md  (ADDED: 2FA requirements with G/W/T scenarios)
  ✓ Created branch feature/add-2fa
  ✓ Committed proposal artifacts
  → Request intent review before implementation

$ /spec-apply add-2fa

  ✓ Implementing task 1.1: add 2FA columns to user table
  ✓ Test: migration up/down
  ✓ Commit: "feat: add 2FA schema columns"
  ... (8 tasks, atomic commits)
  ✓ Full test suite passes
  ✓ PR opened: "add-2fa: Add two-factor authentication via TOTP"

$ /spec-archive add-2fa

  ✓ Merged deltas into openspec/specs/auth-login/spec.md
  ✓ Moved to openspec/changes/archive/2026-04-27-add-2fa/
  ✓ Squash-merged feature branch
  ✓ Deleted local + remote feature/add-2fa
```

#### How it fits with the rest of Codify

```
codify generate ─────▶ AGENTS.md, CONTEXT.md       (project memory)
codify spec ─────────▶ CONSTITUTION.md, SPEC.md... (initial specs)
codify workflows ────▶ /spec-propose, /spec-apply, /spec-archive
  --preset spec-                                   (SDD lifecycle skills)
  driven-change
```

`generate` and `spec` create the **initial state**. `spec-driven-change` workflow then governs **every subsequent change**, keeping the system's specs in lockstep with its code.

#### OpenSpec compatibility

The output structure (`openspec/specs/`, `openspec/changes/`, delta format with ADDED/MODIFIED/REMOVED, GIVEN/WHEN/THEN scenarios) follows the [OpenSpec](https://openspec.dev/) convention. Skills generated by Codify are designed to operate on OpenSpec workspaces seamlessly.

**Codify's value-add over installing OpenSpec directly:**
- **LLM personalization**: `--mode personalized --context "..."` adapts the skills to your stack, tools, and conventions
- **Multi-target**: same SDD methodology delivered for Claude Code or Antigravity
- **Locale support**: English and Spanish skills out of the box
- **Integrated pipeline**: combined with `codify generate` + `codify spec`, you get end-to-end SDD bootstrap

### Skills vs Workflows

| | Skills | Workflows |
|-|--------|-----------|
| **Purpose** | Teach *how* to do a specific task | Orchestrate a *sequence* of tasks |
| **Scope** | Single concern (e.g., "write a commit") | End-to-end process (e.g., "evolve a spec from proposal to merged change") |
| **Invocation** | Agent reads when relevant | User invokes via `/command` |
| **Examples** | Conventional Commits, DDD entity, code review | Spec-driven change lifecycle, bug fix, release cycle |

### Options

```bash
codify workflows [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--preset` `-p` | Workflow preset | *(interactive)* |
| `--target` | Target ecosystem: `claude` or `antigravity` | `antigravity` |
| `--mode` | Generation mode: `static` or `personalized` | *(interactive)* |
| `--install` | Install scope: `global` or `project` | *(interactive)* |
| `--context` | Project description for personalized mode | — |
| `--model` `-m` | LLM model (personalized mode only) | auto-detected |
| `--locale` | Output language (`en`, `es`) | `en` |
| `--output` `-o` | Output directory (overrides `--install`) | target-specific |

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
| `generate_workflows` | Generate workflow files for Claude Code (native skills) or Antigravity (native .md) — supports `static` and `personalized` modes |

All generative tools support `locale` (`en`/`es`) and `model` parameters. `generate_context` and `analyze_project` also accept `with_specs`. `generate_skills` accepts `mode`, `category`, `preset`, `target`, and `project_context`. `generate_workflows` accepts `mode`, `preset`, `target` (`claude`/`antigravity`), and `project_context`.

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

"Generate convention skills for my project"
→ Agent calls generate_skills with mode=static, category=conventions, preset=all

"Create DDD skills adapted to my Go project with Clean Architecture"
→ Agent calls generate_skills with mode=personalized, project_context="Go with DDD..."

"Generate spec-driven-change workflow for Claude Code"
→ Agent calls generate_workflows with target=claude, preset=spec-driven-change, mode=static

"Generate all workflows adapted to my Go project with GitHub Actions"
→ Agent calls generate_workflows with target=claude, mode=personalized, preset=all, project_context="Go with GitHub Actions"

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
│   ├── catalog/         Declarative skill + workflow catalogs and metadata registries
│   ├── shared/          Value objects, domain errors
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Use cases (CQRS)
│   ├── command/         GenerateContext, GenerateSpec, GenerateSkills, GenerateWorkflows
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementations
│   ├── llm/             LLM providers (Claude, Gemini) + prompt builder
│   ├── template/        Template loader (locale + preset + language-aware)
│   ├── scanner/         Project scanner (language, deps, framework, build targets, testing, CI/CD)
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 Entry points
    ├── cli/commands/    generate, analyze, spec, skills, workflows, serve, list
    └── mcp/             MCP server (stdio + HTTP transport, 7 tools)
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
│   │   ├── testing/             Testing: Foundational, TDD, BDD
│   │   └── conventions/         Conventions (conventional commits, semver)
│   ├── workflows/              Workflow templates
│   │   ├── bug_fix.template
│   │   ├── release_cycle.template
│   │   ├── spec_propose.template
│   │   ├── spec_apply.template
│   │   └── spec_archive.template
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

**v1.18.0** 🎉

✅ **Working:**
- Multi-provider LLM support (Anthropic Claude + Google Gemini)
- **Context generation** with streaming (`generate`, `analyze`)
- **Enhanced analyze** — differentiated prompt (factual vs aspirational), enriched scanner with 18+ context files, build target parsing, testing pattern detection, CI/CD pipeline summarization, smart README filtering
- **SDD spec generation** from existing context (`spec`, `--with-specs`)
- **Agent Skills** with dual mode (static/personalized), interactive guided selection, and declarative catalog
- **Skills install** — `--install global` or `--install project` for direct agent path installation
- Skill categories (architecture, testing, conventions) with ecosystem-aware frontmatter (Claude, Codex, Antigravity)
- **Workflows** — multi-step orchestration recipes for Claude Code (native skills) and Antigravity (native annotations)
- **Workflow presets** — spec-driven-change (propose/apply/archive), bug-fix, release-cycle (static + personalized modes, multi-target)
- **Unified interactive UX** — all commands prompt for missing parameters when run in a terminal
- MCP Server mode (stdio + HTTP transport) with 7 tools
- MCP knowledge tools (commit_guidance, version_guidance) — no API key needed
- Preset system (default: DDD/Clean, neutral: generic)
- AGENTS.md standard as root file
- Language-specific idiomatic guides (Go, JavaScript, Python)
- Dependency parsing for 8 languages (Go, JS/TS, Python, Rust, Java, Ruby, PHP, C#)
- Anti-hallucination grounding rules in prompts
- CLI with Cobra + interactive menus (charmbracelet/huh)
- Homebrew formula distribution (macOS/Linux)

🚧 **Coming next:**
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

**What's the difference between Skills and Workflows?**
Skills teach your agent *how* to do a single task (e.g., write a commit message, design a DDD entity). Workflows orchestrate a *sequence* of tasks into an end-to-end process (e.g., the full feature development lifecycle from branch to PR merge). Skills are passive (read when relevant), workflows are active (invoked via `/command`).

**Do I need an API key for workflows?**
Only for personalized mode. Static mode delivers pre-built workflows instantly — no LLM, no API key, no cost.

**Which ecosystems support workflows?**
Claude Code (`--target claude`) and Antigravity (`--target antigravity`). Claude workflows generate native skills (`SKILL.md` with frontmatter) following the official Claude Code skills methodology. Antigravity workflows produce native `.md` files with execution annotations (`// turbo`, `// capture`, etc.).

**What's AI Spec-Driven Development?**
A methodology where you generate context and specifications *before* writing code. Your agent implements a spec, not an improvisation. `generate` creates the blueprint, `spec` creates the implementation plan, and the `spec-driven-change` workflow governs every subsequent change as a tracked spec evolution (propose → apply → archive) with formal deltas, isolated change workspaces, and audit trails.

**Why three phases (propose / apply / archive) instead of one workflow?**
Each phase is a different cognitive mode. *Propose* answers "what should change and why?" without writing code — the LLM stays focused on intent. *Apply* answers "how to make it real?" with the deltas already approved, eliminating spec ambiguity from the implementation context. *Archive* closes the loop deterministically: merge deltas into source-of-truth specs, archive the change for audit, merge the branch. Mixing these phases dilutes attention and produces vague plans + sloppy code.

**Does Codify replace OpenSpec?**
No — it complements it. The `spec-driven-change` preset generates skills that operate on OpenSpec-format workspaces (`openspec/specs/`, `openspec/changes/`, ADDED/MODIFIED/REMOVED deltas with G/W/T scenarios). If you already use OpenSpec, Codify gives you LLM-personalized lifecycle skills tailored to your stack. If you don't, Codify is your zero-config entry point to the methodology — combined with `codify generate` and `codify spec`, you get the full pipeline from blank repo to governed iteration.

## 📚 Documentation

- [📋 AGENTS.md](AGENTS.md) — Project context for AI agents
- [🏛️ Architecture](context/CONTEXT.md) — DDD/Clean Architecture details
- [📝 Changelog](CHANGELOG.md) — Change history
- [📐 Specs](specs/) — Technical specifications (SDD)

## 📄 License

Apache License 2.0 — see [LICENSE](LICENSE).

---

<div align="center">

**Context. Specs. Skills. Workflows. Your agent, fully equipped.** 🧠

*"An agent without context is an intern with root access"*

⭐ If this helped you, give it a star — it keeps us building

[🐛 Report bug](https://github.com/jorelcb/codify/issues) · [💡 Request feature](https://github.com/jorelcb/codify/issues)

</div>