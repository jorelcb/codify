# Feature Specification: La aplicación ve lo que ya tienes instalado

**Feature Branch**: `feat/detectar-modelos-locales`
**Created**: 2026-10-02
**Status**: Draft
**Input**: issue [#62](https://github.com/jorelcb/codify/issues/62)

## Por qué existe este spec

> «Los proveedores ya existentes vía CLI (claude-code, agy, etc.) no se sigue teniendo la más
> remota idea de cómo agregarlos. Incluso, herramientas modernas autodetectan los CLI de IA o LLMs
> ya existentes.»

Y no es una queja abstracta. En la máquina donde se escribió este spec hay **cuatro cosas usables**
al alcance de la mano:

| Lo que hay | Dónde |
|---|---|
| `llama-server` | compilado desde fuente, fuera de cualquier ruta estándar |
| `llama-cli` | al lado |
| `claude` | un agente de línea de órdenes |
| `lms` | LM Studio |

Y lo que la aplicación enseña es: **«No hay backend disponible · `http://localhost:11434`»** —
una dirección por defecto que apunta a Ollama, **que en esa máquina no está instalado**.

> **Esto es lo que cambia el orden del trabajo.** El producto se presenta como la capa que custodia
> tu contexto sin que salga de tu equipo, y ahora mismo **no encuentra el equipo**. Pide una
> dirección que el usuario no sabe, de un servicio que puede no existir, mientras ignora lo que ya
> está instalado.

Las dos rondas de validación con personas fallaron intentando conectar algo **remoto**. Nadie llegó
a probar el camino local, que es el que el producto dice servir primero.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Abrir la aplicación y que ya sepa qué tengo (Priority: P1)

Quien abre la aplicación con un modelo local ya instalado lo ve ofrecido, sin configurar nada.

**Why this priority**: es el caso que el producto dice servir primero y el único que hoy no sirve
en absoluto. Sin esto, trabajar en local exige adivinar una variable de entorno.

**Independent Test**: en una máquina con algo instalado, abrir la aplicación y mirar.

**Acceptance Scenarios**:

1. **Given** un servidor de modelos **corriendo** en la máquina, **When** el usuario abre la
   aplicación, **Then** aparece ofrecido, sin que haya escrito nada.
2. **Given** un servidor instalado pero **no corriendo**, **When** el usuario abre la aplicación,
   **Then** se le dice que está instalado y **qué le falta** para usarlo.
3. **Given** nada instalado, **When** el usuario abre la aplicación, **Then** se le dice qué podría
   instalar, en vez de una dirección que no significa nada.
4. **Given** varios modelos detectados, **When** el usuario mira, **Then** **elige**, y no se le
   impone uno.

---

### User Story 2 - Saber de dónde salió cada cosa (Priority: P2)

Quien ve un modelo ofrecido puede saber **por qué** aparece ahí.

**Why this priority**: un listado que aparece solo, sin explicación, es magia — y la magia en un
producto cuya tesis es «lo que no se puede verificar no se afirma» es una contradicción.

**Acceptance Scenarios**:

1. **Given** un modelo detectado, **When** el usuario lo mira, **Then** ve de dónde salió.
2. **Given** un modelo que el usuario configuró a mano, **When** lo mira junto a los detectados,
   **Then** distingue cuál es cuál.

---

### User Story 3 - Que detectar no cueste la apertura (Priority: P3)

Buscar lo instalado no retrasa el momento en que la aplicación es utilizable.

**Acceptance Scenarios**:

1. **Given** la aplicación arrancando, **When** la detección aún no ha terminado, **Then** la
   interfaz ya responde y lo dice.
2. **Given** una detección que no termina, **When** pasa su límite, **Then** se abandona y se dice,
   sin dejar la aplicación esperando.

---

### Edge Cases

- **Instalado fuera de las rutas habituales**: `llama-server` compilado desde fuente vive en el
  `PATH` del usuario y en ningún sitio estándar. Si solo se miran rutas conocidas, no se encuentra.
- **Un binario que está y no arranca**: existe en el `PATH` y falla al invocarlo.
- **Un servidor que responde y no sirve modelos**: contesta en el puerto y su lista está vacía.
- **Algo que no es lo que parece**: un ejecutable llamado como uno conocido que hace otra cosa.
- **Nada instalado**, que es el caso de quien abre esto por primera vez.

## Requirements *(mandatory)*

### Functional Requirements

**Encontrar**

- **FR-001**: La aplicación MUST buscar, sin que el usuario lo pida, qué modelos locales hay
  disponibles en la máquina.
- **FR-002**: La búsqueda MUST cubrir **tanto servidores en marcha como programas instalados**, y
  MUST distinguir unos de otros: están a distinta distancia de poder usarse.
- **FR-003**: La búsqueda MUST encontrar lo instalado **fuera de las rutas habituales** si está al
  alcance del usuario. Compilar desde fuente es normal en este terreno.
- **FR-004**: Lo detectado MUST ofrecerse para **elegir**, no imponerse. Varias opciones son el
  caso esperado, no la excepción.

**Decir de dónde salió**

- **FR-005**: Cada modelo ofrecido MUST declarar **de dónde se obtuvo**. Un listado sin procedencia
  contradice el principio del proyecto: lo que no se puede verificar, no se afirma.
- **FR-006**: Lo detectado MUST distinguirse de lo que el usuario configuró a mano.

**No estorbar**

- **FR-007**: La detección MUST NOT bloquear el arranque. La aplicación MUST ser utilizable
  mientras ocurre, y MUST decir que está ocurriendo.
- **FR-008**: La detección MUST tener un límite de tiempo declarado, y al pasarlo MUST abandonarse
  diciéndolo.
- **FR-009**: Lo que no se puede usar MUST decirse con **qué falta** para poder usarlo, no
  simplemente omitirse.

**Lo que no puede romperse**

- **FR-010**: Un modelo detectado MUST seguir sometido a la garantía de que nada sale del equipo.
  Detectar **amplía lo que se encuentra, nunca lo que se permite**.
- **FR-011**: La detección MUST NOT ejecutar nada que modifique la máquina. Mirar qué hay no es
  instalar, arrancar ni configurar nada.
- **FR-012**: Toda cadena visible MUST salir del catálogo, en los dos idiomas.

### Key Entities

- **Modelo disponible**: algo con lo que se podría trabajar. Tiene un nombre, una **procedencia** y
  una **distancia**: usable ahora, instalado pero no en marcha, o ausente.
- **Fuente de detección**: un sitio donde mirar. Cada una sabe buscar una clase de cosa y declara
  lo que encuentra.
- **Procedencia**: cómo llegó un modelo a la lista. Detectado por una fuente concreta, o puesto por
  el usuario.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En una máquina con un modelo local instalado, el usuario lo ve ofrecido **sin
  escribir ni configurar nada**.
- **SC-002**: **Cero** menciones de una dirección por defecto cuando no hay nada escuchando en
  ella. Hoy se enseña una que apunta a un servicio que puede no estar instalado.
- **SC-003**: Todo lo ofrecido declara su procedencia: **cero** entradas que aparecen sin explicar
  de dónde salieron.
- **SC-004**: La aplicación responde desde el arranque **aunque la detección no haya terminado**.
- **SC-005**: Detectar **no cambia** lo que puede salir del equipo, verificable de forma
  automática.
- **SC-006**: Una persona con algo instalado llega a generar contexto **sin pasar por ningún
  formulario de conexión**.

## Assumptions

- **Qué se hace con lo detectado** [NEEDS CLARIFICATION: hoy la aplicación habla con los modelos
  locales por HTTP, y rechaza por construcción cualquier dirección que no sea del propio equipo —
  así es como la garantía de cero-egress no puede mentir. Un agente de línea de órdenes como
  `claude` **no tiene dirección**: es un programa que se ejecuta. ¿Entra en este ciclo hablar con
  él, o este ciclo **encuentra** y solo se usa lo que ya se sabe usar?]
- La detección es **de solo lectura**. Mirar el `PATH` y preguntar a un servidor que ya está
  corriendo no cambia nada en la máquina; instalar o arrancar algo sí, y eso lo decide el usuario.
- No se rediseña cómo se añade un proveedor **remoto**. Eso es #61, y este spec no depende de él ni
  al revés.
- La lista de qué buscar es **conocida y declarada**: se mira lo que se sabe mirar. Encontrar algo
  nuevo exige enseñárselo, y eso se dice en vez de aparentar que se descubre solo.

### Dependencies

- `003` — el proveedor local y su garantía de que la dirección es del propio equipo. Este spec se
  apoya en ella; la pregunta abierta es si sigue bastando.
- **#63** depende de este: «qué cuenta como local» cambia si un proveedor puede ser un programa en
  vez de una dirección, y de eso depende que el modo pueda derivarse.

### Out of scope

- Instalar o arrancar nada por cuenta propia. Se dice qué falta; lo hace el usuario.
- Rediseñar el alta de proveedores remotos (#61).
- Que el modo deje de ser una entrada (#63) — depende de este, pero se decide allí.
- Descubrir proveedores que nadie ha enseñado a buscar.
