//! **Application capability ports**: capacidades que el Dominio *nunca nombra* y que
//! existen para que el orquestador haga su trabajo (constitución, Principio I).
//!
//! Todas las firmas hablan tipos de dominio. Ningún tipo de vendor/SDK/HTTP cruza esta
//! frontera: los adapters traducen antes (incluidos los errores → `CoreError`).

use crate::domain::change::{ApprovalDecision, ChangeProposal, Diff};
use crate::domain::error::Result;
use crate::domain::reference::{Reference, ReferenceOrigin, Repository};
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
