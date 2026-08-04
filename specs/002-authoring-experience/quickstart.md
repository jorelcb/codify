# Quickstart — Validación de la experiencia (002)

Escenarios que prueban la experiencia de punta a punta. Complementan los tests automatizados: **el DOM se valida aquí** (ver la desviación documentada en `plan.md`).

## Prerrequisitos

- Toolchain Rust + dependencias de sistema de Tauri.
- Un backend local de modelo (Ollama o `llama.cpp` server) con un modelo instalado. **Uno de los escenarios exige tenerlo apagado.**
- Fixture: un repositorio cuyo `README` **referencia** un `SPEC` hermano (el mismo patrón que la auditoría que originó el proyecto), más una URL privada para el caso de referencia no resuelta.

## Comandos

```bash
cargo test --workspace     # unit, contract y paridad del catálogo de cadenas
cargo run -p codify-app    # levanta la aplicación
```

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
