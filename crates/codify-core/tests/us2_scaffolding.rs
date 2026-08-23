//! **Corregir un supuesto arrastra su andamiaje** (T055, T056) — FR-011.
//!
//! Es la diferencia entre esta herramienta y la que reemplaza. El `resolve` anterior sustituía
//! el marcador y se iba: si el usuario corregía «Kafka → Temporal», la sección «Topics y
//! particiones» seguía ahí, y los nombres derivados también. El documento quedaba peor que
//! antes, porque ahora era **incoherente además de equivocado** — y con aspecto de revisado.
//!
//! La propagación no se implementa como una lista de reglas de reescritura. Es **estructural**:
//! una propuesta lleva la versión completa del archivo, no un parche sobre el marcador. No
//! existe un camino por el que se pueda tocar solo la mención literal.

mod fakes;

use codify_core::application::ports::{CompletionOutput, ProviderStatus};
use codify_core::application::refine::RefineLoop;
use codify_core::domain::context::{ArtifactKind, ContextArtifact, Segment};
use codify_core::domain::session::{AuthoringSession, Mode, SessionId};
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

/// Contexto con un supuesto **y andamiaje que depende de él**: la sección y el nombre del
/// componente solo tienen sentido si el motor es Kafka.
const CON_ANDAMIAJE: &str = "El motor de mensajería es Kafka.\n\
     ## Topics y particiones\n\
     El servicio `kafka-consumer-pool` lee de tres topics particionados.";

/// Lo que un agente **correcto** devuelve: cambia el supuesto y todo lo que colgaba de él.
const CORREGIDO_ENTERO: &str = "El motor de orquestación es Temporal.\n\
     ## Workflows y actividades\n\
     El servicio `temporal-worker-pool` ejecuta tres workflows.";

fn sesion() -> AuthoringSession {
    let mut s = AuthoringSession::start(SessionId::new("s-scaffold"), "/fake/repo", Mode::Local);
    s.set_locale("es");
    s.put_artifact(
        ContextArtifact::new(ArtifactKind::Context, "es").with_segments(vec![Segment::grounded(
            CON_ANDAMIAJE,
            vec!["README.md".into()],
            vec!["el andamiaje del README".into()],
        )]),
    );
    s
}

fn loop_que_responde(after: &str) -> (RefineLoop, Arc<FakeModelProvider>) {
    let provider = Arc::new(FakeModelProvider::local(
        "ollama",
        vec![CompletionOutput::Text(format!(
            r#"{{"proposals":[{{"target":"{}","after":{},"rationale":"corrección del usuario"}}]}}"#,
            ArtifactKind::Context.file_path(),
            serde_json::to_string(after).unwrap(),
        ))],
    ));
    let deps = CoreBuilder::new(Mode::Local)
        .provider(provider.clone())
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[])))
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
    (RefineLoop::new(deps), provider)
}

fn cancel() -> Arc<dyn codify_core::application::ports::Cancellation> {
    use codify_core::application::ports::CancellationFactory;
    FakeCancellationFactory::new().create()
}

/// **FR-011**: corregir el supuesto arrastra la sección y el nombre que dependían de él.
#[tokio::test]
async fn correcting_an_assumption_drags_the_scaffolding_that_depended_on_it() {
    let (refine, _) = loop_que_responde(CORREGIDO_ENTERO);
    let mut s = sesion();

    refine
        .submit_message(&mut s, "No usamos Kafka, es Temporal.", cancel())
        .await
        .unwrap();

    let resultado = s.artifacts()[0].render();
    assert!(
        resultado.contains("Temporal"),
        "el supuesto no se corrigió: {resultado}"
    );
    assert!(
        !resultado.contains("Kafka"),
        "quedó una mención del supuesto viejo: {resultado}"
    );
    assert!(
        !resultado.contains("Topics y particiones"),
        "la SECCIÓN que dependía del supuesto sobrevivió: el documento queda incoherente \
         además de corregido, que es peor — parece revisado.\n{resultado}"
    );
    assert!(
        !resultado.contains("kafka-consumer-pool"),
        "el NOMBRE derivado del supuesto sobrevivió: {resultado}"
    );
}

/// El mecanismo que lo hace posible: el agente recibe el **contexto completo actual**. Sin él
/// no podría saber qué colgaba del supuesto, y la propagación sería adivinanza.
#[tokio::test]
async fn the_agent_sees_the_whole_current_context() {
    let (refine, provider) = loop_que_responde(CORREGIDO_ENTERO);
    let mut s = sesion();

    refine
        .submit_message(&mut s, "No usamos Kafka.", cancel())
        .await
        .unwrap();

    let peticiones = provider.seen.lock().unwrap();
    let enviado = format!("{:?}", peticiones[0]);
    assert!(
        enviado.contains("Topics y particiones"),
        "al agente no se le mostró el andamiaje: no puede propagar lo que no ve"
    );
    assert!(
        enviado.contains(ArtifactKind::Context.file_path()),
        "el contexto va con su ruta delante para que el agente pueda referirse al archivo"
    );
}

/// La propagación es **estructural**: la propuesta trae el archivo entero, así que no existe
/// un camino por el que se pueda tocar solo la mención literal y dejar lo demás.
#[tokio::test]
async fn a_proposal_carries_the_whole_file_not_a_marker_patch() {
    let (refine, _) = loop_que_responde(CORREGIDO_ENTERO);
    let mut s = sesion();

    let outcome = refine
        .submit_message(&mut s, "No usamos Kafka.", cancel())
        .await
        .unwrap();

    let diff = &outcome.proposals[0].diff;
    assert_eq!(
        diff.before,
        sesion().artifacts()[0].render(),
        "el lado 'antes' es el archivo completo tal y como estaba"
    );
    assert!(
        diff.after.contains("Workflows y actividades"),
        "el lado 'después' es el archivo completo ya coherente"
    );
}

/// Un agente que **solo** corrige la mención literal deja el documento incoherente. El loop no
/// puede evitarlo por sí solo —lo redacta el modelo— pero el resultado tiene que ser
/// **visible** en el diff, no colarse como si estuviera resuelto.
#[tokio::test]
async fn a_partial_correction_still_shows_the_leftovers_in_the_diff() {
    let a_medias = "El motor de mensajería es Temporal.\n\
         ## Topics y particiones\n\
         El servicio `kafka-consumer-pool` lee de tres topics particionados.";
    let (refine, _) = loop_que_responde(a_medias);
    let mut s = sesion();

    let outcome = refine
        .submit_message(&mut s, "No usamos Kafka.", cancel())
        .await
        .unwrap();

    let diff = &outcome.proposals[0].diff;
    assert!(
        diff.after.contains("kafka-consumer-pool"),
        "el resto quedó en el archivo propuesto"
    );
    assert!(
        outcome.proposals[0].requires_approval(),
        "una corrección a medias es alto impacto: TIENE que pasar por revisión humana, que es \
         donde el resto se ve y se puede rechazar"
    );
}
