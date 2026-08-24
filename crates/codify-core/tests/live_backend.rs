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
/// `dynamodb` **no** está aquí, y la ausencia es deliberada: `docs/PRD.md` lo afirma con todas
/// las letras. Es la contradicción que el fixture planta a propósito, no una invención — el
/// modelo que lo recoge está siendo fiel a una fuente real. Meterlo en esta lista confundía
/// «contradicho» con «inventado», que son fallos distintos: el primero se señala (FR-008), el
/// segundo no debe ocurrir. Atribuir DynamoDB a quien no lo dice sí es invención, y de eso se
/// encarga la comprobación de citas de más abajo, que es donde corresponde.
const NEGADOS_POR_EL_SPEC: &[&str] = &["kafka", "rabbitmq", "broker", "event sourc"];

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
/// Aborta si el fixture arrastra la salida de una corrida anterior.
///
/// No es celo de limpieza: si `context/` o `AGENTS.md` ya están ahí, el agente los lee como
/// fuentes y puede **fundamentar una afirmación contra su propio output previo**. La medida
/// dejaría de decir lo que se cree que dice, y en verde, que es la peor forma de fallar.
/// Pasó de verdad — la corrida 2 del 2026-08-23 citó `context/CONTEXT.md` como fuente de la
/// contradicción sobre la persistencia, en vez de `docs/PRD.md`.
fn exigir_fixture_limpio() {
    let raiz = fixture();
    let sucios: Vec<_> = ["AGENTS.md", "context"]
        .iter()
        .filter(|n| std::path::Path::new(&raiz).join(n).exists())
        .collect();
    assert!(
        sucios.is_empty(),
        "el fixture arrastra artefactos de una corrida anterior ({sucios:?}): el agente los \
         leería como fuentes y se fundamentaría contra sí mismo. Regenéralo antes de cada \
         corrida con ./scripts/quickstart-fixture.sh"
    );
}

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

/// Un pase completo sobre el fixture, con el grafo real cableado.
///
/// Extraído para que T073 pueda encadenar dos sin duplicar el cableado: lo que ese test
/// comprueba es precisamente qué ocurre en el **segundo**.
async fn correr_pase() -> (
    codify_core::application::service::SessionSnapshot,
    Arc<RecordingAudit>,
) {
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
    (snap, audit)
}

#[tokio::test]
#[ignore = "necesita un backend de modelo local corriendo"]
async fn the_agent_follows_the_reference_instead_of_inventing_an_architecture() {
    exigir_fixture_limpio();

    let (snap, audit) = correr_pase().await;

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

    let segmentos: Vec<_> = snap.artifacts.iter().flat_map(|a| &a.segments).collect();

    // La prueba de fuego: no que se mencionen los términos, sino que se AFIRMEN. Y afirmar es
    // lo que hace un segmento **fundamentado**: un tentativo declara que no lo sabe y una
    // contradicción declara que las fuentes chocan. Escrutar el render entero los metía a
    // todos en el mismo saco y convertía en falta el señalamiento, que es justo lo que este
    // producto existe para hacer.
    let afirmado = segmentos
        .iter()
        .filter(|s| s.is_grounded())
        .map(|s| s.text.as_str())
        .collect::<Vec<_>>()
        .join("\n");
    let inventadas = afirmaciones_inventadas(&afirmado);
    assert!(
        inventadas.is_empty(),
        "el agente AFIRMÓ algo que el SPEC niega — no siguió la referencia.\n\
         Frases (de segmentos FUNDAMENTADOS, no de los señalados): {inventadas:#?}\n\n{contexto}"
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

// ---------------------------------------------------------------------------
// T073 — FR-006d contra un modelo real: dos pases encadenados
// ---------------------------------------------------------------------------

/// Dos pases sobre el **mismo** fixture, sin regenerarlo entre medias: el segundo lee los
/// artefactos que dejó el primero y no puede fundamentarse en ellos.
///
/// Invierte a propósito `exigir_fixture_limpio`. Aquel guardia protege la **medición** —una
/// corrida sucia daba números que no significaban lo que parecían— y por eso aborta. Este test
/// necesita justo ese fixture sucio, porque lo que prueba es que el **producto** resiste lo que
/// el guardia se limitaba a evitar. Por eso el guardia se comprueba al principio, sobre el
/// fixture todavía limpio, y no se vuelve a llamar.
#[tokio::test]
#[ignore = "necesita un backend vivo; CI nunca lo corre"]
async fn a_second_pass_cannot_ground_itself_on_what_the_first_one_wrote() {
    exigir_fixture_limpio();

    // Primer pase: deja `AGENTS.md` y `context/*` dentro del fixture.
    let (primera, _) = correr_pase().await;
    assert!(
        !primera.artifacts.is_empty(),
        "el primer pase debe escribir artefactos, o el segundo no tendría qué releer"
    );
    assert!(
        std::path::Path::new(&fixture()).join("context").exists(),
        "sin `context/` en disco el segundo pase no reproduce el escenario y el test no \
         probaría nada"
    );

    // Segundo pase sobre el fixture ya poblado: el agente leerá lo que escribimos nosotros.
    let (segunda, audit) = correr_pase().await;

    for e in audit.events.lock().unwrap().iter() {
        println!("--- 2º pase, audit {:?}: {}", e.kind, e.payload);
    }

    let mut apoyadas_en_lo_nuestro = Vec::new();
    let (mut fundamentados, mut tentativos) = (0usize, 0usize);

    for seg in segunda.artifacts.iter().flat_map(|a| &a.segments) {
        let fuentes = match &seg.groundedness {
            Groundedness::Grounded { sources, .. } => {
                fundamentados += 1;
                sources.clone()
            }
            Groundedness::Contradiction { sources, .. } => sources.clone(),
            Groundedness::Tentative { .. } => {
                tentativos += 1;
                continue;
            }
        };
        for fuente in fuentes {
            if ArtifactKind::is_canonical_path(&fuente) {
                apoyadas_en_lo_nuestro.push(format!("«{}» citando {fuente}", seg.text));
            }
        }
    }

    println!("--- 2º pase: {fundamentados} fundamentados, {tentativos} tentativos");

    // Que el agente **lea** los artefactos previos depende de lo que decida explorar, y no se
    // puede forzar sin dejar de medir un pase real. Cuando no los lee, este test comprueba el
    // resultado sin haber ejercitado el mecanismo — y eso hay que decirlo, porque un verde que
    // no probó nada se lee igual que uno que sí.
    let leyo_lo_nuestro = audit
        .events
        .lock()
        .unwrap()
        .iter()
        .any(|e| ArtifactKind::is_canonical_path(&e.payload));
    if !leyo_lo_nuestro {
        println!(
            "--- AVISO: el 2º pase no llegó a leer ningún artefacto propio. El escenario no se \
             ejercitó en vivo; quien lo cubre de forma determinista es us1_provenance.rs"
        );
    }

    assert!(
        apoyadas_en_lo_nuestro.is_empty(),
        "el segundo pase se fundamentó en la salida del primero (FR-006d):\n\
         {apoyadas_en_lo_nuestro:#?}"
    );
}
