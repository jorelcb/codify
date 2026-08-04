//! **El loop de authoring** — propio y mínimo (decisión raíz D5), dependiente solo de ports.
//!
//! Corrige el fallo raíz medido en la auditoría del codify anterior: allí la generación era
//! *one-shot* sobre el texto de un archivo y no dereferenciaba sus punteros, así que el
//! modelo rellenaba con el prior genérico y afirmaba una arquitectura falsa. Aquí el agente
//! **navega el repo y sigue las referencias** antes de escribir, y todo lo que no puede
//! verificar queda marcado como tentativo — nunca afirmado.

use crate::application::deps::AuthoringDeps;
use crate::application::ingest::{
    ingest_tools, parse_action, AgentAction, IngestBudget, TOOL_FINALIZE,
};
use crate::application::ports::{
    Cancellation, CompletionOutput, CompletionRequest, EntryKind, Message, Tier, ToolSpec,
};
use crate::domain::audit::{AuditEvent, AuditKind};
use crate::domain::context::{ArtifactKind, ContextArtifact, Groundedness, Segment};
use crate::domain::error::{CoreError, Result};
use crate::domain::reference::{Reference, ReferenceOrigin};
use crate::domain::session::{AuthoringSession, SessionState};
use serde::Deserialize;
use std::sync::Arc;

const INGEST_SYSTEM_PROMPT: &str = "\
Eres un agente que reúne material para documentar un proyecto de software.

Tu tarea es LEER el repositorio y SEGUIR las referencias que encuentres (otros archivos del \
repo aludidos desde el README o las specs, y URLs públicas si dispones de la herramienta) \
hasta entender de verdad el proyecto: su propósito, su arquitectura real, su stack y sus \
convenciones.

Reglas innegociables:
- NUNCA inventes el contenido de un documento que no pudiste leer. Si no lo alcanzas, usa \
  note_unresolved.
- NO asumas una arquitectura por lo que 'suele' hacerse. Verifica en las fuentes.
- Prioriza: README, specs, documentos referenciados, manifiestos y, de forma selectiva, los \
  archivos de código que revelen estructura (entrypoints, interfaces, configuración).
- Cuando tengas material suficiente, llama a finalize con un resumen.";

const GENERATE_SYSTEM_PROMPT: &str = "\
Redactas un archivo de contexto para agentes de código, a partir ÚNICAMENTE del material \
reunido que se te entrega.

Devuelve SOLO un objeto JSON con esta forma:
{\"segments\":[
  {\"text\":\"...\",\"grounded\":[\"fuente1\"]},
  {\"text\":\"...\",\"tentative\":\"por qué no pudo verificarse\"},
  {\"text\":\"...\",\"contradiction\":{\"sources\":[\"a\",\"b\"],\"note\":\"en qué se contradicen\"}}
]}

Reglas innegociables:
- Todo lo que afirmes como hecho DEBE ir en un segmento 'grounded' citando la fuente donde lo \
  leíste.
- Lo que no puedas verificar va en 'tentative'. Es preferible un contexto con huecos \
  declarados a uno completo e inventado.
- Si dos fuentes se contradicen, emite un segmento 'contradiction': señálalo, NO elijas en \
  silencio.";

/// Material reunido durante la ingesta, con su procedencia.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct GatheredSource {
    pub id: String,
    pub content: String,
}

/// Resultado de la ingesta. `omitted` y `budget_exhausted` existen para poder **declarar**
/// lo que quedó fuera: el truncamiento silencioso está prohibido.
#[derive(Debug, Clone, Default)]
pub struct IngestOutcome {
    pub gathered: Vec<GatheredSource>,
    pub unresolved: Vec<Reference>,
    pub omitted: Vec<String>,
    pub budget_exhausted: bool,
    pub summary: String,
    pub interview_mode: bool,
}

impl IngestOutcome {
    pub fn source_ids(&self) -> Vec<String> {
        self.gathered.iter().map(|g| g.id.clone()).collect()
    }
}

// ---------------------------------------------------------------------------
// Parseo de la respuesta de generación (función pura, testeable sin proveedor)
// ---------------------------------------------------------------------------

#[derive(Deserialize)]
struct RawContradiction {
    #[serde(default)]
    sources: Vec<String>,
    #[serde(default)]
    note: String,
}

#[derive(Deserialize)]
struct RawSegment {
    #[serde(default)]
    text: String,
    #[serde(default)]
    grounded: Option<Vec<String>>,
    #[serde(default)]
    tentative: Option<String>,
    #[serde(default)]
    contradiction: Option<RawContradiction>,
}

#[derive(Deserialize)]
struct RawArtifact {
    #[serde(default)]
    segments: Vec<RawSegment>,
}

/// Convierte la salida del modelo en segmentos de dominio.
///
/// Política de seguridad: un segmento **sin** procedencia declarada NO se toma por hecho —
/// se degrada a tentativo. El modelo no puede colar afirmaciones sin fuente.
pub fn parse_segments(raw: &str) -> Result<Vec<Segment>> {
    let cleaned = strip_code_fences(raw);
    let parsed: RawArtifact = serde_json::from_str(&cleaned)
        .map_err(|e| CoreError::Provider(format!("respuesta de generación no parseable: {e}")))?;

    Ok(parsed
        .segments
        .into_iter()
        .map(|s| match (s.grounded, s.tentative, s.contradiction) {
            (_, _, Some(c)) => Segment::contradiction(s.text, c.sources, c.note),
            (Some(sources), _, None) if !sources.is_empty() => Segment::grounded(s.text, sources),
            (_, Some(reason), None) => Segment::tentative(s.text, reason),
            _ => Segment::tentative(s.text, "el modelo no declaró procedencia"),
        })
        .collect())
}

fn strip_code_fences(raw: &str) -> String {
    let t = raw.trim();
    if !t.starts_with("```") {
        return t.to_string();
    }
    let without_open = t
        .trim_start_matches("```json")
        .trim_start_matches("```")
        .trim_start();
    without_open.trim_end_matches("```").trim().to_string()
}

// ---------------------------------------------------------------------------
// El loop
// ---------------------------------------------------------------------------

pub struct AuthoringLoop {
    deps: AuthoringDeps,
    budget: IngestBudget,
}

impl AuthoringLoop {
    pub fn new(deps: AuthoringDeps) -> Self {
        Self {
            deps,
            budget: IngestBudget::default(),
        }
    }

    pub fn with_budget(mut self, budget: IngestBudget) -> Self {
        self.budget = budget;
        self
    }

    fn audit(&self, kind: AuditKind, payload: impl Into<String>) {
        self.deps
            .audit
            .record(AuditEvent::new(self.deps.clock.now_iso(), kind, payload));
    }

    /// El modo local no ofrece la herramienta de red: la ausencia es la garantía.
    fn tools(&self) -> Vec<ToolSpec> {
        ingest_tools(!self.deps.mode.is_local())
    }

    /// Pase completo de US1: ingesta dirigida por el agente + generación grounded.
    pub async fn run(
        &self,
        session: &mut AuthoringSession,
        cancel: Arc<dyn Cancellation>,
    ) -> Result<IngestOutcome> {
        let outcome = self.ingest(session, cancel.clone()).await?;

        // Sin material no se genera nada: la piel abre la entrevista. Generar aquí sería
        // exactamente el pecado que este producto viene a corregir (inventar sin fuentes).
        if outcome.interview_mode {
            return Ok(outcome);
        }

        if cancel.is_cancelled() {
            return Err(CoreError::Cancelled);
        }

        session
            .advance_to(SessionState::Generating)
            .map_err(|e| CoreError::Invalid(e.to_string()))?;
        self.generate(session, &outcome, cancel).await?;
        Ok(outcome)
    }

    /// Llamada al modelo compuesta con la señal de cancelación.
    ///
    /// Es el punto que hace cierta la promesa de FR-023: cancelar **aborta la petición en
    /// vuelo** en vez de esperar a que termine. Con una simple consulta en puntos de control
    /// habría que aguantar la llamada entera, que puede durar decenas de segundos.
    async fn complete_or_cancel(
        &self,
        tier: Tier,
        request: CompletionRequest,
        cancel: &Arc<dyn Cancellation>,
    ) -> Result<CompletionOutput> {
        let provider = self.deps.providers.pick(tier);
        tokio::select! {
            result = provider.complete(request) => result,
            _ = cancel.cancelled() => Err(CoreError::Cancelled),
        }
    }

    /// Fase 1: el agente navega el repo y sigue referencias hasta `finalize` o presupuesto.
    pub async fn ingest(
        &self,
        session: &mut AuthoringSession,
        cancel: Arc<dyn Cancellation>,
    ) -> Result<IngestOutcome> {
        let mut budget = self.budget.clone();
        let mut outcome = IngestOutcome::default();

        let repo = self.deps.navigator.describe().await?;

        // Repo vacío ⇒ modo entrevista: ni se falla ni se inventa.
        if repo.requires_interview() {
            outcome.interview_mode = true;
            outcome.summary = "repositorio vacío: se requiere entrevista".into();
            return Ok(outcome);
        }

        if session.locale().is_none() {
            let detected = self.deps.locale.detect(&repo).await;
            session.set_locale(detected);
        }

        let tools = self.tools();
        let mut messages = vec![Message::user(format!(
            "Repositorio a documentar. Señales estructurales: [{}]. Lenguaje detectado: {}. \
             Empieza listando la raíz.",
            repo.structural_signals.join(", "),
            repo.detected_language
                .clone()
                .unwrap_or_else(|| "desconocido".into())
        ))];

        loop {
            // Punto de control: cancelar entre pasos no espera a nada.
            if cancel.is_cancelled() {
                return Err(CoreError::Cancelled);
            }
            if !budget.tick_step() {
                outcome.budget_exhausted = true;
                self.audit(
                    AuditKind::IngestBudgetExhausted,
                    "límite de pasos alcanzado",
                );
                break;
            }

            let response = self
                .complete_or_cancel(
                    Tier::Cheap,
                    CompletionRequest {
                        system: INGEST_SYSTEM_PROMPT.to_string(),
                        messages: messages.clone(),
                        tools: tools.clone(),
                    },
                    &cancel,
                )
                .await?;

            let calls = match response {
                // Sin tool-calls, el agente ya no explora más.
                CompletionOutput::Text(text) => {
                    outcome.summary = text;
                    break;
                }
                CompletionOutput::ToolCalls(calls) if calls.is_empty() => break,
                CompletionOutput::ToolCalls(calls) => calls,
            };

            let mut finalized = false;
            for call in calls {
                let action = parse_action(&call.name, &call.arguments);
                let observation = self
                    .execute(session, &mut budget, &mut outcome, action)
                    .await;
                messages.push(Message::tool(observation));
                if call.name == TOOL_FINALIZE {
                    finalized = true;
                }
            }
            if finalized {
                break;
            }
        }

        Ok(outcome)
    }

    /// Ejecuta una acción del agente y devuelve la observación que vuelve al modelo.
    async fn execute(
        &self,
        session: &mut AuthoringSession,
        budget: &mut IngestBudget,
        outcome: &mut IngestOutcome,
        action: AgentAction,
    ) -> String {
        match action {
            AgentAction::ListRepo { path } => match self.deps.navigator.list(&path).await {
                Ok(entries) => {
                    let listing: Vec<String> = entries
                        .iter()
                        .map(|e| {
                            let mark = if matches!(e.kind, EntryKind::Dir) {
                                "/"
                            } else {
                                ""
                            };
                            format!("{}{}", e.path, mark)
                        })
                        .collect();
                    format!("entradas: {}", listing.join(", "))
                }
                Err(e) => format!("error listando '{path}': {e}"),
            },

            AgentAction::ReadFile { path } => {
                if !budget.try_read() {
                    outcome.budget_exhausted = true;
                    outcome.omitted.push(path.clone());
                    self.audit(AuditKind::IngestBudgetExhausted, path.clone());
                    return format!(
                        "presupuesto de lectura agotado: '{path}' queda declarado como omitido"
                    );
                }
                match self.deps.navigator.read(&path).await {
                    Ok(file) => {
                        if file.truncated {
                            outcome.omitted.push(format!("{path} (recortado)"));
                        }
                        outcome.gathered.push(GatheredSource {
                            id: path.clone(),
                            content: file.content.clone(),
                        });
                        session.record_reference(Reference::resolved(
                            ReferenceOrigin::LocalPath(path.clone()),
                            file.content.clone(),
                        ));
                        self.audit(AuditKind::ReferenceResolved, path.clone());
                        let truncated_note = if file.truncated {
                            " [CONTENIDO RECORTADO]"
                        } else {
                            ""
                        };
                        format!("contenido de {path}{truncated_note}:\n{}", file.content)
                    }
                    Err(e) => {
                        let reference = Reference::unresolved(
                            ReferenceOrigin::LocalPath(path.clone()),
                            crate::domain::reference::ReferenceState::Inaccessible,
                        );
                        session.record_reference(reference.clone());
                        outcome.unresolved.push(reference);
                        self.audit(AuditKind::ReferenceUnresolved, path.clone());
                        format!("no se pudo leer '{path}': {e}. NO inventes su contenido.")
                    }
                }
            }

            AgentAction::FetchUrl { url } => {
                if self.deps.mode.is_local() {
                    self.audit(AuditKind::EgressBlocked, url.clone());
                    return format!(
                        "modo local: la salida a '{url}' está bloqueada. Decláralo como no resuelto."
                    );
                }
                if !budget.try_fetch() {
                    outcome.budget_exhausted = true;
                    outcome.omitted.push(url.clone());
                    return format!("presupuesto de red agotado: '{url}' queda omitido");
                }
                let origin = ReferenceOrigin::PublicUrl(url.clone());
                let reference = self.deps.resolver.resolve(&origin).await;
                session.record_reference(reference.clone());

                match reference.content() {
                    Some(content) => {
                        outcome.gathered.push(GatheredSource {
                            id: url.clone(),
                            content: content.to_string(),
                        });
                        self.audit(AuditKind::ReferenceResolved, url.clone());
                        format!("contenido de {url}:\n{content}")
                    }
                    None => {
                        let state = reference.state();
                        outcome.unresolved.push(reference);
                        self.audit(AuditKind::ReferenceUnresolved, url.clone());
                        format!(
                            "'{url}' no se pudo resolver ({state:?}). NO inventes su contenido."
                        )
                    }
                }
            }

            AgentAction::NoteUnresolved { what, reason } => {
                self.audit(AuditKind::ReferenceUnresolved, format!("{what}: {reason}"));
                outcome.omitted.push(format!("{what} ({reason})"));
                "anotado como no resuelto".into()
            }

            AgentAction::Finalize { summary } => {
                outcome.summary = summary;
                "material suficiente registrado".into()
            }

            AgentAction::Unknown { name } => {
                format!("la herramienta '{name}' no existe; usa solo las ofrecidas")
            }
        }
    }

    /// Fase 2: genera cada artefacto a partir **solo** del material reunido.
    pub async fn generate(
        &self,
        session: &mut AuthoringSession,
        outcome: &IngestOutcome,
        cancel: Arc<dyn Cancellation>,
    ) -> Result<()> {
        let locale = session.locale().unwrap_or("en").to_string();
        let material = self.render_material(outcome);

        for kind in ArtifactKind::default_set() {
            if cancel.is_cancelled() {
                return Err(CoreError::Cancelled);
            }
            let response = self
                .complete_or_cancel(
                    Tier::Heavy,
                    CompletionRequest {
                        system: GENERATE_SYSTEM_PROMPT.to_string(),
                        messages: vec![Message::user(format!(
                            "Archivo a redactar: {}\nIdioma de salida: {locale}\n\n\
                             === MATERIAL REUNIDO ===\n{material}",
                            kind.file_path()
                        ))],
                        tools: Vec::new(),
                    },
                    &cancel,
                )
                .await?;

            let raw = match response {
                CompletionOutput::Text(t) => t,
                CompletionOutput::ToolCalls(_) => {
                    return Err(CoreError::Provider(
                        "la fase de generación no admite tool-calls".into(),
                    ))
                }
            };

            let segments = parse_segments(&raw)?;
            let artifact =
                ContextArtifact::new(kind, locale.clone()).with_segments(segments.clone());

            // FR-008: una contradicción entre fuentes se señala y queda auditada.
            for seg in &segments {
                if let Groundedness::Contradiction { sources, note } = &seg.groundedness {
                    self.audit(
                        AuditKind::ContradictionDetected,
                        format!("{}: [{}] {}", kind.file_path(), sources.join(" vs "), note),
                    );
                }
            }

            self.audit(AuditKind::ArtifactGenerated, kind.file_path());

            // El contexto no sirve de nada si no sale de la memoria: se escribe y se declara
            // lo ocurrido, incluso si falló (un fallo aislado no aborta el resto).
            let record = self
                .deps
                .writer
                .write(kind.file_path(), &artifact.render())
                .await;
            self.audit(AuditKind::ArtifactWritten, record.summary());
            session.record_write(record);

            session.put_artifact(artifact);
        }

        Ok(())
    }

    /// Material que ve la fase de generación. Incluye de forma explícita lo **no resuelto**
    /// y lo **omitido**, para que el modelo no rellene esos huecos por su cuenta.
    fn render_material(&self, outcome: &IngestOutcome) -> String {
        let mut out = String::new();
        for source in &outcome.gathered {
            out.push_str(&format!(
                "--- FUENTE: {} ---\n{}\n\n",
                source.id, source.content
            ));
        }
        if !outcome.unresolved.is_empty() {
            out.push_str("--- REFERENCIAS NO RESUELTAS (no inventes su contenido) ---\n");
            for r in &outcome.unresolved {
                out.push_str(&format!("- {} ({:?})\n", r.origin().as_str(), r.state()));
            }
            out.push('\n');
        }
        if !outcome.omitted.is_empty() {
            out.push_str("--- NO LEÍDO (fuera del presupuesto) ---\n");
            for o in &outcome.omitted {
                out.push_str(&format!("- {o}\n"));
            }
        }
        out
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parses_grounded_tentative_and_contradiction_segments() {
        let raw = r#"{"segments":[
            {"text":"Motor: Temporal","grounded":["SPEC-30.md"]},
            {"text":"Métricas por definir","tentative":"ninguna fuente lo cubre"},
            {"text":"Persistencia","contradiction":{"sources":["PRD","SPEC"],"note":"uno dice Postgres, otro event-sourced"}}
        ]}"#;
        let segments = parse_segments(raw).unwrap();
        assert_eq!(segments.len(), 3);
        assert!(segments[0].is_grounded());
        assert!(segments[1].is_unattended_tentative());
        assert!(segments[2].is_contradiction());
    }

    #[test]
    fn segment_without_provenance_is_downgraded_to_tentative() {
        let raw = r#"{"segments":[{"text":"Usa RabbitMQ"}]}"#;
        let segments = parse_segments(raw).unwrap();
        assert!(
            segments[0].is_unattended_tentative(),
            "sin fuente declarada no puede pasar por hecho"
        );
    }

    #[test]
    fn empty_grounded_list_is_not_enough_to_claim_a_fact() {
        let raw = r#"{"segments":[{"text":"Usa Kafka","grounded":[]}]}"#;
        assert!(parse_segments(raw).unwrap()[0].is_unattended_tentative());
    }

    #[test]
    fn tolerates_code_fenced_json() {
        let raw = "```json\n{\"segments\":[{\"text\":\"x\",\"grounded\":[\"a\"]}]}\n```";
        assert!(parse_segments(raw).unwrap()[0].is_grounded());
    }

    #[test]
    fn rejects_unparseable_generation_output() {
        assert!(parse_segments("lo siento, no puedo").is_err());
    }
}
