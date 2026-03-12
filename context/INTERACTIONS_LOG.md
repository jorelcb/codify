# Interaction Log - AI Context Generator

## Guiding Principles
This log documents the evolution of the project, serving as its institutional memory. It captures key decisions, problems solved, and lessons learned. New entries should follow the established format.

## Session Format

```markdown
## Session: YYYY-MM-DD - [Descriptive Title]

### Context
- **Version:** [x.y.z]
- **Goal:** [What to resolve or implement]

### Tasks Completed
1. **[Title]** - [Brief description, affected files, result]

### Architectural Decisions
- **ADR-XXX: [Title]**
  - **Decision:** [What was decided]
  - **Reason:** [Justification and trade-offs]
```

---

## Session: 2026-03-06 - v2.1.0: Output Validation and Consistency Fixes

### Context
- **Version:** 2.1.0
- **Goal:** Validate generated output for a sample project (`market-signals-service`) against its business specification and correct all inconsistencies.

### Tasks Completed
1.  **Spec Emulation:** Generated `CONSTITUTION.md`, `SPEC.md`, `PLAN.md`, `TASKS.md` for the test project.
2.  **Gap Analysis:** Systematically compared generated output against business spec, identifying 6 key gaps.
3.  **Consistency Fixes:** Propagated corrections across all 9 generated files, including unifying entity names (`SignalSource`), standardizing timestamps (`timestamp_utc`), and aligning JSON payload fields (`trading_pair`).

### Architectural Decisions

#### ADR-016: Output Validation as a Mandatory Workflow Step
-   **Decision:** Always validate LLM-generated output against the user's original business specifications.
-   **Reason:** LLMs can omit details or introduce subtle inconsistencies between generated files. Common gaps include diluted domain entities and inconsistent naming.

#### ADR-017: Dual Naming for Ports (Publisher/Consumer)
-   **Decision:** Use `SignalPublisher` for the port name in code, but maintain `SignalConsumer` as a conceptual alias in documentation.
-   **Reason:** The business spec defines the contract from the consumer's view, while the implementation is a publisher. Both names are valid in their respective contexts.

---

## Session: 2026-03-05 - v2.1.0: Multi-Locale Support and BDD Fixes

### Context
-   **Version:** 2.1.0
-   **Goal:** Implement multi-locale support (`en`/`es`), add a `--from-file` feature, and fix BDD test instability.

### Tasks Completed
1.  **Locale Support:** Added `--locale` flag. Reorganized templates into `templates/{locale}/` directories. Updated prompts with `[DEFINE]` markers for unspecified domain logic.
2.  **`--from-file` Feature:** Added `-f` flag to read project descriptions from a file. Implemented as a CLI-layer concern only.
3.  **BDD Test Fixes:** Corrected repository save logic to handle duplicate names. Ensured unique names in test steps to prevent state leakage between scenarios. All 14 scenarios (72 steps) now pass reliably.

### Architectural Decisions

#### ADR-015: `--from-file` as a CLI-Only Feature
-   **Decision:** The logic for reading a description from a file is confined to the `interfaces/cli` layer. The application core continues to receive a simple string.
-   **Reason:** The application layer is agnostic to the source of the input string, upholding the separation of concerns.

---

## Session: 2026-02-19 - v1.1.0: Restructure to AGENTS.md Standard & Add Spec Command

### Context
-   **Version:** 1.1.0
-   **Goal:** Align generated output with industry standards (AGENTS.md) and introduce a `spec` command for Spec-Driven Development (SDD).

### Tasks Completed
1.  **Output Restructuring:** Replaced the old file structure (`PROMPT.md`, `SCAFFOLDING.md`) with the `AGENTS.md` standard. `AGENTS.md` is now the root file, with details in the `context/` directory.
2.  **Spec Command:** Implemented `spec <name> --from-context <path>` to generate `specs/` from existing context, enabling the AI SDD workflow.
3.  **Prompt Engineering:** Switched system prompts from Markdown to XML tags (`<role>`, `<task>`) to improve LLM semantic precision, based on Claude's training data.

### Architectural Decisions

#### ADR-011: Adopt AGENTS.md as the Root File Standard
-   **Decision:** Use `AGENTS.md` as the main entry point file in the project root.
-   **Reason:** It is a Linux Foundation standard with wide adoption, providing a common interface for AI agents.

#### ADR-012: Eliminate Agent Persona Files (PROMPT.md)
-   **Decision:** Removed `PROMPT.md`, which defined an "agent persona". Factual coding standards were moved into `AGENTS.md`.
-   **Reason:** Industry standards focus on providing factual context, not defining a personality for the agent, which is an anti-pattern.

#### ADR-013: Spec Generation is Context-Dependent
-   **Decision:** The `spec` command *must* operate on a pre-existing context. It cannot be run standalone.
-   **Reason:** To ensure coherence, implementation specifications must be derived directly from the established architectural blueprint.

#### ADR-014: Use XML Tags in System Prompts
-   **Decision:** Structure system prompts using XML tags. User-provided descriptions remain as-is.
-   **Reason:** Significantly improves Claude's ability to parse the prompt's structure, leading to higher-quality, more consistent output.

---

## Session: 2026-02-19 - v1.0.0: Initial Release

### Context
-   **Version:** 1.0.0
-   **Goal:** Deliver the core functionality: generate context files from a user description via the Anthropic Claude API.

### Architectural Decisions

#### Per-File Generation vs. Single JSON Blob
-   **Decision:** Make separate API calls for each file, with each call returning pure Markdown.
-   **Reason:** A single large JSON response containing all files was prone to truncation and complex parsing. This approach allows for granular progress feedback and richer content per file.

#### Streaming API by Default
-   **Decision:** Use the streaming API for all LLM calls.
-   **Reason:** Prevents gateway timeouts for complex generations and provides a better user experience by showing real-time progress.

---
**CRITICAL:** This log serves as the project's memory. Review recent ADRs before making significant architectural changes to maintain consistency.