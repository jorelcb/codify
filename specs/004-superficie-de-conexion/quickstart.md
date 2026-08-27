# Quickstart — validar la superficie de conexión

Dos partes, y la separación es el punto: **qué comprueba el build** y **qué sigue necesitando a
una persona delante**. Confundirlas fue lo que dejó pasar el defecto que originó este spec — los
tests estaban verdes y la pantalla no se entendía.

## Prerrequisitos

Ninguno más allá del entorno normal del proyecto. Esta feature no necesita backend vivo, ni
fixture, ni proveedor remoto: es presentación.

---

## Parte 1 — Lo que comprueba el build

```bash
cargo test -p codify-app --test ui_contract
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo fmt --all -- --check
```

**Se espera**: 18 tests de contrato de interfaz (14 de antes + 4 nuevos), y el workspace entero en
verde.

| Criterio | Test | Qué significa que pase |
|---|---|---|
| **SC-002** | `ningun_par_de_regiones_comparte_nombre_accesible` | Cero regiones con nombre repetido |
| **SC-005** | `los_campos_del_formulario_caben_en_la_ventana_minima` | Cada campo declara al menos 24 caracteres de ancho mínimo, y el contenedor envuelve |
| **SC-006** | `el_modo_no_puede_discrepar_entre_sus_dos_superficies` + `la_interfaz_no_tiene_su_propia_idea_del_modo` | Un solo escritor por superficie, misma función, sin copia en la interfaz |
| **FR-007** | los tests de catálogo que ya existían | Ninguna cadena se escapó ni quedó en un idioma |

### Verificar que los tests sirven

Un test que no se ha visto fallar no está verificado. Antes de dar el ciclo por bueno, inyectar
cada violación de [contracts/ui-contract.md](./contracts/ui-contract.md) —una por una, y
revertirla— y comprobar que el test que le corresponde **cae**. Si no cae, el test es decorativo.

---

## Parte 2 — Lo que necesita a una persona delante

Estos tres criterios miden si algo **se entiende**, y eso ningún análisis estático lo mide.

```bash
cargo run -p codify-app
```

### SC-001 · Conectar sin preguntar

Poner delante a alguien que no conoce la aplicación y pedirle que conecte un proveedor. Sin
explicar nada.

- **Pasa**: lo consigue sin preguntar qué va en ningún campo.
- **Falla**: pregunta una sola vez.

### SC-003 · Saber dónde estás sin ver

Recorrer la interfaz por regiones con un lector de pantalla, deteniéndose en cada una.

- **Pasa**: en cada parada se puede decir en qué región se está.
- **Falla**: dos regiones suenan igual. *(El test R1 lo previene, pero no comprueba que el nombre
  sea **útil**: «región 4» sería único e inservible.)*

### SC-004 · Señalar la decisión

Pedir, sin explicaciones: «señala qué control decide si algo sale de tu equipo».

- **Pasa**: señala el control de modo, a la primera.
- **Falla**: duda, o señala el indicador de proveedor.

### Comprobación de humo del modo — que ya no mienta

Esta sí se puede hacer sin ayuda, y comprueba el defecto de research D2:

1. Abrir la aplicación. La insignia dice **local**.
2. Desmarcar «solo local».
3. **La insignia cambia a híbrido.** Antes de este ciclo, no cambiaba nunca.
4. Iniciar una sesión y confirmar en la corriente de actividad que el grafo se armó en híbrido.

El paso 4 es el que demuestra que `003`-FR-008a se cumple de punta a punta, cosa que no ocurría.

---

## Qué anotar

Para cada escenario de la parte 2: qué pasó, y si falló, **qué preguntó o dónde dudó la persona**.
Un «falla» sin la pregunta concreta no dice qué arreglar.
