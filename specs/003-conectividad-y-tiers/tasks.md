# Tasks — Conectividad y reparto de modelos

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) · **Issue**: [#42](https://github.com/jorelcb/codify/issues/42)

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).

## Lo que ya existe y no hay que construir

Conviene saberlo antes de empezar, porque tres piezas del diseño ya están en el árbol:

- **`ProviderRegistry` ya enruta por tier** y devuelve un `Routed { provider, degraded_from }`
  (`001`-T046). Con un solo proveedor no distinguía nada; con dos, empieza a significar lo que dice.
- **La degradación ya se declara** (`001`-FR-018): evento de auditoría, campo en el snapshot y
  frase de catálogo. Este spec la usa, no la reescribe.
- **`SessionFailure::Unauthorized` ya existe** (`002`-FR-028), así que un fallo de autorización ya
  puede llegar al usuario distinguido de un fallo del modelo.
- **FR-007 de este spec no tiene tarea, y es correcto**: pide degradar y declararlo, que es
  exactamente `001`-FR-018, entregado en la PR #38. Se dice aquí para que una revisión futura no
  lo cuente como hueco de cobertura.

---

## Phase 1: Setup

- [X] T001 Añadir `oauth2` (device-authorization flow) y `keyring` como dependencias de `crates/codify-core/Cargo.toml`, y **`trybuild` como `dev-dependency`** — es lo que permite afirmar que un programa **no compila**, y sin ello T003 no se puede escribir
- [X] T002 [P] Crear el módulo `crates/codify-core/src/infrastructure/secrets/mod.rs`, declarado en su `mod.rs` padre. **`providers/remote.rs` no se crea aquí**: lo crea T020, que es quien lo llena. Un módulo vacío durante cuatro fases es un stub sin fecha de caducidad

---

## Phase 2: Foundational — el modo en el tipo (BLOQUEA todo lo demás)

**Goal**: que un grafo local no pueda contener un adapter de red. Todo lo demás depende de que
esta forma esté fijada, porque cambia la firma con la que se construye el sistema.

- [X] T003 ⚠️ **Primero**: test de **compilación fallida** con `trybuild` — un programa que llama al método de proveedor remoto sobre `CoreBuilder<Local>` no debe compilar. El caso va en `crates/codify-core/tests/compile_fail/local_no_admite_remoto.rs` y el arnés que lo ejecuta en `crates/codify-core/tests/compile_fail.rs` (depende de T001). Se escribe antes que T004 porque **es quien define la forma del constructor**. **Al escribirlo se vio que no debía importar el adapter remoto**: lo que prueba es que el método no existe, y el argumento da igual — atarlo a T020 habría cruzado una dependencia de fase sin ganar nada
- [X] T004 `CoreBuilder<M>` con `M ∈ {Local, Hybrid}` en `crates/codify-core/src/infrastructure/composition.rs`: el método que acepta un proveedor remoto existe **solo** en `CoreBuilder<Hybrid>` (depende de T003). **`new()` dejó de recibir el `Mode`**: al pasarlo como tipo y como valor, `CoreBuilder::<Local>::new(Mode::Hybrid)` compilaba y el grafo podía decir una cosa mientras el tipo decía otra. El modo sale ahora de `ModoDelGrafo::MODE`, fuente única
- [X] T005 Extender `crates/codify-core/tests/egress_guard.rs`: la comprobación de runtime de `ProviderRegistry::for_mode` **sigue viva** y se prueba explícitamente. Es defensa en profundidad, no redundancia — cubre un proveedor construido por otra vía
- [X] T006 [P] Migrar los puntos de construcción existentes a la firma nueva — `crates/codify-app/src/commands.rs` y los tests que arman el grafo (depende de T004)

**Checkpoint**: escribir un grafo local con salida a la red es un error de compilación.

---

## Phase 3: User Story 1 — Conectar una cuenta una sola vez (P1) 🎯 MVP

**Goal**: que exista un segundo proveedor. Sin esto, el resto del spec no tiene sobre qué operar.

**Independent Test**: conectar una cuenta y comprobar que una generación posterior la usa, sin que
la credencial aparezca en disco, en el registro ni en la interfaz.

### Tests (test-first) ⚠️

- [ ] T007 [P] [US1] Suite de contrato de `CredentialStore` —guardar, obtener, borrar idempotente, `disponible()` sin escribir— corriendo contra el adapter real **y** contra un doble en memoria, en `crates/codify-core/tests/contract_credential_store.rs`
- [ ] T008 [P] [US1] El secreto **no** aparece en la salida de `Debug` ni en un `AuditEvent`, en `crates/codify-core/tests/contract_credential_store.rs`. Es SC-002 y se comprueba buscándolo, no suponiéndolo
- [ ] T009 [P] [US1] Abandonar o denegar la autorización deja el sistema sin conexión a medias, **y desconectar una cuenta impide su uso en la tarea siguiente sin reiniciar** (SC-006) — en `crates/codify-core/tests/us1_connect_account.rs`
- [ ] T010 [P] [US1] Sin almacén disponible, conectar **falla diciendo por qué** y no escribe nada en disco (FR-004), en `crates/codify-core/tests/us1_connect_account.rs`

### Ports y dominio

- [ ] T011 [US1] `CredentialStore` y `AccountConnector` en `crates/codify-core/src/application/ports.rs`, con firmas en tipos de aplicación y **sin** tipo que transporte el secreto fuera del adapter
- [ ] T012 [US1] `ProviderConnection` en `crates/codify-core/src/application/` — sin campo para la credencial, con `tier` declarado al conectar (data-model.md)

### Adapters

- [ ] T013 [US1] `CredentialStore` contra el keyring del sistema en `crates/codify-core/src/infrastructure/secrets/keyring.rs`, **sin respaldo en archivo** (research.md D3) (depende de T011)
- [ ] T014 [US1] `AccountConnector` por device-flow en `crates/codify-core/src/infrastructure/secrets/device_flow.rs`, con **su propio** límite de tiempo — esperar a una persona no es esperar a un modelo
- [ ] T015 [P] [US1] `AccountConnector` por credencial directa en `crates/codify-core/src/infrastructure/secrets/direct.rs` (depende de T011)

### Piel

- [ ] T016 [US1] Comandos `connect_provider`, `complete_connection`, `list_connections` y `disconnect_provider` en `crates/codify-app/src/commands.rs`, con el DTO sin campo para el secreto (contracts/skin-commands.md). **Desconectar borra del almacén y rearma el grafo** (research.md D4): sin lo segundo, la conexión seguiría cableada hasta reiniciar y SC-006 no se cumpliría
- [ ] T017 [US1] Conectar y desconectar en la interfaz — `crates/codify-app/ui/` y claves nuevas en `crates/codify-app/src/strings.rs`, en los dos idiomas

**Checkpoint**: hay un segundo proveedor, y su credencial no está en ningún sitio que podamos leer.

---

## Phase 4: User Story 2 — Que lo barato haga lo frecuente (P2)

**Goal**: cumplir `001`-FR-017, el MUST que originó este spec. Depende de US1: sin segundo
proveedor no hay entre qué repartir.

**Independent Test**: con dos tiers conectados, el refinamiento va al económico y la generación al
de mayor capacidad, comprobable sin leer el código.

### Tests ⚠️

- [ ] T018 [P] [US2] Con dos proveedores de tiers distintos, el refinamiento va al económico y la generación pesada al de mayor capacidad, en `crates/codify-core/tests/us2_tier_routing.rs`
- [ ] T019 [P] [US2] Un fallo de **autorización** llega como `SessionFailure::Unauthorized` y no como fallo del modelo, en `crates/codify-core/tests/us2_tier_routing.rs`

### Implementación

- [ ] T020 [US2] **Crear** e implementar el adapter `ModelProvider` remoto genérico en `crates/codify-core/src/infrastructure/providers/remote.rs`: `is_local()` devuelve `false` y `tier_hint()` el tier **declarado**, no inferido (contracts/ports.md)
- [ ] T021 [US2] Cablear las conexiones guardadas al grafo híbrido en `crates/codify-core/src/infrastructure/composition.rs` (depende de T004, T013, T020)
- [ ] T022 [US2] Eventos `task.routed` —tier y conexión— y **`connection.state_changed`** —conectada, caducada o revocada—, en `crates/codify-core/src/domain/audit.rs` y `crates/codify-app/src/adapters.rs`. Los dos están en el contrato; el segundo se había quedado sin tarea
- [ ] T023 [US2] Mostrar qué tier atendió cada tarea en `crates/codify-app/ui/` (FR-006)

**Checkpoint**: iterar deja de costar lo que cuesta generar.

---

## Phase 5: User Story 3 — Saber qué sale del equipo (P3)

**Goal**: que el modo sea del usuario y que pueda reconstruir qué salió. Es lo que permite que
US1 exista sin romper la promesa del producto.

**Independent Test**: en modo local, comprobar que **no existe** ruta de egress — no que no se use.

### Tests ⚠️

- [ ] T024 [P] [US3] Cambiar de modo **no** afecta a una sesión en curso (FR-008b), en `crates/codify-core/tests/us3_mode.rs`
- [ ] T025 [P] [US3] Tras una sesión, el registro permite reconstruir **qué proveedor atendió cada tarea** (FR-010), en `crates/codify-core/tests/us3_mode.rs`

### Implementación

- [ ] T026 [US3] Comando `set_mode` que **rearma el grafo** sin reiniciar, en `crates/codify-app/src/commands.rs` y `crates/codify-app/src/lib.rs` (depende de T004)
- [ ] T027 [US3] Antes de arrancar una sesión híbrida, mostrar **qué proveedores podrían recibir contenido** del repositorio (FR-009), en `crates/codify-app/ui/`
- [ ] T028 [US3] Evento `mode.changed` y su presentación, incluyendo que la sesión viva no cambia, en `crates/codify-app/src/adapters.rs` y `crates/codify-app/ui/`

**Checkpoint**: el modo local sigue siendo una garantía, no una opción de configuración.

---

## Phase 6: Polish

- [ ] T029 [P] Cada código nuevo —estado de conexión, modo— tiene texto en **los dos idiomas**, con el test que recorre los códigos del núcleo, en `crates/codify-app/tests/ui_contract.rs`
- [ ] T030 [P] Contratos al día: `contracts/ports.md` y `contracts/skin-commands.md` reflejan lo entregado, y `002/contracts/skin-commands.md` recoge los campos nuevos del DTO de sesión si los hubiera
- [ ] T031 Ejecutar la validación de `quickstart.md` (S1–S7) — **S4 y S6 necesitan comprobarse fuera de la aplicación**: buscar la credencial en disco y en el keyring del sistema con sus propias herramientas, porque preguntárselo a la aplicación es preguntarle a la parte interesada

---

## Dependencies & Execution Order

**Fase 2 bloquea todo.** Cambia la firma con la que se construye el sistema, así que cualquier
tarea posterior escrita contra la firma vieja habría que rehacerla.

Dentro de la Fase 2 el orden es estricto: **T003 → T004 → T005/T006**. El test de compilación
fallida va primero porque es quien define la forma del constructor; escribirlo después sería
describir lo que ya se hizo en vez de dejar que decida.

**US1 → US2**: sin segundo proveedor no hay reparto. **US3** no depende de US2 y podría adelantarse
si interesa cerrar antes la parte de garantía.

### Oportunidades de paralelismo

Ocho tareas `[P]`: los tests de cada fase entre sí, T015 con T014, y las dos de Polish.

## Implementation Strategy

**MVP = Fase 2 + US1.** Con eso hay un segundo proveedor conectado y custodiado, y la garantía
local sostenida por el compilador. US2 es lo que convierte eso en un beneficio para el usuario;
US3 lo que lo hace auditable.

## Notes

- [P] = archivos distintos, sin dependencias pendientes.
- Verificar que cada test falla antes de implementar. En T003 «falla» significa **no compila**.
- Commit atómico por tarea o grupo lógico (Conventional Commits; **sin atribución IA**).
- El secreto no se escribe en ningún log, ni siquiera durante el desarrollo.
