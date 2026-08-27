# Contrato — superficie de conexión y modo

Lo que este ciclo se compromete a sostener, y **quién lo comprueba**. Un contrato que nadie
verifica es una intención.

---

## Comandos Tauri

Cambios sobre lo que fija
[`003`/contracts/tauri-commands.md](../../003-conectividad-y-tiers/contracts/tauri-commands.md),
que hay que actualizar en el mismo ciclo.

| Comando | Antes | Después | Por qué |
|---|---|---|---|
| `set_mode(local: bool)` | `-> ()` | `-> ModeDto` | Quien pide el cambio recibe el estado resultante, en vez de asumirlo |
| `mode()` | no existe | `-> ModeDto` | La interfaz necesita leer el modo al arrancar, y no tener copia propia |
| `start_session(request)` | `request.local` decide el modo | lee `AppState.mode`; **`request.local` se elimina** | Un tercer dueño del modo, y el que rompía `003`-FR-008a (research D2) |

```
ModeDto := { local: bool }
```

**Greenfield**: `request.local` se borra. No queda como campo opcional ni ignorado — un campo que
se acepta y no se usa es la siguiente confusión.

---

## Reglas que un test hace cumplir

Cada regla nombra el test que la sostiene. Los cuatro son nuevos; los catorce que ya existen
siguen aplicando sin cambios.

### R1 · Ninguna región comparte nombre — FR-005, FR-006, SC-002

> `ningun_par_de_regiones_comparte_nombre_accesible`

Recoge las regiones de `index.html`, resuelve su `data-i18n-aria` contra `strings.rs`, y falla si
se repite **una clave o un texto**, en cualquiera de los dos idiomas.

**Inyección que debe hacerlo fallar**: darle a `#applied` el nombre de `#decide`.

### R2 · El modo tiene un solo escritor por superficie, y es el mismo — FR-003a, SC-006

> `el_modo_no_puede_discrepar_entre_sus_dos_superficies`

Cuenta los sitios que escriben cada superficie: `dataset.mode =` y el `checked =` de
`#modo-local`. Exige **uno cada uno**, y que ambos estén en la misma función.

**Inyección que debe hacerlo fallar**: pintar la insignia desde `connections.js`.

**Lo que este test no ve**: que la función se llame. Eso lo cubre el test que ya existe,
`ninguna_funcion_de_la_interfaz_queda_sin_llamar`.

### R3 · La interfaz no guarda copia del modo — FR-003a

> `la_interfaz_no_tiene_su_propia_idea_del_modo`

Ningún módulo declara una variable de modo propia; el valor viene siempre de `mode()` o de lo que
devuelve `set_mode`. Sin esto, R2 pasa con dos superficies pintadas a la vez desde un valor
equivocado — que es literalmente lo que ocurre hoy.

**Inyección que debe hacerlo fallar**: reintroducir `local: true` en el estado de `main.js`.

### R4 · Los campos caben — SC-005

> `los_campos_del_formulario_caben_en_la_ventana_minima`

Cada campo de texto del formulario declara `min-width` en `ch` de al menos 24, y su contenedor
envuelve. Un mínimo dentro de un contenedor que no envuelve no da campos legibles: da
desbordamiento.

**Inyección que debe hacerlo fallar**: bajar un campo a `20ch`, o quitar el `flex-wrap`.

---

## Lo que no cambia, y sigue verificado

| Garantía | Test que ya existe |
|---|---|
| Cero cadenas fuera del catálogo (FR-007) | `ningun_texto_visible_escapa_al_catalogo` |
| Ninguna clave huérfana ni sin cablear | `toda_clave_del_catalogo_esta_cableada_o_declarada_como_reservada` |
| Nadie tiene dos dueños de su texto | `ningun_elemento_tiene_dos_duenos_de_su_texto` |
| Quien pinta a mano repinta al cambiar de idioma | `quien_pinta_a_mano_repinta_al_cambiar_de_idioma` |
| El punto de quiebre responsivo es alcanzable | `el_punto_de_quiebre_responsivo_es_alcanzable` |
| Cero-egress estructural | `compile_fail.rs`, `egress_guard.rs` |

Las cuatro claves nuevas —`a11y.connections_region`, `a11y.prefs_region`, `connection.reveal`,
`a11y.mode_group`— entran en los dos idiomas o el segundo test de esta tabla falla.
