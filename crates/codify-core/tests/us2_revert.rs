//! **Deshacer un cambio aplicado sin preguntar** (T055) — FR-008.
//!
//! Lo de bajo riesgo se auto-aplica para no interrumpir por cada nimiedad. El precio de esa
//! comodidad es que el usuario **no dijo que sí**: se le aplicó algo sin consultarle. La
//! compensación es poder deshacerlo, y tiene que ser real — la interfaz lo estuvo anunciando
//! («aplicada… y revertible») antes de que existiera, que es justo la clase de afirmación sin
//! respaldo que este producto se niega a hacer.
//!
//! Deshacer es **solo** para lo auto-aplicado. Lo que pasó por una decisión humana se cambia
//! decidiendo otra vez, no deshaciendo a sus espaldas.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::change::{ProposalId, RiskLevel};
use codify_core::domain::context::ArtifactKind;
use codify_core::domain::session::{Mode, SessionId};
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

const README: &str = "# Proyecto\nMotor: Temporal.";

fn script() -> Vec<CompletionOutput> {
    let generado = r#"{"segments":[{"text":"Motor: Temporal","grounded":["README.md"]}]}"#;
    let mut s = vec![CompletionOutput::ToolCalls(vec![ToolCall {
        id: "c1".into(),
        name: "finalize".into(),
        arguments: r#"{"summary":"listo"}"#.into(),
    }])];
    for _ in 0..4 {
        s.push(CompletionOutput::Text(generado.to_string()));
    }
    // La respuesta del turno de refinamiento.
    s.push(CompletionOutput::Text(format!(
        r#"{{"proposals":[{{"target":"{}","after":"Motor: Temporal (reformateado).","rationale":"retoque"}}]}}"#,
        ArtifactKind::Context.file_path()
    )));
    s
}

fn service(risk: Arc<dyn codify_core::domain::ports::RiskClassifier>) -> ContextAuthoring {
    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", script())))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            README,
        )])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(risk)
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

async fn sesion_con_una_propuesta(
    svc: &ContextAuthoring,
) -> (SessionId, codify_core::domain::change::ChangeProposal) {
    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .expect("la sesión arranca");
    svc.join_session(&id).await.expect("la sesión termina");

    let proposals = svc
        .submit_message(&id, "reformatea el contexto")
        .await
        .expect("el refinamiento produce una propuesta");
    assert_eq!(proposals.len(), 1, "el guion produce exactamente una");
    (id, proposals.into_iter().next().unwrap())
}

fn contexto(snap: &codify_core::application::service::SessionSnapshot) -> String {
    snap.artifacts
        .iter()
        .find(|a| a.kind == ArtifactKind::Context)
        .expect("el artefacto de contexto existe")
        .render()
}

#[tokio::test]
async fn undoing_an_auto_applied_change_restores_the_previous_state() {
    let svc = service(Arc::new(AlwaysLowRisk));
    let (id, propuesta) = sesion_con_una_propuesta(&svc).await;

    assert_eq!(propuesta.risk, RiskLevel::Low);
    assert!(
        propuesta.applied,
        "lo de bajo riesgo se aplica sin preguntar"
    );

    let despues = contexto(&svc.session_state(&id).await.unwrap());
    assert!(
        despues.contains("reformateado"),
        "la propuesta se aplicó: {despues}"
    );

    svc.revert_proposal(&id, &propuesta.id)
        .await
        .expect("deshacer lo auto-aplicado tiene que funcionar");

    let restaurado = contexto(&svc.session_state(&id).await.unwrap());
    assert!(
        !restaurado.contains("reformateado"),
        "deshacer no devolvió el archivo a como estaba: {restaurado}"
    );
}

/// Deshacer deja la propuesta **sin aplicar**, no la borra: sigue constando que se propuso.
#[tokio::test]
async fn undoing_marks_it_unapplied_without_erasing_it() {
    let svc = service(Arc::new(AlwaysLowRisk));
    let (id, propuesta) = sesion_con_una_propuesta(&svc).await;

    svc.revert_proposal(&id, &propuesta.id).await.unwrap();

    let pendientes = svc.pending_proposals(&id).await.unwrap();
    assert!(
        pendientes.iter().any(|p| p.id == propuesta.id),
        "tras deshacer, la propuesta vuelve a estar sin aplicar y sigue existiendo"
    );
}

/// Deshacer dos veces no puede aplicar el «antes» dos veces ni fallar en silencio.
#[tokio::test]
async fn undoing_twice_is_refused_the_second_time() {
    let svc = service(Arc::new(AlwaysLowRisk));
    let (id, propuesta) = sesion_con_una_propuesta(&svc).await;

    svc.revert_proposal(&id, &propuesta.id).await.unwrap();
    assert!(
        svc.revert_proposal(&id, &propuesta.id).await.is_err(),
        "deshacer algo que ya no está aplicado tiene que decirlo, no fingir que funcionó"
    );
}

/// **Solo lo auto-aplicado.** Un cambio de alto impacto pasó por una decisión humana; cambiarlo
/// se hace decidiendo otra vez, no deshaciéndolo a espaldas de quien lo aprobó.
#[tokio::test]
async fn a_high_impact_change_cannot_be_undone_behind_the_users_back() {
    let svc = service(Arc::new(ConservativeRisk));
    let (id, propuesta) = sesion_con_una_propuesta(&svc).await;

    assert_eq!(propuesta.risk, RiskLevel::HighImpact);
    assert!(
        svc.revert_proposal(&id, &propuesta.id).await.is_err(),
        "deshacer algo aprobado por una persona sin volver a preguntarle es exactamente lo que \
         el producto no hace"
    );
}

#[tokio::test]
async fn an_unknown_proposal_is_reported_not_ignored() {
    let svc = service(Arc::new(AlwaysLowRisk));
    let (id, _) = sesion_con_una_propuesta(&svc).await;

    assert!(
        svc.revert_proposal(&id, &ProposalId::new("no-existe"))
            .await
            .is_err(),
        "una propuesta inexistente no puede resolverse en silencio"
    );
}
