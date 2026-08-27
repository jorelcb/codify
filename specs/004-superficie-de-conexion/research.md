# Phase 0 — Investigación

Seis decisiones. La segunda es la que cambió el plan, y no venía en el spec: apareció buscando
cómo se comprueba que dos elementos no discrepen.

---

## D1 · Cómo se comprueba que ningún nombre de región se repite

**Decisión.** Un test estático sobre `index.html`: recoger todo elemento que sea región
—`role="region"` o etiqueta de landmark (`header`, `main`, `footer`, `nav`, `aside`)— leer su
`data-i18n-aria`, y exigir que **ni las claves ni los textos resueltos** se repitan, en los dos
idiomas.

**Por qué las dos cosas.** Comprobar solo las claves deja pasar dos claves distintas con el mismo
texto; comprobar solo los textos deja pasar el caso en que un idioma los distingue y el otro no.
El lector de pantalla oye el texto, no la clave.

**Lo que encuentra hoy** — los dos pares que FR-005 anuncia, y son exactamente estos:

| Clave | La llevan | Por qué duele |
|---|---|---|
| `a11y.provider_region` | `#provider` y `#conexiones` | Quien navega por regiones oye lo mismo en la barra de estado del backend y en la superficie de conexión |
| `a11y.toolbar_region` | `header.bar` y `footer.prefs` | La cabecera y el pie dicen ser la misma cosa |

**Alternativas descartadas.**

- *Auditoría con `axe-core` en un navegador headless.* Cubre mucho más que la unicidad, pero
  exige navegador en CI y una dependencia de JavaScript en un proyecto que hoy no tiene ni
  bundler. El coste no lo paga un requisito que el análisis estático cubre entero.
  **Señal para revisitarlo:** que los nombres de región pasen a componerse en tiempo de ejecución,
  o que aparezcan más requisitos de accesibilidad que el texto plano no alcance.
- *Revisarlo a mano en la validación manual.* Es justo lo que ya falló: los nombres llevan
  duplicados desde `003` y nadie lo notó mirando. FR-006 existe por eso.

---

## D2 · La única fuente del modo — y el tercer dueño que nadie había contado

**El spec suponía dos dueños. Son tres, y el tercero rompe una garantía vigente.**

Lo que hay hoy, verificado leyendo el código:

| Dueño | Dónde | Qué cree que es el modo |
|---|---|---|
| La insignia de la barra | `main.js:79` — `el.mode.dataset.mode = ui.local ? …` | `ui.local`, que se inicializa a `true` en `main.js:46` **y no se reasigna nunca** |
| La casilla del panel | `connections.js:99` — lee `#modo-local.checked` | Lo que el usuario haya marcado en el DOM |
| **El backend** | `commands.rs:531` — `set_mode` escribe `state.mode` | Lo que `set_mode` guardó… y **nadie lee** |

El tercero es el que hace daño. `start_session` (`commands.rs:667`) decide el modo de la sesión
así:

```rust
let mode = if request.local { Mode::Local } else { Mode::Hybrid };
```

y `request.local` viene de `ui.local` (`main.js:219`), que vale `true` siempre. Consecuencias:

1. **La insignia miente por construcción**: dice «local» hagas lo que hagas.
2. **`set_mode` escribe en un estado que ningún lector consulta.**
3. **`003`-FR-008a no se cumple de punta a punta** — «el modo se guarda y se aplica a la siguiente
   sesión»: se guarda, y no se aplica. La sesión siempre se arma en local.

Falla del lado seguro —nunca sale nada— pero eso no lo convierte en una garantía. Es un accidente:
el usuario no puede elegir, que es un defecto distinto de estar protegido.

**Decisión.** La única fuente del modo es **`state.mode`, en el backend**. La interfaz no guarda
ninguna copia:

- `start_session` lee `state.mode` y `request.local` **desaparece** de la petición.
- `set_mode` devuelve el modo resultante; se añade una consulta `mode()` para el arranque.
- **Una sola función** pinta el modo, y pinta **las dos** superficies a la vez.
- Un test exige que `dataset.mode =` y `#modo-local`.`checked =` aparezcan **cada uno en un solo
  sitio**, y que ese sitio sea el mismo. Dos escritores es el defecto; un test que los cuente lo
  impide.

**Alternativas descartadas.**

- *Dejar dos renderizadores independientes y comparar sus salidas en un test de navegador.*
  Detectar la divergencia en vez de imposibilitarla, y con navegador en CI. Este proyecto pone las
  garantías en la estructura precisamente para no tener que vigilarlas.
- *Que la interfaz mantenga `ui.local` y lo sincronice al cambiar.* Es lo que ya hay, sin
  sincronizar. Añadir la sincronización deja el segundo dueño en pie y confía en que nadie olvide
  llamarla — que es exactamente lo que pasó.
- *Quitar la insignia y dejar solo la casilla.* Resolvería la divergencia eliminando una
  superficie, pero la sesión de clarificación decidió lo contrario, y con motivo: informar de un
  vistazo y poder cambiarlo son preguntas distintas.

**Lo que este spec consume y no arregla del todo.** Que `003`-FR-008a estuviera roto es un
defecto de `003`. Se repara aquí porque FR-003a es irrealizable sin repararlo, pero merece **issue
propio**: el arreglo entra en este ciclo, la trazabilidad no puede quedarse en un comentario.

---

## D3 · Cómo se comprueba SC-005 sin medir píxeles

**Decisión.** Expresar el criterio en la unidad que *significa* el criterio: `min-width: 24ch` en
los campos de texto del formulario. `1ch` es el ancho del carácter «0» de la fuente en uso, así
que `24ch` **es** «24 caracteres», no una aproximación en píxeles que depende de la tipografía.

El test comprueba dos cosas: que cada campo declare `min-width` en `ch` de al menos 24, y que su
contenedor **envuelva** (`flex-wrap` o rejilla con `auto-fit`), porque un mínimo dentro de un
contenedor que no envuelve no produce campos legibles: produce desbordamiento horizontal, que es
la otra mitad de SC-005.

**Alternativas descartadas.**

- *Fijar un ancho en píxeles.* Es lo que hay hoy —campos de 147 px— y no dice nada sobre cuántos
  caracteres caben. Cambiar la fuente cambiaría el cumplimiento sin cambiar el CSS.
- *Medir el render de verdad.* Es lo único que mediría el criterio literalmente, y necesita
  navegador. **Señal para revisitarlo:** que aparezcan más criterios que dependan del render real
  y justifiquen el arnés entero.

---

## D4 · Presentación propia para la superficie (FR-001)

**Decisión.** `#conexiones` deja de llevar `class="provider"`. Presentación en bloque, con
jerarquía visible: el modo arriba y separado, la lista de cuentas después, el formulario plegado
al final.

**Por qué no un modificador.** `.provider .provider--conexiones` dejaría las reglas de la barra de
estado como base y las contradicciones encima. FR-001 prohíbe justamente eso: reutilizar la
disposición de una barra de estado para contener un formulario. Un modificador es esa
reutilización con otro nombre.

---

## D5 · Cómo se revela el formulario (FR-002a)

**Decisión.** `<details>` con `<summary>`, sin JavaScript.

**Por qué.** El estado abierto/cerrado tiene **exactamente un dueño**: el atributo `open` del DOM.
No hay variable que sincronizar, ni segundo escritor que se olvide de actualizar — que es la
enfermedad que este mismo spec está tratando en D2. Además trae gratis la semántica de expansión
para el lector de pantalla y el manejo de teclado.

**Precio, y por qué se paga.** El marcador del `<summary>` hay que estilarlo a mano en cada
navegador. Es CSS aburrido a cambio de que un estado no pueda divergir; el proyecto ya ha pagado
tres veces el precio contrario.

**Alternativa descartada.** Un `<button>` que conmuta `hidden` sobre un `<fieldset>`: funciona, y
crea un segundo dueño del mismo hecho.

---

## D6 · Las cuatro claves de catálogo nuevas

`a11y.connections_region` y `a11y.prefs_region` deshacen los dos pares repetidos de D1;
`a11y.mode_group` nombra el grupo del modo.

**El `<summary>` no estrena clave: reutiliza `connection.add`** («Conectar un proveedor»), que es
literalmente el texto que la clarificación puso en la puerta de entrada. Lo que cambia es el
**botón de envío**, que hoy lleva esa misma clave y pasa a `connection.submit` —«Conectar» /
«Connect»—. Dejar las dos cosas con el mismo texto habría dado dos controles idénticos, uno que
abre y otro que envía: peor que la pantalla que este spec viene a arreglar.

En los dos idiomas, y el test que ya existe
—`toda_clave_del_catalogo_esta_cableada_o_declarada_como_reservada`— se encarga de que ninguna
quede huérfana ni sin cablear. No hace falta nada nuevo para eso: FR-007 está cubierto por lo que
ya hay, y la tarea es no romperlo.
