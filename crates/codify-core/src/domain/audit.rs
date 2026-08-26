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
    /// Un artefacto llegó al repositorio (o se intentó): la base de FR-017.
    ArtifactWritten,
    ContradictionDetected,
    IngestBudgetExhausted,
    ProposalMade,
    ProposalApplied,
    ProposalReverted,
    ApprovalCaptured,
    EgressBlocked,
    /// La sesión se canceló; el payload lleva el balance de escrituras.
    SessionCancelled,
    /// La sesión murió, con su motivo como código estable (`002`-FR-028). El payload lleva
    /// además el detalle técnico: sirve para el registro, no para la pantalla.
    SessionFailed,
    /// Qué tier y qué conexión atendieron una tarea (`003`-FR-006/FR-010). Es lo que permite
    /// reconstruir después qué salió del equipo y qué no.
    TaskRouted,
    /// Una conexión cambió de estado: conectada, caducada o revocada (`003`-FR-003).
    ConnectionStateChanged,
    /// No había proveedor del tier pedido y se enrutó a otro (FR-018). Que quede auditado es
    /// lo que permite demostrar que se avisó, en vez de solo afirmarlo.
    TierDegraded,
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
