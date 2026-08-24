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
use crate::domain::change::{
    ChangeProposal, ChangeTarget, ProposalId, ProposalOrigin, RiskLevel, Verdict,
};
use crate::domain::context::{ArtifactKind, ContextArtifact, Groundedness, Segment};
use crate::domain::error::{CoreError, Result};
use crate::domain::reference::{Reference, ReferenceOrigin};
use crate::domain::session::{AuthoringSession, SessionState};
use crate::domain::write::WriteRecord;
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
  {\"text\":\"...\",\"grounded\":[\"fuente1\"],\"quotes\":[\"fragmento textual de fuente1\"]},
  {\"text\":\"...\",\"tentative\":\"por qué no pudo verificarse\"},
  {\"text\":\"...\",\"contradiction\":{\"sources\":[\"a\",\"b\"],\
\"quotes\":[\"fragmento de a\",\"fragmento de b\"],\"note\":\"en qué se contradicen\"}}
]}

Reglas innegociables:
- Todo lo que afirmes como hecho DEBE ir en un segmento 'grounded' citando la fuente donde lo \
  leíste Y el fragmento TEXTUAL que lo dice, en 'quotes'.
- Cada cita de 'quotes' debe estar COPIADA LITERALMENTE del material, no parafraseada ni \
  reconstruida de memoria. Se comprueba una por una contra el texto que se te entregó: la que \
  no aparezca degrada el segmento a tentativo.
- No cites lo que la fuente NO dice. Si una fuente niega algo, la cita es la negación.
- Lo que no puedas verificar va en 'tentative'. Es preferible un contexto con huecos \
  declarados a uno completo e inventado.
- Si dos fuentes se contradicen, emite un segmento 'contradiction' con una cita textual de \
  CADA una: señálalo, NO elijas en silencio.";

/// Material reunido durante la ingesta, con su procedencia.
/// De dónde salió una pieza de material.
///
/// La distinción existe porque **leer** y **fundamentar** dejaron de ser lo mismo (FR-006d):
/// un artefacto de una sesión anterior se sigue leyendo —US3 lo necesita para proponer una
/// actualización en vez de sobrescribir— pero no puede respaldar una afirmación.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MaterialOrigin {
    /// Documento del proyecto. Puede respaldar una afirmación.
    Source,
    /// Artefacto que escribió el propio sistema. Se lee; no fundamenta.
    OwnOutput,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct GatheredSource {
    pub id: String,
    pub content: String,
    pub origin: MaterialOrigin,
}

impl GatheredSource {
    /// Material del proyecto, apto para respaldar una afirmación.
    pub fn source(id: impl Into<String>, content: impl Into<String>) -> Self {
        Self {
            id: id.into(),
            content: content.into(),
            origin: MaterialOrigin::Source,
        }
    }

    /// Artefacto propio: entra al material, no a la procedencia.
    pub fn own_output(id: impl Into<String>, content: impl Into<String>) -> Self {
        Self {
            id: id.into(),
            content: content.into(),
            origin: MaterialOrigin::OwnOutput,
        }
    }

    pub fn is_own_output(&self) -> bool {
        self.origin == MaterialOrigin::OwnOutput
    }
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
    quotes: Vec<String>,
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
    quotes: Vec<String>,
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

/// Longitud mínima —ya normalizada— para que una cita demuestre algo. Un fragmento de tres
/// caracteres aparece en cualquier texto: admitirlo convertiría la comprobación en un trámite.
const CITA_MINIMA: usize = 12;

/// Normalización de la comparación (T063).
///
/// Se equiparan mayúsculas y cualquier racha de espacios, tabuladores o saltos de línea. El
/// modelo reproduce el sentido de una frase, no su maquetación: exigir el byte exacto haría que
/// un salto de línea de más invalidara una cita legítima, y el criterio sería inútilmente
/// estricto. Lo que **no** se normaliza son las palabras — sin eso, la comprobación dejaría de
/// distinguir lo que la fuente dice de lo que se le atribuye, que es justo lo que verifica.
fn normalizar(texto: &str) -> String {
    texto
        .to_lowercase()
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
}

/// Las fuentes citadas que de verdad se leyeron, separadas por si pueden respaldar algo.
///
/// El emparejamiento es tolerante en el identificador (`docs/SPEC-30.md` vale por `SPEC-30.md`)
/// porque el modelo abrevia rutas, y ser estricto ahí castigaría una cita correcta por un
/// prefijo. Es tolerancia sobre **qué** fuente, nunca sobre qué dice.
///
/// Devuelve dos listas y no una filtrada porque la segunda hace falta para **explicarse**: si
/// lo único citado fue salida propia, el motivo del degradado tiene que decir eso y no «la
/// fuente no está en el material leído», que sería falso — sí está.
fn fuentes_leidas<'a>(
    citadas: &'a [String],
    material: &[GatheredSource],
) -> (Vec<(&'a str, String)>, Vec<&'a str>) {
    let mut respaldan = Vec::new();
    let mut propias = Vec::new();

    for c in citadas {
        let cn = normalizar(c);
        let encontrada = material.iter().find(|g| {
            let gn = normalizar(&g.id);
            !cn.is_empty() && (gn.contains(&cn) || cn.contains(&gn))
        });
        match encontrada {
            // FR-006d: se leyó, pero es nuestra. No cuenta como procedencia.
            Some(g) if g.is_own_output() => propias.push(c.as_str()),
            Some(g) => respaldan.push((c.as_str(), normalizar(&g.content))),
            None => {}
        }
    }

    (respaldan, propias)
}

/// Convierte la salida del modelo en segmentos de dominio, **comprobando la procedencia**.
///
/// Política de seguridad (FR-006a): que el modelo declare una fuente no verifica nada. Para que
/// un segmento sea `Grounded`, su cita textual debe aparecer en el material que efectivamente se
/// leyó; lo que no se sostiene se degrada a tentativo **declarando el motivo** (FR-006c). Una
/// `Contradiction` exige además cita comprobable de cada lado (FR-006b).
///
/// El origen es el hallazgo F-1: el sistema afirmó `[PRD vs Makefile]` algo sobre un `Makefile`
/// de dos líneas que sí había leído. La defensa anterior cubría la procedencia **ausente**; esta
/// cubre la **falsa**, que es la que engaña.
pub fn parse_segments(raw: &str, material: &[GatheredSource]) -> Result<Vec<Segment>> {
    let cleaned = strip_code_fences(raw);
    let parsed: RawArtifact = serde_json::from_str(&cleaned)
        .map_err(|e| CoreError::Provider(format!("respuesta de generación no parseable: {e}")))?;

    Ok(parsed
        .segments
        .into_iter()
        .map(|s| match (s.grounded, s.tentative, s.contradiction) {
            (_, _, Some(c)) => verificar_contradiccion(s.text, c, material),
            (Some(sources), _, None) if !sources.is_empty() => {
                verificar_grounded(s.text, sources, s.quotes, material)
            }
            (_, Some(reason), None) => Segment::tentative(s.text, reason),
            _ => Segment::tentative(s.text, "el modelo no declaró procedencia"),
        })
        .collect())
}

/// `Grounded` solo si cada cita aparece en alguna de las fuentes citadas **y leídas**.
fn verificar_grounded(
    text: String,
    sources: Vec<String>,
    quotes: Vec<String>,
    material: &[GatheredSource],
) -> Segment {
    let (leidas, propias) = fuentes_leidas(&sources, material);
    if leidas.is_empty() {
        // Distinguir los dos casos importa: uno es citar lo que no se leyó, el otro es citarse
        // a uno mismo. Confundirlos mandaría al usuario a buscar un archivo que sí está.
        return Segment::tentative(
            text,
            if propias.is_empty() {
                format!(
                    "cita fuentes que no están en el material leído: {}",
                    sources.join(", ")
                )
            } else {
                format!(
                    "solo se apoya en artefactos que escribió el propio sistema ({}): se leen, \
                     pero no son fuente de procedencia",
                    propias.join(", ")
                )
            },
        );
    }

    if quotes.is_empty() {
        return Segment::tentative(
            text,
            "declara la fuente pero no la cita: sin fragmento textual no hay nada que comprobar",
        );
    }

    for cita in &quotes {
        let cn = normalizar(cita);
        if cn.chars().count() < CITA_MINIMA {
            return Segment::tentative(
                text,
                format!("la cita «{cita}» es demasiado corta para demostrar la afirmación"),
            );
        }
        if !leidas.iter().any(|(_, contenido)| contenido.contains(&cn)) {
            return Segment::tentative(
                text,
                format!(
                    "la cita «{cita}» no aparece en {}: la fuente se leyó, pero no dice eso",
                    leidas
                        .iter()
                        .map(|(id, _)| *id)
                        .collect::<Vec<_>>()
                        .join(", ")
                ),
            );
        }
    }

    Segment::grounded(text, sources, quotes)
}

/// Una contradicción exige cita comprobable de **cada** fuente en conflicto (FR-006b): si solo
/// se sostiene un lado, lo que hay es una afirmación, no un choque entre fuentes.
fn verificar_contradiccion(
    text: String,
    c: RawContradiction,
    material: &[GatheredSource],
) -> Segment {
    let (leidas, propias) = fuentes_leidas(&c.sources, material);
    if leidas.len() < 2 || leidas.len() != c.sources.len() {
        return Segment::tentative(
            text,
            if propias.is_empty() {
                format!(
                    "una contradicción necesita al menos dos fuentes leídas; se citaron: {}",
                    c.sources.join(", ")
                )
            } else {
                // Contradecirse con uno mismo no es un conflicto entre fuentes: es el bucle
                // que FR-006d viene a cerrar.
                format!(
                    "un lado del conflicto es salida del propio sistema ({}): no hay dos \
                     fuentes que se contradigan",
                    propias.join(", ")
                )
            },
        );
    }

    for (id, contenido) in &leidas {
        let sostenida = c.quotes.iter().any(|q| {
            let qn = normalizar(q);
            qn.chars().count() >= CITA_MINIMA && contenido.contains(&qn)
        });
        if !sostenida {
            return Segment::tentative(
                text,
                format!(
                    "no hay cita comprobable de {id}: el conflicto está afirmado, no demostrado"
                ),
            );
        }
    }

    Segment::contradiction(text, c.sources, c.quotes, c.note)
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
                        // FR-006d: una ruta canónica es un hueco de salida del sistema, así
                        // que lo que hay ahí es nuestro de una sesión anterior. Se lee igual;
                        // lo que no puede es respaldar una afirmación.
                        outcome
                            .gathered
                            .push(if ArtifactKind::is_canonical_path(&path) {
                                GatheredSource::own_output(path.clone(), file.content.clone())
                            } else {
                                GatheredSource::source(path.clone(), file.content.clone())
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
                        // Una URL nunca es un artefacto propio: el sistema escribe a disco,
                        // no publica. Se marca explícitamente en vez de asumirlo.
                        outcome
                            .gathered
                            .push(GatheredSource::source(url.clone(), content.to_string()));
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

            let segments = parse_segments(&raw, &outcome.gathered)?;
            let artifact =
                ContextArtifact::new(kind, locale.clone()).with_segments(segments.clone());

            // FR-008: una contradicción entre fuentes se señala y queda auditada.
            for seg in &segments {
                if let Groundedness::Contradiction { sources, note, .. } = &seg.groundedness {
                    self.audit(
                        AuditKind::ContradictionDetected,
                        format!("{}: [{}] {}", kind.file_path(), sources.join(" vs "), note),
                    );
                }
            }

            self.audit(AuditKind::ArtifactGenerated, kind.file_path());

            // El contexto no sirve de nada si no sale de la memoria: se escribe y se declara
            // lo ocurrido, incluso si falló (un fallo aislado no aborta el resto).
            //
            // Pero **nunca a ciegas**: si ya había contexto, se propone la actualización y se
            // espera decisión (US3). Un producto cuyo trabajo es custodiar el contexto de un
            // proyecto no puede hacer que el uso repetido sea destructivo.
            let contenido = match self.settle_update(session, kind, artifact.render()).await? {
                Some(aprobado) => aprobado,
                None => {
                    // Rechazado: el archivo se queda como estaba, y consta que fue una
                    // decisión, no un olvido.
                    let record = WriteRecord::skipped(
                        kind.file_path(),
                        self.deps.clock.now_iso(),
                        "el usuario rechazó la actualización: se conserva el contenido previo",
                    );
                    self.audit(AuditKind::ArtifactWritten, record.summary());
                    session.record_write(record);
                    continue;
                }
            };

            let record = self.deps.writer.write(kind.file_path(), &contenido).await;
            self.audit(AuditKind::ArtifactWritten, record.summary());
            session.record_write(record);

            session.put_artifact(artifact);
        }

        Ok(())
    }

    /// Decide qué contenido se escribe cuando **ya había** contexto en el repositorio (US3).
    ///
    /// Devuelve `Some(contenido)` si hay que escribir, `None` si el usuario lo rechazó.
    ///
    /// No hay fusión automática, y es deliberado: el sistema **no puede saber** qué párrafo
    /// escribió una persona y cuál generó él. Adivinarlo produciría pérdidas silenciosas justo
    /// donde más duelen. Así que enseña el diff y pregunta — y si el usuario edita, se escribe
    /// lo suyo.
    async fn settle_update(
        &self,
        session: &mut AuthoringSession,
        kind: ArtifactKind,
        generado: String,
    ) -> Result<Option<String>> {
        let Some(previo) = self.deps.writer.read_existing(kind.file_path()).await? else {
            return Ok(Some(generado)); // primera vez: nada que preservar
        };

        let diff = self.deps.diff.make(&previo, &generado);
        if diff.is_empty() {
            // Regenerar lo idéntico no es un cambio. Preguntar aquí sería ruido puro, y el
            // ruido entrena a aprobar sin leer.
            return Ok(Some(generado));
        }

        let mut proposal = ChangeProposal::new(
            ProposalId::new(format!(
                "{}-update-{}",
                session.id().as_str(),
                kind.file_path()
            )),
            ChangeTarget::Artifact(kind),
            diff,
            RiskLevel::Low, // provisional: lo decide el clasificador
            "el repositorio ya tenía contexto: esto es lo que cambiaría",
            ProposalOrigin::Generation,
        );
        proposal.risk = self.deps.risk.classify(&proposal);
        self.audit(AuditKind::ProposalMade, proposal.id.as_str());

        if !proposal.requires_approval() {
            return Ok(Some(generado));
        }

        let decision = self.deps.prompter.present(&proposal).await?;
        self.audit(
            AuditKind::ApprovalCaptured,
            format!("{}: por {}", proposal.id.as_str(), decision.actor),
        );
        Ok(match decision.verdict {
            Verdict::Approve => Some(generado),
            // Se escribe **lo del usuario**: es la vía por la que conserva lo que la
            // regeneración se habría comido.
            Verdict::Edit(texto) => Some(texto),
            Verdict::Reject => None,
        })
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

    /// Material de registro para los tests: lo que la sesión dice haber leído.
    fn material(pares: &[(&str, &str)]) -> Vec<GatheredSource> {
        pares
            .iter()
            .map(|(id, content)| GatheredSource::source(*id, *content))
            .collect()
    }

    /// Material vacío: para los casos donde lo que se comprueba no es la cita.
    fn sin_material() -> Vec<GatheredSource> {
        Vec::new()
    }

    /// Artefactos que escribió el propio sistema: se leen, pero no fundamentan (FR-006d).
    fn propio(pares: &[(&str, &str)]) -> Vec<GatheredSource> {
        pares
            .iter()
            .map(|(id, content)| GatheredSource::own_output(*id, *content))
            .collect()
    }

    // -----------------------------------------------------------------------
    // T066/T067 — la salida propia se lee, pero no respalda (FR-006d)
    // -----------------------------------------------------------------------

    #[test]
    fn a_segment_backed_only_by_our_own_output_is_downgraded() {
        let mut mat = material(&[("docs/SPEC-30.md", "El motor de workflows es Temporal.")]);
        mat.extend(propio(&[(
            "context/CONTEXT.md",
            "La persistencia se resolvió con DynamoDB.",
        )]));

        let raw = r#"{"segments":[
            {"text":"Persistencia en DynamoDB","grounded":["context/CONTEXT.md"],
             "quotes":["La persistencia se resolvió con DynamoDB"]}
        ]}"#;
        let segments = parse_segments(raw, &mat).unwrap();

        assert!(
            segments[0].is_unattended_tentative(),
            "la cita está ahí, pero el documento lo escribimos nosotros: no es evidencia"
        );
        let m = motivo(&segments[0]);
        assert!(
            m.contains("context/CONTEXT.md"),
            "el motivo debe nombrar el artefacto propio, no hablar de fuentes ausentes: {m:?}"
        );
    }

    #[test]
    fn our_own_output_alongside_a_real_source_does_not_poison_the_segment() {
        let mut mat = material(&[("docs/SPEC-30.md", "El motor de workflows es Temporal.")]);
        mat.extend(propio(&[("context/CONTEXT.md", "El motor es Temporal.")]));

        let raw = r#"{"segments":[
            {"text":"Motor: Temporal","grounded":["context/CONTEXT.md","docs/SPEC-30.md"],
             "quotes":["El motor de workflows es Temporal"]}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_grounded(),
            "FR-006d degrada lo que SOLO se apoya en la salida propia, no lo que la menciona: \
             la cita se sostiene en el SPEC"
        );
    }

    #[test]
    fn a_contradiction_cannot_use_our_own_output_as_one_of_its_sides() {
        let mut mat = material(&[("docs/SPEC-30.md", "Persistencia: PostgreSQL 16.")]);
        mat.extend(propio(&[(
            "context/CONTEXT.md",
            "La persistencia se resolvió con DynamoDB.",
        )]));

        let raw = r#"{"segments":[
            {"text":"Persistencia","contradiction":{"sources":["docs/SPEC-30.md","context/CONTEXT.md"],
             "quotes":["Persistencia: PostgreSQL 16","La persistencia se resolvió con DynamoDB"],
             "note":"chocan"}}
        ]}"#;
        assert!(
            !parse_segments(raw, &mat).unwrap()[0].is_contradiction(),
            "contradecirse con uno mismo no es un conflicto entre fuentes (FR-006b + FR-006d)"
        );
    }

    // -----------------------------------------------------------------------
    // T058 — la cita se comprueba contra el material, no se cree
    // -----------------------------------------------------------------------

    #[test]
    fn grounded_survives_when_its_quote_appears_in_the_cited_source() {
        let mat = material(&[("SPEC-30.md", "El motor de workflows es Temporal.")]);
        let raw = r#"{"segments":[
            {"text":"Motor: Temporal","grounded":["SPEC-30.md"],"quotes":["El motor de workflows es Temporal"]}
        ]}"#;
        let segments = parse_segments(raw, &mat).unwrap();
        assert!(
            segments[0].is_grounded(),
            "la cita está en la fuente citada: debe sostenerse"
        );
    }

    #[test]
    fn grounded_is_downgraded_when_its_quote_is_absent_from_the_material() {
        let mat = material(&[("SPEC-30.md", "El motor de workflows es Temporal.")]);
        let raw = r#"{"segments":[
            {"text":"Usa Kafka","grounded":["SPEC-30.md"],"quotes":["El bus de eventos es Kafka"]}
        ]}"#;
        let segments = parse_segments(raw, &mat).unwrap();
        assert!(
            segments[0].is_unattended_tentative(),
            "una cita que no aparece en la fuente no verifica nada"
        );
        assert!(
            motivo(&segments[0]).contains("no aparece"),
            "el motivo debe decir por qué se degradó, no degradar en silencio: {:?}",
            motivo(&segments[0])
        );
    }

    #[test]
    fn grounded_without_quotes_is_no_longer_enough() {
        let mat = material(&[("SPEC-30.md", "El motor de workflows es Temporal.")]);
        let raw = r#"{"segments":[{"text":"Motor: Temporal","grounded":["SPEC-30.md"]}]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_unattended_tentative(),
            "declarar la fuente sin citarla ya no basta (FR-006a)"
        );
    }

    #[test]
    fn a_quote_found_in_another_source_does_not_verify_the_cited_one() {
        let mat = material(&[
            ("SPEC-30.md", "No hay broker de mensajes."),
            ("PRD.md", "La persistencia es DynamoDB."),
        ]);
        let raw = r#"{"segments":[
            {"text":"Persistencia en DynamoDB","grounded":["SPEC-30.md"],"quotes":["La persistencia es DynamoDB"]}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_unattended_tentative(),
            "atribuir a la fuente equivocada también es inventar procedencia"
        );
    }

    #[test]
    fn citing_a_source_that_was_never_read_does_not_verify() {
        let mat = material(&[("SPEC-30.md", "El motor de workflows es Temporal.")]);
        let raw = r#"{"segments":[
            {"text":"Usa Redis","grounded":["ARQUITECTURA.md"],"quotes":["Usa Redis"]}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_unattended_tentative(),
            "no se puede verificar contra material que nunca se leyó"
        );
    }

    // -----------------------------------------------------------------------
    // T063 — normalización: el formato no decide la verdad
    // -----------------------------------------------------------------------

    #[test]
    fn a_quote_that_differs_only_in_formatting_still_counts() {
        let mat = material(&[("SPEC-30.md", "El motor\n  de workflows   es Temporal.")]);
        let raw = r#"{"segments":[
            {"text":"Motor: Temporal","grounded":["SPEC-30.md"],"quotes":["el   MOTOR de workflows es temporal"]}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_grounded(),
            "espacios, saltos y mayúsculas no cambian lo que la fuente dice"
        );
    }

    #[test]
    fn a_trivially_short_quote_does_not_verify() {
        let mat = material(&[("SPEC-30.md", "El motor de workflows es Temporal.")]);
        let raw = r#"{"segments":[
            {"text":"Usa Kafka","grounded":["SPEC-30.md"],"quotes":["el"]}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_unattended_tentative(),
            "una cita trivial aparece en cualquier texto: no demuestra nada"
        );
    }

    // -----------------------------------------------------------------------
    // T059 — la contradicción exige cita de CADA fuente (FR-006b)
    // -----------------------------------------------------------------------

    #[test]
    fn contradiction_holds_when_each_source_has_a_checkable_quote() {
        let mat = material(&[
            ("SPEC-30.md", "El sistema NO es event-sourced."),
            ("PRD.md", "La persistencia es DynamoDB con event sourcing."),
        ]);
        let raw = r#"{"segments":[
            {"text":"Persistencia","contradiction":{"sources":["SPEC-30.md","PRD.md"],
             "quotes":["El sistema NO es event-sourced","La persistencia es DynamoDB con event sourcing"],
             "note":"una lo niega, la otra lo afirma"}}
        ]}"#;
        assert!(
            parse_segments(raw, &mat).unwrap()[0].is_contradiction(),
            "con cita comprobable de cada lado, la contradicción se sostiene"
        );
    }

    #[test]
    fn contradiction_without_a_quote_from_each_source_is_not_asserted() {
        let mat = material(&[
            ("SPEC-30.md", "El sistema NO es event-sourced."),
            ("PRD.md", "La persistencia es DynamoDB."),
        ]);
        let raw = r#"{"segments":[
            {"text":"Persistencia","contradiction":{"sources":["SPEC-30.md","PRD.md"],
             "quotes":["El sistema NO es event-sourced"],
             "note":"chocan"}}
        ]}"#;
        let segments = parse_segments(raw, &mat).unwrap();
        assert!(
            !segments[0].is_contradiction(),
            "sin cita de PRD.md no hay conflicto demostrado, solo afirmado (FR-006b)"
        );
        assert!(segments[0].is_unattended_tentative());
    }

    /// Qué motivo quedó registrado al degradar. Degradar en silencio sería otro defecto.
    fn motivo(seg: &Segment) -> String {
        match &seg.groundedness {
            Groundedness::Tentative { reason, .. } => reason.clone(),
            otro => panic!("se esperaba tentativo, hay {otro:?}"),
        }
    }

    #[test]
    fn parses_grounded_tentative_and_contradiction_segments() {
        let mat = material(&[
            ("SPEC-30.md", "El motor de workflows es Temporal."),
            ("PRD.md", "La persistencia es Postgres."),
        ]);
        let raw = r#"{"segments":[
            {"text":"Motor: Temporal","grounded":["SPEC-30.md"],"quotes":["El motor de workflows es Temporal"]},
            {"text":"Métricas por definir","tentative":"ninguna fuente lo cubre"},
            {"text":"Persistencia","contradiction":{"sources":["PRD.md","SPEC-30.md"],
             "quotes":["La persistencia es Postgres","El motor de workflows es Temporal"],
             "note":"uno dice Postgres, otro no lo menciona"}}
        ]}"#;
        let segments = parse_segments(raw, &mat).unwrap();
        assert_eq!(segments.len(), 3);
        assert!(segments[0].is_grounded());
        assert!(segments[1].is_unattended_tentative());
        assert!(segments[2].is_contradiction());
    }

    #[test]
    fn segment_without_provenance_is_downgraded_to_tentative() {
        let raw = r#"{"segments":[{"text":"Usa RabbitMQ"}]}"#;
        let segments = parse_segments(raw, &sin_material()).unwrap();
        assert!(
            segments[0].is_unattended_tentative(),
            "sin fuente declarada no puede pasar por hecho"
        );
    }

    #[test]
    fn empty_grounded_list_is_not_enough_to_claim_a_fact() {
        let raw = r#"{"segments":[{"text":"Usa Kafka","grounded":[]}]}"#;
        assert!(parse_segments(raw, &sin_material()).unwrap()[0].is_unattended_tentative());
    }

    #[test]
    fn tolerates_code_fenced_json() {
        let mat = material(&[("a", "el contenido citado vive aquí dentro")]);
        let raw = "```json\n{\"segments\":[{\"text\":\"x\",\"grounded\":[\"a\"],\"quotes\":[\"el contenido citado vive aquí\"]}]}\n```";
        assert!(parse_segments(raw, &mat).unwrap()[0].is_grounded());
    }

    #[test]
    fn rejects_unparseable_generation_output() {
        assert!(parse_segments("lo siento, no puedo", &sin_material()).is_err());
    }
}
