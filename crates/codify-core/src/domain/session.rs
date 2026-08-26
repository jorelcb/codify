//! `AuthoringSession`: agregado raíz del loop de authoring.
//!
//! Invariantes:
//! 1. Las transiciones de estado siguen `Ingesting → Generating → Refining → Approved`
//!    (con `Failed`/`Cancelled` alcanzables desde cualquier estado no terminal).
//! 2. **No se puede aprobar** dejando segmentos tentativos sin atender (FR-013).

use crate::domain::context::ContextArtifact;
use crate::domain::reference::Reference;
use crate::domain::write::WriteRecord;
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
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize, Default)]
pub enum Mode {
    /// El caso por defecto del producto: conectar un remoto es una decisión explícita.
    #[default]
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
    /// Identificador estable para la piel. Es **parte del contrato**: la ventana compone con
    /// él la clave del catálogo (`session.state.<code>`).
    ///
    /// Derivarlo de `Debug` —como se hacía— ataba las claves al nombre de la variante:
    /// renombrar `Approved` habría dejado la interfaz mostrando la clave cruda, sin que
    /// ningún test avisara.
    pub fn code(&self) -> &'static str {
        match self {
            SessionState::Ingesting => "ingesting",
            SessionState::Generating => "generating",
            SessionState::Refining => "refining",
            SessionState::Approved => "approved",
            SessionState::Failed => "failed",
            SessionState::Cancelled => "cancelled",
        }
    }

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

    #[error("no se puede pasar a Failed sin motivo: usa fail(SessionFailure)")]
    FailureNeedsReason,
}

/// Por qué murió una sesión (`002`-FR-028).
///
/// El núcleo devuelve un **código estable** y la piel elige la frase, igual que `ProviderIssue`
/// o `ReferenceState`. Ese desacople es lo que permite que el motivo no nazca redactado en un
/// idioma fijo (SC-009) — y lo que hace que añadir un motivo sin texto rompa un test en vez de
/// aparecer en pantalla como `session.failure.loquesea`.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SessionFailure {
    /// El modelo no contestó a tiempo. Se espera más, o se usa uno más rápido.
    ProviderTimeout,
    /// El backend no responde. Se comprueba que esté levantado.
    ProviderUnavailable,
    /// Contestó algo que no se pudo interpretar. Se revisa el prompt o el modelo.
    ProviderUnparseable,
    /// No se pudo leer el repositorio o escribir en él.
    RepoUnreadable,
    /// La política de cero-egress cortó la operación. En modo local es lo esperado.
    EgressBlocked,
    /// Falta autorización para algo que se intentó.
    Unauthorized,
    /// Cualquier otro fallo del propio sistema.
    Internal,
}

impl SessionFailure {
    pub fn code(&self) -> &'static str {
        match self {
            SessionFailure::ProviderTimeout => "provider_timeout",
            SessionFailure::ProviderUnavailable => "provider_unavailable",
            SessionFailure::ProviderUnparseable => "provider_unparseable",
            SessionFailure::RepoUnreadable => "repo_unreadable",
            SessionFailure::EgressBlocked => "egress_blocked",
            SessionFailure::Unauthorized => "unauthorized",
            SessionFailure::Internal => "internal",
        }
    }

    /// Todas las variantes. La piel la recorre para comprobar que ninguna se queda sin texto.
    pub fn all() -> [SessionFailure; 7] {
        [
            SessionFailure::ProviderTimeout,
            SessionFailure::ProviderUnavailable,
            SessionFailure::ProviderUnparseable,
            SessionFailure::RepoUnreadable,
            SessionFailure::EgressBlocked,
            SessionFailure::Unauthorized,
            SessionFailure::Internal,
        ]
    }
}

impl From<&crate::domain::error::CoreError> for SessionFailure {
    fn from(e: &crate::domain::error::CoreError) -> Self {
        use crate::domain::error::CoreError as E;
        match e {
            E::ProviderTimeout(_) => SessionFailure::ProviderTimeout,
            E::Unavailable(_) => SessionFailure::ProviderUnavailable,
            E::Provider(_) => SessionFailure::ProviderUnparseable,
            E::Storage(_) | E::NotFound(_) => SessionFailure::RepoUnreadable,
            E::EgressBlocked(_) => SessionFailure::EgressBlocked,
            E::Unauthorized(_) => SessionFailure::Unauthorized,
            E::Invalid(_) => SessionFailure::Internal,
            // Cancelar no es fallar: tiene su propio estado y su propio balance (FR-023). Si
            // llega aquí, alguien confundió las dos cosas — se mapea para que el match sea
            // total, no porque tenga sentido.
            E::Cancelled => SessionFailure::Internal,
        }
    }
}

pub struct AuthoringSession {
    id: SessionId,
    repo_root: PathBuf,
    mode: Mode,
    locale: Option<String>,
    state: SessionState,
    artifacts: Vec<ContextArtifact>,
    references: Vec<Reference>,
    writes: Vec<WriteRecord>,
    failure: Option<SessionFailure>,
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
            writes: Vec::new(),
            failure: None,
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

    /// Registra lo ocurrido con un artefacto. Es lo que la sesión reporta al cerrar,
    /// se haya completado o cancelado (FR-017/FR-023).
    pub fn record_write(&mut self, record: WriteRecord) {
        self.writes.push(record);
    }

    pub fn writes(&self) -> &[WriteRecord] {
        &self.writes
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

    /// Avanza de estado. **No admite `Failed`**: para eso está `fail(motivo)`.
    ///
    /// Podría bastar con un campo opcional y la disciplina de rellenarlo, pero esa disciplina ya
    /// falló: el `Err(_)` que descartaba el error llevaba ahí desde el principio y nadie lo notó
    /// hasta que costó cinco corridas diagnosticar un timeout (`002`-FR-028). Que el tipo lo
    /// exija es lo que impide que vuelva a perderse.
    pub fn advance_to(&mut self, to: SessionState) -> Result<(), SessionError> {
        if to == SessionState::Approved {
            return self.approve();
        }
        if to == SessionState::Failed {
            return Err(SessionError::FailureNeedsReason);
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

    /// Da la sesión por fallida **con su motivo** (`002`-FR-028).
    ///
    /// No devuelve `Result`: un fallo no puede fallar. Si el estado no admitiera la transición
    /// y esto rechazara, el motivo se perdería justo cuando más falta hace.
    pub fn fail(&mut self, motivo: SessionFailure) {
        self.state = SessionState::Failed;
        self.failure = Some(motivo);
    }

    /// Por qué murió la sesión, si murió.
    pub fn failure(&self) -> Option<SessionFailure> {
        self.failure
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
