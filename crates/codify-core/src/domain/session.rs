//! `AuthoringSession`: agregado raíz del loop de authoring.
//!
//! Invariantes:
//! 1. Las transiciones de estado siguen `Ingesting → Generating → Refining → Approved`
//!    (con `Failed`/`Cancelled` alcanzables desde cualquier estado no terminal).
//! 2. **No se puede aprobar** dejando segmentos tentativos sin atender (FR-013).

use crate::domain::context::ContextArtifact;
use crate::domain::reference::Reference;
use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};
use thiserror::Error;

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct SessionId(String);

impl SessionId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }
    pub fn as_str(&self) -> &str {
        &self.0
    }
}

/// Modo de operación. `Local` implica la garantía de cero-egress, que se materializa
/// estructuralmente en el composition root (no con un flag consultado en runtime).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Mode {
    Local,
    Hybrid,
}

impl Mode {
    pub fn is_local(&self) -> bool {
        matches!(self, Mode::Local)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SessionState {
    Ingesting,
    Generating,
    Refining,
    Approved,
    Failed,
    Cancelled,
}

impl SessionState {
    pub fn is_terminal(&self) -> bool {
        matches!(
            self,
            SessionState::Approved | SessionState::Failed | SessionState::Cancelled
        )
    }
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum SessionError {
    #[error("transición inválida: {from:?} → {to:?}")]
    InvalidTransition {
        from: SessionState,
        to: SessionState,
    },

    #[error("no se puede aprobar: quedan {0} segmento(s) tentativo(s) sin atender")]
    PendingTentativeSegments(usize),
}

#[derive(Debug, Clone)]
pub struct AuthoringSession {
    id: SessionId,
    repo_root: PathBuf,
    mode: Mode,
    locale: Option<String>,
    state: SessionState,
    artifacts: Vec<ContextArtifact>,
    references: Vec<Reference>,
}

impl AuthoringSession {
    pub fn start(id: SessionId, repo_root: impl AsRef<Path>, mode: Mode) -> Self {
        Self {
            id,
            repo_root: repo_root.as_ref().to_path_buf(),
            mode,
            locale: None,
            state: SessionState::Ingesting,
            artifacts: Vec::new(),
            references: Vec::new(),
        }
    }

    pub fn id(&self) -> &SessionId {
        &self.id
    }
    pub fn repo_root(&self) -> &Path {
        &self.repo_root
    }
    pub fn mode(&self) -> Mode {
        self.mode
    }
    pub fn state(&self) -> SessionState {
        self.state
    }
    pub fn locale(&self) -> Option<&str> {
        self.locale.as_deref()
    }
    pub fn artifacts(&self) -> &[ContextArtifact] {
        &self.artifacts
    }
    pub fn references(&self) -> &[Reference] {
        &self.references
    }

    /// Fija el idioma (auto-detectado u override explícito del usuario — FR-019).
    pub fn set_locale(&mut self, locale: impl Into<String>) {
        self.locale = Some(locale.into());
    }

    pub fn record_reference(&mut self, reference: Reference) {
        self.references.push(reference);
    }

    pub fn put_artifact(&mut self, artifact: ContextArtifact) {
        match self.artifacts.iter_mut().find(|a| a.kind == artifact.kind) {
            Some(existing) => *existing = artifact,
            None => self.artifacts.push(artifact),
        }
    }

    /// Referencias que no se pudieron resolver: se declaran, nunca se inventan (FR-004).
    pub fn unresolved_references(&self) -> Vec<&Reference> {
        self.references
            .iter()
            .filter(|r| !r.is_resolved())
            .collect()
    }

    pub fn unattended_tentative_count(&self) -> usize {
        self.artifacts
            .iter()
            .map(|a| a.unattended_tentative_count())
            .sum()
    }

    pub fn advance_to(&mut self, to: SessionState) -> Result<(), SessionError> {
        if to == SessionState::Approved {
            return self.approve();
        }
        if !Self::can_transition(self.state, to) {
            return Err(SessionError::InvalidTransition {
                from: self.state,
                to,
            });
        }
        self.state = to;
        Ok(())
    }

    /// Cierra la sesión. Falla si queda contexto tentativo sin atender (FR-013).
    pub fn approve(&mut self) -> Result<(), SessionError> {
        if !Self::can_transition(self.state, SessionState::Approved) {
            return Err(SessionError::InvalidTransition {
                from: self.state,
                to: SessionState::Approved,
            });
        }
        let pending = self.unattended_tentative_count();
        if pending > 0 {
            return Err(SessionError::PendingTentativeSegments(pending));
        }
        self.state = SessionState::Approved;
        Ok(())
    }

    fn can_transition(from: SessionState, to: SessionState) -> bool {
        use SessionState::*;
        if from.is_terminal() {
            return false;
        }
        match to {
            Failed | Cancelled => true,
            Generating => from == Ingesting,
            Refining => matches!(from, Generating | Refining),
            Approved => matches!(from, Generating | Refining),
            Ingesting => false,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::context::{ArtifactKind, ContextArtifact, Segment};

    fn session() -> AuthoringSession {
        AuthoringSession::start(SessionId::new("s1"), "/tmp/repo", Mode::Local)
    }

    #[test]
    fn happy_path_transitions() {
        let mut s = session();
        assert_eq!(s.state(), SessionState::Ingesting);
        s.advance_to(SessionState::Generating).unwrap();
        s.advance_to(SessionState::Refining).unwrap();
        s.approve().unwrap();
        assert_eq!(s.state(), SessionState::Approved);
    }

    #[test]
    fn cannot_skip_from_ingesting_to_refining() {
        let mut s = session();
        let err = s.advance_to(SessionState::Refining).unwrap_err();
        assert_eq!(
            err,
            SessionError::InvalidTransition {
                from: SessionState::Ingesting,
                to: SessionState::Refining
            }
        );
    }

    #[test]
    fn cannot_approve_with_unattended_tentative_segments() {
        let mut s = session();
        s.advance_to(SessionState::Generating).unwrap();
        s.put_artifact(
            ContextArtifact::new(ArtifactKind::Context, "es")
                .with_segments(vec![Segment::tentative("broker?", "ninguna fuente")]),
        );
        let err = s.approve().unwrap_err();
        assert_eq!(err, SessionError::PendingTentativeSegments(1));
        assert_ne!(s.state(), SessionState::Approved);
    }

    #[test]
    fn approving_succeeds_once_tentative_segments_are_acknowledged() {
        let mut s = session();
        s.advance_to(SessionState::Generating).unwrap();
        let mut seg = Segment::tentative("broker?", "ninguna fuente");
        seg.acknowledge();
        s.put_artifact(ContextArtifact::new(ArtifactKind::Context, "es").with_segments(vec![seg]));
        s.approve().unwrap();
        assert_eq!(s.state(), SessionState::Approved);
    }

    #[test]
    fn terminal_state_rejects_further_transitions() {
        let mut s = session();
        s.advance_to(SessionState::Cancelled).unwrap();
        assert!(s.advance_to(SessionState::Generating).is_err());
    }
}
