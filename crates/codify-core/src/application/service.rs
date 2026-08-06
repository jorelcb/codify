//! Application Service: el punto de entrada del caso de uso.
//!
//! Es lo que consumen las pieles (Tauri hoy; MCP/CLI mañana). El trait existe porque hay
//! **más de un adaptador primario** previsto — la única razón que justifica una interfaz
//! driving (constitución, Principio I).
//!
//! Nombre sin decoración: `ContextAuthoring` nombra la capacidad, no el patrón.

use crate::application::authoring_loop::{AuthoringLoop, IngestOutcome};
use crate::application::deps::AuthoringDeps;
use crate::application::ports::Cancellation;
use crate::domain::audit::{AuditEvent, AuditKind};
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
        let cancel = self.deps.cancellations.create();

        let authoring = self.build_loop();
        let task_view = view.clone();
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

            *task_view.lock().await = Self::project(&session, &outcome);
        });

        self.sessions.lock().await.insert(
            id.as_str().to_string(),
            Arc::new(SessionEntry {
                view,
                cancel,
                handle: Mutex::new(Some(handle)),
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
