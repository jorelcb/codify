//! Loop de refinamiento **curado** (T037, T038, T056) — FR-010/FR-012/FR-014/FR-015.
//!
//! Reemplaza la peor experiencia de la herramienta anterior: una cola de prompts modales, uno
//! por marcador, con defaults sesgados. Nadie lee 37 preguntas seguidas; se pulsa Enter. El
//! resultado era un contexto "aprobado" que nadie había mirado.
//!
//! Aquí el usuario **conversa**, el agente propone **diffs**, y la interrupción se reserva para
//! lo que de verdad la merece:
//!
//! - `Low` se auto-aplica, queda visible y es revertible.
//! - `HighImpact` **bloquea** hasta que alguien decida a la vista del diff.
//! - Rechazar deja el archivo exactamente como estaba.
//!
//! El sesgo de la interrupción es deliberado: preguntar por todo entrena a aprobar sin leer, y
//! preguntar por nada escribe sin permiso. El `RiskClassifier` decide dónde cae la línea.

use crate::application::authoring_loop::GatheredSource;
use crate::application::deps::AuthoringDeps;
use crate::application::ports::{Cancellation, CompletionOutput, CompletionRequest, Message, Tier};
use crate::domain::audit::{AuditEvent, AuditKind};
use crate::domain::change::{
    ApprovalDecision, ChangeProposal, ChangeTarget, ProposalId, ProposalOrigin, Verdict,
};
use crate::domain::context::{ArtifactKind, ContextArtifact};
use crate::domain::error::{CoreError, Result};
use crate::domain::session::AuthoringSession;
use serde::Deserialize;
use std::sync::Arc;

const REFINE_SYSTEM_PROMPT: &str = "\
Eres el copiloto de authoring de contexto de un repositorio. El usuario te corrige o te pide \
cambios en lenguaje natural y tú propones la NUEVA VERSIÓN COMPLETA de los archivos afectados.

Reglas:
- Cuando el usuario corrige un supuesto, ajusta TAMBIÉN todo lo que dependía de ese supuesto: \
  nombres, secciones y afirmaciones derivadas. Corregir solo la mención literal deja el \
  documento incoherente.
- No inventes nada que no esté respaldado por el repositorio o por lo que el usuario acaba de \
  decirte.
- Responde SOLO con JSON:
  {\"proposals\":[{\"target\":\"<ruta>\",\"after\":\"<contenido completo>\",\"rationale\":\"<por qué>\"}]}";

#[derive(Deserialize)]
struct RawProposal {
    target: String,
    after: String,
    #[serde(default)]
    rationale: String,
}

#[derive(Deserialize)]
struct RawProposals {
    #[serde(default)]
    proposals: Vec<RawProposal>,
}

/// Lo que produjo un turno de conversación.
#[derive(Debug, Clone)]
pub struct RefineOutcome {
    pub proposals: Vec<ChangeProposal>,
}

impl RefineOutcome {
    /// Las que siguen esperando decisión: lo que la piel tiene que mostrar.
    pub fn pending(&self) -> Vec<&ChangeProposal> {
        self.proposals.iter().filter(|p| !p.applied).collect()
    }
}

pub struct RefineLoop {
    deps: AuthoringDeps,
    material: Vec<GatheredSource>,
}

impl RefineLoop {
    pub fn new(deps: AuthoringDeps) -> Self {
        Self {
            deps,
            material: Vec::new(),
        }
    }

    /// El material que la sesión leyó. Un refinamiento aprobado vuelve a pasar por la misma
    /// comprobación de citas que una generación: aprobar no convierte en verificado lo que no
    /// lo estaba (FR-006a).
    pub fn with_material(mut self, material: Vec<GatheredSource>) -> Self {
        self.material = material;
        self
    }

    fn audit(&self, kind: AuditKind, payload: impl Into<String>) {
        self.deps
            .audit
            .record(AuditEvent::new(self.deps.clock.now_iso(), kind, payload));
    }

    /// Un turno de conversación: el usuario dice algo, salen propuestas.
    pub async fn submit_message(
        &self,
        session: &mut AuthoringSession,
        message: &str,
        cancel: Arc<dyn Cancellation>,
    ) -> Result<RefineOutcome> {
        if cancel.is_cancelled() {
            return Err(CoreError::Cancelled);
        }

        let estado_actual = self.render_current(session);
        let elegido = self.deps.providers.pick(Tier::Heavy);
        if let Some(pedido) = elegido.degraded_from {
            self.audit(
                AuditKind::TierDegraded,
                format!(
                    "sin proveedor de tier {pedido:?}: el refinamiento se sirvió con '{}'",
                    elegido.provider.name()
                ),
            );
        }
        let response = tokio::select! {
            r = elegido.provider.complete(CompletionRequest {
                system: REFINE_SYSTEM_PROMPT.to_string(),
                messages: vec![Message::user(format!(
                    "=== CONTEXTO ACTUAL ===\n{estado_actual}\n\n=== EL USUARIO DICE ===\n{message}"
                ))],
                tools: Vec::new(),
            }) => r?,
            _ = cancel.cancelled() => return Err(CoreError::Cancelled),
        };

        let raw = match response {
            CompletionOutput::Text(t) => t,
            CompletionOutput::ToolCalls(_) => {
                return Err(CoreError::Provider(
                    "el refinamiento espera propuestas, no tool-calls".into(),
                ))
            }
        };

        let mut proposals = Vec::new();
        for (i, raw_p) in parse_proposals(&raw)?.into_iter().enumerate() {
            let Some(kind) = kind_for(&raw_p.target) else {
                // Una ruta desconocida no se resuelve adivinando: se declara y se ignora.
                self.audit(
                    AuditKind::ProposalMade,
                    format!("descartada, ruta desconocida: {}", raw_p.target),
                );
                continue;
            };

            let before = session
                .artifacts()
                .iter()
                .find(|a| a.kind == kind)
                .map(|a| a.render())
                .unwrap_or_default();

            let diff = self.deps.diff.make(&before, &raw_p.after);
            if diff.is_empty() {
                continue; // proponer un no-cambio sería ruido
            }

            let mut proposal = ChangeProposal::new(
                ProposalId::new(format!("{}-{}", session.id().as_str(), i)),
                ChangeTarget::Artifact(kind),
                diff,
                crate::domain::change::RiskLevel::Low, // provisional
                raw_p.rationale,
                ProposalOrigin::Refinement,
            );
            proposal.risk = self.deps.risk.classify(&proposal);
            self.audit(AuditKind::ProposalMade, proposal.id.as_str());

            self.settle(session, &mut proposal, kind).await?;
            proposals.push(proposal);
        }

        Ok(RefineOutcome { proposals })
    }

    /// Decide el destino de una propuesta: auto-aplicar o esperar a un humano.
    async fn settle(
        &self,
        session: &mut AuthoringSession,
        proposal: &mut ChangeProposal,
        kind: ArtifactKind,
    ) -> Result<()> {
        if !proposal.requires_approval() {
            // Bajo riesgo: se aplica sin interrumpir, y queda revertible por diseño
            // (el diff conserva ambos lados).
            self.apply(session, proposal, kind, proposal.diff.after.clone())?;
            return Ok(());
        }

        let decision = self.deps.prompter.present(proposal).await?;
        self.record_decision(&decision);
        match decision.verdict {
            Verdict::Approve => self.apply(session, proposal, kind, proposal.diff.after.clone()),
            // Se aplica **el texto del usuario**. Aplicar el del agente convertiría "editar"
            // en "aprobar con pasos de más".
            Verdict::Edit(texto) => self.apply(session, proposal, kind, texto),
            Verdict::Reject => Ok(()), // el artefacto no se toca: FR-015
        }
    }

    /// Escribe la nueva versión en la sesión. **No** toca el disco: eso lo hace la escritura
    /// de artefactos, para que una propuesta aprobada y una generación sigan el mismo camino.
    fn apply(
        &self,
        session: &mut AuthoringSession,
        proposal: &mut ChangeProposal,
        kind: ArtifactKind,
        contenido: String,
    ) -> Result<()> {
        let locale = session.locale().unwrap_or("en").to_string();
        let segments =
            crate::application::authoring_loop::parse_segments(&contenido, &self.material)
                .unwrap_or_else(|_| vec![refined_segment(&contenido)]);

        session.put_artifact(ContextArtifact::new(kind, locale).with_segments(segments));
        proposal.applied = true;
        self.audit(AuditKind::ProposalApplied, proposal.id.as_str());
        Ok(())
    }

    /// Toda decisión queda auditada, incluida la de rechazar: "no se aplicó nada" tiene que
    /// poder demostrarse, no solo afirmarse.
    fn record_decision(&self, decision: &ApprovalDecision) {
        let veredicto = match &decision.verdict {
            Verdict::Approve => "aprobada",
            Verdict::Edit(_) => "editada por el usuario",
            Verdict::Reject => "rechazada",
        };
        self.audit(
            AuditKind::ApprovalCaptured,
            format!(
                "{}: {veredicto} por {}",
                decision.proposal_id.as_str(),
                decision.actor
            ),
        );
    }

    /// Lo que el agente ve del contexto actual. Va con la ruta delante para que pueda
    /// referirse a los archivos por su nombre.
    fn render_current(&self, session: &AuthoringSession) -> String {
        session
            .artifacts()
            .iter()
            .map(|a| format!("--- {} ---\n{}", a.kind.file_path(), a.render()))
            .collect::<Vec<_>>()
            .join("\n\n")
    }
}

/// El resultado de una edición humana no trae procedencia declarada por el modelo, así que
/// **no puede darse por verificado**: entra como tentativo con su motivo. Marcarlo grounded
/// sería inventar una fuente.
fn refined_segment(contenido: &str) -> crate::domain::context::Segment {
    crate::domain::context::Segment::tentative(
        contenido,
        "proviene del refinamiento conversacional; no se ha verificado contra una fuente",
    )
}

fn kind_for(path: &str) -> Option<ArtifactKind> {
    ArtifactKind::default_set()
        .into_iter()
        .find(|k| k.file_path() == path)
}

fn parse_proposals(raw: &str) -> Result<Vec<RawProposal>> {
    let cleaned = strip_fences(raw);
    let parsed: RawProposals = serde_json::from_str(&cleaned)
        .map_err(|e| CoreError::Provider(format!("respuesta de refinamiento no parseable: {e}")))?;
    Ok(parsed.proposals)
}

fn strip_fences(raw: &str) -> String {
    let t = raw.trim();
    if !t.starts_with("```") {
        return t.to_string();
    }
    t.trim_start_matches("```json")
        .trim_start_matches("```")
        .trim_start()
        .trim_end_matches("```")
        .trim()
        .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn an_unknown_target_is_not_guessed() {
        assert!(kind_for("no/existe.md").is_none());
        assert!(kind_for(ArtifactKind::Context.file_path()).is_some());
    }

    #[test]
    fn tolerates_code_fenced_json() {
        let raw = "```json\n{\"proposals\":[{\"target\":\"a\",\"after\":\"b\",\"rationale\":\"c\"}]}\n```";
        assert_eq!(parse_proposals(raw).unwrap().len(), 1);
    }

    #[test]
    fn an_unparseable_response_is_reported_not_ignored() {
        assert!(parse_proposals("lo siento, no puedo").is_err());
    }
}
