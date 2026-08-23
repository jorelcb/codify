//! Application Service: el punto de entrada del caso de uso.
//!
//! Es lo que consumen las pieles (Tauri hoy; MCP/CLI mañana). El trait existe porque hay
//! **más de un adaptador primario** previsto — la única razón que justifica una interfaz
//! driving (constitución, Principio I).
//!
//! Nombre sin decoración: `ContextAuthoring` nombra la capacidad, no el patrón.

use crate::application::authoring_loop::{AuthoringLoop, GatheredSource, IngestOutcome};
use crate::application::deps::AuthoringDeps;
use crate::application::ports::Cancellation;
use crate::domain::audit::{AuditEvent, AuditKind};
use crate::domain::change::{ApprovalDecision, ChangeProposal, ChangeTarget, ProposalId, Verdict};
use crate::domain::context::ContextArtifact;
use crate::domain::error::{CoreError, Result};
use crate::domain::reference::ReferenceState;
use crate::domain::session::{AuthoringSession, Mode, SessionId, SessionState};
use crate::domain::write::WriteRecord;
use async_trait::async_trait;
use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use tokio::sync::Mutex;
use tokio::task::JoinHandle;

#[derive(Debug, Clone)]
pub struct StartSession {
    pub repo_root: PathBuf,
    pub mode: Mode,
    /// Override explícito del idioma; `None` ⇒ auto-detección (FR-019).
    pub locale: Option<String>,
}

/// Referencia no resuelta tal como la ve la piel: origen + motivo, **sin contenido**.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnresolvedReport {
    pub origin: String,
    pub state: ReferenceState,
}

/// Vista de solo lectura de la sesión para la piel.
#[derive(Debug, Clone)]
pub struct SessionSnapshot {
    pub id: SessionId,
    pub state: SessionState,
    pub locale: Option<String>,
    pub artifacts: Vec<ContextArtifact>,
    pub unresolved: Vec<UnresolvedReport>,
    /// Lo que quedó fuera del presupuesto: se declara, nunca se trunca en silencio.
    pub omitted: Vec<String>,
    /// Qué llegó (o no) al repositorio. Es la respuesta a FR-017.
    pub writes: Vec<WriteRecord>,
    pub budget_exhausted: bool,
    pub interview_mode: bool,
    pub unattended_tentative: usize,
    /// Propuestas del refinamiento. Las que siguen sin aplicar son las que esperan decisión.
    pub proposals: Vec<ChangeProposal>,
}

/// Balance de una cancelación: en qué fase se cortó y qué alcanzó a escribirse (FR-023).
#[derive(Debug, Clone)]
pub struct CancelOutcome {
    pub session_id: SessionId,
    pub phase: SessionState,
    pub writes: Vec<WriteRecord>,
}

#[async_trait]
pub trait AuthoringService: Send + Sync {
    /// Arranca la sesión y **retorna de inmediato**: el trabajo sigue en segundo plano.
    ///
    /// Es lo que permite que la interfaz siga viva durante una sesión de minutos (FR-022).
    /// El avance se observa por los eventos de auditoría; el resultado, con `session_state`.
    async fn start_session(&self, request: StartSession) -> Result<SessionId>;

    async fn session_state(&self, id: &SessionId) -> Result<SessionSnapshot>;

    /// Cancela la sesión y devuelve el balance de lo que alcanzó a escribirse (FR-023).
    async fn cancel_session(&self, id: &SessionId) -> Result<CancelOutcome>;

    /// Espera a que la sesión termine. Útil para pieles que quieran bloquear (CLI) y para
    /// los tests; la piel de escritorio no la necesita porque escucha los eventos.
    async fn join_session(&self, id: &SessionId) -> Result<()>;

    async fn set_locale(&self, id: &SessionId, locale: String) -> Result<()>;

    /// Un turno de conversación de refinamiento: el usuario dice algo y salen propuestas.
    ///
    /// Lo de bajo riesgo ya viene aplicado; lo de alto impacto, decidido a través del
    /// `Prompter` — cuyo adapter es la piel (FR-010/FR-012).
    async fn submit_message(&self, id: &SessionId, message: &str) -> Result<Vec<ChangeProposal>>;

    /// Las propuestas que **siguen esperando** decisión.
    async fn pending_proposals(&self, id: &SessionId) -> Result<Vec<ChangeProposal>>;

    /// Registra una decisión sobre una propuesta concreta (FR-014/FR-015).
    async fn decide(&self, id: &SessionId, decision: ApprovalDecision) -> Result<()>;

    /// Deshace un cambio **auto-aplicado por bajo riesgo** (FR-008).
    ///
    /// Es la compensación de no haber preguntado: al usuario se le aplicó algo sin
    /// consultarle, así que tiene que poder devolverlo. Solo aplica a lo auto-aplicado —
    /// lo que pasó por una decisión humana se cambia decidiendo otra vez, no deshaciéndolo
    /// a espaldas de quien lo aprobó.
    async fn revert_proposal(&self, id: &SessionId, proposal_id: &ProposalId) -> Result<()>;

    /// Difiere un fragmento tentativo: el usuario lo ha **visto** y decide dejarlo declarado
    /// como pendiente en vez de resolverlo (FR-014 de `002-authoring-experience`).
    ///
    /// Es por fragmento y no en bloque a propósito. Un botón de «diferir todo» permitiría
    /// despachar sin leer lo que no está verificado, que es exactamente el hábito que este
    /// producto viene a corregir: lo tentativo se difiere **mirándolo**.
    ///
    /// Devuelve cuántos tentativos siguen sin atender.
    async fn defer_tentative(&self, id: &SessionId, path: &str, index: usize) -> Result<usize>;
}

struct SessionEntry {
    view: Arc<Mutex<SessionSnapshot>>,
    cancel: Arc<dyn Cancellation>,
    handle: Mutex<Option<JoinHandle<()>>>,
    /// El material que la sesión **leyó**, contra el que se comprueba toda cita (FR-006a).
    ///
    /// Vive aquí y no en la proyección por dos razones: el agregado muere con la tarea de
    /// ingesta —`submit_message` lo reconstruye desde la vista, sin fuentes—, y el contenido
    /// íntegro de los archivos no tiene nada que hacer en un snapshot destinado a la interfaz.
    material: Arc<Mutex<Vec<GatheredSource>>>,
}

pub struct ContextAuthoring {
    deps: AuthoringDeps,
    sessions: Mutex<HashMap<String, Arc<SessionEntry>>>,
    counter: AtomicU64,
    budget: Option<crate::application::ingest::IngestBudget>,
}

impl ContextAuthoring {
    pub fn new(deps: AuthoringDeps) -> Self {
        Self {
            deps,
            sessions: Mutex::new(HashMap::new()),
            counter: AtomicU64::new(1),
            budget: None,
        }
    }

    pub fn with_budget(mut self, budget: crate::application::ingest::IngestBudget) -> Self {
        self.budget = Some(budget);
        self
    }

    fn next_id(&self) -> SessionId {
        SessionId::new(format!(
            "session-{}",
            self.counter.fetch_add(1, Ordering::SeqCst)
        ))
    }

    fn build_loop(&self) -> AuthoringLoop {
        let l = AuthoringLoop::new(self.deps.clone());
        match &self.budget {
            Some(b) => l.with_budget(b.clone()),
            None => l,
        }
    }

    async fn entry(&self, id: &SessionId) -> Result<Arc<SessionEntry>> {
        self.sessions
            .lock()
            .await
            .get(id.as_str())
            .cloned()
            .ok_or_else(|| CoreError::NotFound(format!("sesión {}", id.as_str())))
    }

    /// Proyecta la sesión y el resultado de la ingesta a la vista que consume la piel.
    fn project(session: &AuthoringSession, outcome: &IngestOutcome) -> SessionSnapshot {
        SessionSnapshot {
            id: session.id().clone(),
            state: session.state(),
            locale: session.locale().map(|s| s.to_string()),
            artifacts: session.artifacts().to_vec(),
            unresolved: session
                .unresolved_references()
                .iter()
                .map(|r| UnresolvedReport {
                    origin: r.origin().as_str().to_string(),
                    state: r.state(),
                })
                .collect(),
            omitted: outcome.omitted.clone(),
            writes: session.writes().to_vec(),
            budget_exhausted: outcome.budget_exhausted,
            interview_mode: outcome.interview_mode,
            unattended_tentative: session.unattended_tentative_count(),
            proposals: Vec::new(),
        }
    }

    fn audit(&self, kind: AuditKind, payload: impl Into<String>) {
        self.deps
            .audit
            .record(AuditEvent::new(self.deps.clock.now_iso(), kind, payload));
    }
}

#[async_trait]
impl AuthoringService for ContextAuthoring {
    async fn start_session(&self, request: StartSession) -> Result<SessionId> {
        // El modo de la sesión debe coincidir con el grafo cableado: si el composition root
        // se armó para `Local`, no se puede pedir `Hybrid` por la puerta de atrás.
        if request.mode != self.deps.mode {
            return Err(CoreError::Invalid(format!(
                "la sesión pide modo {:?} pero el núcleo fue cableado para {:?}",
                request.mode, self.deps.mode
            )));
        }

        let id = self.next_id();
        let mut session = AuthoringSession::start(id.clone(), &request.repo_root, request.mode);
        if let Some(locale) = request.locale {
            session.set_locale(locale);
        }

        // Vista inicial publicada antes de arrancar: la piel puede consultarla enseguida.
        let view = Arc::new(Mutex::new(Self::project(
            &session,
            &IngestOutcome::default(),
        )));
        let material: Arc<Mutex<Vec<GatheredSource>>> = Arc::new(Mutex::new(Vec::new()));
        let cancel = self.deps.cancellations.create();

        let authoring = self.build_loop();
        let task_view = view.clone();
        let task_material = material.clone();
        let task_cancel = cancel.clone();
        let audit = self.deps.audit.clone();
        let clock = self.deps.clock.clone();

        // El trabajo se va a segundo plano: `start_session` no puede quedarse esperando
        // minutos o la interfaz se congela (FR-022).
        let handle = tokio::spawn(async move {
            let result = authoring.run(&mut session, task_cancel).await;

            let outcome = match result {
                Ok(outcome) => outcome,
                Err(CoreError::Cancelled) => {
                    let phase = session.state();
                    let _ = session.advance_to(SessionState::Cancelled);
                    let balance: Vec<String> =
                        session.writes().iter().map(|w| w.summary()).collect();
                    audit.record(AuditEvent::new(
                        clock.now_iso(),
                        AuditKind::SessionCancelled,
                        format!("en {phase:?}; escrituras: [{}]", balance.join(", ")),
                    ));
                    IngestOutcome::default()
                }
                Err(_) => {
                    let _ = session.advance_to(SessionState::Failed);
                    IngestOutcome::default()
                }
            };

            *task_material.lock().await = outcome.gathered.clone();
            *task_view.lock().await = Self::project(&session, &outcome);
        });

        self.sessions.lock().await.insert(
            id.as_str().to_string(),
            Arc::new(SessionEntry {
                view,
                cancel,
                handle: Mutex::new(Some(handle)),
                material,
            }),
        );

        Ok(id)
    }

    async fn session_state(&self, id: &SessionId) -> Result<SessionSnapshot> {
        let entry = self.entry(id).await?;
        let view = entry.view.lock().await;
        Ok(view.clone())
    }

    async fn cancel_session(&self, id: &SessionId) -> Result<CancelOutcome> {
        let entry = self.entry(id).await?;
        let phase = entry.view.lock().await.state;

        entry.cancel.cancel();

        // Esperar a que el loop desenrolle es lo que permite reportar un balance **cierto**:
        // decir "no sé qué se escribió" sería justo lo que FR-023 viene a evitar.
        let handle = entry.handle.lock().await.take();
        if let Some(h) = handle {
            let _ = h.await;
        }

        let view = entry.view.lock().await;
        Ok(CancelOutcome {
            session_id: id.clone(),
            phase,
            writes: view.writes.clone(),
        })
    }

    async fn join_session(&self, id: &SessionId) -> Result<()> {
        let entry = self.entry(id).await?;
        let handle = entry.handle.lock().await.take();
        if let Some(h) = handle {
            let _ = h.await;
        }
        Ok(())
    }

    async fn set_locale(&self, id: &SessionId, locale: String) -> Result<()> {
        let entry = self.entry(id).await?;
        let mut view = entry.view.lock().await;
        view.locale = Some(locale);
        Ok(())
    }

    async fn submit_message(&self, id: &SessionId, message: &str) -> Result<Vec<ChangeProposal>> {
        let entry = self.entry(id).await?;
        let cancel = entry.cancel.clone();

        // El agregado vivió dentro de la tarea del loop y ya no existe; se reconstruye desde
        // la proyección, que es la que sigue siendo la verdad de la sesión. Es el mismo
        // camino que sigue `set_locale`.
        let (mut session, previas) = {
            let view = entry.view.lock().await;
            let mut s = AuthoringSession::start(view.id.clone(), ".", self.deps.mode);
            if let Some(l) = &view.locale {
                s.set_locale(l.clone());
            }
            for a in &view.artifacts {
                s.put_artifact(a.clone());
            }
            (s, view.proposals.clone())
        };

        let material = entry.material.lock().await.clone();
        let refine =
            crate::application::refine::RefineLoop::new(self.deps.clone()).with_material(material);
        let outcome = refine.submit_message(&mut session, message, cancel).await?;

        let mut view = entry.view.lock().await;
        view.artifacts = session.artifacts().to_vec();
        view.unattended_tentative = session.unattended_tentative_count();
        view.proposals = previas
            .into_iter()
            .chain(outcome.proposals.iter().cloned())
            .collect();

        Ok(outcome.proposals)
    }

    async fn pending_proposals(&self, id: &SessionId) -> Result<Vec<ChangeProposal>> {
        let entry = self.entry(id).await?;
        let view = entry.view.lock().await;
        Ok(view
            .proposals
            .iter()
            .filter(|p| !p.applied)
            .cloned()
            .collect())
    }

    async fn decide(&self, id: &SessionId, decision: ApprovalDecision) -> Result<()> {
        let entry = self.entry(id).await?;
        let material = entry.material.lock().await.clone();
        let mut view = entry.view.lock().await;

        let proposal = view
            .proposals
            .iter_mut()
            .find(|p| p.id == decision.proposal_id)
            .ok_or_else(|| {
                CoreError::NotFound(format!("propuesta {}", decision.proposal_id.as_str()))
            })?;

        // Decidir dos veces sobre lo mismo no puede aplicar dos veces: si ya está aplicada,
        // la decisión llega tarde y decirlo es mejor que fingir que surtió efecto.
        if proposal.applied {
            return Err(CoreError::Invalid(format!(
                "la propuesta {} ya se aplicó: no se puede volver a decidir sobre ella",
                decision.proposal_id.as_str()
            )));
        }

        let ChangeTarget::Artifact(kind) = proposal.target.clone() else {
            return Err(CoreError::Invalid(
                "solo se decide sobre artefactos de contexto".into(),
            ));
        };

        let contenido = match &decision.verdict {
            Verdict::Approve => Some(proposal.diff.after.clone()),
            Verdict::Edit(texto) => Some(texto.clone()),
            // Rechazar NO toca el artefacto: es la garantía de FR-015.
            Verdict::Reject => None,
        };

        if let Some(contenido) = contenido {
            proposal.applied = true;
            let locale = view.locale.clone().unwrap_or_else(|| "en".into());
            let segments =
                crate::application::authoring_loop::parse_segments(&contenido, &material)
                    .unwrap_or_else(|_| {
                    vec![crate::domain::context::Segment::tentative(
                        contenido.clone(),
                        "proviene del refinamiento conversacional; no se ha verificado contra una fuente",
                    )]
                });
            let artifact = ContextArtifact::new(kind, locale).with_segments(segments);
            view.artifacts.retain(|a| a.kind != kind);
            view.artifacts.push(artifact);
            view.unattended_tentative = view
                .artifacts
                .iter()
                .map(|a| a.unattended_tentative_count())
                .sum();
        }

        self.audit(
            AuditKind::ApprovalCaptured,
            format!(
                "{}: {} por {}",
                decision.proposal_id.as_str(),
                if decision.is_rejection() {
                    "rechazada"
                } else {
                    "aplicada"
                },
                decision.actor
            ),
        );
        Ok(())
    }

    async fn revert_proposal(&self, id: &SessionId, proposal_id: &ProposalId) -> Result<()> {
        let entry = self.entry(id).await?;
        let material = entry.material.lock().await.clone();
        let mut view = entry.view.lock().await;

        let proposal = view
            .proposals
            .iter_mut()
            .find(|p| &p.id == proposal_id)
            .ok_or_else(|| CoreError::NotFound(format!("propuesta {}", proposal_id.as_str())))?;

        if !proposal.applied {
            return Err(CoreError::Invalid(format!(
                "la propuesta {} no está aplicada: no hay nada que deshacer",
                proposal_id.as_str()
            )));
        }
        if proposal.requires_approval() {
            return Err(CoreError::Invalid(format!(
                "la propuesta {} se aplicó tras una decisión explícita: para cambiarla hay que \
                 volver a decidir, no deshacerla a espaldas de quien la aprobó",
                proposal_id.as_str()
            )));
        }

        let ChangeTarget::Artifact(kind) = proposal.target.clone() else {
            return Err(CoreError::Invalid(
                "solo se deshacen cambios sobre artefactos de contexto".into(),
            ));
        };

        // Se pide el «antes» al motor de diffs en vez de leerlo del propio diff: así la
        // reversión pasa por el port —con su propiedad `apply∘revert = identidad` verificada
        // por contract test— y no por una copia de campo que podría desincronizarse.
        let anterior = self
            .deps
            .diff
            .revert(&proposal.diff.after, &proposal.diff)?;
        proposal.applied = false;
        let proposal_id_str = proposal_id.as_str().to_string();

        let locale = view.locale.clone().unwrap_or_else(|| "en".into());
        let segments = crate::application::authoring_loop::parse_segments(&anterior, &material)
            .unwrap_or_else(|_| {
                vec![crate::domain::context::Segment::tentative(
                    anterior.clone(),
                    "proviene del refinamiento conversacional; no se ha verificado contra una fuente",
                )]
            });
        view.artifacts.retain(|a| a.kind != kind);
        view.artifacts
            .push(ContextArtifact::new(kind, locale).with_segments(segments));
        view.unattended_tentative = view
            .artifacts
            .iter()
            .map(|a| a.unattended_tentative_count())
            .sum();

        self.audit(AuditKind::ProposalReverted, proposal_id_str);
        Ok(())
    }

    async fn defer_tentative(&self, id: &SessionId, path: &str, index: usize) -> Result<usize> {
        let entry = self.entry(id).await?;
        let mut view = entry.view.lock().await;

        {
            let artifact = view
                .artifacts
                .iter_mut()
                .find(|a| a.kind.file_path() == path)
                .ok_or_else(|| CoreError::NotFound(format!("artefacto {path}")))?;

            let segment = artifact.segments.get_mut(index).ok_or_else(|| {
                CoreError::NotFound(format!("fragmento {index} del artefacto {path}"))
            })?;

            // Diferir solo tiene sentido sobre lo tentativo. Aceptarlo sobre un fragmento
            // fundamentado dejaría pasar en silencio un error de la piel — y peor, sugeriría
            // que hay algo que atender donde no lo hay.
            if !matches!(
                segment.groundedness,
                crate::domain::context::Groundedness::Tentative { .. }
            ) {
                return Err(CoreError::Invalid(format!(
                    "el fragmento {index} de {path} no es tentativo: no hay nada que diferir"
                )));
            }

            segment.acknowledge();
        }

        // El contador se recalcula desde los artefactos, no se decrementa: así no puede
        // desincronizarse si alguien difiere dos veces el mismo fragmento.
        view.unattended_tentative = view
            .artifacts
            .iter()
            .map(|a| a.unattended_tentative_count())
            .sum();

        Ok(view.unattended_tentative)
    }
}

impl ContextAuthoring {
    /// Sondea el backend de modelo configurado (FR-019). No falla: informa.
    pub async fn probe_provider(&self) -> crate::application::ports::ProviderStatus {
        let status = self.deps.discovery.probe().await;
        if !status.reachable {
            self.audit(
                AuditKind::IngestBudgetExhausted,
                format!("proveedor no disponible: {}", status.endpoint),
            );
        }
        status
    }
}
