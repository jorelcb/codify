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
use serde::{Deserialize, Serialize};

// ---------------------------------------------------------------------------
// Modelos
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
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

/// Un secreto que el sistema custodia pero no mira.
///
/// `Debug` está redactado a propósito: `003`-FR-002 prohíbe que la credencial llegue a un
/// registro, y la forma más fiable de cumplirlo no es acordarse de no imprimirla, sino que
/// imprimirla no sirva de nada.
#[derive(Clone, PartialEq, Eq)]
pub struct Secreto(String);

impl Secreto {
    pub fn new(valor: impl Into<String>) -> Self {
        Self(valor.into())
    }

    /// Lo expone. El nombre es largo a propósito: quien lo escriba debería notar que lo hace.
    pub fn exponer_para_la_peticion(&self) -> &str {
        &self.0
    }
}

impl std::fmt::Debug for Secreto {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str("Secreto(<redactado>)")
    }
}

/// Dónde vive un secreto dentro del almacén del sistema.
///
/// **No es el secreto**: es la clave con la que pedírselo. Por eso sí se serializa — sin ella,
/// una cuenta no sobreviviría a reiniciar la aplicación (US1, escenario 2).
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ReferenciaDeCredencial(String);

impl ReferenciaDeCredencial {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }
    pub fn as_str(&self) -> &str {
        &self.0
    }
}

/// Custodia secretos **fuera del proceso** (`003`-FR-002).
///
/// Lo nombra la aplicación, no el Dominio: el Dominio de `001` habla de sesión, referencia y
/// artefacto. Una credencial es vocabulario de esta capa.
#[async_trait::async_trait]
pub trait CredentialStore: Send + Sync {
    /// `false` si el almacén del sistema no está disponible. MUST poder responder **sin**
    /// guardar nada: FR-004 exige avisar antes de que el usuario intente conectar.
    fn disponible(&self) -> bool;
    async fn guardar(&self, r: &ReferenciaDeCredencial, s: Secreto) -> Result<()>;
    async fn obtener(&self, r: &ReferenciaDeCredencial) -> Result<Option<Secreto>>;
    /// Idempotente: desconectar dos veces no es un error.
    async fn borrar(&self, r: &ReferenciaDeCredencial) -> Result<()>;
}

/// Lo que el usuario tiene que hacer para autorizar.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Desafio {
    /// Autorización delegada: la aplicación enseña código y dirección, el usuario va fuera.
    Delegada { codigo: String, url: String },
    /// El proveedor no ofrece delegada: hay que pedir la credencial **una sola vez**.
    PideCredencial { instrucciones: String },
}

/// Obtiene una credencial del usuario (`003`-FR-001).
///
/// Dos implementaciones tras una sola frontera: la diferencia entre las vías es **cómo se
/// obtiene** el secreto, no qué se hace con él. Custodia, uso y revocación son idénticas, así
/// que la frontera va donde termina esa diferencia.
#[async_trait::async_trait]
pub trait AccountConnector: Send + Sync {
    async fn iniciar(&self) -> Result<Desafio>;
    /// `respuesta` lleva la credencial cuando el desafío la pedía; para la vía delegada es
    /// `None` y el adapter sondea por su cuenta, **con su propio límite de tiempo**.
    async fn completar(&self, desafio: &Desafio, respuesta: Option<Secreto>) -> Result<Secreto>;
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

/// Por qué el backend no sirve, como **dato** y no como prosa.
///
/// El núcleo no redacta texto para humanos: si devolviera la frase ya escrita, esa frase
/// tendría un idioma fijo y SC-009 —cero cadenas sin traducir— dejaría de ser demostrable.
/// Nombrar el motivo y dejar que la piel lo redacte mantiene la presentación donde va y hace
/// que un test pueda recorrer el catálogo y comprobar que ningún motivo se quedó sin texto.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ProviderIssue {
    /// Responde, pero no tiene ningún modelo instalado.
    NoModels,
    /// No hay nada escuchando en el endpoint.
    NotListening,
    /// El endpoint apunta fuera de la máquina y el modo local no lo admite.
    EndpointNotLocal,
}

impl ProviderIssue {
    /// Identificador estable para que la piel elija el texto. Es parte del contrato.
    pub fn code(&self) -> &'static str {
        match self {
            ProviderIssue::NoModels => "no_models",
            ProviderIssue::NotListening => "not_listening",
            ProviderIssue::EndpointNotLocal => "endpoint_not_local",
        }
    }
}

/// Estado del backend de modelo, para poder guiar al usuario en vez de fallar en silencio.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ProviderStatus {
    pub reachable: bool,
    pub endpoint: String,
    pub models: Vec<String>,
    /// Motivo accionable cuando no sirve. **Nunca vacío** si `reachable == false`.
    pub issue: Option<ProviderIssue>,
}

impl ProviderStatus {
    pub fn reachable(endpoint: impl Into<String>, models: Vec<String>) -> Self {
        Self {
            reachable: true,
            endpoint: endpoint.into(),
            models,
            issue: None,
        }
    }

    pub fn unreachable(endpoint: impl Into<String>, issue: ProviderIssue) -> Self {
        Self {
            reachable: false,
            endpoint: endpoint.into(),
            models: Vec::new(),
            issue: Some(issue),
        }
    }
}

/// Sondea el backend de modelo. **No falla**: un error opaco es justo lo que hay que evitar.
#[async_trait]
pub trait ProviderDiscovery: Send + Sync {
    async fn probe(&self) -> ProviderStatus;
}
