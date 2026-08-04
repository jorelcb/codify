# Implementation Plan: La experiencia de authoring — ver, entender y decidir

**Branch**: `002-authoring-experience` | **Date**: 2026-07-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-authoring-experience/spec.md`

## Summary

Construir la experiencia de la aplicación de escritorio: **flujo cronológico de bloques** donde el usuario ve trabajar al agente, revisa cambios y distingue sin ambigüedad qué está fundamentado, qué es tentativo y qué se contradice.

El hallazgo que domina este plan: **dos de los requisitos más consecuentes no son trabajo de interfaz sino del núcleo.** No se puede "dejar claro qué se escribió al repositorio" (FR-017) porque **hoy el núcleo no escribe nada** — genera los artefactos en memoria y ahí se quedan. Y no se puede cancelar una sesión (FR-023) porque el loop corre entero hasta terminar. Ambas cosas se planifican aquí como **prerrequisitos del núcleo**, antes de tocar un solo pixel.

## Technical Context

**Language/Version**: Rust (edition 2021) para núcleo y piel. Frontend en HTML/CSS/JavaScript sin transpilación.

**Primary Dependencies**:
- Núcleo: las existentes (`tokio`, `serde`, `reqwest`, `similar`, `walkdir`, `async-trait`, `thiserror`) **+ `tokio-util`** (solo en `infrastructure`, para el token de cancelación).
- Piel: `tauri` v2 (ya integrado). **Sin framework de frontend y sin bundler** — decisión razonada en `research.md`.

**Storage**: Archivos del repositorio objetivo. Sin base de datos y **sin persistencia de sesión** (FR-024).

**Testing**: `cargo test`. Unit del dominio; contract tests por port (real + fake); integración del loop con fakes. La **completitud del catálogo de cadenas** (SC-009) se verifica con un test de Rust, no a ojo. El DOM se valida con los escenarios de `quickstart.md`.

**Target Platform**: Escritorio (macOS/Linux/Windows) vía Tauri v2.

**Project Type**: desktop-app sobre library crate (workspace Cargo ya existente).

**Performance Goals**: Interfaz **nunca bloqueada** mientras el agente trabaja (FR-022). La cancelación surte efecto **sin esperar a que termine la llamada al modelo en vuelo**.

**Constraints**: La piel no contiene lógica de dominio (constitución I). Toda acción alcanzable por teclado (FR-025). Sin dependencia del color (FR-026). Cero-egress estructural intacto.

**Scale/Scope**: Un usuario, un repositorio por sesión. La corriente de bloques debe seguir siendo utilizable con cientos de acciones.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Ratificada contra `.specify/memory/constitution.md` **v1.0.0**:

| Principio | Cumplimiento |
|---|---|
| **I. Regla de Dependencia** | ✅ Los ports nuevos (`Cancellation`, `ArtifactWriter`, `ProviderDiscovery`) se declaran en `application/ports.rs` — son capacidades que el Dominio no nombra. Sus adapters viven en `infrastructure/`. La piel sigue siendo Interface Adapter. La fitness function existente lo verifica. |
| **I. Firmas con tipos de dominio** | ✅ Ningún tipo de `tokio-util`, `reqwest` ni Tauri cruza un port. `Cancellation` se expresa como trait propio; el `CancellationToken` concreto queda en `infrastructure`. |
| **II. Test-first + Test Desiderata** | ⚠️ **Desviación documentada**: el frontend no tendrá test automatizado en esta etapa (ver Complexity Tracking). Todo lo demás va test-first, incluido el catálogo de cadenas. |
| **III. Conventional Commits → SemVer** | ✅ Sin atribución de IA. Los cambios de ports son aditivos (MINOR en 0.y.z). |
| **Proyecto — Cero-egress** | ✅ Intacto: la sonda de proveedor solo consulta endpoints loopback en modo local. |
| **Proyecto — Greenfield** | ✅ Sin compatibilidad hacia atrás que preservar. |

**Resultado del gate**: PASS con una desviación justificada (ver Complexity Tracking).

## Project Structure

### Documentation (this feature)

```text
specs/002-authoring-experience/
├── plan.md              # Este archivo
├── research.md          # Fase 0 — decisiones técnicas y alternativas
├── data-model.md        # Fase 1 — entidades de la experiencia
├── quickstart.md        # Fase 1 — escenarios de validación
├── contracts/           # Fase 1
│   ├── core-ports.md        #   ports nuevos del núcleo
│   ├── skin-commands.md     #   comandos y eventos de la piel
│   └── ui-strings.md        #   catálogo de cadenas y contrato de localización
└── tasks.md             # Fase 2 — /speckit-tasks (NO lo crea /speckit-plan)
```

### Source Code (repository root)

Se extiende el workspace existente. **Lo marcado 🆕 es nuevo; el resto ya existe.**

```text
crates/
├── codify-core/src/
│   ├── domain/
│   │   └── write.rs              # 🆕 WriteRecord (el Dominio lo nombra: la sesión lo reporta)
│   ├── application/
│   │   ├── ports.rs              # 🆕 + Cancellation, ArtifactWriter, ProviderDiscovery
│   │   ├── authoring_loop.rs     # 🆕 respeta la cancelación y escribe los artefactos
│   │   ├── deps.rs               # 🆕 + los tres ports nuevos
│   │   └── service.rs            # 🆕 start_session no bloquea; + cancel_session
│   └── infrastructure/
│       ├── cancel.rs             # 🆕 TokenCancellation (tokio-util)
│       ├── repo/writer.rs        # 🆕 FsArtifactWriter
│       ├── providers/probe.rs    # 🆕 sonda del backend local
│       └── composition.rs        # 🆕 cablea los tres ports nuevos
└── codify-app/
    ├── src/
    │   ├── commands.rs           # 🆕 + cancel_session, probe_provider, ui_strings, artifact
    │   ├── strings.rs            # 🆕 catálogo es/en + test de completitud
    │   └── adapters.rs           # (sin cambios de fondo)
    └── ui/
        ├── index.html            # 🆕 estructura semántica y regiones
        ├── main.js               # 🆕 orquestación, atajos de teclado
        ├── stream.js             # 🆕 la corriente cronológica
        ├── artifact.js           # 🆕 vista de artefacto completo (FR-021)
        ├── provider.js           # 🆕 onboarding guiado (FR-019)
        ├── i18n.js               # 🆕 consumo del catálogo de cadenas
        └── styles.css            # 🆕 estados con forma + etiqueta + color, no solo color
```

**Structure Decision**: no se introducen crates nuevos. Los tres ports nuevos son **application capability ports** (el Dominio no los nombra) y sus adapters entran en `infrastructure/`. La piel crece en módulos pequeños por responsabilidad, sin framework.

## Decisiones de diseño principales

### 1. Cancelación (FR-022/FR-023) — el punto de mayor impacto

`start_session` deja de correr el loop completo antes de retornar. Pasa a:

1. **Retornar de inmediato** con el `SessionId`, dejando el loop corriendo en una tarea.
2. La corriente de eventos (que ya existe, vía `AuditSink`) es lo que informa el avance — no hizo falta inventar nada nuevo para eso.
3. Un port **`Cancellation`** se inyecta en el loop. El loop lo consulta en cada punto de control **y lo compone con `tokio::select!` alrededor de la llamada al modelo**, de modo que cancelar **aborta la petición en vuelo** en vez de esperar a que termine.
4. Al cancelar, la sesión pasa a un estado terminal que **declara qué artefactos alcanzaron a escribirse**.

El port es un trait propio del núcleo; el `CancellationToken` de `tokio-util` vive solo en `infrastructure/cancel.rs`. Alternativas y su descarte: `research.md`.

### 2. Escritura de artefactos — deuda de 001 que aflora aquí

Hoy `generate()` termina en `session.put_artifact()` y nada llega al disco. Se añade el port **`ArtifactWriter`** (`write`, `read_existing`) con adapter de filesystem, y el loop lo invoca al cerrar la generación, registrando cada escritura en el log de auditoría — que es como la piel se entera (FR-017).

> Esto cubre las escrituras del **primer** paso de generación. La regla de "no sobrescribir sin diff y aprobación" (FR-014 de 001) es la User Story 3 de aquel spec y **no** se implementa aquí; el writer se diseña para soportarla — `read_existing` existe exactamente para eso.

### 3. Onboarding del proveedor (FR-019)

Port **`ProviderDiscovery`** con `probe()` que devuelve si el backend responde y qué modelos ofrece. El adapter consulta el endpoint OpenAI-compatible; en modo local solo acepta loopback, así que la garantía de cero-egress no se debilita. La piel expone `probe_provider` y presenta el resultado como estado accionable (FR-028).

### 4. Frontend sin framework — decisión, no inercia

**Se mantiene vanilla.** La interfaz son tres superficies (corriente, vista de artefacto, panel de proveedor) sin estado compartido complejo; el estado real vive en el núcleo. Introducir un framework añadiría npm, un bundler y Node al CI —hoy inexistentes— para resolver un problema que aún no tenemos. Criterio explícito para revisitarlo en `research.md`.

### 5. Localización (FR-016b / SC-009)

El catálogo de cadenas vive **en Rust** (`strings.rs`) y se expone con el comando `ui_strings(locale)`. Así la completitud del catálogo es **verificable con un test** ("toda clave existe en es y en en"), en vez de depender de revisar la interfaz a ojo. El idioma de la interfaz es independiente del de los artefactos.

## Complexity Tracking

| Elemento | Por qué se acepta | Alternativa más simple rechazada porque |
|---|---|---|
| **Sin test automatizado del frontend** (desviación del Principio II) | No hay toolchain de JS en el proyecto, y añadirla (Node + runner + CI) para un DOM de pocos cientos de líneas cuesta más de lo que protege. **Mitigación**: toda la lógica testeable (catálogo de cadenas, DTOs, mapeo de fundamento) vive en Rust y sí va test-first; el DOM queda deliberadamente delgado y se valida con los escenarios de `quickstart.md` | Montar un runner de JS ahora añade dependencias y tiempo de CI antes de saber siquiera si la forma de la interfaz es la correcta. **Señal para revisitarlo**: en cuanto el frontend contenga alguna decisión que no sea render puro |
| **`tokio-util` como dependencia nueva** | La cancelación de I/O en vuelo con `select!` es lo que hace que "cancelable en cualquier momento" sea cierto y no un eufemismo | Un `AtomicBool` consultado en puntos de control: cero dependencias, pero cancelar tardaría lo que tarde la llamada al modelo en curso (decenas de segundos). Incumple el espíritu de FR-023 |
| **`start_session` cambia de semántica** (retorna antes de terminar) | Es la única forma de que la interfaz siga viva durante una sesión de minutos | Mantenerlo bloqueante y "resolver" la responsividad en la piel: imposible, el trabajo ocurre dentro de la llamada |
