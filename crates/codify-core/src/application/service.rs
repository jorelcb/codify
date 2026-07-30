//! Application Service: el punto de entrada del caso de uso.
//!
//! Es lo que consumen las pieles (Tauri hoy; MCP/CLI mañana). El trait existe porque hay
//! **más de un adaptador primario** previsto — que es la única razón que justifica una
//! interfaz driving (constitución, Principio I).
//!
//! Nombre sin decoración: `ContextAuthoring` nombra la capacidad, no el patrón.

use crate::application::authoring_loop::{AuthoringLoop, IngestOutcome};
use crate::application::deps::AuthoringDeps;
use crate::domain::context::ContextArtifact;
use crate::domain::error::{CoreError, Result};
use crate::domain::reference::ReferenceState;
use crate::domain::session::{AuthoringSession, Mode, SessionId, SessionState};
use async_trait::async_trait;
use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};
use tokio::sync::Mutex;

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
    pub budget_exhausted: bool,
    pub interview_mode: bool,
    pub unattended_tentative: usize,
}

#[async_trait]
pub trait AuthoringService: Send + Sync {
    /// Arranca la sesión y ejecuta el pase de US1 (ingesta dirigida + generación grounded).
    async fn start_session(&self, request: StartSession) -> Result<SessionId>;
    async fn session_state(&self, id: &SessionId) -> Result<SessionSnapshot>;
    async fn set_locale(&self, id: &SessionId, locale: String) -> Result<()>;
}

struct SessionEntry {
    session: AuthoringSession,
    outcome: IngestOutcome,
}

pub struct ContextAuthoring {
    deps: AuthoringDeps,
    sessions: Mutex<HashMap<String, SessionEntry>>,
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

    fn snapshot(entry: &SessionEntry) -> SessionSnapshot {
        SessionSnapshot {
            id: entry.session.id().clone(),
            state: entry.session.state(),
            locale: entry.session.locale().map(|s| s.to_string()),
            artifacts: entry.session.artifacts().to_vec(),
            unresolved: entry
                .session
                .unresolved_references()
                .iter()
                .map(|r| UnresolvedReport {
                    origin: r.origin().as_str().to_string(),
                    state: r.state(),
                })
                .collect(),
            omitted: entry.outcome.omitted.clone(),
            budget_exhausted: entry.outcome.budget_exhausted,
            interview_mode: entry.outcome.interview_mode,
            unattended_tentative: entry.session.unattended_tentative_count(),
        }
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

        let outcome = self.build_loop().run(&mut session).await?;

        self.sessions
            .lock()
            .await
            .insert(id.as_str().to_string(), SessionEntry { session, outcome });
        Ok(id)
    }

    async fn session_state(&self, id: &SessionId) -> Result<SessionSnapshot> {
        let sessions = self.sessions.lock().await;
        sessions
            .get(id.as_str())
            .map(Self::snapshot)
            .ok_or_else(|| CoreError::NotFound(format!("sesión {}", id.as_str())))
    }

    async fn set_locale(&self, id: &SessionId, locale: String) -> Result<()> {
        let mut sessions = self.sessions.lock().await;
        let entry = sessions
            .get_mut(id.as_str())
            .ok_or_else(|| CoreError::NotFound(format!("sesión {}", id.as_str())))?;
        entry.session.set_locale(locale);
        Ok(())
    }
}
