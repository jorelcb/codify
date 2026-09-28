# Implementation Plan: Conectar una cuenta llega a su fin

**Branch**: `fix/conectar-termina` | **Date**: 2026-09-27 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-conectar-termina/spec.md`

## Summary

Cerrar el camino de conectar una cuenta: **dónde escribir la credencial**, el desenlace visible, y
que lo empezado no se acumule. El backend está entero desde `003`; lo que falta es la mitad de la
interfaz y tres defectos que solo se ven al recorrer el camino de punta a punta
([research.md](./research.md)).

Y una regla nueva que sale del propio defecto: **un camino que nadie puede alcanzar queda declarado
o no existe** (FR-012). Este spec nace de algo construido, con sus tests en verde, que nadie pudo
usar durante dos ciclos.

## Technical Context

**Language/Version**: Rust 1.75+ (edición 2021) · JavaScript ES2022 sin transpilación ni bundler

**Primary Dependencies**: Tauri 2 (piel); `keyring` (almacén del sistema, ya presente). **Ninguna
nueva**

**Storage**: el almacén de credenciales del sistema (`SystemKeyring`), **inyectable**: construirlo
dentro del camino hacía que cualquier test del camino escribiera en el llavero real de quien lo
corriera, y los primeros de este ciclo lo hicieron. El desafío en curso vive en memoria del proceso

**Testing**: `cargo test`. Contrato de interfaz por análisis estático en
`crates/codify-app/tests/ui_contract.rs` (22 tests hoy); comportamiento del núcleo en
`crates/codify-core/tests/`

**Target Platform**: aplicación de escritorio (macOS/Linux/Windows vía Tauri)

**Project Type**: desktop-app — núcleo hexagonal + piel

**Performance Goals**: N/A — el coste lo pone la red del proveedor, no el código

**Constraints**: la credencial no aparece en registros, eventos ni mensajes; cero cadenas visibles
fuera del catálogo, **incluidas las del backend**; cero-egress estructural intacto

**Scale/Scope**: un desafío en curso como máximo. Dos módulos JS, un comando nuevo, dos comandos
retocados, un enum de dominio, y catorce claves de catálogo

**Entregado: doce tests nuevos**, no los siete que este plan previó. Cinco salieron de inyectar
violaciones y ver que no caía nada: el caso literal de FR-006 —que **el proveedor** rechace, no que
el campo esté vacío— no era alcanzable sin un conector inyectable; el almacén ausente no tenía
prueba; y **ningún test impedía escribir en el llavero del operador**, cosa que los primeros de este
ciclo hicieron.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Veredicto | Por qué |
|---|---|---|
| I · Regla de Dependencia | **PASA** | El enum nuevo vive en `application/ports.rs`, junto al `Desafio` que lo lleva. Nada apunta hacia afuera; `arch_deps.rs` no tiene nada que decir. |
| I · Ports y firmas con tipos de dominio | **PASA, y mejora** | `DirectCredential::new(String)` recibe hoy una **frase en español**. Pasa a recibir un tipo de dominio con valores cerrados. Es el principio aplicado a un sitio donde estaba incumplido. |
| I · Nombres sin decoración | **PASA** | `InstruccionDeCredencial`, `DesafioEnCurso`: nombres del problema. |
| II · Test-first | **PASA con obligación** | Cada requisito estrena test **en rojo**, y cada test se verifica **inyectando la violación**. Y esta vez con un cuidado extra: tres de los defectos de este spec sobrevivieron a tests verdes, así que la inyección va **en la capa que cablea**, no donde vive la lógica. |
| II · Test Desiderata | **PASA, con una renuncia declarada** | Los tests siguen siendo rápidos y deterministas. FR-012 se comprueba por análisis estático, que es *specific* pero **no exhaustivo**: caza la clase de camino sin puerta que produjo este defecto, no todas. Se dice en D2 en vez de aparentar lo contrario. |
| III · Conventional Commits, cero atribución de IA | **PASA** | |
| Greenfield | **PASA** | La firma de `DirectCredential` cambia sin capa de compatibilidad; el `HashMap` de desafíos se sustituye, no se complementa. |
| **Cero-egress estructural** | **PASA** | Nada de esto toca el composition root. Una cuenta conectada sigue sin poder entrar en un grafo local: el método no existe. Lo que cambia es que ahora se puede llegar a conectarla. |

### Una nota sobre la credencial

Este ciclo es el primero en que **una credencial real cruza la interfaz**. Tres consecuencias que
el plan asume y los tests vigilan: no se registra en ningún evento ni traza, no queda en el DOM tras
enviarse, y **no se comprueba leyendo el llavero del operador** — la custodia se verifica desde el
proceso de la propia aplicación, que es quien tiene derecho a su entrada. Esa última regla nació de
un incidente real (`003`-T031).

## Project Structure

### Documentation (this feature)

```text
specs/005-conectar-termina/
├── plan.md              # Este archivo
├── research.md          # Fase 0 — seis decisiones, y tres defectos que aparecieron al mirar
├── data-model.md        # Fase 1 — el desafío y su ciclo de vida
├── quickstart.md        # Fase 1 — qué cierra el build y qué exige una persona
├── contracts/
│   └── conexion.md      # Fase 1 — comandos, el port, y las reglas con su inyección
├── checklists/
│   └── requirements.md  # De /speckit-specify, 16/16
└── tasks.md             # Fase 2 — lo crea /speckit-tasks
```

### Source Code (repository root)

```text
crates/codify-core/
└── src/
    ├── application/ports.rs              # `InstruccionDeCredencial` con `code()` y `all()`;
    │                                     #   `Desafio` lleva la clave, no la frase
    └── infrastructure/secrets/direct.rs  # deja de recibir una frase en español

crates/codify-app/
├── src/
│   ├── commands.rs      # `abandon_connection` nuevo; `connect_provider` reemplaza en vez de
│   │                    #   acumular; `complete_connection` **no consume el desafío si falla**
│   └── strings.rs       # 6 claves nuevas × 2 idiomas
├── ui/
│   ├── index.html       # dónde escribir la credencial, y el control de cancelar
│   ├── connections.js   # el camino completo: enviar, cancelar, desenlace
│   └── styles.css       # el estado «en curso»
└── tests/
    ├── ui_contract.rs   # 6 tests nuevos (22 → 28)
    └── conexion_termina.rs  # 6 tests nuevos del camino
```

**Structure Decision**: sin estructura nueva. Un archivo de test nuevo en `codify-app` porque los
dos tests que faltan son **de comportamiento del camino**, no de contrato estático, y meterlos en
`ui_contract.rs` mezclaría dos cosas que se leen distinto.

## Complexity Tracking

> Sin violaciones de la constitución que justificar.

Sí hay **una renuncia** que conviene registrar en vez de descubrirla al revisar:

| Qué no cubre | Por qué se acepta | Qué se descartó |
|---|---|---|
| FR-012 caza **la clase** de camino sin puerta que produjo este defecto —una rama elegida por un parámetro que la interfaz nunca varía— **no todas** las formas de código inalcanzable | Es barato, determinista y ataca el caso real. Pretender cobertura general sería la misma clase de afirmación sin respaldo que este spec viene a corregir | Cobertura de ramas (`cargo-llvm-cov`): mide lo que **los tests** ejercitan, no lo que **la interfaz** alcanza, que es la pregunta de FR-012. Un test puede llamar a la rama delegada y dejarla igual de inalcanzable para el usuario |
