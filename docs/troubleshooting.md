# Troubleshooting

Quick reference for the errors most people hit on first contact with Codify. If something is not in this table, open an issue with: command run, exit code, and stderr. The CHANGELOG and ADRs in this repo document most design decisions — usually the answer is in there.

---

## Common errors

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

---

## Diagnosing further

### `codify` itself fails to start

```bash
codify --version          # confirms binary loaded
which codify              # confirms PATH resolution
codify --help             # confirms cobra registered all commands
```

If `--version` prints but a subcommand fails, the issue is in that subcommand's flow — pass `--help` to it for the flag surface.

### Auto-launch prompt loops

If you keep getting prompted to run `codify config` even after declining:

```bash
ls -la ~/.codify/                          # check the state of the dotdir
cat ~/.codify/config.yml                   # confirm config is actually saved
ls -la ~/.codify/.no-auto-config 2>/dev/null   # confirm marker file (for "skip permanently")
```

If `config.yml` exists but the prompt still fires, the auto-launch logic checks for the marker file or env var. Set one and the prompt stops.

### `codify check` reports drift you don't expect

`check` is hash-based and deterministic — it compares the snapshot in `.codify/state.json` against current SHA256s of artifacts and input signals. If it reports drift:

1. Run `codify check --verbose` (if available) or inspect `.codify/state.json` directly.
2. Confirm the file actually changed (`git diff` against last commit).
3. If you intentionally edited an artifact and want that to be the new baseline: `codify reset-state`.
4. If the drift is from input signals (`go.mod`, `Makefile`, etc.) and you want to regenerate artifacts: `codify update`.

### LLM errors during `generate` / `analyze` / `spec`

```bash
codify usage              # verify recent calls and their costs
echo $ANTHROPIC_API_KEY   # confirm the key is set in the current shell
```

If the key is set but calls fail with auth errors, your key may have expired. Generate a new one in the provider console.

---

## Where to look for more

- [Main README](../README.md) — Overview and full feature documentation.
- [`docs/getting-started.md`](getting-started.md) — 5-minute tour with expected outputs.
- [`docs/command-reference.md`](command-reference.md) — Per-command cheatsheet.
- [`docs/lifecycle-matrix.md`](lifecycle-matrix.md) — Scope/kind/phase decision matrix.
- [`docs/adr/`](adr/) — Architectural Decision Records.
- [CHANGELOG.md](../CHANGELOG.md) — Behavior changes and migrations between versions.
