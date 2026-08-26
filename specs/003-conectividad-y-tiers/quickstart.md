# Quickstart — Conectividad y reparto de modelos

## Cobertura automatizada

| Qué | Cómo se comprueba | Dónde |
|---|---|---|
| El grafo local no admite un proveedor remoto | **no compila** | `tests/compile_fail/` |
| El registro sigue rechazando en runtime | aserción | `tests/egress_guard.rs` |
| Custodia de credenciales | suite de contrato contra adapter real y doble | `tests/contract_credential_store.rs` |
| El secreto no aparece en `Debug` ni en registros | aserción sobre la salida formateada | `tests/contract_credential_store.rs` |
| Reparto por tier | dos proveedores falsos, uno por tier | `tests/us2_tier_routing.rs` |
| Cada motivo de conexión tiene texto en ambos idiomas | recorre los códigos | `codify-app/tests/ui_contract.rs` |

## Escenarios

### S1 — Conectar por autorización delegada (US1 · SC-001)
1. Conectar un proveedor que ofrezca device-flow.
2. **Esperado**: la aplicación da código y dirección; tras autorizar fuera, la cuenta figura
   conectada **sin haber escrito ninguna credencial**.

### S2 — Conectar por credencial (US1 · SC-001, SC-002)
1. Conectar un proveedor que solo admita credencial; introducirla una vez.
2. **Esperado**: queda conectada, y la credencial **no vuelve a mostrarse** en la interfaz ni en
   la configuración.
3. Buscarla en el directorio de configuración y en los registros de la aplicación.
4. **Esperado**: **cero** coincidencias. Es SC-002, y se comprueba buscándola, no suponiéndolo.

### S3 — Reparto por tier (US2 · SC-004)
1. Con dos tiers conectados, refinar y luego generar un artefacto.
2. **Esperado**: el refinamiento fue al tier económico y la generación al de mayor capacidad, y
   ambas cosas se ven en la aplicación sin abrir herramientas de desarrollo.

### S4 — Sin keyring (FR-004)
1. Correr en un entorno donde el almacén del sistema no esté disponible.
2. **Esperado**: se dice, se ofrece seguir en local, y **no** se guarda nada en un archivo. Buscar
   la credencial en disco tras el intento: cero coincidencias.

### S5 — Cambiar de modo con sesión en curso (FR-008b · SC-007)
1. Arrancar una sesión en modo híbrido; a mitad, cambiar a local.
2. **Esperado**: la sesión viva termina con el modo con el que nació y se dice; la siguiente ya es
   local. **Sin reiniciar** la aplicación.

### S6 — Revocar (SC-006)
1. Desconectar una cuenta con la aplicación abierta.
2. **Esperado**: la siguiente tarea **no puede** usarla, sin reiniciar. Y la credencial ya no está
   en el keyring — comprobar con la herramienta del sistema, no con la aplicación.

### S7 — Qué salió del equipo (FR-010 · SC-005)
1. Terminar una sesión en modo híbrido.
2. **Esperado**: el usuario puede decir qué proveedor atendió cada tarea, solo mirando la
   aplicación.

## Mapeo a criterios de éxito

| Escenario | Cubre |
|---|---|
| S1, S2 | SC-001, SC-002 |
| S3 | SC-004 |
| S5 | SC-007 |
| S6 | SC-006 |
| S7 | SC-005 |
| compile_fail + egress_guard | **SC-003** |

> **SC-003 no tiene escenario manual, y es deliberado.** «En modo local el egress es cero» no se
> demuestra mirando una corrida: una corrida limpia es compatible con que exista la ruta. Lo que
> se comprueba es que el programa que la usaría **no compila**. Un escenario manual aquí daría
> confianza sin darla.
