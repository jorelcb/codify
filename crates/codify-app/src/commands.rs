//! Comandos Tauri (T029) — el **driving adapter** de la piel.
//!
//! Superficie definida en `specs/001-context-authoring/contracts/tauri-commands.md`.
//! Cada comando delega en el Application Service del núcleo y traduce sus tipos a DTOs
//! serializables: los tipos de dominio no cruzan hacia la ventana.

use crate::adapters::{
    EventAuditSink, PendingDecisions, StatePayload, SystemClock, WindowPrompter,
};
use codify_core::application::connections::{ConnectionState, ProviderConnection};
use codify_core::application::ports::{
    AccountConnector, CredentialStore, Desafio, ReferenciaDeCredencial, Secreto,
};
use codify_core::application::ports::{ModelProvider, Tier};
use codify_core::application::ports::{ProviderDiscovery, ProviderIssue};
use codify_core::application::service::{
    AuthoringService, ContextAuthoring, SessionSnapshot, StartSession,
};
use codify_core::domain::change::{ApprovalDecision, ProposalId, Verdict};
use codify_core::domain::context::{ContextArtifact, Groundedness};
use codify_core::domain::session::{Mode, SessionId};
use codify_core::infrastructure::cancel::TokenCancellationFactory;
use codify_core::infrastructure::composition::{CoreBuilder, Hybrid, Local, ModoDelGrafo};
use codify_core::infrastructure::diff::engine::SimilarDiffEngine;
use codify_core::infrastructure::diff::risk::ConservativeRiskClassifier;
use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
use codify_core::infrastructure::providers::probe::LocalProviderProbe;
use codify_core::infrastructure::providers::remote::RemoteOpenAiCompatProvider;
use codify_core::infrastructure::repo::locale::HeuristicLocaleDetector;
use codify_core::infrastructure::repo::navigator::FsRepoNavigator;
use codify_core::infrastructure::repo::reference_resolver::FsHttpReferenceResolver;
use codify_core::infrastructure::repo::writer::FsArtifactWriter;
use codify_core::infrastructure::secrets::device_flow::DeviceFlow;
use codify_core::infrastructure::secrets::direct::DirectCredential;
use codify_core::infrastructure::secrets::keyring::SystemKeyring;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::{AppHandle, Emitter, State};

const DEFAULT_ENDPOINT: &str = "http://localhost:11434";
const DEFAULT_MODEL: &str = "qwen2.5-coder";

// ---------------------------------------------------------------------------
// Estado de la aplicación
// ---------------------------------------------------------------------------

/// Un servicio por sesión: el núcleo se cablea contra un repositorio concreto, así que no
/// puede existir un único servicio global.
/// Un desafío de conexión esperando a que el usuario termine.
///
/// Con nombres en vez de una tupla: `(Desafio, Arc<dyn AccountConnector>, Tier, String)` obliga
/// a recordar qué era el cuarto elemento cada vez que se lee.
struct DesafioPendiente {
    desafio: Desafio,
    conector: Arc<dyn AccountConnector>,
    tier: Tier,
    label: String,
    /// A dónde apunta la cuenta. Se guarda porque `complete_connection` lo necesita y no lo
    /// recibe: sin esto la conexión nacía con el host vacío y el proveedor apuntaba a
    /// `https://` (issue #48).
    endpoint: String,
}

#[derive(Default)]
pub struct AppState {
    services: Mutex<HashMap<String, Arc<ContextAuthoring>>>,
    /// Cuentas remotas conectadas (`003`-FR-003). No guardan el secreto: llevan la referencia
    /// con la que pedírselo al almacén del sistema.
    connections: Mutex<Vec<ProviderConnection>>,
    /// Desafíos de conexión a medio completar, por id.
    challenges: Mutex<HashMap<String, DesafioPendiente>>,
    /// El modo elegido por el usuario (`003`-FR-008a). Cambiarlo rearma el grafo de la
    /// **siguiente** sesión; la viva conserva el suyo (FR-008b).
    mode: Mutex<Mode>,
    /// Decisiones que el núcleo está esperando ahora mismo. Es el otro extremo del canal
    /// que `WindowPrompter::present` deja abierto: `decide` lo resuelve.
    pending: PendingDecisions,
}

impl AppState {
    fn remember(&self, id: &str, service: Arc<ContextAuthoring>) {
        if let Ok(mut map) = self.services.lock() {
            map.insert(id.to_string(), service);
        }
    }

    fn lookup(&self, id: &str) -> Option<Arc<ContextAuthoring>> {
        self.services.lock().ok()?.get(id).cloned()
    }

    pub fn pending(&self) -> PendingDecisions {
        self.pending.clone()
    }
}

// ---------------------------------------------------------------------------
// DTOs (la ventana nunca ve tipos de dominio)
// ---------------------------------------------------------------------------

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct StartSessionRequest {
    pub repo_root: String,
    /// `true` ⇒ modo local con cero-egress garantizado por construcción.
    pub local: bool,
    pub locale: Option<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct UnresolvedDto {
    pub origin: String,
    pub state: String,
}

/// Fragmento con su fundamento explícito. Es lo que permite a la UI distinguir sin
/// ambigüedad qué está verificado y qué no (FR-011 de `002-authoring-experience`).
#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SegmentDto {
    pub text: String,
    /// `grounded` | `tentative` | `contradiction`
    pub kind: String,
    pub sources: Vec<String>,
    pub reason: Option<String>,
    pub acknowledged: bool,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ArtifactDto {
    pub path: String,
    pub locale: String,
    pub segments: Vec<SegmentDto>,
    /// `written` | `skipped` | `failed` | `pending`.
    ///
    /// Sin esto, la vista no puede distinguir un archivo que ya está en el repositorio de uno
    /// que solo existe en pantalla — y dar por escrito lo que no lo está es justo la clase de
    /// afirmación sin respaldo que el producto se niega a hacer (FR-017).
    pub write_state: String,
}

/// Una propuesta tal y como la ve la ventana: el diff que se lee y el motivo que lo explica.
#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProposalDto {
    pub id: String,
    pub target: String,
    pub unified: String,
    pub rationale: String,
    /// `low` | `highimpact`. Solo el segundo bloquea.
    pub risk: String,
    pub applied: bool,
}

fn to_proposal_dto(p: &codify_core::domain::change::ChangeProposal) -> ProposalDto {
    use codify_core::domain::change::ChangeTarget;
    ProposalDto {
        id: p.id.as_str().to_string(),
        target: match &p.target {
            ChangeTarget::Artifact(k) => k.file_path().to_string(),
            ChangeTarget::RepoFile(path) => path.clone(),
        },
        unified: p.diff.unified.clone(),
        rationale: p.rationale.clone(),
        risk: p.risk.code().to_string(),
        applied: p.applied,
    }
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SessionSnapshotDto {
    pub id: String,
    pub state: String,
    pub locale: Option<String>,
    pub artifacts: Vec<ArtifactDto>,
    pub unresolved: Vec<UnresolvedDto>,
    pub omitted: Vec<String>,
    pub budget_exhausted: bool,
    pub interview_mode: bool,
    pub unattended_tentative: usize,
    /// Qué llegó (o no) al repositorio (FR-017).
    pub writes: Vec<WriteRecordDto>,
    /// Se generó con un tier inferior al pedido (`001`-FR-018). La interfaz lo declara.
    pub tier_degraded: bool,
    /// Por qué murió la sesión, como código (`002`-FR-028). La frase sale del catálogo: el
    /// núcleo no redacta, para que el motivo siga el idioma activo (SC-009).
    pub failure: Option<String>,
}

/// Las `quotes` que respaldan un segmento **no** se proyectan a la interfaz. No es un olvido:
/// tras FR-006a un segmento `grounded` ya está comprobado contra el material, así que lo que la
/// interfaz muestra hoy es más cierto que antes, no menos. Enseñar la evidencia al usuario es
/// una decisión de presentación —de `002`— con su clave de catálogo y su test de contrato.
fn to_segment_dto(segment: &codify_core::domain::context::Segment) -> SegmentDto {
    match &segment.groundedness {
        Groundedness::Grounded { sources, .. } => SegmentDto {
            text: segment.text.clone(),
            kind: "grounded".into(),
            sources: sources.clone(),
            reason: None,
            acknowledged: true,
        },
        Groundedness::Tentative {
            reason,
            acknowledged,
        } => SegmentDto {
            text: segment.text.clone(),
            kind: "tentative".into(),
            sources: Vec::new(),
            reason: Some(reason.clone()),
            acknowledged: *acknowledged,
        },
        Groundedness::Contradiction { sources, note, .. } => SegmentDto {
            text: segment.text.clone(),
            kind: "contradiction".into(),
            sources: sources.clone(),
            reason: Some(note.clone()),
            acknowledged: false,
        },
    }
}

fn to_artifact_dto(
    artifact: &ContextArtifact,
    writes: &[codify_core::domain::write::WriteRecord],
) -> ArtifactDto {
    use codify_core::domain::write::WriteOutcome;

    let path = artifact.kind.file_path().to_string();
    // La última escritura manda: una sesión puede reintentar sobre el mismo archivo.
    let write_state = writes
        .iter()
        .rev()
        .find(|w| w.path == path)
        .map(|w| match &w.outcome {
            WriteOutcome::Written => "written",
            WriteOutcome::Skipped(_) => "skipped",
            WriteOutcome::Failed(_) => "failed",
        })
        .unwrap_or("pending")
        .to_string();

    ArtifactDto {
        path,
        locale: artifact.locale.clone(),
        segments: artifact.segments.iter().map(to_segment_dto).collect(),
        write_state,
    }
}

fn to_snapshot_dto(snapshot: SessionSnapshot) -> SessionSnapshotDto {
    SessionSnapshotDto {
        id: snapshot.id.as_str().to_string(),
        state: snapshot.state.code().to_string(),
        locale: snapshot.locale,
        artifacts: snapshot
            .artifacts
            .iter()
            .map(|a| to_artifact_dto(a, &snapshot.writes))
            .collect(),
        unresolved: snapshot
            .unresolved
            .iter()
            .map(|u| UnresolvedDto {
                origin: u.origin.clone(),
                state: u.state.code().to_string(),
            })
            .collect(),
        omitted: snapshot.omitted,
        budget_exhausted: snapshot.budget_exhausted,
        interview_mode: snapshot.interview_mode,
        unattended_tentative: snapshot.unattended_tentative,
        writes: snapshot.writes.iter().map(to_write_dto).collect(),
        tier_degraded: snapshot.tier_degraded,
        failure: snapshot.failure.map(|f| f.code().to_string()),
    }
}

// ---------------------------------------------------------------------------
// Composition root de la piel
// ---------------------------------------------------------------------------

fn env_or(key: &str, fallback: &str) -> String {
    std::env::var(key).unwrap_or_else(|_| fallback.to_string())
}

/// Ensambla el núcleo para un repositorio y un modo concretos.
///
/// En modo local el resolver se construye **sin cliente HTTP** y el registro de proveedores
/// rechaza cualquier backend no local: la garantía de cero-egress es del cableado, no de un
/// flag que se consulte más tarde.
fn build_service(
    app: &AppHandle,
    repo_root: &str,
    mode: Mode,
    pending: PendingDecisions,
    remotos: Vec<Arc<dyn ModelProvider>>,
) -> Result<ContextAuthoring, String> {
    let provider = LocalOpenAiCompatProvider::new(
        "local",
        env_or("CODIFY_LOCAL_ENDPOINT", DEFAULT_ENDPOINT),
        env_or("CODIFY_LOCAL_MODEL", DEFAULT_MODEL),
    )
    .map_err(|e| format!("no se pudo preparar el proveedor local: {e}"))?;

    let resolver = if mode.is_local() {
        FsHttpReferenceResolver::local_only(repo_root)
    } else {
        FsHttpReferenceResolver::with_public_web(repo_root)
    };

    // El modo vive en el TIPO del builder (`003`-FR-008a), así que aquí hay que ramificar. La
    // rama no es ceremonia: es el punto exacto donde los dos grafos dejan de ser el mismo, y
    // tenerlo visible es preferible a un `mode` que viaja como dato y se consulta más tarde.
    let deps = match mode {
        // En local los remotos **no se cablean**, y no porque se filtren: el builder de este
        // lado no tiene el método (`003`-FR-008). Si alguien los pasara, no habría dónde
        // ponerlos.
        Mode::Local => cablear(
            CoreBuilder::<Local>::new(),
            app,
            repo_root,
            pending,
            provider,
            resolver,
        )?,
        Mode::Hybrid => {
            let mut b = CoreBuilder::<Hybrid>::new();
            for r in remotos {
                b = b.remote_provider(r);
            }
            cablear(b, app, repo_root, pending, provider, resolver)?
        }
    };

    Ok(ContextAuthoring::new(deps))
}

/// Cablea todo lo que **no** depende del modo.
///
/// Es genérica sobre el estado del grafo para que exista una sola copia de esta lista: si el
/// cableado común se duplicara por rama, olvidar un adapter en una de las dos sería un error
/// silencioso que solo aparecería en ese modo.
#[allow(clippy::too_many_arguments)]
fn cablear<M: ModoDelGrafo>(
    builder: CoreBuilder<M>,
    app: &AppHandle,
    repo_root: &str,
    pending: PendingDecisions,
    provider: LocalOpenAiCompatProvider,
    resolver: FsHttpReferenceResolver,
) -> Result<codify_core::application::deps::AuthoringDeps, String> {
    builder
        .provider(Arc::new(provider))
        .navigator(Arc::new(FsRepoNavigator::new(repo_root)))
        .resolver(Arc::new(resolver))
        .diff(Arc::new(SimilarDiffEngine))
        .risk(Arc::new(ConservativeRiskClassifier))
        .prompter(Arc::new(WindowPrompter::new(app.clone(), pending)))
        .audit(Arc::new(EventAuditSink::new(app.clone())))
        .locale(Arc::new(HeuristicLocaleDetector::new(String::new())))
        .clock(Arc::new(SystemClock))
        .writer(Arc::new(FsArtifactWriter::new(repo_root)))
        .discovery(Arc::new(
            LocalProviderProbe::new(env_or("CODIFY_LOCAL_ENDPOINT", DEFAULT_ENDPOINT))
                .map_err(|e| format!("no se pudo preparar la sonda: {e}"))?,
        ))
        .cancellations(Arc::new(TokenCancellationFactory))
        .build()
        .map_err(|e| format!("no se pudo cablear el núcleo: {e}"))
}

/// Conexión a un proveedor remoto, tal y como la ve la interfaz (`003`-contracts).
///
/// **No tiene campo para el secreto, ni podría tenerlo**: `ProviderConnection` tampoco. Es más
/// fiable que acordarse de no rellenarlo.
#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProviderConnectionDto {
    pub id: String,
    pub label: String,
    /// Solo el host: FR-009 necesita decir **quién** podría recibir contenido, no cómo se llega.
    pub endpoint_host: String,
    pub tier: String,
    pub state: String,
}

fn to_connection_dto(c: &ProviderConnection) -> ProviderConnectionDto {
    ProviderConnectionDto {
        id: c.id.clone(),
        label: c.label.clone(),
        endpoint_host: c.endpoint_host.clone(),
        tier: match c.tier {
            Tier::Cheap => "cheap".into(),
            Tier::Heavy => "heavy".into(),
        },
        state: c.state.code().to_string(),
    }
}

// ---------------------------------------------------------------------------
// Comandos
// ---------------------------------------------------------------------------

/// `003`-FR-001. Devuelve el desafío: código y dirección, o la petición de credencial.
#[tauri::command]
pub async fn connect_provider(
    state: State<'_, AppState>,
    label: String,
    endpoint: String,
    tier: String,
    delegada: bool,
) -> Result<ConnectChallengeDto, String> {
    let tier = match tier.as_str() {
        "heavy" => Tier::Heavy,
        _ => Tier::Cheap,
    };
    let conector: Arc<dyn AccountConnector> = if delegada {
        Arc::new(
            DeviceFlow::new(
                format!("{endpoint}/device/code"),
                format!("{endpoint}/token"),
                env_or("CODIFY_OAUTH_CLIENT_ID", "codify"),
            )
            .map_err(|e| e.to_string())?,
        )
    } else {
        Arc::new(DirectCredential::new(
            "Pega tu credencial: se guarda en el almacén del sistema y no vuelve a mostrarse.",
        ))
    };

    let desafio = conector.iniciar().await.map_err(|e| e.to_string())?;
    let id = format!("conn-{}", uuid_simple());
    let dto = ConnectChallengeDto::from(&id, &desafio);
    if let Ok(mut m) = state.challenges.lock() {
        m.insert(
            id,
            DesafioPendiente {
                desafio,
                conector,
                tier,
                label,
                endpoint: solo_host(&endpoint),
            },
        );
    }
    Ok(dto)
}

/// `003`-FR-001/FR-002. Guarda en el almacén del sistema y devuelve la conexión **sin** secreto.
#[tauri::command]
pub async fn complete_connection(
    state: State<'_, AppState>,
    challenge_id: String,
    secret: Option<String>,
) -> Result<ProviderConnectionDto, String> {
    let DesafioPendiente {
        desafio,
        conector,
        tier,
        label,
        endpoint,
    } = state
        .challenges
        .lock()
        .ok()
        .and_then(|mut m| m.remove(&challenge_id))
        .ok_or_else(|| "ese desafío ya no está en curso".to_string())?;

    let store = SystemKeyring::new();
    if !store.disponible() {
        // FR-004: se dice, y NO se recurre a otro sitio.
        return Err("no hay almacén de credenciales disponible en este sistema".into());
    }

    let secreto = conector
        .completar(&desafio, secret.map(Secreto::new))
        .await
        .map_err(|e| e.to_string())?;

    let referencia = ReferenciaDeCredencial::new(challenge_id.clone());
    store
        .guardar(&referencia, secreto)
        .await
        .map_err(|e| e.to_string())?;

    let conexion = conexion_desde(&challenge_id, &label, &endpoint, tier);
    let dto = to_connection_dto(&conexion);
    if let Ok(mut c) = state.connections.lock() {
        c.push(conexion);
    }
    Ok(dto)
}

/// `003`-FR-003.
#[tauri::command]
pub async fn list_connections(
    state: State<'_, AppState>,
) -> Result<Vec<ProviderConnectionDto>, String> {
    Ok(state
        .connections
        .lock()
        .map(|c| c.iter().map(to_connection_dto).collect())
        .unwrap_or_default())
}

/// `003`-FR-003 y SC-006: borra del almacén y quita la conexión, **sin reiniciar**.
#[tauri::command]
pub async fn disconnect_provider(
    state: State<'_, AppState>,
    connection_id: String,
) -> Result<(), String> {
    let referencia = ReferenciaDeCredencial::new(connection_id.clone());
    SystemKeyring::new()
        .borrar(&referencia)
        .await
        .map_err(|e| e.to_string())?;
    if let Ok(mut c) = state.connections.lock() {
        c.retain(|x| x.id != connection_id);
    }
    Ok(())
}

/// `003`-FR-008a. El modo se guarda y el grafo se rearma en la **siguiente** sesión: la viva
/// conserva el suyo (FR-008b), y por eso este comando no toca ninguna.
#[tauri::command]
pub async fn set_mode(state: State<'_, AppState>, local: bool) -> Result<(), String> {
    if let Ok(mut m) = state.mode.lock() {
        *m = if local { Mode::Local } else { Mode::Hybrid };
    }
    Ok(())
}

/// Convierte las cuentas conectadas en proveedores, pidiendo cada credencial al almacén.
///
/// Una conexión revocada o caducada **no** produce proveedor: SC-006 exige que desconectar surta
/// efecto en la tarea siguiente, y filtrar aquí es lo que lo consigue sin reiniciar.
async fn adapters_remotos(state: &State<'_, AppState>) -> Vec<Arc<dyn ModelProvider>> {
    let conexiones: Vec<ProviderConnection> = state
        .connections
        .lock()
        .map(|c| c.iter().filter(|x| x.usable()).cloned().collect())
        .unwrap_or_default();

    let store = SystemKeyring::new();
    let mut out: Vec<Arc<dyn ModelProvider>> = Vec::new();
    let mut sin_credencial: Vec<String> = Vec::new();
    for c in conexiones {
        // Sin credencial no hay proveedor, **y se dice**: la conexión pasa a
        // `CredentialMissing` para que el panel lo muestre. Antes se omitía en silencio con el
        // argumento de que el panel daría el motivo, y el panel no tenía cómo (issue #48).
        let secreto = store.obtener(c.credential()).await.ok().flatten();
        let Some(secreto) = secreto else {
            sin_credencial.push(c.id.clone());
            continue;
        };
        {
            if let Ok(p) = RemoteOpenAiCompatProvider::new(
                c.label.clone(),
                format!("https://{}", c.endpoint_host),
                env_or("CODIFY_REMOTE_MODEL", "default"),
                c.tier,
                secreto,
            ) {
                out.push(Arc::new(p));
            }
        }
    }

    // Marcar lo que ya no tiene credencial. Se hace después del bucle para no tener el candado
    // de conexiones abierto mientras se habla con el llavero.
    if !sin_credencial.is_empty() {
        if let Ok(mut conns) = state.connections.lock() {
            for c in conns.iter_mut() {
                if sin_credencial.contains(&c.id) {
                    c.state = ConnectionState::CredentialMissing;
                }
            }
        }
    }
    out
}

/// Arma la conexión a partir de lo que el desafío guardó.
///
/// Extraída del comando para que **el cableado se pueda probar**. El defecto de #48 vivía justo
/// aquí —el endpoint llegaba y se tiraba— y no lo veía ningún test porque los que había probaban
/// las piezas por separado: `solo_host` por un lado, `ProviderConnection::new` por otro. Entre
/// las dos no miraba nadie.
fn conexion_desde(id: &str, label: &str, endpoint_host: &str, tier: Tier) -> ProviderConnection {
    ProviderConnection::new(
        id,
        label,
        endpoint_host,
        tier,
        ReferenciaDeCredencial::new(id),
    )
}

/// Solo el host de una URL. FR-009 pide decir **quién** podría recibir contenido, no cómo se
/// llega — y una URL puede llevar credenciales embebidas.
fn solo_host(url: &str) -> String {
    url.trim()
        .trim_start_matches("https://")
        .trim_start_matches("http://")
        .split('/')
        .next()
        .unwrap_or("")
        .to_string()
}

/// Id corto y suficiente: solo tiene que ser único dentro de esta ejecución.
fn uuid_simple() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};
    format!(
        "{:x}",
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_nanos())
            .unwrap_or(0)
    )
}

/// El desafío, tal y como lo ve la interfaz.
#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ConnectChallengeDto {
    pub challenge_id: String,
    /// `"delegada"` o `"credencial"`: la interfaz elige qué enseñar.
    pub kind: String,
    pub code: Option<String>,
    pub url: Option<String>,
    pub instructions: Option<String>,
}

impl ConnectChallengeDto {
    fn from(id: &str, d: &Desafio) -> Self {
        match d {
            Desafio::Delegada { codigo, url } => Self {
                challenge_id: id.into(),
                kind: "delegada".into(),
                code: Some(codigo.clone()),
                url: Some(url.clone()),
                instructions: None,
            },
            Desafio::PideCredencial { instrucciones } => Self {
                challenge_id: id.into(),
                kind: "credencial".into(),
                code: None,
                url: None,
                instructions: Some(instrucciones.clone()),
            },
        }
    }
}

#[tauri::command]
pub async fn start_session(
    app: AppHandle,
    state: State<'_, AppState>,
    request: StartSessionRequest,
) -> Result<String, String> {
    let mode = if request.local {
        Mode::Local
    } else {
        Mode::Hybrid
    };
    let _ = app.emit(
        "session.state_changed",
        StatePayload {
            state: "ingesting".into(),
        },
    );

    // Las conexiones guardadas se convierten en proveedores **solo** si el modo las admite.
    // El `for` de dentro de `build_service` no puede colarlas en un grafo local: allí el método
    // no existe.
    let remotos = adapters_remotos(&state).await;
    let service = Arc::new(build_service(
        &app,
        &request.repo_root,
        mode,
        state.pending(),
        remotos,
    )?);

    let id = service
        .start_session(StartSession {
            repo_root: request.repo_root.into(),
            mode,
            locale: request.locale,
        })
        .await
        .map_err(|e| e.to_string())?;

    state.remember(id.as_str(), service);
    let _ = app.emit(
        "session.state_changed",
        StatePayload {
            state: "generating".into(),
        },
    );
    Ok(id.as_str().to_string())
}

#[tauri::command]
pub async fn session_state(
    state: State<'_, AppState>,
    session_id: String,
) -> Result<SessionSnapshotDto, String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    let snapshot = service
        .session_state(&SessionId::new(session_id))
        .await
        .map_err(|e| e.to_string())?;

    Ok(to_snapshot_dto(snapshot))
}

#[tauri::command]
pub async fn set_locale(
    state: State<'_, AppState>,
    session_id: String,
    locale: String,
) -> Result<(), String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    service
        .set_locale(&SessionId::new(session_id), locale)
        .await
        .map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use codify_core::domain::context::{ArtifactKind, Segment};

    fn tres_estados() -> ContextArtifact {
        ContextArtifact::new(ArtifactKind::Context, "es").with_segments(vec![
            Segment::grounded(
                "Motor: Temporal",
                vec!["SPEC-30.md".into()],
                vec!["el motor es Temporal".into()],
            ),
            Segment::tentative("Métricas por definir", "sin fuente"),
            Segment::contradiction(
                "Persistencia",
                vec!["PRD".into(), "SPEC".into()],
                vec!["Postgres".into(), "event-sourced".into()],
                "chocan",
            ),
        ])
    }

    #[test]
    fn maps_each_groundedness_to_its_own_kind() {
        let artifact = tres_estados();

        let dto = to_artifact_dto(&artifact, &[]);
        assert_eq!(dto.path, "context/CONTEXT.md");
        assert_eq!(dto.segments[0].kind, "grounded");
        assert_eq!(dto.segments[0].sources, vec!["SPEC-30.md"]);
        assert_eq!(dto.segments[1].kind, "tentative");
        assert_eq!(dto.segments[1].reason.as_deref(), Some("sin fuente"));
        assert!(
            !dto.segments[1].acknowledged,
            "lo tentativo nace sin atender"
        );
        assert_eq!(dto.segments[2].kind, "contradiction");
        assert_eq!(dto.segments[2].sources.len(), 2);
    }

    /// **T033** — la vista de artefacto necesita saber si lo que muestra **está en el
    /// repositorio**. Sin esto no puede distinguir un archivo escrito de uno que solo existe
    /// en pantalla, y presentarlos igual sería afirmar algo que no consta.
    #[test]
    fn the_artifact_declares_whether_it_reached_the_repository() {
        use codify_core::domain::write::{WriteOutcome, WriteRecord};

        let artifact = tres_estados();
        let path = artifact.kind.file_path().to_string();

        let sin_escribir = to_artifact_dto(&artifact, &[]);
        assert_eq!(
            sin_escribir.write_state, "pending",
            "sin registro de escritura, el artefacto NO puede darse por escrito"
        );

        let escrito = to_artifact_dto(
            &artifact,
            &[WriteRecord {
                path: path.clone(),
                bytes: 120,
                at: "2026-01-01T00:00:00Z".into(),
                outcome: WriteOutcome::Written,
            }],
        );
        assert_eq!(escrito.write_state, "written");

        let fallido = to_artifact_dto(
            &artifact,
            &[WriteRecord {
                path: path.clone(),
                bytes: 0,
                at: "2026-01-01T00:00:00Z".into(),
                outcome: WriteOutcome::Failed("permiso denegado".into()),
            }],
        );
        assert_eq!(
            fallido.write_state, "failed",
            "un fallo de escritura tiene que verse: silenciarlo haría creer que el archivo está"
        );

        // Un registro de OTRO archivo no puede teñir a este.
        let ajeno = to_artifact_dto(
            &artifact,
            &[WriteRecord {
                path: "otro/ARCHIVO.md".into(),
                bytes: 10,
                at: "2026-01-01T00:00:00Z".into(),
                outcome: WriteOutcome::Written,
            }],
        );
        assert_eq!(ajeno.write_state, "pending");
    }

    /// Una sesión puede reintentar sobre el mismo archivo: manda el **último** resultado.
    #[test]
    fn the_latest_write_wins() {
        use codify_core::domain::write::{WriteOutcome, WriteRecord};

        let artifact = tres_estados();
        let path = artifact.kind.file_path().to_string();
        let record = |outcome| WriteRecord {
            path: path.clone(),
            bytes: 1,
            at: "2026-01-01T00:00:00Z".into(),
            outcome,
        };

        let dto = to_artifact_dto(
            &artifact,
            &[
                record(WriteOutcome::Failed("disco lleno".into())),
                record(WriteOutcome::Written),
            ],
        );
        assert_eq!(
            dto.write_state, "written",
            "el reintento que funcionó manda"
        );
    }

    #[test]
    fn the_three_verdicts_are_understood() {
        assert!(matches!(
            parse_verdict("approve", None).unwrap(),
            Verdict::Approve
        ));
        assert!(matches!(
            parse_verdict("reject", None).unwrap(),
            Verdict::Reject
        ));
        match parse_verdict("edit", Some("mi texto".into())).unwrap() {
            Verdict::Edit(t) => assert_eq!(t, "mi texto"),
            otro => panic!("se esperaba una edición, llegó {otro:?}"),
        }
    }

    /// Editar sin texto, o con texto vacío, **no** puede degradar a aprobar: se aplicaría el
    /// contenido del agente haciendo creer al usuario que aplicó el suyo.
    #[test]
    fn editing_without_text_is_refused_not_downgraded() {
        assert!(parse_verdict("edit", None).is_err());
        assert!(parse_verdict("edit", Some("   ".into())).is_err());
    }

    /// Un veredicto desconocido falla. Interpretarlo como rechazo sería seguro pero silencioso;
    /// como aprobación, escribiría sin permiso.
    #[test]
    fn an_unknown_verdict_is_refused() {
        assert!(parse_verdict("quizá", None).is_err());
        assert!(parse_verdict("", None).is_err());
    }

    #[test]
    fn a_local_session_cannot_be_wired_against_a_remote_endpoint() {
        // El proveedor local rechaza endpoints no loopback en su constructor.
        assert!(LocalOpenAiCompatProvider::new("x", "https://api.remoto.test", "m").is_err());
    }
}

// ---------------------------------------------------------------------------
// Comandos de la Fase 3 (spec 002): cancelación, sonda, cadenas e idioma
// ---------------------------------------------------------------------------

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct WriteRecordDto {
    pub path: String,
    pub bytes: usize,
    /// `written` | `skipped` | `failed`
    pub outcome: String,
    /// Motivo cuando no llegó al disco. Nunca vacío si `outcome != "written"`.
    pub detail: Option<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct CancelOutcomeDto {
    pub session_id: String,
    /// Fase en la que se cortó: el usuario sabe hasta dónde llegó.
    pub phase: String,
    pub writes: Vec<WriteRecordDto>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProviderStatusDto {
    pub reachable: bool,
    pub endpoint: String,
    pub models: Vec<String>,
    /// Motivo como **código**, no como frase: la interfaz elige el texto en su idioma
    /// (`provider.issue.<code>`). Es lo que separa "guiado" de "silencioso" (FR-019) sin
    /// que el núcleo tenga que redactar en un idioma concreto (SC-009).
    pub issue: Option<String>,
}

fn to_write_dto(record: &codify_core::domain::write::WriteRecord) -> WriteRecordDto {
    use codify_core::domain::write::WriteOutcome;
    let (outcome, detail) = match &record.outcome {
        WriteOutcome::Written => ("written", None),
        WriteOutcome::Skipped(why) => ("skipped", Some(why.clone())),
        WriteOutcome::Failed(why) => ("failed", Some(why.clone())),
    };
    WriteRecordDto {
        path: record.path.clone(),
        bytes: record.bytes,
        outcome: outcome.into(),
        detail,
    }
}

/// Cancela la sesión y devuelve el balance de lo que alcanzó a escribirse (FR-023).
#[tauri::command]
pub async fn cancel_session(
    state: State<'_, AppState>,
    session_id: String,
) -> Result<CancelOutcomeDto, String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    let outcome = service
        .cancel_session(&SessionId::new(session_id))
        .await
        .map_err(|e| e.to_string())?;

    Ok(CancelOutcomeDto {
        session_id: outcome.session_id.as_str().to_string(),
        phase: outcome.phase.code().to_string(),
        writes: outcome.writes.iter().map(to_write_dto).collect(),
    })
}

/// Sondea el backend de modelo. No falla: informa con un motivo accionable (FR-019/FR-028).
#[tauri::command]
pub async fn probe_provider(local: bool) -> Result<ProviderStatusDto, String> {
    let endpoint = env_or("CODIFY_LOCAL_ENDPOINT", DEFAULT_ENDPOINT);

    let probe = match LocalProviderProbe::new(&endpoint) {
        Ok(p) => p,
        // Un endpoint no loopback en modo local no es un fallo opaco: es algo que explicar.
        Err(_) => {
            return Ok(ProviderStatusDto {
                reachable: false,
                endpoint,
                models: Vec::new(),
                issue: Some(ProviderIssue::EndpointNotLocal.code().to_string()),
            })
        }
    };

    let status = probe.probe().await;
    let _ = local; // el modo condiciona el cableado, no la forma de sondear
    Ok(ProviderStatusDto {
        reachable: status.reachable,
        endpoint: status.endpoint,
        models: status.models,
        issue: status.issue.map(|i| i.code().to_string()),
    })
}

/// Catálogo de cadenas del idioma pedido (FR-016b).
#[tauri::command]
pub fn ui_strings(locale: String) -> crate::strings::UiStrings {
    crate::strings::strings_for(crate::strings::Locale::parse(&locale))
}

/// Idioma del sistema, con caída a inglés (FR-016b).
#[tauri::command]
pub fn system_locale() -> String {
    crate::strings::system_locale().code().to_string()
}

/// Un artefacto completo, alcanzable en cualquier momento (FR-021).
#[tauri::command]
pub async fn artifact(
    state: State<'_, AppState>,
    session_id: String,
    path: String,
) -> Result<ArtifactDto, String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    let snapshot = service
        .session_state(&SessionId::new(session_id))
        .await
        .map_err(|e| e.to_string())?;

    snapshot
        .artifacts
        .iter()
        .find(|a| a.kind.file_path() == path)
        .map(|a| to_artifact_dto(a, &snapshot.writes))
        .ok_or_else(|| format!("artefacto desconocido: {path}"))
}

/// Un turno de conversación: el usuario escribe, el agente propone.
///
/// **Retorna cuando el turno está resuelto**, no antes. Mientras tanto la ventana recibe un
/// `proposal.new` por cada cambio de alto impacto y responde con `decide`; los de bajo riesgo
/// ni aparecen aquí porque ya se aplicaron (FR-010).
#[tauri::command]
pub async fn submit_message(
    state: State<'_, AppState>,
    session_id: String,
    message: String,
) -> Result<Vec<ProposalDto>, String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    let proposals = service
        .submit_message(&SessionId::new(session_id), &message)
        .await
        .map_err(|e| e.to_string())?;

    Ok(proposals.iter().map(to_proposal_dto).collect())
}

/// Las propuestas que el núcleo está esperando decidir **ahora mismo**.
///
/// Se lee del registro de canales abiertos, no de la sesión: es la única fuente que sabe qué
/// está realmente bloqueado esperando a una persona.
#[tauri::command]
pub fn pending_proposals(state: State<'_, AppState>) -> Result<Vec<String>, String> {
    Ok(state
        .pending
        .lock()
        .map_err(|_| "el registro de decisiones se corrompió".to_string())?
        .keys()
        .cloned()
        .collect())
}

/// Registra la decisión del usuario y **desbloquea** al núcleo (FR-014/FR-015).
#[tauri::command]
pub fn decide(
    state: State<'_, AppState>,
    proposal_id: String,
    verdict: String,
    edited: Option<String>,
) -> Result<(), String> {
    let sender = state
        .pending
        .lock()
        .map_err(|_| "el registro de decisiones se corrompió".to_string())?
        .remove(&proposal_id)
        .ok_or_else(|| format!("nadie está esperando una decisión sobre {proposal_id}"))?;

    let verdict = parse_verdict(&verdict, edited)?;

    // Si el receptor ya no está, el núcleo dejó de esperar (cancelación o cierre). Decirlo es
    // mejor que dar por registrada una decisión que nadie recibió.
    sender
        .send(ApprovalDecision {
            proposal_id: ProposalId::new(proposal_id),
            verdict,
            actor: "usuario".into(),
            at: codify_core::domain::ports::Clock::now_iso(&SystemClock),
        })
        .map_err(|_| "el núcleo dejó de esperar esta decisión".to_string())
}

/// Deshace un cambio auto-aplicado por bajo riesgo (FR-008).
///
/// Es la compensación de no haber preguntado: se aplicó algo sin consultar al usuario, así que
/// tiene que poder devolverlo. El núcleo rechaza deshacer lo que pasó por una decisión humana.
#[tauri::command]
pub async fn revert_proposal(
    state: State<'_, AppState>,
    session_id: String,
    proposal_id: String,
) -> Result<(), String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    service
        .revert_proposal(&SessionId::new(session_id), &ProposalId::new(proposal_id))
        .await
        .map_err(|e| e.to_string())
}

/// Traduce el veredicto que llega de la ventana.
///
/// Un veredicto que no se reconoce **falla**: interpretarlo como "rechazar" sería seguro pero
/// silencioso, y como "aprobar" escribiría sin permiso. Ninguna de las dos es aceptable ante
/// una entrada que no se entiende.
fn parse_verdict(verdict: &str, edited: Option<String>) -> Result<Verdict, String> {
    match verdict {
        "approve" => Ok(Verdict::Approve),
        "reject" => Ok(Verdict::Reject),
        // Editar SIN texto no es editar: aplicar el del agente convertiría "editar" en
        // "aprobar con pasos de más".
        "edit" => edited
            .filter(|t| !t.trim().is_empty())
            .map(Verdict::Edit)
            .ok_or_else(|| "editar exige el texto del usuario".to_string()),
        otro => Err(format!("veredicto desconocido: {otro}")),
    }
}

/// Difiere un fragmento tentativo concreto y devuelve cuántos quedan sin atender (FR-014).
///
/// Es por fragmento: no existe un «diferir todo». Lo que no está verificado se difiere
/// mirándolo, que es la diferencia entre decidir y despachar.
#[tauri::command]
pub async fn defer_tentative(
    state: State<'_, AppState>,
    session_id: String,
    path: String,
    index: usize,
) -> Result<usize, String> {
    let service = state
        .lookup(&session_id)
        .ok_or_else(|| format!("sesión desconocida: {session_id}"))?;

    service
        .defer_tentative(&SessionId::new(session_id), &path, index)
        .await
        .map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests_conexion {
    use super::*;

    /// El camino que ningún test recorría (issue #48).
    ///
    /// Los tests del núcleo prueban `RemoteOpenAiCompatProvider` con una URL explícita — justo el
    /// trozo que sí funcionaba. El defecto vivía en el cableado: el endpoint llegaba a
    /// `connect_provider`, se usaba para armar las URLs del device-flow y se tiraba, así que la
    /// conexión nacía con el host vacío y el proveedor apuntaba a `https://`.
    ///
    /// La lógica probada y el cableado sin probar es la clase de fallo que ha aparecido en las
    /// cinco fases de este proyecto. Este test cierra el hueco en el camino de US1.
    #[test]
    fn el_endpoint_sobrevive_de_la_peticion_a_la_conexion() {
        // El camino real: lo que `connect_provider` guarda es lo que `complete_connection` usa.
        let guardado = solo_host("https://api.example.com/v1/");
        let conexion = conexion_desde("conn-1", "Frontier", &guardado, Tier::Heavy);

        assert!(
            !conexion.endpoint_host.is_empty(),
            "con el host vacío el proveedor apuntaría a `https://` y no podría llamar a nadie"
        );
        assert_eq!(
            format!("https://{}", conexion.endpoint_host),
            "https://api.example.com",
            "es la URL que `adapters_remotos` construye"
        );
    }

    /// FR-009 pide decir **quién** podría recibir contenido, no cómo se llega — y una URL puede
    /// llevar credenciales embebidas.
    #[test]
    fn el_host_no_arrastra_esquema_ruta_ni_credenciales() {
        assert_eq!(
            solo_host("https://usuario:clave@api.example.com/v1"),
            "usuario:clave@api.example.com"
        );
        assert_eq!(solo_host("http://localhost:8080/x"), "localhost:8080");
        assert_eq!(solo_host("  api.example.com  "), "api.example.com");
    }

    /// Una conexión sin credencial **no** se muestra conectada.
    ///
    /// Antes se omitía en silencio al armar el grafo, con el argumento de que el panel daría el
    /// motivo. El panel no tenía cómo: no existía el estado.
    #[test]
    fn una_conexion_sin_credencial_deja_de_ser_usable() {
        let mut c = ProviderConnection::new(
            "conn-1",
            "Frontier",
            "api.example.com",
            Tier::Heavy,
            ReferenciaDeCredencial::new("conn-1"),
        );
        assert!(c.usable());

        c.state = ConnectionState::CredentialMissing;

        assert!(
            !c.usable(),
            "si no se puede usar, no puede seguir diciendo que está conectada"
        );
        assert_eq!(to_connection_dto(&c).state, "credential_missing");
    }
}
