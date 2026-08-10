//! **Repositorio con contexto previo: actualizar sin sobrescribir** (T041) — quickstart S3.
//!
//! Hasta ahora `generate()` escribía sin mirar. Re-ejecutar sobre un repositorio que ya tenía
//! contexto **lo destruía**, y con él lo que un humano hubiera escrito a mano — sin diff, sin
//! aviso, sin forma de recuperarlo. Para una herramienta cuyo trabajo es *custodiar* el
//! contexto de un proyecto, es el peor fallo posible: el uso repetido, que debería ser lo
//! normal, era destructivo.
//!
//! Lo que estos tests fijan es que un archivo existente **no cambia sin que alguien lo haya
//! visto y decidido**. No hay fusión automática: el sistema no puede saber qué párrafo escribió
//! una persona y cuál generó él, así que no lo adivina — lo enseña y pregunta.

mod fakes;

use codify_core::application::ingest::IngestBudget;
use codify_core::application::ports::{CompletionOutput, ProviderStatus, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::context::ArtifactKind;
use codify_core::domain::session::Mode;
use codify_core::domain::write::WriteOutcome;
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

const README: &str = "# Proyecto\nMotor: Temporal.";

/// Lo que un humano dejó escrito a mano en una corrida anterior. Perder esto sin preguntar es
/// exactamente lo que la US3 viene a impedir.
const CONTEXTO_PREVIO: &str = "Motor: Temporal\n\n\
     ## Notas del equipo\n\
     OJO: el worker de facturación NO usa Temporal, corre en cron por decisión de negocio.";

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
    s
}

fn service(
    writer: Arc<FakeArtifactWriter>,
    prompter: Arc<FakePrompter>,
    risk: Arc<dyn codify_core::domain::ports::RiskClassifier>,
) -> ContextAuthoring {
    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", script())))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            README,
        )])))
        .resolver(Arc::new(FakeReferenceResolver::new()))
        .diff(Arc::new(FakeDiffEngine))
        .risk(risk)
        .prompter(prompter)
        .audit(Arc::new(RecordingAudit::default()))
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

/// Un repositorio que **ya tiene** contexto escrito.
fn writer_con_contexto_previo() -> Arc<FakeArtifactWriter> {
    let w = Arc::new(FakeArtifactWriter::new());
    w.files.lock().unwrap().insert(
        ArtifactKind::Context.file_path().into(),
        CONTEXTO_PREVIO.into(),
    );
    w
}

async fn correr(svc: &ContextAuthoring) {
    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .expect("la sesión arranca");
    svc.join_session(&id).await.expect("la sesión termina");
}

fn contenido(w: &FakeArtifactWriter) -> String {
    w.files
        .lock()
        .unwrap()
        .get(ArtifactKind::Context.file_path())
        .cloned()
        .unwrap_or_default()
}

// ---------------------------------------------------------------------------
// S3.1 — se propone como diff, no se reemplaza en silencio
// ---------------------------------------------------------------------------

#[tokio::test]
async fn existing_context_is_proposed_as_a_diff_not_replaced_silently() {
    let writer = writer_con_contexto_previo();
    let prompter = Arc::new(FakePrompter::approving());
    let svc = service(writer.clone(), prompter.clone(), Arc::new(ConservativeRisk));

    correr(&svc).await;

    assert_eq!(
        prompter.presented.lock().unwrap().len(),
        1,
        "sobrescribir contexto existente TIENE que pasar por una decisión: es la diferencia \
         entre custodiar y pisar"
    );
}

// ---------------------------------------------------------------------------
// S3.2 — rechazar preserva lo que había (FR-014/FR-015)
// ---------------------------------------------------------------------------

#[tokio::test]
async fn rejecting_the_update_keeps_the_human_content_intact() {
    let writer = writer_con_contexto_previo();
    let prompter = Arc::new(FakePrompter::rejecting());
    let svc = service(writer.clone(), prompter.clone(), Arc::new(ConservativeRisk));

    correr(&svc).await;

    let final_ = contenido(&writer);
    assert_eq!(
        final_, CONTEXTO_PREVIO,
        "rechazar tiene que dejar el archivo EXACTAMENTE como estaba"
    );
    assert!(
        final_.contains("worker de facturación NO usa Temporal"),
        "se perdió la nota que había escrito una persona: {final_}"
    );
}

/// El balance tiene que **declarar** que no se escribió, y por qué. Callarlo dejaría al usuario
/// creyendo que su decisión no surtió efecto — o peor, que sí se escribió (FR-017).
#[tokio::test]
async fn a_rejected_update_is_declared_as_skipped_not_silently_dropped() {
    let writer = writer_con_contexto_previo();
    let svc = service(
        writer.clone(),
        Arc::new(FakePrompter::rejecting()),
        Arc::new(ConservativeRisk),
    );

    let id = svc
        .start_session(StartSession {
            repo_root: "/fake/repo".into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .unwrap();
    svc.join_session(&id).await.unwrap();

    let snap = svc.session_state(&id).await.unwrap();
    let registro = snap
        .writes
        .iter()
        .find(|w| w.path == ArtifactKind::Context.file_path())
        .expect("tiene que constar un intento sobre el artefacto de contexto");

    assert!(
        matches!(registro.outcome, WriteOutcome::Skipped(_)),
        "un rechazo se declara como omitido, no se calla: {:?}",
        registro.outcome
    );
    assert!(
        !registro.reached_disk(),
        "una escritura rechazada no puede contar como realizada"
    );
}

// ---------------------------------------------------------------------------
// S3.3 — aprobar sí escribe, y editar aplica el texto del usuario
// ---------------------------------------------------------------------------

#[tokio::test]
async fn approving_the_update_writes_it() {
    let writer = writer_con_contexto_previo();
    let svc = service(
        writer.clone(),
        Arc::new(FakePrompter::approving()),
        Arc::new(ConservativeRisk),
    );

    correr(&svc).await;

    assert_ne!(
        contenido(&writer),
        CONTEXTO_PREVIO,
        "aprobar la actualización tiene que aplicarla"
    );
}

/// **El caso que da sentido a toda la historia**: el usuario ve que la regeneración se comería
/// su nota, la edita para conservarla, y su texto es el que se escribe.
#[tokio::test]
async fn editing_lets_the_user_keep_what_the_agent_would_have_dropped() {
    use codify_core::domain::change::Verdict;

    let conservado = "Motor: Temporal\n\n\
         ## Notas del equipo\n\
         OJO: el worker de facturación NO usa Temporal, corre en cron por decisión de negocio.";
    let writer = writer_con_contexto_previo();
    let svc = service(
        writer.clone(),
        Arc::new(FakePrompter::new(
            Vec::new(),
            Verdict::Edit(conservado.into()),
        )),
        Arc::new(ConservativeRisk),
    );

    correr(&svc).await;

    assert!(
        contenido(&writer).contains("worker de facturación NO usa Temporal"),
        "se aplicó el texto del agente en vez del que el usuario decidió conservar: {}",
        contenido(&writer)
    );
}

// ---------------------------------------------------------------------------
// Sin contexto previo: nada cambia
// ---------------------------------------------------------------------------

/// La primera corrida sobre un repositorio limpio **no puede pedir permiso**: no hay nada que
/// preservar, y preguntar donde no hay riesgo entrena a aprobar sin leer.
#[tokio::test]
async fn a_first_run_on_a_clean_repo_asks_nothing() {
    let writer = Arc::new(FakeArtifactWriter::new());
    let prompter = Arc::new(FakePrompter::approving());
    let svc = service(writer.clone(), prompter.clone(), Arc::new(ConservativeRisk));

    correr(&svc).await;

    assert!(
        prompter.presented.lock().unwrap().is_empty(),
        "no había contexto previo: no hay nada que decidir"
    );
    assert!(
        !contenido(&writer).is_empty(),
        "y el contexto se escribió sin fricción"
    );
}

/// Regenerar y obtener **lo mismo** no es un cambio: no interrumpe.
#[tokio::test]
async fn regenerating_identical_content_is_not_a_change() {
    let writer = Arc::new(FakeArtifactWriter::new());
    let prompter = Arc::new(FakePrompter::approving());

    // Primera corrida: escribe.
    let svc = service(writer.clone(), prompter.clone(), Arc::new(ConservativeRisk));
    correr(&svc).await;
    let primera = contenido(&writer);

    // Segunda corrida con el mismo guion: produce exactamente lo mismo.
    let svc2 = service(writer.clone(), prompter.clone(), Arc::new(ConservativeRisk));
    correr(&svc2).await;

    assert_eq!(contenido(&writer), primera);
    assert!(
        prompter.presented.lock().unwrap().is_empty(),
        "regenerar lo idéntico no puede pedir aprobación: sería ruido puro"
    );
}
