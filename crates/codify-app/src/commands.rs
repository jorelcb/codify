//! Comandos Tauri (T029) — el **driving adapter** de la piel.
//!
//! Superficie definida en `specs/001-context-authoring/contracts/tauri-commands.md`.
//! Cada comando delega en el Application Service del núcleo y traduce sus tipos a DTOs
//! serializables: los tipos de dominio no cruzan hacia la ventana.

use crate::adapters::{EventAuditSink, StatePayload, SystemClock, UnavailablePrompter};
use codify_core::application::ports::{ProviderDiscovery, ProviderIssue};
use codify_core::application::service::{
    AuthoringService, ContextAuthoring, SessionSnapshot, StartSession,
};
use codify_core::domain::context::{ContextArtifact, Groundedness};
use codify_core::domain::session::{Mode, SessionId};
use codify_core::infrastructure::cancel::TokenCancellationFactory;
use codify_core::infrastructure::composition::CoreBuilder;
use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
use codify_core::infrastructure::providers::probe::LocalProviderProbe;
use codify_core::infrastructure::repo::locale::HeuristicLocaleDetector;
use codify_core::infrastructure::repo::navigator::FsRepoNavigator;
use codify_core::infrastructure::repo::reference_resolver::FsHttpReferenceResolver;
use codify_core::infrastructure::repo::writer::FsArtifactWriter;
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
#[derive(Default)]
pub struct AppState {
    services: Mutex<HashMap<String, Arc<ContextAuthoring>>>,
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
}

fn to_segment_dto(segment: &codify_core::domain::context::Segment) -> SegmentDto {
    match &segment.groundedness {
        Groundedness::Grounded { sources } => SegmentDto {
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
        Groundedness::Contradiction { sources, note } => SegmentDto {
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
        state: format!("{:?}", snapshot.state).to_lowercase(),
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
                state: format!("{:?}", u.state).to_lowercase(),
            })
            .collect(),
        omitted: snapshot.omitted,
        budget_exhausted: snapshot.budget_exhausted,
        interview_mode: snapshot.interview_mode,
        unattended_tentative: snapshot.unattended_tentative,
        writes: snapshot.writes.iter().map(to_write_dto).collect(),
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
fn build_service(app: &AppHandle, repo_root: &str, mode: Mode) -> Result<ContextAuthoring, String> {
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

    let deps = CoreBuilder::new(mode)
        .provider(Arc::new(provider))
        .navigator(Arc::new(FsRepoNavigator::new(repo_root)))
        .resolver(Arc::new(resolver))
        .diff(Arc::new(NoDiffYet))
        .risk(Arc::new(ConservativeRisk))
        .prompter(Arc::new(UnavailablePrompter))
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
        .map_err(|e| format!("no se pudo cablear el núcleo: {e}"))?;

    Ok(ContextAuthoring::new(deps))
}

/// El motor de diffs entra en US2. Declararlo ausente es más honesto que cablear uno que
/// nadie ejercita todavía.
struct NoDiffYet;

impl codify_core::application::ports::DiffEngine for NoDiffYet {
    fn make(&self, before: &str, after: &str) -> codify_core::domain::change::Diff {
        codify_core::domain::change::Diff {
            unified: String::new(),
            before: before.into(),
            after: after.into(),
        }
    }
    fn apply(
        &self,
        _before: &str,
        diff: &codify_core::domain::change::Diff,
    ) -> codify_core::domain::error::Result<String> {
        Ok(diff.after.clone())
    }
    fn revert(
        &self,
        _after: &str,
        diff: &codify_core::domain::change::Diff,
    ) -> codify_core::domain::error::Result<String> {
        Ok(diff.before.clone())
    }
}

/// Política conservadora: mientras no exista el criterio afinado (spec derivado de FR-012),
/// todo cambio no trivial exige aprobación.
struct ConservativeRisk;

impl codify_core::domain::ports::RiskClassifier for ConservativeRisk {
    fn classify(
        &self,
        proposal: &codify_core::domain::change::ChangeProposal,
    ) -> codify_core::domain::change::RiskLevel {
        if proposal.diff.is_empty() {
            codify_core::domain::change::RiskLevel::Low
        } else {
            codify_core::domain::change::RiskLevel::HighImpact
        }
    }
}

// ---------------------------------------------------------------------------
// Comandos
// ---------------------------------------------------------------------------

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

    let service = Arc::new(build_service(&app, &request.repo_root, mode)?);

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
            Segment::grounded("Motor: Temporal", vec!["SPEC-30.md".into()]),
            Segment::tentative("Métricas por definir", "sin fuente"),
            Segment::contradiction("Persistencia", vec!["PRD".into(), "SPEC".into()], "chocan"),
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
        phase: format!("{:?}", outcome.phase).to_lowercase(),
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
