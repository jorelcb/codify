# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-2.0.3-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.23)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Generate, audit, and evolve your AI agent's context across the whole project lifecycle.** 🏗️

*Because an agent without context is an intern with root access — and stale context is an intern reading three-week-old docs.*

**[English]** | [Español](README_ES.md)

[Quick Start](#-quick-start) · [Config & Bootstrap](#%EF%B8%8F-configuration--bootstrap) · [Context](#-context-generation) · [Specs](#-spec-driven-development) · [Skills](#-agent-skills) · [Workflows](#-workflows) · [Hooks](#-hooks) · [Drift Detection](#-lifecycle-drift-detection) · [Update / Audit / Usage](#-lifecycle-update-audit--usage-tracking) · [Watch](#%EF%B8%8F-lifecycle-foreground-watcher-codify-watch) · [MCP Server](#-mcp-server) · [Language Guides](#-language-specific-guides) · [Architecture](#%EF%B8%8F-architecture) · [Migrating from v1.x](#-migrating-from-v1x)

</div>

---

## 🎯 The Problem

**Two problems, both real.**

**The agent improvises.** You tell it *"Build a payments API in Go with microservices"*, and:
- Folder structures nobody asked for
- Patterns that contradict your architecture
- Decisions you'll revert in the next session
- Zero continuity between sessions

It's not the agent's fault. Without context, it starts from scratch every session.

**Even when context exists, it drifts.** Three weeks ago you wrote a beautiful AGENTS.md. Since then `go.mod` added five deps, the Makefile gained four targets, the README evolved. The AGENTS.md still says what was true on day one. The agent reads it and confidently makes decisions on stale ground.

**Codify equips the agent with context AND keeps that context honest as the codebase moves.** 🛠️

## 💡 The Solution

**Codify** equips your AI agent with six layers it needs to stop improvising:

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

┌─────────────────────────────────┐  ┌─────────────────────────────────────┐
│           Hooks                 │  │            Lifecycle                │
│                                 │  │                                     │
│   Deterministic guardrails      │  │   Maintain artifacts over time      │
│   on tool calls (Edit/Bash)     │  │                                     │
│                                 │  │   config / init                     │
│   hooks                         │  │   check / update / audit / usage    │
└─────────────────────────────────┘  └─────────────────────────────────────┘
       Determinism                              Custodianship
```

- **Context** gives the agent architectural memory — stack, patterns, conventions, domain knowledge
- **Specs** give the agent an implementation plan — features, acceptance criteria, task breakdowns
- **Skills** give the agent reusable abilities — how to commit, version, design entities, review code
- **Workflows** give the agent orchestration recipes — multi-step processes like feature development, bug fixing, releases
- **Hooks** add deterministic guardrails — shell scripts on Claude Code lifecycle events, no LLM in the loop
- **Lifecycle** keeps everything in sync — `config`, `init`, `check`, `update`, `audit`, `usage`, `watch` — drift detection, selective regen, commit auditing, cost transparency, foreground watching

It follows the [AGENTS.md standard](https://github.com/anthropics/AGENTS.md) — an open specification backed by the Linux Foundation for providing AI agents with project context. Files work out of the box with Claude Code, Cursor, Codex, and any agent that reads the standard.

## ✨ Before and after

### 😱 Without Codify

```
Day 1
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

Day 22 (after AGENTS.md was written and never updated)
You: "Add the new circuit-breaker dep we added last week"
Agent: *uses go-retry — never heard of go-resilience because AGENTS.md is stale*

Result: 45 minutes correcting the agent on day 1, two hours on day 22. 😤
```

### 🚀 With Codify

```
Day 1
You: "Create a payments API in Go"

Agent: *reads AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md*
Agent: "I see you use DDD with Clean Architecture, PostgreSQL,
        BDD testing with Godog, and idiomatic Go patterns.
        I'll create the payments endpoint in internal/domain/payment/
        following your patterns and concurrency conventions."

Agent: *reads SKILL.md for conventional commits*
Agent: "Done. Here's the commit following Conventional Commits:
        feat(payment): add payment domain entity with Stripe integration"

Day 22 (you've been editing in the background; AGENTS.md auto-stayed in sync)
You: "Add the new circuit-breaker dep we added last week"
Agent: *reads current AGENTS.md — knows go-resilience is the project's choice*
Agent: "Adding go-resilience following the wrap-with-context pattern from
        IDIOMS.md. Note that codify check ran clean before this session."

Result: Coherent code from the first line, AND from line 1,000. ✨
```

Behind the scenes on day 22, **`codify watch`** has been quietly running, **`codify check`** flagged the `go.mod` change, **`codify update`** refreshed AGENTS.md, and **`codify usage`** shows it cost $0.04 in tokens.

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

### One-time setup (recommended)

The first time you run any interactive Codify command, you'll be offered the option to launch the configuration wizard:

```bash
codify generate
# → Codify isn't configured globally yet. Run interactive setup now? [Yes / No / Skip permanently]
# → Yes launches: codify config (wizard for default preset, locale, model, target)
```

You can also run `codify config` at any time. Configuration persists at `~/.codify/config.yml` and applies as defaults to every subsequent command (flags still override).

**Project bootstrap** with `codify init`:

```bash
cd my-project/
codify init
# → New or existing project?
#   - new      → asks for description (inline or file), runs `generate` internally
#   - existing → scans the codebase, runs `analyze` internally
# → Persists .codify/config.yml + .codify/state.json
```

`init` is the smart entry point that picks the right flow for you. If you prefer to control each step explicitly, use `generate`/`analyze` directly.

### Codify command surface

Every command supports **interactive mode** — run without flags and menus guide you through all options. Or pass flags explicitly for CI/scripting. Both forms read defaults from `~/.codify/config.yml` (user) and `.codify/config.yml` (project) when present, with merge precedence: flags > project > user > built-in defaults.

```bash
# 1. Set your API key (Claude or Gemini) — only needed for LLM-backed commands
export ANTHROPIC_API_KEY="sk-ant-..."   # for Claude (default)
# or
export GEMINI_API_KEY="AI..."           # for Gemini

# ── Bootstrap: configure once, equip a project end-to-end ──
codify config         # User-level wizard (auto-launches first time, opt-out via env / marker / flag)
codify init           # Project-level: new or existing → generate or analyze + state.json

# ── Context: give your agent project memory ──
codify generate            # Description-driven generation
codify analyze             # Scan existing repo and generate context from it

# ── Specs: give your agent an implementation plan ──
codify spec payment-service \
  --from-context ./output/payment-service/

# ── Skills: give your agent reusable abilities ──
codify skills              # No API key for static mode

# ── Workflows: give your agent orchestration recipes ──
codify workflows           # Claude (native skills) or Antigravity (native .md)

# ── Hooks: deterministic guardrails on Claude Code lifecycle events ──
codify hooks               # linting / security-guardrails / convention-enforcement / all

# ── Lifecycle: maintain artifacts over time ──
codify check               # Drift detection between snapshot and FS — no LLM, zero cost
codify update              # Selective regen when input signals drift
codify audit               # Review commits against conventions (rules-only by default; --with-llm opt-in)
codify reset-state         # Recompute snapshot without touching artifacts
codify usage               # Read LLM cost tracking from local files
```

**Free, no API key**: `config`, `init` (when generating from scan only), `check`, `reset-state`, `audit` (rules-only mode), `usage`, `hooks`, `skills` (static mode), `workflows` (static mode), MCP knowledge tools (`commit_guidance`, `version_guidance`, `get_usage`).

**Requires API key**: `generate`, `analyze`, `spec`, `skills --mode personalized`, `workflows --mode personalized`, `update`, `audit --with-llm`.

### Disabling the auto-launch prompt

The first-run prompt is **soft** — it only appears in interactive TTYs and never blocks CI or scripts. Three opt-out paths:

```bash
# Per-invocation: skip just for this run
codify generate --no-auto-config ...

# Per-shell: env variable
export CODIFY_NO_AUTO_CONFIG=1

# Permanently: marker file (created automatically when you choose "Skip permanently")
touch ~/.codify/.no-auto-config
```

### What you'll see

```
🚀 Generating context for: payment-service
  Model: claude-sonnet-4-6
  Preset: clean-ddd
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

## ⚙️ Configuration & Bootstrap

Two complementary commands shape how Codify behaves: **`codify config`** at the user level and **`codify init`** at the project level. Both compose on top of the existing standalone commands; they are smart entry points, not replacements.

### `codify config` — user-level defaults

`codify config` manages your global preferences at `~/.codify/config.yml`. The first time you run any interactive Codify command in a TTY without that file existing, you'll be offered the option to launch the wizard. Three answers: Yes (run wizard), No (use defaults this run), Skip permanently (creates `~/.codify/.no-auto-config` so the prompt never appears again).

| Subcommand | Action |
|---|---|
| `codify config` | Wizard if no config exists; print current config if it does |
| `codify config get <key>` | Read a single value |
| `codify config set <key> <value>` | Update a single value |
| `codify config unset <key>` | Clear a single value |
| `codify config edit` | Open `~/.codify/config.yml` in `$EDITOR` |
| `codify config list` | Print the effective config (with merge applied) |

Valid keys: `preset`, `locale`, `language`, `model`, `target`, `provider`, `project_name`.

### `codify init` — project-level bootstrap

`codify init` asks one question first: is this project new or existing? Based on the answer it routes you to the right flow:

| Answer | Internal flow | What you provide |
|---|---|---|
| **new** | invokes `generate` | project name + description (inline or path to a file) |
| **existing** | invokes `analyze` | project name (auto-detected from cwd, override if you want) |

After that, both branches collect: architectural preset (override of global default), language, locale, output directory, model. Result:

- `.codify/config.yml` — project-scoped defaults that persist for everyone with the repo
- `.codify/state.json` — snapshot of generation state (consumed by lifecycle commands)
- Generated `AGENTS.md` and `context/*.md` written to `output/`

Skills, workflows, and hooks are NOT bundled — `init` prints recommended next-step commands to keep responsibilities focused. Run `codify skills`, `codify workflows`, `codify hooks` separately when you want them.

### Merge precedence

When any command resolves a value (preset, locale, model, etc.):

```
flags > .codify/config.yml > ~/.codify/config.yml > built-in defaults
```

Setting `--preset hexagonal` on the command line wins regardless of what's in either config file. Project-level config wins over user-level. Built-ins fill the gaps.

---

## 🔍 Lifecycle: Drift Detection

Once Codify generates artifacts, the world keeps moving. Dependencies change, README evolves, someone hand-edits `AGENTS.md`. Without active checking, the artifacts drift silently out of sync with the project.

`codify check` and its companion `codify reset-state` solve this without an LLM: SHA256 hashes of artifacts and input signals, captured at generation time and compared at check time. **Zero LLM cost. Zero network. Fully deterministic.**

### `codify check` — detect drift in CI or locally

```bash
codify check                    # human-readable report; exit 1 on significant drift
codify check --strict           # any drift (including minor) triggers exit 1
codify check --json             # machine-readable JSON for CI pipelines
codify check -o ./output/my-project   # if artifacts live elsewhere than cwd
```

**What it detects:**

| Drift kind | Severity | What it means |
|---|---|---|
| `artifact_modified` | significant | A generated file (e.g. AGENTS.md) was edited after generation |
| `artifact_missing` | significant | A file present in the snapshot is gone from disk |
| `signal_changed` | significant | An input signal (`go.mod`, `Makefile`, `README.md`, etc.) changed — your context may be stale |
| `signal_removed` | significant | A tracked signal is no longer on disk |
| `artifact_new` | minor | A new artifact appeared since the snapshot |
| `signal_added` | minor | A new signal appeared (informational) |

**Exit codes:**

- `0` — no significant drift (or no drift at all)
- `1` — significant drift (default) or any drift (with `--strict`)
- `2` — no `.codify/state.json` exists (project not bootstrapped)

**CI usage example (GitHub Actions):**

```yaml
- name: Verify Codify artifacts are in sync
  run: codify check --strict
```

A non-zero exit fails the job, so PRs that change dependencies without regenerating context are caught automatically.

### `codify reset-state` — accept current FS as the new baseline

When you intentionally edit `AGENTS.md` (e.g. you tightened a constraint by hand) and want Codify to consider that the new truth:

```bash
codify reset-state              # recompute state.json from current FS, atomic write
codify reset-state --dry-run    # preview only, no changes
```

The command is read-only over your artifacts — it never modifies AGENTS.md or context files. It only updates `state.json` (with backup at `.bak`). Subsequent `check` runs compare against the new baseline.

### How drift detection works under the hood

Every successful `codify generate` / `codify analyze` / `codify init` writes `.codify/state.json` containing:

- Project metadata (name, preset, language, locale, target)
- Git context (commit, branch, remote, dirty status)
- Artifacts: SHA256 + size + generation timestamp for each generated file
- Input signals: SHA256 of well-known files (`go.mod`, `Makefile`, `README.md`, etc.)

`codify check` recomputes this snapshot from the current FS and diffs the two. The whole operation is local, fast (<100ms typical), and fully reproducible.

---

## 🔄 Lifecycle: Update, Audit & Usage Tracking

Three commands build on drift detection to close the gap between "Codify generated artifacts once" and "Codify maintains them as the project evolves": `update` regenerates selectively, `audit` reviews commits against documented conventions, `usage` exposes LLM cost.

### `codify update` — selective regeneration

Once `codify check` flags drift, `codify update` does the actual refresh:

```bash
codify update                    # detect drift, regenerate via analyze if needed
codify update --dry-run          # show what would change without LLM cost
codify update --force            # regenerate even on minor drift
codify update --accept-current   # keep current FS as new baseline (alias for reset-state)
codify update --no-tracking      # skip usage recording for this invocation
```

**Behavior matrix:**

| Drift state | Without `--force` | With `--force` |
|---|---|---|
| No drift | no-op, exit 0, no LLM call | no-op, exit 0 |
| Only minor drift (`artifact_new`, `signal_added`) | report and exit 0 | regenerate |
| Significant drift in signals (e.g. `go.mod` changed) | regenerate via analyze | regenerate via analyze |
| Only hand-edits to artifacts (no signal changes) | refuses with exit 1; suggests `--accept-current` | regenerate (loses edits) |

The "hand-edit refusal" exists deliberately — if you tightened a constraint in AGENTS.md by hand, regenerating would silently lose it.

### `codify audit` — review commits against conventions

`audit` evaluates recent git commits against project conventions:

```bash
codify audit                     # last 20 commits, rules-only (zero LLM cost)
codify audit --since main~50     # all commits since main~50
codify audit --strict            # any finding (incl. minor) fails the run
codify audit --json              # machine-readable for CI pipelines
codify audit --with-llm          # heuristic mode — sends commits + AGENTS.md to LLM (records usage)
```

**Rules-only checks (deterministic, zero cost):**

| Finding | Severity | Description |
|---|---|---|
| `commit_invalid_type` | significant | Header doesn't match `type[scope][!]: subject` or uses an unknown type |
| `commit_trivial` | significant | Message is a placeholder (`wip`, `fix`, `update`, etc.) |
| `commit_header_too_long` | minor | Header exceeds 72 characters |
| `protected_branch_direct` | significant | Direct commit on `main` / `master` / `develop` / `production` (no merge commit detected) |

Recognized commit types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.

### `codify usage` — LLM cost transparency

Every successful and failed LLM call (from `generate`, `analyze`, `update`, `spec`, `skills`, `workflows`, etc.) is automatically recorded with token counts and cost. Read the log with:

```bash
codify usage                       # current project's spending
codify usage --global              # aggregate across all your projects
codify usage --since 7d            # last 7 days only
codify usage --by command          # break down by command name
codify usage --by model            # break down by model name
codify usage --json                # full JSON for scripting
codify usage --reset               # archive current log and start fresh
```

**Sample output:**

```
Codify Usage — project scope (.codify/usage.json)
════════════════════════════════════════════════════════════
Total cost:     $0.42 (42 cents)
Total calls:    17
Total input:    142.3K tokens
Total output:   31.8K tokens

By command:
  generate                  $0.12   2 calls
  audit                     $0.18   8 calls
  update                    $0.10   6 calls
  spec                      $0.02   1 call
```

**Pricing transparency:** the cost is computed using a public list-price table embedded at `internal/domain/usage/pricing.go` (version `2026-05`). It reflects what Anthropic and Google publish on their pricing pages. If you have negotiated discounts, the figure shown is an upper bound — useful for comparison, not for invoicing.

**Disabling tracking — three options, any one suffices:**

```bash
# 1. Per-invocation: skip just for this run
codify update --no-tracking

# 2. Per-shell: env variable
export CODIFY_NO_USAGE_TRACKING=1

# 3. Permanently: marker file
touch ~/.codify/.no-usage-tracking
```

When tracking is disabled, no entries are written. The `codify usage` command will report zero (because nothing was recorded), but works fine.

### CI integration — GitHub Actions pattern

A drop-in workflow that runs `codify check` + `codify audit` on every pull request:

```yaml
# .github/workflows/codify.yml
name: Codify drift + audit
on: [pull_request]

jobs:
  codify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 50      # codify audit needs commit history

      - name: Install Codify
        run: |
          go install github.com/jorelcb/codify/cmd/codify@latest
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      - name: Verify generated artifacts are in sync
        run: codify check --strict

      - name: Audit recent commits
        run: codify audit --since origin/main --strict
```

Both `check` and `audit` (rules-only mode, the default) are deterministic and free — no API key required. `update` and `audit --with-llm` require `ANTHROPIC_API_KEY` or `GEMINI_API_KEY`.

---

## 👁️ Lifecycle: Foreground Watcher (`codify watch`)

`codify watch` keeps drift detection running in the background of your editor session. It re-runs `check` automatically when any file registered in `.codify/state.json` changes — input signals (e.g. `go.mod`, `Makefile`, `README.md`) and generated artifacts (`AGENTS.md`, `context/*.md`).

```bash
codify watch                         # default 2s debounce, report-only
codify watch --debounce 500ms        # tighter debounce for fast feedback
codify watch --auto-update --strict  # aggressively keep artifacts in sync
```

**Behavior:**
- Loads `.codify/state.json` once at startup; exits 2 if missing
- Subscribes via `fsnotify` to the parent dirs of registered paths (no recursive walk)
- Debounces events (default 2s) — five rapid saves trigger one drift check, not five
- Prints drift reports to stdout and keeps watching
- `--auto-update` runs `codify update` when significant drift is detected (records LLM usage)
- `Ctrl+C` exits cleanly

### Why foreground (not daemon)

`codify watch` is intentionally a **foreground process**, not a system daemon. It has no `--detach`, no PID file, no signal-driven reload. This decision is documented in [ADR-008](docs/adr/0008-watch-model-decision.md). The summary:

- **PID file management, signal handling, log rotation, OS service integration** are all hard problems and out of scope for a single-maintainer project. Users who need persistence can wrap with `tmux` / `nohup` / `systemd` / their preferred process supervisor.
- **The realistic use case is short-lived** — you start `watch` when you start coding, you stop it when you stop. Hours, not weeks.
- **Scope is naturally bounded** — only the ~20 paths in `state.json` are watched.

### Wrapping in a process supervisor

If you do want long-running watch:

```bash
# tmux session that survives terminal close
tmux new-session -d -s codify-watch "cd $(pwd) && codify watch"
tmux attach -t codify-watch         # to inspect; Ctrl+B then D to detach

# systemd user unit (~/.config/systemd/user/codify-watch.service)
[Unit]
Description=Codify watch for %i
[Service]
WorkingDirectory=%h/projects/%i
ExecStart=/usr/local/bin/codify watch --debounce 5s
Restart=on-failure
[Install]
WantedBy=default.target

# nohup for a quick session-survival
nohup codify watch > codify-watch.log 2>&1 &
```

### Alternative — git-hook integration with `codify check`

For users whose mental model is "validate at git commit" rather than "validate while editing", `codify check` is the right tool — it's a one-shot deterministic check designed for CI and git hooks. Integrate via your preferred hook manager:

**lefthook (`lefthook.yml`):**
```yaml
pre-commit:
  commands:
    codify-check:
      run: codify check --strict
```

**pre-commit (`.pre-commit-config.yaml`):**
```yaml
repos:
  - repo: local
    hooks:
      - id: codify-check
        name: Codify drift detection
        entry: codify check --strict
        language: system
        pass_filenames: false
```

**watchexec (foreground alternative on the same FS-event basis):**
```bash
watchexec -w go.mod -w Makefile -w README.md -- codify check
```

Codify itself doesn't generate these configs — the integration is short and project-specific enough that copy-paste is the right primitive (per [ADR-008](docs/adr/0008-watch-model-decision.md)).

---

## 📋 Context Generation

The foundation. Generates files following the [AGENTS.md](https://github.com/anthropics/AGENTS.md) standard that give your agent deep project memory.

### When to use `generate` vs `analyze`

| Situation | Use | Why |
|---|---|---|
| Greenfield project (no code yet) | `codify generate` | You provide the description; the LLM generates context against it |
| Existing repo with code in it | `codify analyze` | The scanner extracts factual signals (deps, build targets, CI, frameworks) and feeds them as ground truth — much higher accuracy than a manual description |
| Existing repo + you want to override what the scanner detects | `codify analyze` first, then edit, then `codify reset-state` | Scan-first, hand-tune second |
| You have a detailed design doc | `codify generate --from-file ./docs/design.md` | Treat the file's content as the description |
| In doubt | `codify init` | Asks "new or existing?" and routes you to the right flow internally |

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
| `--preset` | `-p` | Template preset (`neutral`, `clean-ddd`, `hexagonal`, `event-driven`) | *(interactive)* |
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
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `architecture` | `clean-ddd` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `hexagonal` | Port definition, Adapter pattern, Dependency inversion, Hexagonal integration test |
| `architecture` | `event-driven` | Command handler, Domain event, Event projection, Saga orchestrator, Event idempotency |
| `testing` | `foundational` | Test Desiderata — Kent Beck's 12 properties of good tests |
| `testing` | `tdd` | Test-Driven Development — Red-Green-Refactor *(includes foundational)* |
| `testing` | `bdd` | Behavior-Driven Development — Given/When/Then *(includes foundational)* |
| `conventions` | `conventional-commit` | Conventional Commits |
| `conventions` | `semantic-versioning` | Semantic Versioning |
| `conventions` | `all` | All convention skills combined |

The four `architecture` presets mirror the four `--preset` options for context generation, so skills installed for `hexagonal` line up with AGENTS.md/CONTEXT.md generated under `--preset hexagonal`.

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

#### Adopting SDD on an existing codebase

For brownfield projects (mature codebases without formal specs), the adoption path is different — specs should emerge from the **real** behavior of the code, not from aspirations. Follow this sequence:

```
1. codify analyze ./my-project           → AGENTS.md, CONTEXT.md, ... (factual context from scan)
2. openspec init                         → empty openspec/ workspace
3. codify workflows                      → /spec-propose, /spec-apply, /spec-archive
     --preset spec-driven-change
     --target claude --install project
4. From your agent, prompt:
   "Read AGENTS.md and CONTEXT.md, then reverse-engineer OpenSpec specs
    from the source code under a change named 'baseline'. Identify
    capability boundaries from the codebase structure. Use ADDED
    requirements with GIVEN/WHEN/THEN scenarios derived from real
    behavior, not aspirational design."
5. /spec-archive baseline                → consolidate baseline specs into openspec/specs/
```

This pattern (the [OpenSpec retrofitting mode](https://openspec.dev/)) produces **factual** specs validated against existing code rather than projections from a description. After the baseline is archived, every new change goes through the standard `/spec-propose → /spec-apply → /spec-archive` lifecycle. Codify's role here is to provide the context (`analyze`) and the lifecycle skills (`workflows --preset spec-driven-change`); the baseline retrofit itself is a one-shot prompt against your agent, not a separate Codify command — keeping responsibilities clean and avoiding overlap with OpenSpec's tooling.

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

## 🪝 Hooks

Hooks are **deterministic guardrails** for Claude Code. Where skills (prompts) and workflows (orchestration) rely on the LLM doing the right thing, hooks are shell scripts that **always** run on lifecycle events (`PreToolUse`, `PostToolUse`, etc.) — they enforce rules every single time, by exit code.

The three artifact layers complement each other:

| Layer | Mechanism | When does it run? | Determinism |
|---|---|---|---|
| **Skills** | Prompt loaded into context | When agent or user invokes | Depends on LLM |
| **Workflows** | Multi-skill lifecycle | User invokes via slash command | Depends on LLM |
| **Hooks** | Shell scripts on events | Every matching tool call | 100% (exit codes) |

### Preset catalog

| Preset | Event | Purpose |
|---|---|---|
| `linting` | `PostToolUse` (Edit\|Write) | Auto-format and lint files using the right tool per language (Prettier/ESLint, ruff/black, gofmt/gofumpt, rustfmt, rubocop, shfmt). Tools detected via `command -v` — skipped silently if not installed. |
| `security-guardrails` | `PreToolUse` (Bash, Edit\|Write) | Block dangerous Bash commands (`rm -rf /`, `git push --force` to main, `curl \| bash`, fork bombs, fs-formatting) and protect sensitive files (`.env*`, `secrets/`, `.git/`, lockfiles, private keys, CI configs). |
| `convention-enforcement` | `PreToolUse` (Bash with `if`) | Validate commit messages against Conventional Commits 1.0.0 (header ≤72 chars, valid type, no trivial placeholders) and block direct/force pushes to protected branches (`main`, `master`, `develop`, `production`, `release/*`). Requires Claude Code v2.1.85+. |
| `all` | (combined) | All three preset bundles merged into a single `hooks.json` |

### Activation modes

| Flag | Behavior |
|---|---|
| `--install project` (default in interactive) | Merge into `.claude/settings.json` and copy scripts to `.claude/hooks/`. Backs up the existing settings file before any modification. Idempotent — running it twice adds zero handlers the second time. |
| `--install global` | Same as project, but targets `~/.claude/settings.json` and `~/.claude/hooks/` for all projects |
| `--output PATH` | **Preview mode** — writes a standalone `{PATH}/hooks.json` + `{PATH}/hooks/*.sh` bundle for inspection or manual merge. Does NOT touch `settings.json`. Use this if you want to review the proposed changes before activating |
| `--dry-run` | Prints the proposed `settings.json` after merge, exits 0, writes nothing |

### Output layout

```
~/.claude/                      OR   ./.claude/
├── settings.json   (merged)         ├── settings.json   (merged)
├── settings.json.codify-backup-…    ├── settings.json.codify-backup-…
└── hooks/                            └── hooks/
    ├── lint.sh                            ├── lint.sh
    ├── block-dangerous-commands.sh        ├── block-dangerous-commands.sh
    ├── protect-sensitive-files.sh         ├── protect-sensitive-files.sh
    ├── validate-commit-message.sh         ├── validate-commit-message.sh
    └── check-protected-branches.sh        └── check-protected-branches.sh
```

### Interactive mode

```bash
codify hooks
# → Select preset (linting, security-guardrails, convention-enforcement, all)
# → Select locale (en, es)
# → Select activation mode (project / global / preview)
```

### CLI mode

```bash
# Activate everything for the current project (default flow)
codify hooks --preset all --install project

# Globally for all your projects
codify hooks --preset all --install global

# Preview only (write bundle, don't touch settings.json)
codify hooks --preset linting --output ./tmp/preview

# See the proposed merge without writing anything
codify hooks --preset all --install project --dry-run

# Spanish stderr messages
codify hooks --preset linting --install project --locale es
```

### Verify activation

```bash
claude
> /hooks
```

### Rollback

Every install backs up the previous `settings.json` to `settings.json.codify-backup-<timestamp>`. To roll back:

```bash
mv .claude/settings.json.codify-backup-<timestamp> .claude/settings.json
```

### Requirements

- **Bash** + **jq** (Linux/macOS native; Windows requires Git Bash or WSL)
- **Claude Code v2.1.85+** (only for the `convention-enforcement` preset, which uses the `if` field on hook handlers)

### Honest limitations

The bash scripts use regex patterns, not AST parsing. They stop **careless** agent commands, not motivated adversaries — sophisticated obfuscation (e.g. `eval $(echo b3JtIC1yZiAv | base64 -d)`) can bypass detection. For stronger guarantees use a dedicated tool like [bash-guardian](https://github.com/RoaringFerrum/claude-code-bash-guardian). The scripts are short and deliberately editable: extend the pattern arrays to match your project's specific risk model.

### Options

```bash
codify hooks [flags]
```

| Flag | Description | Default |
|---|---|---|
| `--preset` `-p` | `linting`, `security-guardrails`, `convention-enforcement`, or `all` | *(interactive)* |
| `--locale` | Output language for stderr (`en` or `es`) | `en` |
| `--install` | Install scope: `global` or `project` (auto-activates) | *(interactive — default `project`)* |
| `--output` `-o` | Preview directory: write standalone bundle, no settings change | — |
| `--dry-run` | Print the proposed `settings.json` merge but write nothing | `false` |

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
| `generate_hooks` | Generate Claude Code hook bundles (deterministic guardrails). Static-only, Claude-only. Outputs `hooks.json` + `.sh` scripts for manual merge into `settings.json` |

All generative tools support `locale` (`en`/`es`) and `model` parameters. `generate_context` and `analyze_project` also accept `with_specs`. `generate_skills` accepts `mode`, `category`, `preset`, `target`, and `project_context`. `generate_workflows` accepts `mode`, `preset`, `target` (`claude`/`antigravity`), and `project_context`. `generate_hooks` accepts `preset` (`linting`/`security-guardrails`/`convention-enforcement`/`all`), `locale`, and `output` — no model or context (static-only).

#### Read-only tools (no API key needed)

| Tool | Description |
|------|-------------|
| `commit_guidance` | Conventional Commits spec and behavioral context for generating proper commit messages |
| `version_guidance` | Semantic Versioning spec and behavioral context for determining version bumps |
| `get_usage` | Read LLM cost tracking from local `.codify/usage.json` (project) or `~/.codify/usage.json` (global). Pure file read, no LLM call. Parameters: `scope` (`project`/`global`), `since` (e.g. `7d`/`24h`), `by` (`command`/`model`/`provider`) |

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

"Generate Claude Code hooks to block dangerous commands and enforce conventional commits"
→ Agent calls generate_hooks with preset=all (or security-guardrails + convention-enforcement)

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

Choose the architectural philosophy for your context. Codify ships **4 presets**:

| Preset | Focus | When to use |
|---|---|---|
| `neutral` *(default)* | No architectural opinion — structure adapts to the project | Greenfield exploration, scripts, tools, anywhere you want minimal opinion baked in |
| `clean-ddd` | DDD + Clean Architecture + BDD + Layered domain | Long-lived business systems, domain-rich logic, teams comfortable with layered architecture |
| `hexagonal` | Ports & Adapters — lighter than clean-ddd | Apps with strong external integration concerns, swappable infra, simpler than full DDD |
| `event-driven` | CQRS + Event Sourcing + Sagas | Async systems, multi-service coordination, event-first domains, audit trails |

```bash
# Default — no architectural opinion
codify generate my-api -d "Inventory REST API in Go"

# Clean + DDD
codify generate my-api -d "Inventory REST API in Go" --preset clean-ddd

# Hexagonal — ports & adapters
codify generate my-payments -d "Payment service" --preset hexagonal

# Event-driven — CQRS + ES + sagas
codify generate my-orders -d "Order processing" --preset event-driven
```

### `--from-file` — Rich descriptions from files

For detailed project descriptions (design docs, RFCs, 6-pagers), use `--from-file` instead of `--description`:

```bash
codify generate my-api \
  --from-file ./docs/project-description.md \
  --language go
```

The file content becomes the project description. Supports any text format — markdown, plain text, etc. Mutually exclusive with `--description`.

## 🚀 Migrating from v1.x

Codify v2.0 has **one breaking change**. Everything else (multi-target support for Claude/Codex/Antigravity, all commands, all flags, all config keys) keeps working identically.

### What changed

| v1.x | v2.0 |
|---|---|
| `--preset default` (deprecated alias resolving to `clean-ddd` with warning) | **Removed** — returns a clear error with migration instructions |
| Default value of `--preset` flag: `clean-ddd` | **`neutral`** (no architectural opinion baked in) |
| `default` accepted in `~/.codify/config.yml` | Returns the same error at config load |

The change in default reflects a project decision documented in [ADR-001](docs/adr/0001-default-preset-transition.md): Codify's "default" used to be DDD/Clean — opinionated. v2.0 makes the default architecturally neutral, so the agent gets a clean baseline unless you explicitly choose a stance.

### Migration steps

**If you used `--preset default` explicitly:**

```bash
# Before (v1.x):
codify generate my-api -d "..." --preset default

# After (v2.0): use clean-ddd (same behavior as v1.x default)
codify generate my-api -d "..." --preset clean-ddd

# OR adopt the new default explicitly:
codify generate my-api -d "..." --preset neutral
```

**If you ran `codify generate` without `--preset` and want to keep v1.x behavior:**

Two options:

```bash
# Option A — pass --preset clean-ddd on every invocation
codify generate my-api -d "..." --preset clean-ddd

# Option B — set it as your global default (recommended for CI/scripts)
codify config set preset clean-ddd
```

**If your `~/.codify/config.yml` has `preset: default`:**

```bash
# Edit the file or use the CLI:
codify config set preset clean-ddd   # to keep v1.x behavior
codify config set preset neutral     # to adopt v2.0 default
```

### What did NOT change

- All targets remain supported: `claude`, `codex`, `antigravity` (per [ADR-009](docs/adr/0009-antigravity-deprecation-reversal.md), reversing the v1.26 deprecation plan)
- All commands work identically — `generate`, `analyze`, `spec`, `skills`, `workflows`, `hooks`, `config`, `init`, `check`, `update`, `audit`, `usage`, `watch`, `reset-state`
- All other flags, all output formats, all MCP tools (10 total)
- Config schema, state.json schema, usage.json schema — unchanged
- Pricing table version, locale options, language options — unchanged

If you don't pass `--preset` explicitly anywhere, the only observable difference is that newly-generated AGENTS.md/CONTEXT.md will be architecture-agnostic instead of DDD-flavored. Existing artifacts are not affected; `codify check` won't flag drift just because the version changed.

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
    └── mcp/             MCP server (stdio + HTTP transport, 10 tools)
```

### Template system

```
templates/
├── en/                          English locale
│   ├── neutral/                 Default preset — no architectural opinion
│   │   ├── agents.template
│   │   ├── context.template
│   │   ├── interactions.template
│   │   └── development_guide.template
│   ├── clean-ddd/               DDD + Clean Architecture + BDD
│   │   └── (same files)
│   ├── hexagonal/               Ports & Adapters
│   │   └── (same files)
│   ├── event-driven/            CQRS + Event Sourcing + Sagas
│   │   └── (same files)
│   ├── spec/                    Specification templates (AI SDD)
│   │   ├── constitution.template
│   │   ├── spec.template
│   │   ├── plan.template
│   │   └── tasks.template
│   ├── skills/                  Agent Skills templates (static + LLM guides)
│   │   ├── neutral/             Architecture: review, testing, API design, refactoring
│   │   ├── clean-ddd/           Architecture: DDD entity, layer, BDD, CQRS, Hexagonal port
│   │   ├── hexagonal/           Architecture: port, adapter, dependency inversion, integration test
│   │   ├── event-driven/        Architecture: command handler, domain event, projection, saga, idempotency
│   │   ├── testing/             Testing: Foundational, TDD, BDD
│   │   └── conventions/         Conventions (conventional commits, semver)
│   ├── workflows/              Workflow templates
│   │   ├── bug_fix.template
│   │   ├── release_cycle.template
│   │   ├── spec_propose.template
│   │   ├── spec_apply.template
│   │   └── spec_archive.template
│   ├── hooks/                  Hook bundle templates
│   │   ├── linting/
│   │   ├── security-guardrails/
│   │   └── convention-enforcement/
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

**v2.0.3**

The full surface in one snapshot — anything checked here is shipped, tested, and behaves as documented above.

**Context layer**
- ✅ `generate` — context from a description (4 files, +1 with `--language`)
- ✅ `analyze` — context from an existing repo via project scanner (18+ context-file patterns, build-target parsing, CI/CD detection, framework + dependency parsing for 8 languages)
- ✅ `spec` + `--with-specs` flag — SDD specs (CONSTITUTION, SPEC, PLAN, TASKS)
- ✅ Streaming output, anti-hallucination grounding rules, output validators (`[DEFINE]` markers, frontmatter, code-fence balance)
- ✅ Anthropic prompt caching across per-file generation loop

**Behavior layer**
- ✅ `skills` — 4 architecture presets (mirroring context presets) + testing + conventions; static + personalized modes; multi-ecosystem (claude, codex, antigravity)
- ✅ `workflows` — spec-driven-change, bug-fix, release-cycle; static + personalized; claude (native skills) + antigravity (native annotations)
- ✅ `hooks` — linting, security-guardrails, convention-enforcement; auto-install with backup + idempotent merge; `--output` preview and `--dry-run`

**Bootstrap layer**
- ✅ `config` — user-level config wizard with auto-launch SOFT (TTY-gated, triple opt-out); `get` / `set` / `unset` / `edit` / `list` subcommands
- ✅ `init` — project-level smart router (new vs existing) that delegates to `generate` or `analyze`

**Lifecycle layer**
- ✅ `check` — drift detection (artifact_modified, signal_changed, etc.) — deterministic, no LLM
- ✅ `update` — selective regeneration via `analyze`; refuses to overwrite hand-edits without `--force`
- ✅ `audit` — Conventional Commits + protected branches (rules-only, free) + `--with-llm` heuristic mode (records usage)
- ✅ `usage` — local LLM cost tracking (`.codify/usage.json` + `~/.codify/usage.json`); `--global`, `--since`, `--by`, `--json`, `--reset`
- ✅ `watch` — foreground file watcher with debounce, optional `--auto-update`
- ✅ `reset-state` — recompute snapshot without touching artifacts

**MCP server**
- ✅ 10 tools: 7 generative (context/specs/analyze/skills/workflows/hooks/usage) + 3 read-only (commit_guidance/version_guidance/get_usage)
- ✅ stdio + HTTP transports; parameter enums for stricter agent validation; no API key needed for read-only tools

**Distribution**
- ✅ Homebrew tap (`brew install jorelcb/tap/codify`)
- ✅ `go install github.com/jorelcb/codify/cmd/codify@latest`
- ✅ Pre-built binaries in GitHub Releases

**Quality**
- ✅ 9 BDD packages with 30+ scenarios; pure unit tests across domain + infrastructure
- ✅ DDD/Clean Architecture internal layout (the project eats its own dog food)

**Known boundaries (intentional, not roadmap):**
- No daemon mode for `watch` — wrap with tmux/nohup/systemd if needed (per [ADR-008](docs/adr/0008-watch-model-decision.md))
- No `pkg/codify` Go library — embedding via process boundary (CLI/MCP) is the contract (per [ADR-003](docs/adr/0003-no-public-go-library.md))
- Hooks are Claude Code-only (the underlying primitive doesn't exist for codex/antigravity)

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

## 🆘 Troubleshooting

Quick reference for the errors most people hit on first contact.

| Error / Symptom | Cause | Fix |
|---|---|---|
| `ANTHROPIC_API_KEY or GEMINI_API_KEY environment variable is required` | LLM-backed command without an API key in the env | `export ANTHROPIC_API_KEY=...` (or Gemini); for read-only commands like `check`, `audit --rules-only`, `usage`, none is needed |
| `preset 'default' was removed in Codify v2.0.0...` | Carried `--preset default` from a v1.x script or `~/.codify/config.yml` | `codify config set preset clean-ddd` (v1.x behavior) or `... preset neutral` (v2.0 default). Or pass `--preset clean-ddd` explicitly |
| `No snapshot at .codify/state.json...` (exit 2) on `check` / `update` / `watch` | Project not bootstrapped — never ran `init` / `generate` / `analyze` | Run one of those first, or `codify reset-state` if `state.json` was deleted by accident |
| `codify update` refuses with "Only hand-edits to generated artifacts detected" | You edited AGENTS.md by hand and `update` doesn't want to overwrite intent | `codify update --accept-current` (= `reset-state`) to make your edits the new baseline, or `--force` to regenerate (loses edits) |
| `codify watch` exits with "no watchable directories" | `state.json` exists but its registered paths are all missing | `codify reset-state` to recompute against the current FS |
| `Codify isn't configured globally yet. Run interactive setup now?` prompt blocks scripts | Auto-launch SOFT prompt fires in a TTY without `~/.codify/config.yml` | Pass `--no-auto-config`, or `export CODIFY_NO_AUTO_CONFIG=1`, or `touch ~/.codify/.no-auto-config` |
| `codify hooks` works but Claude Code doesn't run them | `.claude/settings.json` not loaded by your Claude Code version | Check Claude Code is v2.1.85+ (required for `convention-enforcement`); verify with `claude /hooks` |
| `audit --with-llm` falls back to rules-only with WARNING | Missing API key OR LLM call failed | Same fix as the API-key error; rules-only still produced its findings |
| Hooks scripts skip silently (e.g. `lint.sh` does nothing) | Required tool (gofmt, ruff, prettier, etc.) not installed | `command -v <tool>` to verify; install whichever you want enforced |

If you hit something that's not in this table, open an issue with: command run, exit code, and stderr. The CHANGELOG and ADRs in this repo document most design decisions — usually the answer is in there.

## 📚 Documentation

- [📋 AGENTS.md](AGENTS.md) — Project context for AI agents
- [🏛️ Architecture](context/CONTEXT.md) — DDD/Clean Architecture details
- [📝 Changelog](CHANGELOG.md) — Change history
- [📐 Specs](specs/) — Technical specifications (SDD)

## 📄 License

Apache License 2.0 — see [LICENSE](LICENSE).

---

<div align="center">

**Context. Specs. Skills. Workflows. Hooks. Lifecycle. Your agent, fully equipped — and kept honest.** 🧠

*"An agent without context is an intern with root access — and stale context is an intern reading three-week-old docs"*

⭐ If this helped you, give it a star — it keeps us building

[🐛 Report bug](https://github.com/jorelcb/codify/issues) · [💡 Request feature](https://github.com/jorelcb/codify/issues)

</div>