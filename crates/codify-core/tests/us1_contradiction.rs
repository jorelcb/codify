//! **FR-008** — cuando dos fuentes se contradicen, el sistema lo **señala**; no elige en
//! silencio ni promedia las versiones.

mod fakes;

use codify_core::application::authoring_loop::GatheredSource;
use codify_core::application::ports::{CompletionOutput, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::audit::AuditKind;
use codify_core::domain::context::Groundedness;
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::{CoreBuilder, Local};
use fakes::*;
use std::sync::Arc;

const PRD: &str = "PRD-00: la persistencia del Run es PostgreSQL.";
const SPEC: &str = "SPEC-30: el Run es event-sourced en Temporal; no hay base de datos propia.";

fn call(id: &str, name: &str, args: &str) -> ToolCall {
    ToolCall {
        id: id.into(),
        name: name.into(),
        arguments: args.into(),
    }
}

#[tokio::test]
async fn contradiction_between_sources_is_surfaced_and_audited() {
    // FR-006b: cada lado del conflicto va con su cita textual, o no se afirma.
    let generated = r#"{"segments":[
        {"text":"Persistencia del Run","contradiction":{"sources":["PRD-00.md","SPEC-30.md"],
         "quotes":["la persistencia del Run es PostgreSQL","el Run es event-sourced en Temporal"],
         "note":"PRD dice PostgreSQL; SPEC dice event-sourced sin base de datos"}}
    ]}"#;

    let mut script = vec![
        CompletionOutput::ToolCalls(vec![call("c1", "read_file", r#"{"path":"PRD-00.md"}"#)]),
        CompletionOutput::ToolCalls(vec![call("c2", "read_file", r#"{"path":"SPEC-30.md"}"#)]),
        CompletionOutput::ToolCalls(vec![call("c3", "finalize", r#"{"summary":"leído"}"#)]),
    ];
    for _ in 0..4 {
        script.push(CompletionOutput::Text(generated.to_string()));
    }

    let audit = Arc::new(RecordingAudit::default());
    let deps = CoreBuilder::<Local>::new()
        .provider(Arc::new(FakeModelProvider::local("ollama", script)))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[
            ("PRD-00.md", PRD),
            ("SPEC-30.md", SPEC),
        ])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(Arc::new(ConservativeRisk))
        .prompter(Arc::new(FakePrompter::approving()))
        .audit(audit.clone())
        .locale(Arc::new(FixedLocale("es")))
        .clock(Arc::new(FixedClock))
        .writer(Arc::new(FakeArtifactWriter::new()))
        .discovery(Arc::new(FakeProviderDiscovery(
            codify_core::application::ports::ProviderStatus::reachable(
                "http://localhost:11434",
                vec!["fake-model".into()],
            ),
        )))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .unwrap();

    let svc = ContextAuthoring::new(deps);
    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    // start_session ya no bloquea (FR-022): esperar el trabajo es explícito.
    svc.join_session(&id).await.unwrap();

    let snapshot = svc.session_state(&id).await.unwrap();
    let artifact = &snapshot.artifacts[0];

    // 1. La contradicción sobrevive como estado de dominio.
    let contradiction = artifact
        .segments
        .iter()
        .find(|s| s.is_contradiction())
        .expect("el segmento de contradicción debe preservarse");

    match &contradiction.groundedness {
        Groundedness::Contradiction {
            sources,
            quotes,
            note,
        } => {
            assert_eq!(sources.len(), 2, "deben citarse ambas fuentes en conflicto");
            assert_eq!(quotes.len(), 2, "y una cita textual por cada una (FR-006b)");
            assert!(note.contains("PostgreSQL"));
        }
        other => panic!("groundedness inesperada: {other:?}"),
    }

    // 2. No se presenta como hecho: el render la marca de forma distinguible.
    let rendered = artifact.render();
    assert!(rendered.contains("CONTRADICCIÓN"));
    assert!(rendered.contains("PRD-00.md"));
    assert!(rendered.contains("SPEC-30.md"));

    // 3. Queda auditada.
    assert!(
        audit
            .events
            .lock()
            .unwrap()
            .iter()
            .any(|e| e.kind == AuditKind::ContradictionDetected),
        "la contradicción debe quedar en el log de auditoría"
    );
}

/// Una contradicción **no** es un segmento grounded: no puede colarse como afirmación firme.
#[tokio::test]
async fn contradiction_is_never_treated_as_grounded_fact() {
    let material = vec![
        GatheredSource::source("a", "la fuente a sostiene una cosa"),
        GatheredSource::source("b", "la fuente b sostiene la contraria"),
    ];
    let raw = r#"{"segments":[{"text":"x","grounded":["a"],"contradiction":{"sources":["a","b"],
        "quotes":["la fuente a sostiene una cosa","la fuente b sostiene la contraria"],"note":"chocan"}}]}"#;
    let segments =
        codify_core::application::authoring_loop::parse_segments(raw, &material).unwrap();
    assert!(segments[0].is_contradiction());
    assert!(
        !segments[0].is_grounded(),
        "la contradicción gana sobre el grounded declarado"
    );
}
