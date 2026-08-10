//! Repositorio objetivo y referencias aludidas desde el material leído.
//!
//! Invariante central (SC-006): una referencia **no resuelta nunca lleva contenido**. El tipo
//! hace imposible fabricarlo: sólo el constructor `resolved` acepta contenido.

use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ReferenceOrigin {
    /// Ruta relativa a otro archivo dentro del repositorio.
    LocalPath(String),
    /// URL pública (v1: sin autenticación — FR-003).
    PublicUrl(String),
}

impl ReferenceOrigin {
    pub fn as_str(&self) -> &str {
        match self {
            ReferenceOrigin::LocalPath(p) => p,
            ReferenceOrigin::PublicUrl(u) => u,
        }
    }

    pub fn is_remote(&self) -> bool {
        matches!(self, ReferenceOrigin::PublicUrl(_))
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ReferenceState {
    Resolved,
    Inaccessible,
    /// Requiere autenticación: fuera de v1. Se reporta, jamás se inventa (FR-003/FR-004).
    RequiresAuth,
    /// Fuera del presupuesto de muestreo: se declara, no se trunca en silencio.
    OutOfScope,
}

impl ReferenceState {
    /// Identificador estable para la piel, por el mismo motivo que en `SessionState`.
    pub fn code(&self) -> &'static str {
        match self {
            ReferenceState::Resolved => "resolved",
            ReferenceState::Inaccessible => "inaccessible",
            ReferenceState::RequiresAuth => "requires_auth",
            ReferenceState::OutOfScope => "out_of_scope",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Reference {
    origin: ReferenceOrigin,
    state: ReferenceState,
    content: Option<String>,
}

impl Reference {
    /// Única vía para construir una referencia **con** contenido.
    pub fn resolved(origin: ReferenceOrigin, content: impl Into<String>) -> Self {
        Self {
            origin,
            state: ReferenceState::Resolved,
            content: Some(content.into()),
        }
    }

    /// Construye una referencia no resuelta. El contenido queda forzosamente vacío:
    /// no existe forma de fabricarlo desde el dominio.
    pub fn unresolved(origin: ReferenceOrigin, state: ReferenceState) -> Self {
        debug_assert!(
            state != ReferenceState::Resolved,
            "unresolved() no admite el estado Resolved"
        );
        let state = if state == ReferenceState::Resolved {
            ReferenceState::Inaccessible
        } else {
            state
        };
        Self {
            origin,
            state,
            content: None,
        }
    }

    pub fn origin(&self) -> &ReferenceOrigin {
        &self.origin
    }

    pub fn state(&self) -> ReferenceState {
        self.state
    }

    pub fn is_resolved(&self) -> bool {
        self.state == ReferenceState::Resolved
    }

    /// Contenido disponible solo si la referencia está resuelta.
    pub fn content(&self) -> Option<&str> {
        match self.state {
            ReferenceState::Resolved => self.content.as_deref(),
            _ => None,
        }
    }
}

/// El repositorio objetivo sobre el que se autora el contexto.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Repository {
    pub root: PathBuf,
    pub is_empty: bool,
    pub detected_language: Option<String>,
    pub structural_signals: Vec<String>,
}

impl Repository {
    pub fn new(root: impl AsRef<Path>) -> Self {
        Self {
            root: root.as_ref().to_path_buf(),
            is_empty: false,
            detected_language: None,
            structural_signals: Vec::new(),
        }
    }

    /// Un repo vacío deriva al modo entrevista: no se falla ni se inventa.
    pub fn requires_interview(&self) -> bool {
        self.is_empty
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn unresolved_reference_never_exposes_content() {
        let r = Reference::unresolved(
            ReferenceOrigin::PublicUrl("https://example.test/private".into()),
            ReferenceState::RequiresAuth,
        );
        assert!(!r.is_resolved());
        assert_eq!(r.content(), None);
        assert_eq!(r.state(), ReferenceState::RequiresAuth);
    }

    #[test]
    fn resolved_reference_exposes_its_content() {
        let r = Reference::resolved(
            ReferenceOrigin::LocalPath("SPEC-30.md".into()),
            "sin broker",
        );
        assert!(r.is_resolved());
        assert_eq!(r.content(), Some("sin broker"));
    }

    #[test]
    fn empty_repository_requires_interview_mode() {
        let mut repo = Repository::new("/tmp/x");
        repo.is_empty = true;
        assert!(repo.requires_interview());
    }
}
