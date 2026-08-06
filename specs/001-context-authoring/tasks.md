---
description: "Task list — Context Authoring (codify-NG)"
---

# Tasks: Context Authoring — de repositorio a contexto vivo (codify-NG)

**Input**: Design documents from `specs/001-context-authoring/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, constitution v1.0.0

**Tests**: INCLUIDOS y **test-first** — la constitución (Principio II) manda TDD Red-Green-Refactor + BDD outside-in como **[NON-NEGOTIABLE]**. Cada test debe escribirse y **fallar** antes de su implementación.

**Organization**: por user story (US1/US2/US3), cada una implementable y testeable de forma independiente.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: paralelizable (archivos distintos, sin dependencias pendientes)
- **[Story]**: US1/US2/US3 (solo en fases de user story)
- Ruta de archivo exacta en cada tarea

## Path Conventions
Workspace Cargo (plan.md): `crates/codify-core/` (biblioteca hexagonal) + `crates/codify-app/` (piel Tauri) + `tests/` en la raíz.

---

## 🔗 Dependencias con el spec `002-authoring-experience`

Este spec **no vive aislado**: parte de su trabajo desbloquea al otro, y parte del suyo ya
resolvió tareas de aquí. Se documenta en ambos lados para que planificar uno solo no lleve a
conclusiones falsas.

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).
> Este archivo sigue siendo la fuente de verdad de la **ejecución tarea a tarea**; los issues lo
> son de las **dependencias entre épicas**, porque allí se ven desde los dos lados a la vez.
> Épicas de este spec: **US2 → #4** · **US3 → #7**.

### Lo que este spec BLOQUEA

**La US2 de aquí (T031–T040 y T055–T057) bloquea la US2 de `002`** (revisar diffs, sus
T039–T044): no se puede construir la interfaz de revisión de propuestas mientras no exista el
loop de refinamiento que las genera.

Evidencia en el código: no existen `application/refine.rs` ni `infrastructure/diff/`, el
servicio no expone `submit_message`/`pending_proposals`/`decide`, y la piel lleva dos
marcadores que **fallan a propósito** — `UnavailablePrompter` y `NoDiffYet`— para que la
ausencia sea visible en vez de silenciosa.

### Lo que `002` ya nos entregó

La Fase 2 de `002` añadió al núcleo tres capacidades que **estas tareas pueden dar por hechas**:

| Capacidad | Qué habilita aquí |
|---|---|
| Port `ArtifactWriter` con `read_existing` | **US3 (T042)**: detectar artefactos previos ya no requiere construir nada; el port existe y se diseñó justamente para esto |
| Los artefactos se **escriben a disco** | Antes se quedaban en memoria. La US3 opera sobre archivos reales |
| `start_session` **ya no bloquea** y hay `cancel_session`/`join_session` | Cualquier tarea que asuma la semántica bloqueante anterior está desactualizada |

### ⚠️ Tareas de este spec que `002` ya resolvió

Verificado contra el código, no contra la documentación. **Revisar antes de retomarlas**:

- **T030** (UI mínima: stream de actividad + render grounded/tentative) — **superada** por la
  US1 de `002` (su T025–T032 y T051–T054). La interfaz existe y va bastante más allá.
- **T047** (comando `set_locale`) — **ya existe** en `codify-app/src/commands.rs`.
- **T048** — la mitad de "evento `egress.blocked` en UI" **ya está** (`002`-T051); la
  *persistencia del `AuditSink`* sigue pendiente.
- **T049** (modo entrevista para repo vacío) — **hecho**: el núcleo lo devuelve y la interfaz
  lo presenta.
- **T050 / T052** (validación del quickstart y verificación de fitness functions en CI) —
  **solapan** con T049/T050 de `002`. Conviene ejecutarlos una sola vez, no dos.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Inicializar workspace y toolchain.

- [X] T001 Crear workspace Cargo con miembros `codify-core` y `codify-app` en `Cargo.toml` (raíz) y esqueleto de crates
- [X] T002 [P] Declarar dependencias del core (tokio, serde/serde_json, reqwest, similar, oauth2, keyring, async-trait, thiserror) en `crates/codify-core/Cargo.toml`
- [X] T003 [P] Inicializar Tauri v2 (tauri, tauri-build) + scaffold del frontend web en `crates/codify-app/` y `crates/codify-app/ui/`
- [X] T004 [P] Configurar `rustfmt.toml` y clippy (`deny(warnings)`) en la raíz del repo
- [X] T005 [P] Esqueleto de CI (fmt + clippy + test) en `.github/workflows/ci.yml`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Dominio puro, ports, composition root y las fitness functions de la constitución. **⚠️ Ninguna user story puede empezar hasta cerrar esta fase.**

- [X] T006 [P] Layout de módulos `domain/`, `application/`, `infrastructure/` en `crates/codify-core/src/lib.rs`
- [X] T007 [P] `AuthoringSession` (agregado) + máquina de estados Ingesting→Generating→Refining→Approved en `crates/codify-core/src/domain/session.rs`
- [X] T008 [P] `Repository` + `Reference` (+ estado Resolved/Inaccessible/RequiresAuth/OutOfScope) en `crates/codify-core/src/domain/reference.rs`
- [X] T009 [P] `ContextArtifact` + `Segment` (grounded/tentative) en `crates/codify-core/src/domain/context.rs`
- [X] T010 [P] `ChangeProposal` + `RiskLevel` + `ApprovalDecision` en `crates/codify-core/src/domain/change.rs`
- [X] T011 [P] `AuditEvent` (append-only) en `crates/codify-core/src/domain/audit.rs`
- [X] T012 Definir ports (driving `AuthoringService`; driven `ModelProvider`/`RepoNavigator`/`ReferenceResolver`/`DiffEngine`/`RiskClassifier`/`Prompter`/`AuditSink`/`LocaleDetector`) en `crates/codify-core/src/domain/ports/` y `crates/codify-core/src/application/ports/` — firmas con tipos de dominio únicamente (depende de T007-T011)
- [X] T013 [P] Fakes in-memory de todos los driven ports en `crates/codify-core/tests/fakes/mod.rs`
- [X] T014 Composition root + `ProviderRegistry` con cableado **solo-local** (cero-egress estructural) en `crates/codify-core/src/infrastructure/composition.rs` (depende de T012)
- [X] T015 [P] **Fitness function de dependencias** (aserta: `domain`/`application` con cero imports de `infrastructure`) en `.github/workflows/ci.yml` + `crates/codify-core/tests/arch_deps.rs` (constitución I, [NN])
- [X] T016 [P] **Harness de cero-egress** (falla si hay salida a host no-local en modo local) en `tests/integration/egress_guard.rs` (constitución proyecto, [NN])

**Checkpoint**: Fundación lista — las user stories pueden comenzar.

---

## Phase 3: User Story 1 - Contexto grounded desde el repo y sus referencias (Priority: P1) 🎯 MVP

**Goal**: Leer repo + seguir referencias (docs + estructura + muestreo selectivo de código; local + URLs públicas) y generar los artefactos de contexto fieles a la SSOT, marcando lo tentativo.

**Independent Test**: Apuntar al fixture (README→SPEC "sin broker/event-sourced") en modo local; verificar contexto grounded, referencias declaradas, y cero-egress.

### Tests for User Story 1 (test-first) ⚠️
- [X] T017 [P] [US1] Escenario de aceptación BDD (quickstart S1) en `tests/integration/us1_grounded.rs`
- [X] T018 [P] [US1] Contract test `RepoNavigator` (real fs + fake) en `tests/contract/repo_navigator.rs`
- [X] T019 [P] [US1] Contract test `ReferenceResolver` (local + URL pública; RequiresAuth reportado, no inventado) en `tests/contract/reference_resolver.rs`
- [X] T020 [P] [US1] Contract test `ModelProvider` local (openai-compat/Ollama; `is_local`) en `tests/contract/model_provider.rs`
- [X] T021 [P] [US1] Unit: presupuesto de ingesta + declarar-omitido + detección de idioma en `crates/codify-core/tests/ingest_unit.rs`
- [X] T053 [P] [US1] Escenario: dos fuentes contradictorias ⇒ el sistema señala la contradicción (no elige en silencio) en `tests/integration/us1_contradiction.rs` [FR-008]

### Implementation for User Story 1
- [X] T022 [P] [US1] Adapter `RepoNavigator` (fs) en `crates/codify-core/src/infrastructure/repo/navigator.rs`
- [X] T023 [P] [US1] Adapter `ReferenceResolver` (fs local + http público vía reqwest; estados de no-resuelto) en `crates/codify-core/src/infrastructure/repo/reference_resolver.rs`
- [X] T024 [P] [US1] Adapters `ModelProvider` locales (openai-compat cubre Ollama + llama.cpp-server) en `crates/codify-core/src/infrastructure/providers/local.rs`
- [X] T025 [P] [US1] Adapter `LocaleDetector` en `crates/codify-core/src/infrastructure/repo/locale.rs`
- [X] T026 [US1] Tools del agente (list_repo/read_file/fetch_url/note_unresolved/finalize) + estrategia de muestreo en `crates/codify-core/src/application/ingest.rs` (depende de T022-T024)
- [X] T027 [US1] Loop de authoring: pase ingest+generate → `ContextArtifact` con segmentos grounded/tentative en `crates/codify-core/src/application/authoring_loop.rs` (depende de T026, T012)
- [X] T054 [US1] Detección y señalización de contradicción entre fuentes (evento/segmento de contradicción) en `crates/codify-core/src/application/authoring_loop.rs` (depende de T027) [FR-008]
- [X] T028 [US1] `AuthoringService::start_session`/`session_state` (resultado US1) en `crates/codify-core/src/application/service.rs`
- [X] T029 [US1] Comandos Tauri `start_session`/`session_state` + eventos `agent.activity`/`reference.unresolved` en `crates/codify-app/src/commands.rs`
- [X] T030 [US1] ~~UI mínima: iniciar sesión, stream de actividad, render de artefactos~~ — **superada por la US1 de `002`** (su T025-T032, T051-T054), que entrega bastante más

**Checkpoint**: US1 funcional e independiente — **MVP** entregable.

---

## Phase 4: User Story 2 - Refinamiento conversacional con diffs curados (Priority: P2)

**Goal**: Refinar el contexto en loop conversacional; auto-aplicar bajo riesgo, aprobar explícito el alto impacto; todo revertible.

**Independent Test**: Desde un contexto con hueco/supuesto incorrecto, refinar y aprobar/rechazar diffs (quickstart S2).

### Tests for User Story 2 (test-first) ⚠️
- [ ] T031 [P] [US2] Escenario de aceptación BDD (quickstart S2) en `tests/integration/us2_refine.rs`
- [ ] T032 [P] [US2] Contract test `DiffEngine` (property: apply∘revert = identidad) en `tests/contract/diff_engine.rs`
- [ ] T033 [P] [US2] Contract test `RiskClassifier` (default conservador: no-trivial ⇒ HighImpact) en `tests/contract/risk_classifier.rs`
- [ ] T034 [P] [US2] Contract test `Prompter` (solo HighImpact bloquea) en `tests/contract/prompter.rs`
- [ ] T055 [P] [US2] Escenario: al corregir un supuesto, el diff ajusta el andamiaje dependiente (nombres/secciones), no solo el marcador, en `tests/integration/us2_scaffolding.rs` [FR-011]

### Implementation for User Story 2
- [ ] T035 [P] [US2] Adapter `DiffEngine` (crate `similar`) make/apply/revert en `crates/codify-core/src/infrastructure/diff/engine.rs`
- [ ] T036 [P] [US2] `RiskClassifier` conservador v1 en `crates/codify-core/src/infrastructure/diff/risk.rs`
- [ ] T037 [US2] Loop curado en `refine.rs`: submit_message → propose_change → classify → auto-aplica Low / requiere aprobación HighImpact; tool `ask_user` en `crates/codify-core/src/application/refine.rs` (depende de T035, T036)
- [ ] T038 [US2] Manejo de `ApprovalDecision` (approve/edit/reject; reject ⇒ no escribe) en `crates/codify-core/src/application/refine.rs`
- [ ] T056 [US2] Propagación de la corrección al andamiaje dependiente (secciones/nombres afectados) en `crates/codify-core/src/application/refine.rs` (depende de T037) [FR-011]
- [ ] T057 [US2] Aserto de cierre: transición a Approved deja 0 marcadores pendientes (resueltos o diferidos explícitos) en `tests/integration/us2_refine.rs` [FR-013]
- [ ] T039 [US2] `AuthoringService::submit_message`/`pending_proposals`/`decide` en `crates/codify-core/src/application/service.rs`
- [ ] T040 [US2] Comandos Tauri `submit_message`/`pending_proposals`/`decide` + eventos `proposal.new`/`agent.token`; UI de diff + approve/edit/reject en `crates/codify-app/src/commands.rs` y `crates/codify-app/ui/`

**Checkpoint**: US1 + US2 funcionan de forma independiente.

---

## Phase 5: User Story 3 - Repo con contexto previo: actualización sin sobrescribir (Priority: P3)

**Goal**: Re-ejecutar sobre un repo con contexto existente → proponer actualización como diff, cero sobrescrituras silenciosas.

**Independent Test**: Re-ejecutar sobre repo con `AGENTS.md` existente → diff de actualización + aprobación (quickstart S3).

### Tests for User Story 3 (test-first) ⚠️
- [ ] T041 [P] [US3] Escenario de aceptación BDD (quickstart S3: no-clobber) en `tests/integration/us3_update.rs`

### Implementation for User Story 3
- [ ] T042 [US3] Detectar artefactos de contexto existentes en `start_session` y derivar al flujo de actualización en `crates/codify-core/src/application/authoring_loop.rs` — **usa `ArtifactWriter::read_existing`, que ya existe** (entregado por `002`-Fase 2)
- [ ] T043 [US3] Producir `ChangeProposal` de actualización preservando contenido humano (reusa `DiffEngine`/aprobación de US2) en `crates/codify-core/src/application/refine.rs`
- [ ] T044 [US3] Cablear el camino de contexto-previo en `start_session` + mensajería de UI en `crates/codify-app/`

**Checkpoint**: Las tres user stories funcionan de forma independiente.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Capacidades transversales y hardening (no bloquean el MVP).

- [ ] T045 [P] `ModelProvider` remoto + OAuth device-flow + keyring (`connect_provider`/`list_connections`) en `crates/codify-core/src/infrastructure/providers/remote.rs` y `crates/codify-core/src/infrastructure/secrets/keyring.rs`
- [ ] T046 [P] Routing de tiers (cheap/heavy) + degradación transparente (FR-018) en `crates/codify-core/src/application/routing.rs`
- [X] T047 [P] Comando `set_locale` (override de idioma, FR-019) en `crates/codify-app/src/commands.rs` — **ya existe** (entregado con T029)
- [ ] T048 [P] Persistencia de `AuditSink` en `crates/codify-core/src/infrastructure/audit/sink.rs` — el *evento `egress.blocked` en UI* **ya lo entregó `002`-T051**; queda solo la persistencia
- [X] T049 Modo entrevista para repo vacío (edge case) — **hecho**: el núcleo lo devuelve (`interview_mode`) y la interfaz de `002` lo presenta
- [ ] T050 [P] Ejecutar validación quickstart.md (S1–S5) end-to-end — **solapa con `002`-T049**: ejecutar una sola vez
- [ ] T051 [P] Empaquetado Tauri (macOS/Linux/Windows) + docs en `docs/`
- [ ] T052 Verificar en CI las fitness functions verdes — **solapa con `002`-T050**: ejecutar una sola vez

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (P1)**: sin dependencias.
- **Foundational (P2)**: depende de Setup — **BLOQUEA** todas las user stories.
- **User Stories (P3-P5)**: dependen de Foundational. Luego pueden ir en paralelo o en orden de prioridad P1→P2→P3.
- **Polish (P6)**: depende de las user stories deseadas.

### User Story Dependencies
- **US1 (P1)**: tras Foundational — sin dependencias de otras stories. Es el MVP.
- **US2 (P2)**: tras Foundational — introduce `DiffEngine`/`RiskClassifier`/aprobación; independientemente testeable.
- **US3 (P3)**: tras Foundational — **reusa** `DiffEngine`/aprobación de US2 para el diff de actualización; testeable de forma independiente con su propio fixture.

### Within Each User Story
- Los tests se escriben y **fallan** antes de implementar (constitución II).
- Domain/entidades → adapters (infra) → application (loop/servicio) → comandos Tauri/UI.

### Parallel Opportunities
- Setup: T002-T005 en paralelo.
- Foundational: T007-T011 (entidades) y T013/T015/T016 en paralelo; T012 tras entidades; T014 tras T012.
- US1: tests T017-T021 en paralelo; adapters T022-T025 en paralelo; luego T026→T027→T028→T029→T030.
- US2: tests T031-T034 en paralelo; T035/T036 en paralelo; luego T037→T038→T039→T040.
- Con equipo: US1/US2/US3 en paralelo tras Foundational (US3 coordina con US2 por reuso del diff).

---

## Parallel Example: User Story 1

```bash
# Tests de US1 en paralelo (deben fallar primero):
Task: "Contract test RepoNavigator en tests/contract/repo_navigator.rs"
Task: "Contract test ReferenceResolver en tests/contract/reference_resolver.rs"
Task: "Contract test ModelProvider local en tests/contract/model_provider.rs"

# Adapters de US1 en paralelo:
Task: "RepoNavigator fs en crates/codify-core/src/infrastructure/repo/navigator.rs"
Task: "ReferenceResolver en crates/codify-core/src/infrastructure/repo/reference_resolver.rs"
Task: "ModelProvider local en crates/codify-core/src/infrastructure/providers/local.rs"
```

---

## Implementation Strategy

### MVP First (solo US1)
1. Phase 1 Setup → 2. Phase 2 Foundational (crítica) → 3. Phase 3 US1 → 4. **VALIDAR** US1 independiente (quickstart S1, cero-egress) → 5. Demo.

### Incremental Delivery
Setup+Foundational → US1 (MVP) → US2 → US3, cada una testeada e integrada sin romper las previas. Polish (P6) al final o intercalado por necesidad (OAuth remoto solo cuando se requiera el tier frontier).

---

## Notes
- [P] = archivos distintos, sin dependencias pendientes.
- Verificar que cada test falla antes de implementar (constitución II, [NN]).
- Commit atómico por tarea/grupo lógico (Conventional Commits; **sin atribución IA**; constitución III, [NN]).
- Detenerse en cada checkpoint para validar la story de forma independiente.
- Evitar: tareas vagas, conflictos en el mismo archivo, dependencias cross-story que rompan la independencia.
- BDD Discovery (Three Amigos / Example Mapping) se front-cargó en `spec.md` + `quickstart.md` (los ejemplos ya acordados). En build individual se acepta conscientemente arrancar en Formulation; si se forma equipo, correr Example Mapping por story antes de sus tests (constitución II).
