# Contracts — Comandos y eventos de la piel

Extiende la superficie ya implementada en T029 (`001/contracts/tauri-commands.md`). La piel **renderiza y captura decisiones**; toda regla vive en el núcleo.

## Comandos (frontend → backend)

| Comando | Input | Output | Estado |
|---|---|---|---|
| `start_session` | `{ repoRoot, local, locale? }` | `{ sessionId }` **inmediato** | ⚠️ cambia: ya no espera a que termine el trabajo |
| `session_state` | `{ sessionId }` | `SessionSnapshotDto` | ⚠️ crece — ver abajo |
| `set_locale` | `{ sessionId, locale }` | `ack` | ya existe |
| `cancel_session` | `{ sessionId }` | `CancelOutcomeDto` | 🆕 FR-023 |
| `artifact` | `{ sessionId, path }` | `ArtifactDto` (**+ `writeState`**) | 🆕 FR-021 |
| `defer_tentative` | `{ sessionId, path, index }` | `number` (cuántos quedan) | 🆕 FR-014 |
| `submit_message` | `{ sessionId, message }` | `ProposalDto[]` | 🆕 FR-010 — **retorna cuando el turno está resuelto** |
| `pending_proposals` | — | `string[]` (ids) | 🆕 lo que el núcleo espera **ahora mismo** |
| `decide` | `{ proposalId, verdict, edited? }` | `ack` | 🆕 FR-014/FR-015 |
| `revert_proposal` | `{ sessionId, proposalId }` | `ack` | 🆕 FR-008 — **solo lo auto-aplicado** |
| `probe_provider` | `{ local }` | `ProviderStatusDto` | 🆕 FR-019 |
| `ui_strings` | `{ locale }` | `{ locale, entries }` | 🆕 FR-016b |

### `SessionSnapshotDto` — campos

`id` · `state` · `locale` · `artifacts` · `unresolved` · `omitted` · `budgetExhausted` ·
`interviewMode` · `unattendedTentative` · `writes` (FR-017) · `tierDegraded` (`001`-FR-018) ·
`failure` (FR-028, en curso).

Se enumeran porque la fila «**+ `writes`**» llevaba tiempo describiendo un DTO de once campos
como si tuviera uno más que antes. Un contrato que solo dice «crece» no permite notar que dejó
de ser cierto — y de hecho no lo permitió: `tierDegraded`, `interviewMode` y
`unattendedTentative` llevaban tiempo fuera.
| `system_locale` | — | `{ locale }` (`es`\|`en`, cae a `en`) | 🆕 FR-016b |

## Eventos (backend → frontend)

Los cuatro primeros **ya existen**; se añaden dos.

| Evento | Payload | Cuándo |
|---|---|---|
| `agent.activity` | `{ action, target }` | El agente listó, leyó, generó o detectó algo |
| `reference.resolved` | `{ action, target }` | Se leyó una fuente |
| `reference.unresolved` | `{ target, reason }` | No se pudo resolver — se declara, no se inventa |
| `egress.blocked` | `{ action, target }` | Salida bloqueada en modo local |
| `session.state_changed` | `{ state }` | Transición de la máquina de estados |
| `artifact.written` | `{ path, bytes, outcome }` | 🆕 Un artefacto llegó al repositorio (FR-017) |
| `session.cancelled` | `{ phase, writes }` | 🆕 La sesión se canceló y este es el balance (FR-023) |
| `proposal.new` | `{ id, target, unified, rationale, risk: `low`\|`high_impact` }` | 🆕 El núcleo **está bloqueado** esperando una decisión sobre este cambio |

## Reglas de la piel

- **Nada de lógica de dominio.** Qué es alto impacto, qué está fundamentado y qué se escribe lo decide el núcleo. La piel elige *cómo se ve*.
- **La corriente es append-only**: los bloques no se reescriben ni se reordenan. Espeja el log de auditoría del que nacen.
- **Los tres estados de fundamento se distinguen por etiqueta + forma + color** (FR-026). Quitar el color no puede volver la interfaz ambigua.
- **Toda acción tiene camino de teclado** (FR-025): iniciar, cancelar, recorrer la corriente, abrir un artefacto, cerrar la vista.
- **Ningún texto va incrustado en la vista**: todo sale del catálogo de `ui_strings` (FR-016b).
- **Ningún fallo se muestra crudo** (FR-028): se traduce a qué pasó y qué hacer.
- **El `Prompter` es la piel, y por eso el port es `async`**: `present()` emite `proposal.new` y **espera** en un canal que `decide` resuelve. El turno sigue siendo una unidad — `submit_message` retorna cuando todo quedó decidido. Si el canal se cierra sin respuesta (ventana cerrada, sesión cancelada), la propuesta **no** se aplica.
- **El panel de decisión vive fuera de la corriente**: la corriente es append-only y registra qué pasó; lo que espera decisión cambia. Mezclarlos obligaría a reescribir bloques ya emitidos.
- **Deshacer es solo para lo auto-aplicado**: si el producto se permite no consultar por lo de bajo riesgo, el precio es que deshacerlo esté a mano. Lo que pasó por una decisión humana se cambia **decidiendo otra vez**, no deshaciéndolo a espaldas de quien lo aprobó.
- **Todo enum que cruza a la ventana viaja como código estable**, nunca derivado de `Debug`: `RiskLevel` (`low` | `high_impact`), `SessionState` (`ingesting` | `generating` | `refining` | `approved` | `failed` | `cancelled`), `ReferenceState` (`resolved` | `inaccessible` | `requires_auth` | `out_of_scope`) y `ProviderIssue`. Atar el contrato al nombre de una variante lo rompería al renombrarla, **sin aviso** — y en el caso de `SessionState` la ventana compone con él la clave del catálogo (`session.state.<code>`), así que un renombrado dejaría la clave cruda en pantalla.
- **Diferir es por fragmento**: `defer_tentative` exige un índice concreto. No existe «diferir todo» — despachar sin leer es el hábito que el producto corrige, no uno que facilite.
- **Un archivo en pantalla no es un archivo en el repositorio**: `writeState` (`written` | `pending` | `failed` | `skipped`) viaja con el artefacto, y la vista lo declara.

## Contrato de estados de la interfaz

| Estado de sesión | Qué muestra la corriente | Acciones disponibles |
|---|---|---|
| *(sin sesión)* | Invitación + estado del proveedor | iniciar, configurar proveedor |
| `ingesting` / `generating` | Bloques en vivo | **cancelar**, abrir artefacto ya generado, desplazarse |
| `refining` | Bloques + propuestas *(US2, fuera de este alcance)* | — |
| `approved` | Balance final + artefactos escritos | abrir artefactos, iniciar otra sesión |
| `cancelled` / `failed` | **Balance de escrituras** + motivo | abrir lo que sí se escribió, reintentar |
