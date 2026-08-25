//! Adapters de la piel: implementaciones de los ports que el núcleo necesita del borde.
//!
//! La piel es **driving adapter** (invoca casos de uso) y a la vez implementa los callbacks
//! que el núcleo declara como ports. No contiene lógica de dominio: traduce.

use codify_core::application::ports::{AuditSink, Prompter, Question};
use codify_core::domain::audit::{AuditEvent, AuditKind};
use codify_core::domain::change::{ApprovalDecision, ChangeProposal};
use codify_core::domain::error::{CoreError, Result};
use serde::Serialize;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::{AppHandle, Emitter};
use tokio::sync::oneshot;

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
            // `002`-FR-028. El payload trae `codigo: detalle`; la piel traduce el código y
            // guarda el detalle para el registro, nunca para la pantalla.
            AuditKind::SessionFailed => {
                let (target, reason) = Self::split_reason(&event.payload);
                let _ = self
                    .app
                    .emit("session.failed", UnresolvedPayload { target, reason });
            }
            // `001`-FR-018. El payload lleva el detalle técnico —qué tier faltaba, con qué se
            // sirvió— para el registro; la frase que ve el usuario sale del catálogo.
            AuditKind::TierDegraded => {
                let _ = self.app.emit(
                    "tier.degraded",
                    ActivityPayload {
                        action: "degradado".into(),
                        target: event.payload,
                    },
                );
            }
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

/// Puente entre el núcleo y la persona: **este es el adapter del port `Prompter`**.
///
/// El núcleo pide una decisión con `present()` y **espera**. La piel no puede responderle en
/// el acto —la respuesta la da un humano mirando un diff—, así que aquí se parte en dos:
/// se emite `proposal.new` a la ventana y se aguarda en un canal que el comando `decide`
/// resuelve cuando el usuario pulsa.
///
/// Es lo que permite que el turno siga siendo una unidad: `submit_message` retorna cuando
/// todo quedó decidido, no antes. Y es la razón de que el port sea `async`.
pub struct WindowPrompter {
    app: AppHandle,
    pending: PendingDecisions,
}

/// Decisiones que el núcleo está esperando, por id de propuesta.
pub type PendingDecisions = Arc<Mutex<HashMap<String, oneshot::Sender<ApprovalDecision>>>>;

#[derive(Serialize, Clone)]
pub struct ProposalPayload {
    pub id: String,
    pub target: String,
    pub unified: String,
    pub rationale: String,
    pub risk: String,
}

impl WindowPrompter {
    pub fn new(app: AppHandle, pending: PendingDecisions) -> Self {
        Self { app, pending }
    }
}

#[async_trait::async_trait]
impl Prompter for WindowPrompter {
    async fn ask(&self, _question: Question) -> Result<String> {
        // El loop de refinamiento todavía no usa la herramienta `ask_user`: propone cambios,
        // no hace preguntas. Declararlo ausente es más honesto que devolver una respuesta
        // vacía que el agente tomaría por buena.
        Err(CoreError::Unavailable(
            "el agente todavía no formula preguntas: propone cambios".into(),
        ))
    }

    async fn present(&self, proposal: &ChangeProposal) -> Result<ApprovalDecision> {
        let (tx, rx) = oneshot::channel();
        let id = proposal.id.as_str().to_string();
        self.pending
            .lock()
            .map_err(|_| CoreError::Storage("el registro de decisiones se corrompió".into()))?
            .insert(id.clone(), tx);

        let target = match &proposal.target {
            codify_core::domain::change::ChangeTarget::Artifact(k) => k.file_path().to_string(),
            codify_core::domain::change::ChangeTarget::RepoFile(p) => p.clone(),
        };
        let _ = self.app.emit(
            "proposal.new",
            ProposalPayload {
                id: id.clone(),
                target,
                unified: proposal.diff.unified.clone(),
                rationale: proposal.rationale.clone(),
                risk: proposal.risk.code().to_string(),
            },
        );

        // Si el canal se cierra sin respuesta —la ventana se cerró, la sesión se canceló— la
        // propuesta NO se aplica. Ante la duda, no tocar el repositorio del usuario.
        rx.await.map_err(|_| CoreError::Cancelled)
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
