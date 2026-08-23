//! **Validación contra un backend real** — quickstart S1 de `001`, y la pregunta que originó
//! el proyecto.
//!
//! Está marcado `#[ignore]` a propósito: necesita un modelo vivo, así que CI **nunca** lo corre.
//! Se lanza a mano cuando hay uno delante:
//!
//! ```bash
//! ./scripts/quickstart-fixture.sh
//! CODIFY_LOCAL_ENDPOINT=http://127.0.0.1:8080 \
//!   cargo test -p codify-core --test live_backend -- --ignored --nocapture
//! ```
//!
//! Lo que prueba no es el mecanismo —eso ya está cubierto por los tests con proveedor de
//! guion— sino **la calidad de la salida**, que es justo lo que un guion no puede demostrar
//! (SC-001). El fixture está construido para que quedarse en el README produzca una
//! arquitectura inventada: habla de "orquestar trabajos de larga duración", lo que invita a
//! suponer un broker de mensajes. El SPEC hermano dice explícitamente que **no lo hay**.
//!
//! Si el contexto generado menciona un broker, el agente no siguió la referencia.

use codify_core::application::ingest::IngestBudget;
use codify_core::application::service::{AuthoringService, ContextAuthoring, StartSession};
use codify_core::domain::context::{ArtifactKind, Groundedness};
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::CoreBuilder;
use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
use codify_core::infrastructure::repo::locale::HeuristicLocaleDetector;
use codify_core::infrastructure::repo::navigator::FsRepoNavigator;
use codify_core::infrastructure::repo::reference_resolver::FsHttpReferenceResolver;
use codify_core::infrastructure::repo::writer::FsArtifactWriter;
use std::sync::Arc;

mod fakes;
use fakes::{FakePrompter, FixedClock, RecordingAudit};

fn endpoint() -> String {
    std::env::var("CODIFY_LOCAL_ENDPOINT").unwrap_or_else(|_| "http://localhost:11434".into())
}

fn fixture() -> String {
    std::env::var("CODIFY_FIXTURE").unwrap_or_else(|_| "/tmp/codify-fixture".into())
}

/// Términos que el SPEC **niega**. Mencionarlos no es inventar — el propio SPEC los nombra
/// para descartarlos, y un contexto fiel hará lo mismo. Lo que delata la invención es
/// **afirmarlos**: decir que hay un broker, no decir que no lo hay.
const NEGADOS_POR_EL_SPEC: &[&str] = &["kafka", "rabbitmq", "broker", "event sourc", "dynamodb"];

/// Marcas de negación en español e inglés, más las de tentativo: si aparecen en la misma
/// frase, el término está siendo descartado o declarado sin verificar, no afirmado.
const NIEGA: &[&str] = &[
    "no ",
    "not ",
    "sin ",
    "ningún",
    "ningun",
    "never",
    "instead of",
    "en vez de",
    "tentativo",
    "tentative",
    "rather than",
    // Descartar en ambos idiomas: el modelo responde en el idioma que detecte, así que la
    // lista tiene que cubrir los dos o produce falsos positivos.
    "descart",
    "discard",
    "ruled out",
    "evaluated and",
    "se evaluó",
];

/// Busca afirmaciones de lo que el SPEC niega, frase a frase.
///
/// Contar apariciones sueltas daría falsos positivos con «**no** hay broker», que es
/// exactamente lo que un contexto correcto dice.
/// Todo el material del fixture, normalizado igual que lo normaliza el núcleo. Sirve para
/// comprobar **desde fuera** lo que `parse_segments` promete desde dentro.
fn material_del_fixture() -> String {
    fn recoger(dir: &std::path::Path, out: &mut String) {
        let Ok(entradas) = std::fs::read_dir(dir) else {
            return;
        };
        for e in entradas.flatten() {
            let ruta = e.path();
            if ruta.is_dir() {
                recoger(&ruta, out);
            } else if let Ok(c) = std::fs::read_to_string(&ruta) {
                out.push_str(&c);
                out.push('\n');
            }
        }
    }
    let mut todo = String::new();
    recoger(std::path::Path::new(&fixture()), &mut todo);
    todo.to_lowercase()
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
}

fn afirmaciones_inventadas(contexto: &str) -> Vec<String> {
    contexto
        .split(['.', '\n'])
        .map(str::trim)
        .filter(|f| !f.is_empty())
        .filter(|frase| {
            let bajo = frase.to_lowercase();
            NEGADOS_POR_EL_SPEC.iter().any(|t| bajo.contains(t))
                && !NIEGA.iter().any(|n| bajo.contains(n))
        })
        .map(|f| f.to_string())
        .collect()
}

#[tokio::test]
#[ignore = "necesita un backend de modelo local corriendo"]
async fn the_agent_follows_the_reference_instead_of_inventing_an_architecture() {
    let repo = fixture();
    let model = std::env::var("CODIFY_MODEL").unwrap_or_else(|_| "default".into());

    let audit = Arc::new(RecordingAudit::default());
    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(
            LocalOpenAiCompatProvider::new("local", endpoint(), model)
                .expect("el endpoint debe ser loopback"),
        ))
        .navigator(Arc::new(FsRepoNavigator::new(&repo)))
        .resolver(Arc::new(FsHttpReferenceResolver::with_public_web(&repo)))
        .diff(Arc::new(
            codify_core::infrastructure::diff::engine::SimilarDiffEngine,
        ))
        .risk(Arc::new(
            codify_core::infrastructure::diff::risk::ConservativeRiskClassifier,
        ))
        .prompter(Arc::new(FakePrompter::approving()))
        .audit(audit.clone())
        .locale(Arc::new(HeuristicLocaleDetector::new(String::new())))
        .clock(Arc::new(FixedClock))
        .writer(Arc::new(FsArtifactWriter::new(&repo)))
        .discovery(Arc::new(
            codify_core::infrastructure::providers::probe::LocalProviderProbe::new(endpoint())
                .expect("sonda local"),
        ))
        .cancellations(Arc::new(
            codify_core::infrastructure::cancel::TokenCancellationFactory,
        ))
        .build()
        .expect("el grafo local se cablea");

    let svc = ContextAuthoring::new(deps).with_budget(IngestBudget::new(40, 8, 60));
    let id = svc
        .start_session(StartSession {
            repo_root: repo.clone().into(),
            mode: Mode::Local,
            locale: None,
        })
        .await
        .expect("la sesión arranca");
    svc.join_session(&id).await.expect("la sesión termina");

    let snap = svc.session_state(&id).await.expect("hay snapshot");

    // Diagnóstico antes de asertar: si el modelo no produjo lo esperado, hay que poder ver
    // QUÉ produjo, no solo que faltaba.
    println!("--- estado de la sesión: {:?}", snap.state);
    println!(
        "--- artefactos generados: {:?}",
        snap.artifacts
            .iter()
            .map(|a| a.kind.file_path())
            .collect::<Vec<_>>()
    );
    for e in audit.events.lock().unwrap().iter() {
        println!("--- audit {:?}: {}", e.kind, e.payload);
    }

    let contexto = snap
        .artifacts
        .iter()
        .find(|a| a.kind == ArtifactKind::Context)
        .unwrap_or_else(|| {
            panic!(
                "no se generó el artefacto de contexto. Estado: {:?}. Generados: {:?}",
                snap.state,
                snap.artifacts
                    .iter()
                    .map(|a| a.kind.file_path())
                    .collect::<Vec<_>>()
            )
        })
        .render();

    println!("\n===== CONTEXTO GENERADO =====\n{contexto}\n=============================\n");
    println!("--- referencias no resueltas: {:?}", snap.unresolved);
    println!("--- omitido: {:?}", snap.omitted);
    println!("--- escrituras: {:?}", snap.writes.len());

    // La prueba de fuego: no que se mencionen los términos, sino que se AFIRMEN.
    let inventadas = afirmaciones_inventadas(&contexto);
    assert!(
        inventadas.is_empty(),
        "el agente AFIRMÓ algo que el SPEC niega — no siguió la referencia.\n\
         Frases: {inventadas:#?}\n\n{contexto}"
    );

    // Y la otra mitad: que las negaciones del SPEC lleguen al contexto. Un texto que
    // simplemente omitiera el tema también pasaría el test de arriba.
    let bajo = contexto.to_lowercase();
    assert!(
        bajo.contains("broker") || bajo.contains("event sourc"),
        "el contexto ni afirma ni NIEGA lo que el SPEC descarta: omitirlo deja al siguiente \
         agente sin la información que evita el error.\n{contexto}"
    );
    assert!(
        contexto.to_lowercase().contains("temporal"),
        "el contexto no recoge lo que el SPEC sí dice (Temporal):\n{contexto}"
    );

    // -----------------------------------------------------------------------
    // Fase 7 (T065) — que F-1 no se reproduzca, y que la defensa no sea vacua
    // -----------------------------------------------------------------------

    let material = material_del_fixture();
    let segmentos: Vec<_> = snap.artifacts.iter().flat_map(|a| &a.segments).collect();
    let (mut fundamentados, mut tentativos, mut sin_respaldo) = (0usize, 0usize, Vec::new());

    for seg in &segmentos {
        let citas = match &seg.groundedness {
            Groundedness::Grounded { quotes, .. } => {
                fundamentados += 1;
                quotes.clone()
            }
            Groundedness::Contradiction { quotes, .. } => quotes.clone(),
            Groundedness::Tentative { .. } => {
                tentativos += 1;
                Vec::new()
            }
        };
        for cita in citas {
            let cn = cita
                .to_lowercase()
                .split_whitespace()
                .collect::<Vec<_>>()
                .join(" ");
            if !material.contains(&cn) {
                sin_respaldo.push(format!("«{cita}» en: {}", seg.text));
            }
        }
    }

    println!("--- procedencia: {fundamentados} fundamentados, {tentativos} tentativos");

    // Esto es F-1: una afirmación presentada como verificada cuya cita no está en lo leído.
    assert!(
        sin_respaldo.is_empty(),
        "F-1 SE REPRODUCE: hay citas que no aparecen en el material leído.\n{sin_respaldo:#?}"
    );

    // Y el reverso, sin el cual lo de arriba se cumpliría degradándolo todo: si el material
    // nunca llegara a la verificación, cero segmentos sobrevivirían y el test de arriba
    // pasaría igual. Un pase que no fundamenta NADA no es una victoria, es una tubería rota.
    assert!(
        fundamentados > 0,
        "ningún segmento quedó fundamentado sobre un fixture con fuentes legibles: \
         revisa que el material leído llegue a la verificación, no que el modelo falle.\n{contexto}"
    );
}
