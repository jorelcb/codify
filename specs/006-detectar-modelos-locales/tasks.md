# Tasks — La aplicación ve lo que ya tienes instalado

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) · **Contrato**: [contracts/deteccion.md](./contracts/deteccion.md) · **Issue**: [#62](https://github.com/jorelcb/codify/issues/62)

> **Estado de alto nivel**: [issue #9 · Roadmap](https://github.com/jorelcb/codify/issues/9) · **Épica**: [#56](https://github.com/jorelcb/codify/issues/56)

## Antes de empezar

**El «por qué» de este ciclo es comprobable.** En la máquina donde se escribió el spec hay
`llama-server`, `llama-cli`, `claude` y `lms` en el `PATH`, y la aplicación enseña
«No hay backend disponible · `http://localhost:11434`» — una dirección hacia un Ollama que ahí no
está instalado.

**Dos piezas ya existen y no hay que construirlas**: `LocalProviderProbe`, que **rechaza por
construcción** cualquier dirección que no sea del propio equipo, y `ProviderIssue`, que ya sabe
decir «no escucha» y «sin modelos» con códigos estables.

**Dos reglas que este ciclo decide NO haciendo** ([research D3, D4](./research.md)):

| | |
|---|---|
| **No se ejecuta** ningún programa encontrado | Invocar un binario del `PATH` elegido por su nombre, para decidir si enseñar una fila, es más riesgo del que compra |
| **No se cablea como local** lo que no pueda sostenerlo | Un servidor de loopback lo sostiene por construcción. Un binario del `PATH` solo lo declara |

**Y una que conviene no malinterpretar**: que `claude` aparezca en la lista **es correcto**. El
producto quiere los dos caminos. Lo que sería un defecto es que apareciera como local.

**Formato**: `[ID] [P?] [Story] Descripción con ruta`.

---

## Fase 1: Preparación

- [ ] T001 Registrar la línea de base — `cargo test --workspace` (**251** verdes) y
      `cargo test -p codify-app --test ui_contract` (28 verdes), para que una caída posterior sea
      atribuible

---

## Fase 2: Fundacional (bloquea a las tres historias)

Los tipos primero, porque **dos requisitos se cumplen por construcción y no por vigilancia**: una
procedencia que no se puede omitir, y una clasificación que no se puede declarar a la ligera.

### Tests primero

- [ ] T002 Escribir `nada_marcado_como_local_infiere_fuera` en
      `crates/codify-core/tests/deteccion.rs` — un modelo solo se marca «infiere dentro» si viene
      de una fuente que lo sostiene; lo que solo lo declara se enseña y no se cablea. **Debe
      fallar**: los tipos no existen (R2, FR-010, FR-010a, SC-005)
- [ ] T003 [P] Escribir `toda_distancia_tiene_texto_en_ambos_idiomas` en
      `crates/codify-app/tests/ui_contract.rs`, recorriendo `Distancia::all()` contra el catálogo.
      **Debe fallar** (R4, FR-012)

### Implementación

- [ ] T004 Crear `ModeloDisponible`, `Procedencia`, `Distancia` y `DondeInfiere` en
      `crates/codify-core/src/application/deteccion.rs`. **La procedencia no puede omitirse**: lo
      sostiene el tipo, no un test (FR-005). `Distancia` y `DondeInfiere` con `code()` y `all()`,
      como `ProviderIssue`
- [ ] T005 Definir el port `DetectorDeModelos` en `crates/codify-core/src/application/deteccion.rs`.
      Vive en `application/` porque
      **el dominio no nombra «detectar»**: es una capacidad, y la constitución separa esos dos
      casos (FR-001)
- [ ] T006 [P] Añadir a `crates/codify-app/src/strings.rs` las claves de las distancias y de dónde
      infiere, en los dos idiomas (FR-012)

**Punto de control**: 251 tests siguen verdes, más los dos nuevos. Nada de la interfaz ha cambiado.

---

## Fase 3: User Story 1 — Abrir y que ya sepa qué tengo (P1) 🎯 MVP

**Meta**: que quien tiene un modelo instalado lo vea ofrecido sin configurar nada.

**Prueba independiente**: abrir la aplicación en una máquina con algo instalado y mirar.

**Depende de**: Fase 2.

### Tests primero

- [ ] T007 [US1] Escribir `la_deteccion_no_ejecuta_nada` en
      `crates/codify-app/tests/ui_contract.rs` — ningún módulo de `infrastructure/deteccion/`
      construye un proceso hijo. **Debe fallar** mientras el módulo no exista, y **debe seguir
      cayendo** si alguien añade la invocación (R1, FR-011)
- [ ] T008 [US1] Escribir en `crates/codify-core/tests/deteccion.rs` que las dos fuentes producen
      la **distancia** que les toca —servidor que responde → usable; binario en el `PATH` →
      instalado— con fuentes de mentira. **Debe fallar** (FR-002)
- [ ] T009 [US1] Escribir `todo_motivo_de_indisponibilidad_tiene_texto_y_salida` en
      `crates/codify-app/tests/ui_contract.rs` — cada distancia que no sea «usable» tiene texto **y
      salida**: qué falta para poder usarlo. **Debe fallar** (R4, FR-009)
- [ ] T010 [US1] Escribir `la_deteccion_arranca_sola` en
      `crates/codify-app/tests/ui_contract.rs` — el arranque invoca la detección sin que nadie la
      pida, y lo detectado se enseña como **varias opciones**, no como un valor único. **Debe
      fallar** (R7, FR-001, FR-004, FR-010b)
- [ ] T011 [US1] Escribir `el_path_se_recorre_entero` en `crates/codify-core/tests/deteccion.rs` —
      con un `PATH` de mentira que incluya un directorio no estándar, lo que hay dentro se
      encuentra. **Debe fallar** (R8, FR-003)

### Implementación

- [ ] T012 [US1] Crear la fuente de servidores en
      `crates/codify-core/src/infrastructure/deteccion/servidores.rs`, sobre `LocalProviderProbe` y
      una lista **declarada** de puertos de loopback. Es la única que sostiene la clasificación por
      construcción. Un servidor que responde sin modelos ya lo dice `ProviderIssue::NoModels`, así
      que ese caso límite queda cubierto por lo que hay ([research D2](./research.md))
- [ ] T013 [US1] Crear la fuente de programas en
      `crates/codify-core/src/infrastructure/deteccion/programas.rs`: recorre el `PATH` buscando
      nombres declarados y reporta **presencia y ruta**. **No ejecuta nada** — ni para saber la
      versión (FR-003, FR-011, [research D3](./research.md))
- [ ] T014 [US1] Añadir el comando de detección en `crates/codify-app/src/commands.rs` y
      registrarlo en `crates/codify-app/src/lib.rs` (FR-001)
- [ ] T015 [US1] Enseñar lo detectado al arrancar en `crates/codify-app/ui/index.html` y
      `crates/codify-app/ui/provider.js`, **para elegir**, no imponiendo uno. Encontrar algo no lo
      da por aceptado: se enseña y se usa si el usuario lo elige (FR-004, FR-010b, FR-013)
- [ ] T016 [US1] Decir **qué falta** para lo que no se puede usar todavía, en
      `crates/codify-app/ui/provider.js` y `crates/codify-app/src/strings.rs`: instalado pero sin
      arrancar, y nada instalado (FR-009)
- [ ] T017 [US1] Verificar T007 y T008 por inyección: invocar un binario encontrado en
      `crates/codify-core/src/infrastructure/deteccion/programas.rs`; y devolver la distancia
      equivocada. Revertir cada una

**Punto de control**: abrir la aplicación en esta máquina debería enseñar `llama-server`,
`llama-cli`, `claude` y `lms`, hoy invisibles.

---

## Fase 4: User Story 2 — Saber de dónde salió cada cosa (P2)

**Meta**: que nada aparezca en la lista sin explicar por qué está ahí.

**Prueba independiente**: mirar la lista y preguntar, de cada fila, de dónde salió.

**Depende de**: Fase 2.

### Test primero

- [ ] T018 [US2] Escribir `todo_modelo_disponible_declara_su_procedencia` en
      `crates/codify-app/tests/ui_contract.rs` — la piel **enseña** la procedencia. Que no se pueda
      construir sin ella lo sostiene el tipo; este test cubre lo que el tipo no alcanza. **Debe
      fallar** (R3, FR-005, FR-006, SC-003)

### Implementación

- [ ] T019 [US2] Enseñar la procedencia de cada modelo en `crates/codify-app/ui/provider.js`, y
      distinguir lo detectado de lo que puso el usuario (FR-005, FR-006)
- [ ] T020 [P] [US2] Añadir las claves de procedencia a `crates/codify-app/src/strings.rs`, en los
      dos idiomas (FR-012)
- [ ] T021 [US2] Verificar T018 por inyección: enseñar la lista sin la procedencia, en
      `crates/codify-app/ui/provider.js`. Revertir

**Punto de control**: ninguna fila de la lista aparece sin decir de dónde salió. Una lista mágica
contradice el principio rector del proyecto.

---

## Fase 5: User Story 3 — Que detectar no cueste la apertura (P3)

**Meta**: que buscar no retrase el momento en que la aplicación sirve.

**Prueba independiente**: con una fuente que nunca contesta, abrir y cronometrar.

**Depende de**: Fase 2.

### Test primero

- [ ] T022 [US3] Escribir `la_deteccion_respeta_su_presupuesto` en
      `crates/codify-core/tests/deteccion.rs` — con una fuente que nunca contesta, la detección
      termina dentro de los **3 s** y lo declara. **Debe fallar** (R6, FR-007, FR-008, SC-004)

### Implementación

- [ ] T023 [US3] Correr las fuentes **a la vez** con un presupuesto común de 3 s en
      `crates/codify-core/src/application/deteccion.rs`, abandonando y declarándolo lo que no
      llegue ([research D5](./research.md))
- [ ] T024 [US3] Pintar la interfaz **antes** de que la detección termine, diciendo que está
      mirando, en `crates/codify-app/ui/provider.js` (FR-007)
- [ ] T025 [US3] Verificar T022 por inyección: esperar a la fuente lenta sin presupuesto, en
      `crates/codify-core/src/application/deteccion.rs`. Revertir

**Punto de control**: la aplicación responde desde el arranque aunque no haya terminado de mirar.

---

## Fase 6: Retirar la dirección inventada

No cuelga de ninguna historia: es lo que deja de hacerse. Enseñar una dirección hacia un servicio
que puede no estar instalado es **afirmar algo sin comprobarlo**, en un producto cuya tesis es que
eso no se hace.

- [ ] T026 Escribir `ninguna_direccion_por_defecto_sobrevive` en
      `crates/codify-app/tests/ui_contract.rs` — no existe ninguna dirección de proveedor escrita
      como valor por defecto. **Debe fallar**: hoy está `DEFAULT_ENDPOINT` (R5, FR-014, SC-002)
- [ ] T027 Escribir `la_interfaz_no_pide_una_direccion` en
      `crates/codify-app/tests/ui_contract.rs` — ningún campo pide una dirección de proveedor
      local, y `CODIFY_LOCAL_ENDPOINT` no aparece ni en `ui/` ni en el catálogo. **Debe fallar**
      (R9, FR-013, FR-015)
- [ ] T028 Retirar `DEFAULT_ENDPOINT` de `crates/codify-app/src/commands.rs` y hacer que
      `probe_provider` sondee **lo detectado** en vez de una dirección inventada (FR-014)
- [ ] T029 Dejar `CODIFY_LOCAL_ENDPOINT` como **anulación de desarrollo**, declarada en
      `crates/codify-app/src/commands.rs` y pasando por el mismo sondeo que todo lo demás, sin
      trato de favor. Fuera de la interfaz y de todo lo que el usuario tenga que leer (FR-015)
- [ ] T030 Verificar T026 por inyección: devolver una constante con una dirección en
      `crates/codify-app/src/commands.rs`. Revertir

---

## Fase 7: Cierre, y lo que no cierra el build

- [ ] T031 Dejar el árbol en verde: `cargo test --workspace`, `cargo clippy --workspace
      --all-targets -- -D warnings`, `cargo fmt --all -- --check`. Se esperan **261** tests
- [ ] T032 [P] Actualizar `specs/003-conectividad-y-tiers/contracts/skin-commands.md` con el
      comando de detección y la desaparición de la dirección por defecto — el contrato que cambia
      es el del crate que se toca
- [ ] T033 [persona] **SC-001**: abrir la aplicación en una máquina con algo instalado y ver que
      aparece, **sin escribir ni configurar nada** — cero pasos de configuración (SC-007). En esta
      máquina deberían salir cuatro cosas
- [ ] T034 [persona] **SC-002**: con nada escuchando, **ninguna dirección** aparece en pantalla. Se
      dice qué podría instalarse
- [ ] T035 [persona] **SC-006**: generar contexto de un repositorio **sin pasar por la superficie
      de conexión**. Es el criterio que mide si este ciclo sirve: las dos validaciones anteriores
      fallaron intentando conectar algo remoto, y **nadie llegó a probar el camino local**
- [ ] T036 [persona] Comprobar que si aparece `claude` se enseña como lo que es —programa de la
      máquina, **inferencia fuera**—. Que esté en la lista es correcto; que figurara como local no

> **T033–T036 no se marcan `[X]` porque el CI pase.** Y hay un motivo extra en este ciclo: los
> tests usan fuentes de mentira, así que **nadie ha comprobado que el descubrimiento real
> funcione** hasta que alguien abra esto en su máquina. Está declarado en el plan como renuncia.

---

## Dependencias y orden

```
Fase 1 (línea de base)
   ↓
Fase 2 (los tipos y el port)  ← bloquea a las tres historias
   ↓
   ├── Fase 3 · US1 (P1) ─┐
   ├── Fase 4 · US2 (P2) ─┼→ Fase 6 (retirar la dirección) → Fase 7 (cierre)
   └── Fase 5 · US3 (P3) ─┘
```

### Entre historias

Independientes en lo conceptual. US2 y US3 tocan `provider.js`, que US1 reescribe: en paralelo hay
conflictos de texto, no de diseño.

**La Fase 6 va después de US1** porque retirar la dirección por defecto deja la aplicación sin
forma de encontrar un modelo hasta que la detección exista.

### Dentro de cada historia

Test en rojo → implementación → **inyección de la violación**. En este ciclo dos inyecciones
importan más que el resto: la de T007 —que no se ejecute nada encontrado— y la de T002 —que nada
remoto se cuele como local—. Son las que protegen algo que no se recupera.

> **Por qué este hallazgo se repite.** Tres ciclos seguidos con requisitos implementados y sin
> test. La causa no son las instancias: escribo las tareas de implementación recorriendo los
> requisitos, y las de test recorriendo el contrato — de modo que **lo que el contrato no nombra se
> queda sin red**. La corrección es escribir el contrato recorriendo los requisitos, no al revés.

### Oportunidades de paralelismo

| Fase | `[P]` | Por qué |
|---|---|---|
| 2 | T003 | Test de catálogo, ajeno a los tipos |
| 2 | T006 | `strings.rs` |
| 4 | T020 | `strings.rs` |
| 7 | T032 | Documento de contrato |

**Cadena crítica**: T004 → T005 → T012/T013. Los tipos, el port y las fuentes: partirlo deja el
árbol sin compilar entre medias.

---

## Estrategia

### MVP

**Fases 1 + 2 + 3**. Entrega lo que el issue #62 pide: que la aplicación vea lo que ya tienes.

### Entrega incremental

Cada fase deja el árbol verde. **Pero la Fase 6 no es opcional**: mientras siga la dirección
inventada, la aplicación sigue afirmando algo que no comprobó — y eso es lo que el producto dice no
hacer.

### Resumen

| Fase | Tareas | Historia |
|---|---|---|
| 1 · Preparación | T001 | — |
| 2 · Fundacional | T002–T006 | — |
| 3 · US1 (P1) | T007–T017 | Abrir y que ya sepa qué tengo |
| 4 · US2 (P2) | T018–T021 | Saber de dónde salió cada cosa |
| 5 · US3 (P3) | T022–T025 | Que detectar no cueste la apertura |
| 6 · La dirección inventada | T026–T030 | — |
| 7 · Cierre | T031–T036 | — |

**Total: 36 tareas.** 32 las cierra el build; **4 necesitan a una persona con su máquina delante**.
