//! **`001`-FR-018** — si el tier de mayor capacidad no está, el sistema degrada **y lo declara**.
//!
//! Degradar sin avisar es peor que fallar: el usuario recibe un contexto de calidad reducida
//! creyéndolo del mismo nivel, y no tiene forma de saberlo. FR-018 pide las dos mitades, y
//! durante mucho tiempo solo estuvo la primera — el enrutado caía al proveedor disponible y un
//! comentario del código llegó a afirmar que la degradación «se declara en la capa de
//! aplicación», donde no la declaraba nadie.

mod fakes;

use codify_core::application::deps::ProviderRegistry;
use codify_core::application::ports::{CompletionOutput, Tier, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::audit::AuditKind;
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::{CoreBuilder, Local};
use fakes::*;
use std::sync::Arc;

const README: &str = "# Proyecto\nMotor: Temporal.";

fn guion() -> Vec<CompletionOutput> {
    let generado = r#"{"segments":[
        {"text":"Motor: Temporal","grounded":["README.md"],"quotes":["Motor: Temporal."]}
    ]}"#;
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
        s.push(CompletionOutput::Text(generado.to_string()));
    }
    s
}

#[test]
fn con_el_tier_pedido_disponible_no_hay_degradacion() {
    let registry = ProviderRegistry::for_mode(
        Mode::Local,
        vec![
            Arc::new(FakeModelProvider::local("cheap", Vec::new())),
            Arc::new(FakeModelProvider::local("heavy", Vec::new()).with_tier(Tier::Heavy)),
        ],
    )
    .expect("registro local válido");

    let elegido = registry.pick(Tier::Heavy);
    assert_eq!(elegido.provider.name(), "heavy");
    assert!(
        elegido.degraded_from.is_none(),
        "había proveedor del tier pedido: nada que declarar"
    );
}

#[test]
fn sin_el_tier_pedido_la_degradacion_queda_registrada_en_el_tipo() {
    let registry = ProviderRegistry::for_mode(
        Mode::Local,
        vec![Arc::new(FakeModelProvider::local("solo-cheap", Vec::new()))],
    )
    .expect("registro local válido");

    let elegido = registry.pick(Tier::Heavy);
    assert_eq!(elegido.provider.name(), "solo-cheap");
    assert_eq!(
        elegido.degraded_from,
        Some(Tier::Heavy),
        "el que enruta tiene que poder saber que degradó, o no puede avisarlo"
    );
}

/// De punta a punta: la degradación llega al usuario, no se queda en el enrutador.
#[tokio::test]
async fn la_degradacion_se_declara_en_la_auditoria_y_en_la_vista() {
    let audit = Arc::new(RecordingAudit::default());
    let deps = CoreBuilder::<Local>::new()
        // Un solo proveedor, de tier Cheap: la generación pide Heavy y no lo hay.
        .provider(Arc::new(FakeModelProvider::local("ollama", guion())))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            README,
        )])))
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
                vec!["fake".into()],
            ),
        )))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .expect("el grafo local debe armarse");

    let svc = ContextAuthoring::new(deps);
    let id = svc
        .start_session(StartSession {
            repo_root: ".".into(),
            mode: Mode::Local,
            locale: Some("es".into()),
        })
        .await
        .expect("la sesión arranca");
    svc.join_session(&id).await.expect("la sesión termina");

    let vista = svc.session_state(&id).await.expect("hay vista");
    assert!(
        vista.tier_degraded,
        "el usuario debe poder saber que el contexto se generó con calidad reducida (FR-018)"
    );

    let eventos = audit.events.lock().unwrap();
    assert!(
        eventos.iter().any(|e| e.kind == AuditKind::TierDegraded),
        "y tiene que quedar auditado: «se avisó» debe poder demostrarse, no solo afirmarse"
    );
}
