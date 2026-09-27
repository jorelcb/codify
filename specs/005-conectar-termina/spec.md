# Feature Specification: Conectar una cuenta llega a su fin

**Feature Branch**: `fix/conectar-termina`
**Created**: 2026-09-25
**Status**: Draft
**Input**: issue [#54](https://github.com/jorelcb/codify/issues/54)

## Por qué existe este spec

Alguien que no conocía la aplicación intentó conectar un proveedor. Escribió un nombre, pulsó
**Conectar**, y no ocurrió nada. Escribió también una dirección, volvió a pulsar, y tampoco.

No era impresión suya. `connect_provider` abre un desafío, la interfaz enseña una frase —«se guarda
en el almacén del sistema y no vuelve a mostrarse»— y **ahí se acaba el camino**: no hay campo
donde escribir la credencial, y `complete_connection` está registrado en el backend sin que ningún
JavaScript lo invoque. El desafío se abre y nunca se completa.

> **Lo raro es dónde está el hueco.** El backend está entero: `complete_connection` recibe la
> credencial, la entrega al conector, la guarda en el almacén del sistema y crea la conexión. El
> DTO del desafío ya lleva su identificador, su vía y sus instrucciones. **Lo que falta es la
> mitad de la interfaz**, y por eso todos los tests pasaban: un comando que nadie llama compila y
> pasa los suyos.

`004` reorganizó esta superficie para que se entendiera. Se entiende, y no funciona. Su criterio
SC-001 —conectar sin preguntar— quedó declarado en fallo porque **no se puede medir si alguien
conecta sin ayuda mientras conectar sea imposible**. Este spec es lo que lo desbloquea.

## Clarifications

### Session 2026-09-26

- Q: ¿Entra la vía delegada en este ciclo? → A: **No: solo la vía directa, y la delegada queda declarada como camino sin puerta.** No se aplaza ningún requisito, porque hoy no hay ninguno que la cubra: `003` construyó el motor de las dos vías y solo cableó una. Lo que sí se hace es **dejar de tener un camino invisible** — el mismo defecto que trajo este spec, y que el test de comandos huérfanos no caza porque la delegada no es un comando sino una rama dentro de uno, elegida por un parámetro que la interfaz nunca varía.

- Q: Las instrucciones del desafío nacen como texto en el núcleo, que no sabe en qué idioma está la interfaz. ¿Qué viaja? → A: **Una clave de catálogo, no texto.** El conector sabe *qué* hay que decirle al usuario, no *cómo* decirlo en su idioma. La frontera se queda donde este proyecto la pone siempre —el núcleo decide, la piel redacta— y la clave entra en los tests de catálogo que ya existen, sin inventar nada.
- Q: ¿Quién cierra un desafío que no se completa? → A: **Las dos cosas: control explícito para cancelar, y reemplazo al volver a pedir conectar.** Responden a casos distintos: el control a «cambié de idea», el reemplazo al que de verdad ocurre —pulsar «Conectar» dos veces—. Solo con el control, el segundo intento dejaría huérfano al primero; solo con el reemplazo, no habría forma de decir «déjalo». Una cuenta no puede tener dos desafíos compitiendo.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pegar la credencial y quedar conectado (Priority: P1)

Quien tiene una clave de su proveedor la pega, y ve que la cuenta queda conectada.

**Why this priority**: es el único camino alcanzable hoy desde la interfaz, y el que bloquea
`004`-SC-001. Sin esto, la funcionalidad entera de `003` es inalcanzable.

**Independent Test**: dar a alguien una clave y pedirle que conecte la cuenta, sin explicar nada.

**Acceptance Scenarios**:

1. **Given** el formulario relleno, **When** el usuario pide conectar, **Then** aparece **dónde
   escribir la credencial**, con instrucciones de dónde encontrarla.
2. **Given** la credencial escrita, **When** el usuario la envía, **Then** la cuenta aparece en la
   lista de conectadas, y el campo de la credencial deja de estar a la vista.
3. **Given** una credencial que el proveedor rechaza, **When** el usuario la envía, **Then** se
   dice que fue rechazada y **el desafío sigue en pie**: se puede corregir sin volver a empezar.
4. **Given** una cuenta recién conectada, **When** el usuario mira la lista, **Then** la ve con el
   nombre que le puso, no con un identificador interno.

---

### User Story 2 - Saber por qué no se pudo (Priority: P2)

Cuando la conexión no sale, el usuario sabe **qué pasó** y **qué puede hacer**.

**Why this priority**: sin esto, todo fallo se parece al defecto que originó este spec — pulsar y
que no pase nada. La diferencia entre «falló» y «no hizo nada» es lo que decide si alguien vuelve a
intentarlo.

**Independent Test**: provocar cada fallo y comprobar que la pantalla dice cuál fue y ofrece salida.

**Acceptance Scenarios**:

1. **Given** un sistema sin almacén de credenciales, **When** el usuario intenta conectar, **Then**
   se le dice que no lo hay y que **puede seguir trabajando en local** (`003`-FR-004).
2. **Given** un desafío que ya no está en curso —por tiempo o por haberlo abandonado—, **When** el
   usuario envía la credencial, **Then** se le dice, y puede empezar de nuevo.
3. **Given** cualquiera de esos fallos, **When** el usuario lo lee, **Then** el texto está en su
   idioma, como el resto de la aplicación.

---

### User Story 3 - Abandonar sin dejar rastro (Priority: P3)

Quien empieza a conectar y cambia de idea puede irse, y el sistema no se queda esperando.

**Why this priority**: no bloquea a nadie, pero un desafío abierto retiene un conector y las
credenciales a medio camino. Un camino que solo tiene entrada acaba acumulando gente dentro.

**Acceptance Scenarios**:

1. **Given** un desafío en curso, **When** el usuario usa el control de cancelar, **Then** deja de
   estar en curso y sus datos a medias no se conservan.
2. **Given** un desafío abandonado, **When** el usuario vuelve a pedir conectar, **Then** empieza
   uno nuevo, sin arrastrar nada del anterior.
3. **Given** un desafío en curso, **When** el usuario pide conectar otra vez esa misma cuenta,
   **Then** el anterior queda reemplazado y **solo uno** sigue en curso.

---

### Edge Cases

- **Credencial vacía**: enviar sin escribir nada no debe consumir el desafío.
- **Dos intentos seguidos**: el segundo reemplaza al primero (FR-009). Es el caso que de verdad
  ocurre, y hoy deja dos dentro del sistema sin que nada lo diga.
- **El almacén desaparece a mitad**: el sistema tenía almacén al abrir el desafío y no al
  completarlo.
- **Una credencial larga**: una clave de proveedor puede pasar de cien caracteres y no debe romper
  la presentación que `004` dejó (`004`-SC-005).

## Requirements *(mandatory)*

### Functional Requirements

**Que el camino termine**

- **FR-001**: Cuando el desafío pide una credencial, la interfaz MUST ofrecer **dónde escribirla** y
  un modo de enviarla.
- **FR-002**: Enviar la credencial MUST completar el desafío: la credencial queda bajo custodia del
  almacén del sistema y la cuenta pasa a estar conectada.
- **FR-003**: Al terminar, la cuenta MUST aparecer en la lista de conectadas **sin reiniciar** y sin
  que el usuario tenga que refrescar nada.
- **FR-004**: Ningún comando del backend MUST quedar registrado sin que la interfaz lo invoque. Un
  comando que nadie llama compila, pasa sus tests y no hace nada — que es exactamente cómo este
  defecto sobrevivió desde `003`.

**Que el fallo se entienda**

- **FR-005**: Cada motivo por el que la conexión no puede completarse MUST tener texto propio y
  **salida declarada**: qué puede hacer el usuario a continuación.
- **FR-006**: Una credencial rechazada MUST NOT consumir el desafío: se puede corregir sin rehacer
  el formulario.
- **FR-007**: Un desafío que ya no está en curso MUST decirse como tal, distinguido de una
  credencial equivocada.

**Que no se acumule lo empezado**

- **FR-008**: Un desafío MUST poder abandonarse **desde un control visible**, y al hacerlo MUST NOT
  quedar retenido nada de lo que llevaba.
- **FR-009**: Pedir conectar de nuevo MUST **reemplazar** al desafío anterior de esa cuenta, no
  añadirse a él. Una cuenta MUST NOT tener dos desafíos en curso a la vez — hoy pulsar «Conectar»
  dos veces deja dos dentro, y solo sale el que se complete.

**Lo que no puede romperse**

- **FR-010**: Toda cadena visible MUST salir del catálogo, en los dos idiomas — **incluidas las
  que hoy nacen en el backend**. Las instrucciones del desafío están escritas como literal en
  español dentro del código del comando, y hoy no se ven solo porque la interfaz las ignora.
- **FR-010a**: Lo que el desafío lleva para el usuario MUST ser una **clave de catálogo**, no texto
  ya redactado. El núcleo sabe qué hay que decir; en qué idioma decirlo es de la piel, que es la
  única que conoce el idioma elegido. Una cadena redactada en el núcleo solo puede estar en un
  idioma, y en este proyecto eso ya tiene nombre: una cadena fuera del catálogo.
- **FR-011**: La credencial MUST NOT aparecer en registros, eventos ni mensajes de error, ni
  persistir en la interfaz una vez enviada (`003`-FR-002).

**Que no queden caminos invisibles**

- **FR-012**: Un camino del núcleo que la interfaz **no puede alcanzar** MUST estar declarado como
  tal, con el issue que lo recoge, y la declaración MUST comprobarse de forma automática. Vale para
  una rama elegida por un parámetro que la interfaz nunca varía, no solo para un comando sin
  invocar. Es la generalización de lo que produjo este spec: algo construido, con sus tests en
  verde, sin que nadie pudiera llegar a ello durante dos ciclos.

### Key Entities

- **Desafío**: una conexión empezada y no terminada. Tiene identificador, vía —credencial directa
  o autorización delegada— y **la clave de catálogo** de lo que se espera del usuario, no su texto.
  Vive mientras dure el intento.
- **Vía de conexión**: cómo se obtiene el secreto. Lo que cambia entre vías es **cómo se obtiene**,
  no qué se hace con él: custodia, uso y revocación son idénticas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona que no conoce la aplicación conecta una cuenta **de punta a punta** sin
  ayuda. Es el criterio que `004` dejó declarado en fallo.
- **SC-002**: **Cero** comandos del backend sin quien los invoque, verificable de forma automática.
- **SC-003**: **Cero** cadenas visibles fuera del catálogo, contando también las que nacen en el
  backend: lo que cruza del núcleo a la piel son claves, nunca frases. Verificable de forma
  automática.
- **SC-004**: Todo motivo de fallo de la conexión tiene texto y salida en **los dos idiomas**,
  verificable de forma automática.
- **SC-005**: Desde que se envía la credencial hasta que la cuenta aparece o se explica el fallo,
  el usuario **nunca** se queda sin señal de que algo está ocurriendo: el control de envío declara
  que está en curso y no admite un segundo envío mientras tanto.
- **SC-006**: Tras cancelar un desafío, **cero** desafíos quedan en curso. Y en ningún momento hay
  **más de uno** por cuenta.
- **SC-007**: **Cero** caminos del núcleo alcanzables solo en teoría: o la interfaz los alcanza, o
  están declarados con su issue. Verificable de forma automática.

## Assumptions

- **La vía delegada queda fuera de este ciclo, y declarada.** Su código existe y funciona; lo que
  no existe es la puerta: la interfaz pide siempre la vía directa. Este ciclo cierra el camino que
  la gente sí encuentra, y **hace visible** que el otro no tiene entrada — ver FR-012. Construirle
  una puerta ahora sería además apostar a que sobreviva a #56, que puede cambiar por dónde se entra
  a conectar.
- El almacén de credenciales sigue siendo **el del sistema, sin alternativa**: si no lo hay, se
  dice y se sigue en local. No se inventa un sustituto, que es lo que `003`-FR-004 decidió.
- No se rediseña la puerta. Que el formulario sea uno o dos, y que aparezca donde aparece, es
  materia de #56: **una conexión que se abre y nunca se cierra está rota bajo cualquier diseño**.
- La presentación que dejó `004` se conserva. Lo que este spec añade —dónde escribir la credencial,
  el desenlace— vive dentro de esa superficie.

### Dependencies

- `003`-FR-001 a FR-004 (obtener, custodiar y declarar la ausencia de almacén) — este spec los
  **ejercita por primera vez desde la interfaz**.
- `004` — la superficie donde esto ocurre, y su SC-001, que este spec desbloquea.

### Out of scope

- **Construir la puerta de la autorización delegada.** Su motor existe en el núcleo desde `003`;
  este ciclo declara que no tiene entrada (FR-012) y le abre issue, pero no la cablea.
- Rediseñar cómo se añade un proveedor, o unificar local y remoto — eso es #56.
- Autodetectar proveedores ya instalados — #56 también.
- Cambiar qué se hace con la credencial una vez obtenida: custodia, uso y revocación son de `003` y
  funcionan.
- El silencio de la sesión cuando el backend muere a mitad (#51). Es la misma familia —no saber si
  algo sigue vivo— pero ocurre en otra superficie.
