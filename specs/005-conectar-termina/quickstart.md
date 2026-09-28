# Quickstart — validar que conectar termina

Dos partes. La separación importa: este spec existe porque 239 tests en verde convivieron con un
camino que no llegaba a ninguna parte.

## Prerrequisitos

- Un sistema **con almacén de credenciales** para el camino feliz. Sin él se ejercita el escenario
  de FR-004, que también hay que probar.
- Una credencial cualquiera: no hace falta que sea válida en un proveedor real para los escenarios
  de esta lista.

---

## Parte 1 — Lo que cierra el build

```bash
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo fmt --all -- --check
```

**Se espera**: **251** tests (239 de antes + 12 nuevos), workspace entero en verde.

| Criterio | Test |
|---|---|
| **SC-002** | `ningun_comando_del_backend_queda_sin_invocar`, con `complete_connection` **fuera** de la lista de deuda |
| **SC-003** | `ningun_texto_visible_escapa_al_catalogo` + `toda_instruccion_de_credencial_tiene_texto_en_ambos_idiomas` |
| **SC-004** | los tests de catálogo que recorren los motivos de fallo |
| **SC-005** (su mitad automatizable) | `el_envio_no_admite_un_segundo_intento` |
| **SC-006** | `pedir_conectar_dos_veces_deja_uno_solo` |
| **SC-007** | `ningun_camino_del_nucleo_queda_sin_puerta` |
| **FR-006** | `una_credencial_rechazada_no_pierde_el_desafio` |
| **FR-011** | `la_credencial_no_sobrevive_al_envio` |

### Verificar que los tests sirven

Inyectar cada violación de [contracts/conexion.md](./contracts/conexion.md), una por una, y
comprobar que **cae el test que le toca** — no cualquiera. Revertir cada una.

Este ciclo tiene un motivo extra para no saltárselo: tres de sus defectos sobrevivieron a suites
verdes, y uno de los tests de `004` pasó en verde antes de servir para nada por un fallo de análisis
propio.

---

## Parte 2 — Lo que necesita a una persona

```bash
cargo run -p codify-app
```

### SC-001 · Conectar de punta a punta, sin ayuda

**Es el criterio que `004` dejó declarado en fallo**, y no se hereda como aprobado: se vuelve a
medir ahora que el camino existe.

Dar a alguien que no conoce la aplicación una credencial y pedirle que conecte una cuenta. Sin
explicar nada.

- **Pasa**: lo consigue, y la cuenta aparece en la lista.
- **Falla**: pregunta, o se queda sin saber si pasó algo. **Anotar sus palabras** — un «falla» sin
  la frase concreta no dice qué arreglar.

### SC-005 · Nunca sin señal

Enviar la credencial y observar.

- **Pasa**: en todo momento se sabe que algo está ocurriendo, y no se puede enviar dos veces.
- **Falla**: hay un instante en que la pantalla no dice nada — indistinguible de «pulsé y no pasó
  nada», que es el defecto que originó este spec.

### Los fallos, uno a uno

| Escenario | Cómo provocarlo | Qué debe verse |
|---|---|---|
| Credencial rechazada | una clave inválida | se dice, y **el desafío sigue en pie**: se corrige sin rehacer nada |
| Sin almacén | un sistema sin llavero | se dice, y se ofrece seguir en local |
| Desafío caducado | cancelar y luego enviar | se distingue de una credencial equivocada |
| Cancelar | el control de cancelar | vuelve al estado limpio, sin nada retenido |

Los cuatro en **los dos idiomas**, cambiando el idioma de la interfaz sin reiniciar.

### Custodia — cómo NO se comprueba

Que la credencial quedó bajo el almacén del sistema **no se verifica leyendo el llavero del
operador**. Lanzar las herramientas del sistema contra él disparó una alerta de seguridad
corporativa y un diálogo pidiendo la contraseña (`003`-T031, 2026-08-27) — y ese diálogo *era* la
evidencia: un proceso ajeno no puede leer la entrada.

Lo que sí se comprueba: que la credencial **no está** donde no debe —ni en el proyecto, ni en la
configuración, ni en ningún registro o evento— y que la cuenta sigue funcionando en la sesión
siguiente, que es lo que demuestra que se guardó.
