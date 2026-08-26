# Contracts — Comandos y eventos que la piel gana

Extiende la superficie de `002/contracts/skin-commands.md`.

| Comando | Input | Output | Estado |
|---|---|---|---|
| `connect_provider` | `{ provider, tier }` | `ConnectChallengeDto` \| `ack` | 🆕 FR-001 |
| `complete_connection` | `{ challengeId, secret? }` | `ProviderConnectionDto` | 🆕 FR-001 |
| `list_connections` | — | `ProviderConnectionDto[]` | 🆕 FR-003 |
| `disconnect_provider` | `{ connectionId }` | `ack` | 🆕 FR-003 |
| `set_mode` | `{ mode }` | `ack` | 🆕 FR-008a — **rearma el grafo** |

### `ProviderConnectionDto` — campos

`id` · `label` · `tier` (`cheap` \| `heavy`) · `state` (`connected` \| `expired` \| `revoked`) ·
`endpointHost`.

**Entregado.** El comando de conexión devuelve además un `ConnectChallengeDto` con
`challengeId`, `kind` (`delegada` \| `credencial`), y según la vía `code`+`url` o
`instructions`. Ninguno de los dos tiene campo para el secreto.

**No lleva la credencial, ni un campo donde pudiera ir.** Se enumeran los campos, y no se dice
«crece», por lo aprendido en `002`: un contrato que solo dice que crece no permite notar que dejó
de ser cierto.

`endpointHost` es el **host**, no la URL con credenciales embebidas: es lo que FR-009 necesita
para que el usuario vea qué proveedores podrían recibir contenido, sin exponer más.

### Eventos

| Evento | Cuándo | Contenido |
|---|---|---|
| `connection.state_changed` | conectada, caducada o revocada | id y estado nuevo |
| `mode.changed` | tras rearmar el grafo | modo nuevo, y que la sesión viva no cambia (FR-008b) |
| `task.routed` | cada tarea atendida | qué tier y qué conexión — la base de FR-006 y FR-010 |

### Reglas

1. Ningún comando ni evento MUST llevar el secreto. El DTO no tiene campo para él, que es más
   fiable que acordarse de no rellenarlo.
2. `set_mode` MUST NOT afectar a una sesión en curso (FR-008b) y MUST NOT exigir reiniciar
   (SC-007).
3. El catálogo gana `connection.*` y `mode.*` en **los dos idiomas**, con el test que ya recorre
   los códigos del núcleo — el mismo patrón de `SessionFailure` y `ProviderIssue`.
