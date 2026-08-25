//! Comandos Tauri (T029) — el **driving adapter** de la piel.
//!
//! Superficie definida en `specs/001-context-authoring/contracts/tauri-commands.md`.
//! Cada comando delega en el Application Service del núcleo y traduce sus tipos a DTOs
//! serializables: los tipos de dominio no cruzan hacia la ventana.

use crate::adapters::{
    EventAuditSink, PendingDecisions, StatePayload, SystemClock, WindowPrompter,
};
use codify_core::application::ports::{ProviderDiscovery, ProviderIssue};
use codify_core::application::service::{
    AuthoringService, ContextAuthoring, SessionSnapshot, StartSession,
};
use codify_core::domain::change::{ApprovalDecision, ProposalId, Verdict};
use codify_core::domain::context::{ContextArtifact, Groundedness};
use codify_core::domain::session::{Mode, SessionId};
use codify_core::infrastructure::cancel::TokenCancellationFactory;
use codify_core::infrastructure::composition::CoreBuilder;
use codify_core::infrastructure::diff::engine::SimilarDiffEngine;
use codify_core::infrastructure::diff::risk::ConservativeRiskClassifier;
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

    let deps = CoreBuilder::new(mode)
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
        .map_err(|e| format!("no se pudo cablear el núcleo: {e}"))?;

    Ok(ContextAuthoring::new(deps))
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

    let service = Arc::new(build_service(
        &app,
        &request.repo_root,
        mode,
        state.pending(),
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
