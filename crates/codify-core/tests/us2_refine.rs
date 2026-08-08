//! **Refinamiento conversacional con diffs curados** (T031, T057) — quickstart S2, FR-010/012/014/015.
//!
//! Este es el escenario que justifica toda la user story. La herramienta anterior resolvía los
//! huecos con 37 prompts modales seguidos, uno por marcador, con defaults sesgados: el usuario
//! acababa pulsando Enter sin leer. Lo que se prueba aquí es que **conversar** produce el mismo
//! resultado sin esa cola, y que ningún cambio con sustancia llega al archivo sin que alguien
//! lo haya aprobado a la vista del diff.
//!
//! Las tres reglas que sostienen "curado":
//! - `Low` se auto-aplica y queda revertible.
//! - `HighImpact` **bloquea** hasta decisión explícita.
//! - Rechazar deja el archivo **exactamente** como estaba.

mod fakes;

use codify_core::application::ports::{CompletionOutput, ProviderStatus};
use codify_core::application::refine::RefineLoop;
use codify_core::domain::change::{ChangeTarget, RiskLevel, Verdict};
use codify_core::domain::context::{ArtifactKind, ContextArtifact, Segment};
use codify_core::domain::session::{AuthoringSession, Mode, SessionId, SessionState};
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

/// El contexto de partida: un supuesto **equivocado** (Kafka) y un hueco sin verificar.
fn contexto_con_un_supuesto_erroneo() -> ContextArtifact {
    ContextArtifact::new(ArtifactKind::Context, "es").with_segments(vec![
        Segment::grounded(
            "El motor de orquestación es Kafka.",
            vec!["README.md".into()],
        ),
        Segment::tentative(
            "Las métricas de negocio están por definir.",
            "ninguna fuente las menciona",
        ),
    ])
}

fn sesion() -> AuthoringSession {
    let mut s = AuthoringSession::start(SessionId::new("s-refine"), "/fake/repo", Mode::Local);
    s.set_locale("es");
    s.put_artifact(contexto_con_un_supuesto_erroneo());
    s
}

/// Respuesta del modelo: propone reescribir el artefacto corrigiendo el supuesto.
fn propone(after: &str, rationale: &str) -> Vec<CompletionOutput> {
    vec![CompletionOutput::Text(format!(
        r#"{{"proposals":[{{"target":"{}","after":{},"rationale":{}}}]}}"#,
        ArtifactKind::Context.file_path(),
        serde_json::to_string(after).unwrap(),
        serde_json::to_string(rationale).unwrap(),
    ))]
}

fn loop_con(script: Vec<CompletionOutput>, prompter: Arc<FakePrompter>) -> RefineLoop {
    loop_con_riesgo(script, prompter, Arc::new(ConservativeRisk))
}

fn loop_con_riesgo(
    script: Vec<CompletionOutput>,
    prompter: Arc<FakePrompter>,
    risk: Arc<dyn codify_core::domain::ports::RiskClassifier>,
) -> RefineLoop {
    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", script)))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(risk)
        .prompter(prompter)
        .audit(Arc::new(RecordingAudit::default()))
        .locale(Arc::new(FixedLocale("es")))
        .clock(Arc::new(FixedClock))
        .writer(Arc::new(FakeArtifactWriter::new()))
        .discovery(Arc::new(FakeProviderDiscovery(ProviderStatus::reachable(
            "http://localhost:11434",
            vec!["fake".into()],
        ))))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .unwrap();
    RefineLoop::new(deps)
}

const CORREGIDO: &str =
    "El motor de orquestación es Temporal.\nLas métricas de negocio están por definir.";

// ---------------------------------------------------------------------------
// S2.2 — corregir en lenguaje natural produce un diff, no una cola de prompts
// ---------------------------------------------------------------------------

#[tokio::test]
async fn correcting_in_plain_language_produces_a_reviewable_diff() {
    let prompter = Arc::new(FakePrompter::approving());
    let refine = loop_con(
        propone(CORREGIDO, "el usuario corrigió el motor de orquestación"),
        prompter.clone(),
    );
    let mut s = sesion();

    let outcome = refine
        .submit_message(&mut s, "El motor no es Kafka, es Temporal.", cancel())
        .await
        .expect("el refinamiento no puede fallar por una corrección normal");

    assert_eq!(
        outcome.proposals.len(),
        1,
        "una corrección produce una propuesta, no una pregunta por cada marcador"
    );
    let p = &outcome.proposals[0];
    assert!(
        !p.diff.unified.trim().is_empty(),
        "la propuesta llega como diff revisable, no como texto sustituido a ciegas"
    );
    assert!(
        !p.rationale.trim().is_empty(),
        "el usuario tiene que poder leer POR QUÉ se propone el cambio"
    );
    assert_eq!(p.target, ChangeTarget::Artifact(ArtifactKind::Context));
}

// ---------------------------------------------------------------------------
// S2.3 — HighImpact bloquea; Low se auto-aplica
// ---------------------------------------------------------------------------

#[tokio::test]
async fn a_high_impact_change_blocks_until_someone_decides() {
    let prompter = Arc::new(FakePrompter::approving());
    let refine = loop_con(propone(CORREGIDO, "cambia el motor"), prompter.clone());
    let mut s = sesion();

    let outcome = refine
        .submit_message(&mut s, "El motor es Temporal.", cancel())
        .await
        .unwrap();

    assert_eq!(outcome.proposals[0].risk, RiskLevel::HighImpact);
    assert_eq!(
        prompter.presented.lock().unwrap().len(),
        1,
        "un cambio de alto impacto TIENE que pasar por la decisión de alguien"
    );
}

/// **La regla que el `Prompter` no puede cumplir por sí solo** (ver `contract_prompter.rs`):
/// un cambio de bajo riesgo no interrumpe. Preguntar por todo entrena a aprobar sin leer, que
/// es precisamente el fallo que esta user story corrige.
#[tokio::test]
async fn a_low_risk_change_never_interrupts() {
    let prompter = Arc::new(FakePrompter::approving());
    // Mismo texto salvo espacios: el clasificador conservador lo considera Low.
    let solo_espacios =
        "El motor de orquestación es   Kafka.\n\nLas métricas de negocio están por definir.";
    // Se fija la clasificación a Low a propósito: lo que se prueba aquí es qué hace el LOOP
    // ante un cambio de bajo riesgo, no cómo se decide que algo lo es (eso es
    // `contract_risk_classifier.rs`).
    let refine = loop_con_riesgo(
        propone(solo_espacios, "reformateo"),
        prompter.clone(),
        Arc::new(AlwaysLowRisk),
    );
    let mut s = sesion();

    let outcome = refine
        .submit_message(&mut s, "reformatea el archivo", cancel())
        .await
        .unwrap();

    assert_eq!(outcome.proposals[0].risk, RiskLevel::Low);
    assert!(
        prompter.presented.lock().unwrap().is_empty(),
        "lo de bajo riesgo NO interrumpe: si preguntara por todo, el usuario aprobaría sin leer"
    );
    assert!(
        outcome.proposals[0].applied,
        "lo de bajo riesgo se auto-aplica (FR-010) y queda revertible"
    );
}

// ---------------------------------------------------------------------------
// S2.4 — rechazar deja el archivo intacto (FR-014 / FR-015)
// ---------------------------------------------------------------------------

#[tokio::test]
async fn rejecting_leaves_the_artifact_exactly_as_it_was() {
    let prompter = Arc::new(FakePrompter::rejecting());
    let refine = loop_con(propone(CORREGIDO, "cambia el motor"), prompter.clone());
    let mut s = sesion();
    let antes = s.artifacts()[0].render();

    let outcome = refine
        .submit_message(&mut s, "El motor es Temporal.", cancel())
        .await
        .unwrap();

    assert!(
        !outcome.proposals[0].applied,
        "una propuesta rechazada no se aplica"
    );
    assert_eq!(
        s.artifacts()[0].render(),
        antes,
        "rechazar tiene que dejar el archivo EXACTAMENTE como estaba (FR-015)"
    );
}

/// Editar aplica **el texto del usuario**, no el del agente. Si aplicara el del agente,
/// "editar" sería aprobar con pasos de más.
#[tokio::test]
async fn editing_applies_the_users_text_not_the_agents() {
    let mio = "El motor de orquestación es Temporal (Cadence en los workers legacy).";
    let prompter = Arc::new(FakePrompter::new(Vec::new(), Verdict::Edit(mio.into())));
    let refine = loop_con(propone(CORREGIDO, "cambia el motor"), prompter.clone());
    let mut s = sesion();

    refine
        .submit_message(&mut s, "El motor es Temporal.", cancel())
        .await
        .unwrap();

    assert!(
        s.artifacts()[0].render().contains("Cadence"),
        "se aplicó el texto del agente en vez del que escribió el usuario"
    );
}

// ---------------------------------------------------------------------------
// T057 — aserto de cierre
// ---------------------------------------------------------------------------

/// **FR-013**: no se puede dar por aprobada una sesión con marcadores sin atender. O se
/// resuelven, o se difieren explícitamente — pero nadie cierra en silencio sobre lo no
/// verificado.
#[tokio::test]
async fn a_session_cannot_be_approved_with_unattended_markers() {
    let mut s = sesion();
    s.advance_to(SessionState::Generating).unwrap();
    s.advance_to(SessionState::Refining).unwrap();

    assert!(
        s.unattended_tentative_count() > 0,
        "el fixture parte con un tentativo sin atender"
    );
    assert!(
        s.approve().is_err(),
        "aprobar con marcadores sin atender daría por verificado lo que nadie miró (FR-013)"
    );

    // Al diferirlo explícitamente, cerrar pasa a ser legítimo.
    let mut artifact = s.artifacts()[0].clone();
    for seg in artifact.segments.iter_mut() {
        seg.acknowledge();
    }
    s.put_artifact(artifact);

    assert_eq!(s.unattended_tentative_count(), 0);
    assert!(
        s.approve().is_ok(),
        "una vez diferidos a sabiendas, la sesión puede cerrarse"
    );
}

fn cancel() -> Arc<dyn codify_core::application::ports::Cancellation> {
    use codify_core::application::ports::CancellationFactory;
    FakeCancellationFactory::new().create()
}
