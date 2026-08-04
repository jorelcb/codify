//! **Cancelación de una sesión en curso** (T013/T014) — FR-022/FR-023, SC-008.
//!
//! Lo que se prueba no es que exista un botón, sino las dos propiedades que hacen cierta la
//! promesa: que cancelar **surta efecto sin esperar** a la llamada al modelo en vuelo, y que
//! el usuario quede sabiendo **qué alcanzó a escribirse**.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::session::{Mode, SessionState};
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;
use std::time::{Duration, Instant};

fn call(id: &str, name: &str, args: &str) -> ToolCall {
    ToolCall {
        id: id.into(),
        name: name.into(),
        arguments: args.into(),
    }
}

/// Guion largo: hay trabajo de sobra por delante cuando se cancele.
fn long_script() -> Vec<CompletionOutput> {
    let mut script = Vec::new();
    for i in 0..20 {
        script.push(CompletionOutput::ToolCalls(vec![call(
            &format!("c{i}"),
            "read_file",
            r#"{"path":"README.md"}"#,
        )]));
    }
    script
}

struct Harness {
    service: Arc<ContextAuthoring>,
    writer: Arc<FakeArtifactWriter>,
    audit: Arc<RecordingAudit>,
}

fn harness(script: Vec<CompletionOutput>, delay: Duration) -> Harness {
    let writer = Arc::new(FakeArtifactWriter::new());
    let audit = Arc::new(RecordingAudit::default());

    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(
            FakeModelProvider::local("ollama", script).with_delay(delay),
        ))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            "# Proyecto\nVer SPEC.md",
        )])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(Arc::new(ConservativeRisk))
        .prompter(Arc::new(FakePrompter::approving()))
        .audit(audit.clone())
        .locale(Arc::new(FixedLocale("es")))
        .clock(Arc::new(FixedClock))
        .writer(writer.clone())
        .discovery(Arc::new(FakeProviderDiscovery(ProviderStatus::reachable(
            "http://localhost:11434",
            vec!["fake".into()],
        ))))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .expect("grafo local válido");

    Harness {
        service: Arc::new(ContextAuthoring::new(deps).with_budget(IngestBudget::new(50, 5, 50))),
        writer,
        audit,
    }
}

/// T013 — cancelar durante la ingesta detiene el loop y la sesión queda en `Cancelled`.
#[tokio::test]
async fn cancelling_during_ingestion_stops_the_loop_and_reports_the_balance() {
    let h = harness(long_script(), Duration::from_millis(30));

    let id = h
        .service
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();

    // Dejar que arranque y cancelar a mitad.
    tokio::time::sleep(Duration::from_millis(60)).await;
    let outcome = h.service.cancel_session(&id).await.unwrap();

    h.service.join_session(&id).await.ok();
    let snapshot = h.service.session_state(&id).await.unwrap();

    assert_eq!(snapshot.state, SessionState::Cancelled, "estado terminal");
    assert!(
        outcome.phase == SessionState::Ingesting || outcome.phase == SessionState::Generating,
        "el balance declara en qué fase se canceló: {:?}",
        outcome.phase
    );

    // El usuario sabe qué pasó con el repositorio sin inspeccionar archivos (SC-008).
    assert_eq!(
        outcome.writes.len(),
        snapshot.writes.len(),
        "el balance del corte y el de la sesión coinciden"
    );
    assert!(
        h.audit
            .events
            .lock()
            .unwrap()
            .iter()
            .any(|e| e.kind == codify_core::domain::audit::AuditKind::SessionCancelled),
        "la cancelación queda auditada"
    );
}

/// T014 — cancelar **durante la llamada al modelo** la aborta sin esperar a que termine.
///
/// Es la propiedad que separa "cancelable" de un eufemismo: con una consulta en puntos de
/// control habría que esperar el retardo completo de la petición en vuelo.
#[tokio::test]
async fn cancelling_aborts_the_in_flight_model_call_without_waiting_for_it() {
    const CALL_DELAY: Duration = Duration::from_secs(5);
    let h = harness(long_script(), CALL_DELAY);

    let id = h
        .service
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();

    // Dentro de la primera llamada al modelo, que tarda 5 s.
    tokio::time::sleep(Duration::from_millis(50)).await;

    let started = Instant::now();
    h.service.cancel_session(&id).await.unwrap();
    h.service.join_session(&id).await.ok();
    let elapsed = started.elapsed();

    assert!(
        elapsed < Duration::from_secs(2),
        "cancelar tardó {elapsed:?}: debe abortar la petición en vuelo, no esperarla"
    );
    assert_eq!(
        h.service.session_state(&id).await.unwrap().state,
        SessionState::Cancelled
    );
}

/// Cancelar antes de que se escriba nada deja un balance vacío — y eso también se declara.
#[tokio::test]
async fn cancelling_early_reports_an_empty_balance_rather_than_silence() {
    let h = harness(long_script(), Duration::from_millis(200));

    let id = h
        .service
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();

    let outcome = h.service.cancel_session(&id).await.unwrap();
    h.service.join_session(&id).await.ok();

    assert!(outcome.writes.is_empty(), "no alcanzó a escribir nada");
    assert!(
        h.writer.files.lock().unwrap().is_empty(),
        "el repositorio quedó intacto"
    );
}
