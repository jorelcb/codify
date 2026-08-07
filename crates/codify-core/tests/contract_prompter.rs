//! **Contract test del port `Prompter`** (T034) — FR-012/FR-014/FR-015.
//!
//! `Prompter` es el borde humano, y su adapter real **es la piel** (Interface Adapter). Por eso
//! aquí solo corre contra el fake: el adapter real vive en `codify-app` y se prueba allí.
//!
//! Nota sobre "solo `HighImpact` bloquea": esa regla **no la cumple el `Prompter`**, la cumple
//! el loop, que decide a quién presentar. Un `Prompter` al que le presenten un cambio de bajo
//! riesgo responderá igual — no es su trabajo negarse. La regla se asserta donde vive, en
//! `us2_refine.rs`, contando qué propuestas llegaron a `presented`.
//!
//! Lo que sí es contrato del port: que la decisión sea **sobre la propuesta que se presentó**,
//! y que una respuesta editada llegue con el contenido del usuario, no con el del agente.

mod fakes;

use codify_core::application::ports::{Prompter, Question};
use codify_core::domain::change::{
    ChangeProposal, ChangeTarget, Diff, ProposalId, ProposalOrigin, RiskLevel, Verdict,
};
use codify_core::domain::context::ArtifactKind;
use fakes::FakePrompter;

fn propuesta(id: &str) -> ChangeProposal {
    ChangeProposal::new(
        ProposalId::new(id),
        ChangeTarget::Artifact(ArtifactKind::Context),
        Diff {
            unified: "-Kafka\n+Temporal".into(),
            before: "Kafka".into(),
            after: "Temporal".into(),
        },
        RiskLevel::HighImpact,
        "el usuario corrigió el motor",
        ProposalOrigin::Refinement,
    )
}

/// La suite que **todo** adapter de `Prompter` debe pasar.
async fn prompter_contract(prompter: &dyn Prompter, label: &str) {
    let p = propuesta("p-42");
    let decision = prompter
        .present(&p)
        .await
        .unwrap_or_else(|e| panic!("[{label}] presentar no puede fallar: {e}"));

    assert_eq!(
        decision.proposal_id, p.id,
        "[{label}] la decisión tiene que referirse a la propuesta presentada: si no, se podría \
         aplicar un cambio aprobando otro"
    );
    assert!(
        !decision.actor.trim().is_empty(),
        "[{label}] toda decisión declara quién la tomó — es lo que la hace auditable"
    );
    assert!(
        !decision.at.trim().is_empty(),
        "[{label}] toda decisión declara cuándo se tomó"
    );

    let respuesta = prompter
        .ask(Question {
            text: "¿Qué motor de orquestación usa el proyecto?".into(),
            suggestions: vec!["Temporal".into(), "Kafka".into()],
        })
        .await
        .unwrap_or_else(|e| panic!("[{label}] preguntar no puede fallar: {e}"));
    let _ = respuesta; // el contenido lo decide el humano; aquí solo importa que responde
}

#[tokio::test]
async fn contract_holds_for_the_in_memory_fake() {
    prompter_contract(&FakePrompter::approving(), "fake-aprobador").await;
    prompter_contract(&FakePrompter::rejecting(), "fake-rechazador").await;
}

/// Rechazar es una respuesta de primera clase, no un error ni un silencio.
#[tokio::test]
async fn rejection_comes_back_as_a_decision() {
    let decision = FakePrompter::rejecting()
        .present(&propuesta("p-1"))
        .await
        .expect("rechazar no es un fallo");

    assert!(
        decision.is_rejection(),
        "un rechazo tiene que llegar como veredicto, no como Err: distinguir 'el usuario dijo \
         que no' de 'algo se rompió' es lo que permite no reintentar a ciegas"
    );
}

/// Editar conserva **el texto del usuario**. Si el loop se quedara con el del agente, "editar"
/// sería aprobar con pasos extra.
#[tokio::test]
async fn an_edit_carries_the_user_text() {
    let editado = "Motor de orquestación: Temporal (Cadence en los workers legacy).";
    let prompter = FakePrompter::new(Vec::new(), Verdict::Edit(editado.into()));

    let decision = prompter.present(&propuesta("p-7")).await.unwrap();
    match decision.verdict {
        Verdict::Edit(texto) => assert_eq!(texto, editado),
        otro => panic!("se esperaba una edición, llegó {otro:?}"),
    }
}
