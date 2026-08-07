//! Clasificador de riesgo conservador, v1 (T036) — FR-012.
//!
//! **Por defecto, bloquea.** Solo se deja pasar sin aprobación lo que se puede demostrar que no
//! cambia nada.
//!
//! La asimetría es el punto, no un detalle de implementación: equivocarse hacia `HighImpact`
//! cuesta una confirmación de más; equivocarse hacia `Low` escribe en el repositorio del
//! usuario algo que nunca aprobó. Un producto cuyo principio rector es "lo que no se puede
//! verificar no se afirma" no puede tener un clasificador optimista.
//!
//! El criterio fino de "bajo riesgo" —qué reformulaciones son realmente inocuas— queda para un
//! spec derivado. Hasta entonces, esta política prefiere molestar a colarse.

use crate::domain::change::{ChangeProposal, RiskLevel};
use crate::domain::ports::RiskClassifier;

pub struct ConservativeRiskClassifier;

impl ConservativeRiskClassifier {
    /// ¿El cambio es puramente cosmético? Solo cuenta como tal si, ignorando espacios en
    /// blanco, los dos lados dicen **exactamente** lo mismo.
    ///
    /// Nótese lo que NO se mira: la justificación que acompaña a la propuesta. Si el veredicto
    /// dependiera de ella, bastaría con que el modelo escribiera «cambio menor» para saltarse
    /// la aprobación — el clasificador estaría confiando en la parte interesada.
    fn is_whitespace_only(proposal: &ChangeProposal) -> bool {
        let normalizar = |s: &str| s.split_whitespace().collect::<Vec<_>>().join(" ");
        normalizar(&proposal.diff.before) == normalizar(&proposal.diff.after)
    }
}

impl RiskClassifier for ConservativeRiskClassifier {
    fn classify(&self, proposal: &ChangeProposal) -> RiskLevel {
        if proposal.diff.is_empty() || Self::is_whitespace_only(proposal) {
            RiskLevel::Low
        } else {
            RiskLevel::HighImpact
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::change::{ChangeTarget, Diff, ProposalId, ProposalOrigin};
    use crate::domain::context::ArtifactKind;

    fn propuesta(before: &str, after: &str) -> ChangeProposal {
        ChangeProposal::new(
            ProposalId::new("p"),
            ChangeTarget::Artifact(ArtifactKind::Context),
            Diff {
                unified: String::new(),
                before: before.into(),
                after: after.into(),
            },
            RiskLevel::Low,
            "da igual lo que diga",
            ProposalOrigin::Refinement,
        )
    }

    #[test]
    fn reindenting_does_not_require_approval() {
        let p = propuesta("- uno\n-   dos\n", "-  uno\n-  dos\n");
        assert_eq!(ConservativeRiskClassifier.classify(&p), RiskLevel::Low);
    }

    #[test]
    fn changing_a_single_word_does_require_approval() {
        let p = propuesta("Motor: Kafka", "Motor: Temporal");
        assert_eq!(
            ConservativeRiskClassifier.classify(&p),
            RiskLevel::HighImpact
        );
    }

    /// Añadir contenido nunca es inocuo: un párrafo nuevo puede afirmar algo sin fuente.
    #[test]
    fn adding_content_requires_approval() {
        let p = propuesta("Motor: Temporal", "Motor: Temporal\nEscala a 10k workers.");
        assert_eq!(
            ConservativeRiskClassifier.classify(&p),
            RiskLevel::HighImpact
        );
    }

    /// Borrar tampoco: perder una afirmación fundamentada es tan grave como inventar una.
    #[test]
    fn removing_content_requires_approval() {
        let p = propuesta(
            "Motor: Temporal\nPersistencia: PostgreSQL",
            "Motor: Temporal",
        );
        assert_eq!(
            ConservativeRiskClassifier.classify(&p),
            RiskLevel::HighImpact
        );
    }
}
