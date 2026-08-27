# Data Model — Conectividad y reparto de modelos

## ProviderConnection (aplicación)

Una cuenta remota autorizada. Vive en `application/`, no en `domain/`: el Dominio de `001` habla
de sesión, referencia y artefacto — «cuenta conectada» es vocabulario de esta capa.

- **Campos**: `id`, `label` (lo que el usuario ve), `endpoint`, `tier` (`Cheap` | `Heavy`),
  `state` (`Connected` | `Expired` | `Revoked` | `CredentialMissing`).
- **`CredentialMissing`** existe porque faltaba (issue #48): la credencial puede desaparecer del
  almacén —el usuario limpia su llavero, otra aplicación la borra— y sin este estado la conexión
  seguía mostrándose `Connected` mientras el sistema la omitía en silencio al armar el grafo. Un
  estado que miente es peor que uno que falta.
- **Invariante que sostiene FR-002**: **no contiene la credencial**. Lleva la referencia con la
  que pedírsela al almacén. Un `ProviderConnection` serializado a la interfaz o a un registro no
  puede filtrar un secreto porque no lo tiene.
- **Regla**: el `tier` se **declara al conectar**. El sistema no puede saber si un endpoint sirve
  un modelo caro o barato, y adivinarlo produciría un reparto arbitrario.

## Mode (dominio, ya existe — cambia de significado)

`Local` | `Hybrid`. Hoy es un dato que se compara en tiempo de ejecución; pasa a ser además el
**parámetro de tipo del constructor del grafo**.

- **Invariante nueva (FR-008/008a)**: un grafo construido como `Local` **no puede** contener un
  proveedor capaz de salir a la red. No es que se rechace: el método que lo aceptaría no existe en
  `CoreBuilder<Local>`.
- **Transición**: cambiar de modo **reconstruye** el grafo. Una sesión en curso conserva el suyo
  (FR-008b).

## Tier (dominio, ya existe — pasa a tener dos habitantes)

`Cheap` | `Heavy`. Existía desde `001` pero con un solo proveedor real detrás, así que enrutar no
distinguía nada. Con dos, `ProviderRegistry::pick` empieza a significar lo que dice, y la
degradación declarada de `001`-FR-018 deja de ser el caso habitual.

## Credential (nunca en el modelo)

Se nombra aquí para dejar constancia de que **no es una entidad**. El secreto vive en el almacén
del sistema operativo y el resto del sistema maneja solo la referencia con la que pedirlo. No hay
un tipo que lo transporte, no se serializa y no aparece en ningún DTO — la forma más fiable de que
algo no se filtre es que no exista donde podría filtrarse.
