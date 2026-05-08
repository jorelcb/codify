# Command Reference

Cheatsheet of every Codify command, grouped by lifecycle phase. For deep examples, screenshots, and option walkthroughs see the corresponding section in the [main README](../README.md).

For the conceptual phase model see [`docs/lifecycle-matrix.md`](lifecycle-matrix.md). For a 5-minute end-to-end tour see [`docs/getting-started.md`](getting-started.md).

---

## 🚀 Bootstrap (one-time setup)

| Command | What it does | Scope | API key? | Detail |
|---|---|---|---|---|
| `codify config` | User-level wizard. Sets default target ecosystem, preset, locale, model. Persists to `~/.codify/config.yml`. Optionally installs global skills/workflows/hooks. | Workstation | No | [README §Bootstrap](../README.md#-bootstrap-phase-one-time-setup) |
| `codify config get <key>` | Print a single config value. | Workstation | No | — |
| `codify config set <key> <value>` | Set a single config value (no wizard). | Workstation | No | — |
| `codify config unset <key>` | Clear a single config value. | Workstation | No | — |
| `codify config edit` | Open `~/.codify/config.yml` in `$EDITOR`. | Workstation | No | — |
| `codify config list` | Print the full effective configuration. | Workstation | No | — |
| `codify init` | Project-level smart entry point. Asks `new`/`existing` and routes to `generate`/`analyze`. Persists `.codify/config.yml` + `.codify/state.json`. | Project | Depends on path (`new` requires LLM; `existing` does not) | [README §Bootstrap](../README.md#-bootstrap-phase-one-time-setup) |

---

## 🧰 Equip (install context, skills, workflows, hooks, specs)

| Command | What it does | Scope | API key? | Detail |
|---|---|---|---|---|
| `codify generate` | Generate context files (AGENTS.md, CONTEXT.md, etc.) from a description. Usually invoked indirectly by `init`. | Project | Yes | [README §Context](../README.md#-context-generation) |
| `codify analyze <path>` | Generate context by scanning an existing repo. Usually invoked indirectly by `init`. | Project | Yes | [README §Context](../README.md#-context-generation) |
| `codify spec <project>` | Generate SDD specification files (CONSTITUTION, SPEC, PLAN, TASKS) from existing context. | Project | Yes | [README §Specs](../README.md#-spec-driven-development) |
| `codify spec --with-specs` | Full pipeline: context generation + spec generation in one command. | Project | Yes | — |
| `codify skills` | Install reusable AI agent skills. Interactive by default; `--install global\|project` for non-interactive. Static and personalized modes. | Workstation + Project | No (static) / Yes (personalized) | [README §Skills](../README.md#-agent-skills) |
| `codify workflows` | Install workflow files (multi-step recipes: bug-fix, release-cycle, spec-driven-change). Same `--install` pattern. | Workstation + Project | No (static) / Yes (personalized) | [README §Workflows](../README.md#-workflows) |
| `codify hooks` | Install Claude Code hook bundles (linting, security-guardrails, convention-enforcement, all). Claude only. | Workstation + Project | No | [README §Hooks](../README.md#-hooks) |

> **Future (v2.0, ADR-0010)**: `codify skills`/`workflows`/`hooks` interactive UX is replaced by a unified `codify catalog` command. The non-interactive `--install` flags remain available for automation/CI.

---

## 🔧 Maintain (ongoing lifecycle)

| Command | What it does | Scope | API key? | Detail |
|---|---|---|---|---|
| `codify check` | Detect drift between `.codify/state.json` and the current project (input signals + artifacts). No LLM, deterministic. Exits non-zero on drift — wire into CI. | Project | No | [README §Drift Detection](../README.md#-lifecycle-drift-detection) |
| `codify update` | Selectively regenerate stale artifacts based on the drift report. Refuses to overwrite hand-edits unless `--force`. | Project | Yes | [README §Update](../README.md#-lifecycle-update-audit--usage-tracking) |
| `codify update --accept-current` | Alias for `reset-state`: accept current FS as the new baseline. | Project | No | — |
| `codify audit` | Score recent commits against Conventional Commits and protected-branch rules. Rules-only by default. | Project | No (rules) / Yes (`--with-llm`) | [README §Audit](../README.md#-lifecycle-update-audit--usage-tracking) |
| `codify watch` | Foreground watcher: re-runs `check` on file changes. Wrap in tmux/systemd/nohup for session survival. | Project | No | [README §Watch](../README.md#%EF%B8%8F-lifecycle-foreground-watcher-codify-watch) |
| `codify usage` | Report LLM token usage and cost from local tracking files (`.codify/usage.json` or `~/.codify/usage.json`). | Workstation + Project | No | [README §Usage](../README.md#-lifecycle-update-audit--usage-tracking) |
| `codify resolve [files...]` | Interactively fill `[DEFINE: ...]` markers in existing artifacts using the LLM. Skip mode and diff preview supported. | Project | Yes | [README §Resolve](../README.md#%EF%B8%8F-lifecycle-marker-resolution-codify-resolve) |
| `codify reset-state` | Recompute `.codify/state.json` from the current FS without touching artifacts. | Project | No | — |

---

## ⚙️ System

| Command | What it does | Scope | API key? | Detail |
|---|---|---|---|---|
| `codify serve` | Start as MCP (Model Context Protocol) server. Exposes tools for Claude Code, Codex CLI, Gemini CLI. Transports: stdio (default), HTTP. | — | Some tools yes (generative), some no (read-only) | [README §MCP Server](../README.md#-mcp-server) |
| `codify list` | List generated projects. | Workstation | No | — |
| `codify --version` | Print the binary version. | — | No | — |
| `codify --help` | Print the phase diagram + grouped command index. | — | No | — |

---

## Common flags (apply to most commands)

| Flag | Effect |
|---|---|
| `--locale <en\|es>` | Output language for generated artifacts and CLI messages where applicable. |
| `--preset <name>` | Architectural preset (`clean-ddd`, `neutral`, etc.). Overrides project/global config. |
| `--language <go\|typescript\|python\|...>` | Target language for idiomatic guides. |
| `--model <id>` | LLM model override (e.g. `claude-sonnet-4-6`, `gemini-2.5-flash`). |
| `--target <claude\|codex\|antigravity>` | Target ecosystem for skill/workflow/hook delivery. |
| `--no-auto-config` | Skip the soft auto-launch of `codify config` for this invocation. |
| `--from-file <path>` (where applicable) | Read description / input from a file instead of inline prompt. |
| `--with-specs` (on `generate`/`analyze`) | Run the spec pipeline after context generation in a single invocation. |

---

## Configuration merge precedence

When the same key is set in multiple places, this is the order (highest wins):

1. **CLI flags** (e.g. `--preset clean-ddd`)
2. **Project config** (`.codify/config.yml`)
3. **User config** (`~/.codify/config.yml`)
4. **Built-in defaults**

---

## Where to look next

- [`docs/getting-started.md`](getting-started.md) — End-to-end tour with expected outputs.
- [`docs/lifecycle-matrix.md`](lifecycle-matrix.md) — Scope/kind/phase decision matrix.
- [`docs/troubleshooting.md`](troubleshooting.md) — Common errors and diagnosis.
- [`docs/adr/`](adr/) — Architectural Decision Records.
- [Main README](../README.md) — Overview, examples, full feature documentation.
