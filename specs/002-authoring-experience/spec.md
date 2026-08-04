# Feature Specification: La experiencia de authoring — ver, entender y decidir

**Feature Branch**: `002-authoring-experience`

**Created**: 2026-07-30

**Status**: Draft

**Input**: User description: "US1 con la piel: la experiencia de la aplicación de escritorio — cómo el usuario ve trabajar al agente, revisa y decide sobre los cambios, y distingue de un vistazo qué está fundamentado, qué es tentativo y qué se contradice."

> **Relación con `001-context-authoring`**: aquel spec define **qué hace** el núcleo (leer el repo, seguir referencias, generar y refinar) y **qué superficie** expone la piel (`contracts/tauri-commands.md`). Este spec define **cómo se vive** eso: qué ve el usuario, en qué orden, y cómo decide. Los prerrequisitos técnicos (scaffold de la piel y cableado de comandos/eventos, tareas T003 y T029 de 001) **no** son alcance de este documento.

## Clarifications

### Session 2026-07-30

- Q: ¿Modelo de presentación: bloques cronológicos o documento + chat? → A: **Bloques cronológicos** (corriente estilo terminal moderna). Se acepta la tensión que abre con US3 y se cierra con FR-021.
- Q: ¿La aplicación es editor o revisor? → A: **Revisor**: el usuario aprueba, edita la propuesta o rechaza; no escribe directamente sobre los artefactos.
- Q: ¿La aplicación guía la conexión del proveedor de modelo? → A: **Guiado**: detecta el backend local, ayuda a elegir modelo y avisa si falta algo.
- Q: ¿Qué puede hacer el usuario mientras el agente trabaja? → A: La interfaz **sigue siendo utilizable y la sesión es cancelable** en cualquier momento, declarando qué alcanzó a escribirse.
- Q: ¿La sesión sobrevive al cierre de la aplicación? → A: **No hay persistencia** en esta etapa; lo escrito permanece en el repositorio. Marcado como **revisitable** cuando US2 traiga propuestas pendientes de aprobación.
- Q: ¿En qué idioma está la interfaz? → A: **Localizable español/inglés desde v1**, siguiendo el idioma del sistema por defecto. Es independiente del idioma de los artefactos.
- Q: ¿Qué alcance de accesibilidad? → A: **Operable por teclado + marcado semántico**, sin depender del color. Un programa formal de conformidad (WCAG) queda fuera de esta etapa.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ver trabajar al agente sin quedarse a ciegas (Priority: P1)

El usuario apunta la aplicación a un repositorio y **ve al agente trabajar en vivo**: qué archivos abre, qué referencias sigue, qué logra resolver y qué no. No espera frente a un spinner opaco preguntándose si la herramienta está pensando o colgada.

**Why this priority**: la confianza en el resultado se construye viendo el proceso. Un usuario que vio al agente abrir el `SPEC` referenciado cree el contexto que produjo; uno que solo vio una barra de progreso, no. Es además el primer momento en que la herramienta se distingue de la anterior, que no mostraba nada. Sin esto, el resto de la experiencia no tiene sobre qué apoyarse.

**Independent Test**: iniciar una sesión sobre un repositorio de prueba y verificar que el usuario puede narrar, solo mirando la pantalla, qué leyó el agente y qué no pudo resolver — sin abrir logs ni una terminal.

**Acceptance Scenarios**:

1. **Given** una sesión recién iniciada, **When** el agente explora el repositorio, **Then** el usuario ve aparecer en orden las acciones que realiza (qué lista, qué lee, qué referencia sigue), con nombres de archivo reconocibles.
2. **Given** una referencia que no se pudo resolver, **When** el agente la encuentra, **Then** el usuario la ve señalada como no resuelta **con su motivo**, y esa señal permanece visible al final de la sesión.
3. **Given** una sesión en modo local, **When** el usuario la observa, **Then** puede confirmar de un vistazo que está en modo local y que nada salió hacia la nube.
4. **Given** una exploración que agota su presupuesto, **When** termina la ingesta, **Then** el usuario ve **qué quedó sin leer**, no un resultado que aparenta ser completo.

---

### User Story 2 - Decidir sobre los cambios con el diff a la vista (Priority: P2)

El usuario **revisa los cambios propuestos como diffs** y decide: aprobar, editar o rechazar. Lo de bajo riesgo ya está aplicado y visible con opción de revertir; lo de alto impacto espera su decisión. Refina conversando, no llenando un formulario.

**Why this priority**: es el reemplazo directo de la peor experiencia de la herramienta anterior — una secuencia de prompts modales, uno por hueco, con valores por defecto que empujaban al error. P2 porque US1 ya entrega valor observable; esta lo vuelve accionable.

**Independent Test**: partir de un contexto con un supuesto incorrecto, corregirlo conversando, y llegar a un contexto aprobado sin que la aplicación haya exigido responder una pregunta modal por cada hueco.

**Acceptance Scenarios**:

1. **Given** un cambio de alto impacto propuesto, **When** el usuario lo revisa, **Then** ve el diff y la razón del cambio, y **nada se escribe** hasta que decide.
2. **Given** un cambio de bajo riesgo, **When** el agente lo aplica, **Then** el usuario lo ve reflejado y puede **revertirlo** sin deshacer el resto del trabajo.
3. **Given** un diff propuesto, **When** el usuario lo rechaza, **Then** el archivo queda como estaba y la conversación continúa sin perder el hilo.
4. **Given** varias propuestas pendientes, **When** el usuario las revisa, **Then** puede ver cuántas quedan y avanzar entre ellas sin perder el contexto de lo ya decidido.

---

### User Story 3 - Leer el contexto sabiendo en qué se apoya (Priority: P3)

Al leer el contexto generado, el usuario distingue **sin esfuerzo y sin ambigüedad** tres cosas distintas: lo que está fundamentado en una fuente (y en cuál), lo que es tentativo (y por qué no se pudo verificar), y dónde las fuentes se contradicen (y en qué).

**Why this priority**: es la promesa central del producto hecha visible. Si el usuario no puede ver la diferencia, la garantía que el núcleo sostiene en el tipo se pierde en la pantalla. P3 porque depende de que ya exista contenido que leer (US1) y decisiones que tomar (US2).

**Independent Test**: mostrar a alguien que no participó en la sesión un artefacto generado y pedirle que señale qué afirmaciones están verificadas y cuáles no; debe acertar sin instrucción previa.

**Acceptance Scenarios**:

1. **Given** un artefacto con contenido fundamentado y tentativo, **When** el usuario lo lee, **Then** distingue ambos tipos sin tener que buscarlos ni interpretar una convención de texto.
2. **Given** una afirmación fundamentada, **When** el usuario quiere comprobarla, **Then** puede ver **de qué fuente** salió.
3. **Given** una contradicción entre fuentes, **When** el usuario la encuentra, **Then** ve **qué fuentes chocan y en qué**, presentada como una decisión abierta y no como un hecho.
4. **Given** un contexto con puntos tentativos sin atender, **When** el usuario intenta darlo por cerrado, **Then** la aplicación se lo advierte y le permite resolverlos o diferirlos explícitamente.

---

### Edge Cases

- **Repositorio vacío**: no hay nada que mostrar ni que generar → la aplicación conduce una **entrevista** en lugar de mostrar una pantalla vacía o un error.
- **Sesión larga**: la exploración puede producir muchas acciones; el usuario debe poder seguir el hilo sin ahogarse en ruido ni perder lo importante entre lo trivial.
- **Muchas propuestas pendientes**: la revisión no puede degradar a una cola interminable de decisiones (el fallo que este producto viene a corregir).
- **El proveedor de modelo falla o no está disponible**: el usuario entiende qué pasó y qué puede hacer, sin mensajes técnicos crudos ni una pantalla congelada.
- **El usuario cierra la aplicación a mitad de sesión**: la sesión termina (no se reanuda, FR-024); debe quedar claro **antes de cerrar** qué se escribió al repositorio y qué se perderá.
- **Ventana pequeña**: el diff y la conversación siguen siendo utilizables.

## Requirements *(mandatory)*

### Functional Requirements

**Transparencia del trabajo del agente**
- **FR-001**: La aplicación MUST mostrar en vivo las acciones del agente durante la ingesta (qué lista, qué lee, qué referencia sigue), identificadas por nombres reconocibles para el usuario.
- **FR-002**: La aplicación MUST mostrar el estado de la sesión en todo momento (explorando, generando, refinando, cerrada) sin que el usuario tenga que inferirlo.
- **FR-003**: La aplicación MUST mostrar las referencias **no resueltas** con su motivo, y mantenerlas visibles como parte del resultado — no solo como un aviso pasajero.
- **FR-004**: La aplicación MUST declarar lo que quedó **sin leer** cuando la exploración se acota, de modo que el resultado nunca aparente ser más completo de lo que es.
- **FR-005**: La aplicación MUST indicar de forma persistente si la sesión corre en **modo local**, y señalar cualquier intento de salida a la red bloqueado.

**Revisión y decisión**
- **FR-006**: La aplicación MUST presentar cada cambio propuesto como un **diff legible**, acompañado de la razón del cambio.
- **FR-007**: La aplicación MUST impedir que un cambio de alto impacto se escriba antes de la decisión del usuario, y MUST permitir **aprobar, editar o rechazar**.
- **FR-008**: La aplicación MUST mostrar los cambios de bajo riesgo ya aplicados y ofrecer **revertirlos individualmente**.
- **FR-009**: La aplicación MUST permitir refinar el contexto **conversando en lenguaje natural**, sin exigir una interacción modal por cada punto pendiente.
- **FR-010**: La aplicación MUST mostrar cuántas decisiones quedan pendientes y permitir avanzar entre ellas sin perder lo ya decidido.

**Legibilidad del fundamento**
- **FR-011**: La aplicación MUST distinguir visualmente y sin ambigüedad tres estados del contenido: **fundamentado**, **tentativo** y **en contradicción**.
- **FR-012**: La aplicación MUST permitir ver, para un contenido fundamentado, **de qué fuente** proviene.
- **FR-013**: La aplicación MUST mostrar, para una contradicción, **qué fuentes chocan y en qué**, presentándola como decisión abierta.
- **FR-014**: La aplicación MUST advertir al usuario si intenta cerrar la sesión con puntos tentativos **sin atender**, y permitirle resolverlos o diferirlos de forma explícita.

**Alcance y control del usuario**
- **FR-015**: La aplicación MUST permitir al usuario elegir el repositorio objetivo y el modo de la sesión antes de comenzar.
- **FR-016**: La aplicación MUST permitir cambiar el **idioma de los artefactos**, sobrescribiendo el detectado automáticamente.
- **FR-016b**: La **interfaz** MUST estar disponible en **español e inglés**, tomando por defecto el idioma del sistema operativo y permitiendo cambiarlo. Su idioma es **independiente** del de los artefactos: un usuario con el sistema en inglés puede generar contexto en español. Todo texto visible MUST provenir de un catálogo de cadenas, no estar incrustado en la vista.
- **FR-017**: La aplicación MUST dejar claro, en todo momento, **qué se escribió al repositorio** y qué sigue siendo una propuesta.
- **FR-018**: La aplicación es un **revisor, no un editor**: el usuario MUST poder aprobar, **editar la propuesta** del agente o rechazarla, pero NO escribe directamente sobre los artefactos. Cuando quiera aportar contenido propio, lo hace **conversando** y el agente lo convierte en una propuesta revisable. Así todo cambio conserva su rastro y su fundamento declarado.
- **FR-019**: La aplicación MUST **guiar la conexión de un proveedor de modelo**: detectar si hay un backend local disponible, permitir elegir el modelo, y explicar en términos accionables qué falta cuando no hay ninguno — nunca quedarse en silencio sin explicar por qué no ocurre nada.

**Modelo de presentación**
- **FR-020**: La experiencia MUST organizarse como un **flujo cronológico de bloques**: cada acción del agente, cada respuesta y cada propuesta de cambio es un bloque en una corriente que avanza, legible de arriba abajo como el relato de lo que pasó.
- **FR-021**: Como contrapeso del flujo cronológico, la aplicación MUST permitir **ver un artefacto completo en cualquier momento**, con su fundamento visible, sin obligar al usuario a reconstruirlo recorriendo la corriente hacia atrás. *(Cierra la tensión conocida entre FR-020 y la User Story 3.)*

**Trabajo en curso**
- **FR-022**: La interfaz MUST permanecer **utilizable mientras el agente trabaja**: el usuario puede leer la corriente, desplazarse y consultar lo ya producido. La aplicación no se congela durante una sesión que puede durar minutos.
- **FR-023**: El usuario MUST poder **cancelar la sesión en curso** en cualquier momento, y la aplicación MUST declarar entonces **qué alcanzó a escribirse** al repositorio y qué no.
- **FR-024**: La sesión **no persiste** entre ejecuciones: cerrar la aplicación la termina. Lo ya escrito permanece en el repositorio, y al reabrir se inicia una sesión nueva. La aplicación MUST dejar claro, antes de cerrarse con trabajo en curso, **qué se perderá**.

**Accesibilidad y manejo de fallos**
- **FR-025**: Toda acción MUST ser alcanzable **con el teclado**, sin ratón: iniciar y cancelar una sesión, recorrer la corriente, abrir un artefacto y decidir sobre una propuesta.
- **FR-026**: La distinción entre estados de fundamento (grounded / tentativo / contradicción) MUST NOT depender **únicamente del color**: debe sobrevivir a daltonismo y a una captura en escala de grises.
- **FR-027**: La interfaz MUST usar **estructura semántica** (encabezados, listas, regiones, etiquetas de control), de modo que un lector de pantalla obtenga soporte base sin un programa de accesibilidad dedicado.
- **FR-028**: Ante un fallo del proveedor de modelo o de la lectura del repositorio, la aplicación MUST explicar **qué ocurrió y qué puede hacer el usuario**, sin exponer mensajes técnicos crudos ni dejar la interfaz en un estado indeterminado.

### Key Entities

- **Sesión de authoring**: el trabajo en curso sobre un repositorio; tiene un estado observable y un modo (local o híbrido).
- **Actividad del agente**: cada acción observable durante la ingesta (listar, leer, seguir una referencia, no poder resolverla).
- **Propuesta de cambio**: unidad que el usuario revisa; lleva un diff, una razón y un nivel de riesgo que determina si bloquea o no.
- **Decisión**: el veredicto del usuario sobre una propuesta (aprobar, editar, rechazar) y su rastro.
- **Artefacto de contexto**: el documento generado, compuesto de fragmentos con distinto grado de fundamento.
- **Fragmento**: unidad de contenido con su estado — fundamentado (con fuentes), tentativo (con motivo) o en contradicción (con las fuentes en conflicto).
- **Referencia no resuelta**: algo que el agente aludió pero no pudo leer, con el motivo por el que no pudo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un usuario que observó la sesión puede enumerar correctamente **qué fuentes leyó el agente** y **cuáles no pudo resolver**, sin consultar logs ni terminal.
- **SC-002**: Una persona ajena a la sesión identifica correctamente el estado (fundamentado / tentativo / contradicción) de **al menos el 90 %** de los fragmentos que se le señalan, **sin instrucción previa** sobre la convención visual.
- **SC-003**: Llevar un repositorio desde "sin contexto" hasta "contexto aprobado" ocurre **en una sola aplicación**, sin abrir un editor externo ni una terminal.
- **SC-004**: El refinamiento se completa **sin ninguna interacción modal bloqueante por hueco**: las decisiones que se piden son solo las de alto impacto.
- **SC-005**: **Cero** escrituras al repositorio sin que el usuario pueda saber que ocurrieron: todo cambio aplicado es visible y revertible.
- **SC-006**: Un usuario en modo local puede confirmar, **solo mirando la interfaz**, que la sesión no envió nada a la nube.
- **SC-007**: Con la ventana en su tamaño mínimo soportado, el diff y la conversación **siguen siendo utilizables** (sin recorte de contenido ni desplazamiento horizontal).
- **SC-008**: Cancelar una sesión en curso deja el repositorio en un estado que el usuario **puede describir con exactitud** a partir de lo que la aplicación le muestra — nunca en un estado que deba averiguar inspeccionando archivos.
- **SC-009**: La interfaz se muestra íntegramente en el idioma activo: **cero** cadenas sin traducir en español o inglés, verificable recorriendo la aplicación con cada idioma.

## Assumptions

- **Alcance**: este spec cubre la **experiencia** de la aplicación de escritorio. El scaffold de la piel y el cableado de comandos/eventos son prerrequisitos técnicos que viven en `001-context-authoring` (T003, T029) y **no** se re-especifican aquí.
- **La piel no contiene lógica de dominio**: renderiza lo que el núcleo expone y captura decisiones. Toda regla (qué es alto impacto, qué está fundamentado) proviene del núcleo, ya especificado en 001.
- **Comportamientos ya decididos en 001** que esta experiencia debe reflejar, no redefinir: loop curado (bajo riesgo se auto-aplica, alto impacto requiere aprobación, todo revertible), cero-egress en modo local, idioma auto-detectado con override, y que lo no verificable se marca en vez de afirmarse.
- **Un solo usuario, un solo repositorio por sesión**; no hay colaboración concurrente ni multi-proyecto simultáneo en esta etapa.
- **Escritorio**: la experiencia se diseña para pantalla de computador con teclado y ratón; no hay soporte táctil ni móvil.
- **Accesibilidad**: alcance fijado en FR-025 a FR-027 (operable por teclado, sin depender del color, marcado semántico). Un programa formal de conformidad (auditoría con lectores de pantalla, WCAG) queda **fuera** de esta etapa.

### Dependencies

- El núcleo de `001-context-authoring` expone el estado de la sesión, las actividades del agente, las propuestas de cambio y los artefactos con su fundamento (contrato en `001/contracts/tauri-commands.md`).
- Requiere al menos un proveedor de modelo disponible para que haya algo que observar.

## Alternativas de experiencia descartadas

Se registran con su razón y con la señal que justificaría revisitarlas. Ninguna se descartó por
ser mala: se descartaron por no ser la apuesta de esta etapa.

### En lugar de bloques cronológicos (FR-020)

| Alternativa | Qué ofrecía | Por qué no ahora | Qué la haría revisitable |
|---|---|---|---|
| **Documento central + chat lateral** | Los artefactos siempre visibles; ideal para leer el contexto y ver su fundamento (US3), y para revisar diffs en su sitio | Diluye lo que más diferencia al producto en esta etapa: **ver trabajar al agente**. La actividad quedaría en un panel secundario compitiendo por atención | Si al usar la app se observa que la gente pasa más tiempo **leyendo y corrigiendo** el contexto que **observando** cómo se construye, este modelo se vuelve el correcto |
| **Híbrido por fase** (corriente durante la ingesta → documento al generar) | Cada fase con la vista que le sirve | Dos modelos mentales y una transición que hay que diseñar muy bien para no desorientar; complejidad alta para un v1 que aún no sabe cómo se usa la herramienta | Si FR-021 resulta insuficiente y leer artefactos dentro de la corriente se siente forzado, este es el siguiente paso natural |

> **Riesgo asumido y su mitigación**: la corriente cronológica hace que los artefactos generados
> "queden atrás". Se acepta conscientemente y se contrapesa con **FR-021** (ver un artefacto
> completo en cualquier momento). Si esa mitigación no basta en la práctica, la decisión a
> revisitar es esta, no FR-021.

### En lugar de revisor (FR-018)

| Alternativa | Qué ofrecía | Por qué no ahora | Qué la haría revisitable |
|---|---|---|---|
| **Editor**: escribir directamente sobre los artefactos | Más libertad y sensación de control inmediato | Rompe la trazabilidad que sostiene la promesa del producto: el contenido escrito a mano **no tiene fundamento declarado** y habría que reconciliarlo con las propuestas del agente. Se convertiría en un editor de Markdown con un agente al lado, que es otro producto | Si los usuarios terminan copiando el contexto a su editor para retocarlo, la fricción de "todo pasa por una propuesta" está siendo demasiado alta y hay que abrir la edición directa — con una respuesta explícita a qué *groundedness* tiene lo que el humano escribe |

### En lugar de sesión efímera (FR-024)

| Alternativa | Qué ofrecía | Por qué no ahora | Qué la haría revisitable |
|---|---|---|---|
| **Sesión reanudable** (se guarda y continúa al reabrir) | Nada del trabajo se pierde al cerrar | El valor durable ya vive en el **repositorio**: los artefactos escritos permanecen. La sesión es el andamio para producirlos, y persistirla añade estado que hay que versionar y migrar | **US2**: en cuanto existan propuestas pendientes de aprobación, cerrar la aplicación destruiría decisiones del usuario todavía no aplicadas. Ese es el momento de revisitarlo — está marcado explícitamente |
| **Solo historial** (se conserva la auditoría, no se reanuda) | Permite consultar qué pasó en sesiones anteriores | No resuelve la pérdida de trabajo y sí introduce almacenamiento | Si aparece una necesidad real de auditoría entre sesiones (p. ej. justificar ante terceros cómo se generó un contexto) |

### En lugar de onboarding guiado (FR-019)

| Alternativa | Qué ofrecía | Por qué no ahora | Qué la haría revisitable |
|---|---|---|---|
| **Asumir el proveedor ya configurado** | Alcance mínimo para v1 | El modo por defecto del producto es local, y un usuario que abre la app sin backend corriendo **no vería nada ni sabría por qué**. El primer minuto decide si la herramienta se vuelve a abrir | Solo si el público objetivo pasara a ser exclusivamente gente con su entorno ya montado (p. ej. distribución interna en un equipo con setup estándar) |
