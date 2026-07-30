//! Log de auditoría **append-only** de la sesión de authoring.
//!
//! Espeja el ethos del `INTERACTIONS_LOG`: los hechos se agregan, nunca se reescriben.
//! Es la base de la trazabilidad del loop (qué leyó el agente, qué propuso, qué se aplicó,
//! qué salida de red se bloqueó).

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AuditKind {
    ReferenceResolved,
    ReferenceUnresolved,
    ArtifactGenerated,
    ContradictionDetected,
    IngestBudgetExhausted,
    ProposalMade,
    ProposalApplied,
    ProposalReverted,
    ApprovalCaptured,
    EgressBlocked,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct AuditEvent {
    pub at: String,
    pub kind: AuditKind,
    pub payload: String,
}

impl AuditEvent {
    pub fn new(at: impl Into<String>, kind: AuditKind, payload: impl Into<String>) -> Self {
        Self {
            at: at.into(),
            kind,
            payload: payload.into(),
        }
    }
}

/// Colección append-only. No expone borrado ni mutación de entradas previas.
#[derive(Debug, Default, Clone)]
pub struct AuditLog {
    events: Vec<AuditEvent>,
}

impl AuditLog {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn append(&mut self, event: AuditEvent) {
        self.events.push(event);
    }

    pub fn events(&self) -> &[AuditEvent] {
        &self.events
    }

    pub fn len(&self) -> usize {
        self.events.len()
    }

    pub fn is_empty(&self) -> bool {
        self.events.is_empty()
    }

    pub fn count_of(&self, kind: AuditKind) -> usize {
        self.events.iter().filter(|e| e.kind == kind).count()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn log_is_append_only_and_counts_by_kind() {
        let mut log = AuditLog::new();
        log.append(AuditEvent::new(
            "t0",
            AuditKind::ArtifactGenerated,
            "AGENTS.md",
        ));
        log.append(AuditEvent::new(
            "t1",
            AuditKind::EgressBlocked,
            "example.test",
        ));
        log.append(AuditEvent::new(
            "t2",
            AuditKind::EgressBlocked,
            "other.test",
        ));

        assert_eq!(log.len(), 3);
        assert_eq!(log.count_of(AuditKind::EgressBlocked), 2);
        // El primer evento sigue intacto: no hay reescritura.
        assert_eq!(log.events()[0].payload, "AGENTS.md");
    }
}
