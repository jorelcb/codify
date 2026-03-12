# AI Context Generator - Technical Plan

## High-Level Architecture

The system uses a **Clean Architecture with Domain-Driven Design (DDD)** pattern. It is divided into four distinct layers with a strict inward-pointing dependency rule.

- **Interfaces (`interfaces/`):** Entry points like the Cobra CLI and a future MCP server. Adapts external requests into application-layer commands/DTOs.
- **Application (`application/`):** Contains use cases (commands/queries) that orchestrate domain logic. It depends only on the Domain layer.
- **Infrastructure (`infrastructure/`):** Implements external concerns like LLM API clients (Anthropic, Gemini), filesystem writers, and template loaders. It implements interfaces defined in the Domain layer.
- **Domain (`domain/`):** The core of the application, containing pure, technology-agnostic business logic, entities (`Project`), and service interfaces (`LLMProvider`, `FileWriter`).

## Technical Decisions

| Decision | Chosen Option | Discarded Alternatives | Justification |
|---|---|---|---|
| Generation Granularity | Per-file LLM generation | Single LLM call returning a large JSON object | Avoids token limits, prevents JSON parsing failures, and allows granular progress feedback. |
| Context Root File | `AGENTS.md` | Custom root files (`CLAUDE.md`, `.projectrc`) | Adheres to the Linux Foundation standard for AI agents, maximizing tool compatibility. |
| Prompting Strategy | XML tags in system prompts | Markdown-only prompts | Improves the LLM's semantic understanding and output structure, especially for Claude models. |
| LLM Integration | Multi-provider LLM factory (Claude/Gemini) | A single, hardcoded LLM provider | Allows user flexibility without changing core application logic by abstracting provider selection. |
| Specification Generation | `spec` command is context-dependent | Standalone `spec` command | Ensures generated specifications are consistent with the established architectural context. |

## Data Model

### Main Entities

- **`Project` Entity:**
  - **Description:** Represents the software project being generated or analyzed.
  - **Key Attributes:**
    - Name (e.g., `my-api`)
    - Description (e.g., "A new API")
    - `[DEFINE: Other attributes like Language, Framework, etc.]`
  - **Invariants:** `[DEFINE: Business rules, e.g., project name must be a valid directory name]`

### API Contracts

The primary interface is the CLI, not a traditional web API.

- **`generate` command:**
  - **Usage:** `ai-context-generator generate <project-name> --description "<text>"`
  - **Inputs:** Project name (argument), Description (flag).
  - **Output:** Writes generated files to `output/<project-name>/` and returns a `GenerationResult` (file paths, tokens used) to stdout.
- **MCP Server API:** `[DEFINE: Specification for MCP endpoints, requests, and responses]`

## Data Flow

The flow for the main `generate` command is as follows:

1.  **Interfaces:** The Cobra CLI command receives the project name and description from the user.
2.  **Application:** The CLI handler creates a `GenerateContextCommand` DTO and passes it to the corresponding application service handler.
3.  **Infrastructure:** The `TemplateLoader` is invoked to fetch the required structural guides from the `/templates` directory.
4.  **Domain:** A core `ProjectGenerator` service orchestrates the generation, using domain logic.
5.  **Infrastructure:** The `LLMProvider` factory selects the appropriate client (Claude/Gemini) and makes API calls for each file, guided by the templates.
6.  **Infrastructure:** The `FileWriter` adapter writes the streamed LLM response for each file to the disk.
7.  **Application:** The handler compiles a `GenerationResult` and returns it to the CLI for display to the user.

## Testing Strategy

| Level | Scope | Framework | Target Coverage |
|---|---|---|---|
| Unit | Domain entities, value objects, application services | `testify/assert` | Domain: 90%, Application: 80% |
| Integration | Infrastructure adapters (LLM clients, file writers) | `[DEFINE]` | `[DEFINE]` |
| BDD/E2E | High-level features (`generate`, `spec` commands) | `godog` | Critical use cases |

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| External LLM API Unavailability | Medium | High | Implement retry with exponential backoff for transient errors. Implement a Circuit Breaker pattern for sustained outages. |
| Inconsistent LLM Output | Medium | Medium | Use structured XML tags in prompts to guide the model. Version prompts and track performance. |
| API Key Leakage | Low | Critical | Never commit secrets. Strictly use environment variables (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY`) for credentials. |
| LLM Token Limits | Low | High | Use a per-file generation strategy instead of a single large request to stay within context window limits. |