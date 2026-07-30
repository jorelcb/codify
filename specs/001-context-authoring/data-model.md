# Data Model — Context Authoring (codify-NG)

Entidades del **dominio** (crate `codify-core`, capa `domain/`). Puras, sin I/O. Los tipos concretos de infraestructura (clientes HTTP, etc.) NO viven aquí.

## AuthoringSession (agregado raíz del loop)
Representa una sesión de authoring sobre un repo objetivo.
- **Campos**: `id`, `repo_root` (path), `mode` (`Local` | `Hybrid`), `locale` (auto-detectado | override), `state`, `provider_wiring_ref`.
- **Estados** (máquina): `Ingesting → Generating → Refining → Approved` (+ `Failed`, `Cancelled`). `Refining` es un loop que puede repetir.
- **Invariantes**:
  - En `mode = Local` el grafo de proveedores no contiene adapters de red (garantía estructural; SC-007).
  - No puede pasar a `Approved` con segmentos tentativos **sin atender** (resueltos o explícitamente diferidos) — FR-013.

## Repository (target)
- **Campos**: `root`, `is_empty` (bool), `detected_language`, `structural_signals` (deps/manifests/layout).
- **Regla**: `is_empty ⇒` la sesión entra en **modo entrevista** (no falla, no inventa).

## Reference
Documento aludido desde el material leído.
- **Campos**: `origin` (`LocalPath` | `PublicUrl`), `target`, `state` (`Resolved` | `Inaccessible` | `RequiresAuth` | `OutOfScope`), `content?`.
- **Reglas**: `RequiresAuth ⇒` fuera de v1 → se reporta, **no** se inventa contenido (FR-003/004, SC-006). `OutOfScope` = fuera del presupuesto de muestreo (declarado, no truncado en silencio).

## ContextArtifact
Cada archivo de contexto generado.
- **Campos**: `kind` (`Agents` | `Context` | `DevelopmentGuide` | `InteractionsLog` | `Idioms`), `segments: [Segment]`, `locale`.
- **Segment**: `{ text, groundedness: Grounded { sources: [ReferenceRef] } | Tentative { reason } }`.
- **Reglas**: todo `Tentative` es distinguible del `Grounded` en el render (SC-002); `Idioms` solo cuando hay lenguaje aplicable detectable.

## ChangeProposal (diff)
Unidad de cambio propuesta sobre un artefacto o el repo.
- **Campos**: `id`, `target` (artefacto/archivo), `diff` (hunks), `risk: RiskLevel`, `rationale`, `origin` (`Generation` | `Refinement`).
- **RiskLevel**: `Low` | `HighImpact` (clasificado por `RiskClassifier`; default conservador en v1).
- **Reglas** (loop curado, FR-010/012): `Low ⇒` auto-aplicable + visible/revertible; `HighImpact ⇒` requiere `ApprovalDecision` explícita **antes** de escribir. Todo cambio es auditable y revertible.

## ApprovalDecision
- **Campos**: `proposal_id`, `verdict` (`Approve` | `Edit(new_diff)` | `Reject`), `actor`, `at`.
- **Regla**: `Reject ⇒` el estado del repo no cambia; nunca se pierde trabajo humano sin verlo (FR-014/015).

## ModelConnection (config, borde de infraestructura; se referencia desde la sesión)
- **Campos**: `id`, `backend` (`Ollama` | `LlamaCpp` | `OpenAiCompat` | `Remote`), `endpoint?`, `auth` (`LocalEndpoint` | `ApiKey(ref)` | `OAuth(token_ref)`), `is_local` (bool).
- **Regla**: `is_local` determina la elegibilidad en modo `Local`.

## ModelTierRouting (política)
- **Campos**: `cheap_tier → ConnectionRef` (alta frecuencia: refinamiento, clasificación), `heavy_tier → ConnectionRef` (generación grounded).
- **Regla**: si `heavy_tier` no disponible → degradación transparente declarada (FR-018).

## AuditEvent (append-only)
- **Campos**: `at`, `kind` (`ReferenceResolved` | `ArtifactGenerated` | `ProposalMade` | `ProposalApplied` | `ProposalReverted` | `ApprovalCaptured` | `EgressBlocked`), `payload`.
- **Regla**: append-only (espeja el ethos del `INTERACTIONS_LOG`); base de la trazabilidad del loop.

## Relaciones (resumen)
`AuthoringSession 1—1 Repository`; `Session 1—* Reference`; `Session 1—* ContextArtifact`; `ContextArtifact 1—* Segment`; `Session 1—* ChangeProposal`; `ChangeProposal 0..1—1 ApprovalDecision`; `Session 1—* AuditEvent`; `Session —* ModelConnection` (vía `ModelTierRouting`).
