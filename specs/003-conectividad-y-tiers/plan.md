# Implementation Plan: Conectividad y reparto de modelos

**Branch**: `feat/conectividad-y-tiers` | **Date**: 2026-08-26 | **Spec**: [spec.md](./spec.md)
**Input**: issue [#42](https://github.com/jorelcb/codify/issues/42)

## Summary

Añadir un segundo proveedor —remoto— y repartir el trabajo entre tiers, **sin que el modo local
deje de ser incapaz de salir a la red**. Lo primero es rutina; lo segundo es el plan.

## Technical Context

**Language/Version**: Rust 1.98 (workspace `codify-core` + `codify-app`)
**Primary Dependencies**: `reqwest` (ya presente), `tokio`, `oauth2` (device-authorization flow),
`keyring` (almacén del sistema operativo)
**Storage**: credenciales en el keyring del sistema; **nada** en archivos del proyecto
**Testing**: `cargo test --workspace`; fitness functions en `crates/codify-core/tests/`
**Target Platform**: escritorio macOS / Linux / Windows
**Project Type**: aplicación de escritorio con núcleo hexagonal reutilizable
**Performance Goals**: no hay objetivo nuevo; el reparto por tier existe para bajar coste y
latencia percibida, y eso lo mide SC-004, no un número de milisegundos
**Constraints**: cero-egress estructural en modo local **[NON-NEGOTIABLE]**
**Scale/Scope**: un usuario, un repositorio a la vez, dos tiers

## Constitution Check

| Principio | Cómo lo cumple este diseño |
|---|---|
| **Arquitectura hexagonal** | El adapter remoto es infraestructura y satisface el port `ModelProvider` que ya existe. Aparecen **dos capability ports nuevos** en `application/`: custodiar una credencial y conectar una cuenta. Ninguno lo nombra el Dominio, así que ninguno vive ahí. |
| **Ports en la capa que los nombra** | «Credencial» y «cuenta conectada» son vocabulario de aplicación, no de dominio: el Dominio de `001` habla de sesión, referencia y artefacto. |
| **Nombres sin decoración** | `CredentialStore`, `AccountConnector`. Sin `…Port`, sin `I…`. |
| **Test-first** | Los tests de cada tarea van antes, y el de FR-008 es **de compilación**: se escribe primero y debe fallar al compilar. |
| **Cero-egress estructural** | Es el eje del plan. Ver abajo. |
| **Greenfield** | `CoreBuilder` cambia de forma sin ruta de migración. |

**Resultado**: sin violaciones. La sección de complejidad recoge el único punto que la añade a
propósito.

## La decisión que ordena el resto: cómo se sostiene FR-008

Hoy el cero-egress se sostiene con dos comprobaciones **en tiempo de ejecución**:
`ProviderRegistry::for_mode` rechaza un proveedor no local si el modo es `Local`, y
`LocalOpenAiCompatProvider::new` rechaza un endpoint que no sea loopback. Bastaban porque **no
existía ningún adapter capaz de salir**: la garantía la daba la ausencia.

En cuanto exista uno, esas comprobaciones siguen siendo correctas pero dejan de ser suficientes
para la palabra «estructuralmente». Un rechazo en tiempo de ejecución dice *«no lo hace»*; FR-008
pide *«no puede»*.

**Decisión: el modo pasa a ser un parámetro de tipo del constructor del grafo.**

`CoreBuilder<Local>` **no tiene** el método que acepta un proveedor remoto. No es que lo rechace:
no existe. Intentarlo es un **error de compilación**, y ese es el enunciado más fuerte de
«imposible» que este lenguaje permite escribir.

Las dos comprobaciones de runtime **se quedan** —defensa en profundidad, y siguen cubriendo el
caso de un proveedor construido por otra vía—, pero dejan de ser lo único.

Consecuencia que hay que aceptar: cambiar de modo deja de ser mutar un campo y pasa a ser
construir otro grafo, que es exactamente lo que FR-008a pide y lo que la clarificación del
2026-08-26 decidió.

## Project Structure

### Documentation (this feature)

```
specs/003-conectividad-y-tiers/
├── spec.md
├── plan.md              # este archivo
├── research.md          # decisiones y alternativas descartadas
├── data-model.md        # entidades del dominio y de aplicación
├── quickstart.md        # escenarios de validación
├── contracts/
│   ├── ports.md         # CredentialStore, AccountConnector, ModelProvider remoto
│   └── skin-commands.md # comandos y eventos que la piel gana
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```
crates/codify-core/
├── src/
│   ├── application/
│   │   ├── ports.rs             # + CredentialStore, AccountConnector
│   │   └── deps.rs              # ProviderRegistry: tier real, no un solo proveedor
│   └── infrastructure/
│       ├── composition.rs       # CoreBuilder<M>: el modo, en el tipo
│       ├── providers/
│       │   └── remote.rs        # nuevo: adapter remoto
│       └── secrets/
│           └── keyring.rs       # nuevo: CredentialStore contra el SO
└── tests/
    ├── egress_guard.rs          # se extiende
    ├── compile_fail/            # nuevo: el grafo local no admite remoto
    └── contract_credential_store.rs  # nuevo: suite hex-integration

crates/codify-app/
├── src/commands.rs              # connect_provider, list_connections, disconnect
└── ui/                          # conectar cuenta, ver modo, ver qué salió
```

## Complexity Tracking

| Lo que añade complejidad | Por qué se acepta | Qué se descartó |
|---|---|---|
| Parámetro de tipo en `CoreBuilder` | Es lo que convierte FR-008 de afirmación en propiedad comprobable por el compilador. Sin él, «estructuralmente imposible» sería una forma de hablar. | Seguir solo con la comprobación de runtime — deja la garantía en «no lo hace» |
| Dos vías de autenticación | Decidido en la clarificación del 2026-08-26: sin la vía de credencial, este spec no sirve a los proveedores que lo motivan | Solo device-flow (deja fuera a los frontier mayoritarios); solo credencial (contradice la constitución) |
| Dependencia del keyring del SO | FR-002 y FR-004 no se pueden cumplir sin un almacén que no sea nuestro | Cifrar un archivo propio: mueve el problema a dónde guardar la llave |
