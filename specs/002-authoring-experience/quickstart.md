# Quickstart — Validación de la experiencia (002)

Escenarios que prueban la experiencia de punta a punta. Complementan los tests automatizados: **el DOM se valida aquí** (ver la desviación documentada en `plan.md`).

## Prerrequisitos

- Toolchain Rust + dependencias de sistema de Tauri.
- Un backend local de modelo (Ollama o `llama.cpp` server) con un modelo instalado. **Uno de los escenarios exige tenerlo apagado.**
- Fixture: un repositorio cuyo `README` **referencia** un `SPEC` hermano (el mismo patrón que la auditoría que originó el proyecto), más una URL privada para el caso de referencia no resuelta.

## Comandos

```bash
cargo test --workspace     # unit, contract, paridad del catálogo y contrato de interfaz
cargo run -p codify-app    # levanta la aplicación
```

> **Qué de esto ya no depende de mirar.** S5–S8 estaban escritos como inspección manual. Sus
> propiedades **mecánicas** están hoy fijadas en `crates/codify-app/tests/ui_contract.rs` y las
> verifica el build: orden de tabulación, indicador de foco, alcanzabilidad del punto de quiebre
> responsivo, texto fuera del catálogo, elementos con dos dueños de su texto, y motivos del
> proveedor sin traducir.
>
> El recorrido humano **sigue valiendo** —nadie automatiza «se entiende»— pero ya no es la única
> red. La primera pasada real, midiendo la interfaz en un navegador en vez de leyéndola, encontró
> cuatro defectos; los tests existen para que no vuelvan.

> **Hallazgos de la primera pasada con modelo real** (2026-08-23): están en
> `001/quickstart.md`, y **uno es de este spec** — la sesión puede terminar en `Failed` sin
> artefactos y **sin motivo**, contra lo que exige FR-028 ([#24](https://github.com/jorelcb/codify/issues/24)).
> Los demás: [#23](https://github.com/jorelcb/codify/issues/23) y [#25](https://github.com/jorelcb/codify/issues/25).

## Cobertura automatizada

**S5, S6 y S7 ya no dependen de mirar**: sus propiedades mecánicas las verifica el build en
`crates/codify-app/tests/ui_contract.rs`. S8 tiene cubierta la presentación. Los cuatro
primeros exigen un backend real y una persona — son justo los que preguntan «¿se entiende lo
que está pasando?», y eso no lo automatiza nadie.

| Escenario | Estado | Detalle |
|---|---|---|
| S1 · Ver trabajar al agente | 🧑 **humano** | el núcleo está cubierto (`us1_grounded.rs`, `us1_nonblocking.rs`); SC-001 mide lo que una persona logra *enumerar* mirando |
| S2 · Cancelar a mitad | 🧑 **humano** | el núcleo sí: `us1_cancellation.rs` (3), incluida la cancelación de la llamada en vuelo |
| S3 · Leer el fundamento | 🧑 **humano** | render verificado en navegador; SC-002 exige que alguien **ajeno** acierte ≥90 % sin instrucción |
| S4 · Onboarding del proveedor | 🧑 **humano** | `contract_provider_discovery.rs` y el test de motivos cubren el mecanismo; falta apagar el backend y ver si guía |
| S5 · Idioma | ✅ **automatizado** | `ningun_texto_visible_escapa_al_catalogo`, `quien_pinta_a_mano_repinta_al_cambiar_de_idioma`, `ningun_elemento_tiene_dos_duenos_de_su_texto`, paridad del catálogo |
| S6 · Solo teclado | ✅ **automatizado** | `no_hay_tabindex_positivo`, `el_foco_siempre_deja_rastro_visible` |
| S7 · Ventana mínima | ✅ **automatizado** | `el_punto_de_quiebre_responsivo_es_alcanzable` |
| S8 · Repositorio vacío | ◐ **parcial** | `el_repositorio_vacio_tiene_presentacion_propia`; la entrevista con un modelo real, pendiente |

**Fixture**: `./scripts/quickstart-fixture.sh`.

## Escenarios

### S1 — Ver trabajar al agente (US1 · SC-001)
1. Abrir la app, elegir el repositorio fixture en **modo local**, iniciar.
2. **Esperado**: los bloques aparecen **en vivo y en orden** (lista, lee el README, sigue la referencia al SPEC). La interfaz **no se congela**: se puede desplazar mientras el agente trabaja.
3. **Esperado**: la referencia privada aparece como **no resuelta con su motivo**, y sigue visible al terminar.
4. **Esperado**: el modo local es visible de un vistazo; cualquier intento de salida aparece bloqueado.
5. **Criterio**: alguien que solo miró la pantalla puede enumerar qué leyó el agente y qué no. Sin logs, sin terminal.

### S2 — Cancelar a mitad de trabajo (FR-022/023 · SC-008)
1. Iniciar una sesión y pulsar **cancelar** mientras el agente genera.
2. **Esperado**: la cancelación surte efecto **sin esperar** a que termine la llamada al modelo en curso.
3. **Esperado**: la sesión queda en `cancelled` y muestra el **balance de escrituras**: qué artefactos llegaron al repositorio y cuáles no.
4. **Criterio**: el usuario puede describir el estado del repositorio **sin inspeccionar archivos**. Contrastar con `git status`.

### S3 — Leer el contexto y su fundamento (US3 · SC-002)
1. Con la sesión terminada, abrir un artefacto completo **desde cualquier punto** de la corriente.
2. **Esperado**: los tres estados (fundamentado / tentativo / contradicción) se distinguen **con etiqueta, forma y color**; las fuentes de lo fundamentado son consultables.
3. **Prueba dura de FR-026**: poner la pantalla en **escala de grises** — la distinción debe seguir siendo inequívoca.
4. **Criterio**: alguien ajeno a la sesión acierta el estado de ≥90 % de los fragmentos **sin instrucción previa**.

### S4 — Onboarding del proveedor (FR-019 · FR-028)
1. **Apagar el backend local** y abrir la aplicación.
2. **Esperado**: la app dice **que no hay backend y qué hacer al respecto**. No se queda en silencio ni muestra un error crudo.
3. Encender el backend y volver a sondear: **esperado**, aparecen los modelos disponibles y se puede elegir uno.

### S5 — Idioma de la interfaz (FR-016b · SC-009)
1. Recorrer la aplicación completa en **español** y luego en **inglés**.
2. **Esperado**: **cero** claves crudas y cero textos en el idioma equivocado.
3. **Esperado**: cambiar el idioma de la interfaz **no** cambia el idioma de los artefactos generados (comprobar que un artefacto en español sigue en español con la interfaz en inglés).

### S6 — Solo teclado (FR-025 · FR-027)
1. Recorrer un flujo completo **sin tocar el ratón**: elegir repositorio, iniciar, cancelar, abrir un artefacto, cerrarlo.
2. **Esperado**: toda acción es alcanzable, el foco es siempre visible, y el orden de tabulación sigue el orden visual.

### S7 — Ventana mínima (SC-007)
1. Reducir la ventana a su tamaño mínimo soportado.
2. **Esperado**: la corriente y la vista de artefacto siguen siendo utilizables, sin recorte de contenido ni desplazamiento horizontal.

### S8 — Repositorio vacío (edge case)
1. Apuntar a un directorio vacío.
2. **Esperado**: la aplicación **conduce una entrevista**; no muestra una pantalla en blanco ni un error, y **no inventa** contexto.

## Mapeo a criterios de éxito

| Escenario | Cubre |
|---|---|
| S1 | SC-001, SC-006 |
| S2 | SC-005, SC-008 |
| S3 | SC-002 |
| S4 | FR-019, FR-028 |
| S5 | SC-009 |
| S6 | FR-025, FR-027 |
| S7 | SC-007 |
| S1+S3 | SC-003 (todo en una sola aplicación) |
