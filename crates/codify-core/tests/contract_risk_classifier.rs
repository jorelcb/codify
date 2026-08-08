//! **Contract test del port `RiskClassifier`** (T033) — FR-012.
//!
//! La política de v1 es **conservadora por defecto**: lo que no se pueda demostrar trivial se
//! trata como alto impacto y bloquea. La asimetría es deliberada. Equivocarse hacia
//! `HighImpact` cuesta una confirmación de más; equivocarse hacia `Low` escribe en el
//! repositorio del usuario algo que nunca aprobó — y eso es exactamente lo que el producto
//! existe para no hacer.
//!
//! El criterio fino de "bajo riesgo" queda para un spec derivado. Lo que esta suite fija es la
//! **dirección del sesgo**, que es lo que no debe cambiarse sin discutirlo.

mod fakes;

use codify_core::domain::change::{
    ChangeProposal, ChangeTarget, Diff, ProposalId, ProposalOrigin, RiskLevel,
};
use codify_core::domain::context::ArtifactKind;
use codify_core::domain::ports::RiskClassifier;
use codify_core::infrastructure::diff::risk::ConservativeRiskClassifier;
use fakes::ConservativeRisk;

fn propuesta(before: &str, after: &str, rationale: &str) -> ChangeProposal {
    ChangeProposal::new(
        ProposalId::new("p"),
        ChangeTarget::Artifact(ArtifactKind::Context),
        Diff {
            unified: format!("-{before}\n+{after}"),
            before: before.into(),
            after: after.into(),
        },
        RiskLevel::Low, // el valor de partida es irrelevante: lo decide el clasificador
        rationale,
        ProposalOrigin::Refinement,
    )
}

/// La suite que **todo** clasificador de v1 debe pasar.
fn risk_classifier_contract(clf: &dyn RiskClassifier, label: &str) {
    // Un no-cambio no puede pedir aprobación: interrumpir sin motivo entrena a aprobar sin leer.
    let vacia = propuesta("mismo texto", "mismo texto", "sin cambios");
    assert_eq!(
        clf.classify(&vacia),
        RiskLevel::Low,
        "[{label}] una propuesta sin cambio real no puede bloquear"
    );

    // Cualquier cambio con sustancia bloquea, por pequeño que parezca.
    let sustancial = propuesta(
        "Motor de orquestación: Kafka.",
        "Motor de orquestación: Temporal.",
        "el usuario corrigió el motor",
    );
    assert_eq!(
        clf.classify(&sustancial),
        RiskLevel::HighImpact,
        "[{label}] v1 es conservadora: lo no trivial bloquea (FR-012)"
    );

    // Reformatear no cambia lo que el documento afirma: interrumpir por espacios en blanco
    // gastaría la atención del usuario justo donde no hay nada que decidir.
    let solo_espacios = propuesta("Motor:   Kafka.", "Motor: Kafka.", "reformateo");
    assert_eq!(
        clf.classify(&solo_espacios),
        RiskLevel::Low,
        "[{label}] un cambio de solo espacios no altera ninguna afirmación"
    );

    // Y el sesgo es el que decide los casos dudosos.
    let dudosa = propuesta("una línea", "una línea distinta", "reformulación menor");
    assert_eq!(
        clf.classify(&dudosa),
        RiskLevel::HighImpact,
        "[{label}] ante la duda, bloquear: equivocarse hacia Low escribe sin aprobación"
    );
}

#[test]
fn contract_holds_for_the_real_adapter() {
    risk_classifier_contract(&ConservativeRiskClassifier, "real");
}

#[test]
fn contract_holds_for_the_in_memory_fake() {
    risk_classifier_contract(&ConservativeRisk, "fake");
}

/// La clasificación **no depende del texto de la justificación**: si dependiera, bastaría con
/// que el modelo escribiera «cambio menor» para saltarse la aprobación.
#[test]
fn the_rationale_cannot_lower_the_risk() {
    let honesta = propuesta("Kafka", "Temporal", "cambia el motor de orquestación");
    let zalamera = propuesta("Kafka", "Temporal", "cambio trivial, sin impacto, seguro");

    assert_eq!(
        ConservativeRiskClassifier.classify(&honesta),
        ConservativeRiskClassifier.classify(&zalamera),
        "el mismo cambio no puede clasificarse distinto según cómo se describa"
    );
}

/// Lo que se clasifica es el **contenido del diff**, no el artefacto al que apunta: cambiar
/// una línea de `AGENTS.md` no es menos serio que cambiarla en `CONTEXT.md`.
#[test]
fn the_target_does_not_change_the_verdict() {
    let mut en_contexto = propuesta("Kafka", "Temporal", "corrección");
    let mut en_agentes = propuesta("Kafka", "Temporal", "corrección");
    en_contexto.target = ChangeTarget::Artifact(ArtifactKind::Context);
    en_agentes.target = ChangeTarget::Artifact(ArtifactKind::Agents);

    assert_eq!(
        ConservativeRiskClassifier.classify(&en_contexto),
        ConservativeRiskClassifier.classify(&en_agentes)
    );
}
