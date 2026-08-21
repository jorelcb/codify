# Quickstart — Validación end-to-end (Context Authoring)

Guía de validación que prueba el loop de contexto. No incluye implementación; referencia `spec.md`, `data-model.md` y `contracts/`.

## Prerrequisitos
- Toolchain Rust + Tauri v2 dev deps.
- Un backend de modelo **local**: Ollama (o llama.cpp-server) corriendo con un modelo instalado.
- Fixture: `tests/fixtures/lumen-like/` — repo con `README.md` que **referencia** un `SPEC` local (que declara "sin broker, event-sourced sobre Temporal") y una URL pública; opcionalmente una URL privada para el caso RequiresAuth.
- Un monitor de red para el aserto de cero-egress (o el `AuditSink` con eventos `egress.blocked`).

## Comandos
```bash
cargo test -p codify-core                 # unit + contract tests por port
cargo test --test integration             # loop end-to-end con fakes
cargo tauri dev                            # levanta la piel para validación manual
```

## Cobertura automatizada

Los cinco escenarios están cubiertos por tests de integración del núcleo. Lo que **sigue
necesitando una persona con un modelo real** es la *calidad de la salida*, no el mecanismo: un
proveedor con guion no puede demostrar SC-001 («≥90 % consistente con la SSOT»).

| Escenario | Tests que lo cubren | Qué queda para el humano |
|---|---|---|
| S1 · Grounded + cero-egress | `us1_grounded.rs` (5), `egress_guard.rs` (8) | leer el `CONTEXT` generado por un modelo real y juzgar si es fiel al SPEC |
| S2 · Refinamiento curado | `us2_refine.rs` (6), `us2_scaffolding.rs` (4), `us2_revert.rs` (5) | que la conversación *fluya*: eso no lo mide un guion |
| S3 · Contexto previo | `us3_update.rs` (7) | — cubierto |
| S4 · Referencia no resuelta | `us1_grounded.rs::unresolved_private_reference…`, `contract_ports.rs::requires_auth…` | — cubierto |
| S5 · Idioma | `us1_grounded.rs::locale_is_autodetected_and_can_be_overridden` | — cubierto |

**Fixture**: `./scripts/quickstart-fixture.sh` lo genera. El README referencia un SPEC hermano
que **contradice** lo que un modelo asumiría leyendo solo el README — si el contexto generado
menciona un broker de mensajes, el agente se lo inventó.

## Escenarios

### S1 — Grounded + cero-egress (US1, SC-001/002/006/007)
1. Conectar solo un proveedor **local**; iniciar sesión en `mode = local` sobre el fixture.
2. **Esperado**: el agente lista/lee el repo, sigue el README→SPEC (evento `agent.activity`), y el `CONTEXT` generado refleja "**no broker / event-sourced / Temporal**" — **no** una arquitectura genérica inventada.
3. **Esperado**: lo no verificado queda marcado **tentativo** (distinguible); referencias no resueltas **declaradas**, no inventadas.
4. **Esperado (cero-egress)**: 0 conexiones salientes a hosts no-locales; cualquier intento → `egress.blocked` en el audit.

### S2 — Refinamiento curado con diffs (US2, FR-010/012)
1. Partir de un contexto con un hueco y un supuesto incorrecto.
2. Corregir en lenguaje natural; **esperado**: el agente propone un diff que integra la corrección **y ajusta el andamiaje dependiente** (nombres/secciones), no solo el hueco.
3. Un cambio `HighImpact` **bloquea** hasta aprobación; un cambio `Low` se **auto-aplica** y es **revertible**.
4. **Rechazar** un diff → el archivo **no** cambia (FR-014/015).

### S3 — Repo con contexto previo (US3, SC-005)
1. Re-ejecutar sobre un repo que ya tiene `AGENTS.md`.
2. **Esperado**: se presenta un **diff de actualización** con aprobación; **0 sobrescrituras silenciosas**; contenido humano preservado hasta decisión.

### S4 — Referencia no resuelta (FR-003/004, SC-006)
1. El README referencia una **URL privada** (requiere auth).
2. **Esperado**: se reporta `reference.unresolved: requires_auth`; el contenido faltante **no** se fabrica; el contexto marca ese punto como tentativo.

### S5 — Idioma (FR-019)
1. Fixture mayoritariamente en español → **esperado**: contexto generado en español (auto-detección).
2. `set_locale(en)` → **esperado**: regenera/ajusta en inglés (override).

## Criterios de aprobación (mapeo a Success Criteria)
- S1 ⇒ SC-001 (≥90% consistente con SSOT), SC-002, SC-006, SC-007.
- S2 ⇒ SC-004 (conversación + diffs, 0 marcadores sin atender al cerrar).
- S3 ⇒ SC-005 (0 clobbers).
- S1–S5 en `mode = local` ⇒ offline-capable.
