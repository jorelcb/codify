# Contrato — conectar una cuenta hasta el final

Lo que este ciclo se compromete a sostener, y **quién lo comprueba**.

---

## Comandos Tauri

Cambios sobre [`003`/contracts/skin-commands.md](../../003-conectividad-y-tiers/contracts/skin-commands.md), que hay que actualizar en el mismo ciclo.

| Comando | Antes | Después | Por qué |
|---|---|---|---|
| `connect_provider` | acumula un desafío en un mapa | **reemplaza** el que hubiera | Una cuenta no puede tener dos compitiendo, y hoy pulsar dos veces deja dos dentro |
| `complete_connection` | consume el desafío **antes** de intentar | lo consume **solo al terminar bien** | Una credencial rechazada obligaba a rehacer el formulario entero |
| `abandon_connection` | no existe | `-> ()` | FR-008: se puede decir «déjalo» |

El DTO del desafío cambia un campo: `instructions` deja de llevar una frase y lleva un **código**,
que la piel traduce. Greenfield: el campo no se duplica ni se deja por compatibilidad.

## Port del núcleo

```
Desafio::PideCredencial { instruccion: InstruccionDeCredencial }
```

`InstruccionDeCredencial` expone `code()` y `all()`, como `ProviderIssue`, `SessionFailure` y
`ConnectionState`. `DirectCredential::new` deja de recibir una `String` —hoy recibe una frase en
español— y recibe el tipo de dominio.

---

## Reglas que un test hace cumplir

Cinco nuevas. Las veintidós que ya existen siguen aplicando.

### R1 · Toda instrucción tiene texto en los dos idiomas — FR-010a, SC-003

> `toda_instruccion_de_credencial_tiene_texto_en_ambos_idiomas`

Recorre `InstruccionDeCredencial::all()` contra el catálogo. Un valor nuevo sin traducir no pasa.

**Inyección**: añadir un valor al enum sin darle texto.

### R2 · Ningún camino del núcleo queda sin puerta — FR-012, SC-007

> `ningun_camino_del_nucleo_queda_sin_puerta`

Para cada selector de rama conocido, recoge los **valores literales que la interfaz puede enviar**.
Una rama inalcanzable con cualquiera de ellos debe estar en la lista de caminos declarados, **con su
issue**.

**Inyecciones**: quitar la vía delegada de la lista de declarados; sacar de la lista algo que la
interfaz ya alcanza.

**Lo que no cubre**: caza *esta clase* —un selector que la piel fija— no todo código inalcanzable.
Dicho en [research D2](../research.md), no descubierto al revisar.

### R3 · La credencial no se queda en la pantalla — FR-011

> `la_credencial_no_sobrevive_al_envio`

El campo de la credencial se vacía al enviarse, y ningún módulo la guarda en una variable de
módulo. Es la misma regla que `004` aplicó al modo: sin copia no hay nada que se quede atrás.

**Inyección**: conservar el valor tras enviarlo.

### R4 · Un desafío rechazado sigue en pie — FR-006

> `una_credencial_rechazada_no_pierde_el_desafio`

Test del camino: desafío abierto → credencial rechazada → **sigue en curso** → credencial buena →
conectado. **Hoy este test no puede pasar** ([research D4](../research.md)).

**Inyección**: volver a consumir el desafío antes de intentar.

### R5 · No hay dos desafíos a la vez — FR-009, SC-006

> `pedir_conectar_dos_veces_deja_uno_solo`

Pedir conectar dos veces seguidas deja **uno**; cancelar deja **cero**.

**Inyección**: volver a un mapa en vez de un `Option` — y que el test lo note, no el tipo.

---

## Lo que no cambia, y sigue verificado

| Garantía | Quién |
|---|---|
| Ningún comando del backend sin invocar | `ningun_comando_del_backend_queda_sin_invocar` — **`complete_connection` sale de la lista de deuda en este ciclo** |
| Cero cadenas visibles fuera del catálogo | `ningun_texto_visible_escapa_al_catalogo` + R1 para las del backend |
| La superficie sigue siendo legible con la ventana al mínimo | `los_campos_del_formulario_caben_en_la_ventana_minima` |
| Nombres de región únicos, esquema de títulos | los tests de `004` |
| Cero-egress estructural | `compile_fail.rs`, `egress_guard.rs` |

Las seis claves nuevas entran en los dos idiomas o el test de catálogo falla.
