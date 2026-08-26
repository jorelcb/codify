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
Workspace Cargo (plan.md): `crates/codify-core/` (biblioteca hexagonal) + `crates/codify-app/` (piel Tauri).

> **Los tests de integración van en la RAÍZ de `crates/<crate>/tests/`**, en archivos planos
> (`contract_*.rs`, `us1_*.rs`). No en subdirectorios: **Cargo no descubre tests anidados**, así
> que `tests/contract/foo.rs` no se ejecutaría nunca — y un test que no corre es peor que
> ninguno, porque aparenta cobertura. Los tests unitarios viven en su propio archivo fuente,
> en un `mod tests`.

---

## 🔗 Dependencias con el spec `002-authoring-experience`

Este spec **no vive aislado**: parte de su trabajo desbloquea al otro, y parte del suyo ya
resolvió tareas de aquí. Se documenta en ambos lados para que planificar uno solo no lleve a
conclusiones falsas.

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).
> Este archivo sigue siendo la fuente de verdad de la **ejecución tarea a tarea**; los issues lo
> son de las **dependencias entre épicas**, porque allí se ven desde los dos lados a la vez.
> Épicas de este spec: **US2 → #4** · **US3 → #7**.

### ✅ Lo que este spec bloqueaba — ya no

**La US2 de aquí está entregada**, y con ella la **US2 de `002` (issue #5) ya se cerró también**.
Queda desbloqueada la **US3 de aquí (issue #7)**, que reusa su `DiffEngine` y su flujo de
aprobación.

Lo que ahora existe y aquellas tareas pueden dar por hecho:

| Pieza | Dónde |
|---|---|
| Motor de diffs con `apply∘revert = identidad` | `infrastructure/diff/engine.rs` |
| Clasificador de riesgo conservador | `infrastructure/diff/risk.rs` |
| Loop curado de refinamiento | `application/refine.rs` |
| `submit_message` / `pending_proposals` / `decide` | `application/service.rs` |

`UnavailablePrompter` **ya no existe**: la T040 de `002` lo reemplazó por `WindowPrompter`, que
emite `proposal.new` a la ventana y espera la decisión en un canal. El circuito está completo.

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
- [X] T016 [P] **Harness de cero-egress** (falla si hay salida a host no-local en modo local) en `crates/codify-core/tests/egress_guard.rs` (constitución proyecto, [NN])

**Checkpoint**: Fundación lista — las user stories pueden comenzar.

---

## Phase 3: User Story 1 - Contexto grounded desde el repo y sus referencias (Priority: P1) 🎯 MVP

**Goal**: Leer repo + seguir referencias (docs + estructura + muestreo selectivo de código; local + URLs públicas) y generar los artefactos de contexto fieles a la SSOT, marcando lo tentativo.

**Independent Test**: Apuntar al fixture (README→SPEC "sin broker/event-sourced") en modo local; verificar contexto grounded, referencias declaradas, y cero-egress.

### Tests for User Story 1 (test-first) ⚠️
- [X] T017 [P] [US1] Escenario de aceptación BDD (quickstart S1) en `crates/codify-core/tests/us1_grounded.rs`
- [X] T018 [P] [US1] Contract test `RepoNavigator` (real fs + fake) en `crates/codify-core/tests/contract_ports.rs` (sección `repo_navigator`)
- [X] T019 [P] [US1] Contract test `ReferenceResolver` (local + URL pública; RequiresAuth reportado, no inventado) en `crates/codify-core/tests/contract_ports.rs` (sección `reference_resolver`)
- [X] T020 [P] [US1] Contract test `ModelProvider` local (openai-compat/Ollama; `is_local`) en `crates/codify-core/tests/contract_ports.rs` (sección `model_provider`)
- [X] T021 [P] [US1] Unit: presupuesto de ingesta + declarar-omitido + detección de idioma en `crates/codify-core/src/application/ingest.rs` (módulo `tests`)
- [X] T053 [P] [US1] Escenario: dos fuentes contradictorias ⇒ el sistema señala la contradicción (no elige en silencio) en `crates/codify-core/tests/us1_contradiction.rs` [FR-008]

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
- [X] T031 [P] [US2] Escenario de aceptación BDD (quickstart S2) en `crates/codify-core/tests/us2_refine.rs`
- [X] T032 [P] [US2] Contract test `DiffEngine` (property: apply∘revert = identidad) en `crates/codify-core/tests/contract_diff_engine.rs` — el fake era **más permisivo** que el adapter real (aceptaba aplicar sobre cualquier texto); la suite compartida lo destapó y se subió al contrato
- [X] T033 [P] [US2] Contract test `RiskClassifier` (default conservador: no-trivial ⇒ HighImpact) en `crates/codify-core/tests/contract_risk_classifier.rs`
- [X] T034 [P] [US2] Contract test `Prompter` (solo HighImpact bloquea) en `crates/codify-core/tests/contract_prompter.rs` — nota: la regla «solo HighImpact bloquea» no la cumple el `Prompter` sino el loop, y se asserta en `us2_refine.rs`, donde vive
- [X] T055 [P] [US2] Escenario: al corregir un supuesto, el diff ajusta el andamiaje dependiente (nombres/secciones), no solo el marcador, en `crates/codify-core/tests/us2_scaffolding.rs` [FR-011]

### Implementation for User Story 2
- [X] T035 [P] [US2] Adapter `DiffEngine` (crate `similar`) make/apply/revert en `crates/codify-core/src/infrastructure/diff/engine.rs`
- [X] T036 [P] [US2] `RiskClassifier` conservador v1 en `crates/codify-core/src/infrastructure/diff/risk.rs`
- [X] T037 [US2] Loop curado en `refine.rs`: submit_message → propose_change → classify → auto-aplica Low / requiere aprobación HighImpact; tool `ask_user` en `crates/codify-core/src/application/refine.rs` (depende de T035, T036)
- [X] T038 [US2] Manejo de `ApprovalDecision` (approve/edit/reject; reject ⇒ no escribe) en `crates/codify-core/src/application/refine.rs`
- [X] T056 [US2] Propagación de la corrección al andamiaje dependiente (secciones/nombres afectados) en `crates/codify-core/src/application/refine.rs` (depende de T037) [FR-011]
- [X] T057 [US2] Aserto de cierre: transición a Approved deja 0 marcadores pendientes (resueltos o diferidos explícitos) en `crates/codify-core/tests/us2_refine.rs` [FR-013]
- [X] T039 [US2] `AuthoringService::submit_message`/`pending_proposals`/`decide` en `crates/codify-core/src/application/service.rs`
- [→] T040 [US2] ~~Comandos Tauri + UI de diff/approve/edit/reject~~ — **remitida a la Fase 5 de `002` (issue #5)**. Describía la misma piel que las T039–T044 de allí: comandos, `Prompter` real, bloque de diff, captura de decisión, contador de pendientes y conversación. Aquí se queda **solo el núcleo**; la interfaz es del spec que la tiene por objeto. El único detalle que `002` no nombraba, `submit_message`, se anotó en su T039

**Checkpoint**: US1 + US2 funcionan de forma independiente.

---

## Phase 5: User Story 3 - Repo con contexto previo: actualización sin sobrescribir (Priority: P3) ✅ COMPLETA

**Goal**: Re-ejecutar sobre un repo con contexto existente → proponer actualización como diff, cero sobrescrituras silenciosas.

**Independent Test**: Re-ejecutar sobre repo con `AGENTS.md` existente → diff de actualización + aprobación (quickstart S3).

### Tests for User Story 3 (test-first) ⚠️
- [X] T041 [P] [US3] Escenario de aceptación BDD (quickstart S3: no-clobber) en `crates/codify-core/tests/us3_update.rs`

### Implementation for User Story 3
- [X] T042 [US3] Detectar artefactos de contexto existentes en `start_session` y derivar al flujo de actualización en `crates/codify-core/src/application/authoring_loop.rs` — **usa `ArtifactWriter::read_existing`, que ya existe** (entregado por `002`-Fase 2) — el flujo vive en `generate()`: antes de escribir consulta `read_existing`; si hay contexto previo distinto, propone en vez de pisar
- [X] T043 [US3] Producir `ChangeProposal` de actualización preservando contenido humano (reusa `DiffEngine`/aprobación de US2) en `crates/codify-core/src/application/refine.rs` — **sin fusión automática**, y es deliberado: el sistema no puede saber qué párrafo escribió una persona y cuál generó él. Adivinarlo produciría pérdidas silenciosas justo donde más duelen; así que enseña el diff y pregunta, y si el usuario edita se escribe lo suyo
- [X] T044 [US3] Cablear el camino de contexto-previo en `start_session` + mensajería de UI en `crates/codify-app/` — **no hizo falta interfaz nueva**: la propuesta de actualización viaja por el mismo `proposal.new` → panel de decisión que montó la US2 de `002`. Verificado en navegador

**Checkpoint**: Las tres user stories funcionan de forma independiente.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Capacidades transversales y hardening (no bloquean el MVP).

> **Una casilla vacía no dice lo mismo en todos los casos.** Aquí conviven tres cosas distintas,
> y confundirlas es lo que hace que un spec envejezca mal:
>
> - **defecto** — falta trabajo, nada lo bloquea. Se puede tomar hoy.
> - **decisión sin tomar** — el trabajo está claro, lo que falta es saber qué se quiere. Pedirlo
>   sin decidir produciría la respuesta equivocada bien implementada.
> - **remitida** (`[→]`) — se aborda en otro spec, con enlace. **No** es lo mismo que descartada:
>   sigue dentro del alcance y su issue lo dice.
>
> **Corrección del 2026-08-25**: T045 y T048 estuvieron clasificadas aquí como «decisión sin
> tomar» mientras sostenían requisitos **MUST** sin cumplir. Eso las sacó de facto del alcance
> sin que nadie lo decidiera, y los resúmenes pasaron a decir que el proyecto solo tenía un
> hallazgo menor pendiente. Ambas están **dentro de v1** y remitidas a specs propios.
> - **fuera de nuestro alcance** — depende de algo que el equipo no tiene, y ninguna cantidad de
>   trabajo lo desbloquea.
>
> **T051 y T053 son operativas, no derivadas de un requisito.** `001` especifica cómo se autora
> el contexto, no cómo se distribuye la aplicación. Que no tengan FR detrás es correcto, y se
> dice aquí para que la próxima revisión no lo cuente como hueco.

- [→] T045 [P] `ModelProvider` remoto + OAuth device-flow + keyring — **remitida al spec `003`** ([#42](https://github.com/jorelcb/codify/issues/42)), **dentro de v1**. Junto con T046 cubre FR-016 y FR-017, dos **MUST** que este spec promete y el código no cumple. Sale de aquí porque toca la garantía de cero-egress —`Mode::Hybrid` se vuelve real— y eso merece su propio ciclo, no una novena fase de un spec del que ya hemos entrado y salido demasiadas veces
- [X] T046 [P] Degradación **declarada** entre tiers (`001`-FR-018) — **hecho**. Ojo: esto cierra FR-018, **no FR-017**: enrutar y declarar la degradación funciona, pero hoy solo hay un tier real entre el que elegir. FR-017 quedó **cumplido** en el spec `003` (PR #46): con dos tiers conectados, el reparto distingue de verdad. `ProviderRegistry::pick` devuelve un `Routed { provider, degraded_from }` en vez del proveedor a secas, para que la degradación **no se pueda ignorar por descuido**: quien enruta recibe el dato en la mano. Se audita (`AuditKind::TierDegraded`), viaja al snapshot (`tier_degraded`) y la interfaz lo declara con `provider.tier_degraded`. Tres tests nuevos en `crates/codify-core/tests/us1_tier_degradation.rs`
- [X] T047 [P] Comando `set_locale` (override de idioma, FR-019) en `crates/codify-app/src/commands.rs` — **ya existe** (entregado con T029)
- [→] T048 [P] Persistencia de `AuditSink` — **remitida al spec `004`** ([#43](https://github.com/jorelcb/codify/issues/43)), **dentro de v1**. El *evento `egress.blocked` en UI* ya lo entregó `002`-T051. Sale de *Polish* porque no tenía FR que la sostuviera, y su pregunta de fondo —demostrar meses después que un cambio se aplicó— es la puerta al ciclo de vida que el proyecto declara como dirección
- [X] T049 Modo entrevista para repo vacío (edge case) — **hecho**: el núcleo lo devuelve (`interview_mode`) y la interfaz de `002` lo presenta
- [~] T050 [P] Ejecutar validación quickstart.md (S1–S5) end-to-end — **pasada ejecutada el 2026-08-23** con `llama.cpp server` + `qwen2.5-7b-q4km`, tres corridas (arnés `live_backend.rs`). S1, S2, S3 y S5 ejercitados; **cinco hallazgos** documentados en `quickstart.md`. **F-1 (#23) está cerrado** por la Fase 7, verificado en una segunda pasada con `Qwen2.5-32B` y tres corridas limpias; siguen abiertos #24 y #25, más **#34**, que salió de esa segunda pasada. **Queda S4 sin ejercitar**: el agente nunca intenta la URL privada, y decidir si debería es la pregunta abierta de #25
- [ ] T051 [P] Empaquetado Tauri firmado (macOS/Linux/Windows) + docs en `docs/` — **fuera de nuestro alcance y de las últimas del proyecto**. `tauri.conf.json` ya trae `bundle.targets: "all"` e iconos; falta una matriz de CI en los tres sistemas, y sobre todo **certificados que no tenemos**: Apple Developer para firmar y notarizar, y firma de Windows. Sin ellos el binario se distribuye igual, pero Gatekeeper y SmartScreen avisan al usuario. Es una compra, no trabajo de código
- [ ] T053 [P] Distribución por Homebrew — **va antes que T051**. Se reutiliza **al 100% el nombre `codify`**, reapuntando la fórmula existente de `jorelcb/homebrew-tap`, que hoy sirve el binario Go de `codify-og`. Es la vía real de distribución mientras no haya certificados, y un `.app` instalado por cask no arrastra la misma fricción que un `.dmg` descargado del navegador — conviene confirmarlo al llegar
- [X] T052 Verificar en CI las fitness functions verdes — ejecutado una sola vez junto a `002`-T050: verificado sobre el commit `68ed5df` (run 31377803807, `completed/success`); el paso «Tests (incluye fitness functions de arquitectura y cero-egress)» pasó, y `cargo test --workspace` recorre `arch_deps.rs` y `egress_guard.rs` por vivir en la raíz de `tests/`. **Limitación declarada**: los logs del run no se pudieron leer desde esta máquina (**token de `gh` inválido**, no falta de permisos), así que la evidencia es la conclusión del paso, no la lectura línea a línea

---

---

## Phase 7: Verificación de procedencia — FR-006a/b/c (issue #23)

**Goal**: que `Grounded` signifique algo comprobable. Hoy el modelo declara la fuente y el
núcleo le cree; tras esto, una afirmación solo es `Grounded` si su **cita textual aparece en el
material leído**.

**Independent Test**: sembrar una respuesta del modelo que cite una fuente real atribuyéndole
algo que no dice → el segmento **se degrada a tentativo** con el motivo, en vez de presentarse
como verificado.

> **Origen**: hallazgo F-1 de la pasada con modelo real (2026-08-23). El sistema registró
> `[PRD vs Makefile] «el Makefile solo soporta PostgreSQL 16»` sobre un `Makefile` de dos líneas.
> La fuente **sí se había leído** — lo inventado era lo que se le atribuía.

> **Consecuencia esperada y aceptada**: tras esto, la primera corrida parecerá **menos**
> fundamentada que hoy — más segmentos en tentativo. Es la señal de que funciona.

### Tests (test-first) ⚠️

- [X] T058 [P] Unit: `parse_segments` degrada a tentativo un `grounded` cuya cita **no aparece** en el material, y conserva el `Grounded` cuando sí aparece — en `crates/codify-core/src/application/authoring_loop.rs` (módulo `tests`)
- [X] T059 [P] Unit: una `contradiction` sin cita comprobable **de cada** fuente no se afirma (FR-006b) — en `crates/codify-core/src/application/authoring_loop.rs` (módulo `tests`)
- [X] T060 [P] Escenario: el caso exacto de F-1 — citar una fuente **leída** atribuyéndole algo ausente ⇒ tentativo con motivo — en `crates/codify-core/tests/us1_provenance.rs`

### Dominio

- [X] T061 `Groundedness::Grounded` y `Contradiction` ganan `quotes: Vec<String>` en `crates/codify-core/src/domain/context.rs`, con los constructores y `render()` al día (data-model.md)

### Verificación en el núcleo

- [X] T062 `parse_segments` recibe el **material leído** y comprueba cada cita; lo que no se sostiene se degrada a `Tentative` declarando el motivo — en `crates/codify-core/src/application/authoring_loop.rs` (depende de T061). **Ojo: 5 puntos de llamada** — `authoring_loop.rs:477`, `refine.rs:193`, `service.rs:397` y `:471`, más los tests
- [X] T063 Decidir y documentar la **normalización** de la comparación (espacios, mayúsculas, saltos de línea) en el propio `parse_segments`: una cita que solo difiere en formato debe seguir contando, o el criterio será inútilmente estricto

### Prompt

- [X] T064 El esquema de salida pasa a exigir `quotes` junto a `grounded` y en cada lado de `contradiction`, con la instrucción de que la cita sea **textual** — en `GENERATE_SYSTEM_PROMPT` de `crates/codify-core/src/application/authoring_loop.rs` y alineado con `contracts/agent-tools.md`

### Cierre

- [X] T065 Correr el arnés `live_backend` contra un modelo real y comprobar que F-1 **no se reproduce** — **hecho el 2026-08-23** con `Qwen2.5-32B-Instruct-Q4_K_M`, tres corridas completas regenerando el fixture entre cada una **(SC-001, medido sobre ≥3 corridas)**: **cero citas sin respaldo** en las tres, y entre 11 y 24 segmentos fundamentados por corrida, que es lo que descarta que la defensa funcione degradándolo todo. Detalle y los tres defectos que salieron por el camino, en `quickstart.md`

**Checkpoint**: `Grounded` es una afirmación comprobada, no una declarada. El principio rector
del proyecto deja de depender de la buena fe del modelo.


## Phase 8: La salida propia no fundamenta — FR-006d (issue #34)

**Goal**: que el sistema no pueda apoyarse en lo que él mismo escribió. Hoy un artefacto de una
sesión anterior entra al material como cualquier archivo, y una cita suya **verifica**: la
comprobación de la Fase 7 no ve nada raro, porque la cita sí aparece en lo leído.

**Independent Test**: sembrar material donde la única fuente citada es `context/CONTEXT.md` ⇒ el
segmento **se degrada a tentativo** nombrando el motivo, aunque la cita esté literalmente ahí.

> **Origen**: hallazgo de la segunda pasada con modelo real (2026-08-23). Al encadenar corridas
> sin regenerar el fixture, una sesión resolvió la contradicción sobre la persistencia entre
> `docs/SPEC-30.md` y **`context/CONTEXT.md` — su propia salida previa** — en vez de
> `docs/PRD.md`, y la presentó como procedencia verificada. En verde.

> **Relación con la Fase 7**: aquella impide **atribuir a una fuente algo que no dice**; esta
> impide **tratar como fuente algo que no lo es**. Son los dos lados del mismo principio, y la
> primera no puede detectar lo segundo por construcción.

### Tests (test-first) ⚠️

- [X] T066 [P] Unit: un `grounded` cuya **única** fuente citada es un artefacto propio se degrada a tentativo, y el motivo nombra el artefacto — en `crates/codify-core/src/application/authoring_loop.rs` (módulo `tests`)
- [X] T067 [P] Unit: un `grounded` que **además** cita una fuente real con cita comprobable **se mantiene** — FR-006d degrada lo que *solo* se apoya en la salida propia, no lo que la menciona — en `crates/codify-core/src/application/authoring_loop.rs` (módulo `tests`)
- [X] T068 [P] Escenario: el caso exacto de la medición — `context/CONTEXT.md` como lado de una contradicción ⇒ no se afirma (FR-006b + FR-006d) — en `crates/codify-core/tests/us1_provenance.rs`

### Dominio

- [X] T069 `ArtifactKind` sabe reconocer sus propias rutas: `is_canonical_path(&str) -> bool` en `crates/codify-core/src/domain/context.rs`, cubriendo **las cinco** —`Idioms` incluido, que no está en `default_set()`— y normalizando el separador. Es del Dominio porque es él quien nombra esas rutas (`file_path()`)

### Verificación en el núcleo

- [X] T070 `GatheredSource` distingue **fuente** de **artefacto propio**, y la ingesta lo clasifica al reunir el material — en `crates/codify-core/src/application/authoring_loop.rs` y `crates/codify-core/src/application/ingest.rs` (depende de T069)
- [X] T071 `fuentes_leidas` deja de admitir artefactos propios como respaldo, y el motivo del degradado lo dice — en `crates/codify-core/src/application/authoring_loop.rs` (depende de T070). **Ojo**: el material se lee igual; lo que cambia es qué puede sostener una afirmación

### Contratos

- [X] T072 [P] `data-model.md` recoge el origen en `GatheredSource`, y `contracts/ports.md` y `contracts/agent-tools.md` dicen que la salida propia no fundamenta — en `specs/001-context-authoring/`

### Cierre

- [X] T073 Arnés en vivo: **dos pases encadenados sin regenerar el fixture** — **hecho el 2026-08-24** con `Qwen2.5-32B`: el segundo pase dejó **18 segmentos fundamentados y ninguno citando una ruta canónica**, y resolvió la contradicción contra `docs/PRD.md` vs `docs/SPEC-30.md`, las fuentes correctas. **Matiz honesto**: ese pase no llegó a leer ningún artefacto propio, así que comprobó el resultado sin ejercitar el mecanismo en vivo; el test avisa por consola cuando eso pasa, y quien lo cubre de forma determinista es `us1_provenance.rs`

**Checkpoint**: el sistema no puede citarse a sí mismo. El bucle donde lo afirmado ayer respalda
lo de hoy queda cerrado, y `context/` vuelve a ser lo que US3 necesita —material para proponer
una actualización— sin ser evidencia.


## Dependencies & Execution Order

### Fase 7 (verificación de procedencia) — dónde encaja

**No depende de ninguna user story pendiente**: toca el dominio, el parseo y el prompt, que ya
existen. Se puede tomar de inmediato.

Dentro de la fase el orden sí importa: **T058–T060 (tests) → T061 (dominio) → T062–T063
(verificación) → T064 (prompt) → T065 (cierre con modelo real)**. T062 no puede escribirse antes
que T061 porque necesita el campo `quotes`; y T064 va después de T062 porque cambiar el prompt
sin tener la verificación deja al modelo produciendo un campo que nadie mira.

### Fase 8 (la salida propia no fundamenta) — dónde encaja

**Depende de la Fase 7**: T071 modifica `fuentes_leidas`, que nace allí. Fuera de eso no toca
ninguna user story pendiente.

El orden interno: **T066–T068 (tests) → T069 (dominio) → T070 (clasificación) → T071
(verificación) → T072 (contratos) → T073 (cierre en vivo)**. T070 no puede escribirse antes que
T069 porque necesita saber qué es una ruta canónica, y T071 va después de T070 porque sin el
origen en el material no hay nada que excluir.

T072 va con `[P]`: es documentación y no bloquea a T073.

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
Task: "Contract test RepoNavigator en crates/codify-core/tests/contract_ports.rs"
Task: "Contract test ReferenceResolver en crates/codify-core/tests/contract_ports.rs"
Task: "Contract test ModelProvider local en crates/codify-core/tests/contract_ports.rs"

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

## Trazabilidad — qué valida cada criterio

| | Validado por | Estado |
|---|---|---|
| **SC-001** ≥90 % consistente con la SSOT | T065 (3 corridas en vivo) | ✅ |
| **SC-002** 0 afirmaciones no verificadas como hecho | T058–T060, T066–T068 | ✅ |
| **SC-003** de «sin contexto» a «aprobado» en una sesión | T050 (S1–S3) | 🔶 falta S4 |
| **SC-004** refinamiento por conversación + diffs | US2 (T031–T040) | ✅ |
| **SC-005** 0 sobrescrituras silenciosas | US3 (T041–T044) | ✅ |
| **SC-006** nunca inventa una referencia no resuelta | T050 (S2) | ✅ |
| **SC-007** cero egress en modo local | `egress_guard.rs`, T052 | ✅ |

FR-018 (degradación declarada) no tiene SC propio: lo cubre SC-002 por la vía de no presentar
como firme lo que no lo es.

Los FR sin ID citado en su tarea **están cubiertos** — verificado contra el código, no contra el
papel. La cita se añade al tocar cada tarea, no en una pasada masiva que nadie revisaría y que
convertiría la trazabilidad en ceremonia.

## Notes
- [P] = archivos distintos, sin dependencias pendientes.
- Verificar que cada test falla antes de implementar (constitución II, [NN]).
- Commit atómico por tarea/grupo lógico (Conventional Commits; **sin atribución IA**; constitución III, [NN]).
- Detenerse en cada checkpoint para validar la story de forma independiente.
- Evitar: tareas vagas, conflictos en el mismo archivo, dependencias cross-story que rompan la independencia.
- BDD Discovery (Three Amigos / Example Mapping) se front-cargó en `spec.md` + `quickstart.md` (los ejemplos ya acordados). En build individual se acepta conscientemente arrancar en Formulation; si se forma equipo, correr Example Mapping por story antes de sus tests (constitución II).
