//! **Diferir lo tentativo de forma explícita** (T037) — FR-014 de `002-authoring-experience`.
//!
//! El producto no puede cerrar una sesión fingiendo que todo está verificado, pero tampoco
//! puede secuestrar al usuario hasta que resuelva cada hueco. La salida es que **decida**: lo
//! difiere a sabiendas, y queda declarado como pendiente.
//!
//! Lo que estos tests fijan es que diferir sea un acto **dirigido** —un fragmento concreto que
//! el usuario ha mirado— y no un interruptor que despache todo de golpe.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::context::ArtifactKind;
use codify_core::domain::session::{Mode, SessionId};
use codify_core::infrastructure::composition::{CoreBuilder, Local};
use fakes::*;
use std::sync::Arc;

const README: &str = "# Proyecto\nMotor: Temporal.";

/// El modelo devuelve un artefacto con un fundamentado y **dos tentativos**.
fn script() -> Vec<CompletionOutput> {
    let generated = r#"{"segments":[
        {"text":"Motor: Temporal","grounded":["README.md"],"quotes":["Motor: Temporal."]},
        {"text":"Métricas por definir","tentative":"no hay fuente que lo respalde"},
        {"text":"Despliegue por definir","tentative":"no se encontró manifiesto"}
    ]}"#;
    // El README se lee de verdad: desde FR-006a, citar una fuente que la sesión nunca abrió
    // ya no produce un segmento fundamentado, y el guion tiene que reflejar un pase real.
    let mut s = vec![
        CompletionOutput::ToolCalls(vec![ToolCall {
            id: "c1".into(),
            name: "read_file".into(),
            arguments: r#"{"path":"README.md"}"#.into(),
        }]),
        CompletionOutput::ToolCalls(vec![ToolCall {
            id: "c2".into(),
            name: "finalize".into(),
            arguments: r#"{"summary":"listo"}"#.into(),
        }]),
    ];
    for _ in 0..4 {
        s.push(CompletionOutput::Text(generated.to_string()));
    }
    s
}

fn service() -> ContextAuthoring {
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
    ContextAuthoring::new(deps).with_budget(IngestBudget::new(10, 2, 20))
}

async fn finished_session(svc: &ContextAuthoring) -> SessionId {
    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .expect("la sesión arranca");
    svc.join_session(&id).await.expect("la sesión termina");
    id
}

fn context_path() -> &'static str {
    ArtifactKind::Context.file_path()
}

/// Encuentra el índice del primer fragmento tentativo sin atender del artefacto de contexto.
async fn first_unattended(svc: &ContextAuthoring, id: &SessionId) -> usize {
    let snap = svc.session_state(id).await.unwrap();
    let artifact = snap
        .artifacts
        .iter()
        .find(|a| a.kind.file_path() == context_path())
        .expect("el artefacto de contexto existe");
    artifact
        .segments
        .iter()
        .position(|s| s.is_unattended_tentative())
        .expect("hay al menos un tentativo sin atender")
}

#[tokio::test]
async fn deferring_a_segment_attends_only_that_one() {
    let svc = service();
    let id = finished_session(&svc).await;

    let antes = svc.session_state(&id).await.unwrap().unattended_tentative;
    assert!(antes >= 2, "el guion produce dos tentativos; había {antes}");

    let index = first_unattended(&svc, &id).await;
    let quedan = svc
        .defer_tentative(&id, context_path(), index)
        .await
        .expect("diferir un tentativo funciona");

    assert_eq!(
        quedan,
        antes - 1,
        "diferir uno atiende exactamente uno: lo contrario permitiría despachar sin mirar"
    );

    // Y el fragmento concreto queda marcado como diferido, no borrado ni reescrito.
    let snap = svc.session_state(&id).await.unwrap();
    let artifact = snap
        .artifacts
        .iter()
        .find(|a| a.kind.file_path() == context_path())
        .unwrap();
    assert!(
        !artifact.segments[index].is_unattended_tentative(),
        "el fragmento diferido ya no cuenta como sin atender"
    );
    assert!(
        !artifact.segments[index].text.is_empty(),
        "diferir no borra el contenido: queda declarado como pendiente"
    );
}

#[tokio::test]
async fn deferring_twice_does_not_double_count() {
    let svc = service();
    let id = finished_session(&svc).await;
    let index = first_unattended(&svc, &id).await;

    let primera = svc
        .defer_tentative(&id, context_path(), index)
        .await
        .unwrap();
    let segunda = svc
        .defer_tentative(&id, context_path(), index)
        .await
        .unwrap();

    assert_eq!(
        primera, segunda,
        "diferir dos veces el mismo fragmento no puede descontar dos"
    );
}

#[tokio::test]
async fn a_grounded_segment_cannot_be_deferred() {
    let svc = service();
    let id = finished_session(&svc).await;

    let snap = svc.session_state(&id).await.unwrap();
    let artifact = snap
        .artifacts
        .iter()
        .find(|a| a.kind.file_path() == context_path())
        .unwrap();
    let grounded = artifact
        .segments
        .iter()
        .position(|s| s.is_grounded())
        .expect("el guion produce un fragmento fundamentado");

    let result = svc.defer_tentative(&id, context_path(), grounded).await;
    assert!(
        result.is_err(),
        "diferir algo verificado sugeriría que hay algo que atender donde no lo hay"
    );
}

#[tokio::test]
async fn unknown_artifact_or_segment_is_reported_not_ignored() {
    let svc = service();
    let id = finished_session(&svc).await;

    assert!(
        svc.defer_tentative(&id, "no/existe.md", 0).await.is_err(),
        "un artefacto inexistente no puede resolverse en silencio"
    );
    assert!(
        svc.defer_tentative(&id, context_path(), 9_999)
            .await
            .is_err(),
        "un índice fuera de rango no puede resolverse en silencio"
    );
}
