# Feature Specification: La superficie de conexión y modo

**Feature Branch**: `fix/superficie-de-conexion`
**Created**: 2026-08-27
**Status**: Draft
**Input**: issue [#52](https://github.com/jorelcb/codify/issues/52)

## Por qué existe este spec

Nace de **mirar la aplicación**, no de leer un test. Al validar a mano los escenarios de
quickstart, la superficie que `003` añadió —modo, conexiones y el formulario para conectar una
cuenta— resultó estar apelotonada en **una sola fila horizontal**: casilla de modo, lista de
cuentas y tres campos de 147 px seguidos, sin separación ni jerarquía.

Los tests decían la verdad: las claves están cableadas, el texto existe en los dos idiomas,
ninguna cadena escapa al catálogo. **Todo eso pasa. Y aun así no se entiende al mirarlo.**

> **Lo que este spec añade y no existía.** `002` exige estructura semántica (FR-027) y que la
> ventana mínima siga siendo usable (SC-007), pero **ningún requisito dice cómo debe presentarse
> la superficie de conexión**. Sin uno, «está mal maquetado» es una opinión y el arreglo no es
> verificable. Escribirlo es la mitad del trabajo.

## Clarifications

### Session 2026-08-27

- Q: ¿Dónde vive el formulario para conectar, sabiendo que el caso por defecto es no tener ninguna cuenta? → A: **Aparece al pulsar** «Conectar un proveedor», en la misma ventana. Enseñar tres campos vacíos de forma permanente hace que el caso común pague por el raro. Se descartó el diálogo modal: `002` los evita por principio (SC-004).
- Q: El modo se muestra en dos sitios —insignia en la barra y casilla en el panel—. ¿Qué hacemos? → A: **Los dos, con una sola fuente de verdad.** Informar de un vistazo y poder cambiarlo son preguntas distintas; lo que no puede ocurrir es que discrepen, y un test debe impedirlo. Es el patrón que este proyecto ya aplicó tras el defecto de `#provider-status`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conectar sin adivinar qué va en cada campo (Priority: P1)

Alguien que nunca ha usado la aplicación quiere conectar un proveedor. Ve un formulario donde
cada campo dice qué espera, y entiende qué va a pasar al enviarlo.

**Why this priority**: es la superficie que `003` entregó y que hoy no cumple su función. Sin
esto, la funcionalidad existe y no se puede usar.

**Independent Test**: poner a alguien que no conoce la aplicación delante y pedirle que conecte
un proveedor, sin explicarle nada. Si pregunta qué va en un campo, el escenario falla.

**Acceptance Scenarios**:

1. **Given** la aplicación abierta sin cuentas conectadas, **When** el usuario mira la
   superficie de conexión, **Then** ve el modo y la ausencia de cuentas, **sin campos de
   formulario ocupando sitio**.
2. **Given** esa misma pantalla, **When** el usuario pulsa «Conectar un proveedor», **Then**
   aparece el formulario, y cada campo lleva su etiqueta asociada de forma inequívoca.
3. **Given** el formulario a la vista, **When** el usuario lo recorre, **Then** el orden de los
   campos corresponde al orden en que hay que rellenarlos.
4. **Given** una cuenta conectada, **When** el usuario mira la lista, **Then** distingue la
   lista de cuentas del formulario para añadir una nueva.

---

### User Story 2 - El modo se ve como lo que es: una decisión con consecuencias (Priority: P2)

El control que decide si el contenido del repositorio puede salir del equipo **no se parece** a
un ajuste más de configuración.

**Why this priority**: es la única decisión de esta superficie con consecuencias de privacidad.
Hoy comparte fila y aspecto con el indicador de estado del backend, que es información pasiva.

**Independent Test**: pedir a alguien que señale, sin explicaciones, qué control de la pantalla
decide si algo sale de su equipo.

**Acceptance Scenarios**:

1. **Given** la aplicación abierta, **When** el usuario mira la pantalla, **Then** el control de
   modo se distingue de la configuración del proveedor sin necesidad de leer las etiquetas
   completas.
2. **Given** el modo con remotos permitido, **When** hay cuentas conectadas, **Then** se ve
   **junto al control de modo** qué proveedores podrían recibir contenido (`003`-FR-009).

---

### User Story 3 - Saber dónde estás navegando sin ver (Priority: P3)

Quien recorre la interfaz por regiones con un lector de pantalla puede decir en cuál está.

**Why this priority**: es un incumplimiento de `002`-FR-027 que ya existe en la aplicación
publicada; no bloquea a las otras dos, pero es el único de los tres que rompe un requisito
vigente.

**Independent Test**: enumerar las regiones de la interfaz y comprobar que ninguna comparte
nombre con otra.

**Acceptance Scenarios**:

1. **Given** la interfaz cargada, **When** se enumeran sus regiones, **Then** cada una tiene un
   nombre accesible **distinto de todas las demás**.
2. **Given** un lector de pantalla recorriendo regiones, **When** llega a la superficie de
   conexión, **Then** oye un nombre que la identifica, no el de la sección vecina.

---

### Edge Cases

- **Sin cuentas conectadas** (el caso por defecto): la superficie no debe ocupar espacio ni
  atención proporcionales a una funcionalidad que el usuario no está usando.
- **Muchas cuentas conectadas**: la lista crece sin empujar el formulario fuera de la vista ni
  desbordar horizontalmente.
- **Ventana en su tamaño mínimo**: el formulario sigue siendo utilizable, en los términos que ya
  fija `002`-SC-007.
- **Un campo con contenido largo** (una dirección de proveedor extensa): no rompe la fila ni
  empuja a los demás fuera.

## Requirements *(mandatory)*

### Functional Requirements

**Presentación**

- **FR-001**: La superficie de conexión y modo MUST tener **presentación propia**, distinta de la
  del indicador de estado del proveedor. MUST NOT reutilizar la disposición de una barra de
  estado para contener un formulario.
- **FR-002**: Cada campo del formulario de conexión MUST llevar una **etiqueta asociada y
  legible**, colocada de modo que no pueda confundirse con la del campo contiguo. «Legible» no se
  cuantifica a propósito: lo comprobable de forma automática es que la etiqueta esté **asociada** y
  no se confunda con la contigua; que además se entienda lo mide SC-001, con una persona delante.
  Ponerle un número aquí daría una falsa sensación de cobertura.
- **FR-002a**: El formulario MUST NOT ocupar espacio cuando no se está usando: se revela al pedir
  conectar un proveedor. El caso por defecto del producto —sin cuentas— es el que debe quedar más
  limpio.
- **FR-003**: El control de **modo** MUST distinguirse visualmente de la configuración del
  proveedor, de forma que se identifique como una decisión y no como un dato. MUST permanecer
  visible aunque el formulario esté plegado: comprobar el modo no es una tarea infrecuente.
- **FR-003a**: El modo se muestra en **dos sitios** —la insignia de la barra y el control del
  panel— y ambos MUST leer del **mismo estado**. Una comprobación automática MUST impedir que
  puedan discrepar; que dos elementos afirmen cosas distintas sobre el mismo hecho es el defecto
  que este proyecto ya corrigió una vez en el indicador de proveedor.
- **FR-004**: La lista de cuentas conectadas MUST distinguirse del formulario para añadir una.

> **FR-003 y FR-003a sirven a un control cuya existencia está en revisión.** La validación con
> persona del 2026-08-29 concluyó que el modo **no debería ser una decisión del usuario** sino una
> consecuencia de qué proveedores hay configurados: el deseo era conectar modelos locales, y se
> interpretó como tener un modo local. Lo que estos dos requisitos piden —que el control se
> distinga y que sus dos superficies no discrepen— sigue siendo correcto **mientras el control
> exista**. Si deja de existir, caen con él y no por incumplirse. Decisión en **#56**.

**Accesibilidad**

- **FR-005**: Cada región de la interfaz MUST tener un **nombre accesible único**. Dos regiones
  MUST NOT compartir nombre — cumple `002`-FR-027, hoy incumplido en dos pares.
- **FR-006**: La comprobación de unicidad MUST ser **automática**: un nombre repetido MUST hacer
  fallar la verificación, no depender de que alguien lo note mirando.

**Lo que no cambia**

- **FR-007**: Toda cadena visible MUST seguir saliendo del catálogo, en los dos idiomas, sin
  claves huérfanas — las garantías que `002` ya tiene y que esta reorganización MUST NOT romper.

### Key Entities

- **Superficie de conexión**: la zona que agrupa modo, cuentas conectadas y el formulario para
  añadir una. Hoy no existe como unidad: está fundida con el indicador de proveedor.
- **Región**: una zona de la interfaz con nombre accesible propio, recorrible por un lector de
  pantalla.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona que no conoce la aplicación conecta un proveedor **sin preguntar** qué
  va en cada campo.
- **SC-002**: **Cero** regiones con nombre accesible repetido, verificable de forma automática.
- **SC-003**: Una persona navegando **solo por regiones** puede decir correctamente en cuál está.
- **SC-004**: Una persona señala correctamente, sin explicaciones, qué control decide si algo
  sale de su equipo.
- **SC-005**: Con la ventana en su tamaño mínimo, el formulario de conexión sigue siendo
  utilizable: **cero** desbordamiento horizontal, y cada campo de texto muestra al menos
  **24 caracteres** de su contenido sin recortar.
- **SC-006**: La insignia de modo y el control de modo **nunca** muestran estados distintos,
  verificable de forma automática.

## Assumptions

- La superficie sigue viviendo **en la ventana principal**, no en una ventana aparte ni en un
  diálogo. Conectar un proveedor es infrecuente, pero ver el modo no lo es: sacarlo de la vista
  principal escondería justo lo que decide si algo sale del equipo.
- **24 caracteres** es el mínimo por el que una dirección de proveedor deja de ser adivinable. Es
  un número elegido, no medido; si aparece evidencia de que hace falta más, se revisa.
- El caso por defecto del producto sigue siendo **sin cuentas conectadas**, así que la superficie
  se diseña para que ese caso sea el que menos ruido produce.
- No se rediseña el resto de la interfaz. Este spec toca la superficie de conexión y los nombres
  de región; lo demás se queda como está.

### Dependencies

- `002`-FR-027 (estructura semántica) y `002`-SC-007 (ventana mínima) — este spec los sostiene y
  concreta, no los reemplaza.
- `003`-FR-009 (ver qué proveedores podrían recibir contenido) — su presentación es parte de esta
  superficie.

### Out of scope

- Rediseñar la corriente de actividad, el panel de decisión o la vista de artefacto.
- Cambiar qué campos pide el formulario de conexión: se presentan mejor los que ya hay.
- Cualquier cambio en el comportamiento de conexión, custodia o reparto — eso es `003`, **con
  una excepción declarada abajo**.

**La excepción.** `start_session` deja de leer `request.local` y pasa a leer el modo guardado. Eso
es comportamiento de reparto, y entra igual porque **FR-003a es irrealizable sin ello**: «ambas
superficies leen el mismo estado» sería falso mientras exista un tercer lector con su propia
copia. Repara de paso `003`-FR-008a, hoy incumplido de punta a punta — el modo se guardaba y no se
aplicaba. Está razonado en [research.md D2](./research.md) y registrado en la tabla de Complexity
Tracking del plan; **el defecto lleva issue propio** para que el arreglo sea trazable fuera de este
spec.
