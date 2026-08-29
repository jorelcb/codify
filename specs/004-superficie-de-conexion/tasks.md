# Tasks — La superficie de conexión y modo

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) · **Contrato**: [contracts/ui-contract.md](./contracts/ui-contract.md) · **Issue**: [#52](https://github.com/jorelcb/codify/issues/52)

> **Estado de alto nivel y dependencias vivas**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9).

## Antes de empezar: dos cosas que conviene saber

**1. Esto no es solo maquetación.** Al planificar apareció que el modo tiene **tres** dueños y que
el tercero rompe `003`-FR-008a: `start_session` decide el modo de la sesión desde `request.local`,
que la interfaz fija a `true` y no reasigna jamás. `set_mode` escribe en un estado que nadie lee.
Toda la Fase 4 existe por eso ([research.md D2](./research.md)).

**2. Los tres criterios que no puede cerrar el build.** SC-001, SC-003 y SC-004 miden si algo **se
entiende**. Están en la Fase 6 marcados `[persona]` y **no se marcan `[X]` porque el CI pase** —
justo lo que dejó pasar el defecto original: 231 tests verdes sobre una pantalla ilegible.

**Formato**: `[ID] [P?] [Story] Descripción con ruta`. `[P]` = paralelizable (otro archivo, sin
dependencias pendientes).

**Rutas**: todo bajo `crates/codify-app/` salvo donde se diga.

---

## Fase 1: Preparación

Sin dependencias que instalar: esta feature no añade ninguna. Lo único que hace falta es una línea
de base a la que atribuir cualquier regresión.

- [X] T001 Registrar la línea de base — `cargo test --workspace` (231 verdes) y
      `cargo test -p codify-app --test ui_contract` (14 verdes) — anotando los números en el commit
      de arranque, para que una caída posterior sea atribuible

---

## Fase 2: Fundacional (bloquea a las tres historias)

**Por qué bloquea**: las tres historias editan la misma sección. Separarla una vez, antes, evita
tres reescrituras que se pisan.

- [X] T002 Separar `#conexiones` de la barra de estado en `ui/index.html`: deja de llevar
      `class="provider"` y pasa a su propia clase; el `<h2>` deja de estar oculto y encabeza la
      sección (FR-001)
- [X] T003 Abrir el bloque de presentación propia de la superficie en `ui/styles.css`: disposición
      en bloque con las tres partes en orden —modo, cuentas, formulario—, **sin** heredar de
      `.provider` (FR-001, [research D4](./research.md))

**Punto de control**: la aplicación levanta y la sección se ve como una zona distinta de la barra
de estado. Los 14 tests siguen verdes; el de R5 aún no existe.

---

## Fase 3: User Story 1 — Conectar sin adivinar qué va en cada campo (P1) 🎯 MVP

**Meta**: que alguien que no conoce la aplicación conecte un proveedor sin preguntar qué va en
ningún campo.

**Prueba independiente**: poner a esa persona delante y pedírselo, sin explicar nada. Si pregunta
una sola vez, la historia falla.

### Test primero

- [X] T004 [US1] Escribir `los_campos_del_formulario_caben_en_la_ventana_minima` en
      `tests/ui_contract.rs` — cada campo de texto del formulario declara `min-width` en `ch` de al
      menos 24, y su contenedor envuelve. **Debe fallar** al escribirlo (R4, SC-005)
- [X] T005 [US1] Escribir `la_superficie_de_conexion_tiene_forma_propia` en
      `tests/ui_contract.rs` — `#conexiones` no lleva la clase de la barra de estado, el formulario
      está en un contenedor plegado por defecto, y la lista de cuentas queda fuera de él. **Debe
      fallar** (R5, FR-001, FR-002a, FR-004)

### Implementación

- [X] T006 [P] [US1] Añadir `connection.submit` («Conectar» / «Connect») a `src/strings.rs` y
      repuntar el botón `#conectar` de `ui/index.html` a esa clave, dejando `connection.add`
      —«Conectar un proveedor»— libre para el `<summary>` que despliega el formulario
- [X] T007 [US1] Envolver el formulario de conexión en `<details><summary>` en `ui/index.html`, de
      modo que quede plegado por defecto (FR-002a, [research D5](./research.md))
- [X] T008 [US1] Dar a cada campo su etiqueta asociada e inequívoca en `ui/index.html`, en el orden
      en que hay que rellenarlos, y separar visiblemente la lista de cuentas del formulario
      (FR-002, FR-004)
- [X] T009 [US1] Estilar el formulario en `ui/styles.css`: `min-width: 24ch` en los campos de
      texto y contenedor que envuelve, para que el mínimo no produzca desbordamiento (SC-005,
      [research D3](./research.md)). Cubrir además los dos casos límite que el mínimo **no**
      resuelve: un `max-width` que impida que un valor largo empuje a los demás fuera de la fila, y
      una lista de cuentas que crezca con su propio desplazamiento en vez de empujar al formulario
      fuera de la vista
- [X] T010 [US1] Verificar T004 inyectando la violación en `ui/styles.css`: bajar un campo a
      `20ch`, comprobar que el test cae, revertir; repetir quitando el `flex-wrap`. Verificar
      también R5 devolviéndole a `#conexiones` la clase `provider` en `ui/index.html`. **Un test
      que no se ha visto fallar no está verificado**

**Punto de control**: sin cuentas conectadas, la pantalla enseña el modo y «ninguna cuenta», y
ningún campo de formulario ocupa sitio. Al pulsar el desplegable aparece el formulario con sus
etiquetas.

---

## Fase 4: User Story 2 — El modo se ve como lo que es (P2)

**Meta**: que el control que decide si algo sale del equipo no se parezca a un ajuste más — y que
**diga la verdad**, que hoy no la dice.

**Prueba independiente**: pedir a alguien que señale, sin explicaciones, qué control decide si algo
sale de su equipo.

**Depende de**: Fase 2. No depende de la Fase 3.

### Tests primero

- [X] T011 [P] [US2] Escribir `el_modo_no_puede_discrepar_entre_sus_dos_superficies` en
      `tests/ui_contract.rs` — un solo sitio escribe `dataset.mode`, un solo sitio escribe el
      `checked` de `#modo-local`, y ambos son la misma función. **Debe fallar** (R2, SC-006)
- [X] T012 [P] [US2] Escribir `la_interfaz_no_tiene_su_propia_idea_del_modo` en
      `tests/ui_contract.rs` — ningún módulo declara estado de modo propio. **Debe fallar hoy por
      `ui.local`** (R3, FR-003a)

### Implementación — la fuente única

- [X] T013 [US2] Añadir el comando `mode() -> ModeDto` y hacer que `set_mode` devuelva el modo
      resultante, en `src/commands.rs` (contrato §Comandos)
- [X] T014 [US2] Hacer que `start_session` lea `AppState.mode` en vez de `request.local`, y
      **eliminar** `local` de `StartSessionRequest`, en `src/commands.rs`. Es lo que repara
      `003`-FR-008a: hasta ahora el modo se guardaba y no se aplicaba
- [X] T015 [US2] Concentrar en `ui/main.js` la **única** función que pinta el modo, pintando las
      dos superficies —insignia y casilla— desde el valor que devuelve el backend; borrar
      `ui.local` (FR-003a)
- [X] T016 [US2] Quitar a `ui/connections.js` toda opinión sobre el modo: el manejador de la
      casilla invoca `set_mode` y delega el repintado en la función de T015
- [X] T017 [US2] Comprobar que `provider.refresh(...)` y cualquier otro consumidor del modo en
      `ui/main.js` leen la fuente única, no una copia — el defecto se propagó por ahí

### Implementación — la presentación

- [X] T018 [P] [US2] Agrupar el control de modo en `ui/index.html` con un nombre accesible
      propio. **Se resolvió sin clave nueva**: `aria-labelledby` apunta al encabezado visible del
      grupo, así que el nombre tiene un solo dueño. La clave `a11y.mode_group` que el plan preveía
      habría sido una segunda copia del mismo texto — el defecto que este spec vino a quitar
- [X] T019 [US2] Distinguir el modo del resto en `ui/styles.css`: se identifica como decisión sin
      leer las etiquetas completas, y permanece visible con el formulario plegado (FR-003)
- [X] T020 [US2] Colocar los receptores posibles **junto al control de modo** cuando el modo
      admite remotos y hay cuentas conectadas, en `ui/connections.js` y `ui/index.html`
      (`003`-FR-009)

### Cierre de la historia

- [X] T021 [US2] Verificar T011 y T012 inyectando: pintar la insignia desde `connections.js` (debe
      caer R2) y reintroducir `local: true` en el estado de `main.js` (debe caer R3). Revertir cada
      una
- [X] T022 [P] [US2] Actualizar
      `specs/003-conectividad-y-tiers/contracts/tauri-commands.md` con la firma nueva de
      `set_mode`, el comando `mode()` y la desaparición de `request.local` — el contrato que cambia
      es el del crate que se toca, no el del spec en que se trabaja

**Punto de control**: desmarcar «solo local» cambia la insignia **y** el modo de la siguiente
sesión. Antes de este ciclo no cambiaba ninguna de las dos cosas.

---

## Fase 5: User Story 3 — Saber dónde estás navegando sin ver (P3)

**Meta**: que ninguna región suene igual que otra.

**Prueba independiente**: enumerar las regiones y comprobar que ninguna comparte nombre.

**Depende de**: Fase 2. No depende de las Fases 3 ni 4.

### Test primero

- [X] T023 [US3] Escribir `ningun_par_de_regiones_comparte_nombre_accesible` en
      `tests/ui_contract.rs` — recoge las regiones de `index.html`, resuelve su `data-i18n-aria`
      contra `src/strings.rs`, y falla si se repite **una clave o un texto**, en cualquiera de los
      dos idiomas. **Debe fallar con los dos pares que hay hoy** (R1, FR-005, FR-006, SC-002)

### Implementación

- [X] T024 [P] [US3] Añadir `a11y.connections_region` y `a11y.prefs_region` a `src/strings.rs`, en
      los dos idiomas
- [X] T025 [US3] Cablearlas en `ui/index.html`: `#conexiones` deja `a11y.provider_region` y
      `footer.prefs` deja `a11y.toolbar_region` ([data-model.md §Región](./data-model.md))
- [X] T026 [US3] Verificar T023 inyectando la violación en `ui/index.html`: dar a `#applied` el
      nombre de `#decide`, comprobar que el test cae, revertir

**Punto de control**: siete regiones, siete nombres distintos, en los dos idiomas.

---

## Fase 6: Cierre y lo que no puede cerrar el build

- [X] T027 Dejar el árbol en verde: `cargo test --workspace`, `cargo clippy --workspace
      --all-targets -- -D warnings`, `cargo fmt --all -- --check`. Se esperan **19** tests de
      contrato de interfaz. Es aquí donde se cierra **FR-007**: no tiene tarea propia porque no
      pide construir nada —pide **no romper** las garantías de catálogo de `002`—, y quien lo
      comprueba son los tests que ya existen
- [X] T028 Comprobación de humo del modo con la aplicación levantada, según
      [quickstart.md](./quickstart.md): la insignia cambia al desmarcar, y la sesión siguiente se
      arma en híbrido. Es la demostración de punta a punta de que `003`-FR-008a se cumple
- [ ] T029 Abrir issue por el defecto de `003` que este ciclo consume —`set_mode` escribía en un
      estado que nadie leía— para que el arreglo quede trazable fuera de este spec
- [ ] T029 Abrir issue por el defecto de `003` que este ciclo consume —`set_mode` escribía en un
      estado que nadie leía— para que el arreglo quede trazable fuera de este spec
- [ ] T030 [persona] **SC-001 — FALLA (2026-08-29).** Encontró «Connect a provider». **No supo qué
      hacer al empezar** —ni el autor—. Escribió `claude` en *Name*, pulsó **Connect**: no ocurre
      nada. Escribió algo en *Provider address*, pulsó Connect: tampoco. **Causa raíz: conectar un
      proveedor es imposible.** `connect_provider` devuelve un desafío, la interfaz enseña una
      frase, y ahí acaba: `complete_connection` está registrado en el backend y **ningún JavaScript
      lo invoca**. No hay campo donde escribir la credencial. Ningún texto de ayuda arregla esto —
      es un hueco funcional de `003`
- [ ] T031 [persona] **SC-003 — FALLA (2026-08-29).** Los nombres se leen y son distintos, pero
      (a) la numeración del rotor no corresponde: tres entradas con «(2.)»; (b) `Generated File`
      apunta a un espacio diminuto e invisible —el diálogo de artefacto lleva `role="region"` y
      figura como landmark **estando cerrado**—; (c) se ve apeñuscado y superpuesto con otras
      secciones. **El test R1 pasaba mientras esto fallaba**: contaba el diálogo como una región
      legítima, porque medía el HTML y el HTML es lo que está mal
- [ ] T032 [persona] **SC-004 — el control se encuentra, el concepto no se entiende (2026-08-29).**
      Señaló la casilla «Local only» a la primera y la activó y desactivó varias veces, así que
      FR-003 cumple su enunciado. Pero preguntó: «¿qué es? ¿qué hace? Entiendo lo que dice, pero no
      el concepto, qué relación tiene con el uso, qué pasa en cada opción». Las frases que lo
      explican —`mode.local_hint` y `mode.hybrid_hint`— existen y viven **solo en el tooltip de la
      insignia**, invisibles. Una decisión que no se entiende no es una decisión

> **T030–T032 no se marcan `[X]` porque el CI pase.** Necesitan a una persona delante. Si el ciclo
> se cierra sin ellas, se dice en el PR — no se dan por hechas.

---

## Dependencias y orden

### Entre fases

```
Fase 1 (línea de base)
   ↓
Fase 2 (separar la sección)  ← bloquea a las tres historias
   ↓
   ├── Fase 3 · US1 (P1) ─┐
   ├── Fase 4 · US2 (P2) ─┼→ Fase 6 (cierre)
   └── Fase 5 · US3 (P3) ─┘
```

### Entre historias

Las tres son **independientes entre sí** una vez hecha la Fase 2. Editan zonas distintas del mismo
archivo, así que en paralelo hay que resolver conflictos de texto, no de diseño.

### Dentro de cada historia

Test en rojo → implementación → inyección de la violación. En ese orden, y el tercer paso no es
opcional: es lo que distingue un test de un adorno.

### Oportunidades de paralelismo

| Fase | Tareas `[P]` | Por qué |
|---|---|---|
| 3 | T006 | `strings.rs`, mientras el resto toca HTML/CSS |
| 4 | T011 y T012 | Dos tests, un archivo, sin dependencia entre ellos |
| 4 | T018, T022 | Catálogo y documento de contrato, ajenos a `commands.rs` |
| 5 | T024 | `strings.rs` antes de cablear |

**Cadena crítica**: T013 → T014 → T015 → T016 → T017. Es una sola costura —la fuente única del
modo— y partirla deja el sistema con dos verdades a medias, que es peor que el estado actual.

---

## Estrategia

### MVP

**Fase 1 + Fase 2 + Fase 3 (US1)**. Entrega lo que el issue #52 pide en su titular: una superficie
legible donde se entiende qué va en cada campo.

### Entrega incremental

Cada fase deja el árbol verde y la aplicación usable. Se puede parar después de cualquiera.

**Pero la Fase 4 no es opcional en la práctica**: mientras no se haga, la insignia de modo miente y
`003`-FR-008a sigue roto. Es presentación en su enunciado y corrección en su efecto.

### Resumen

| Fase | Tareas | Historia |
|---|---|---|
| 1 · Preparación | T001 | — |
| 2 · Fundacional | T002–T003 | — |
| 3 · US1 (P1) | T004–T010 | Conectar sin adivinar |
| 4 · US2 (P2) | T011–T022 | El modo dice la verdad y se ve como decisión |
| 5 · US3 (P3) | T023–T026 | Nombres de región únicos |
| 6 · Cierre | T027–T032 | Verde, trazabilidad y los tres de persona |

**Total: 32 tareas.** 29 las cierra el build; **3 necesitan a una persona delante**.
