# Contrato — detectar lo que ya está instalado

Lo que este ciclo se compromete a sostener, y **quién lo comprueba**.

---

## El port

```
DetectorDeModelos
  └── detectar(presupuesto) -> [ModeloDisponible]
```

Vive en `application/`: el dominio no nombra «detectar», es una capacidad. Lo que cruza son tipos
propios —nombre, procedencia, distancia, dónde infiere— **nunca rutas ni cadenas sueltas**.

Cada fuente es un adapter. Añadir una es añadir un adapter y **declararla en el catálogo**.

## Comandos Tauri

| Comando | Antes | Después |
|---|---|---|
| detectar modelos | no existe | devuelve lo encontrado, con procedencia y distancia |
| `probe_provider` | sondea `CODIFY_LOCAL_ENDPOINT` o una dirección inventada | sondea **lo detectado** |

**`DEFAULT_ENDPOINT` desaparece.** Greenfield: no se deja por compatibilidad.

---

## Reglas que un test hace cumplir

### R1 · La detección no ejecuta nada de lo que encuentra — FR-011

> `la_deteccion_no_ejecuta_nada`

Ningún módulo de detección lanza un proceso. Comprobable leyendo el código: no aparece la
construcción de un proceso hijo en `infrastructure/deteccion/`.

**Inyección**: invocar un binario encontrado para preguntarle su versión.

**Por qué es una regla y no una costumbre.** Lanzar un binario del `PATH` elegido por su nombre es
ejecutar código arbitrario para decidir si enseñar una fila en una lista. El riesgo no lo
controlamos y la pregunta no lo necesita ([research D3](../research.md)).

### R2 · Nada clasificado como local infiere fuera del equipo — SC-005

> `nada_marcado_como_local_infiere_fuera`

Un modelo solo se marca «infiere dentro» si viene de una fuente que **sostiene** esa propiedad por
construcción. Lo que la declara sin sostenerla se enseña, **no se cablea como local**.

**Inyecciones**: marcar como local algo que llega de la fuente de programas; hacer que el sondeo
acepte una dirección que no sea del propio equipo.

### R3 · Todo lo ofrecido declara su procedencia — FR-005, SC-003

> `todo_modelo_disponible_declara_su_procedencia`

No se puede construir un `ModeloDisponible` sin procedencia — y eso lo sostiene el tipo, no el
test. El test cubre lo que el tipo no alcanza: que la piel la **enseñe**.

**Inyección**: enseñar la lista sin la procedencia.

### R4 · Toda distancia y todo motivo tienen texto en los dos idiomas — FR-012

> `toda_distancia_tiene_texto_en_ambos_idiomas`
> `todo_motivo_de_indisponibilidad_tiene_texto_y_salida`

Recorren `all()` contra el catálogo, como `ProviderIssue`, `SessionFailure` y `ConnectionState`. Un
valor nuevo sin traducir no pasa.

**Inyección**: añadir un valor sin texto.

### R5 · No queda ninguna dirección inventada — FR-014, SC-002

> `ninguna_direccion_por_defecto_sobrevive`

No existe una dirección de proveedor escrita en el código como valor por defecto. Hoy hay una
—`http://localhost:11434`— que apunta a un servicio que puede no estar instalado.

**Inyección**: devolver una constante con una dirección.

### R6 · La detección tiene un límite y no bloquea — FR-007, FR-008, SC-004

> `la_deteccion_respeta_su_presupuesto`

Con una fuente que nunca contesta, la detección termina dentro del presupuesto y lo declara.

**Inyección**: una fuente que duerme más que el presupuesto y un detector que la espera.

### R7 · La detección arranca sola y ofrece opciones — FR-001, FR-004, FR-010b

> `la_deteccion_arranca_sola`

El arranque invoca la detección sin que nadie la pida, y lo detectado se enseña como **varias
opciones**, no como un valor único que se impone.

**Inyección**: quitar la invocación del arranque; enseñar un solo modelo en vez de una lista.

### R8 · El `PATH` se recorre entero — FR-003

> `el_path_se_recorre_entero`

Con un `PATH` de mentira que incluya un directorio no estándar, lo que hay dentro se encuentra.
Compilar desde fuente es normal en este terreno: en la máquina donde se escribió este spec,
`llama-server` vive fuera de toda ruta habitual.

**Inyección**: recorrer solo un conjunto fijo de directorios conocidos.

### R9 · La interfaz no pide una dirección — FR-013, FR-015

> `la_interfaz_no_pide_una_direccion`

Ningún campo pide una dirección de proveedor local, y `CODIFY_LOCAL_ENDPOINT` **no aparece** ni en
`ui/` ni en el catálogo: es una anulación de desarrollo, no algo que el usuario deba leer.

**Inyección**: devolver un campo de dirección a la interfaz; mencionar la variable en el catálogo.

---

## Lo que no cambia, y sigue verificado

| Garantía | Quién |
|---|---|
| Cero cadenas visibles fuera del catálogo | `ningun_texto_visible_escapa_al_catalogo` |
| Ningún comando del backend sin invocar | `ningun_comando_del_backend_queda_sin_invocar` |
| Ningún camino del núcleo sin puerta | `ningun_camino_del_nucleo_queda_sin_puerta` |
| Ningún test escribe en el llavero real | `ningun_test_escribe_en_el_llavero_real` |
| Cero-egress estructural **en modo local** | `compile_fail.rs`, `egress_guard.rs` |
