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

## Clarifications

### Session 2026-10-02

- Q: ¿Entra en este ciclo hablar con un agente de línea de órdenes, y qué cuenta como «local»? → A: **Detectar todo; cablear en este ciclo solo lo que puede probarse local.** Lo que decide la clasificación no es dónde corre el programa sino **dónde ocurre la inferencia**: `llama-server` y `lms` corren local y infieren local; `claude` corre local e infiere en remoto. Son caminos distintos, no uno bueno y uno malo.
- **Corrección del operador, no una pregunta del agente**: «nada sale del equipo» no es una condición que todo proveedor deba cumplir. **Suponerlo era un error de lectura.** El objetivo del producto siempre fue **ofrecer los dos caminos**: integración con modelos que cumplen esa condición, y también con los de inferencia remota. La constitución ya lo dice acotado —cero-egress estructural **en modo local**—; quien lo generalizó fue el agente. Un proveedor remoto no es una violación: es la otra mitad del producto. **Lo que sí es un defecto es que algo remoto se presente como local.**
- Q: Hoy el único modo de decir dónde está tu modelo es una variable de entorno, con una dirección por defecto. ¿Cómo convive con la detección? → A: **No convive: la detección es el único modo en que un usuario integra un modelo local.** Pedir una variable de entorno para usar algo que ya está instalado es trasladarle al usuario el trabajo de descubrir, y **ninguna herramienta comparable lo hace**. La variable sobrevive únicamente como anulación para desarrollo y pruebas, declarada como tal; el producto deja de mencionarla y deja de inventarse una dirección por defecto.

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
- **Un binario que está y no arranca**: se enseña como instalado igualmente. Saber si arranca
  exigiría **ejecutarlo**, y esta feature decide no hacerlo. Lo descubre el usuario al usarlo.
- **Un servidor que responde y no sirve modelos**: contesta en el puerto y su lista está vacía.
- **Algo que no es lo que parece**: un ejecutable llamado como uno conocido que hace otra cosa
  **se enseñará como si lo fuera**. Distinguirlos exigiría ejecutarlo. Es lo que esta decisión
  acepta no saber, razonado en [research D3](./research.md).
- **Nada instalado**, que es el caso de quien abre esto por primera vez.
- **Una anulación de desarrollo apuntando a algo que no responde**: se trata como cualquier otra
  cosa que no responde, sin trato de favor por venir del entorno.

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

**Cómo se integra un modelo local**

- **FR-013**: La detección MUST ser **el único modo** en que un usuario integra un modelo local. No
  MUST exigirse configurar nada para usar algo que ya está instalado.
- **FR-014**: La aplicación MUST NOT ofrecer una dirección por defecto cuando no hay nada
  escuchando en ella. Hoy enseña una que apunta a un servicio que puede no estar instalado, y eso
  es afirmar algo sin comprobarlo — justo lo que el producto dice no hacer.
- **FR-015**: Una anulación por entorno MAY existir **para desarrollo y pruebas**, y MUST estar
  declarada como tal. MUST NOT aparecer en la interfaz ni en nada que el usuario tenga que leer
  para usar el producto.

**Lo que no puede romperse**

- **FR-010**: Cada modelo detectado MUST clasificarse por **dónde ocurre la inferencia**, no por
  dónde corre el programa. Un programa que vive en la máquina y consulta un servicio remoto **es
  remoto**, y presentarlo como local sería el defecto — no tenerlo.
- **FR-010a**: La clasificación MUST poder **sostenerse**, no declararse. Hoy un proveedor local lo
  es porque su dirección es del propio equipo y el constructor rechaza cualquier otra. Lo que se
  cablee en este ciclo MUST conservar esa propiedad; lo que no pueda sostenerla **no se cablea como
  local**.
- **FR-010b**: Detectar **no decide** por el usuario. Encontrar un modelo de inferencia remota no lo
  pone en marcha ni lo da por aceptado: se enseña, se dice lo que es, y se usa si el usuario lo
  elige. Los dos caminos existen a propósito.
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
- **SC-004**: La aplicación responde desde el arranque **aunque la detección no haya terminado**, y
  la detección no pasa de **3 segundos** antes de abandonarse diciéndolo.
- **SC-007**: **Cero** pasos de configuración para usar un modelo que ya está instalado. Lo que el
  usuario tiene que leer para empezar a trabajar en local es **nada**.
- **SC-005**: **Cero** modelos clasificados como locales cuya inferencia ocurra fuera del equipo,
  verificable de forma automática. Detectar amplía lo que se encuentra; lo que cambia qué sale es
  lo que el usuario elige, no lo que la aplicación descubre.
- **SC-006**: Una persona con algo instalado llega a generar contexto **sin pasar por ningún
  formulario de conexión**.

## Assumptions

- **Este ciclo detecta todo y cablea solo lo que puede probarse local.** Un agente de línea de
  órdenes de inferencia remota se encuentra, se enseña y se clasifica como remoto; que además se
  pueda **usar** por esa vía es trabajo de otro ciclo, porque hablar con un subproceso necesita un
  adapter que hoy no existe. Encontrarlo y nombrarlo correctamente ya es más de lo que hay.
- La detección es **de solo lectura**. Mirar el `PATH` y preguntar a un servidor que ya está
  corriendo no cambia nada en la máquina; instalar o arrancar algo sí, y eso lo decide el usuario.
- **El producto ofrece los dos caminos a propósito**: modelos cuya inferencia no sale del equipo, y
  modelos de inferencia remota. La garantía estructural de cero-egress rige **el primero**, que es
  lo que la constitución dice. Tratarla como una condición que todo proveedor debe cumplir dejaría
  fuera la mitad del producto.
- No se rediseña cómo se añade un proveedor **remoto**. Eso es #61, y este spec no depende de él ni
  al revés.
- La lista de qué buscar es **conocida y declarada**: se mira lo que se sabe mirar. Encontrar algo
  nuevo exige enseñárselo, y eso se dice en vez de aparentar que se descubre solo.
- **El límite de la detección son 3 segundos.** Es un número elegido, no medido: por encima de eso
  la espera se nota al abrir, y por debajo se corre el riesgo de descartar un servidor que tarda en
  contestar. Si aparece evidencia de que hace falta otro, se revisa.

### Dependencies

- `003` — el proveedor local y su garantía de que la dirección es del propio equipo. Este spec se
  apoya en ella; la pregunta abierta es si sigue bastando.
- **#63** depende de este: «qué cuenta como local» cambia si un proveedor puede ser un programa en
  vez de una dirección, y de eso depende que el modo pueda derivarse.

### Out of scope

- Instalar o arrancar nada por cuenta propia. Se dice qué falta; lo hace el usuario.
- Rediseñar el alta de proveedores remotos (#61).
- **Usar** un modelo de inferencia remota a través de un subproceso. Se detecta y se clasifica;
  hablarle necesita un adapter de subproceso y es un ciclo propio.
- Que el modo deje de ser una entrada (#63) — depende de este, pero se decide allí.
- Descubrir proveedores que nadie ha enseñado a buscar.
