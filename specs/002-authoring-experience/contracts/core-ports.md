# Contracts — Ports nuevos del núcleo

Tres **application capability ports** (el Dominio no los nombra). Firmas ilustrativas; sus adapters viven en `infrastructure/`. Todas hablan tipos de dominio: ningún tipo de `tokio-util`, `reqwest` ni Tauri cruza esta frontera.

## `Cancellation` — abortar una sesión en curso (FR-022/FR-023)

```rust
#[async_trait]
pub trait Cancellation: Send + Sync {
    /// Consulta barata para los puntos de control del loop.
    fn is_cancelled(&self) -> bool;

    /// Resuelve cuando se cancela. Permite componer con `select!` y **abortar la
    /// petición al modelo en vuelo**, en vez de esperar a que termine.
    async fn cancelled(&self);
}
```

**Reglas de contrato**
- Una vez cancelado, `is_cancelled()` devuelve `true` para siempre (no se "descancela").
- `cancelled()` puede esperarse desde varios sitios a la vez y todos deben despertar.
- Cancelar **nunca** deja el repositorio en un estado que el usuario no pueda conocer: la sesión reporta sus `WriteRecord`.
- Adapter real: `TokenCancellation` (envuelve `tokio_util::sync::CancellationToken`). Fake: bandera en memoria, para tests deterministas.

## `ArtifactWriter` — llevar los artefactos al repositorio (FR-017)

```rust
#[async_trait]
pub trait ArtifactWriter: Send + Sync {
    async fn write(&self, path: &str, content: &str) -> Result<WriteRecord>;
    /// Contenido actual del archivo, si existe. Es lo que hará posible
    /// "no sobrescribir sin diff y aprobación" (US3 de 001).
    async fn read_existing(&self, path: &str) -> Result<Option<String>>;
}
```

**Reglas de contrato**
- `path` es **relativo a la raíz del repositorio**. Rutas absolutas o con `..` se rechazan con `CoreError::Invalid` — las mismas defensas que el navegador de repositorio.
- Crea los directorios intermedios que falten (`context/` puede no existir).
- Devuelve un `WriteRecord` por llamada; el loop los acumula y la sesión los reporta.
- Un fallo de escritura **no aborta** la sesión entera: se registra como `Failed(motivo)` y se sigue con el resto de artefactos. Perder un archivo es malo; perder los otros tres por eso es peor.

## `ProviderDiscovery` — sondear el backend de modelo (FR-019/FR-028)

```rust
#[async_trait]
pub trait ProviderDiscovery: Send + Sync {
    async fn probe(&self) -> ProviderStatus;
}

pub struct ProviderStatus {
    pub reachable: bool,
    pub endpoint: String,
    pub models: Vec<String>,
    /// Motivo accionable cuando no sirve. Nunca vacío si `reachable == false`.
    pub issue: Option<ProviderIssue>,
}

/// El motivo como **dato**, no como prosa.
pub enum ProviderIssue {
    NoModels,          // responde, pero no hay ningún modelo instalado
    NotListening,      // no hay nada escuchando en el endpoint
    EndpointNotLocal,  // el endpoint sale de la máquina y el modo local no lo admite
}

impl ProviderIssue {
    /// Identificador estable con el que la piel elige el texto. Es parte del contrato.
    pub fn code(&self) -> &'static str;
}
```

**Reglas de contrato**
- `probe()` **no falla**: devuelve `reachable: false` con un `issue` que la piel traduce a qué hacer. Un error opaco es exactamente lo que FR-019 viene a evitar.
- El núcleo **no redacta prosa para humanos**. Devolver la frase ya escrita le fijaría un idioma y volvería SC-009 —cero cadenas sin traducir— indemostrable; además sería presentación colándose en la aplicación. El núcleo nombra el motivo con un código estable y la piel elige el texto en `provider.issue.<code>`.
- En modo local el `endpoint` debe ser loopback; el adapter reutiliza la validación existente del proveedor local.
- No produce efectos: no genera, no descarga, no modifica nada.

## Cambios en la superficie existente

| Elemento | Cambio | Motivo |
|---|---|---|
| `AuthoringService::start_session` | **Retorna el `SessionId` de inmediato**; el loop corre en segundo plano | FR-022: la interfaz no puede quedarse bloqueada minutos |
| `AuthoringService::cancel_session(id)` | **Nuevo** | FR-023 |
| `SessionSnapshot` | **+ `writes: Vec<WriteRecord>`** | FR-017: declarar qué llegó al repositorio |
| `AuthoringDeps` | **+ los tres ports nuevos** | Inyección por constructor en el composition root |
| `AuditKind` | **+ `ArtifactWritten`, `SessionCancelled`** | La piel se entera por el canal que ya existe |

> Todos los cambios son **aditivos** salvo la semántica de `start_session`, que cambia de bloqueante a no bloqueante. Es greenfield y la única consumidora es nuestra propia piel.

## Verificación (contract tests, real + fake)

- `Cancellation`: cancelar durante la ingesta detiene el loop y la sesión queda en `Cancelled` reportando sus escrituras; cancelar **durante una llamada al modelo** la aborta sin esperar a que termine.
- `ArtifactWriter`: escribe y relee lo escrito; rechaza rutas fuera del repositorio; crea directorios intermedios; un fallo aislado no arrastra al resto.
- `ProviderDiscovery`: con backend ausente devuelve `reachable: false` **con `issue` presente**; nunca devuelve `Err`. Los códigos de `ProviderIssue` son distintos entre sí —dos que colisionaran se presentarían como el mismo problema— y **cada uno tiene texto en ambos idiomas**, verificado recorriendo el catálogo.
