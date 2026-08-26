# Contracts — Ports

## `CredentialStore` (capability port, `application/`)

Custodia secretos fuera del proceso. Lo nombra la aplicación, no el Dominio.

```
guardar(referencia, secreto) -> Result<()>
obtener(referencia) -> Result<Option<Secreto>>
borrar(referencia) -> Result<()>
disponible() -> bool
```

**Reglas de contrato**

1. `disponible()` MUST poder responder **sin** guardar nada: FR-004 exige poder avisar antes de
   que el usuario intente conectar.
2. Un adapter MUST NOT caer a un almacenamiento alternativo si el del sistema falla. Falla y se
   dice — ver D3 en `research.md`.
3. `borrar` MUST ser idempotente: desconectar dos veces no es un error.
4. El secreto MUST NOT aparecer en `Debug`, en logs ni en mensajes de error. Los tipos que lo
   transportan implementan `Debug` redactado.

**Suite de contrato**: `contract_credential_store.rs`, corriendo contra el adapter real y contra
un doble en memoria — el patrón que ya usan los otros ports de este repositorio.

## `AccountConnector` (capability port, `application/`)

Obtiene una credencial del usuario. Dos implementaciones tras una sola frontera (D4).

```
iniciar(proveedor) -> Result<Desafio>     // device-flow: código + URL; credencial: solicitud
completar(desafio, respuesta?) -> Result<ProviderConnection>
```

**Reglas de contrato**

1. El sondeo del device-flow MUST tener su propio límite de tiempo, **no** el de generación:
   esperar a una persona y esperar a un modelo son cosas distintas.
2. Abandonar o denegar MUST dejar el sistema como estaba, sin conexión a medias (US1, escenario 4).
3. `completar` MUST guardar por `CredentialStore` y devolver una conexión **sin** el secreto.

## `ModelProvider` remoto (adapter, `infrastructure/`)

No hay port nuevo: satisface el que ya existe desde `001`.

**Reglas de contrato**

1. `is_local()` MUST devolver `false`. Es lo que `ProviderRegistry::for_mode` usa para rechazarlo
   en modo local — la defensa en profundidad que se mantiene junto a la del tipo.
2. `tier_hint()` MUST devolver el tier **declarado al conectar**, no uno inferido.
3. Un fallo de autorización MUST distinguirse de un fallo del modelo: llegan como
   `SessionFailure` distintos (`002`-FR-028) porque piden cosas distintas del usuario.

## Lo que el tipo garantiza y ningún test necesita comprobar

`CoreBuilder<Local>` no expone el método que acepta un proveedor remoto. Se prueba con un test
**de compilación fallida**, no con una aserción: lo que se comprueba es que el programa no existe.
