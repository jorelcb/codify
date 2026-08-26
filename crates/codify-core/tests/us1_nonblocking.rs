//! **`start_session` no bloquea** (T019) — FR-022.
//!
//! Es lo que permite que la interfaz siga viva durante una sesión de minutos. Ninguna técnica
//! del lado de la piel puede lograrlo: el trabajo ocurre *dentro* de la llamada.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::session::{Mode, SessionState};
use codify_core::infrastructure::composition::{CoreBuilder, Local};
use fakes::*;
use std::sync::Arc;
use std::time::{Duration, Instant};

fn slow_service(delay: Duration) -> ContextAuthoring {
    let script: Vec<CompletionOutput> = (0..10)
        .map(|i| {
            CompletionOutput::ToolCalls(vec![ToolCall {
                id: format!("c{i}"),
                name: "read_file".into(),
                arguments: r#"{"path":"README.md"}"#.into(),
            }])
        })
        .collect();

    let deps = CoreBuilder::<Local>::new()
        .provider(Arc::new(
            FakeModelProvider::local("ollama", script).with_delay(delay),
        ))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            "# Proyecto",
        )])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(Arc::new(ConservativeRisk))
        .prompter(Arc::new(FakePrompter::approving()))
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

    ContextAuthoring::new(deps).with_budget(IngestBudget::new(20, 3, 20))
}

/// Con un modelo que tarda 300 ms por llamada y 10 llamadas por delante, `start_session`
/// debe volver en milisegundos — no en segundos.
#[tokio::test]
async fn start_session_returns_before_the_work_is_done() {
    let svc = slow_service(Duration::from_millis(300));

    let started = Instant::now();
    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    let elapsed = started.elapsed();

    assert!(
        elapsed < Duration::from_millis(200),
        "start_session tardó {elapsed:?}: debe retornar sin esperar al trabajo"
    );

    // Y la sesión ya existe y es consultable mientras trabaja.
    let snapshot = svc.session_state(&id).await.unwrap();
    assert!(
        !snapshot.state.is_terminal(),
        "la sesión sigue en curso: {:?}",
        snapshot.state
    );

    svc.cancel_session(&id).await.ok();
    svc.join_session(&id).await.ok();
}

/// El avance es observable mientras corre: no hay que esperar al final para saber algo.
#[tokio::test]
async fn progress_is_observable_while_the_session_runs() {
    let svc = slow_service(Duration::from_millis(80));

    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();

    tokio::time::sleep(Duration::from_millis(120)).await;
    let mid = svc.session_state(&id).await.unwrap();
    assert!(
        matches!(
            mid.state,
            SessionState::Ingesting | SessionState::Generating
        ),
        "estado observable a mitad de camino: {:?}",
        mid.state
    );

    svc.cancel_session(&id).await.ok();
    svc.join_session(&id).await.ok();
}
