//! Ports que **el Dominio nombra** (constitución, Principio I: regla de ubicación).
//!
//! Son políticas puras y capacidades que el dominio necesita para expresar sus invariantes.
//! Todo lo que representa un *efecto externo* vive en `application::ports`.

use crate::domain::change::{ChangeProposal, RiskLevel};

/// Clasifica el riesgo de una propuesta. Es una **política de dominio**: `RiskLevel` es un
/// value object del dominio y de él depende si el loop auto-aplica o pide aprobación.
///
/// v1 usa una política conservadora; el criterio fino de "bajo riesgo" se define en un
/// spec derivado (FR-012).
pub trait RiskClassifier: Send + Sync {
    fn classify(&self, proposal: &ChangeProposal) -> RiskLevel;
}

/// El dominio nombra el tiempo (fecha de los `AuditEvent` y de las decisiones), pero no
/// sabe leerlo: se inyecta.
pub trait Clock: Send + Sync {
    fn now_iso(&self) -> String;
}
