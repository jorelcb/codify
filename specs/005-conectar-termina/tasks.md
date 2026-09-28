# Tasks — Conectar una cuenta llega a su fin

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) · **Contrato**: [contracts/conexion.md](./contracts/conexion.md) · **Issue**: [#54](https://github.com/jorelcb/codify/issues/54)

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).

## Antes de empezar: qué está hecho y qué engaña

**El backend está entero.** `complete_connection` recibe la credencial, la entrega al conector, la
guarda en el almacén del sistema y crea la conexión. El DTO del desafío ya lleva identificador, vía
e instrucciones. **Nada de eso hay que construirlo.**

**Y aun así hay tres defectos**, que solo se ven recorriendo el camino entero
([research D1, D3, D4](./research.md)):

| | Qué pasa hoy |
|---|---|
| El desafío se consume **antes** de intentar | Una credencial rechazada obliga a rehacer el formulario |
| Las instrucciones son **una frase en español** en el backend | No se ven porque la interfaz ignora el campo, y el test de cadenas sueltas no mira ahí |
| El estado es un **mapa** | Pulsar «Conectar» dos veces deja dos desafíos dentro |

**Lo que este ciclo NO hace**: construir la puerta de la autorización delegada. Su motor existe
desde `003` y no tiene entrada; aquí solo deja de ser invisible (FR-012).

**Formato**: `[ID] [P?] [Story] Descripción con ruta`. `[P]` = otro archivo, sin dependencias
pendientes.

---

## Fase 1: Preparación

- [X] T001 Registrar la línea de base — `cargo test --workspace` (239 verdes) y
      `cargo test -p codify-app --test ui_contract` (22 verdes), para que una caída posterior sea
      atribuible

---

## Fase 2: Fundacional (bloquea a las tres historias)

Dos cambios de tipo. Van antes porque las tres historias construyen encima, y porque el segundo
convierte una invariante en algo que **no se puede escribir**.

### Test primero

- [X] T002 Escribir `toda_instruccion_de_credencial_tiene_texto_en_ambos_idiomas` en
      `crates/codify-app/tests/ui_contract.rs`, recorriendo `InstruccionDeCredencial::all()` contra
      el catálogo. **Debe fallar**: el tipo todavía no existe (R1, FR-010a)

### Implementación

- [X] T003 Crear `InstruccionDeCredencial` en `crates/codify-core/src/application/ports.rs` con
      `code()` y `all()`, y hacer que `Desafio::PideCredencial` lo lleve **en vez de la frase**. Es
      el patrón de `ProviderIssue`, `SessionFailure` y `ConnectionState`
- [X] T004 Cambiar `DirectCredential::new` en
      `crates/codify-core/src/infrastructure/secrets/direct.rs` para recibir el tipo de dominio, y
      **borrar el literal en español** de `connect_provider` en `crates/codify-app/src/commands.rs`
- [X] T005 [P] Añadir las claves de instrucción a `crates/codify-app/src/strings.rs`, en los dos
      idiomas (FR-010, SC-003)
- [X] T006 Sustituir el mapa de desafíos por `Mutex<Option<DesafioEnCurso>>` en
      `crates/codify-app/src/commands.rs`. Dos desafíos a la vez deja de ser un estado que hay que
      impedir y pasa a ser uno que **no se puede nombrar** ([research D3](./research.md))

**Punto de control**: 239 tests siguen verdes, más el de T002. La aplicación levanta y el
formulario sigue donde `004` lo dejó.

---

## Fase 3: User Story 1 — Pegar la credencial y quedar conectado (P1) 🎯 MVP

**Meta**: que alguien con una clave conecte una cuenta y la vea en la lista.

**Prueba independiente**: dar a alguien una credencial y pedirle que conecte, sin explicar nada.

**Depende de**: Fase 2.

### Tests primero

- [X] T007 [P] [US1] Escribir `la_credencial_no_sobrevive_al_envio` en
      `crates/codify-app/tests/ui_contract.rs` — el campo se vacía al enviarse y ningún módulo la
      guarda en una variable. Misma regla que `004` aplicó al modo: sin copia no hay nada que se
      quede atrás. **Debe fallar** (R3, FR-011)
- [X] T008 [US1] Escribir el camino feliz en `crates/codify-app/tests/conexion_termina.rs`: abrir
      desafío → enviar credencial → la cuenta queda conectada y el desafío deja de estar en curso.
      **Debe fallar** (FR-002, FR-003)
- [X] T009 [US1] Escribir `el_envio_no_admite_un_segundo_intento` en
      `crates/codify-app/tests/ui_contract.rs` — el manejador del envío deshabilita su control
      mientras la credencial viaja, y lo vuelve a habilitar al terminar. **Debe fallar** (R7,
      SC-005)

### Implementación

- [X] T010 [US1] Añadir al plegable de `crates/codify-app/ui/index.html` **dónde escribir la
      credencial** y su control de envío, con las instrucciones que llegan del desafío, dentro de
      la superficie que `004` dejó (FR-001, [research D6](./research.md))
- [X] T011 [US1] Cablear el envío en `crates/codify-app/ui/connections.js`: invocar
      `complete_connection` con el identificador del desafío y la credencial. **Es la llamada que
      no existía**, y por la que conectar era imposible
- [X] T012 [US1] Al terminar bien, en `crates/codify-app/ui/connections.js`: refrescar la lista sin
      reiniciar, vaciar el campo de la credencial y devolver el plegable a su estado limpio
      (FR-003, FR-011)
- [X] T013 [P] [US1] Añadir a `crates/codify-app/src/strings.rs` las claves del envío y del
      desenlace, en los dos idiomas
- [X] T014 [US1] Declarar el estado «en curso» en `crates/codify-app/ui/connections.js` y
      `crates/codify-app/ui/styles.css`: mientras la credencial viaja se dice, y **no se admite un
      segundo envío**. Cierra SC-005 y de paso la carrera de dos envíos del mismo desafío
      ([research D5](./research.md))
- [X] T015 [US1] Sacar `complete_connection` de la lista de comandos sin invocar en
      `crates/codify-app/tests/ui_contract.rs` — el test avisa si se olvida, y saldar la deuda es
      parte de darla por cerrada (FR-004, SC-002)

### Cierre de la historia

- [X] T016 [US1] Verificar T007 y T008 por inyección: conservar el valor de la credencial tras
      enviarla en `crates/codify-app/ui/connections.js`; y quitar la llamada a `complete_connection`.
      Revertir cada una. **Un test que no se ha visto fallar no está verificado**

**Punto de control**: se conecta una cuenta de punta a punta y aparece en la lista.

---

## Fase 4: User Story 2 — Saber por qué no se pudo (P2)

**Meta**: que cada fallo diga qué pasó y qué se puede hacer.

**Prueba independiente**: provocar cada fallo y comprobar que la pantalla lo dice y ofrece salida.

**Depende de**: Fase 2. No depende de la Fase 3, aunque comparte archivos con ella.

### Test primero

- [X] T017 [US2] Escribir `una_credencial_rechazada_no_pierde_el_desafio` en
      `crates/codify-app/tests/conexion_termina.rs`: desafío abierto → credencial rechazada →
      **sigue en curso** → credencial buena → conectado. **Hoy no puede pasar**, y ninguno lo
      intentaba (R4, FR-006)

### Implementación

- [X] T018 [US2] Consumir el desafío **solo al terminar bien** en
      `crates/codify-app/src/commands.rs`. Hoy sale del estado antes de intentar nada, así que un
      rechazo lo destruye ([research D4](./research.md)). Cubrir además el caso límite de la
      **credencial vacía**: enviar sin escribir nada no consume el desafío ni llega al proveedor
- [X] T019 [US2] Distinguir los motivos de fallo en `crates/codify-app/src/commands.rs`: credencial
      rechazada · sin almacén · desafío que ya no está en curso. Cada uno con su código estable,
      como `SessionFailure` (FR-005, FR-007). Incluir el caso de que el almacén **desaparezca a
      mitad**: había almacén al abrir el desafío y no al completarlo, que es un fallo distinto de no
      haberlo tenido nunca
- [X] T020 [P] [US2] Dar a cada motivo texto **y salida** en `crates/codify-app/src/strings.rs`, en
      los dos idiomas (SC-004). «Qué puede hacer ahora» no es opcional: sin eso, un fallo se parece al
      defecto que originó este spec
- [X] T021 [US2] Enseñar el motivo y su salida en `crates/codify-app/ui/connections.js`, dejando el
      formulario **utilizable** para corregir sin rehacerlo
- [X] T022 [US2] Verificar T017 por inyección: volver a consumir el desafío antes de intentar, en
      `crates/codify-app/src/commands.rs`. Revertir

**Punto de control**: equivocarse de credencial se corrige sin rehacer el formulario.

---

## Fase 5: User Story 3 — Abandonar sin dejar rastro (P3)

**Meta**: que lo empezado y no terminado no se acumule.

**Prueba independiente**: pedir conectar dos veces y cancelar; comprobar cuántos desafíos quedan.

**Depende de**: Fase 2 (el `Option` es donde vive esta garantía).

### Test primero

- [X] T023 [US3] Escribir `pedir_conectar_dos_veces_deja_uno_solo` en
      `crates/codify-app/tests/conexion_termina.rs`: dos peticiones dejan **uno**; cancelar deja
      **cero**. **Debe fallar** antes de T025 (R5, FR-009, SC-006)

### Implementación

- [X] T024 [US3] Añadir el comando `abandon_connection` en `crates/codify-app/src/commands.rs` y
      registrarlo en `crates/codify-app/src/lib.rs` (FR-008)
- [X] T025 [US3] Hacer que `connect_provider` **reemplace** el desafío anterior en
      `crates/codify-app/src/commands.rs` (FR-009)
- [X] T026 [US3] Añadir el control de cancelar en `crates/codify-app/ui/index.html` y cablearlo en
      `crates/codify-app/ui/connections.js`, volviendo al estado limpio
- [X] T027 [P] [US3] Añadir la clave del control de cancelar a `crates/codify-app/src/strings.rs`,
      en los dos idiomas
- [X] T028 [US3] Verificar T023 por inyección: volver a un mapa en
      `crates/codify-app/src/commands.rs` y comprobar que **el test lo nota** — no el compilador.
      Revertir

**Punto de control**: cancelar deja el sistema como estaba; pulsar dos veces deja un solo desafío.

---

## Fase 6: Que no queden caminos invisibles

No pertenece a ninguna historia: es la regla que sale del propio defecto. Algo construido, con sus
tests en verde, que nadie pudo alcanzar durante dos ciclos.

- [X] T029 Escribir `ningun_camino_del_nucleo_queda_sin_puerta` en
      `crates/codify-app/tests/ui_contract.rs`: para cada selector de rama conocido, los valores
      literales que la interfaz puede enviar; una rama inalcanzable con todos ellos debe estar en
      la lista de **caminos declarados con su issue**, y declarando la lista de selectores
      conocidos —hoy con un solo miembro, el parámetro `delegada` (R2, FR-012, SC-007)
- [X] T030 Abrir el issue de la vía delegada: su motor existe desde `003`, la interfaz manda
      siempre la vía directa, y hasta ahora nada lo decía. **Va antes de T031**, que necesita su
      número
- [X] T031 Declarar la autorización delegada como camino sin puerta en la lista de
      `crates/codify-app/tests/ui_contract.rs`, **con el número de issue que abrió T030**
- [X] T032 Verificar T029 por inyección en `crates/codify-app/tests/ui_contract.rs`: quitar la
      delegada de la lista de declarados; y meter en la lista algo que la interfaz sí alcanza.
      Revertir cada una

---

## Fase 7: Cierre, y lo que no cierra el build

- [X] T033 Dejar el árbol en verde: `cargo test --workspace`, `cargo clippy --workspace
      --all-targets -- -D warnings`, `cargo fmt --all -- --check`. Se esperan **246** tests
- [X] T034 [P] Actualizar
      `specs/003-conectividad-y-tiers/contracts/skin-commands.md` con `abandon_connection`, el
      cambio de `complete_connection` y el campo de instrucción que pasa a ser un código — el
      contrato que cambia es el del crate que se toca
- [X] T035 Marcar `004`-T030 en `specs/004-superficie-de-conexion/tasks.md` según el resultado de
      T037: era inalcanzable por este mismo defecto y queda medible al cerrarlo. Se hace **aquí**
      porque el barrido cruzado corre después del merge, y para entonces ya sería deriva
- [~] T036 Recorrer los cuatro fallos de [quickstart.md](./quickstart.md) con la aplicación
      levantada, **en los dos idiomas**, cambiando el idioma sin reiniciar.
      **Parcial (2026-09-28).** Verificado sobre el árbol de accesibilidad de la app real que el
      camino **existe**: el desafío se abre, la instrucción llega traducida del catálogo, el campo
      de la credencial es de tipo seguro —el sistema lo devuelve como `[redacted]`, así que no se
      expone ni al árbol— y están Enviar y Cancelar. **Queda recorrer los fallos**: conducir la
      ventana por accesibilidad resultó inestable —desaparece del árbol tras un clic, sin crash ni
      log— y no se diagnosticó. Se hace con la persona de T037, que ya tendrá la app delante.
      **El envío con éxito no se ejercitó a propósito**: escribe un secreto en el llavero del
      operador, y eso lo decide quien es dueño del llavero
- [ ] T037 [persona] **SC-001**: alguien que no conoce la aplicación conecta una cuenta de punta a
      punta, sin ayuda. **Es el criterio que `004` dejó declarado en fallo** y no se hereda como
      aprobado: se vuelve a medir ahora que el camino existe. Anotar **sus palabras** si falla
- [ ] T038 [persona] **SC-005**: nunca hay un instante sin señal al enviar la credencial. Un envío
      mudo es indistinguible de «pulsé y no pasó nada», que es el defecto que originó este spec

> **T037 y T038 no se marcan `[X]` porque el CI pase.** Y T037 cierra además `004`-T030, que quedó
> abierta esperando justo a esto.

---

## Dependencias y orden

```
Fase 1 (línea de base)
   ↓
Fase 2 (los dos cambios de tipo)  ← bloquea a las tres historias
   ↓
   ├── Fase 3 · US1 (P1) ─┐
   ├── Fase 4 · US2 (P2) ─┼→ Fase 6 (caminos invisibles) → Fase 7 (cierre)
   └── Fase 5 · US3 (P3) ─┘
```

### Entre historias

Independientes en lo conceptual, **acopladas por archivo**: las tres tocan `connections.js`,
`commands.rs` y `strings.rs`. En paralelo hay conflictos de texto, no de diseño — pero conviene
hacerlas en orden si va una sola persona.

### Dentro de cada historia

Test en rojo → implementación → **inyección de la violación**. El tercer paso no es opcional: tres
de los defectos de este spec sobrevivieron a suites verdes, y un test de `004` pasó en verde antes
de servir para nada por un fallo de análisis propio.

> **Citar el requisito al escribir la tarea, no al revisarla.** Este hallazgo salió del
> `/speckit-analyze` de `004` y del de `005`. Arreglar las instancias no arregla la costumbre: una
> tarea sin cita está cubierta y no se puede demostrar, que a efectos de una revisión es lo mismo
> que no estarlo.

### Oportunidades de paralelismo

| Fase | `[P]` | Por qué |
|---|---|---|
| 2 | T005 | `strings.rs`, mientras el resto toca el núcleo |
| 3 | T007, T013 | Un test y el catálogo, ajenos al camino |
| 4 | T020 | Catálogo |
| 5 | T027 | Catálogo |
| 7 | T034 | Documento de contrato |

**Cadena crítica**: T003 → T004 → T006. Es el cambio de forma de los tipos, y partirlo deja el
árbol sin compilar entre medias.

---

## Estrategia

### MVP

**Fases 1 + 2 + 3**. Entrega lo que el issue #54 pide en su titular: que conectar termine. Y
desbloquea `004`-SC-001, que lleva un ciclo declarado en fallo.

### Entrega incremental

Cada fase deja el árbol verde. **Pero la Fase 4 no es opcional en la práctica**: sin ella,
equivocarse de credencial obliga a rehacer el formulario, que es el tipo de fricción que hace
abandonar justo en el paso que este spec existe para arreglar.

### Resumen

| Fase | Tareas | Historia |
|---|---|---|
| 1 · Preparación | T001 | — |
| 2 · Fundacional | T002–T006 | — |
| 3 · US1 (P1) | T007–T016 | Pegar la credencial y quedar conectado |
| 4 · US2 (P2) | T017–T022 | Saber por qué no se pudo |
| 5 · US3 (P3) | T023–T028 | Abandonar sin dejar rastro |
| 6 · Caminos invisibles | T029–T032 | — |
| 7 · Cierre | T033–T038 | — |

**Total: 38 tareas.** 36 las cierra el build; **2 necesitan a una persona delante**.
