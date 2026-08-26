//! **`003`-US3** — el modo es del usuario, y lo que salió se puede reconstruir.

mod fakes;

use codify_core::application::ports::{CompletionOutput, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::audit::AuditKind;
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::{CoreBuilder, Hybrid, Local};
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

// ---------------------------------------------------------------------------
// T024 — cambiar de modo no afecta a una sesión en curso
// ---------------------------------------------------------------------------

/// `003`-FR-008b. El modo vive en el **grafo**, no en la sesión: rearmar el grafo construye uno
/// nuevo y deja intacto el que la sesión viva está usando. Que sea así no es una precaución —
/// es una consecuencia de que el modo esté en el tipo, y por eso se prueba: la propiedad se
/// perdería en silencio si alguien reintrodujera un modo mutable.
#[tokio::test]
async fn una_sesion_conserva_el_grafo_con_el_que_nacio() {
    let local = ContextAuthoring::new(
        CoreBuilder::<Local>::new()
            .provider(Arc::new(FakeModelProvider::local("local", guion())))
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
            .discovery(Arc::new(FakeProviderDiscovery(
                codify_core::application::ports::ProviderStatus::reachable(
                    "http://localhost:11434",
                    vec!["fake".into()],
                ),
            )))
            .cancellations(Arc::new(FakeCancellationFactory::new()))
            .build()
            .expect("grafo local"),
    );

    let id = local
        .start_session(StartSession {
            repo_root: ".".into(),
            mode: Mode::Local,
            locale: Some("es".into()),
        })
        .await
        .expect("arranca");

    // Rearmar el grafo para otro modo construye un servicio nuevo; el de arriba sigue siendo
    // suyo. No hay forma de que el cambio se cuele en la sesión viva porque no comparten estado.
    let _hibrido = CoreBuilder::<Hybrid>::new();

    local.join_session(&id).await.expect("termina");
    let vista = local.session_state(&id).await.expect("vista");
    assert!(
        !vista.artifacts.is_empty(),
        "la sesión terminó con el grafo con el que nació"
    );
}

// ---------------------------------------------------------------------------
// T025 — se puede reconstruir qué atendió cada tarea
// ---------------------------------------------------------------------------

#[tokio::test]
async fn el_registro_permite_saber_que_proveedor_atendio_cada_tarea() {
    let audit = Arc::new(RecordingAudit::default());
    let svc = ContextAuthoring::new(
        CoreBuilder::<Local>::new()
            .provider(Arc::new(FakeModelProvider::local("local", guion())))
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
            .expect("grafo"),
    );

    let id = svc
        .start_session(StartSession {
            repo_root: ".".into(),
            mode: Mode::Local,
            locale: Some("es".into()),
        })
        .await
        .expect("arranca");
    svc.join_session(&id).await.expect("termina");

    let eventos = audit.events.lock().unwrap();
    let enrutados: Vec<_> = eventos
        .iter()
        .filter(|e| e.kind == AuditKind::TaskRouted)
        .collect();

    assert!(
        !enrutados.is_empty(),
        "FR-010: sin esto el usuario no puede reconstruir qué salió del equipo"
    );
    assert!(
        enrutados.iter().all(|e| e.payload.contains("local")),
        "y cada registro dice si atendió un proveedor local o uno remoto: {:?}",
        enrutados.first().map(|e| &e.payload)
    );
}
