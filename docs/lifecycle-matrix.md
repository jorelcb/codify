# Lifecycle Matrix · scope × kind × phase

> Quick reference: which Codify command applies in which combination of **scope** (workstation / project) and **kind** (new / existing). Ordered by lifecycle phase: **Bootstrap → Equip → Maintain**.
>
> See `codify --help` for the diagram view of these phases. See ADR-0007 for the rationale of the bootstrap entry points (`config`, `init`).

---

## The two axes

- **Scope** — *where* the operation lands.
  - **Workstation**: developer machine, persisted under `~/.codify/`, `~/.claude/`, `~/.gemini/`. Done once per laptop.
  - **Project**: a single repository, persisted under `.codify/`, `.claude/`, `.gemini/`. Done once per project (then maintained over time).

- **Kind** — *state* of the project at bootstrap time.
  - **New (greenfield)**: no code yet. Codify generates context from a description.
  - **Existing (brownfield)**: code already in place. Codify scans the repo and generates context from what it finds.

Workstation has no `kind` axis — it is always a one-time setup of the developer's machine.

---

## Matrix

### Bootstrap phase (one-time)

|  | New (greenfield) | Existing (brownfield) |
|---|---|---|
| **Workstation** | `codify config` — wizard for global defaults (target ecosystem, model, locale, preset) + opt-in install of global skills/workflows/hooks | (same — workstation setup is kind-agnostic) |
| **Project** | `codify init` → asks "new", routes to `generate` from a description; persists `.codify/config.yml` and `.codify/state.json` | `codify init` → asks "existing", routes to `analyze` which scans the repo |

> `codify init` is a smart entry point. It asks `new vs existing` interactively and routes to the right pipeline. Use it instead of calling `generate`/`analyze` directly when you want the guided flow.

### Equip phase (per need, repeatable)

These commands install or generate AI agent equipment. They are not tied to a specific *kind* once Bootstrap is done — the same command works for projects that started greenfield or brownfield.

| Command | Scope supported | Purpose |
|---|---|---|
| `codify generate` | Project (new) | Generate context from a description. Usually invoked indirectly by `init`. |
| `codify analyze` | Project (existing) | Generate context by scanning the repo. Usually invoked indirectly by `init`. |
| `codify spec` | Project | Generate SDD specification files (CONSTITUTION, SPEC, PLAN, TASKS) from existing context. |
| `codify skills` | Workstation + Project | Install reusable AI agent skills. `--install global` for workstation, `--install project` for current project. |
| `codify workflows` | Workstation + Project | Install workflow files (multi-step recipes). Same `--install` pattern. |
| `codify hooks` | Workstation + Project | Install Claude Code hook bundles. Same `--install` pattern. Claude only. |

> Future: under v2.0 (Pista D, ADR-0010), `skills`/`workflows`/`hooks` interactive UX is replaced by a unified `codify catalog` command. The non-interactive flags remain available for automation.

### Maintain phase (ongoing)

These commands operate on an already-bootstrapped project. They have no greenfield/brownfield distinction — both lead to the same maintained state.

| Command | Scope | Purpose |
|---|---|---|
| `codify check` | Project | Detect drift between `.codify/state.json` and the current project. |
| `codify update` | Project | Selectively regenerate stale artifacts based on the drift report. |
| `codify audit` | Project | Score recent commits against Conventional Commits and protected-branch rules. `--with-llm` for richer findings. |
| `codify watch` | Project | Foreground watcher: re-runs `check` on file changes. |
| `codify usage` | Workstation + Project | Report LLM usage and cost from local tracking files. |
| `codify resolve` | Project | Interactively fill in `[DEFINE: ...]` markers in existing artifacts. |
| `codify reset-state` | Project | Recompute `.codify/state.json` from the current FS without touching artifacts. |

---

## Recommended sequencing

### First-time developer on a laptop

```
1. Bootstrap (workstation)   →  codify config
2. Bootstrap (project)        →  codify init
3. Equip (optional)           →  codify spec / skills / workflows / hooks
4. Maintain (ongoing)         →  codify check / update / audit / watch / usage
```

### Existing developer adopting Codify on an existing repo

```
1. Bootstrap (workstation)   →  codify config        (skip if already done globally)
2. Bootstrap (project)        →  codify init          (choose "existing")
3. Equip (optional)           →  same as above
4. Maintain (ongoing)         →  same as above
```

### CI / automation context (no TTY)

Codify falls back to non-interactive flags when stdin is not a TTY. `codify config` is silently skipped (built-in defaults apply). Use the dedicated commands (`generate`, `analyze`, `skills --install ...`) with explicit flags. See `codify <command> --help` for the flag surface.

---

## Cross-references

- `codify --help` — phase diagram and command listing grouped by phase.
- ADR-0007 — naming and auto-launch rationale for the bootstrap commands.
- ADR-0010 — upcoming `catalog` command that consolidates `skills`/`hooks` interactive surface.
