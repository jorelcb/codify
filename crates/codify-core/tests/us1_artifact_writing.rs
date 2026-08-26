//! **Los artefactos llegan al repositorio** (T016) — FR-017, SC-005.
//!
//! Cierra la deuda que destapó el diseño de 002: el núcleo generaba el contexto y lo dejaba
//! en memoria. Un producto que no entrega su resultado no sirve de nada, y "declarar qué se
//! escribió" no tenía sobre qué operar.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::audit::AuditKind;
use codify_core::infrastructure::composition::{CoreBuilder, Local};
use fakes::*;
use std::sync::Arc;

const README: &str = "# Proyecto\nMotor: Temporal. Ver SPEC.md";

fn call(id: &str, name: &str, args: &str) -> ToolCall {
    ToolCall {
        id: id.into(),
        name: name.into(),
        arguments: args.into(),
    }
}

fn script() -> Vec<CompletionOutput> {
    let generated = r#"{"segments":[{"text":"Motor: Temporal","grounded":["README.md"]}]}"#;
    let mut s = vec![
        CompletionOutput::ToolCalls(vec![call("c1", "read_file", r#"{"path":"README.md"}"#)]),
        CompletionOutput::ToolCalls(vec![call("c2", "finalize", r#"{"summary":"listo"}"#)]),
    ];
    for _ in 0..4 {
        s.push(CompletionOutput::Text(generated.to_string()));
    }
    s
}

fn service(writer: Arc<FakeArtifactWriter>, audit: Arc<RecordingAudit>) -> ContextAuthoring {
    let deps = CoreBuilder::<Local>::new()
        .provider(Arc::new(FakeModelProvider::local("ollama", script())))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            README,
        )])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(Arc::new(ConservativeRisk))
        .prompter(Arc::new(FakePrompter::approving()))
        .audit(audit)
        .locale(Arc::new(FixedLocale("es")))
        .clock(Arc::new(FixedClock))
        .writer(writer)
        .discovery(Arc::new(FakeProviderDiscovery(ProviderStatus::reachable(
            "http://localhost:11434",
            vec!["fake".into()],
        ))))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .unwrap();
    ContextAuthoring::new(deps).with_budget(IngestBudget::new(10, 2, 20))
}

#[tokio::test]
async fn generated_artifacts_reach_the_repository_and_each_write_is_audited() {
    let writer = Arc::new(FakeArtifactWriter::new());
    let audit = Arc::new(RecordingAudit::default());
    let svc = service(writer.clone(), audit.clone());

    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: codify_core::domain::session::Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    svc.join_session(&id).await.unwrap();

    // 1. Los cuatro artefactos por defecto están escritos, en sus rutas de dominio.
    //    Se extrae bajo el candado y se suelta enseguida: nada de guards vivos cruzando awaits.
    let (paths, agents_content) = {
        let files = writer.files.lock().unwrap();
        let paths: Vec<String> = files.keys().cloned().collect();
        let agents = files.get("AGENTS.md").cloned();
        (paths, agents)
    };

    assert_eq!(paths.len(), 4, "escritos: {paths:?}");
    for expected in [
        "AGENTS.md",
        "context/CONTEXT.md",
        "context/DEVELOPMENT_GUIDE.md",
        "context/INTERACTIONS_LOG.md",
    ] {
        assert!(paths.iter().any(|p| p == expected), "falta {expected}");
    }

    // 2. El contenido es el renderizado del artefacto, no un marcador de posición.
    assert!(agents_content.unwrap_or_default().contains("Temporal"));

    // 3. Cada escritura queda auditada: es como la piel se entera (FR-017).
    let written = {
        let events = audit.events.lock().unwrap();
        events
            .iter()
            .filter(|e| e.kind == AuditKind::ArtifactWritten)
            .count()
    };
    assert_eq!(written, 4);

    // 4. La sesión reporta el balance.
    let snapshot = svc.session_state(&id).await.unwrap();
    assert_eq!(snapshot.writes.len(), 4);
    assert!(
        snapshot.writes.iter().all(|w| w.reached_disk()),
        "todas llegaron al disco"
    );
}

/// Un fallo de escritura **no** aborta la sesión: los demás artefactos igual se escriben.
#[tokio::test]
async fn a_failing_write_is_reported_without_dragging_the_others_down() {
    let writer = Arc::new(FakeArtifactWriter::new().failing_on("context/CONTEXT.md"));
    let audit = Arc::new(RecordingAudit::default());
    let svc = service(writer.clone(), audit);

    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: codify_core::domain::session::Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    svc.join_session(&id).await.unwrap();

    let snapshot = svc.session_state(&id).await.unwrap();
    assert_eq!(snapshot.writes.len(), 4, "se intentaron los cuatro");

    let failed: Vec<_> = snapshot
        .writes
        .iter()
        .filter(|w| !w.reached_disk())
        .collect();
    assert_eq!(failed.len(), 1, "solo falló uno");
    assert_eq!(failed[0].path, "context/CONTEXT.md");
    assert!(
        failed[0].summary().contains("falló"),
        "el motivo viaja al usuario: {}",
        failed[0].summary()
    );
    let written_count = { writer.files.lock().unwrap().len() };
    assert_eq!(written_count, 3, "los otros tres sí se escribieron");
}
