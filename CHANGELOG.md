# Changelog - Codify

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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