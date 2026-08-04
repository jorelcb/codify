# Research — La experiencia de authoring (002)

Fase 0. Decisiones técnicas con su racional y las alternativas evaluadas. **No quedan NEEDS CLARIFICATION**: las cuatro aclaraciones abiertas se cerraron en `/speckit-clarify` y están en el spec.

## D-1 · Cómo se cancela una sesión en curso

- **Decisión**: port propio **`Cancellation`** en `application/ports.rs`, con dos métodos: `is_cancelled()` (consulta barata en puntos de control) y `cancelled()` (futuro que resuelve al cancelar). El loop lo compone con `tokio::select!` alrededor de la llamada al modelo. El `CancellationToken` de `tokio-util` vive **solo** en `infrastructure/cancel.rs`.
- **Rationale**: es lo único que hace verdadera la promesa de FR-023. Una llamada de generación a un modelo local puede tardar decenas de segundos; sin abortar el I/O en vuelo, "cancelar" significaría "quedarse mirando la pantalla un rato más". Mantener el port abstracto conserva el núcleo testeable con un fake y evita que `tokio-util` cruce hacia `application`.
- **Alternativas**:
  - *`Arc<AtomicBool>` consultado en puntos de control*: cero dependencias y más simple, pero la cancelación solo surte efecto cuando vuelve la llamada en curso. Rechazada por incumplir el espíritu del requisito.
  - *Abortar la tarea de tokio desde fuera (`JoinHandle::abort`)*: brutal y, sobre todo, **deja al usuario sin saber qué alcanzó a escribirse** — justo lo que FR-023 exige informar.
  - *Exponer `CancellationToken` directamente en el port*: filtraría un tipo de infraestructura a través de la frontera del núcleo (constitución I).

## D-2 · `start_session` deja de ser bloqueante

- **Decisión**: `start_session` valida, arranca el loop en una tarea y **retorna el `SessionId` de inmediato**. El avance se sigue por los eventos que ya emite el `AuditSink`; el resultado final se consulta con `session_state`.
- **Rationale**: es la única forma de cumplir FR-022. El trabajo ocurre *dentro* de la llamada actual, así que ninguna técnica del lado de la piel puede desbloquear la interfaz.
- **Consecuencia asumida**: el contrato del comando cambia. Como el proyecto es greenfield y la única consumidora es nuestra propia piel, no hay compatibilidad que preservar.
- **Alternativas**: *streaming del resultado por el propio comando* (más complejo en Tauri y redundante: el canal de eventos ya existe y ya funciona).

## D-3 · Escritura de artefactos al repositorio

- **Decisión**: port **`ArtifactWriter`** con `write(path, content)` y `read_existing(path)`. Adapter de filesystem que respeta la raíz del repositorio (mismas defensas que el navegador: nada de rutas absolutas ni escapes con `..`). Cada escritura se registra en el log de auditoría.
- **Rationale**: **el núcleo no escribía nada** — los artefactos se quedaban en memoria. Sin esto, FR-017 ("dejar claro qué se escribió") no tiene objeto y el producto no entrega su resultado.
- **`read_existing` desde el principio**: la User Story 3 de 001 (no sobrescribir sin diff y aprobación) lo necesitará. Añadirlo ahora cuesta una línea; retrofitearlo obligaría a cambiar el port.
- **Alternativas**: *que la piel escriba los archivos* — rechazada de plano: pondría una regla de dominio (qué se escribe y dónde) en un Interface Adapter, contra la constitución.

## D-4 · Descubrimiento del proveedor de modelo

- **Decisión**: port **`ProviderDiscovery`** con `probe()` → estado (alcanzable / no alcanzable) + modelos disponibles. El adapter consulta el endpoint OpenAI-compatible (`/v1/models`, con respaldo en la API nativa de Ollama).
- **Rationale**: FR-019 exige explicar qué falta en vez de quedarse en silencio. La sonda tiene que ser un port porque el resultado alimenta una decisión de la aplicación, no solo un mensaje.
- **Cero-egress**: la sonda reutiliza la misma validación de loopback del proveedor local, así que no abre una vía de salida por la puerta de atrás.
- **Alternativas**: *intentar una generación de prueba* — cara y con efectos; *no sondear y esperar al primer fallo* — es exactamente la experiencia que FR-019 viene a evitar.

## D-5 · Frontend sin framework

- **Decisión**: **se mantiene HTML/CSS/JS plano, sin bundler ni npm.** Módulos ES nativos (`<script type="module">`) para separar responsabilidades.
- **Rationale**: la interfaz son tres superficies (corriente, artefacto, proveedor) y **el estado vive en el núcleo**, no en el cliente; el frontend es render + captura de eventos. Un framework traería npm, un bundler y Node al CI —hoy inexistentes— a cambio de resolver un problema que todavía no tenemos. La constitución premia la simplicidad justificada, no la ceremonia.
- **Criterio explícito para revisitarlo** (cualquiera de los tres):
  1. aparece estado de cliente que haya que sincronizar entre varias vistas;
  2. la corriente necesita virtualización por volumen;
  3. el DOM a mano supera ~600 líneas o empieza a repetir patrones de componente.
- **Alternativas**: *Svelte/Solid* (excelente ergonomía, pero añaden build); *htmx* (pensado para HTML sobre la red, no para una app local dirigida por eventos); *web components sin librería* (podría entrar si se cumple el criterio 3, sin salir de "sin bundler").

## D-6 · Localización verificable

- **Decisión**: el catálogo de cadenas vive **en Rust** (`codify-app/src/strings.rs`) y se expone por el comando `ui_strings(locale)`. La interfaz toma el idioma del sistema por defecto y permite cambiarlo.
- **Rationale**: SC-009 pide **cero cadenas sin traducir**, y eso solo es verificable si el catálogo es un dato inspeccionable por un test ("toda clave existe en ambos idiomas"). Si las cadenas viven sueltas en el DOM, el criterio se vuelve una revisión manual — justo lo que la constitución evita.
- **Alternativas**: *archivos JSON en `ui/`* (se podrían testear, pero quedan fuera del alcance de `cargo test` y del gate de CI); *librería i18n de JS* (contradice D-5).

## D-7 · Accesibilidad sin depender del color

- **Decisión**: cada estado de fundamento se comunica con **tres señales redundantes**: etiqueta textual, forma/icono, y color. Estructura semántica (`<article>`, `<h*>`, regiones con `aria-live` para la corriente). Toda acción con atajo de teclado y foco visible.
- **Rationale**: FR-026 exige sobrevivir a daltonismo y a una captura en escala de grises; la etiqueta textual es lo que lo garantiza de verdad. La región `aria-live` es lo que convierte una corriente que crece en algo que un lector de pantalla puede anunciar.
- **Alternativas**: *solo color con buen contraste* — incumple el requisito; *ARIA exhaustivo* — desproporcionado sin auditoría real con lectores de pantalla (fuera de alcance por decisión del spec).

## Unknowns resueltos

- **Idioma del sistema en Tauri**: se obtiene en el arranque y se pasa al frontend; si no es `es` ni `en`, se cae a `en`.
- **Volumen de la corriente**: sin virtualización en v1. Con el presupuesto de ingesta actual el orden de magnitud son decenas de bloques, no miles. Si crece, entra el criterio 2 de D-5.
- **Estado terminal por cancelación**: se reutiliza `SessionState::Cancelled`, que **ya existe** en el dominio y ya es terminal — no hace falta tocar la máquina de estados.
