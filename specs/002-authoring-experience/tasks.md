---
description: "Task list — La experiencia de authoring (002)"
---

# Tasks: La experiencia de authoring — ver, entender y decidir

**Input**: Design documents from `specs/002-authoring-experience/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, constitution v1.0.0

**Tests**: INCLUIDOS y **test-first** — la constitución (Principio II) lo manda como [NON-NEGOTIABLE]. Excepción declarada en `plan.md`: el DOM no tiene test automatizado en esta etapa; se valida con `quickstart.md`.

**Organization**: por user story. **La Fase 2 es la que carga el peso**: dos de los requisitos más consecuentes del spec (cancelación y escritura de artefactos) son trabajo del **núcleo**, no de la interfaz.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: paralelizable (archivos distintos, sin dependencias pendientes)
- **[Story]**: US1/US2/US3 (solo en fases de user story)
- Ruta de archivo exacta en cada tarea

## ✅ La dependencia externa que condicionaba el orden — resuelta

**La US2 de este spec estuvo bloqueada por la US2 del spec `001`** (el loop de refinamiento): no
se puede construir la revisión de propuestas mientras no exista nada que las genere. Por eso el
orden de entrega se apartó del orden de prioridad, deliberadamente:

**US1 (P1) → US3 (P3) → US2 (P2)** ← se llegó aquí

`001`-US2 **está entregada**: existen el motor de diffs, el clasificador de riesgo, el loop
curado y los tres métodos del servicio (`submit_message`/`pending_proposals`/`decide`). La US2
de aquí puede empezar.

### Lo que este spec le entregó a `001`

La relación no es solo de bloqueo: la **Fase 2 de aquí** añadió al núcleo capacidades que
tareas de `001` pueden dar por hechas — port `ArtifactWriter` con `read_existing` (que su US3
necesitaba construir), la escritura real de artefactos a disco, y `start_session` no
bloqueante con `cancel_session`/`join_session`.

Además, la **US1 de aquí superó varias tareas de `001`** (su T030, T047, T048 parcial, T049).
El detalle está documentado en `001/tasks.md`, sección «Dependencias con el spec 002».

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).
> Este archivo sigue siendo la fuente de verdad de la **ejecución tarea a tarea**; los issues lo
> son de las **dependencias entre épicas**, porque allí se ven desde los dos lados a la vez.
> Épicas de este spec: **US3 → #6** ✅ · **US2 → #5** (desbloqueada: #4 entregada). Solapes de pulido: #8.

---

## Phase 1: Setup

**Purpose**: preparar el terreno. El workspace ya existe; esto es mínimo.

- [X] T001 Añadir `tokio-util` (feature `rt`) a `[workspace.dependencies]` en `Cargo.toml` y a `crates/codify-core/Cargo.toml`

---

## Phase 2: Foundational — prerrequisitos del NÚCLEO (Blocking)

**Purpose**: sin esto, la interfaz no tiene nada que mostrar ni que cancelar. **⚠️ Ninguna user story puede empezar antes.**

**Contexto**: hoy el núcleo **no escribe los artefactos a disco** (se quedan en memoria) y **no se puede cancelar**. Ambas cosas se resuelven aquí.

### Definición de ports y tipos

- [X] T002 [P] Tipo de dominio `WriteRecord` (path, bytes, at, outcome Written/Skipped/Failed) en `crates/codify-core/src/domain/write.rs` y registrarlo en `domain/mod.rs`
- [X] T003 [P] Ampliar `AuditKind` con `ArtifactWritten` y `SessionCancelled` en `crates/codify-core/src/domain/audit.rs` (cambio aditivo)
- [X] T004 Ports `Cancellation`, `ArtifactWriter` y `ProviderDiscovery` (+ `ProviderStatus`) en `crates/codify-core/src/application/ports.rs`, según `contracts/core-ports.md` (depende de T002)

### Tests de contrato (deben fallar antes de implementar) ⚠️

- [X] T005 [P] Fakes de los tres ports nuevos en `crates/codify-core/tests/fakes/mod.rs` (bandera en memoria para cancelación; writer en mapa; sonda scriptable)
- [X] T006 [P] Contract test `Cancellation`: una vez cancelado no se "descancela"; varios esperadores de `cancelled()` despiertan — en `crates/codify-core/tests/contract_cancellation.rs`
- [X] T007 [P] Contract test `ArtifactWriter` (real fs + fake): escribe y relee; **rechaza rutas absolutas y con `..`**; crea directorios intermedios; un fallo aislado no arrastra al resto — en `crates/codify-core/tests/contract_artifact_writer.rs`
- [X] T008 [P] Contract test `ProviderDiscovery`: sin backend devuelve `reachable:false` **con `issue` presente** y **nunca `Err`** — en `crates/codify-core/tests/contract_provider_discovery.rs` (el campo pasó de `detail` a `issue` en T048: el núcleo ya no redacta la frase, nombra el motivo)

### Adapters

- [X] T009 [P] `TokenCancellation` (envuelve `tokio_util::sync::CancellationToken`) en `crates/codify-core/src/infrastructure/cancel.rs`
- [X] T010 [P] `FsArtifactWriter` con las mismas defensas de ruta que el navegador, en `crates/codify-core/src/infrastructure/repo/writer.rs`
- [X] T011 [P] Sonda del backend (endpoint OpenAI-compatible, respaldo en API de Ollama; solo loopback en modo local) en `crates/codify-core/src/infrastructure/providers/probe.rs`

### Integración en el loop y el servicio

- [X] T012 Añadir los tres ports a `AuthoringDeps` en `crates/codify-core/src/application/deps.rs` y cablearlos en `CoreBuilder` (`infrastructure/composition.rs`) (depende de T004, T009-T011)
- [X] T013 Test: cancelar durante la ingesta detiene el loop y la sesión queda en `Cancelled`, reportando sus escrituras — en `crates/codify-core/tests/us1_cancellation.rs`
- [X] T014 Test: cancelar **durante la llamada al modelo** la aborta sin esperar a que termine — en `crates/codify-core/tests/us1_cancellation.rs` (proveedor fake con retardo)
- [X] T015 El loop respeta la cancelación: comprueba en cada punto de control y compone `tokio::select!` alrededor de la llamada al modelo, en `crates/codify-core/src/application/authoring_loop.rs` (depende de T013, T014)
- [X] T016 Test: al terminar la generación los artefactos **están en disco** y cada escritura queda auditada — en `crates/codify-core/tests/us1_artifact_writing.rs`
- [X] T017 El loop escribe los artefactos vía `ArtifactWriter` y acumula `WriteRecord`, emitiendo `AuditKind::ArtifactWritten` por cada uno, en `crates/codify-core/src/application/authoring_loop.rs` (depende de T016)
- [X] T018 `start_session` **deja de bloquear** (arranca el loop en una tarea y retorna el `SessionId`) y se añade `cancel_session` **que devuelve `CancelOutcome` (fase + escrituras acumuladas)**; `SessionSnapshot` gana `writes` — en `crates/codify-core/src/application/service.rs`
- [X] T019 Test: `start_session` retorna antes de que el trabajo termine y `session_state` refleja el avance — en `crates/codify-core/tests/us1_nonblocking.rs`

**Checkpoint**: el núcleo escribe, se puede cancelar y no bloquea. Las user stories pueden empezar.

---

## Phase 3: User Story 1 — Ver trabajar al agente (Priority: P1) 🎯 MVP

**Goal**: el usuario ve en vivo qué lee el agente y qué no logra resolver, con la interfaz utilizable y la sesión cancelable.

**Independent Test**: iniciar una sesión sobre el fixture y poder narrar, solo mirando la pantalla, qué leyó el agente y qué no — sin logs ni terminal (quickstart S1).

### Catálogo de cadenas y comandos

- [X] T020 [P] [US1] Test de paridad del catálogo: mismas claves en `es` y `en`, ningún valor vacío — en `crates/codify-app/src/strings.rs` (módulo `tests`)
- [X] T021 [US1] Catálogo de cadenas `es`/`en` con espacios `session.*`, `stream.*`, `provider.*`, `error.*`, `a11y.*`, **`locale.*`** (idioma de interfaz vs. de artefactos) y **`mode.*`** (local/híbrido), y `system_locale()` con caída a `en` — en `crates/codify-app/src/strings.rs` (depende de T020)
- [X] T022 [P] [US1] Comandos `cancel_session` (**devuelve `CancelOutcomeDto`**) y `probe_provider` en `crates/codify-app/src/commands.rs`, según `contracts/skin-commands.md`
- [X] T023 [P] [US1] Comandos `ui_strings` y `system_locale` en `crates/codify-app/src/commands.rs`
- [X] T024 [US1] Emitir los eventos nuevos `artifact.written` y `session.cancelled` desde `EventAuditSink` en `crates/codify-app/src/adapters.rs`

### Interfaz

- [X] T025 [P] [US1] Estructura semántica y regiones (`<main>`, `<article>`, `aria-live` para la corriente) en `crates/codify-app/ui/index.html`
- [X] T026 [P] [US1] Consumo del catálogo de cadenas (sin texto incrustado en la vista) en `crates/codify-app/ui/i18n.js`
- [X] T027 [US1] La corriente cronológica de bloques append-only, con **etiqueta + forma + color** por tipo, en `crates/codify-app/ui/stream.js` (depende de T026)
- [X] T028 [US1] Onboarding guiado del proveedor: estado, elección de modelo y **qué hacer cuando falta algo**, en `crates/codify-app/ui/provider.js`
- [X] T029 [US1] Orquestación: iniciar, **cancelar**, suscripción a eventos y presentación accionable de fallos (FR-028), en `crates/codify-app/ui/main.js` (depende de T027, T028)
- [X] T051 [US1] Indicador **persistente** de modo local en la barra + render del bloque `egress.blocked` cuando se intenta una salida, en `crates/codify-app/ui/main.js` y `crates/codify-app/ui/stream.js` [FR-005, SC-006]
- [X] T030 [US1] Estilos de los estados con señal redundante y foco visible, en `crates/codify-app/ui/styles.css`
- [X] T031 [US1] Atajos y orden de tabulación para iniciar, cancelar y recorrer la corriente, en `crates/codify-app/ui/main.js`
- [X] T032 [US1] Mostrar el **balance de escrituras** al terminar o cancelar (FR-017/FR-023), en `crates/codify-app/ui/main.js`
- [X] T052 [US1] Presentar lo que quedó **sin leer** al terminar la ingesta (`omitted` + `budgetExhausted` del snapshot), de modo que el resultado nunca aparente ser completo, en `crates/codify-app/ui/main.js` [FR-004]
- [X] T053 [US1] Confirmación al cerrar la aplicación con una sesión en curso, declarando **qué se perderá** y qué ya está en el repositorio, en `crates/codify-app/ui/main.js` (usa el evento de cierre de ventana de Tauri) [FR-024]
- [X] T054 [US1] Control para cambiar el **idioma de los artefactos** (comando `set_locale`), visiblemente distinto del selector de idioma de la interfaz, en `crates/codify-app/ui/main.js` [FR-016]

**Checkpoint**: US1 completa — **MVP de la experiencia**. Validar con quickstart S1, S2, S4, S6.

---

## Phase 4: User Story 3 — Leer el contexto sabiendo en qué se apoya (Priority: P3)

**Goal**: el usuario abre un artefacto completo en cualquier momento y distingue sin ambigüedad qué está fundamentado, qué es tentativo y qué se contradice.

**Independent Test**: mostrar un artefacto a alguien ajeno a la sesión y pedirle que señale qué está verificado; debe acertar sin instrucción previa (quickstart S3).

> Se entrega **antes que la US2** porque aquella está bloqueada por el spec 001 (ver arriba). No es un cambio de prioridad: es la única secuencia posible.

- [X] T033 [P] [US3] Test: el DTO del artefacto expone los tres estados con sus fuentes/motivo y el estado de escritura — en `crates/codify-app/src/commands.rs` (módulo `tests`)
- [X] T034 [US3] Comando `artifact` que devuelve el artefacto completo (`ArtifactDto` con `writeState`) en `crates/codify-app/src/commands.rs` — **el comando ya existía** (entregado de refilón con T029); lo que faltaba era `writeState`, sin el cual la vista no puede distinguir un archivo escrito de uno que solo existe en pantalla
- [X] T035 [US3] Vista de artefacto completo, alcanzable **en cualquier momento** sin recorrer la corriente (FR-021), en `crates/codify-app/ui/artifact.js` — se puede abrir **durante** la sesión: cada evento `artifact.written` lo añade a la lista, así FR-021 es un contrapeso real y no una consolación al final
- [X] T036 [US3] Render de los tres estados con **etiqueta textual** como señal primaria, fuentes consultables y nota de contradicción, en `crates/codify-app/ui/artifact.js` (depende de T035)
- [X] T037 [US3] Aviso al intentar cerrar con tentativos sin atender, con opción de diferirlos explícitamente (FR-014 del spec 002), en `crates/codify-app/ui/main.js` — requirió **caso de uso nuevo** en el núcleo: `AuthoringService::defer_tentative`, por fragmento y no en bloque (ver abajo)
- [X] T038 [US3] Añadir las claves `artifact.*` al catálogo, incluidas las etiquetas de los tres estados, en `crates/codify-app/src/strings.rs` — **las nueve ya existían**, reservadas desde T048; se añadieron las de diferir, estado de escritura y aviso de cierre

### Decisiones que salieron de la US3

**Diferir es por fragmento, no en bloque.** FR-014 pide poder «resolverlos o diferirlos». Un
botón de *diferir todo* habría sido más cómodo y exactamente el hábito que este producto viene a
corregir: despachar sin leer. El caso de uso `defer_tentative(sesión, ruta, índice)` obliga a
señalar **cuál**, y «Revisarlos» en el aviso de cierre lleva el foco al primero sin atender —
donde está el control. Cuatro tests fijan que diferir uno atienda exactamente uno, que hacerlo
dos veces no descuente dos, y que sobre algo verificado falle.

**Un fragmento diferido sigue marcado como sin verificar.** No se le asciende: se le añade
«pendiente, a sabiendas». Diferir es una decisión sobre qué hacer, no un cambio en lo que consta.

**El estado de escritura es parte del artefacto.** `writeState` (`written` | `pending` |
`failed` | `skipped`) se deriva de la última escritura registrada para esa ruta. Sin él, un
archivo en pantalla y un archivo en el repositorio se veían igual — afirmar lo segundo sin
respaldo es justo lo que el producto no hace (FR-017).

**Un defecto propio, encontrado midiendo.** Al cambiar de idioma con la vista abierta, las
etiquetas de fundamento se quedaban en el idioma anterior: `i18n.apply()` no alcanza al texto
escrito a mano. Tercera vez que aparece esta familia. Ahora hay un test que la persigue como
clase —`quien_pinta_a_mano_repinta_al_cambiar_de_idioma`— y no instancia a instancia.

**Checkpoint**: US1 + US3 funcionando. Validar con quickstart S3 (incluida la prueba en **escala de grises**).

---

## Phase 5: User Story 2 — Decidir sobre los cambios con el diff a la vista (Priority: P2) ✅ COMPLETA

**Goal**: revisar propuestas como diffs y decidir: aprobar, editar o rechazar.

**⚠️ Bloqueada por el spec 001**: requiere el loop de refinamiento (`001/tasks.md` T031–T040 y T055–T057), que genera las propuestas. **No empezar antes.**

- [X] T039 [US2] Comandos `submit_message`, `pending_proposals` y `decide` + evento `proposal.new`, en `crates/codify-app/src/commands.rs` (001-US2 **ya entregada**) — **`agent.token` no entró**: el streaming token a token exige que el turno deje de ser una unidad bloqueante, que es justo la decisión que sostiene el puente del `Prompter`. Extraído a T056 — `submit_message` y los eventos llegan aquí desde la T040 de `001`, que describía esta misma piel y quedó remitida a esta fase
- [X] T040 [US2] `Prompter` real de la piel que reemplaza a `UnavailablePrompter`, en `crates/codify-app/src/adapters.rs` (001-US2 **ya entregada**)
- [X] T041 [US2] Bloque de propuesta con diff legible y razón del cambio, en `crates/codify-app/ui/stream.js` — el bloque vive en la corriente como **registro** (append-only) y el diff revisable en el panel de decisión, que sí cambia; mezclarlos obligaría a reescribir bloques ya emitidos
- [X] T042 [US2] Captura de decisión (aprobar / editar / rechazar), en `crates/codify-app/ui/main.js` — editar es en **dos pasos**: abrir el editor y confirmar. Enviar a la primera aplicaría un texto vacío
- [X] T043 [US2] Contador de decisiones pendientes y navegación entre ellas sin perder lo ya decidido, en `crates/codify-app/ui/main.js`
- [X] T055 [US2] **Deshacer** un cambio auto-aplicado (FR-008): `revert_proposal` en el servicio, comando en la piel y control en `ui/applied.js` — extraído de T042, que lo prometía sin entregarlo. Solo aplica a lo de bajo riesgo: lo aprobado por una persona se cambia decidiendo otra vez, no a sus espaldas
- [ ] T056 Streaming token a token (`agent.token`) durante el refinamiento — **requiere decidir antes** si el turno debe dejar de bloquear; no es solo añadir un evento
- [X] T044 [US2] Refinamiento conversacional en lenguaje natural (sin interacción modal por hueco), en `crates/codify-app/ui/main.js` — el compositor se bloquea mientras el turno se resuelve: un segundo mensaje encima del primero produciría propuestas sobre un contexto a punto de cambiar

---

## Phase 6: Polish & Cross-Cutting

> **Cómo se verificaron S5–S8.** Estaban escritas como inspección manual. Marcarlas hechas por
> haberlas mirado una vez es justo lo que la constitución rechaza: no se verifica contra una
> fuente y se pudre al siguiente cambio de UI. Se hicieron dos cosas:
>
> 1. **Medición empírica**: la interfaz es HTML/CSS/módulos ES planos, así que se cargó en un
>    navegador con el puente de Tauri simulado y se **midió** — orden de tabulación contra
>    orden visual, desbordamiento a la anchura mínima real, recorrido del DOM en ambos idiomas.
>    Eso encontró **cuatro defectos** que ninguna lectura del código habría dado por seguros.
> 2. **Checks durables** en `crates/codify-app/tests/ui_contract.rs` (8 tests), cada uno
>    verificado inyectando su violación. No sustituyen al recorrido humano —nadie automatiza
>    «se entiende»— pero clavan lo que sí es mecánico, que es lo que se rompe solo.

- [X] T045 [P] Repaso de accesibilidad: recorrer un flujo completo **solo con teclado** y verificar foco y orden de tabulación (quickstart S6) — **sin defectos**: orden de tabulación == orden visual, cero `tabindex` positivos, ninguna acción solo-ratón, y anillo de foco real (`2px solid`, offset 2px) confirmado con pulsación de tecla auténtica. Fijado por `no_hay_tabindex_positivo` y `el_foco_siempre_deja_rastro_visible`
- [X] T046 [P] Verificar la ventana en tamaño mínimo: sin recorte ni desplazamiento horizontal (quickstart S7, SC-007) — **defecto**: `minWidth: 820` volvía **inalcanzable** la media query de `720px`; las reglas de «ventana pequeña» eran código muerto que aparentaba cubrir SC-007. Bajado a `720`, comprobado sin desbordamiento hasta 640px con contenido de peor caso. Fijado por `el_punto_de_quiebre_responsivo_es_alcanzable`
- [X] T047 [P] Modo entrevista para repositorio vacío en la interfaz (quickstart S8) — **presentación hecha, entrevista pendiente**. Salía como un bloque gris indistinguible del ruido y con el estado «terminada», que sugiere un resultado inexistente (FR-004). Ahora tiene tipo de bloque propio, estado propio (`session.state.interview`) y dice **qué hacer**. La entrevista conversacional ya es posible: la **US2 (#5)** entregó la caja de texto, y el compositor se habilita también en modo entrevista — que era el caso donde más falta hacía y donde estaba apagado. Fijado por `el_repositorio_vacio_tiene_presentacion_propia`
- [X] T048 [P] Recorrer la aplicación en **ambos idiomas** buscando claves crudas (quickstart S5, SC-009) — **tres defectos**, todos de la misma familia: texto escrito a mano que `i18n.apply()` no alcanza. (a) «sin sesión» se congelaba en el idioma de arranque; (b) `#provider-status` tenía **dos dueños** —`data-i18n` y `provider.js`—, así que tras cambiar de idioma el panel decía «comprobando…» con el glifo en ✓, mintiendo sobre si había backend; el test destapó una tercera instancia en el botón `#action`, que habría podido decir «Iniciar» con la sesión en curso; (c) el «qué hacer» del proveedor venía **redactado en español desde el núcleo**. Fijado por `ningun_texto_visible_escapa_al_catalogo`, `ningun_elemento_tiene_dos_duenos_de_su_texto` y `todo_motivo_del_proveedor_tiene_texto_en_ambos_idiomas`

### Cambio de diseño que salió de T048

El núcleo **dejó de redactar prosa para humanos**. `ProviderStatus.detail: Option<String>` —una
frase ya escrita, y por tanto en un idioma fijo— pasó a `issue: Option<ProviderIssue>`, un motivo
nombrado con código estable que la piel traduce (`provider.issue.<code>`).

No es solo i18n: el núcleo redactando texto de interfaz es presentación colándose en la
aplicación. Con el motivo como dato, SC-009 vuelve a ser **demostrable** —un test recorre los
motivos y comprueba que ninguno se quedó sin texto en ninguno de los dos idiomas— en vez de
depender de que alguien mire la pantalla en inglés.
- [~] T049 Ejecutar la validación completa de `quickstart.md` (S1–S8) — **S5, S6 y S7 automatizados** y verificados (`ui_contract.rs`); **S8 parcial** (presentación cubierta, la entrevista con modelo real no). **S1–S4 pendientes del operador**: exigen Ollama y un repositorio con referencias, y miden lo que una persona logra entender mirando — SC-001 y SC-002 no los decide un test. Fixture listo: `./scripts/quickstart-fixture.sh`. Detalle por escenario en la tabla de `quickstart.md`
- [X] T050 Verificar en CI que fmt, clippy y las fitness functions siguen verdes — verificado sobre el commit `68ed5df` (run 31377803807, `completed/success`); el paso «Tests (incluye fitness functions de arquitectura y cero-egress)» pasó, y `cargo test --workspace` recorre `arch_deps.rs` y `egress_guard.rs` por vivir en la raíz de `tests/`. **Limitación declarada**: los logs del run no se pudieron leer desde esta máquina (**token de `gh` inválido**, no falta de permisos), así que la evidencia es la conclusión del paso, no la lectura línea a línea

---

## Phase 7: Un fallo que se puede entender — FR-028 (issue #24)

**Goal**: que `Failed` deje de ser un callejón sin salida. Hoy la sesión puede terminar con cero
artefactos y **sin decir por qué**, contra lo que FR-028 exige: explicar qué ocurrió y qué puede
hacer el usuario.

**Independent Test**: provocar un fallo del proveedor ⇒ la vista expone un **código de motivo**
y la interfaz muestra una frase del catálogo más un siguiente paso, nunca el error crudo.

> **Origen**: hallazgo F-2 de la pasada con modelo real, y **medido** el 2026-08-24 durante la
> Fase 7 de `001`: diagnosticar un simple *timeout* de cliente costó **cinco corridas perdidas**
> de entre 4 y 9 minutos, más una hipótesis falsa perseguida hasta el final. No por difícil —
> porque sin motivo cualquier hipótesis vale lo mismo y se persiguen por plausibilidad en vez de
> por evidencia.

> **El defecto es de una línea**: `crates/codify-core/src/application/service.rs` hace
> `Err(_) => { advance_to(Failed) }`. El motivo **existe** —`CoreError` tiene siete variantes
> tipadas— y se descarta con un comodín.

> **Reparto**: el requisito incumplido es de `002`, pero el motivo nace en el núcleo. Las tareas
> que tocan `codify-core` lo dicen; es el mismo reparto que la degradación de tier (`001`-T046).

### Tests (test-first) ⚠️

- [ ] T057 [P] Cada variante de `SessionFailure` tiene texto **y siguiente paso** en los dos idiomas — en `crates/codify-app/tests/ui_contract.rs`, siguiendo el patrón de `todo_motivo_del_proveedor_tiene_texto_en_ambos_idiomas`
- [ ] T058 [P] Una sesión que muere por fallo del proveedor expone el **código**, no el mensaje crudo, en `crates/codify-core/tests/us1_session_failure.rs`
- [ ] T059 [P] El caso medido: un *timeout* del proveedor produce un motivo distinguible de «el modelo respondió algo no parseable» — sin esa distinción el hallazgo de #24 se repite — en `crates/codify-core/tests/us1_session_failure.rs`

### Dominio (núcleo)

- [ ] T060 `SessionFailure` con `code()` estable, derivable desde `CoreError`, en `crates/codify-core/src/domain/session.rs`. Sigue el precedente de `ProviderIssue`/`ReferenceState`/`RiskLevel`: el núcleo devuelve un código, la piel elige la frase — es lo que permite que el motivo no nazca redactado en un idioma fijo
- [ ] T061 `AuthoringSession` transporta el motivo al pasar a `Failed`, y `advance_to` deja de aceptar ese estado sin él — en `crates/codify-core/src/domain/session.rs` (depende de T060). **Que el tipo lo exija** es lo que impide que vuelva a perderse

### Núcleo → vista

- [ ] T062 `service.rs` deja de descartar el error con `Err(_)`: lo mapea a `SessionFailure`, lo audita (`AuditKind::SessionFailed`) y lo publica en `SessionSnapshot.failure` — en `crates/codify-core/src/application/service.rs` (depende de T061)

### Interfaz

- [ ] T063 El DTO lleva el código y la interfaz lo resuelve contra el catálogo: frase + siguiente paso, nunca el crudo — en `crates/codify-app/src/commands.rs`, `crates/codify-app/src/strings.rs` y `crates/codify-app/ui/main.js`. **Ojo**: `error.session_failed` es hoy la frase genérica; queda como respaldo solo para fallos sin código, o se retira si no queda ninguno

### Contratos

- [ ] T064 [P] `contracts/skin-commands.md` documenta el campo `failure` del snapshot, y `contracts/ui-strings.md` recoge las claves `session.failure.*` en su superficie correspondiente — en `specs/002-authoring-experience/contracts/`

### Cierre

- [ ] T065 Provocar un fallo real contra un backend inalcanzable y comprobar que la interfaz explica y ofrece salida — escenario **S9** en `specs/002-authoring-experience/quickstart.md`

**Checkpoint**: un fallo deja de ser un misterio. El coste que #24 se cobró en la Fase 7 de `001`
—cinco corridas para encontrar un timeout— no se vuelve a pagar.


## Dependencies & Execution Order

### Phase Dependencies
- **Setup (1)**: sin dependencias.
- **Foundational (2)**: depende de Setup — **BLOQUEA todo lo demás**. Es el grueso del riesgo técnico.
- **US1 (3)**: tras Foundational.
- **US3 (4)**: tras Foundational. Reutiliza el catálogo y la orquestación de US1, así que en la práctica va después.
- **US2 (5)**: tras Foundational **y** tras `001-US2`. Bloqueada por un spec externo.
- **Polish (6)**: al final.

### Within Each Story
Tests primero (deben fallar), luego dominio/ports, luego adapters, luego comandos, luego interfaz.

### Parallel Opportunities
- Foundational: T002/T003 en paralelo; T005-T008 (tests) en paralelo; T009-T011 (adapters) en paralelo.
- US1: T022/T023 en paralelo; T025/T026 en paralelo.
- Polish: T045-T048 en paralelo.
- **No paralelizable**: T012→T015→T017→T018 es una cadena — cada eslabón toca el loop o el servicio.

---

## Parallel Example: Fase 2 (Foundational)

```bash
# Los tres adapters, en paralelo (archivos distintos):
Task: "TokenCancellation en infrastructure/cancel.rs"
Task: "FsArtifactWriter en infrastructure/repo/writer.rs"
Task: "Sonda del proveedor en infrastructure/providers/probe.rs"
```

---

## Implementation Strategy

### MVP (US1)
Setup → **Foundational completa** → US1 → validar con quickstart S1/S2/S4/S6. Ahí ya se ve trabajar al agente, se puede cancelar y se sabe qué se escribió al repositorio.

### Incremental
1. **Fase 2** es el hito de mayor valor aunque no se vea: cierra la deuda de que el producto no entregaba sus archivos.
2. **US1** hace visible el trabajo del agente.
3. **US3** hace legible el fundamento.
4. **US2** cuando 001-US2 desbloquee.

---

## Notes
- [P] = archivos distintos, sin dependencias pendientes.
- Verificar que cada test falla antes de implementar (constitución II, [NN]).
- Commit atómico por tarea o grupo lógico; Conventional Commits, **sin atribución de IA** (constitución III, [NN]).
- El DOM no tiene test automatizado por decisión documentada en `plan.md`; por eso las tareas de interfaz se cierran validando con `quickstart.md`, no con `cargo test`.
