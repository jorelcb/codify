# Contracts — Comandos Tauri (piel fase 1)

Superficie del **driving adapter** de la piel Tauri: comandos que el frontend web invoca (`invoke`) y eventos que el backend emite al frontend. Cada comando delega en `AuthoringService` (ver `ports.md`); la piel **no** contiene lógica de dominio.

## Comandos (frontend → backend)

| Comando | Input | Output | Delega en |
|---|---|---|---|
| `start_session` | `{ repo_root, mode: "local"\|"hybrid", locale? }` | `{ session_id }` | `AuthoringService::start_session` |
| `session_state` | `{ session_id }` | `SessionState` (`ingesting`\|`generating`\|`refining`\|`approved`\|…) | `session_state` |
| `submit_message` | `{ session_id, message }` | `ack` | `submit_message` (loop de refinamiento) |
| `pending_proposals` | `{ session_id }` | `[ChangeProposal]` (con `risk`, `diff`) | `pending_proposals` |
| `decide` | `{ session_id, proposal_id, verdict: "approve"\|"edit"\|"reject", edited_diff? }` | `ack` | `decide` |
| `set_locale` | `{ session_id, locale }` | `ack` | `set_locale` (override, FR-019) |
| `connect_provider` | `{ backend, endpoint?, auth: "local"\|"api_key"\|"oauth" }` | inicia flujo (device-code si oauth) | infra `providers/` (D3) |
| `list_connections` | `—` | `[ModelConnection]` (con `is_local`) | infra |
| `set_mode` | `{ session_id, mode }` | `ack` (revalida elegibilidad de proveedores) | infra |

## Eventos (backend → frontend)

| Evento | Payload | Cuándo |
|---|---|---|
| `session.state_changed` | `{ session_id, state }` | Transición de la máquina de estados |
| `agent.token` | `{ session_id, text }` | Streaming de la respuesta del modelo (UX fluida) |
| `agent.activity` | `{ session_id, action, target }` | El agente listó/leyó/trajo algo (transparencia de ingesta D4) |
| `proposal.new` | `{ session_id, proposal }` | Nuevo diff propuesto (Low = auto-aplicado y mostrado; HighImpact = requiere decisión) |
| `reference.unresolved` | `{ session_id, reference, state }` | Referencia inaccesible/requires-auth (FR-004) |
| `egress.blocked` | `{ session_id, host }` | Intento de salida bloqueado en modo local (auditoría SC-007) |

## Reglas de la piel
- La piel **renderiza** diffs y **captura** aprobaciones; nunca genera/aplica el diff (eso es del core, D6).
- Para `HighImpact`, la UI **bloquea** la escritura hasta la decisión; para `Low`, muestra el cambio ya aplicado con opción de **revertir**.
- La piel implementa el port `Prompter` traduciendo `ask`/`present` a interacciones de UI.
