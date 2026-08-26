# Feature Specification: Conectividad y reparto de modelos

**Feature Branch**: `feat/conectividad-y-tiers`
**Created**: 2026-08-25
**Status**: Draft
**Input**: issue [#42](https://github.com/jorelcb/codify/issues/42) — decidido dentro del alcance de v1 el 2026-08-25.

## Por qué existe este spec

Dos cosas que el proyecto se debe a sí mismo y que no están:

- **`001`-FR-017**, un **MUST incumplido**: usar un tier económico para lo frecuente y reservar el
  caro para la generación pesada. Hoy hay un solo tier, así que enrutar no significa nada.
- La **constitución** exige ser *provider-agnóstico*: «local llama.cpp/Ollama **+ remotos vía
  OAuth/"conectar cuenta"»*. El adapter remoto no existe.

Van juntos porque separados no sirven: el reparto no tiene entre qué repartir sin un segundo
proveedor, y un proveedor remoto sin reparto obliga a elegir a mano en cada tarea.

> **Lo que este spec NO relaja.** El cero-egress en modo local es **[NON-NEGOTIABLE]** en la
> constitución y hoy es *estructural*: el registro rechaza proveedores no locales al construirse y
> el proveedor local rechaza endpoints que no sean loopback. Ninguna de las dos comprobaciones
> depende de que alguien se acuerde. Este spec introduce un modo que **sí** sale a la red; la
> parte difícil no es añadirlo, es que el modo local siga siendo **incapaz** de egress, no
> meramente configurado para no hacerlo.

## Clarifications

### Session 2026-08-26

- Q: ¿Qué significa «conectar cuenta», dado que los frontier mayoritarios autentican con clave y no con OAuth device-flow? → A: **Ambos, según el proveedor.** Device-flow donde exista; clave introducida una vez y custodiada en el almacén del sistema donde no. Lo que `research.md` de `001` descartó fue quedarse *solo* con claves —«incumple D3»—, y D3 es *auth pluggable*: varios mecanismos, no la ausencia de uno.
- Q: Con remotos permitidos, ¿qué contenido del repositorio puede salir? → A: **Todo el material reunido.** El consentimiento es por **modo**, explícito y previo. Se descartó una lista de permitidos por fuente: produce control aparente —un archivo mal clasificado basta para filtrar— y la respuesta real a «esto no debe salir» ya existe y es de primera clase, el modo local.
- Q: ¿Cuándo se elige el modo, si de eso depende que el cero-egress siga siendo estructural? → A: **Al cambiarlo se rearma el grafo de objetos.** Un grafo local no contiene adapter remoto, así que no hay ruta que auditar. Elegirlo por sesión obligaría a un grafo capaz de salir a la red, y «imposible» se degradaría a «no usado esta vez».

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conectar un proveedor remoto una sola vez (Priority: P1)

El usuario conecta una cuenta de un proveedor de modelos desde la aplicación: autoriza en su
navegador si el proveedor lo permite, o introduce su credencial **una sola vez** si no. A partir
de ahí la custodia el sistema operativo, y el usuario no vuelve a verla ni a escribirla.

**Why this priority**: sin esto no hay segundo proveedor, y sin segundo proveedor no hay nada más
en este spec. Es el MVP.

**Independent Test**: conectar una cuenta y comprobar que una generación posterior la usa, sin que
la credencial aparezca en disco en claro, en el registro ni en la interfaz.

**Acceptance Scenarios**:

1. **Given** un proveedor que ofrece autorización delegada, **When** el usuario lo conecta,
   **Then** la aplicación le da un código y una dirección, él autoriza fuera, y al volver la
   cuenta figura conectada.
1b. **Given** un proveedor que solo admite credencial, **When** el usuario la introduce,
   **Then** queda guardada por el sistema y **no vuelve a mostrarse** — ni en la interfaz, ni en
   la configuración, ni en el registro.
2. **Given** una cuenta conectada, **When** el usuario reinicia la aplicación, **Then** sigue
   conectada sin volver a autorizar.
3. **Given** una cuenta conectada, **When** el usuario la desconecta, **Then** la credencial se
   elimina del almacén del sistema y ninguna tarea posterior puede usarla.
4. **Given** una autorización que el usuario abandona o deniega, **When** vuelve a la aplicación,
   **Then** ve qué pasó y puede reintentar, sin quedar en un estado a medias.

---

### User Story 2 - Que lo barato haga lo frecuente (Priority: P2)

El usuario itera en el refinamiento sin pagar el precio de la generación pesada. El sistema manda
cada tarea al tier que le corresponde, y **dice cuál usó**.

**Why this priority**: es el `001`-FR-017 incumplido, y el que convierte «tengo dos proveedores»
en un beneficio. Depende de US1.

**Independent Test**: con dos proveedores conectados, comprobar que una pregunta de refinamiento
va al económico y una generación de artefacto al de mayor capacidad, y que ambas cosas se pueden
comprobar sin leer el código.

**Acceptance Scenarios**:

1. **Given** dos tiers disponibles, **When** el usuario refina, **Then** la interacción va al tier
   económico.
2. **Given** dos tiers disponibles, **When** se genera un artefacto grounded, **Then** va al de
   mayor capacidad.
3. **Given** el tier de mayor capacidad no disponible, **When** se genera, **Then** el sistema
   degrada al disponible **y lo declara** (`001`-FR-018, ya implementado).
4. **Given** una sesión terminada, **When** el usuario mira el registro, **Then** puede saber qué
   tier atendió cada cosa.

---

### User Story 3 - Saber, y poder demostrar, qué sale del equipo (Priority: P3)

El usuario elige entre trabajar 100 % local o permitir el uso de proveedores remotos, y en
cualquier momento puede saber qué salió y qué no.

**Why this priority**: es lo que permite que exista US1 sin romper la promesa del producto. No
bloquea a las otras dos, pero sin ella el modo local deja de ser una garantía y pasa a ser una
opción de configuración.

**Independent Test**: en modo local, comprobar que **no existe** una ruta por la que el contenido
del repositorio pueda salir — no que no salga en una corrida concreta.

**Acceptance Scenarios**:

1. **Given** modo local, **When** el usuario intenta usar un proveedor remoto, **Then** el sistema
   lo impide y explica por qué, en vez de salir a la red.
2. **Given** modo con remotos permitido, **When** arranca la sesión, **Then** el usuario ve **antes
   de empezar** qué proveedores podrían recibir contenido del repositorio.
3. **Given** cualquier modo, **When** la sesión termina, **Then** el usuario puede enumerar qué
   proveedor atendió cada tarea.

---

### Edge Cases

- **La credencial caduca o se revoca a mitad de sesión**: el sistema lo reporta como lo que es y
  ofrece reconectar; no reintenta en bucle ni presenta el fallo como problema del modelo.
- **El almacén de credenciales del sistema no está disponible** (Linux sin *keyring*, sesión sin
  desbloquear): se dice, y se ofrece seguir en modo local. **No** se cae a guardar la credencial
  en un archivo.
- **Dos proveedores del mismo tier**: se usa uno y se declara cuál; no se reparte carga.
- **Ningún proveedor conectado**: la aplicación sigue siendo usable en local, que es el caso por
  defecto y no un modo degradado.
- **El usuario cambia de modo con una sesión en curso**: la sesión viva termina con el modo con
  el que nació (FR-008b); el grafo se rearma para la siguiente, y se dice.

## Requirements *(mandatory)*

### Functional Requirements

**Conectar y custodiar**

- **FR-001**: El sistema MUST permitir conectar proveedores remotos por **dos vías, según lo que
  el proveedor ofrezca**: autorización delegada en el navegador donde exista, y credencial
  introducida **una sola vez** donde no. El sistema MUST NOT exigir volver a introducirla, ni
  mostrarla después de guardada.
- **FR-002**: Las credenciales MUST guardarse en el almacén seguro del sistema operativo. El
  sistema MUST NOT escribirlas en archivos de configuración, registros ni la interfaz.
- **FR-003**: El sistema MUST permitir listar y **desconectar** cuentas, y desconectar MUST
  eliminar la credencial del almacén.
- **FR-004**: Si el almacén seguro no está disponible, el sistema MUST decirlo y ofrecer continuar
  en modo local; MUST NOT recurrir a un almacenamiento menos seguro.

**Repartir**

- **FR-005**: El sistema MUST enrutar cada tarea al tier que le corresponde —económico para
  interacciones frecuentes y de bajo riesgo, de mayor capacidad para la generación pesada
  grounded— cumpliendo `001`-FR-017.
- **FR-006**: El reparto MUST ser observable: el usuario MUST poder saber qué tier atendió cada
  tarea sin recurrir a herramientas de desarrollo.
- **FR-007**: Cuando el tier pedido no esté disponible, el sistema MUST degradar y **declararlo**,
  reutilizando el mecanismo de `001`-FR-018.

**No romper la promesa**

- **FR-008**: En modo local, el sistema MUST hacer **estructuralmente imposible** el egress: no
  basta con no configurar proveedores remotos, MUST NOT existir una ruta por la que se alcancen.
  Verificable por una comprobación automática que falle si aparece una.
- **FR-008a**: El modo MUST fijarse al **construir el grafo de objetos**, y cambiarlo MUST
  reconstruirlo. Un grafo local MUST NOT contener un adapter capaz de salir a la red — que no
  esté es lo que hace a FR-008 demostrable; que esté y no se use, no.
- **FR-008b**: Una sesión en curso MUST terminar con el modo con el que nació. Cambiar el modo
  MUST NOT afectar a lo que ya está corriendo.
- **FR-009**: Antes de que una sesión pueda enviar contenido del repositorio a un proveedor
  remoto, el usuario MUST haber elegido explícitamente ese modo, y MUST poder ver qué proveedores
  podrían recibirlo. El alcance es **todo el material que la sesión reúna**: el sistema MUST NOT
  ofrecer un permiso parcial por fuente, porque no podría sostener la promesa que insinúa.
- **FR-010**: El sistema MUST registrar qué proveedor atendió cada tarea, de forma que el usuario
  pueda reconstruir después qué salió del equipo y qué no.

### Key Entities

- **Conexión de proveedor**: una cuenta remota autorizada; tiene proveedor, estado (conectada /
  caducada / revocada) y tier declarado. **No contiene la credencial**: la custodia el sistema.
- **Tier**: la clase de capacidad de un proveedor —económico o de mayor capacidad— sobre la que se
  decide el reparto.
- **Modo de sesión**: local (cero egress garantizado) o con remotos permitidos. Elegido por el
  usuario, no inferido.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un usuario conecta una cuenta remota y la usa **sin escribir ni pegar ninguna
  credencial** en la aplicación.
- **SC-002**: **Cero** credenciales recuperables del disco, los registros o la interfaz tras
  conectar una cuenta.
- **SC-003**: En modo local, el egress de contenido del repositorio hacia proveedores remotos es
  **cero**, y esa imposibilidad es **verificable automáticamente** —no por inspección de una
  corrida—. Es `001`-SC-007 sostenido tras introducir el modo remoto.
- **SC-004**: Con dos tiers conectados, el **100 %** de las interacciones de refinamiento van al
  económico y el **100 %** de las generaciones pesadas al de mayor capacidad.
- **SC-005**: Tras una sesión, el usuario puede decir correctamente **qué proveedor atendió cada
  tarea**, solo mirando la aplicación.
- **SC-006**: Desconectar una cuenta impide su uso **inmediatamente**, sin reiniciar.
- **SC-007**: Cambiar de modo **no exige reiniciar** la aplicación, y una sesión en curso no
  cambia de modo a mitad.

## Assumptions

- El usuario tiene cuenta con al menos un proveedor remoto compatible, y un navegador donde
  autorizar.
- El modo local sigue siendo **el caso por defecto**: conectar un remoto es una decisión
  explícita, no el camino recomendado. El producto se define por la garantía local.
- Introducir una credencial una sola vez es aceptable **si el sistema la custodia**; lo que no lo
  es —y FR-002 lo prohíbe— es que quede en un archivo del proyecto, en la configuración o en un
  registro.
- El reparto por tier es **automático según el tipo de tarea**, no configurable tarea por tarea.
  Se asume que quien conecta dos proveedores quiere que el sistema decida; si aparece demanda de
  control fino, se revisita.
- No se reparte carga entre proveedores del mismo tier: con dos, se usa uno y se declara cuál.

### Dependencies

- `001`-FR-018 (degradación declarada entre tiers) — **ya implementado**, PR #38.
- `001`-FR-020 y `001`-SC-007 (modo local cero-egress) — este spec debe sostenerlos, no
  reemplazarlos.
- Almacén de credenciales del sistema operativo.

### Out of scope

- Reparto de carga o *failover* entre proveedores del mismo tier.
- Control de coste o presupuesto por proveedor.
- Elegir el proveedor tarea por tarea desde la interfaz.
- **Reintroducir** la credencial en cada sesión, o mostrarla tras guardarla.
- Permisos de egress por fuente: descartados en la sesión de clarificación del 2026-08-26.
