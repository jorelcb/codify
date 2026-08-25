//! **`002`-FR-028** — un fallo tiene que poder entenderse.
//!
//! Hoy la sesión puede morir con cero artefactos y sin decir por qué: `service.rs` hacía
//! `Err(_) => advance_to(Failed)` y descartaba el motivo con un comodín, teniendo `CoreError`
//! siete variantes tipadas.
//!
//! El coste de eso está medido. El 2026-08-24, durante la Fase 7 de `001`, diagnosticar un
//! simple *timeout* de cliente costó **cinco corridas** de entre 4 y 9 minutos y una hipótesis
//! falsa —truncamiento por ventana de contexto— perseguida hasta el final. No por difícil:
//! porque sin motivo cualquier hipótesis vale lo mismo, y se persiguen por plausibilidad en vez
//! de por evidencia.

mod fakes;

use codify_core::domain::error::CoreError;
use codify_core::domain::session::{
    AuthoringSession, Mode, SessionFailure, SessionId, SessionState,
};

fn sesion() -> AuthoringSession {
    AuthoringSession::start(SessionId::new("s1"), ".", Mode::Local)
}

// ---------------------------------------------------------------------------
// T058 — el motivo se expone como código, no como mensaje crudo
// ---------------------------------------------------------------------------

#[test]
fn una_sesion_que_muere_expone_el_codigo_y_no_el_mensaje_crudo() {
    let mut s = sesion();
    s.fail(SessionFailure::ProviderUnavailable);

    assert_eq!(s.state(), SessionState::Failed);
    assert_eq!(
        s.failure().map(|f| f.code()),
        Some("provider_unavailable"),
        "la piel elige la frase a partir del código; el núcleo no redacta en un idioma fijo"
    );
}

#[test]
fn el_estado_fallido_no_se_alcanza_sin_motivo() {
    let mut s = sesion();
    // `advance_to` ya no admite `Failed`: el único camino es `fail(motivo)`.
    let intento = s.advance_to(SessionState::Failed);
    assert!(
        intento.is_err(),
        "poder llegar a Failed sin motivo es exactamente cómo se perdió durante meses"
    );
    assert!(s.failure().is_none());
}

// ---------------------------------------------------------------------------
// T059 — el caso medido: distinguir un timeout de una respuesta ilegible
// ---------------------------------------------------------------------------

#[test]
fn un_timeout_no_se_confunde_con_una_respuesta_no_parseable() {
    let timeout = SessionFailure::from(&CoreError::ProviderTimeout("local tardó demasiado".into()));
    let ilegible = SessionFailure::from(&CoreError::Provider("respuesta sin choices[0]".into()));

    assert_eq!(timeout.code(), "provider_timeout");
    assert_eq!(ilegible.code(), "provider_unparseable");
    assert_ne!(
        timeout.code(),
        ilegible.code(),
        "sin esta distinción el hallazgo de #24 se repite: son diagnósticos opuestos —esperar \
         más frente a revisar el prompt— y confundirlos cuesta corridas enteras"
    );
}

#[test]
fn cada_variante_de_coreerror_tiene_un_motivo_de_sesion() {
    let casos = [
        (CoreError::Provider("x".into()), "provider_unparseable"),
        (CoreError::ProviderTimeout("x".into()), "provider_timeout"),
        (CoreError::Unavailable("x".into()), "provider_unavailable"),
        (CoreError::EgressBlocked("x".into()), "egress_blocked"),
        (CoreError::Storage("x".into()), "repo_unreadable"),
        (CoreError::NotFound("x".into()), "repo_unreadable"),
        (CoreError::Invalid("x".into()), "internal"),
        (CoreError::Unauthorized("x".into()), "unauthorized"),
    ];
    for (err, esperado) in casos {
        assert_eq!(
            SessionFailure::from(&err).code(),
            esperado,
            "cada error del núcleo tiene que llegar al usuario como algo que pueda leer: {err:?}"
        );
    }
}

/// Los códigos son parte del contrato con la piel: cambiarlos rompe el catálogo en silencio.
#[test]
fn los_codigos_son_estables_y_sin_duplicados() {
    let todos = SessionFailure::all();
    let mut codigos: Vec<_> = todos.iter().map(|f| f.code()).collect();
    let antes = codigos.len();
    codigos.sort_unstable();
    codigos.dedup();
    assert_eq!(
        codigos.len(),
        antes,
        "dos motivos con el mismo código harían que la interfaz mostrara la frase equivocada"
    );
    assert!(codigos.iter().all(|c| !c.is_empty()));
}

// ---------------------------------------------------------------------------
// El adapter: que el timeout se PRODUZCA, no solo que se mapee
// ---------------------------------------------------------------------------

/// Un servidor que acepta la conexión y **nunca contesta**.
///
/// Este test existe por un hueco que destapó la comprobación por inyección: al desactivar
/// `e.is_timeout()` en el proveedor, los cinco tests de arriba seguían en verde. Todos prueban
/// el **mapeo** `CoreError → SessionFailure`; ninguno miraba dónde se decide de verdad cuál de
/// los dos errores se construye. Sin esto, la distinción que costó cinco corridas podía
/// desaparecer del adapter sin que nada se enterara.
#[tokio::test]
async fn un_backend_que_no_contesta_produce_provider_timeout_y_no_un_error_generico() {
    use codify_core::application::ports::{CompletionRequest, Message, ModelProvider};
    use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
    use std::time::Duration;

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0")
        .await
        .expect("puerto libre");
    let puerto = listener.local_addr().expect("dirección").port();
    tokio::spawn(async move {
        // Aceptar y callar: es lo que hace un modelo que está pensando demasiado.
        while let Ok((socket, _)) = listener.accept().await {
            std::mem::forget(socket);
        }
    });

    let provider = LocalOpenAiCompatProvider::con_tiempo(
        "lento",
        format!("http://127.0.0.1:{puerto}"),
        "cualquiera",
        Duration::from_millis(150),
    )
    .expect("loopback es válido");

    let err = provider
        .complete(CompletionRequest {
            system: "x".into(),
            messages: vec![Message::user("y")],
            tools: Vec::new(),
        })
        .await
        .expect_err("no puede responder: nadie contesta al otro lado");

    assert!(
        matches!(err, CoreError::ProviderTimeout(_)),
        "esperar más y revisar el prompt son diagnósticos opuestos; el adapter es quien sabe \
         cuál fue, y aquí se comprueba que lo dice: {err:?}"
    );
    assert_eq!(SessionFailure::from(&err).code(), "provider_timeout");
}
