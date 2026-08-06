# Data Model — La experiencia de authoring (002)

Este feature **casi no crea dominio nuevo**: la mayor parte del modelo ya existe en `001` (`AuthoringSession`, `Segment`, `Groundedness`, `Reference`, `ChangeProposal`, `AuditEvent`). Aquí se registra (a) lo que se añade al núcleo y (b) las entidades de presentación que vive la piel.

## Añadido al núcleo

### `WriteRecord` — qué llegó al repositorio
Hace verificable FR-017 y FR-023 ("declarar qué alcanzó a escribirse").
- **Campos**: `path` (relativo al repo), `bytes`, `at`, `outcome` (`Written` | `Skipped` | `Failed(motivo)`).
- **Reglas**: se emite un registro **por artefacto**; el conjunto de registros es lo que la sesión reporta al cerrar, se haya completado o cancelado.
- **Relación**: `AuthoringSession 1—* WriteRecord`.

### `ProviderStatus` — resultado de sondear el backend
Alimenta el onboarding guiado (FR-019) y la presentación de fallos (FR-028).
- **Campos**: `reachable` (bool), `endpoint`, `models` (lista, vacía si no alcanzable), `issue` (`ProviderIssue?` — el motivo cuando no sirve).
- **`ProviderIssue`**: `NoModels` | `NotListening` | `EndpointNotLocal`. Es un **motivo nombrado, no una frase**: el núcleo no redacta texto para humanos, así que expone un código estable (`code()`) y la piel elige la cadena en `provider.issue.<code>`.
- **Reglas**: en modo local el `endpoint` **debe** ser loopback; sondear no abre una vía de salida. Si `reachable == false`, `issue` **nunca** falta — un fallo opaco es justo lo que FR-019 viene a evitar.

### `CancelOutcome` — cómo terminó una cancelación
- **Campos**: `at`, `phase` (en qué estaba la sesión), `writes` (los `WriteRecord` acumulados).
- **Regla**: cancelar lleva la sesión a `SessionState::Cancelled`, que **ya existe** en el dominio y ya es terminal.

### `AuditKind` — variantes nuevas
Se amplía el enum existente con `ArtifactWritten` y `SessionCancelled`, para que la piel se entere por el mismo canal que ya usa. No se altera ninguna variante previa (cambio aditivo).

## Entidades de presentación (viven en la piel)

Son **proyecciones de lectura**: no contienen reglas, solo lo necesario para renderizar.

### `Block` — la unidad de la corriente cronológica (FR-020)
- **Campos**: `seq` (orden), `at`, `kind`, `title`, `target?`, `detail?`.
- **`kind`**: `activity` | `resolved` | `unresolved` | `contradiction` | `written` | `egress-blocked` | `error` | `cancelled`.
- **Reglas**: la corriente es **append-only** — espeja la naturaleza del log de auditoría del que nace. Cada `kind` se distingue por **etiqueta + forma + color**, nunca solo por color (FR-026).

### `ArtifactView` — el artefacto completo (FR-021)
- **Campos**: `path`, `locale`, `segments`, `writeState` (¿está en disco?).
- **`segments`**: los `SegmentDto` que la piel **ya recibe** hoy, con `kind` (`grounded`/`tentative`/`contradiction`), `sources`, `reason`, `acknowledged`.
- **Regla**: alcanzable en cualquier momento sin recorrer la corriente hacia atrás.

### `ProviderPanel` — el estado del proveedor (FR-019)
- **Campos**: `status` (de `ProviderStatus`), `selectedModel`, `nextStep` (qué debe hacer el usuario si algo falta).
- **Regla**: cuando no hay backend, `nextStep` **nunca** puede quedar vacío: es lo que distingue "guiado" de "silencioso".

### `UiStrings` — el catálogo de cadenas (FR-016b)
- **Campos**: `locale` (`es` | `en`), `entries` (clave → texto).
- **Regla verificable (SC-009)**: **toda clave existe en ambos idiomas**. Es un test, no una revisión visual.

## Estados de sesión que la interfaz debe saber pintar

Reutiliza la máquina de estados existente de `001`; no se añaden estados nuevos:

`Ingesting` → `Generating` → `Refining` → `Approved`, con `Failed` y `Cancelled` alcanzables desde cualquier estado no terminal.

- Durante `Ingesting`/`Generating` la interfaz **sigue viva** y ofrece cancelar (FR-022/023).
- `Cancelled` y `Failed` deben mostrar el **balance de escrituras** (`WriteRecord`), no solo el rótulo del estado.
- Intentar cerrar con tentativos sin atender se advierte antes de llegar a `Approved` (FR-014 del spec 002 / invariante ya implementada en el dominio).
