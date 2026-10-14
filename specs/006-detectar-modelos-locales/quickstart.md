# Quickstart — validar que la aplicación ve lo que tienes

Dos partes, y la segunda pesa más que de costumbre: **la detección depende de la máquina**, así que
el build puede verificar la lógica y no el descubrimiento.

## Prerrequisitos

Nada que instalar para la parte 1. Para la parte 2, una máquina con **algo** instalado — y vale la
pena probarla también en una donde **no haya nada**, que es el caso de quien abre esto por primera
vez.

---

## Parte 1 — Lo que cierra el build

```bash
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo fmt --all -- --check
```

| Criterio | Test |
|---|---|
| **FR-011** | `la_deteccion_no_ejecuta_nada` |
| **SC-005** | `nada_marcado_como_local_infiere_fuera` |
| **SC-003** | `todo_modelo_disponible_declara_su_procedencia` |
| **FR-012** | los que recorren `all()` contra el catálogo |
| **SC-002** | `ninguna_direccion_por_defecto_sobrevive` |
| **SC-004** | `la_deteccion_respeta_su_presupuesto` |

### Verificar que los tests sirven

Inyectar cada violación de [contracts/deteccion.md](./contracts/deteccion.md), una por una, y
comprobar que **cae el que le toca** — no cualquiera. Revertir cada una.

Dos merecen atención especial, porque son las que protegen algo que no se recupera: que no se
ejecute un binario encontrado, y que nada remoto se cuele como local.

---

## Parte 2 — Lo que necesita una máquina de verdad

```bash
cargo run -p codify-app
```

**Los tests usan fuentes de mentira.** Eso los hace rápidos y deterministas, y significa que
**nadie ha comprobado que el descubrimiento real funcione** hasta que alguien abra esto en su
máquina. Está dicho en el plan como renuncia, no se descubre aquí.

### SC-001 · Que ya sepa qué tengo

Abrir la aplicación en una máquina con algo instalado, y mirar.

- **Pasa**: aparece lo que hay, sin haber escrito ni configurado nada.
- **Falla**: no aparece, o aparece algo que no está.

En la máquina donde se escribió este spec debería encontrar **`llama-server`, `llama-cli`,
`claude` y `lms`** — los cuatro en el `PATH`, ninguno visible hoy para la aplicación.

### SC-002 · Que no se invente nada

- **Pasa**: con nada escuchando, **no se menciona ninguna dirección**. Se dice qué podría
  instalarse.
- **Falla**: aparece `http://localhost:11434` o cualquier otra dirección que nadie comprobó.

### SC-006 · Llegar a generar contexto sin pasar por ningún formulario

Con algo local detectado, generar contexto de un repositorio **sin tocar la superficie de
conexión**.

Es el criterio que de verdad mide si este ciclo sirve: las dos rondas de validación anteriores
fallaron intentando conectar algo **remoto**, y nadie llegó a probar el camino local — que es el
que el producto dice servir primero.

### Clasificación · lo que hay que mirar con cuidado

Si aparece **`claude`**, comprobar que se enseña como lo que es: un programa de la máquina cuya
**inferencia ocurre fuera**. Que esté en la lista es correcto —el producto quiere los dos caminos—;
que apareciera como local no lo sería.

### Y el caso de la máquina vacía

Probarlo también sin nada instalado. **Pasa** si dice qué podría instalarse. **Falla** si se queda
en blanco o enseña una dirección.
