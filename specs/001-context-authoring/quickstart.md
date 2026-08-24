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

## Hallazgos de la primera pasada con modelo real

Ejecutada el 2026-08-23 con `llama.cpp server` + `qwen2.5-7b-q4km`, sobre el fixture, **tres
corridas**. El arnés que la reproduce es `crates/codify-core/tests/live_backend.rs`
(marcado `#[ignore]`: necesita un backend vivo, CI nunca lo corre).

### Lo que funciona

El agente **sigue la referencia**: `ReferenceResolved: docs/SPEC-30.md` en las tres corridas, y
en las tres reprodujo las negaciones del SPEC («no message broker», «no event-sourcing»). **La
regresión que originó el proyecto no se reproduce.** Los cuatro artefactos llegan a disco.

### Lo que no

| # | Hallazgo | Gravedad |
|---|---|---|
| **F-1** | El sistema **atribuye a una fuente algo que esa fuente no dice**. Se registró la contradicción `[docs/PRD.md vs Makefile] «el Makefile solo soporta PostgreSQL 16»` — el `Makefile` del fixture tiene dos líneas y **cero** menciones a PostgreSQL. Causa de diseño: `parse_segments` degradaba a tentativo lo que llegaba *sin* procedencia, pero **no verificaba la procedencia que sí llegaba** | ✅ **cerrado en la Fase 7**: la cita se comprueba contra el material leído (`us1_provenance.rs`) |
| **F-2** | La sesión puede terminar en `Failed` con **cero artefactos y sin motivo** en el snapshot (1 de 3 corridas, 124 s). FR-028 exige explicar qué pasó y qué hacer | alta |
| **F-3** | Contexto en **idioma mezclado**: fixture íntegramente en español, salida mayormente en inglés con frases sueltas en español. FR-019 / S5 | media |
| **F-4** | **Contenido duplicado**: el mismo bloque de seis frases repetido dentro de un artefacto | media |
| **F-5** | La **URL privada nunca se intenta** (`unresolved: []`), así que **S4 no llega a ejercitarse**. El escenario existe pero el flujo no lo dispara | media |

## Segunda pasada — verificación de procedencia (Fase 7)

Ejecutada el 2026-08-23 con `llama.cpp server` + **`Qwen2.5-32B-Instruct-Q4_K_M`**, regenerando
el fixture antes de cada corrida. **Tres corridas completas, las tres en verde.**

| | fundamentados | tentativos | referencias | contradicción | citas sin respaldo |
|---|---|---|---|---|---|
| 1 | 11 | 15 | 3 | ✅ | **0** |
| 2 | 24 | 5 | 3 | ✅ | **0** |
| 3 | 14 | 11 | 3 | ✅ | **0** |

**F-1 no se reproduce.** Ninguna afirmación presentada como verificada se apoya en algo que las
fuentes no dicen, y la comprobación no es vacua: entre 11 y 24 segmentos sobreviven fundamentados
por corrida. Ese contraste importa — una tubería rota que degradara todo daría cero citas sin
respaldo exactamente igual, y por eso el arnés exige además que **algo** quede en pie.

La dispersión entre corridas (de 11/15 a 24/5) es del modelo, no del mecanismo. Es la razón de
que SC-001 se mida sobre varias corridas y no sobre una.

### Tres defectos que salieron por el camino

Ninguno estaba en el plan de la fase, y los tres se arreglaron antes de cerrarla:

1. **Fixture contaminado.** Encadenar corridas sin regenerar hace que el agente lea los
   artefactos de la anterior como si fueran del repositorio. Una corrida se fundamentó en
   `context/CONTEXT.md`, su propia salida previa. El arnés ahora lo aborta. **Queda abierta la
   pregunta de producto**: en uso real `context/` existe de sesiones anteriores, y el agente lo
   leerá igual.
2. **El arnés confundía «contradicho» con «inventado».** Tenía `dynamodb` en la lista de
   términos que delatan invención, pero `docs/PRD.md` lo afirma: es la contradicción deliberada
   del fixture. Y escrutaba el render completo, marcadores incluidos, así que **señalar** un
   conflicto contaba como cometerlo. Ahora solo mira los segmentos fundamentados.
3. **Timeout de 120 s en el proveedor local.** Cortaba generaciones sanas —690 tokens a 5,8 t/s
   cuando saltó el reloj— y la sesión caía a `Failed`. La petición de citas alarga la salida, así
   que **esta misma fase hizo el fallo más frecuente**. Subido a 900 s.

El tercero costó cinco corridas perdidas, y no por difícil: `Failed` no dice por qué (**F-2**,
issue #24). Un timeout de cliente es trivial de diagnosticar cuando se nombra.

### Advertencia sobre el instrumento

El arnés dio **dos falsos positivos** antes de servir: primero buscaba subcadenas (marcaba «**no**
hay broker» como invención), luego le faltaba «discarded» porque el modelo respondió en inglés y
la lista de negaciones estaba en español. La versión actual mide **afirmación, no mención**, pero
el mismo sesgo puede seguir ahí: estos hallazgos merecen confirmarse a ojo antes de actuar.

### Determinismo

Tres corridas, tres resultados distintos (una limpia, una fallo total, una con contradicciones
fabricadas). **SC-001 («≥90 % consistente con la SSOT») necesita medirse sobre varias corridas**,
no sobre una.

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
