# Getting Started with Codify

A 5-minute tour from zero to a fully bootstrapped project. Each step maps to a lifecycle phase: **Bootstrap → Equip → Maintain**.

For the conceptual overview of the phases, see [`docs/lifecycle-matrix.md`](lifecycle-matrix.md). For the command-by-command reference see [`docs/command-reference.md`](command-reference.md).

---

## 1. Install

```bash
# Homebrew (macOS/Linux — no Go required)
brew tap jorelcb/tap
brew install codify

# Or via go install
go install github.com/jorelcb/codify/cmd/codify@latest

# Or download pre-built binaries from GitHub Releases
# https://github.com/jorelcb/codify/releases
```

Verify the install:

```bash
codify --version
codify --help        # phase diagram + command index
```

---

## 2. Bootstrap your workstation (`codify config`)

A one-time setup of your laptop. Picks the default target ecosystem (Claude / Codex / Antigravity), preset, locale, and model. Persists to `~/.codify/config.yml` and applies as defaults to every later command.

```bash
codify config
```

What you'll see:

```
Codify · Bootstrap (workstation)
════════════════════════════════
? Default target ecosystem
> Claude Code (recommended — full support: skills, workflows, hooks)
  Codex (skills only)
  Antigravity (skills + workflows)

? Default model: claude-sonnet-4-6
? Default locale: en
? Default architectural preset: clean-ddd

✓ Saved /Users/<you>/.codify/config.yml

? Install global skills now? [y/N] N   (default: skip — you can do it later)
? Install global workflows now? [y/N] N
? Install global hooks now? [y/N] N

✓ Workstation defaults saved.

Next steps
──────────

Bootstrap (per project):
  codify init       Bootstrap a project (new or existing) using these defaults

Update workstation defaults later:
  codify config     Re-run this wizard
  codify config set <key> <value>
```

> **Soft auto-launch:** the first time you run any interactive Codify command without a global config, this wizard is *offered* automatically (never forced). Three opt-out paths: `--no-auto-config` flag, `CODIFY_NO_AUTO_CONFIG=1` env, or `~/.codify/.no-auto-config` marker file.

---

## 3. Bootstrap your project (`codify init`)

`init` is the smart entry point per project. It asks `new` or `existing` and routes accordingly.

### Greenfield path

```bash
mkdir my-payments-api && cd my-payments-api
codify init
```

What you'll see:

```
Codify · Bootstrap (project)
════════════════════════════
? Is this a new project or an existing one?
> new — describe the project, generate context
  existing — scan the codebase, generate context from what's there

? Project name: payments-api
? How do you want to provide the description?
> inline (prompt now)
  file (path to a file with the description)

? Project description:
> Payments API in Go with microservices, Stripe integration, DDD/Clean Architecture, BDD tests with Godog.

? Architectural preset is 'clean-ddd' (from global default). Override? [y/N] N
? Language: go
? Output directory: .

--- Bootstrapping context ---

🚀 Generating context for: payments-api
  [1/5] Generating AGENTS.md... ✓
  [2/5] Generating CONTEXT.md... ✓
  [3/5] Generating INTERACTIONS_LOG.md... ✓
  [4/5] Generating DEVELOPMENT_GUIDE.md... ✓
  [5/5] Generating IDIOMS.md... ✓

✓ Project bootstrapped successfully.

Next steps
──────────

Equip (when you need more agent equipment):
  codify spec       Generate SDD specification files from this context
  codify skills     Re-run the interactive skills installer
  codify workflows  Re-run the interactive workflows installer
  codify hooks      Re-run the interactive hooks installer

Maintain (as your project evolves):
  codify check      Detect drift between artifacts and current project state
  codify update     Regenerate stale artifacts from the drift report
  codify audit      Score commits against conventions (--with-llm for richer findings)
  codify watch      Foreground watcher: re-runs check on file changes
  codify usage      LLM token + cost summary across runs
```

### Brownfield path

```bash
cd my-existing-go-service/    # already has code
codify init
# → choose "existing"
# → scans the repo, infers stack/dependencies/structure
# → generates context from what it finds (no description needed)
```

The output and Next steps are identical to the greenfield path — only the input differs.

---

## 4. Equip your project (optional, repeatable)

After bootstrap, equip the project with the layers you actually need. Each command is independent and skippable.

```bash
# SDD specification files (CONSTITUTION, SPEC, PLAN, TASKS)
codify spec payments-api --from-context ./output/payments-api/

# Reusable agent skills (architecture, testing, conventions)
codify skills

# Multi-step workflow recipes (bug-fix, release-cycle, spec-driven-change)
codify workflows

# Deterministic guardrails on Claude Code lifecycle events
codify hooks
```

---

## 5. Maintain your project (ongoing)

These commands operate on the bootstrapped project. They detect drift, regenerate stale artifacts, audit commits, and keep cost transparent.

```bash
codify check           # Drift detection — no LLM, zero cost
codify update          # Selective regen when input signals drift
codify audit           # Review commits against conventions (--with-llm opt-in)
codify watch           # Foreground watcher: re-runs check on file changes
codify usage           # LLM token + cost summary
codify resolve         # Interactively fill [DEFINE: ...] markers in artifacts
```

### Wiring `codify check` into CI

```yaml
# .github/workflows/codify.yml
name: codify-check
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go install github.com/jorelcb/codify/cmd/codify@latest
      - run: codify check    # exit 1 if drift detected
```

---

## API keys

Only LLM-backed commands need a key:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."   # for Claude (default)
# or
export GEMINI_API_KEY="AI..."           # for Gemini
```

| No API key required | API key required |
|---|---|
| `config`, `init` (existing scan), `check`, `reset-state`, `audit` (rules-only), `usage`, `hooks`, `skills` (static), `workflows` (static), MCP read-only tools | `generate`, `analyze`, `spec`, `skills --mode personalized`, `workflows --mode personalized`, `update`, `audit --with-llm` |

---

## Where to next

- [`docs/lifecycle-matrix.md`](lifecycle-matrix.md) — Which command applies to **workstation vs project** and **greenfield vs brownfield**.
- [`docs/command-reference.md`](command-reference.md) — Cheatsheet of every command, grouped by phase.
- [`docs/troubleshooting.md`](troubleshooting.md) — Common errors and how to fix them.
- [`docs/adr/`](adr/) — Architectural Decision Records for the design choices behind Codify.
- [Main README](../README.md) — Overview, before/after, full feature documentation.
