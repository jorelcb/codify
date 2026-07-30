# Feature Specification: Context Authoring — de repositorio a contexto vivo (codify-NG)

**Feature Branch**: `001-context-authoring`

**Created**: 2026-07-27

**Status**: Draft

**Input**: User description: "Onboarding de authoring de codify-NG, acotado al LOOP DE CONTEXTO: un agente lee el repo y sigue sus referencias para generar los archivos de contexto, y luego los refina en un loop conversacional donde el usuario aprueba diffs. Catálogo/instalación, init de SDD y monitor de ciclo de vida quedan como etapas futuras controladas, fuera de este spec."

> **Contexto de producto (no es alcance de comportamiento):** codify-NG es un rediseño greenfield. Se entrega como app standalone (Tauri, fase 1) sobre un **core hexagonal, UI-agnóstico y provider-agnóstico**, de modo que la app Tauri sea la primera de varias "pieles" (futuras: MCP/embebido, CLI) sin reescribir el dominio. Este spec describe **qué observa y hace el usuario**, no el stack.

## Clarifications

### Session 2026-07-27

- Q: ¿En qué idioma se generan los archivos de contexto? → A: **Auto-detectar** del contenido existente del repo (README/docs), con **override** explícito del usuario.
- Q: ¿Se soporta un modo totalmente local sin egress a la nube? → A: Sí — **modo 100% local de primera clase** con **cero egress** garantizado cuando se selecciona; el tier frontier/nube es opcional.
- Q: ¿Qué lee el sistema para fundamentar (profundidad de ingesta)? → A: **Docs + señales estructurales + muestreo selectivo de código** (entrypoints, interfaces/puertos, configs); no el repo completo.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Contexto grounded desde el repo y sus referencias (Priority: P1)

El usuario apunta codify a un repositorio existente. El agente **lee el repo y sigue las referencias que encuentra** —README, specs locales, y documentos enlazados por ruta relativa o URL— para reunir contexto amplio antes de escribir nada. Con esa base genera los archivos de contexto (AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md, INTERACTIONS_LOG.md, IDIOMS.md cuando aplica), reflejando las decisiones **reales** documentadas en las fuentes en lugar de rellenar con supuestos genéricos.

**Why this priority**: Es el fallo raíz medido en la auditoría del codify actual: al recibir solo el texto de un archivo y no dereferenciar sus punteros, generó una arquitectura coherente-pero-falsa (inventó broker/BD para un servicio que es event-sourced sobre Temporal). Seguir las referencias es lo que separa "andamiaje plausible" de "contexto fiel". Este story, aun **sin** el refinamiento conversacional, ya supera a la herramienta actual y es el MVP.

**Independent Test**: Apuntar codify a un repo cuyo README referencia una SSOT (p.ej. un `SPEC` hermano y/o docs enlazadas). Verificar que el contexto generado refleja las decisiones de esa SSOT (motor, persistencia, capas reales) y **no** una arquitectura genérica inventada; y que todo lo que no pudo verificarse queda marcado como tentativo, no afirmado como hecho.

**Acceptance Scenarios**:

1. **Given** un repo con `README` que enlaza un `SPEC` local y docs por URL, **When** el usuario inicia el authoring de contexto, **Then** el agente incorpora el contenido de esas referencias (según el alcance de red acordado) y el contexto resultante no contradice la SSOT.
2. **Given** una decisión clave presente en la SSOT (p.ej. "sin broker, event-sourced"), **When** se genera el contexto, **Then** esa decisión aparece correctamente y **ninguna** afirmación contraria se presenta como hecho.
3. **Given** un aspecto que ninguna fuente cubre, **When** se genera el contexto, **Then** ese punto queda **marcado como tentativo/por-definir**, distinguible del contenido grounded.

---

### User Story 2 - Refinamiento conversacional con diffs aprobables (Priority: P2)

Partiendo del contexto generado (que puede tener huecos o supuestos), el usuario lo **refina en un loop conversacional tipo-LLM**: el agente hace preguntas agrupadas y con sentido, propone los cambios como **diffs que el usuario aprueba, edita o rechaza**, e itera. Reemplaza explícitamente la experiencia actual de `codify resolve`: modal, TTY-locked, 37 prompts secuenciales uno-por-uno con defaults sesgados que empujan al error.

**Why this priority**: Es el segundo hallazgo de la auditoría (F-1/F-4/F-7): incluso el único mecanismo de recuperación que codify tiene hoy es de mala UX, no automatizable, y no repara lo que no marca. La calidad de la muestra #2 dependió de que el humano supiera contradecir a la herramienta. P2 porque US1 ya entrega valor; el refinamiento lo eleva a "contexto aprobado por el usuario".

**Independent Test**: Con un contexto que contiene supuestos incorrectos y huecos, correr el loop de refinamiento y llegar a un contexto sin marcadores pendientes mediante aprobación de diffs, en una conversación fluida (sin prompt modal por cada hueco), corrigiendo los supuestos equivocados.

**Acceptance Scenarios**:

1. **Given** un contexto generado con un supuesto equivocado, **When** el usuario lo corrige en la conversación, **Then** el agente propone un diff que integra la corrección naturalmente y **también** ajusta el andamiaje relacionado (nombres/secciones que dependían del supuesto), no solo el hueco puntual.
2. **Given** un diff propuesto, **When** el usuario lo edita o rechaza, **Then** el sistema respeta la decisión y no lo aplica sin aprobación.
3. **Given** varios huecos relacionados, **When** el usuario responde, **Then** el sistema agrupa/aplica en lote y no exige una interacción modal aislada por cada marcador.

---

### User Story 3 - Repo con contexto previo: actualización sin sobrescribir (Priority: P3)

El usuario re-ejecuta el authoring sobre un repo que **ya tiene** archivos de contexto (de una corrida previa o hechos a mano). El sistema los detecta, propone una **actualización como diff** y nunca sobrescribe a ciegas.

**Why this priority**: Habilita el crecimiento controlado y el uso repetido (la puerta hacia la custodia del ciclo de vida en etapas futuras), pero no es necesario para el primer valor. P3.

**Independent Test**: Re-ejecutar sobre un repo con `AGENTS.md` existente y verificar que se presenta un diff de actualización con aprobación, en vez de un reemplazo silencioso.

**Acceptance Scenarios**:

1. **Given** un repo con contexto previo, **When** se inicia el authoring, **Then** el sistema muestra qué cambiaría como diff y pide aprobación antes de escribir.
2. **Given** contenido humano valioso en un archivo existente, **When** se regenera, **Then** ese contenido no se pierde sin que el usuario lo vea y decida.

---

### Edge Cases

- **Repo vacío / sin README ni código:** no hay referencias que seguir → el flujo debe degradar a un modo de **entrevista** (el agente pregunta lo mínimo para arrancar), no fallar ni inventar.
- **Referencia inaccesible o que requiere autenticación** (p.ej. URL a un repo privado): el sistema debe reportarlo explícitamente y continuar con lo disponible, marcando lo no cubierto — nunca alucinar el contenido faltante.
- **Fuentes contradictorias** entre sí: el sistema señala la contradicción al usuario en vez de elegir en silencio.
- **Sin proveedor frontier disponible (offline):** el authoring pesado debe degradar de forma transparente (avisar y ofrecer continuar solo con el tier local, con calidad reducida declarada) en lugar de fallar opacamente.
- **Repo muy grande / monorepo:** el sistema debe acotar qué lee y declarar qué dejó fuera (sin truncamiento silencioso).
- **El usuario rechaza todos los diffs:** el estado del repo no cambia y no se pierde trabajo.

## Requirements *(mandatory)*

### Functional Requirements

**Ingesta y seguimiento de referencias**
- **FR-001**: El sistema MUST fundamentar la generación leyendo: (a) **documentos** (README, specs locales, docs); (b) **señales estructurales** (layout de directorios, manifiestos de dependencias, configuración); y (c) un **muestreo selectivo de código fuente** (entrypoints, interfaces/puertos, configuración) — no necesariamente el repo completo. MUST declarar qué quedó fuera del muestreo (sin truncamiento silencioso).
- **FR-002**: El sistema MUST detectar y **seguir referencias** que aparezcan en el material leído (rutas relativas a otros archivos del repo; documentos enlazados), incorporando su contenido al contexto de trabajo.
- **FR-003**: En v1 el sistema MUST seguir referencias a **URLs externas públicas** (fetch de red sin autenticación), además de las referencias locales/in-repo. Las referencias que requieran **autenticación** (repos/documentos privados) quedan **fuera de v1**: se reportan como no resueltas (FR-004) y su soporte es una **etapa futura** (spec derivado).
- **FR-004**: El sistema MUST declarar explícitamente qué referencias **no** pudo resolver (inaccesibles, requieren auth, fuera de alcance), sin sustituir su contenido por suposiciones.

**Generación de contexto**
- **FR-005**: El sistema MUST generar el conjunto de archivos de contexto: AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md, INTERACTIONS_LOG.md e IDIOMS.md (este último cuando se detecta lenguaje aplicable).
- **FR-006**: El sistema MUST fundamentar las afirmaciones en las fuentes leídas y NO afirmar como hecho ninguna arquitectura/decisión que no haya podido verificar.
- **FR-007**: El sistema MUST marcar de forma distinguible el contenido **inferido o tentativo** frente al contenido **grounded** (verificado contra fuentes), de modo que el usuario y cualquier agente lector sepan qué es firme y qué no.
- **FR-008**: Cuando las fuentes se contradicen, el sistema MUST señalar la contradicción en vez de resolverla en silencio.
- **FR-019**: El sistema MUST **auto-detectar el idioma** de salida a partir del contenido existente del repo (README/docs) y generar los artefactos de contexto en ese idioma, permitiendo un **override explícito** del usuario.

**Refinamiento conversacional**
- **FR-009**: El sistema MUST ofrecer un loop de refinamiento conversacional donde el agente pregunta e itera y el usuario responde en lenguaje natural (no un prompt modal aislado por cada hueco).
- **FR-010**: El sistema MUST representar todo cambio como un **diff visible, auditable y revertible**. Los cambios **ambiguos o de alto impacto** MUST requerir **aprobación explícita** antes de escribir; los de **bajo riesgo** pueden aplicarse y mostrarse para revisión posterior (modelo curado, FR-012). El usuario siempre puede **editar o revertir** cualquier cambio.
- **FR-011**: Al integrar una respuesta del usuario, el sistema MUST ajustar también el **andamiaje dependiente** (nombres de componentes, secciones, flujos que asumían lo corregido), no solo el marcador puntual.
- **FR-012**: El loop MUST ser **curado**: el sistema **auto-aplica** los cambios de bajo riesgo y **solo pide aprobación explícita** en los cambios ambiguos o de alto impacto (todos los cambios permanecen visibles, auditables y revertibles — FR-010). La **clasificación precisa de "bajo riesgo" vs. "alto impacto"** es un criterio a definir en un **spec derivado**.
- **FR-013**: El sistema MUST poder cerrar el refinamiento sin dejar marcadores/huecos sin atender (resueltos o explícitamente diferidos como tentativos).

**Seguridad de escritura / repos existentes**
- **FR-014**: El sistema MUST NOT sobrescribir archivos de contexto existentes sin mostrar el diff y obtener aprobación del usuario.
- **FR-015**: El sistema MUST preservar el trabajo del usuario: ningún contenido humano existente se descarta sin que el usuario lo vea y decida.

**Comportamiento de modelos (agnóstico de proveedor)**
- **FR-016**: El sistema MUST operar contra proveedores de modelo configurables por el usuario, incluyendo la opción de un proveedor **local**, sin acoplarse a un proveedor específico.
- **FR-017**: El sistema MUST poder usar un tier **económico/de baja latencia** para las interacciones de alta frecuencia y bajo riesgo (preguntas de refinamiento, clasificación) y reservar el tier de mayor capacidad para la generación pesada grounded, de forma que iterar tenga fricción/costo marginal bajos.
- **FR-018**: Si el tier de mayor capacidad no está disponible, el sistema MUST degradar de forma transparente (avisar y declarar la calidad reducida) en lugar de fallar de forma opaca.
- **FR-020**: El sistema MUST soportar un **modo totalmente local** en el que, al seleccionarlo, **ningún** contenido del repositorio ni de sus referencias sale hacia servicios de nube (cero egress garantizado); el tier frontier/nube es **opcional** y solo se usa con configuración/consentimiento explícito.

### Key Entities

- **Repositorio objetivo**: el proyecto sobre el que se autora el contexto; puede estar vacío o existente.
- **Referencia**: un documento aludido desde el material leído — ruta local in-repo o enlace externo; tiene un estado (resuelta / inaccesible / fuera-de-alcance).
- **Artefacto de contexto**: cada archivo generado (AGENTS/CONTEXT/DEVELOPMENT_GUIDE/INTERACTIONS_LOG/IDIOMS); contiene contenido grounded y/o tentativo.
- **Marcador de incertidumbre**: señal que distingue contenido inferido/por-definir del grounded.
- **Propuesta de cambio (diff)**: unidad de cambio presentada al usuario para aprobar/editar/rechazar.
- **Decisión de aprobación**: el resultado de la revisión del usuario sobre un diff.
- **Proveedor de modelo (tier)**: origen de inferencia configurable (local económico / frontier), seleccionado por tipo de tarea.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Dado un repo cuyo README referencia una SSOT, ≥90% de las afirmaciones arquitectónicas del contexto generado son **consistentes con esa SSOT** (vs. el baseline actual, que las contradice/aluciona).
- **SC-002**: **0** afirmaciones no verificadas se presentan como hecho: el 100% del contenido no-grounded queda marcado como tentativo.
- **SC-003**: Un usuario lleva un repo desde "sin contexto" hasta "contexto aprobado por él" en **una sola sesión, sin cambiar de herramienta**. Objetivo de tiempo: **p50 < 10 min para un repo ≤ 2.000 archivos en modo local** — **indicativo, no-gate de v1** (se re-cuantifica cuando exista telemetría real).
- **SC-004**: El refinamiento se completa mediante **conversación + aprobación de diffs**, sin exigir una interacción modal aislada por cada hueco; al cerrar quedan **0 marcadores sin atender**.
- **SC-005**: Re-ejecutar sobre un repo con contexto previo produce **0 sobrescrituras silenciosas** (siempre diff + aprobación).
- **SC-006**: El sistema **nunca inventa** el contenido de una referencia no resuelta: el 100% de las referencias inaccesibles se reportan como tales.
- **SC-007**: En **modo totalmente local**, el egress de contenido del repo/referencias hacia servicios de nube es **cero** (verificable: 0 bytes salientes a proveedores de nube durante una sesión de authoring).

## Assumptions

- **Alcance acotado (crecimiento controlado):** este spec cubre **solo el loop de contexto**. Quedan **fuera de alcance, como etapas futuras en specs derivados**: (a) equipar con catálogo (skills/plugins/hooks, fuentes locales + marketplaces externos); (b) inicialización/configuración del framework SDD (spec-kit u openspec); (c) monitor de ciclo de vida (hook post-commit que evalúa staleness). Se agregan como capacidades nuevas sobre el mismo core, sin reescribir el dominio.
- **Modelo de interacción surface-agnostic** a nivel de comportamiento: agente lee/pregunta/itera; el usuario aprueba diffs. La superficie de fase 1 es la app Tauri, pero el comportamiento aquí descrito no depende de ella.
- **Reparto de modelos:** el usuario configura sus proveedores; el sistema enruta por tipo de tarea (local económico para alta frecuencia; frontier para authoring pesado). No se embebe un modelo en el binario.
- **Conjunto de artefactos de contexto** = AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md, INTERACTIONS_LOG.md, IDIOMS.md; IDIOMS solo cuando hay lenguaje detectable aplicable.
- **Repo vacío** → modo entrevista (mínimas preguntas para arrancar), no fallo ni invención.
- **SDD por defecto** (cuando se aborde en etapa futura) = spec-kit, con opción de openspec — contexto, no alcance de este spec.

### Dependencies

- Acceso de lectura al repositorio objetivo.
- Acceso a uno o más proveedores de modelo (al menos uno; local y/o frontier).
- Para seguir referencias externas **públicas**: acceso de red (FR-003). Referencias con autenticación (privadas) quedan fuera de v1.
