//! **FR-006a/b/c** — verificar es una comprobación del sistema, no una declaración del modelo.
//!
//! Este fichero reproduce el hallazgo **F-1** de la pasada con modelo real del 2026-08-23. El
//! sistema afirmó, como contradicción fundamentada:
//!
//! > `[PRD vs Makefile] «el Makefile solo soporta PostgreSQL 16»`
//!
//! sobre un `Makefile` de **dos líneas** que no menciona ninguna base de datos. La fuente sí se
//! había leído: lo inventado era lo que se le atribuía. La defensa que ya existía cubría la
//! procedencia **ausente** —un segmento sin fuente se degradaba—, y por eso no vio nada: aquí la
//! fuente estaba, y era real. Lo falso era la cita.

mod fakes;

use codify_core::application::authoring_loop::{parse_segments, GatheredSource};
use codify_core::application::ports::{CompletionOutput, ToolCall};
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::context::Groundedness;
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

/// El `Makefile` del fixture, tal cual: dos líneas, ninguna base de datos.
const MAKEFILE: &str = "build:\n\tcargo build --release\n";
const PRD: &str = "PRD-00: la persistencia del Run es DynamoDB.";

fn material() -> Vec<GatheredSource> {
    vec![
        GatheredSource {
            id: "Makefile".into(),
            content: MAKEFILE.into(),
        },
        GatheredSource {
            id: "PRD-00.md".into(),
            content: PRD.into(),
        },
    ]
}

fn motivo(g: &Groundedness) -> String {
    match g {
        Groundedness::Tentative { reason, .. } => reason.clone(),
        otro => panic!("se esperaba tentativo, hay {otro:?}"),
    }
}

/// El caso F-1 exacto: se cita una fuente **leída** atribuyéndole algo que no dice.
#[test]
fn f1_attributing_absent_content_to_a_read_source_is_not_grounded() {
    let raw = r#"{"segments":[
        {"text":"Base de datos","contradiction":{"sources":["PRD-00.md","Makefile"],
         "quotes":["la persistencia del Run es DynamoDB","el Makefile solo soporta PostgreSQL 16"],
         "note":"el PRD dice DynamoDB; el Makefile solo soporta PostgreSQL 16"}}
    ]}"#;

    let segments = parse_segments(raw, &material()).unwrap();

    assert!(
        !segments[0].is_contradiction(),
        "F-1 se reproduce: se afirmó una contradicción con una cita que el Makefile no contiene"
    );
    assert!(
        segments[0].is_unattended_tentative(),
        "lo no comprobado se degrada, no se descarta (FR-006c)"
    );
    let motivo = motivo(&segments[0].groundedness);
    assert!(
        motivo.contains("Makefile"),
        "el motivo debe nombrar la fuente que no sostiene la cita: {motivo:?}"
    );
}

/// La mitad verdadera no salva a la falsa: el PRD **sí** dice DynamoDB, y aun así el segmento
/// entero se degrada. Admitirlo a medias dejaría en pie la afirmación sobre el `Makefile`.
#[test]
fn one_true_quote_does_not_carry_a_false_one() {
    let raw = r#"{"segments":[
        {"text":"Persistencia en DynamoDB, limitada a PostgreSQL 16 por el Makefile",
         "grounded":["PRD-00.md","Makefile"],
         "quotes":["la persistencia del Run es DynamoDB","solo soporta PostgreSQL 16"]}
    ]}"#;
    let segments = parse_segments(raw, &material()).unwrap();
    assert!(
        !segments[0].is_grounded(),
        "una cita inventada contamina el segmento entero: no hay verificación parcial"
    );
}

/// Y el reverso, que es lo que evita que la defensa sea un simple «degrádalo todo»: cuando la
/// cita **sí** está, el segmento se sostiene.
#[test]
fn a_quote_actually_present_keeps_the_segment_grounded() {
    let raw = r#"{"segments":[
        {"text":"El build se hace con cargo en modo release","grounded":["Makefile"],
         "quotes":["cargo build --release"]}
    ]}"#;
    let segments = parse_segments(raw, &material()).unwrap();
    assert!(
        segments[0].is_grounded(),
        "la cita está en el Makefile: degradarla haría inútil la comprobación"
    );
}

/// El mismo caso, pero de punta a punta: el modelo lee de verdad el `Makefile` a través del
/// loop y aun así se inventa lo que dice. Comprobarlo sobre `parse_segments` no bastaría —
/// hay que ver que el material que llega a la verificación es el que la sesión leyó.
#[tokio::test]
async fn end_to_end_the_loop_verifies_against_what_it_actually_read() {
    let generado = r#"{"segments":[
        {"text":"Solo soporta PostgreSQL 16","grounded":["Makefile"],
         "quotes":["el Makefile solo soporta PostgreSQL 16"]}
    ]}"#;

    let mut script = vec![
        CompletionOutput::ToolCalls(vec![ToolCall {
            id: "c1".into(),
            name: "read_file".into(),
            arguments: r#"{"path":"Makefile"}"#.into(),
        }]),
        CompletionOutput::ToolCalls(vec![ToolCall {
            id: "c2".into(),
            name: "finalize".into(),
            arguments: r#"{"summary":"leído"}"#.into(),
        }]),
    ];
    for _ in 0..4 {
        script.push(CompletionOutput::Text(generado.to_string()));
    }

    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", script)))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "Makefile", MAKEFILE,
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
                vec!["fake-model".into()],
            ),
        )))
        .cancellations(Arc::new(FakeCancellationFactory::new()))
        .build()
        .expect("el grafo local debe armarse");

    let service = ContextAuthoring::new(deps);
    let id = service
        .start_session(StartSession {
            repo_root: ".".into(),
            mode: Mode::Local,
            locale: Some("es".into()),
        })
        .await
        .expect("la sesión debe arrancar");
    service.join_session(&id).await.expect("debe terminar");

    let vista = service.session_state(&id).await.expect("debe haber vista");
    let todos: Vec<_> = vista.artifacts.iter().flat_map(|a| &a.segments).collect();
    assert!(
        !todos.is_empty(),
        "el pase debe producir segmentos; si no, el test no está comprobando nada"
    );
    assert!(
        todos.iter().all(|s| !s.is_grounded()),
        "ningún segmento puede quedar fundamentado: la cita no está en el Makefile que se leyó"
    );
}
