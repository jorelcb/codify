# Implementation Plan: La superficie de conexión y modo

**Branch**: `fix/superficie-de-conexion` | **Date**: 2026-08-27 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-superficie-de-conexion/spec.md`

## Summary

Dar **presentación propia** a la superficie que agrupa modo, cuentas conectadas y el formulario
de conexión; revelar el formulario solo cuando se pide; dar **nombre accesible único** a cada
región; y hacer que las dos superficies del modo no puedan discrepar.

Lo último resultó ser más grande de lo que el spec suponía. Al buscar cómo comprobar FR-003a
apareció que el modo **no tiene dos dueños: tiene tres**, y que el tercero rompe una garantía
vigente. Está documentado en [research.md](./research.md) (D2) y decide la forma del arreglo: la
única fuente del modo pasa a ser el estado del backend, y la interfaz deja de tener uno propio.

## Technical Context

**Language/Version**: Rust 1.75+ (edición 2021) · JavaScript ES2022 sin transpilación ni bundler

**Primary Dependencies**: Tauri 2 (piel); ninguna nueva — esta feature no añade dependencias

**Storage**: N/A (el modo vive en memoria del proceso; su persistencia es de `003`)

**Testing**: `cargo test` — la verificación es **análisis estático** sobre `index.html`,
`ui/*.js`, `styles.css`, `tauri.conf.json` y `src/strings.rs`, en
`crates/codify-app/tests/ui_contract.rs` (700 líneas, 14 tests hoy)

**Target Platform**: aplicación de escritorio (macOS/Linux/Windows vía Tauri)

**Project Type**: desktop-app — núcleo hexagonal + piel

**Performance Goals**: N/A — no hay trabajo en caliente en esta superficie

**Constraints**: la ventana no baja de `minWidth` (`tauri.conf.json`); cero cadenas fuera del
catálogo; cero-egress estructural intacto

**Scale/Scope**: una sección del HTML, dos módulos JS, el CSS de esa sección, dos comandos Tauri,
cuatro claves de catálogo nuevas en dos idiomas, y cuatro tests nuevos

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Veredicto | Por qué |
|---|---|---|
| I · Regla de Dependencia | **PASA** | Todo el cambio vive en `crates/codify-app` (piel) y en sus tests. El dominio y la aplicación no se tocan. `arch_deps.rs` no tiene nada que decir aquí. |
| I · Ports y firmas | **N/A** | No se añade ni se altera ningún port. |
| I · Nombres sin decoración | **PASA** | Los nombres nuevos (`modo`, `superficie de conexión`) son del dominio del problema. |
| II · Test-first | **PASA con obligación** | Cada FR nuevo estrena su test **en rojo** antes del arreglo. Y cada test se verifica **inyectando la violación**: un test que nunca se ha visto fallar no está verificado — es la lección que este ciclo ha repetido seis veces. |
| II · Test Desiderata | **PASA** | Los tests son *inspiring*, *fast*, *deterministic* y *specific*: análisis estático, sin navegador, sin red, señalando el archivo y la línea. Sacrifican *behavioral* — comprueban la estructura que produce el comportamiento, no el comportamiento; ver D1 y D2 para el precio y por qué se paga. |
| III · Conventional Commits, cero atribución de IA | **PASA** | |
| Greenfield | **PASA** | `request.local` se elimina, no se deja por compatibilidad. |
| **Cero-egress estructural** | **PASA, con una advertencia que hay que decir en voz alta** | Ver abajo. |

### La advertencia sobre cero-egress

Hoy la aplicación **siempre** construye el grafo en modo local, porque el único sitio que decide
el modo de una sesión lee un valor de la interfaz que nadie actualiza (D2). Es decir: la
seguridad que se observa hoy es **un accidente**, no la garantía.

Este plan repara el defecto, y al hacerlo el modo híbrido pasa a ser realmente alcanzable. Eso
**no debilita** la garantía: sigue realizada donde la constitución exige, en el composition root
—`CoreBuilder<Local>` no tiene método `remote_provider`, y no compila quien lo intente— y
verificada por `compile_fail.rs` y `egress_guard.rs`. Lo que cambia es que empieza a *ejercitarse*
el camino que ya estaba escrito.

Decirlo importa porque la lectura ingenua es la contraria: «antes nada salía, ahora sí puede». La
verdad es que antes **el usuario no podía elegir**, que es un defecto distinto y no una virtud.

## Project Structure

### Documentation (this feature)

```text
specs/004-superficie-de-conexion/
├── plan.md              # Este archivo
├── research.md          # Fase 0 — seis decisiones, con lo descartado
├── data-model.md        # Fase 1 — el modo, la región, la superficie
├── quickstart.md        # Fase 1 — qué comprueba el build y qué exige una persona
├── contracts/
│   └── ui-contract.md   # Fase 1 — comandos, regiones y las reglas que un test hace cumplir
├── checklists/
│   └── requirements.md  # De /speckit-specify
└── tasks.md             # Fase 2 — lo crea /speckit-tasks
```

### Source Code (repository root)

```text
crates/codify-app/
├── src/
│   ├── commands.rs      # `set_mode` devuelve el modo; `mode()` nuevo; `start_session` lee
│   │                    #   `state.mode` en vez de creerle a la petición
│   └── strings.rs       # 4 claves nuevas × 2 idiomas
├── ui/
│   ├── index.html       # `#conexiones` deja de ser `.provider`; formulario en `<details>`;
│   │                    #   nombres de región únicos
│   ├── styles.css       # presentación propia; `min-width: 24ch` en los campos
│   ├── connections.js   # deja de tener opinión sobre el modo
│   └── main.js          # único sitio que pinta el modo, en las dos superficies
├── tests/
│   └── ui_contract.rs   # 4 tests nuevos (ver contrato)
└── tauri.conf.json      # sin cambios — `minWidth` ya está declarado
```

**Structure Decision**: no se crea estructura nueva. La feature es una reorganización dentro de
`crates/codify-app`, con su verificación en el arnés de contrato de interfaz que ya existe y que
ya sabe leer los cinco archivos que hacen falta.

## Complexity Tracking

> Sin violaciones de la constitución que justificar.

Sí hay **un desbordamiento de alcance** que conviene registrar aquí en vez de esconderlo en un
commit:

| Qué se sale del spec | Por qué se hace igual | Qué se descartó |
|---|---|---|
| `start_session` deja de leer `request.local` y pasa a leer `state.mode` | Sin esto, FR-003a es literalmente irrealizable: «ambas superficies leen el mismo estado» sería falso mientras exista un tercer lector con su propia copia. Además repara `003`-FR-008a, hoy incumplido de punta a punta. | Dejarlo fuera y anotarlo como deuda: convertiría este spec en cosmética sobre un defecto funcional, y FR-003a en una afirmación que el test no puede sostener. |

Este defecto es de `003`, no de `004`. Merece **issue propio** para que quede trazable — el spec
lo consume, no lo esconde.
