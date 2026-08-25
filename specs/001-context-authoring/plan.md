# Implementation Plan: Context Authoring — de repositorio a contexto vivo (codify-NG)

**Branch**: `001-context-authoring` | **Date**: 2026-07-27 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/001-context-authoring/spec.md`

## Summary

Loop de authoring de contexto de codify-NG: un agente **lee el repo y sigue sus referencias** (docs + estructura + muestreo selectivo de código; referencias locales + URLs públicas), **genera** los artefactos de contexto en el idioma auto-detectado, y los **refina en un loop conversacional curado** donde el usuario aprueba diffs. Se entrega como **app Tauri standalone** (fase 1) construida sobre un **core Rust hexagonal** reutilizable (crate de biblioteca) que habilita pieles futuras (MCP, CLI) sin reescribir el dominio. Conectividad a modelos **multi-backend** (locales llama.cpp/Ollama + remotos vía OAuth), con **modo 100% local de cero-egress garantizado estructuralmente**.

## Technical Context

**Language/Version**: Rust (edition 2021; MSRV a fijar, objetivo ≥ 1.79). El core y la piel Tauri son 100% Rust; el frontend de Tauri es web (HTML/CSS/TS) servido por el WebView.

**Primary Dependencies**:
- **Core (crate `codify-core`)**, dependency-light: `tokio` (async), `serde`/`serde_json` (schemas), `reqwest` (HTTP a proveedores locales/remotos), `similar` (diff), `oauth2` (device-authorization flow), `keyring` (tokens en el keychain del SO). Sin framework de agentes.
- **Piel (crate `codify-app`)**: **Tauri v2** (shell Rust + WebView). Linkea `codify-core` in-process.

**Storage**: Archivos locales — los artefactos de contexto viven en el repo objetivo del usuario. Config de la app + credenciales OAuth en el **keychain del SO** (`keyring`) y config dir estándar. **Sin base de datos** (coherente con el ethos del propio dominio Lumen que auditamos).

**Testing**: `cargo test` — unit del dominio (puro, sin I/O), **contract tests por port** parametrizados por adapter (real vs fake in-memory, patrón hex-integration-test), e integración del loop con proveedor y filesystem fake. GUI: pruebas mínimas de los comandos Tauri (driving adapter).

**Target Platform**: Desktop (macOS/Linux/Windows vía Tauri v2). El core es platform-agnóstico.

**Project Type**: **desktop-app** (Tauri) sobre una **library crate** reutilizable → workspace Cargo multi-crate.

**Performance Goals**: Interactivo (turnos de refinamiento con tier local ~instantáneos). Sesión de authoring **p50 < 10 min para repo ≤ 2.000 archivos en modo local** — objetivo indicativo, **no-gate de v1** (SC-003).

**Constraints**: **Cero-egress estructural** en modo local (SC-007) — garantizado en el composition root, no por flag runtime; **provider-agnóstico**; **offline-capable** con el tier local; nunca inventar contenido de referencias no resueltas (SC-006); marcar lo tentativo vs. grounded (SC-002).

**Scale/Scope**: Single-user desktop. Repos hasta mediano/monorepo con **muestreo acotado y declarado** (D4). Sin índice de embeddings en v1.

**Verificación de procedencia** (FR-006a/b/c, añadido tras el clarify del 2026-08-23): `grounded`
exige una **cita textual comprobable** contra el material leído, no una fuente declarada. Nace de
un fallo real en la primera pasada con modelo real, donde el sistema atribuyó a un `Makefile` de
dos líneas algo que no decía. Ver `spec.md` § Clarifications.

**La salida propia no fundamenta** (FR-006d, clarify del 2026-08-24): un artefacto que escribió
el sistema se lee —US3 lo necesita para proponer una actualización— pero no respalda un
`grounded`; se reconoce por ruta canónica. Cierra el hueco por el lado contrario a FR-006a:
aquel impide atribuir a una fuente algo que no dice, este impide tratar como fuente algo que no
lo es. Nace de la segunda pasada con modelo real, donde una sesión se fundamentó en su propio
`context/CONTEXT.md` previo.

**La degradación entre tiers se declara** (FR-018, T046): `ProviderRegistry::pick` devuelve un
`Routed { provider, degraded_from }` en vez del proveedor a secas. Devolver el proveedor hacía
que ignorar la degradación fuera el camino de menor esfuerzo —y así estuvo, con un comentario
afirmando lo contrario—; el tipo obliga ahora a decidir qué se hace con el dato. De ahí salen el
evento de auditoría, el campo del snapshot y la frase de catálogo.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Ratificada contra `.specify/memory/constitution.md` **v1.0.0**:

| Principio de la constitución | Cumplimiento en este plan |
|---|---|
| **I. Arquitectura Hexagonal (Regla de Dependencia, ports propiedad del core, nombrado semántico)** | ✅ Core Domain/Application/Infrastructure; deps hacia adentro; ports en la capa que los nombra; firmas con tipos de dominio; composition root explícito; fitness function de dependencias planificada en CI. |
| **II. Testing test-first + Test Desiderata** | ✅ Contract tests por port (fakes + real), unit puro del dominio, integración del loop; property test apply∘revert. Disciplina TDD/BDD a aplicar en `tasks`. |
| **III. Conventional Commits → SemVer (cero atribución IA)** | ✅ Se acata en el flujo de trabajo del repo. |
| **Proyecto — Cero-egress estructural** | ✅ Composition root cablea solo proveedores locales en modo local + test de red (SC-007). |
| **Proyecto — Greenfield** | ✅ El plumbing Go queda como referencia; no se arrastra. |

**Resultado del gate**: PASS (sin violaciones que justificar; ver Complexity Tracking).

## Project Structure

### Documentation (this feature)

```text
specs/001-context-authoring/
├── plan.md              # Este archivo
├── research.md          # Fase 0 — decisiones técnicas + alternativas
├── data-model.md        # Fase 1 — entidades, estados, invariantes
├── quickstart.md        # Fase 1 — guía de validación end-to-end
├── contracts/           # Fase 1 — ports, comandos Tauri, tool-schemas del agente
│   ├── ports.md
│   ├── tauri-commands.md
│   └── agent-tools.md
└── tasks.md             # Fase 2 — /speckit-tasks (NO lo crea /speckit-plan)
```

### Source Code (repository root)

Workspace Cargo multi-crate. El **core** es una biblioteca hexagonal; las **pieles** la linkean.

```text
Cargo.toml                        # workspace
crates/
├── codify-core/                  # Biblioteca hexagonal reutilizable (D1+D2)
│   └── src/
│       ├── domain/               # Puro. Sin I/O, sin red, sin proveedor.
│       │   ├── context.rs        #   ContextArtifact, segmentos grounded/tentativo
│       │   ├── reference.rs      #   Reference + estado (resolved/inaccessible/out-of-scope)
│       │   ├── change.rs         #   ChangeProposal (diff), ApprovalDecision, RiskLevel
│       │   ├── session.rs        #   AuthoringSession (máquina de estados)
│       │   ├── audit.rs          #   AuditEvent (append-only)
│       │   └── ports/            #   Ports que el dominio NOMBRA
│       ├── application/          # Orquestación del loop (determinista salvo puertos)
│       │   ├── authoring_loop.rs #   El loop agéntico propio y mínimo (D5)
│       │   ├── ingest.rs         #   Estrategia de ingesta dirigida por el agente (D4)
│       │   ├── refine.rs         #   Loop curado (auto-aplica bajo riesgo / aprueba alto)
│       │   └── ports/            #   Capability ports (el dominio NO los nombra)
│       └── infrastructure/       # Adapters (driven). Único lugar con I/O.
│           ├── providers/        #   ModelProvider: ollama, llamacpp, openai_compat, oauth
│           ├── repo/             #   RepoNavigator (list/read), ReferenceResolver (fs+http)
│           ├── diff/             #   DiffEngine (similar), Applier/Reverter
│           ├── secrets/          #   keyring
│           └── composition.rs    #   Composition root (cablea modo local vs remoto)
└── codify-app/                   # Piel Tauri (fase 1) — linkea codify-core in-process
    ├── src/                      #   Driving adapter: comandos Tauri + prompter
    └── ui/                       #   Frontend web (render de diffs, aprobación, chat)

# Pieles futuras (NO en v1, documentadas para preservar la topología C):
# crates/codify-skin-mcp/   — fachada MCP (embeber en Claude Code/Cursor)
# crates/codify-skin-cli/   — CLI headless

tests/
├── contract/                     # Contract tests por port (real vs fake)
├── integration/                  # Loop end-to-end con fakes
└── fixtures/                     # Repos de prueba (incl. uno con README→SPEC referenciado)
```

**Structure Decision**: Workspace con `codify-core` (biblioteca hexagonal) + `codify-app` (Tauri). El core NO conoce Tauri ni ningún protocolo; las pieles son driving adapters. Esto materializa D1+D2 (in-process para Tauri; mismo crate reutilizable por pieles futuras) y el ethos hexagonal (composition root en `infrastructure/composition.rs`, ports en la capa que los nombra).

## Complexity Tracking

*Sin violaciones de gate que justificar.* Notas de complejidad aceptadas conscientemente:

| Elemento | Por qué se acepta | Alternativa más simple rechazada porque |
|---|---|---|
| Workspace multi-crate (core + app) | Habilita la topología C (pieles futuras linkean el core) | Un solo crate app-monolítico ataría el dominio a Tauri y bloquearía las pieles MCP/CLI |
| OAuth device-flow + keyring | D3 exige "conectar cuenta", no solo API key | Solo API keys sería más simple pero incumple el requisito del operador |
| Cero-egress estructural (registry por modo) | Garantía verificable (SC-007), no confianza en un flag | Un flag runtime `--local` es más simple pero puede filtrar por bug |
