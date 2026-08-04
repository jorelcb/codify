//! Adapters de la piel: implementaciones de los ports que el núcleo necesita del borde.
//!
//! La piel es **driving adapter** (invoca casos de uso) y a la vez implementa los callbacks
//! que el núcleo declara como ports. No contiene lógica de dominio: traduce.

use codify_core::application::ports::{AuditSink, Prompter, Question};
use codify_core::domain::audit::{AuditEvent, AuditKind};
use codify_core::domain::change::{ApprovalDecision, ChangeProposal};
use codify_core::domain::error::{CoreError, Result};
use serde::Serialize;
use tauri::{AppHandle, Emitter};

#[derive(Serialize, Clone)]
pub struct ActivityPayload {
    pub action: String,
    pub target: String,
}

#[derive(Serialize, Clone)]
pub struct UnresolvedPayload {
    pub target: String,
    pub reason: Option<String>,
}

#[derive(Serialize, Clone)]
pub struct StatePayload {
    pub state: String,
}

/// Reenvía el log de auditoría del núcleo a la ventana como eventos.
///
/// Es la costura que hace visible el trabajo del agente (FR-001 de
/// `002-authoring-experience`): el núcleo ya audita cada lectura, cada referencia no resuelta
/// y cada salida bloqueada — la piel solo lo convierte en algo que se puede mirar.
pub struct EventAuditSink {
    app: AppHandle,
}

impl EventAuditSink {
    pub fn new(app: AppHandle) -> Self {
        Self { app }
    }

    /// Los payloads de auditoría con forma `"algo: motivo"` traen el motivo incorporado.
    fn split_reason(payload: &str) -> (String, Option<String>) {
        match payload.split_once(": ") {
            Some((target, reason)) => (target.to_string(), Some(reason.to_string())),
            None => (payload.to_string(), None),
        }
    }

    fn activity(&self, action: &str, target: &str) {
        let _ = self.app.emit(
            "agent.activity",
            ActivityPayload {
                action: action.into(),
                target: target.into(),
            },
        );
    }
}

impl AuditSink for EventAuditSink {
    fn record(&self, event: AuditEvent) {
        match event.kind {
            AuditKind::ReferenceResolved => {
                let _ = self.app.emit(
                    "reference.resolved",
                    ActivityPayload {
                        action: "leído".into(),
                        target: event.payload,
                    },
                );
            }
            AuditKind::ReferenceUnresolved => {
                let (target, reason) = Self::split_reason(&event.payload);
                let _ = self
                    .app
                    .emit("reference.unresolved", UnresolvedPayload { target, reason });
            }
            AuditKind::EgressBlocked => {
                let _ = self.app.emit(
                    "egress.blocked",
                    ActivityPayload {
                        action: "bloqueado".into(),
                        target: event.payload,
                    },
                );
            }
            AuditKind::ArtifactGenerated => self.activity("generado", &event.payload),
            AuditKind::ArtifactWritten => {
                let (target, reason) = Self::split_reason(&event.payload);
                let _ = self.app.emit(
                    "artifact.written",
                    UnresolvedPayload {
                        target,
                        reason, // detalle del registro: bytes, u omitido/falló con su motivo
                    },
                );
            }
            AuditKind::SessionCancelled => {
                let _ = self.app.emit(
                    "session.cancelled",
                    ActivityPayload {
                        action: "cancelada".into(),
                        target: event.payload,
                    },
                );
            }
            AuditKind::ContradictionDetected => {
                self.activity("contradicción entre fuentes", &event.payload)
            }
            AuditKind::IngestBudgetExhausted => {
                self.activity("presupuesto agotado", &event.payload)
            }
            AuditKind::ProposalMade => self.activity("propuesta", &event.payload),
            AuditKind::ProposalApplied => self.activity("aplicado", &event.payload),
            AuditKind::ProposalReverted => self.activity("revertido", &event.payload),
            AuditKind::ApprovalCaptured => self.activity("decisión", &event.payload),
        }
    }
}

/// Prompter provisional. El refinamiento conversacional es **US2**; hasta entonces el núcleo
/// no debe poder pedir decisiones a través de una superficie que aún no existe: fallar
/// explícitamente es preferible a aprobar algo en silencio.
pub struct UnavailablePrompter;

#[async_trait::async_trait]
impl Prompter for UnavailablePrompter {
    async fn ask(&self, _question: Question) -> Result<String> {
        Err(CoreError::Unavailable(
            "el refinamiento conversacional llega en US2".into(),
        ))
    }

    async fn present(&self, _proposal: &ChangeProposal) -> Result<ApprovalDecision> {
        Err(CoreError::Unavailable(
            "la revisión de propuestas llega en US2".into(),
        ))
    }
}

/// Reloj del sistema para los sellos de tiempo de la auditoría.
pub struct SystemClock;

impl codify_core::domain::ports::Clock for SystemClock {
    fn now_iso(&self) -> String {
        // Sin dependencias de fecha: segundos desde epoch, suficiente para ordenar y auditar.
        let secs = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0);
        format!("epoch:{secs}")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn splits_reason_when_the_payload_carries_one() {
        let (target, reason) = EventAuditSink::split_reason("docs/SPEC.md: requiere auth");
        assert_eq!(target, "docs/SPEC.md");
        assert_eq!(reason.as_deref(), Some("requiere auth"));
    }

    #[test]
    fn keeps_payload_intact_when_there_is_no_reason() {
        let (target, reason) = EventAuditSink::split_reason("README.md");
        assert_eq!(target, "README.md");
        assert_eq!(reason, None);
    }
}
