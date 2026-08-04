//! **Application capability ports**: capacidades que el Dominio *nunca nombra* y que
//! existen para que el orquestador haga su trabajo (constitución, Principio I).
//!
//! Todas las firmas hablan tipos de dominio. Ningún tipo de vendor/SDK/HTTP cruza esta
//! frontera: los adapters traducen antes (incluidos los errores → `CoreError`).

use crate::domain::change::{ApprovalDecision, ChangeProposal, Diff};
use crate::domain::error::Result;
use crate::domain::reference::{Reference, ReferenceOrigin, Repository};
use crate::domain::write::WriteRecord;
use async_trait::async_trait;

// ---------------------------------------------------------------------------
// Modelos
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Tier {
    /// Alta frecuencia y bajo riesgo: refinamiento, clasificación, monitoreo.
    Cheap,
    /// Generación pesada y grounded.
    Heavy,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Role {
    User,
    Assistant,
    Tool,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Message {
    pub role: Role,
    pub content: String,
}

impl Message {
    pub fn user(content: impl Into<String>) -> Self {
        Self {
            role: Role::User,
            content: content.into(),
        }
    }
    pub fn assistant(content: impl Into<String>) -> Self {
        Self {
            role: Role::Assistant,
            content: content.into(),
        }
    }
    pub fn tool(content: impl Into<String>) -> Self {
        Self {
            role: Role::Tool,
            content: content.into(),
        }
    }
}

/// Herramienta ofrecida al modelo. El schema viaja como texto JSON para no acoplar el
/// port a una librería de serialización concreta.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ToolSpec {
    pub name: String,
    pub description: String,
    pub json_schema: String,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ToolCall {
    pub id: String,
    pub name: String,
    /// Argumentos como texto JSON.
    pub arguments: String,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CompletionRequest {
    pub system: String,
    pub messages: Vec<Message>,
    pub tools: Vec<ToolSpec>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CompletionOutput {
    Text(String),
    ToolCalls(Vec<ToolCall>),
}

/// Conexión a un backend de modelo (local `llama.cpp`/Ollama, o remoto). Decisión raíz D3.
#[async_trait]
pub trait ModelProvider: Send + Sync {
    async fn complete(&self, request: CompletionRequest) -> Result<CompletionOutput>;

    /// Determina la elegibilidad en modo `Local`: el composition root **no admite**
    /// proveedores no locales cuando la sesión es local (cero-egress estructural).
    fn is_local(&self) -> bool;

    fn tier_hint(&self) -> Tier;

    fn name(&self) -> &str;
}

// ---------------------------------------------------------------------------
// Navegación del repositorio (ingesta dirigida por el agente — D4)
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum EntryKind {
    File,
    Dir,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Entry {
    pub path: String,
    pub kind: EntryKind,
    pub size: u64,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct FileContent {
    pub path: String,
    pub content: String,
    /// `true` cuando se recortó: obliga a declararlo, nunca truncar en silencio.
    pub truncated: bool,
}

#[async_trait]
pub trait RepoNavigator: Send + Sync {
    async fn list(&self, path: &str) -> Result<Vec<Entry>>;
    async fn read(&self, path: &str) -> Result<FileContent>;
    /// Señales estructurales del repo (manifiestos, layout) para fundamentar sin leerlo todo.
    async fn describe(&self) -> Result<Repository>;
}

/// Resuelve referencias aludidas. v1: rutas locales y URLs **públicas**; lo que requiera
/// autenticación se reporta como no resuelto, jamás se inventa (FR-003/FR-004/SC-006).
#[async_trait]
pub trait ReferenceResolver: Send + Sync {
    async fn resolve(&self, origin: &ReferenceOrigin) -> Reference;
}

// ---------------------------------------------------------------------------
// Diff / aprobación / auditoría (D6)
// ---------------------------------------------------------------------------

pub trait DiffEngine: Send + Sync {
    fn make(&self, before: &str, after: &str) -> Diff;
    fn apply(&self, before: &str, diff: &Diff) -> Result<String>;
    fn revert(&self, after: &str, diff: &Diff) -> Result<String>;
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Question {
    pub text: String,
    pub suggestions: Vec<String>,
}

/// Borde humano. Su **adapter es la piel** (Interface Adapter): la piel renderiza el diff y
/// captura la decisión. El core sigue dependiendo solo de esta abstracción.
#[async_trait]
pub trait Prompter: Send + Sync {
    async fn ask(&self, question: Question) -> Result<String>;
    /// Solo se invoca para propuestas de alto impacto: son las que bloquean la escritura.
    async fn present(&self, proposal: &ChangeProposal) -> Result<ApprovalDecision>;
}

pub trait AuditSink: Send + Sync {
    fn record(&self, event: crate::domain::audit::AuditEvent);
}

/// Detecta el idioma dominante del repo (FR-019). El override del usuario lo aplica la sesión.
#[async_trait]
pub trait LocaleDetector: Send + Sync {
    async fn detect(&self, repo: &Repository) -> String;
}

// ---------------------------------------------------------------------------
// Cancelación, escritura y descubrimiento de proveedor (spec 002)
// ---------------------------------------------------------------------------

/// Señal de cancelación de una sesión en curso.
///
/// Se expresa como trait propio para que el token concreto (`tokio-util`) no cruce hacia el
/// núcleo. Los dos métodos cubren usos distintos: `is_cancelled` para los puntos de control
/// del loop, y `cancelled` para componer con `select!` y **abortar la petición al modelo en
/// vuelo** en vez de esperar a que termine.
#[async_trait]
pub trait Cancellation: Send + Sync {
    fn is_cancelled(&self) -> bool;

    /// Resuelve cuando se cancela. Puede esperarse desde varios sitios a la vez.
    async fn cancelled(&self);

    /// Dispara la cancelación. La señal es un objeto compartido entre quien cancela
    /// (el servicio, a petición de la piel) y quien la observa (el loop).
    fn cancel(&self);
}

/// Crea una señal de cancelación **por sesión**.
///
/// Hace falta una factoría porque la señal no puede compartirse entre sesiones —cancelar
/// una abortaría todas— y solo la infraestructura sabe construir la señal concreta.
pub trait CancellationFactory: Send + Sync {
    fn create(&self) -> std::sync::Arc<dyn Cancellation>;
}

/// Lleva los artefactos generados al repositorio.
///
/// Sin este port el núcleo produce contexto que nunca sale de la memoria — el producto no
/// entregaría su resultado.
#[async_trait]
pub trait ArtifactWriter: Send + Sync {
    /// Escribe el contenido. Devuelve el registro de lo ocurrido, incluso si falló:
    /// un fallo aislado **no** debe abortar el resto de la sesión.
    async fn write(&self, path: &str, content: &str) -> WriteRecord;

    /// Contenido actual del archivo, si existe. Es lo que hará posible "no sobrescribir sin
    /// diff y aprobación" cuando llegue esa historia.
    async fn read_existing(&self, path: &str) -> Result<Option<String>>;
}

/// Estado del backend de modelo, para poder guiar al usuario en vez de fallar en silencio.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ProviderStatus {
    pub reachable: bool,
    pub endpoint: String,
    pub models: Vec<String>,
    /// Motivo accionable cuando no responde. **Nunca vacío** si `reachable == false`.
    pub detail: Option<String>,
}

impl ProviderStatus {
    pub fn reachable(endpoint: impl Into<String>, models: Vec<String>) -> Self {
        Self {
            reachable: true,
            endpoint: endpoint.into(),
            models,
            detail: None,
        }
    }

    pub fn unreachable(endpoint: impl Into<String>, detail: impl Into<String>) -> Self {
        Self {
            reachable: false,
            endpoint: endpoint.into(),
            models: Vec::new(),
            detail: Some(detail.into()),
        }
    }
}

/// Sondea el backend de modelo. **No falla**: un error opaco es justo lo que hay que evitar.
#[async_trait]
pub trait ProviderDiscovery: Send + Sync {
    async fn probe(&self) -> ProviderStatus;
}
