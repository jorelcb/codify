//! **US1 — escenario de aceptación** (quickstart S1).
//!
//! Es el test de regresión del fallo raíz medido en la auditoría del codify anterior: el
//! README es delgado y *apunta* a un SPEC hermano; la herramienta debe **seguir el puntero**
//! y fundamentar el contexto en lo que ese SPEC dice ("sin broker, event-sourced sobre
//! Temporal") en vez de rellenar con el prior genérico ("message broker + base de datos").

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::reference::ReferenceState;
use codify_core::domain::session::{Mode, SessionState};
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

const README: &str = "# Ejecución — Workers Python\n\
Componente del proyecto Lumen. La especificación de este servicio vive en \
SPEC-30-Ejecucion-Workers-Python.md. El PRD y la arquitectura son normativos y viven en \
lumen-docs: https://example.test/privado/ARQ-01.md";

const SPEC: &str = "# SPEC-30\n\
Workers en Python sobre **Temporal** que alojan el bounded context de EJECUCIÓN.\n\
No hay broker ni cola de mensajes: el despacho lo hace Temporal por task queues.\n\
El estado del Run no vive en una fila de base de datos: vive como event history.";

fn call(id: &str, name: &str, args: &str) -> ToolCall {
    ToolCall {
        id: id.into(),
        name: name.into(),
        arguments: args.into(),
    }
}

/// Guion del modelo: explora, sigue la referencia al SPEC, intenta la URL privada,
/// finaliza y luego redacta los cuatro artefactos.
fn scripted_model() -> Vec<CompletionOutput> {
    let generated = r#"{"segments":[
        {"text":"Motor de durabilidad: Temporal. Los Runs son workflows, no tareas en una cola.","grounded":["SPEC-30-Ejecucion-Workers-Python.md"]},
        {"text":"El estado vive como event history de Temporal; no hay base de datos propia.","grounded":["SPEC-30-Ejecucion-Workers-Python.md"]},
        {"text":"Stack de observabilidad por definir.","tentative":"ninguna fuente leída lo cubre"}
    ]}"#;

    let mut script = vec![
        CompletionOutput::ToolCalls(vec![call("c1", "list_repo", r#"{"path":""}"#)]),
        CompletionOutput::ToolCalls(vec![call("c2", "read_file", r#"{"path":"README.md"}"#)]),
        CompletionOutput::ToolCalls(vec![call(
            "c3",
            "read_file",
            r#"{"path":"SPEC-30-Ejecucion-Workers-Python.md"}"#,
        )]),
        CompletionOutput::ToolCalls(vec![call(
            "c4",
            "fetch_url",
            r#"{"url":"https://example.test/privado/ARQ-01.md"}"#,
        )]),
        CompletionOutput::ToolCalls(vec![call(
            "c5",
            "finalize",
            r#"{"summary":"material reunido"}"#,
        )]),
    ];
    // Una respuesta de generación por artefacto del conjunto por defecto.
    for _ in 0..4 {
        script.push(CompletionOutput::Text(generated.to_string()));
    }
    script
}

fn service(
    script: Vec<CompletionOutput>,
) -> (
    Arc<ContextAuthoring>,
    Arc<RecordingAudit>,
    Arc<FakeModelProvider>,
) {
    let provider = Arc::new(FakeModelProvider::local("ollama", script));
    let audit = Arc::new(RecordingAudit::default());
    let deps = CoreBuilder::new(Mode::Local)
        .provider(provider.clone())
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[
            ("README.md", README),
            ("SPEC-30-Ejecucion-Workers-Python.md", SPEC),
        ])))
        .resolver(Arc::new(FakeReferenceResolver::new().failing(
            "https://example.test/privado/ARQ-01.md",
            ReferenceState::RequiresAuth,
        )))
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
        .expect("grafo local válido");

    let svc = Arc::new(ContextAuthoring::new(deps).with_budget(IngestBudget::new(10, 3, 20)));
    (svc, audit, provider)
}

/// El corazón de la regresión: el material que llega a la fase de generación **debe**
/// contener lo que decía el documento referenciado, no solo el README delgado.
#[tokio::test]
async fn follows_the_reference_and_grounds_generation_in_the_spec() {
    let (svc, _audit, provider) = service(scripted_model());

    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .expect("la sesión debe completarse");
    // start_session ya no bloquea (FR-022): esperar el trabajo es explícito.
    svc.join_session(&id).await.unwrap();

    let snapshot = svc.session_state(&id).await.unwrap();
    assert_eq!(snapshot.state, SessionState::Generating);
    assert_eq!(
        snapshot.artifacts.len(),
        4,
        "se generan los 4 artefactos por defecto"
    );

    // El prompt de generación vio el contenido del SPEC referenciado.
    let seen = provider.seen.lock().unwrap();
    let generation_prompt = seen
        .iter()
        .rev()
        .find(|r| r.tools.is_empty())
        .expect("debe existir al menos una llamada de generación")
        .messages[0]
        .content
        .clone();

    assert!(
        generation_prompt.contains("no hay broker") || generation_prompt.contains("No hay broker"),
        "el material de generación debe incluir el SPEC referenciado (anti-starvation)"
    );
    assert!(generation_prompt.contains("event history"));
}

#[tokio::test]
async fn generated_context_carries_grounded_and_tentative_segments() {
    let (svc, _audit, _p) = service(scripted_model());
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

    assert!(artifact.segments.iter().any(|s| s.is_grounded()));
    assert!(
        artifact
            .segments
            .iter()
            .any(|s| s.is_unattended_tentative()),
        "lo no verificado queda marcado, no afirmado"
    );

    let rendered = artifact.render();
    assert!(rendered.contains("Temporal"));
    assert!(
        rendered.contains("TENTATIVO"),
        "lo tentativo es distinguible en el render"
    );
    assert!(
        !rendered.to_lowercase().contains("rabbitmq"),
        "no debe aparecer el prior genérico de broker"
    );
}

/// La URL privada no se resuelve: se **declara**, jamás se fabrica su contenido (SC-006).
#[tokio::test]
async fn unresolved_private_reference_is_reported_never_fabricated() {
    let (svc, audit, _p) = service(scripted_model());
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
    // En modo local la salida está bloqueada: la referencia remota no se resuelve.
    assert!(
        audit
            .events
            .lock()
            .unwrap()
            .iter()
            .any(|e| { matches!(e.kind, codify_core::domain::audit::AuditKind::EgressBlocked) }),
        "el intento de salida en modo local debe quedar auditado"
    );
    assert!(
        snapshot
            .artifacts
            .iter()
            .all(|a| !a.render().contains("ARQ-01 dice")),
        "no se inventa el contenido de la referencia no resuelta"
    );
}

/// En modo local el agente ni siquiera recibe la herramienta de red (cero-egress, SC-007).
#[tokio::test]
async fn local_mode_does_not_offer_the_network_tool() {
    let (svc, _audit, provider) = service(scripted_model());
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

    let seen = provider.seen.lock().unwrap();
    let ingest_request = seen
        .iter()
        .find(|r| !r.tools.is_empty())
        .expect("hubo ingesta");
    assert!(
        !ingest_request.tools.iter().any(|t| t.name == "fetch_url"),
        "en modo local la herramienta de red no debe existir"
    );
}

/// El idioma se auto-detecta cuando el usuario no lo fija (FR-019).
#[tokio::test]
async fn locale_is_autodetected_and_can_be_overridden() {
    let (svc, _audit, _p) = service(scripted_model());
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

    assert_eq!(
        svc.session_state(&id).await.unwrap().locale.as_deref(),
        Some("es")
    );

    svc.set_locale(&id, "en".into()).await.unwrap();
    assert_eq!(
        svc.session_state(&id).await.unwrap().locale.as_deref(),
        Some("en")
    );
}

/// Un repositorio vacío deriva a entrevista: ni falla ni inventa.
#[tokio::test]
async fn empty_repository_switches_to_interview_mode() {
    let provider = Arc::new(FakeModelProvider::local("ollama", vec![]));
    let deps = CoreBuilder::new(Mode::Local)
        .provider(provider)
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[])))
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
                vec!["fake-model".into()],
            ),
        )))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .unwrap();

    let svc = ContextAuthoring::new(deps);
    let id = svc
        .start_session(StartSession {
            repo_root: "/vacio".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    // start_session ya no bloquea (FR-022): esperar el trabajo es explícito.
    svc.join_session(&id).await.unwrap();

    assert!(svc.session_state(&id).await.unwrap().interview_mode);
}

/// El presupuesto agotado se **declara**: nada se omite en silencio.
#[tokio::test]
async fn exhausted_budget_is_declared_not_silent() {
    let generated = r#"{"segments":[{"text":"Motor: Temporal","grounded":["README.md"]}]}"#;
    let mut script = vec![
        CompletionOutput::ToolCalls(vec![call("c1", "read_file", r#"{"path":"README.md"}"#)]),
        CompletionOutput::ToolCalls(vec![call(
            "c2",
            "read_file",
            r#"{"path":"SPEC-30-Ejecucion-Workers-Python.md"}"#,
        )]),
        CompletionOutput::ToolCalls(vec![call("c3", "finalize", r#"{"summary":"parcial"}"#)]),
    ];
    for _ in 0..4 {
        script.push(CompletionOutput::Text(generated.to_string()));
    }
    let provider = Arc::new(FakeModelProvider::local("ollama", script));
    let deps = CoreBuilder::new(Mode::Local)
        .provider(provider)
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[
            ("README.md", README),
            ("SPEC-30-Ejecucion-Workers-Python.md", SPEC),
        ])))
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
                vec!["fake-model".into()],
            ),
        )))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .unwrap();

    // Solo 1 lectura permitida: la segunda debe quedar declarada como omitida.
    let svc = ContextAuthoring::new(deps).with_budget(IngestBudget::new(1, 1, 5));
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
    assert!(snapshot.budget_exhausted);
    assert!(
        snapshot.omitted.iter().any(|o| o.contains("SPEC-30")),
        "lo no leído debe declararse: {:?}",
        snapshot.omitted
    );
}
