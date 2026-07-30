//! Propuestas de cambio (diffs) y su aprobación.
//!
//! Reglas del loop **curado** (FR-010/FR-012): `Low` se auto-aplica y queda visible y
//! revertible; `HighImpact` exige aprobación explícita **antes** de escribir.

use crate::domain::context::ArtifactKind;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct ProposalId(String);

impl ProposalId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }
    pub fn as_str(&self) -> &str {
        &self.0
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum RiskLevel {
    Low,
    HighImpact,
}

impl RiskLevel {
    /// Solo el alto impacto bloquea la escritura a la espera del usuario.
    pub fn requires_approval(&self) -> bool {
        matches!(self, RiskLevel::HighImpact)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ProposalOrigin {
    Generation,
    Refinement,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChangeTarget {
    Artifact(ArtifactKind),
    RepoFile(String),
}

/// Diff con ambos lados presentes: hace la reversión total y verificable
/// (propiedad `revert(apply(before)) == before`).
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Diff {
    pub unified: String,
    pub before: String,
    pub after: String,
}

impl Diff {
    pub fn is_empty(&self) -> bool {
        self.before == self.after
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ChangeProposal {
    pub id: ProposalId,
    pub target: ChangeTarget,
    pub diff: Diff,
    pub risk: RiskLevel,
    pub rationale: String,
    pub origin: ProposalOrigin,
    /// `true` cuando ya se aplicó (auto-aplicación de bajo riesgo o aprobación).
    pub applied: bool,
}

impl ChangeProposal {
    pub fn new(
        id: ProposalId,
        target: ChangeTarget,
        diff: Diff,
        risk: RiskLevel,
        rationale: impl Into<String>,
        origin: ProposalOrigin,
    ) -> Self {
        Self {
            id,
            target,
            diff,
            risk,
            rationale: rationale.into(),
            origin,
            applied: false,
        }
    }

    pub fn requires_approval(&self) -> bool {
        self.risk.requires_approval()
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum Verdict {
    Approve,
    /// El usuario edita el contenido propuesto antes de aplicarlo.
    Edit(String),
    Reject,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ApprovalDecision {
    pub proposal_id: ProposalId,
    pub verdict: Verdict,
    pub actor: String,
    pub at: String,
}

impl ApprovalDecision {
    pub fn approve(id: ProposalId, actor: impl Into<String>, at: impl Into<String>) -> Self {
        Self {
            proposal_id: id,
            verdict: Verdict::Approve,
            actor: actor.into(),
            at: at.into(),
        }
    }

    pub fn reject(id: ProposalId, actor: impl Into<String>, at: impl Into<String>) -> Self {
        Self {
            proposal_id: id,
            verdict: Verdict::Reject,
            actor: actor.into(),
            at: at.into(),
        }
    }

    pub fn is_rejection(&self) -> bool {
        matches!(self.verdict, Verdict::Reject)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn diff(before: &str, after: &str) -> Diff {
        Diff {
            unified: format!("-{before}\n+{after}"),
            before: before.into(),
            after: after.into(),
        }
    }

    #[test]
    fn only_high_impact_requires_approval() {
        let low = ChangeProposal::new(
            ProposalId::new("p1"),
            ChangeTarget::Artifact(ArtifactKind::Agents),
            diff("a", "b"),
            RiskLevel::Low,
            "typo",
            ProposalOrigin::Refinement,
        );
        let high = ChangeProposal::new(
            ProposalId::new("p2"),
            ChangeTarget::Artifact(ArtifactKind::Context),
            diff("a", "b"),
            RiskLevel::HighImpact,
            "cambia la arquitectura descrita",
            ProposalOrigin::Refinement,
        );
        assert!(!low.requires_approval());
        assert!(high.requires_approval());
    }

    #[test]
    fn rejection_is_detected() {
        let d = ApprovalDecision::reject(ProposalId::new("p1"), "user", "2026-07-27");
        assert!(d.is_rejection());
    }
}
