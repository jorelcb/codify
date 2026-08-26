//! **`001`-FR-017 / `003`-US2** — que lo barato haga lo frecuente.
//!
//! Este es el MUST que originó el spec `003`. El enrutado existía desde `001`, pero con un solo
//! proveedor real detrás no distinguía nada: `pick(Heavy)` y `pick(Cheap)` devolvían el mismo.
//! Con dos, empieza a significar lo que dice.

mod fakes;

use codify_core::application::deps::ProviderRegistry;
use codify_core::application::ports::{CompletionRequest, Message, ModelProvider, Secreto, Tier};
use codify_core::domain::error::CoreError;
use codify_core::domain::session::{Mode, SessionFailure};
use codify_core::infrastructure::providers::remote::RemoteOpenAiCompatProvider;
use fakes::*;
use std::sync::Arc;

fn registro_con_dos_tiers() -> ProviderRegistry {
    ProviderRegistry::for_mode(
        Mode::Hybrid,
        vec![
            Arc::new(FakeModelProvider::local("local-barato", Vec::new())),
            Arc::new(FakeModelProvider::remote("frontier").with_tier(Tier::Heavy)),
        ],
    )
    .expect("modo híbrido admite un remoto")
}

#[test]
fn el_refinamiento_va_al_tier_economico() {
    let elegido = registro_con_dos_tiers().pick(Tier::Cheap);
    assert_eq!(elegido.provider.name(), "local-barato");
    assert!(
        elegido.degraded_from.is_none(),
        "había proveedor del tier pedido: no hay nada que declarar"
    );
}

#[test]
fn la_generacion_pesada_va_al_de_mayor_capacidad() {
    let elegido = registro_con_dos_tiers().pick(Tier::Heavy);
    assert_eq!(elegido.provider.name(), "frontier");
    assert!(elegido.degraded_from.is_none());
}

/// Con un solo tier el enrutado degrada **y lo declara** (`001`-FR-018). Se comprueba aquí para
/// que quede claro que ese camino sigue vivo cuando el usuario no ha conectado un remoto — que
/// es el caso por defecto del producto.
#[test]
fn con_un_solo_tier_se_degrada_y_se_dice() {
    let solo_local = ProviderRegistry::for_mode(
        Mode::Local,
        vec![Arc::new(FakeModelProvider::local("local", Vec::new()))],
    )
    .unwrap();
    let elegido = solo_local.pick(Tier::Heavy);
    assert_eq!(elegido.degraded_from, Some(Tier::Heavy));
}

// ---------------------------------------------------------------------------
// T019 — un fallo de autorización no es un fallo del modelo
// ---------------------------------------------------------------------------

/// Piden cosas distintas del usuario —reconectar frente a reintentar—, así que llegan como
/// motivos distintos. Es `002`-FR-028 aplicado al caso nuevo que introduce este spec.
#[tokio::test]
async fn una_credencial_rechazada_llega_como_fallo_de_autorizacion() {
    let servidor = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let puerto = servidor.local_addr().unwrap().port();
    tokio::spawn(async move {
        while let Ok((mut socket, _)) = servidor.accept().await {
            tokio::spawn(async move {
                use tokio::io::AsyncWriteExt;
                let _ = socket
                    .write_all(b"HTTP/1.1 401 Unauthorized\r\nContent-Length: 0\r\n\r\n")
                    .await;
            });
        }
    });

    let provider = RemoteOpenAiCompatProvider::new(
        "frontier",
        format!("http://127.0.0.1:{puerto}"),
        "modelo",
        Tier::Heavy,
        Secreto::new("sk-caducada"),
    )
    .expect("adapter");

    let err = provider
        .complete(CompletionRequest {
            system: "x".into(),
            messages: vec![Message::user("y")],
            tools: Vec::new(),
        })
        .await
        .expect_err("la credencial no vale");

    assert!(
        matches!(err, CoreError::Unauthorized(_)),
        "una credencial caducada pide reconectar, no reintentar: {err:?}"
    );
    assert_eq!(SessionFailure::from(&err).code(), "unauthorized");
}
