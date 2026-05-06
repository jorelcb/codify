# Changelog - Codify

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0] - 2026-05-06 - Interactive resolver redesign — LLM-driven prompts, validator, TODO anchors, diff preview, standalone `codify resolve`

> v2.0.5 introduced interactive `[DEFINE]` resolution as a single CLI helper file (`define_resolver.go`, ~200 lines mixing orchestration, LLM call, literal substitute, and prompt loop). v2.1.0 turns that helper into a real bounded context: domain ports, an application-layer command, infrastructure adapters for enrichment and sanitization, and CLI adapters for the prompter and the diff previewer — plus a first-class `codify resolve` subcommand. The UX shifts from "type your answer cold" to "the LLM proposes grounded suggestions you can pick by number, override, or skip", and a post-rewrite validator catches the LLM altering markers it was told to leave alone.

### Added
- **`codify resolve` standalone command.** Marker resolution previously only fired at the tail of `generate` / `analyze` / `init`. v2.1.0 promotes it to a first-class subcommand that operates on existing files. File selection is explicit:
  ```
  codify resolve AGENTS.md CONTEXT.md   # explicit files
  codify resolve --all                  # walk cwd, pick every file with markers
  codify resolve --since=HEAD~5         # files changed in git since <ref>
  ```
  Without args/flags the command exits with a friendly error explaining the three selection modes. Flags: `--no-enrich`, `--no-preview`, `--skip-mode=todo|verbatim` (default `todo`), `--dry-run` (no writes, just reports what would change), `--locale`, `--model`. Use case: shipped v2.0.x context files with markers left intact (declined the inline prompt, or generated before v2.0.5) and want to fill them in without re-running an LLM-backed `generate`. The walk skips `.git`, `node_modules`, `vendor`, `.codify` and binary files (NUL byte in first 4KB).
- **MarkerEnricher (LLM-driven prompts with grounded suggestions).** Before walking the prompter loop, the resolver makes one LLM call per file (`command="resolve-enrich"` for usage tracking) that turns each `[DEFINE: ...]` into a friendlier prompt. The user now sees a numbered list:
  ```
  ── AGENTS.md (2 markers) ──

       40  ## Currency Configuration
       41
    ▸  42  The supported currency is [DEFINE: ISO 4217 currency code], using two

      ¿Qué moneda usa la aplicación?
      (fintech context inferred from line 12)
      Suggestions:
        1) USD [default]
        2) EUR
        3) MXN
      Your answer (1-N, text, Enter for default, s to skip)
  ```
  Input parser accepts: integer 1-N (pick suggestion), free text, Enter (uses default if present, else skips), `s` / `skip` (case-insensitive, explicit skip). When enrichment fails (provider error, malformed JSON, sanitizer rejected everything) or no API key is configured, the prompt degrades to the legacy v2.0.5 form (`Your input for L42 (Enter to skip)`) — never a hard failure.
- **Suggestion sanitizer (anti-hallucination).** The enricher's output passes through a sanitizer (`internal/infrastructure/resolver/sanitizer.go`) that drops suggestions which look invented: URLs (any `http://`, `https://`, `ftp://`, `file://`, `git://` or net/url-recognizable form), file paths (leading `/`, `./`, `../`), multi-line strings, markdown-fenced text, values longer than 50 chars. Deduplicates case-insensitively, caps at 3 kept suggestions, and verifies the LLM's `default` matches one of the survivors (otherwise dropped). Question and rationale are truncated to 280 bytes with an ASCII ellipsis.
- **TODO-anchor skip mode as the new default.** Skipping a marker no longer leaves `[DEFINE: ...]` verbatim — it's replaced with a date-stamped TODO comment in the file's native syntax:
  - `.md` / `.html` / `.htm` / `.xml` → `<!-- TODO 2026-05-06: ISO 4217 code -->`
  - `.go` / `.js` / `.ts` / `.jsx` / `.tsx` / `.java` / `.kt` / `.rs` / `.c` / `.cpp` / `.cc` / `.h` / `.hpp` / `.swift` / `.cs` → `// TODO 2026-05-06: ISO 4217 code`
  - `.py` / `.rb` / `.sh` / `.bash` / `.yml` / `.yaml` / `.toml` / `.ini` → `# TODO 2026-05-06: ISO 4217 code`
  - other extensions → fall back to verbatim (we never write a comment in a syntax we don't recognize)
  
  Rationale: `[DEFINE]` is a Codify-internal token that confuses humans, IDEs, and LLMs that didn't generate the file. `TODO` is universally understood, greppable, surfaced by IDE TODO panels, and the date stamp lets you spot stale ones. Hint preserved after the colon so intent is recoverable. Opt out per invocation with `--skip-mode=verbatim` (CLI) or by passing `service.SkipModeVerbatim` in `ResolveRequest`. `SkipModeUnset` (zero value) resolves to TODO so existing callers get the new default automatically.
- **Post-rewrite validator (anti-hallucination).** After the LLM rewrites the file with the user's answers, `service.ValidateRewrite` re-scans the output and classifies markers by frequency comparison (line numbers shift naturally during rewrite, text counts are the stable signal):
  - `Resolved` — answered AND disappeared (legitimate)
  - `Skipped` — unanswered AND still present (legitimate)
  - `NotApplied` — answered BUT still present (LLM ignored the input)
  - `Lost` — unanswered BUT disappeared (LLM hallucinated a fix to a skipped marker)
  - `Spurious` — present in output but never in input (invented)
  
  When `HasIssues()` returns true (`NotApplied`, `Lost`, or `Spurious` non-empty), the orchestrator logs the breakdown to stderr and falls back to deterministic literal substitution. The user's answers are never lost — worst case the LLM-prose-integrated path is downgraded to the literal path.
- **Diff preview before write.** New `service.DiffPreviewer` port. After the rewrite (LLM or literal) and `ApplySkipMode` produce the proposed bytes but BEFORE writing to disk, the previewer renders a small unified-style diff (changed lines + 2 lines of context) and asks:
  ```
  About to rewrite AGENTS.md:
        line 41
    ▸ - The supported currency is [DEFINE: ISO 4217 currency code], using two
    ▸ + The supported currency is USD, using two
        line 43

  Apply changes? Apply / Discard (keep file as-is) / Edit before applying
  ```
  - **Apply** (default) → write proposed content
  - **Discard** → file untouched, `FilesDiscarded++`
  - **Edit** → write proposed content to a temp file with the file's extension (so `$EDITOR` picks the right syntax highlighting), invoke the editor synchronously (falling back to `vim` / `vi` / `nano` if `$EDITOR` is unset), read the saved bytes back, write those
  
  Per-file: a user resolving 3 files can apply the first two and discard the third without losing the first two's work. Editor failures degrade to "apply proposed content with stderr warning" — preview never blocks the resolve. CLI flag `--no-preview` skips the step entirely.
- **`EvaluationRequest.CacheableSystem` flag.** The enricher makes one LLM call per file with the same system prompt across calls in a session. The Anthropic provider now reads `CacheableSystem` and attaches `CacheControl: ephemeral` to the system block when set, so subsequent calls within the 5-minute TTL window reuse the prompt cache instead of re-billing. Gemini provider ignores the flag (its caching API has a 4096-token minimum that most of our system prompts do not meet). Backwards compatible — existing callers (audit `--with-llm`) leave the flag at zero and see no change.

### Changed
- **`define_resolver.go` reduced to a thin adapter (~30 lines).** The v2.0.5 file (~200 lines mixing concerns) is split following Clean Architecture:
  - `internal/domain/service/resolver.go` — types (`MarkerHit`, `EnrichedMarker`, `PromptedAnswer`, `ResolveDelta`, `SkipMode`); ports (`InteractivePrompter`, `MarkerEnricher`, `RewriteValidator`, `DiffPreviewer`); pure functions (`ScanMarkers`, `LiteralSubstitute`, `ValidateRewrite`, `ApplySkipMode`, `SkipReplacement`)
  - `internal/application/command/resolve_markers.go` — `ResolveMarkersCommand` orchestrator (enrich → prompt → rewrite → validate → skip-mode → preview → write); injectable file IO, stderr, and `today()` for tests; `ResolveRequest` / `ResolveResult`
  - `internal/infrastructure/resolver/enricher.go` — `LLMEnricher` wrapping `service.LLMProvider`, JSON schema + parser, fallback paths
  - `internal/infrastructure/resolver/sanitizer.go` — `SanitizeFinding` and helpers (URL/path/multiline/length filters, dedup, default validation)
  - `internal/interfaces/cli/commands/resolver_prompter.go` — `HuhPrompter` with two UI branches (legacy + enriched), `ParseEnrichedInput`
  - `internal/interfaces/cli/commands/resolver_previewer.go` — `HuhDiffPreviewer`, hand-rolled `renderUnifiedDiff` (no external diff dependency), editor invocation with fallback chain
  - `internal/interfaces/cli/commands/resolve.go` — `NewResolveCmd`, `discoverFiles` (`--all` / `--since=<ref>` / explicit), dry-run shim
  - `internal/interfaces/cli/commands/define_resolver.go` — thin adapter (~30 lines) that wires huh prompter + enricher + previewer and delegates to `ResolveMarkersCommand`. Existing callers in `generate.go` / `analyze.go` / `init.go` are unchanged.
- **Default rewrite path unchanged.** LLM rewrite still preferred when a provider is available, literal substitute on LLM failure. The validator and skip-mode pass run on top — they don't replace the existing two-path logic.
- **Existing `resolve-defines` command tag preserved.** The rewrite call still uses `command="resolve-defines"` for usage tracking (legacy name kept for continuity in `.codify/usage.json`). The new enrichment call uses `command="resolve-enrich"` so the two appear as separate line items.
- **MCP tool surface unchanged.** `codify resolve` is CLI-only in this release. Exposing `resolve_markers` as an MCP tool was deferred — agents calling MCP need a non-interactive variant first (no prompter, no editor invocation), and that surface is not designed yet.

### Fixed
- **UTF-8 byte-budget bug in suggestion truncation.** Initial implementation used `…` (3 bytes in UTF-8) as the ellipsis but counted bytes as if it were 1 char, causing assertions on `len(string)` to be off by 2. Switched to ASCII `...` so the byte count matches the visual count exactly.

### Internal
- New BDD package `tests/bdd/resolve_markers/` with 7 scenarios: single-file resolve writes the answer, multi-file rewrites count, skip-all-with-TODO produces the markdown anchor, verbatim opt-out preserves `[DEFINE]`, top-level decline leaves files untouched, LLM hallucination triggers literal fallback, diff-preview discard leaves the file untouched.
- ~37 new unit tests across `internal/domain/service`, `internal/application/command`, `internal/infrastructure/resolver`, `internal/interfaces/cli/commands`. Suite green end-to-end.
- 14 commits on `main` since v2.0.6 (1 refactor, 7 feat, 6 test, plus the release commit). Conventional-commits compliant; no squash, phase-by-phase commits preserved for archaeology.
- New file layout: `internal/infrastructure/resolver/` is a new package; the existing `internal/infrastructure/llm/` is unchanged.

### Notes
- TODO anchors include the ISO date (`2026-05-06`), no namespace tag — keeps them indistinguishable from human-written `TODO:` comments by tooling, but distinguishable by date when paired with `git blame`. If a `(codify)` or similar namespace becomes useful later, it can be added without breaking existing files.
- Prompt caching only applies to Anthropic. Gemini callers pay full system-prompt input cost per enrichment call. Estimated overhead for a 3-file generation with markers in all 3: ~$0.003 with Anthropic + caching, ~$0.025 with Gemini, $0 without provider.
- `codify resolve --dry-run` writes `(dry-run) would write <path>` to stdout per file but does not invoke the diff preview by default — pass it explicitly if you want both.

## [2.0.6] - 2026-05-06 - Audit error clarity and global hooks path fix

### Fixed
- **`codify audit` error reporting.** `cmd.Output()` was discarding `stderr`, so three distinct git failures collapsed into the same opaque `git log failed: exit status 128`. A new `runGit` helper now captures stderr and translates the common modes:
  - **Empty repository (no commits yet)** -> reported as `Audited 0 commits` with no findings, instead of an error.
  - **Not a git repository** -> clear message pointing the user to run `codify audit` from inside a git repo.
  - **Unknown revision** in `--since` -> surfaces git's literal stderr (e.g. `git: unknown revision or path not in the working tree`).
  - **Other failures** -> first stderr line as context, instead of just the exit code.

  `currentBranch` and `CollectCommitsForLLM` (used by `--with-llm`) also route through `runGit`, so audit error quality is consistent across the rule-based and LLM paths.

- **`codify hooks --install global` produced broken handlers.** Hook scripts were copied to `~/.claude/hooks/`, but the `command` strings written into `settings.json` kept the project-scoped template form `"$CLAUDE_PROJECT_DIR"/.claude/hooks/...`. Result: in any project that didn't also have a project-scoped install, every hook handler failed with `No such file or directory`. A new `rewriteHookCommandsToHome` step now rewrites those commands to `"$HOME"/.claude/hooks/...` before the merge into `settings.json`, but only when scope is `global`. Project scope is left untouched. Two regression tests cover both branches.

## [2.0.5] - 2026-05-06 - Interactive [DEFINE] marker resolution (LLM rewrite by default, literal fallback)

### Added
- **Interactive resolution of `[DEFINE]` markers at the end of `codify generate` / `codify analyze` / `codify init`.** Previously the user got a list of *"L42  [DEFINE: ISO 4217 currency code]"* and had to alt-tab to their editor to fill each one in by hand. Now after the file listing prints, if any markers were emitted across the generated files, Codify offers:
  ```
  Found 7 [DEFINE] marker(s) across 3 file(s).
  Resolve them interactively now? (Y/n) _
  ```
  On `Y`, it walks file-by-file, marker-by-marker, showing 5 lines of surrounding context per spot:
  ```
  ── AGENTS.md (2 markers) ──

       40  ## Currency Configuration
       41
    ▸  42  The supported currency is [DEFINE: ISO 4217 currency code], using two
       43  decimal places...

  Your input for L42 (Enter to skip)
  > USD
  ```
  Empty input = skip; the marker stays verbatim. Per-file, after collecting all answers, the file is rewritten in one of two modes:

  - **B (default)**: when an LLM provider is available (which it always is during `generate`/`analyze` since those require an API key), the file content + the marker→answer map are sent to the LLM with a strict editor prompt that integrates each answer naturally into the surrounding sentence/paragraph and preserves all other content character-for-character. Output is the rewritten file.
  - **A (fallback)**: when the LLM call fails (network, rate limit, etc.) — falls back to literal 1:1 substitution: each `[DEFINE: ...]` is replaced verbatim with the user's answer. Less polished, never loses the user's work.

  Top-level decline (`n`) leaves all markers intact for manual editing — same UX as v2.0.4.

  New file: `internal/interfaces/cli/commands/define_resolver.go` (orchestration + literal substitute + LLM rewrite call). Wired into `runGenerateWithMode` so all three commands (`generate`, `analyze`, `init`) inherit it via the existing delegation chain.

### Internal
- LLM rewrite call uses the existing `provider.EvaluatePrompt` interface (the same one `audit --with-llm` uses). Usage is recorded under `command="resolve-defines"`, so `codify usage` shows the cost separately from the main generation pass.

## [2.0.4] - 2026-05-06 - Constructive [DEFINE] marker reporting

### Changed
- **`[DEFINE]` validation messages reframed and made actionable.** During context generation, the validator inspects each generated file and surfaces spots where the LLM emitted a `[DEFINE: ...]` placeholder — these are the gaps the model didn't have enough info to fill in (anti-hallucination by design). Previously the message was:
  ```
  AGENTS.md has 1 [DEFINE] marker(s) the user must resolve
  ```
  This was both judgmental ("user must resolve") and useless (no line, no marker text — the user had to grep). New format:
  ```
  AGENTS.md — 1 spot needs your input:
    L42  [DEFINE: ISO 4217 currency code]
  ```
  Each marker now appears with its 1-based line number and verbatim text, so the user can jump straight to it. Tone shifted from accusation to collaboration ("needs your input" — the LLM flagged a gap because the description didn't cover that concept; that's the system working, not user failure).

- **Prompt instructions tightened to ban bare `[DEFINE]`.** All prompt builder paths (generate, analyze, personalized) now require the form `[DEFINE: <what is missing>]` with a concrete hint after the colon. A bare `[DEFINE]` told the user nothing about what to fill in. Existing regex still matches both forms (backward compatible for previously-generated files).

### Internal
- `ValidationResult.DefineMarkers` changed from `[]string` to `[]DefineMarker{Text, Line}`. Internal package — no external consumers. Tests updated to verify line + text capture.

## [2.0.3] - 2026-05-06 - `codify init` opt-in installs + clearer skill bundle labels

### Added
- **`codify init` now offers project-scoped skills/workflows/hooks installs at the end of bootstrap.** Mirroring the global flow that `codify config` runs in v2.0.1, after `init` finishes generate/analyze it walks through the same opt-in steps but with `--install project` semantics:
  1. Skills (per category — architecture/testing/conventions, `skip` default)
  2. Workflows (claude or antigravity targets only — bug-fix / release-cycle / spec-driven-change / all / `skip` default)
  3. Hooks (claude only — linting / security-guardrails / convention-enforcement / all / `skip` default)

  Previously `init` just printed *"Recommended next steps: codify skills…"* and exited — the user had to run three separate commands. Now they all run inline in the same wizard. Each step is opt-in with `skip` as default, so a user who only wants the context can hit Enter through and end up with the same artifacts as before.

  The closing message was also cleaned up — it referenced "lifecycle commands arrive starting v1.23" (long since shipped), now lists the available `check / update / audit / watch / usage` commands directly.

### Changed
- **Preset labels in skill prompts now show file count.** `codify config` and `codify init` previously showed labels like *"Clean + DDD (DDD, BDD, CQRS, Hexagonal port skill)"* — clear about *what's* in the bundle but not how *many* skills the bundle contains, so users picking one preset and getting 4-5 SKILL.md files would think *"all of them got installed"*. Labels now end with `— N skill(s)` so the size is upfront. Same treatment applied to workflow preset labels.

### Internal
- `internal/interfaces/cli/commands/config_install.go` refactored — the previous `promptInstallGlobalSkills` / `installGlobalSkill` / `promptInstallGlobalHooks` helpers now take a `scope` parameter (`"global"` / `"project"`). New `promptInstallWorkflows` / `installWorkflow` helpers added. Both `config.go` (global) and `init.go` (project) call into the same code paths — one source of truth for the opt-in install UX.

## [2.0.2] - 2026-05-06 - SOFT auto-launch fires on bare `codify`

### Fixed
- **`codify` (no subcommand) now triggers the SOFT first-time prompt.** Previously, running `codify` alone with no `~/.codify/config.yml` would print the help text and exit silently — `cobra` short-circuits to its default help when no `Run`/`RunE` is defined on the root command, and `PersistentPreRunE` never fires. Combined with the v2.0.1 whitelist (which didn't include the root command name), the most natural first-touch invocation post-install was the *one* path where the auto-launch never had a chance to run.

  Two-line fix:
  - `rootCmd.RunE` now explicitly delegates to `cmd.Help()`. Cobra now treats the root as a runnable command, so `PersistentPreRunE` fires before help is printed.
  - `isInteractiveSuitable` whitelist includes `"codify"` (the root command name) alongside the 14 subcommands from v2.0.1.

  `--help` and `--version` still short-circuit before `RunE` per cobra semantics, so they don't trigger the prompt.

## [2.0.1] - 2026-05-06 - `codify config` wizard fixes

User feedback on the v2.0 wizard surfaced three real issues. All fixed in this patch.

### Fixed
- **`promptModel` no longer auto-selects silently.** When only one provider's API key is set in the environment, the wizard now prints `→ Auto-selected model 'X' (only one provider key found; set Y to choose another)` to stderr instead of returning the only available model with zero output. When only `ANTHROPIC_API_KEY` is set, the picker still appears (Sonnet vs Opus is a real choice); when only `GEMINI_API_KEY`/`GOOGLE_API_KEY` is set, the wizard short-circuits with the explanatory line.
- **SOFT first-time prompt now fires on more commands.** `isInteractiveSuitable` was previously gated to `generate, analyze, spec, skills, workflows, hooks, init`. Natural first-contact commands like `codify usage`, `codify check`, etc. would silently skip the *"Codify isn't configured globally yet — run interactive setup now?"* prompt. Whitelist widened to include `usage, check, audit, update, watch, list, reset-state`. `--help`, `--version`, `serve`, and `config` itself remain excluded by design.

### Changed
- **Wizard reordered + global skills/hooks install opt-in.** The `codify config` wizard previously asked `preset → locale → target → model`, with architecture preset as the *first* question. Architecture preset is project-scoped by nature — what global config really benefits from is letting the user opt into agent-wide skills/hooks. New flow:
  1. Default target ecosystem (claude/codex/antigravity)
  2. Default model (with auto-select notice when applicable)
  3. Default locale (en/es)
  4. Default architectural posture for new projects (still kept — applies to project commands run without `--preset`)
  5. **NEW**: Global skills install — one prompt per catalog category (`architecture`, `testing`, `conventions`), with `skip` as the default. Picks trigger a static-mode global install (`~/.claude/skills/`, `~/.codex/skills/`, `~/.gemini/antigravity/skills/`). No LLM, no API key. Power users still run `codify skills` for finer control.
  6. **NEW** (Claude only): Global hooks bundle install — `linting / security-guardrails / convention-enforcement / all / skip`. Reuses `InstallHooksCommand` to merge into `~/.claude/settings.json` and copy scripts to `~/.claude/hooks/`. Skipped for codex/antigravity because hooks are a Claude Code feature.

  Defaults are saved to `~/.codify/config.yml` BEFORE the install steps run, so a failure in step 5 or 6 doesn't lose the wizard's main output. New helpers live in `internal/interfaces/cli/commands/config_install.go`; `config.go` keeps the wizard skeleton focused.

### Notes
- No breaking changes. Existing `~/.codify/config.yml` files are untouched. Re-running `codify config` on an already-configured system still prints the current config (does NOT re-launch the wizard) — same as before.
- No new tests: all three changes are interactive prompt paths (huh-driven), out of scope for the existing unit + BDD suites.

## [2.0.0] - 2026-05-05 - Lifecycle Custodian (rebrand)

> **Codify is no longer just a one-shot generator.** v2.0 is the moment when the project formally re-positions itself as a **lifecycle custodian** for AI agent context: it generates artifacts, then keeps them honest as the codebase evolves. Same six layers (Context, Specs, Skills, Workflows, Hooks, Lifecycle) that landed across v1.21–v1.25 — now framed and documented as a coherent product.

### What v2.0 actually changes

This is the **smallest possible breaking-change major release**. The narrative shift is large; the code surface change is one flag default.

#### Breaking
- **`--preset default` is removed.** Passing it returns an error with migration instructions, not a silent fallback. Removed from the CLI flag default value, the MCP tool enums, the validation maps, and `LegacyPresetMapping`. (Per [ADR-001](docs/adr/0001-default-preset-transition.md) phase 3.)
- **The default value of `--preset` is now `neutral`** instead of `clean-ddd`. New AGENTS.md/CONTEXT.md generations made without `--preset` will be architecturally neutral — no DDD/Clean opinion baked in unless you ask for it. Existing artifacts are unaffected; `codify check` won't flag this as drift.

That's the entire functional breaking change list. Everything else continues to work.

### Migration

A single CLI invocation covers most cases. See the [Migrating from v1.x](README.md#-migrating-from-v1x) section in README for full guidance:

```bash
# Match v1.x default behavior (DDD/Clean)
codify config set preset clean-ddd

# OR adopt the new default (no architectural opinion)
codify config set preset neutral
```

Anywhere you previously passed `--preset default`, replace with `--preset clean-ddd` (same behavior) or `--preset neutral` (the new default).

### What did NOT change

- All targets remain supported: `claude`, `codex`, `antigravity`. The v1.26 antigravity deprecation was reversed via [ADR-009](docs/adr/0009-antigravity-deprecation-reversal.md).
- All commands, flags, MCP tools (10 total), config/state/usage schemas, pricing table — unchanged.

### Rebrand: README and positioning

The README was rewritten to reflect the lifecycle custodian framing without changing the underlying capabilities. New tagline: *"Generate, audit, and evolve your AI agent's context across the whole project lifecycle."* Problem section now articulates the drift problem explicitly. Before/After example extended to "Day 22" showing the watcher/check/update flow.

### Cancelled before shipping

- **`codify migrate` command** — was planned for v1.26. Cancelled with the antigravity deprecation (ADR-009).
- **v1.26 release entirely** — cancelled. v1.25 → v2.0 directly.

### Decisions

- **v2.0 is the rebrand, not a code rewrite.** The lifecycle custodian identity was built incrementally across v1.21–v1.25 (per [ADR-006](docs/adr/0006-incremental-release-model.md)).
- **Minimum breaking surface.** v2.0 touches one flag default, the minimum required for the rebrand to be honest.
- **Migration guide is documentation, not a command.**

### Versions

`.version` 1.25.0 → 2.0.0; README + README_ES badges; MCP `serverVersion`. Total BDD packages: 9, 30+ scenarios — all green.

## [1.25.0] - 2026-05-05 - Lifecycle: foreground watcher (`codify watch`)

### Added
- **`codify watch`** — foreground file watcher that monitors paths registered in `.codify/state.json` and re-runs drift detection when they change. Designed for active development sessions, NOT a background daemon. Exits cleanly on Ctrl+C. Behavior:
  - Loads `.codify/state.json` once at startup; exits 2 if missing
  - Subscribes via `fsnotify` to the parent dirs of registered input_signals + artifacts
  - Debounces events (default 2s) before firing drift detection
  - Prints drift reports to stdout, keeps watching
  - `--auto-update` fires `codify update` on detected drift (records LLM usage)
  - Without `--auto-update`, drift is informational only — user runs `check`/`update` manually
- Flags: `--debounce <duration>` (e.g. `500ms`, `2s`), `--auto-update`, `--strict`, `--no-tracking`, `--output`.
- New package `internal/infrastructure/watch` — wraps `fsnotify` with a debouncer, scope-limited to paths from `state.json`. Pure: takes paths + callback, returns events. Tested with temp dirs (5 unit tests + 6 BDD scenarios).
- New BDD package `tests/bdd/watch_loop` covering startup, debounce coalescing, scope filtering, clean cancellation, and construction errors.
- New ADR documenting the architectural decision: [ADR-008 — `codify watch` model](docs/adr/0008-watch-model-decision.md).

### Decisions (per ADR-008)
- **Foreground, not daemon.** No PID file management, no `--detach`, no signal handling beyond `Ctrl+C`. Users who need persistence wrap with `tmux` / `nohup` / `systemd` — Codify intentionally stays out of the daemon business.
- **Scope-limited watching.** Only the paths in `state.json` are watched (input_signals + artifacts). No recursive walk, no ignore patterns. Bounded ~20 files for a typical project.
- **In-house implementation, not config generation.** Modelo B (generating configs for `lefthook`/`pre-commit`/`watchexec`) was evaluated and explicitly NOT chosen as the primary mechanism. Rationale in ADR-008. The README now documents `codify check` integration patterns for those tools as a complementary option for git-hook-driven validation.
- **No reload-on-config-change.** If `state.json` changes mid-run, watch does NOT auto-reload. User must `Ctrl+C` and re-run. This is intentional: reload-on-config-change is a daemon feature we explicitly skipped.

### Changed
- MCP server version bumped to 1.25.0.
- `go.mod` adds `github.com/fsnotify/fsnotify v1.10.x` as a direct dependency.

### Notes
- `codify watch` does not currently invoke `update` interactively — `--auto-update` is non-interactive and records LLM usage automatically. To preview changes without running an LLM, run `codify watch` (read-only) and use the manual `update --dry-run` afterward.
- Debounce default 2s is conservative. Editors that save-on-keystroke (e.g. Zed with auto-save) may benefit from `--debounce 500ms` or `1s`. CI environments that batch FS events may want `--debounce 5s`.

## [1.24.1] - 2026-05-05 - audit --with-llm + README sync

### Added
- **`codify audit --with-llm`** is now real (no longer a stub). When the flag is set, after running the rules-only baseline the audit additionally:
  - Collects recent commits with header, body, and file stats
  - Loads AGENTS.md from the project root
  - Sends a structured prompt to the configured LLM provider asking it to flag commits that don't align with the documented conventions
  - Parses the JSON response into `agents_alignment_issue` findings, marked `Heuristic=true` and tagged `(heuristic)` in human output
  - Records the LLM call in `.codify/usage.json` and `~/.codify/usage.json` like every other LLM-backed command
- New `EvaluatePrompt` method on `service.LLMProvider` for one-shot text completion (different from the multi-file `GenerateContext` flow). Implemented in Anthropic, Gemini, and Mock providers.
- `internal/infrastructure/audit/llm_prompt.go` — system prompt + user prompt builder + JSON parser with markdown-fence stripping. Heuristic findings always have `Heuristic=true`. Invalid severity values fall back to `minor` to avoid double-counting noise.
- 4 new BDD scenarios under `tests/bdd/audit_rules` covering the LLM JSON parser (happy path, fenced output, invalid severity fallback, non-JSON rejection). Total audit BDD scenarios: 11.
- 7 new unit tests in `internal/infrastructure/audit/llm_prompt_test.go` covering parser edge cases.

### Changed
- `audit --with-llm` no longer prints the v1.24.0 NOTICE about being unimplemented; the flag is a real opt-in that requires `ANTHROPIC_API_KEY` or `GEMINI_API_KEY`. If the API key is missing or the LLM call fails, the audit emits a WARNING and falls back to rules-only output (exit code reflects the rules findings).
- `audit` accepts new flags: `--model` (override default LLM) and `--no-tracking` (skip usage recording for this invocation).
- README and README_ES synced with the actual v1.24 surface (the inconsistencies that accreted across v1.19–v1.24):
  - "four things" / "cuatro cosas" → "six layers" / "seis capas" with an updated diagram showing Hooks and Lifecycle as first-class layers
  - Quick Start section: old "Five ways" inventory expanded into a complete command surface section showing Bootstrap + Context + Specs + Skills + Workflows + Hooks + Lifecycle, plus a "free vs API-key-required" breakdown
  - Top-of-file table of contents updated to link Configuration & Bootstrap, Hooks, Drift Detection, and Update/Audit/Usage sections (previously invisible from the TOC)
  - Sample output `Preset: default` → `Preset: clean-ddd` (the rename landed in v1.21; the example was stale)
  - MCP tools table now includes `get_usage` (added in v1.24.0); knowledge-tools section renamed to "Read-only tools" since `get_usage` is read-only but isn't really "knowledge"
- MCP server version bumped to 1.24.1.

### Decisions
- **Heuristic findings are always tagged.** The LLM mode never produces findings that look identical to the deterministic rules — every LLM-sourced finding has `Heuristic=true`, uses the dedicated `agents_alignment_issue` kind, and prints with a `(heuristic)` suffix in human output. This keeps the trust boundary visible.
- **The LLM call augments, never replaces, rules-only.** Both passes always run; LLM findings are appended to rule findings. If LLM fails, we degrade gracefully with a WARNING — never blocking the deterministic baseline.
- **Severity validation on parse.** If the LLM returns an unrecognized severity (e.g. "critical"), the parser silently downgrades to `minor`. Better to under-report than to surface fake-significant noise and burn user trust.

## [1.24.0] - 2026-05-05 - Lifecycle: update + audit + usage tracking

### Added
- **`codify update`** — selective regeneration when input signals drift. Internally runs `check`; if drift is detected, delegates to `analyze` to refresh artifacts (records LLM usage). Detects the "user hand-edited AGENTS.md" case and refuses to regenerate without `--force` (preserves intent). Flags: `--dry-run`, `--force`, `--accept-current` (alias for `reset-state`), `--no-tracking`. Exit codes: 0 no-op or success, 1 hand-edits without `--force`, 2 missing snapshot.
- **`codify audit`** — review recent commits against project conventions.
  - **Default mode (rules-only):** deterministic, zero-cost, no LLM. Validates Conventional Commits format (type[scope][!]: subject), rejects unknown types, flags trivial messages (`wip`, `fix`, `update`, etc.), enforces 72-char header limit, detects direct commits to protected branches (`main`, `master`, `develop`, `production`).
  - **`--with-llm`:** opt-in heuristic mode. **Planned for v1.24.1**; in v1.24.0 falls back to rules-only with a NOTICE.
  - Exit codes: 0 clean, 1 significant findings (or any with `--strict`).
  - Flags: `--since <ref>`, `--limit N`, `--strict`, `--rules-only`, `--with-llm`, `--json`.
- **`codify usage`** — read LLM cost tracking from local files. Reports total cost (USD cents), token counts (input/output/cache), call counts. Subcommand options:
  - `--global` to report `~/.codify/usage.json` (cross-project aggregate)
  - `--since 7d|24h|30m` to filter by recency
  - `--by command|model|provider` to group totals
  - `--json` for machine-readable output
  - `--reset` archives current log as `.bak.<timestamp>` and starts fresh
- **Usage tracking now records every successful and failed LLM call** automatically from both Anthropic and Gemini providers. Each entry captures: timestamp, command, provider, model, input/output/cache tokens, computed cost (using the public list-price pricing table), duration, success flag, project name, and `pricing_table_version`.
- **Triple opt-out** for usage tracking (per [ADR-005](docs/adr/0005-llm-usage-tracking.md)):
  - Per-invocation flag: `--no-tracking` (on `update`, others as needed)
  - Persistent env: `CODIFY_NO_USAGE_TRACKING=1`
  - Persistent marker: `~/.codify/.no-usage-tracking`
- **Pricing table** embedded at `internal/domain/usage/pricing.go` with version `2026-05`. Covers Claude Sonnet 4.6, Opus 4.6/4.7, and Gemini 3.1 Pro Preview. Each entry records the table version so historical reports remain interpretable when prices change.
- **MCP tool `get_usage`** — read-only access to usage logs from agents. Parameters: `scope` (project/global), `since`, `by`. No LLM call inside; pure file read.
- New domain packages:
  - `internal/domain/usage` — `Entry`, `Totals`, `Log` types + `CostCents()` calculator + pricing table
  - `internal/domain/audit` — `Kind`, `Severity`, `Finding`, `Report` types
- New infrastructure packages:
  - `internal/infrastructure/usage` — `Repository`, `Recorder`, paths, opt-out resolution
  - `internal/infrastructure/audit` — deterministic rules engine (Conventional Commits parser, protected-branch detector, trivial-message matcher)
- Both Anthropic and Gemini providers gained a `recordUsage()` shim called after every `GenerateContext` invocation (success or failure). The shim is best-effort; tracking errors never break the parent command.
- New BDD packages with 14 scenarios total:
  - `tests/bdd/audit_rules` — 7 scenarios (valid CC, breaking change, invalid type, header length, trivial messages, generic non-CC, merge commit detection)
  - `tests/bdd/usage_tracking` — 7 scenarios (record/read roundtrip, cost calculation for known/unknown models, opt-out via env/flag, append accumulation)

### Changed
- MCP server version bumped to 1.24.0; total tool count from 9 to 10.

### Deferred
- **`codify audit --with-llm`** is implemented as a clear stub returning a NOTICE message. Full implementation in v1.24.1 once we settle on the prompt structure and JSON-output contract.
- **GitHub Action** as a separately published repo (`codify/check-action@v1`): documented as a workflow YAML pattern in README; the published action shipped to `marketplace` deferred to a v1.24.x patch.

### Decisions
- **Cost is informational, not authoritative.** The pricing table reflects public list prices and may not match negotiated discounts. Users with custom pricing should treat the `cost_usd_cents` field as a relative indicator.
- **Tracking is on by default.** A privacy-first default (off by default) was rejected because the value of cost visibility outweighs the friction of opt-out. The triple opt-out is robust enough that any user concerned about local-only telemetry can disable it permanently.
- **Per-call recording, not aggregated.** Each LLM invocation is a separate entry; aggregates are recomputed on read. This avoids consistency bugs between entries and totals when files are hand-edited.
- **Atomic file writes.** Both `usage.json` and `state.json` use `.tmp` + rename; `state.json` adds `.bak` backup, but `usage.json` is append-only and reproducible (regenerable via `--reset`) so no backup is taken on each save.

## [1.23.0] - 2026-05-05 - Lifecycle: drift detection (`check` + `reset-state`)

### Added
- **`codify check`** — drift detection. Compares `.codify/state.json` (snapshot) against the current FS state and reports any divergence: `artifact_modified`, `artifact_missing`, `artifact_new`, `signal_changed`, `signal_added`, `signal_removed`. Fully deterministic — no LLM calls, no network, zero cost. Suitable for CI:
  - exit `0` when no significant drift
  - exit `1` when significant drift detected (default) or any drift (with `--strict`)
  - exit `2` when no `.codify/state.json` exists (project not bootstrapped)
  - flags: `--strict`, `--output <dir>`, `--json`
- **`codify reset-state`** — recompute `.codify/state.json` from the current FS without touching artifacts. Use case: user intentionally edited `AGENTS.md` and wants to accept the edits as the new baseline (avoids re-running an LLM). Read-only over generated files; only updates `state.json`. Atomic write with `.bak` backup. Flag `--dry-run` shows what would be recomputed.
- New `internal/infrastructure/snapshot` package: `Build()` constructs a complete `State` from the FS (artifacts + input signals + git context). Pure / deterministic. Hashes via SHA256.
- New `internal/domain/drift` package: types for drift entries (`Kind`, `Severity`, `Entry`, `Report`) and `SeverityOf` classification.
- New `internal/infrastructure/drift` package: `Detector.Detect()` compares snapshot vs current FS and produces a report.
- New BDD package `tests/bdd/drift_detection` with 8 scenarios covering no drift, modified artifact, missing artifact, new artifact, signal changed, signal removed, multi-drift, and severity classification.
- Snapshot writing is now wired into `generate` and `analyze` automatically — every successful generation persists `.codify/state.json`. Lifecycle commands operate against this without further setup.
- `init` was simplified: removed manual state.json creation and now relies on `generate`/`analyze` to write the initial snapshot, then re-writes once with the project target metadata that those commands don't have access to.

### Changed
- `init.go` no longer constructs `state.json` by hand; uses the shared `writeProjectSnapshot` helper that calls `snapshot.Build()`.
- MCP server version bumped to 1.23.0.

### Decisions
- **Drift detection is deterministic.** No LLM is involved — comparison is pure SHA256 + filesystem stat. This is the line between v1.23 (free, local-only) and v1.24 (`codify update` + `codify audit` use LLM optionally).
- **Severity classification is fixed**, not configurable in v1.23: `artifact_modified`, `artifact_missing`, `signal_changed`, `signal_removed` are *significant*; `artifact_new`, `signal_added` are *minor*. `--strict` flips minor into failure mode for users who want zero tolerance.
- **state.json is single source of truth.** The same snapshot writer is used by generate, analyze, init, and reset-state — no parallel state-tracking mechanisms.
- **Idempotent re-writes.** Calling `writeProjectSnapshot` multiple times is safe; each call overwrites `state.json` atomically with `.bak` backup.

## [1.22.0] - 2026-05-05 - Bootstrap UX (`config` + `init`)

### Added
- **`codify config`** — user-level configuration command (`~/.codify/config.yml`). Without args: launches interactive wizard if config doesn't exist, prints current config if it does. Subcommands: `get <key>`, `set <key> <value>`, `unset <key>`, `edit` (opens `$EDITOR`), `list`. Valid keys: `preset`, `locale`, `language`, `model`, `target`, `provider`, `project_name`.
- **`codify init`** — project-level bootstrap command. Asks "new vs existing" and routes internally to `generate` (with description inline or from file) or `analyze` (scanner + LLM). Persists `.codify/config.yml` and `.codify/state.json`. Prints recommended next steps for skills/workflows/hooks instead of bundling them (composition over mega-command, per ADR-007).
- **Auto-launch SOFT** of the config wizard on first run (per [ADR-007](docs/adr/0007-bootstrap-commands-naming.md)). Trigger: TTY interactive + interactive-suitable command + no `~/.codify/config.yml` + no opt-out. Three opt-out mechanisms: flag `--no-auto-config` (per-invocation), env `CODIFY_NO_AUTO_CONFIG=1`, marker file `~/.codify/.no-auto-config` (created by selecting "skip permanently").
- New `internal/domain/config` package: `Config` struct with YAML schema, `Merge` for precedence chain, `Get`/`Set`/`Unset` helpers, `BuiltinDefaults`, schema versioning.
- New `internal/infrastructure/config` package: `Repository` for atomic save/load, automatic `.bak` backup before overwriting, `LoadEffective()` resolves the chain (builtin < user < project), path discovery helpers (`UserConfigPath`, `ProjectConfigPath`, `UserNoAutoConfigMarker`).
- New `internal/domain/state` and `internal/infrastructure/state` packages: `state.json` schema (per [ADR-004](docs/adr/0004-state-json-schema.md)) and atomic JSON repository. v1.22 only WRITES; lifecycle commands (v1.23+) consume.
- BDD test package `tests/bdd/config_merge` with 8 scenarios covering precedence chain, roundtrip persistence, backup creation, and key validation.
- Unit tests for both new domain (`config_test.go`) and infrastructure (`repository_test.go`) packages with edge cases for missing files, precedence ordering, and key handling.

### Changed
- `runGenerateInteractive` and `runAnalyzeInteractive` now call `loadEffectiveConfig()` at startup and apply config defaults to unset flag values via `applyConfigDefaults()`. Flags retain priority; the interactive prompt still kicks in if both flag and config leave a field empty in TTY mode. Files: `config_merge.go` (new helper) and the existing CLI commands.
- Root command now has `PersistentPreRunE` that invokes `MaybeAutoLaunchConfig` before each subcommand. Side-effect-free if the config already exists or the command is not interactive-suitable.
- MCP server version bumped to 1.22.0.

### Decisions documented
- Bootstrap commands coexist with the existing standalone commands (`generate`, `analyze`, `skills`, `workflows`, `hooks`) — they are NOT a replacement. CLI flag-driven flow remains the automation surface for CI/MCP per ADR-007.
- `config` is the user-level command and supports CRUD subcommands (idiomatic to `git config`/`npm config`/`gh config`); `init` is the project-level command and is composed (not duplicated) on top of `generate`/`analyze`.

## [1.21.0] - 2026-05-05 - Architectural diversity (4 presets)

### Added
- New context preset `hexagonal` — Ports & Adapters architecture, lighter than clean-ddd. Templates in `templates/{en,es}/hexagonal/` (agents, context, development_guide, interactions)
- New context preset `event-driven` — CQRS + Event Sourcing + Sagas. Templates in `templates/{en,es}/event-driven/`
- New skills under `architecture/hexagonal`: `port_definition`, `adapter_pattern`, `dependency_inversion`, `hex_integration_test` (en + es)
- New skills under `architecture/event-driven`: `command_handler`, `domain_event`, `event_projection`, `saga_orchestrator`, `event_idempotency` (en + es)
- Architecture skills catalog now maps 1:1 with context presets (4 architecture preset names: `neutral`, `clean-ddd`, `hexagonal`, `event-driven`)
- 7 ADRs documenting the strategic shift in `docs/adr/`: default preset transition, antigravity deprecation, no public Go library, state.json schema, LLM usage tracking, incremental release model, bootstrap commands naming
- Updated unit tests in `skills_catalog_test.go` covering all 4 architecture presets and the legacy alias mapping

### Changed
- Renamed preset `default` → `clean-ddd` (templates moved via `git mv` to preserve history; covers both `templates/{en,es}/default/` and `templates/{en,es}/skills/default/`)
- Skill option `clean` renamed to `clean-ddd` — mappings updated across catalog, MCP enums, and CLI
- Default value of `--preset` flag is now `clean-ddd` (was `default`); behavior unchanged this release (DDD/Clean templates)
- Interactive preset menu reordered: `neutral` listed first as the recommended choice; lists all 4 presets
- MCP tool `generate_context` and `analyze_project` now expose `preset` as an enum: `neutral`, `clean-ddd`, `hexagonal`, `event-driven`, `workflow`, `default` (last one deprecated)
- README and README_ES rewrote the Presets section: new comparative table of 4 presets, deprecation notice for `default`, examples for each preset

### Deprecated
- `--preset default` is deprecated. It still works in v1.x and resolves to `clean-ddd` with a stderr warning. Removed in v2.0 per [ADR-001](docs/adr/0001-default-preset-transition.md). The default value of `--preset` will then change to `neutral`.
- The `clean` skill option name is also deprecated as a backward-compat alias; prefer `clean-ddd`. Both resolve to the same templates during v1.x.

### Migration notes
- CI scripts that pass `--preset default` continue to work; consider updating to `--preset clean-ddd` to silence the warning. Before v2.0, decide whether you want `clean-ddd` (current behavior) or `neutral` (the new default) and pass it explicitly.
- No template content changed for clean-ddd; only the directory was renamed. Existing generated AGENTS.md outputs are unaffected.

## [1.20.0] - 2026-05-04 - Codebase audit cleanup + auto-activated hooks

### Added
- **Hooks auto-activation**: `codify hooks --install project` (or `--install global`) now merges the bundle directly into `.claude/settings.json` and copies scripts to `.claude/hooks/`. Replaces the v1.19.0 ceremony of writing to `./codify-hooks/` for manual merge.
- `--dry-run` flag on `codify hooks` that prints the proposed `settings.json` merge without writing anything
- New `internal/infrastructure/settings` package with idempotent merge, automatic backup before overwrite, atomic write via `.tmp` + rename, and discovery helpers (`GlobalSettingsPath`, `ProjectSettingsPath`, `GlobalHooksDir`, `ProjectHooksDir`, `ResolveScope`)
- New `InstallHooksCommand` orchestrator at `internal/application/command/install_hooks.go` (idempotent across runs; reports added vs skipped handlers and conflicting scripts)
- LLM output validators (`internal/infrastructure/llm/validators.go`) detecting `[DEFINE]` markers, unbalanced code fences, missing frontmatter, and missing required `disable-model-invocation` / `allowed-tools` fields for workflow-skills. Wired into both Anthropic and Gemini providers; warnings surface to stderr after each generation
- Anthropic prompt caching: system prompt is marked with `CacheControlEphemeral` so the per-file generation loop reuses the cached system prompt instead of paying for it on every guide
- Mock LLM provider (`internal/infrastructure/llm/mock_provider.go`) for testing application/command orchestration without hitting real APIs
- MCP parameter enums for `generate_skills`, `generate_workflows`, `generate_hooks`, plus `install_scope` parameter on `generate_hooks` (`global` / `project` / `preview`, default `preview`)
- `[DEFINE]` heuristic warnings printed by the CLI when generated content contains unresolved markers
- New catalog helpers `AllSkillPresetNames` and `WorkflowPresetNames` for MCP enum population
- Output examples (`<output_example>` blocks) embedded in the personalized skills, workflows, and workflow-skills system prompts (few-shot anchoring)
- Anti-hallucination grounding rules now applied uniformly to skills, workflows, and workflow-skills modes (previously only generate/analyze/spec had them)
- Locale fallback warning to stderr when the requested locale is unsupported
- New tests: `settings_test.go`, `install_hooks_test.go`, `validators_test.go`, `provider_factory_test.go`, plus expanded coverage of `application/dto`, `application/command`, and `domain/catalog`

### Changed (BREAKING — minor)
- `codify hooks --install project|global` now auto-activates immediately. The previous v1.19.0 default of writing to `./codify-hooks/` for manual merge is moved to `--output PATH` (preview mode)
- Antigravity skills paths corrected to `.agent/skills/` (singular) — was `.agents/skills/` in v1.19.0. Aligns with the existing `.agent/workflows/` convention
- `codify spec` no longer hard-requires `--from-context` at the flag level; the path is prompted in interactive mode (still required overall)
- LLM dispatch fails loudly on unknown `Mode` values instead of silently falling back to `generate` mode

### Fixed
- `mkfs` regex bypass in `security-guardrails/block-dangerous-commands.sh`: `mkfs.ext4` and `mkfs.btrfs` were passing through (regex matched only alpha characters)
- `jq` parse failures in hook scripts now fail closed (exit 2) for blocking PreToolUse hooks instead of silently allowing the operation; the non-blocking lint hook still exits 0 but logs to stderr
- `ProjectContext` is now validated as non-empty in providers when mode requires it (skills/workflows/workflow-skills); previously the LLM received an empty placeholder and invented stack details
- `workflow-skills` system prompt now documents the optional frontmatter fields `agent`, `user-invocable`, and `context` (audit fix #6)
- `promptModel` no longer offers Anthropic options when no `ANTHROPIC_API_KEY` is set, and no Gemini options when no `GEMINI_API_KEY`/`GOOGLE_API_KEY` is set. Returns an explicit error if no key at all is found, instead of presenting unusable choices

### Removed
- Dead code: `validPresets` map in `internal/interfaces/mcp/server.go`
- Manual merge ceremony for hooks: the README sections describing copy/paste of `hooks.json` into `settings.json` are gone (preview mode kept as escape hatch with explicit messaging)

## [1.19.0] - 2026-05-04 - Claude Code hook bundles (deterministic guardrails)

### Added
- New `codify hooks` command that generates Claude Code hook bundles — `hooks.json` (block to merge into `settings.json`) plus auxiliary `.sh` scripts referenced by it
- Three preset categories: `linting` (auto-format on Edit/Write), `security-guardrails` (block dangerous commands and protect sensitive files), `convention-enforcement` (Conventional Commits + protected branches via the `if` field)
- `all` preset that merges the three bundles into a single `hooks.json` (per-event arrays unioned; Claude Code dedupes identical commands automatically)
- Templates in `templates/{en,es}/hooks/{preset}/` with localized stderr messages (English / Spanish)
- New MCP tool `generate_hooks` exposing the same functionality to AI agents (parameters: `preset`, `locale`, `output`)
- New domain bounded context: `internal/domain/catalog/hook_catalog.go` (`HookCategories`, `HookMetadata`, `FindHookCategory`, `HookPresetNames`)
- New DTO `internal/application/dto/hook_config.go` with `ValidHookPresets` map
- New application command `internal/application/command/deliver_hooks.go` with merge-aware delivery (handles `all` preset by reading multiple template directories)
- BDD test suite `tests/bdd/hook_catalog/` with 9 scenarios
- Unit tests `internal/domain/catalog/hook_catalog_test.go` covering preset resolution, metadata, and category lookup

### Design decisions
- **Static-only**: hooks are catalog-driven; no LLM personalization (universal patterns don't benefit from project-specific tuning)
- **Claude-only target**: hooks are exclusive to Claude Code — Antigravity and Codex have no equivalent
- **Standalone output**: codify never auto-merges into `settings.json`; the user merges manually after review (zero risk of clobbering existing config)
- **Default output is `./codify-hooks/`** instead of `.claude/hooks/` so the user immediately sees the generated bundle as something requiring activation
- **Bash + jq required**: scripts use `jq` to parse the hook JSON input (Linux/macOS native; Windows requires Git Bash or WSL)
- **Regex-based detection**: simple `grep -E` rather than AST parsing — sufficient to stop careless agents, not motivated adversaries (documented honestly in script comments)

### Changed
- MCP server registers a new tool (`generate_hooks`) bringing the total to 8 tools
- MCP server version bumped to 1.19.0
- README/README_ES badges updated to v1.19.0

## [1.18.0] - 2026-04-27 - Spec-driven change lifecycle workflow

### Added
- New workflow preset `spec-driven-change` generating three Claude Code skills (`/spec-propose`, `/spec-apply`, `/spec-archive`) for the OpenSpec-compatible SDD lifecycle
- Templates: `spec_propose.template`, `spec_apply.template`, `spec_archive.template` in en/es locales
- BDD scenario for `spec-driven-change` resolution (3 mappings)
- Unit test coverage for the new preset's multi-template mapping

### Changed
- `Resolve("all")` now returns 5 mappings (2 single-file presets + 3 from `spec-driven-change`)
- CLI `--preset` flag description updated to list `spec-driven-change`, `bug-fix`, `release-cycle`
- README workflow catalog table updated to list current presets only
- Workflow category interactive menu now lists 3 options
- MCP server version bumped to 1.18.0

### Removed
- `feature-development` preset and its templates (`feature_development.template` in en/es) — replaced by `spec-driven-change` which is a strict superset that absorbs Git mechanics (branch, commits, PR, merge) and adds formal proposal artifacts (`openspec/changes/<id>/proposal.md`, `design.md`, `tasks.md`, spec deltas)
- `feature_development` entry from `WorkflowMetadata` and `fileOutputNames` map

## [1.17.0] - 2026-04-23 - Multi-target workflow skill consolidation

### Changed
- Unified workflow command routing for both Claude and Antigravity targets through `DeliverStaticWorkflowsCommand` and `GenerateWorkflowsCommand`
- Claude target install paths: `~/.claude/skills/` (global), `.claude/skills/` (project)
- MCP server routing simplified: both targets handled by single command pair
- `BuildWorkflowSkillSystemPrompt()` for personalized SKILL.md generation aligned with Claude's native skill format

## [1.16.0] - 2026-04-14 - Enhanced analyze with enriched scanner and differentiated prompt

### Added
- Differentiated system prompt for `analyze` command: scan data treated as factual ground truth (`<scan_trust>` section), reducing unnecessary `[DEFINE]` markers on detected signals
- Mode propagation: `ProjectConfig.Mode` flows through `GenerationRequest` to provider switch (`"analyze"` mode)
- Expanded context file detection from 7 to 18+ files: CONTRIBUTING.md, ARCHITECTURE.md, .claude/CLAUDE.md, .editorconfig, .github/CODEOWNERS, openapi.yaml/json, swagger.yaml/json, schema.graphql, CHANGELOG.md (truncated to 50 lines)
- Glob-based context file discovery: `.cursor/rules/*.md`, `docs/adr/*.md`, `proto/*.proto`
- Large context file truncation: 200-line limit with `[... truncated ...]` marker
- Makefile target parsing: extracts real target names (excludes `.PHONY`, comments)
- Taskfile task parsing: extracts task names from `tasks:` section (supports Taskfile.yml/yaml)
- `BuildTargets` field in `ScanResult` — formatted as `**Build Targets:**` section for LLM
- Testing pattern detection: test files (`*_test.go`, `*.spec.ts`, `*.test.js`, `*.feature`), frameworks from deps (godog, Jest, Vitest, Mocha, RSpec), coverage config (codecov, jest.config, pytest.ini, .nycrc)
- CI/CD workflow summarization: parses GitHub Actions (.yml) and GitLab CI files, extracts triggers and job names
- `CIWorkflowSummary` struct with File, Triggers, Jobs — formatted as `**CI/CD Pipelines:**` section
- Dependency parsing for Rust (`Cargo.toml` [dependencies]), Java (`pom.xml` artifactId), Ruby (`Gemfile` gems)
- Framework detection for Java (Spring Boot, Quarkus, Micronaut), Ruby (Rails, Sinatra, Hanami), Rust (Rocket)
- Smart README filtering: removes badges, HTML comments, Table of Contents sections, collapses excessive blank lines — applied before truncation for 100 lines of meaningful content
- 23 new unit tests across scanner and prompt builder

### Changed
- `analyze` command now uses `runGenerateWithMode("analyze")` instead of `runGenerate()` — MCP handler updated accordingly
- `FormatAsDescription()` output enriched with Build Targets, Testing Patterns, and CI/CD Pipelines sections
- README filtering applied before line-count truncation (100 useful lines instead of 100 raw lines)

## [1.15.0] - 2026-04-13 - Claude Code native workflow skills

### Added
- Claude Code native skill generation: `--target claude` produces SKILL.md files with frontmatter (`name`, `description`, `disable-model-invocation`, `allowed-tools`)
- `StripAnnotationLines()` — removes Antigravity execution annotations (`// turbo`, `// capture:`, `// if`) from workflow content
- `BuildWorkflowSkillSystemPrompt()` — LLM prompt for annotation-to-prose skill generation
- `workflow-skills` mode routing in both LLM providers (Anthropic + Gemini)
- MCP server routes Claude target to skill generation commands
- BDD: 9 new scenarios for annotation stripping (23 total workflow scenarios, 103 steps)
- Unit tests: annotation stripping tests + skill frontmatter tests

### Changed
- Claude target output: `{workflow}/SKILL.md` in `.claude/skills/` with native frontmatter
- Claude install paths: `~/.claude/skills/` (global), `.claude/skills/` (project)
- Interactive target prompt: "Claude Code (native skill)" replaces "Claude Code (SKILL.md workflows)"
- `BuildClaudeWorkflowSystemPrompt` replaced by `BuildWorkflowSkillSystemPrompt`
- Workflows CLI help text updated for native skill structure
- READMEs rewritten: workflows section reflects native skills for Claude target

## [1.14.0] - 2026-03-27 - Multi-target workflows (Claude Code + Antigravity)

### Added
- `--target` flag on `workflows` command: `claude` (native skill with SKILL.md frontmatter) or `antigravity` (native .md with execution annotations)
- `GenerateWorkflowFrontmatter(name, target)` — target-aware YAML frontmatter generation
- `BuildClaudeWorkflowSystemPrompt()` — LLM prompt with annotation-to-prose translation table
- Claude target output: `{workflow}/SKILL.md` in subdirectories with `user-invocable: true` frontmatter
- `Target` field in `WorkflowConfig` DTO with `ValidWorkflowTargets` validation
- `target` parameter on `generate_workflows` MCP tool
- Interactive target ecosystem prompt in workflows CLI
- Target-aware install paths: Claude → `.claude/skills/`, Antigravity → `.agent/workflows/`
- BDD: 3 new scenarios for Claude frontmatter (14 scenarios, 59 steps total)
- Unit tests: `TestGenerateWorkflowFrontmatter_Claude`, `TestGenerateWorkflowFrontmatter_UnknownClaude`

### Changed
- Workflows section in READMEs rewritten for multi-target support
- Hero/tagline updated to 4 pillars: Context, Specs, Skills, Workflows
- Go badge corrected from 1.21+ to 1.23+
- FAQ expanded with Skills vs Workflows, workflow API key, ecosystem questions

## [1.13.1] - 2026-03-27 - Rename skill category "workflow" to "conventions"

### Changed
- Renamed skill category `workflow` to `conventions` to eliminate naming ambiguity with the `workflows` command (Antigravity orchestration)
- Template directories renamed: `templates/{locale}/skills/workflow/` → `conventions/`
- Legacy mapping preserved: `--category workflow` still resolves to `conventions`

### Fixed
- Version references in READMEs and MCP server updated to 1.13.1

## [1.13.0] - 2026-03-25 - Antigravity Workflows command

### Added
- `workflows` command: generates multi-step Antigravity workflow files with execution annotations
- Workflow catalog (`internal/domain/catalog/workflow_catalog.go`) as separate bounded context from skills
- Three workflow presets: `feature-development`, `bug-fix`, `release-cycle`
- Workflow templates for both locales (en/es) with Antigravity annotations (`// turbo`, `// parallel`, `// capture`, `// if`)
- `WorkflowConfig` DTO with validation
- `DeliverStaticWorkflowsCommand` and `GenerateWorkflowsCommand` in application layer
- `BuildPersonalizedWorkflowsSystemPrompt()` and `BuildWorkflowsUserMessage()` in PromptBuilder
- `generate_workflows` MCP tool
- BDD test suite: `tests/bdd/workflow_catalog/` with 11 scenarios, 43 steps
- Interactive UX for workflows (preset, mode, locale, install scope)
- Install scopes: `global` (`~/.gemini/antigravity/global_workflows/`) and `project` (`.agent/workflows/`)

## [1.12.0] - 2026-03-25 - Testing skill category

### Added
- Testing skill category with 3 exclusive presets: `foundational`, `tdd`, `bdd`
- `foundational.template`: Kent Beck's Test Desiderata — 12 properties of good tests as trade-offs
- `tdd.template`: Part 1 (Desiderata) + Part 2 (Red-Green-Refactor, Three Laws, strategies)
- `bdd.template`: Part 1 (Desiderata) + Part 2 (Discovery/Formulation/Automation, Given/When/Then, Gherkin)
- TDD and BDD presets include foundational content as Part 1 (UX labels show "includes foundational")
- Templates for both locales (en/es)

## [1.11.0] - 2026-03-20 - Unified interactive UX and skills --install

### Added
- Unified interactive prompts: all commands (`generate`, `analyze`, `spec`, `skills`) prompt for missing flags when run in a terminal
- `--install` flag on `skills` command: `global` (agent home path) or `project` (current directory)
- Shared interactive helpers in `internal/interfaces/cli/commands/interactive.go` (charmbracelet/huh)
- `cmd.Flags().Visit()` pattern with explicit flag map to distinguish user-provided from defaults

### Fixed
- `DefaultModel()` bug: always returned Claude default even when another model was explicitly provided
- Translated all Spanish comments in CLI/DTO files to English

## [1.10.0] - 2026-03-18 - Dual-mode skills (static + personalized)

### Added
- Personalized skills mode: LLM adapts skill content to user's specific project context
- Static skills mode: delivers pre-built templates with ecosystem frontmatter (no API key needed)
- `GenerateSkillsCommand` for personalized mode, `DeliverStaticSkillsCommand` for static mode
- `BuildSkillsSystemPrompt()` and `BuildSkillsUserMessage()` in PromptBuilder
- `Mode` and `ProjectContext` fields in `SkillsConfig` DTO
- Skills generation mode (`"skills"`) in both AnthropicProvider and GeminiProvider

## [1.9.0] - 2026-03-16 - Extended interactive skill prompts

### Added
- All skills configuration options accessible via interactive menus (category, preset, mode, target, install, locale, model, project context)

## [1.8.0] - 2026-03-16 - Interactive skill categorization with catalog registry

### Added
- Declarative skill catalog (`internal/domain/catalog/skills_catalog.go`) with `SkillCategory`, `SkillOption`, `ResolvedSelection`, `SkillMetadata`
- Two categories: `architecture` (exclusive) and `workflow` (non-exclusive, renamed to `conventions` in v1.13.1)
- Interactive category → preset selection using charmbracelet/huh
- `SkillMetadata` registry for ecosystem-specific frontmatter (name, description, triggers)

## [1.7.1] - 2026-03-15 - Homebrew formula fix

### Fixed
- Switch from Homebrew cask to formula for proper macOS quarantine handling

## [1.7.0] - 2026-03-15 - Workflow skills preset and MCP knowledge tools

### Added
- Workflow skill preset: `conventional_commit.template` and `semantic_versioning.template` (both locales)
- MCP knowledge tools: `commit_guidance` and `version_guidance` — load embedded templates, no API key needed
- `loadKnowledgeTemplate()` helper for direct template content delivery

## [1.6.0] - 2026-03-15 - Output defaults and ecosystem paths

### Changed
- `generate` command now outputs to current directory by default (instead of `output/` subdirectory)
- Skills output paths: `.claude/skills/` for claude, `.agents/skills/` for codex/antigravity

## [1.5.0] - 2026-03-14 - Rebrand to Codify

### Changed
- Module renamed from `ai-context-generator` to `github.com/jorelcb/codify`
- Binary entry point: `cmd/codify/`
- All CLI references updated to `codify`

## [1.4.0] - 2026-03-14 - Embedded templates, auto-detect provider, Homebrew distribution

### Added
- GoReleaser v2 config with cross-compilation (macOS/Linux, arm64/amd64)
- GitHub Actions: CI (tests on push/PR) and release (on tag push)
- Homebrew tap distribution (`brew tap jorelcb/tap && brew install codify`)
- Embedded templates via `embed.FS` — binary works from any directory
- Auto-detect LLM provider from available API keys when `--model` is not specified

### Fixed
- Templates not found when running installed binary outside project root
- `ANTHROPIC_API_KEY environment variable is required` error when only Gemini key was set
- Version/ldflags not injected in CLI `--version` output

## [1.3.0] - 2026-03-12 - Agent Skills generation

### Added
- `skills` command: generates reusable Agent Skills (SKILL.md) based on architectural presets
- Multi-ecosystem support: `--target claude|codex|antigravity` with ecosystem-specific YAML frontmatter
- Default preset skills: DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port
- Neutral preset skills: code review, test strategy, safe refactoring, API design
- Skill templates for both locales (en/es)
- `GenerateSkillsCommand`, `SkillsConfig` DTO, `BuildSkillsSystemPrompt()`
- `generate_skills` MCP tool

## [1.2.0] - 2026-03-11 - Multi-provider LLM, MCP server, analyze command, HTTP transport

### Added
- **Gemini LLM provider**: Google Gemini API with streaming via `google.golang.org/genai` SDK v1.49.0
- **Provider factory**: `llm.NewProvider()` resolves by model prefix (`gemini-*` → Gemini, else → Anthropic)
- **MCP Server mode**: `serve` command with stdio + HTTP transport strategy
- MCP tools: `generate_context`, `generate_specs`, `analyze_project`
- `analyze` command: scans existing projects (language, framework, 20+ framework detection, config signals)
- `--with-specs` flag on `generate` and `analyze`
- `mcp-go` v0.45.0 dependency

## [1.1.0] - 2026-03-06 - Locale support, anti-hallucination, legacy cleanup

### Added
- Multi-locale support: `--locale en|es` flag
- Templates reorganized into `templates/{locale}/{preset}/`
- Language-specific `idioms.template` files (Go, JavaScript, Python)
- `<grounding_rules>` in system prompts with `[DEFINE]` markers
- `--from-file` / `-f` flag on generate command

### Changed
- System prompts rewritten in English with locale-controlled output

### Removed
- Legacy template directories, bash tests, unused domain/template layer

## [1.0.0] - 2026-02-19 - First stable release (AGENTS.md standard + spec command)

### Added
- `spec` command: generates SDD specifications from existing context
- `agents.template`: root file following AGENTS.md standard
- XML tags in system prompts

### Changed
- Output restructured: AGENTS.md at root, details in context/

## [0.2.0] - 2026-02-19 - DDD architecture with CLI

### Added
- Full DDD/Clean Architecture implementation
- CLI with Cobra (generate, list commands)
- Template system with configurable loader

## [0.1.0] - 2026-02-19 - Initial alpha

### Added
- Context file generation using Anthropic Claude API with streaming
- Per-file generation (independent API calls per output file)
- AnthropicProvider with official SDK
- Value objects with validation
- Unit tests for all components
