# Implementation Plan: La aplicación ve lo que ya tienes instalado

**Branch**: `feat/detectar-modelos-locales` | **Date**: 2026-10-02 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-detectar-modelos-locales/spec.md`

## Summary

Que la aplicación mire qué modelos hay en la máquina y los ofrezca, en vez de pedir una dirección
que el usuario no sabe. Dos clases de hallazgo —**servidores en marcha** y **programas
instalados**— a distinta distancia de poder usarse, y cada uno diciendo de dónde salió.

La decisión que ordena el diseño: **lo que clasifica a un proveedor es dónde ocurre la inferencia,
no dónde corre el programa**. Y en este ciclo solo se cablea lo que puede **sostener** esa
clasificación por construcción, no lo que la declara ([research.md](./research.md) D4).

## Technical Context

**Language/Version**: Rust 1.75+ (edición 2021) · JavaScript ES2022 sin transpilación ni bundler

**Primary Dependencies**: Tauri 2 (piel); `reqwest` y `tokio`, ya presentes. **Ninguna nueva** —
mirar el `PATH` es biblioteca estándar

**Storage**: nada persistente. Lo detectado se recalcula; guardarlo sería una cuarta fuente de
verdad sobre un hecho que cambia solo

**Testing**: `cargo test`. El núcleo con adapters de mentira; la piel por análisis estático en
`crates/codify-app/tests/ui_contract.rs` (28 tests hoy)

**Target Platform**: escritorio (macOS/Linux/Windows). La detección **es dependiente del sistema**:
dónde se instalan las cosas no es igual en los tres

**Project Type**: desktop-app — núcleo hexagonal + piel

**Performance Goals**: la detección entera cabe en **3 s** y no bloquea el arranque (SC-004)

**Constraints**: **no se ejecuta ningún programa encontrado** (D3) · nada clasificado como local
puede inferir fuera del equipo (SC-005) · cero cadenas visibles fuera del catálogo

**Scale/Scope**: un puñado de fuentes de detección —hoy dos—, un port nuevo, el endpoint por
defecto retirado, y lo que la piel enseña al arrancar

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Veredicto | Por qué |
|---|---|---|
| I · Regla de Dependencia | **PASA** | El port de detección vive en `application/` —el dominio no nombra «detectar», es una capacidad—; los adapters en `infrastructure/`. `arch_deps.rs` lo verifica. |
| I · Ports en la capa que los nombra | **PASA** | Es un driven port de capacidad, no de dominio. La constitución separa justo esos dos casos. |
| I · Firmas con tipos de dominio | **PASA** | Lo que cruza el port son tipos propios, no `String` ni `PathBuf` sueltos. |
| I · Nombres sin decoración | **PASA** | `ModeloDisponible`, `FuenteDeDeteccion`, `Procedencia`. |
| II · Test-first | **PASA con obligación** | Cada requisito estrena test en rojo, y cada test se verifica **inyectando la violación**. En este ciclo hay dos que importan más: que no se ejecute nada encontrado, y que nada remoto se cuele como local. |
| II · Test Desiderata | **PASA, con una renuncia declarada** | La detección depende del sistema, así que los tests usan fuentes de mentira: rápidos y deterministas, a cambio de no probar el descubrimiento **real**. Eso lo cierra una persona con su máquina (SC-001), y se dice en vez de aparentarlo. |
| III · Conventional Commits, cero atribución de IA | **PASA** | |
| Greenfield | **PASA** | El endpoint por defecto se retira, no se deja por compatibilidad. |
| **Cero-egress estructural en modo local** | **PASA, y se lee en su alcance** | Ver abajo. |

### Sobre el alcance del cero-egress

La constitución dice **«cero-egress estructural en modo local»**, y eso es exactamente lo que
significa: rige **el camino local**, no todos los proveedores. El producto quiere **los dos
caminos** — modelos cuya inferencia no sale del equipo, y modelos de inferencia remota.

Hace falta decirlo aquí porque durante este mismo ciclo el agente lo leyó de más, llegó a tratar un
CLI de inferencia remota como un agujero de seguridad, y evaluó opciones de diseño por si «tocaban
el principio». Eso estrecha el producto a la mitad.

**Lo que este plan sí sostiene**: que nada clasificado como local infiera fuera del equipo. El
defecto no es tener `claude` instalado — es presentarlo como local.

## Project Structure

### Documentation (this feature)

```text
specs/006-detectar-modelos-locales/
├── plan.md              # Este archivo
├── research.md          # Fase 0 — seis decisiones, con lo descartado
├── data-model.md        # Fase 1 — modelo disponible, fuente, procedencia, distancia
├── quickstart.md        # Fase 1 — qué cierra el build y qué exige una máquina de verdad
├── contracts/
│   └── deteccion.md     # Fase 1 — el port, el catálogo de lo detectable, y las reglas
├── checklists/
│   └── requirements.md  # De /speckit-specify, 16/16
└── tasks.md             # Fase 2 — lo crea /speckit-tasks
```

### Source Code (repository root)

```text
crates/codify-core/
└── src/
    ├── application/
    │   ├── deteccion.rs     # `ModeloDisponible`, `Procedencia`, `Distancia`, el port
    │   └── ports.rs         # `ProviderIssue` gana los motivos nuevos
    └── infrastructure/
        └── deteccion/
            ├── servidores.rs   # puertos conocidos, sobre `LocalProviderProbe`
            └── programas.rs    # el `PATH`, **sin ejecutar nada**

crates/codify-app/
├── src/
│   ├── commands.rs      # comando de detección; **`DEFAULT_ENDPOINT` desaparece**
│   └── strings.rs       # claves nuevas × 2 idiomas
├── ui/                  # lo detectado al arrancar, con su procedencia y qué falta
└── tests/
    └── ui_contract.rs   # las reglas estáticas

crates/codify-core/tests/
└── deteccion.rs         # el comportamiento, con fuentes de mentira
```

**Structure Decision**: un módulo nuevo por capa —`application/deteccion.rs` e
`infrastructure/deteccion/`— porque son varias fuentes con la misma forma y meterlas en
`providers/` las confundiría con los proveedores mismos. Detectar y proveer son cosas distintas:
una encuentra candidatos, la otra habla con ellos.

## Complexity Tracking

> Sin violaciones de la constitución que justificar.

Dos renuncias que conviene registrar en vez de descubrirlas al revisar:

| Qué no se hace | Por qué se acepta | Qué se descartó |
|---|---|---|
| **No se ejecuta ningún programa encontrado**, ni para saber su versión | Lanzar un binario del `PATH` para decidir si enseñarlo es más riesgo del que compra: lo único que hace falta es «está instalado», y lo confirma el usuario al usarlo | Invocarlo con `--version` y un límite de tiempo. Se descarta por el mismo motivo por el que no se lee el llavero del operador: no hace falta para la comprobación que toca |
| Los tests no prueban el **descubrimiento real**, solo la lógica sobre fuentes de mentira | La detección depende del sistema; un test que mire el `PATH` de verdad daría resultados distintos en cada máquina y en CI | Un test de integración con un `PATH` sintético. Posible y útil: queda para cuando haya más de dos fuentes y el valor lo justifique |
