//! **Contract test del port `Cancellation`** (T006), contra el adapter real y el fake.
//!
//! Lo que se asserta es comportamiento, no implementación: una vez cancelado no se
//! "descancela", y todos los que esperan despiertan. Ese segundo punto es el que permite
//! abortar la petición al modelo en vuelo en vez de esperar a que termine.

mod fakes;

use codify_core::application::ports::Cancellation;
use codify_core::infrastructure::cancel::TokenCancellation;
use fakes::FakeCancellation;
use std::sync::Arc;
use std::time::Duration;

async fn cancellation_contract(signal: Arc<dyn Cancellation>, cancel: impl FnOnce(), label: &str) {
    assert!(!signal.is_cancelled(), "[{label}] nace sin cancelar");

    // Tres esperadores concurrentes: todos deben despertar.
    let waiters: Vec<_> = (0..3)
        .map(|_| {
            let s = signal.clone();
            tokio::spawn(async move { s.cancelled().await })
        })
        .collect();

    tokio::time::sleep(Duration::from_millis(20)).await;
    cancel();

    for (i, w) in waiters.into_iter().enumerate() {
        tokio::time::timeout(Duration::from_secs(2), w)
            .await
            .unwrap_or_else(|_| panic!("[{label}] el esperador {i} no despertó"))
            .unwrap();
    }

    assert!(signal.is_cancelled(), "[{label}] queda cancelado");
}

#[tokio::test]
async fn cancellation_contract_holds_for_the_real_token() {
    let token = Arc::new(TokenCancellation::new());
    let t = token.clone();
    cancellation_contract(token.clone(), move || t.cancel(), "token").await;
}

#[tokio::test]
async fn cancellation_contract_holds_for_the_fake() {
    let fake = Arc::new(FakeCancellation::new());
    let f = fake.clone();
    cancellation_contract(fake.clone(), move || f.cancel(), "fake").await;
}

/// Una vez cancelado no hay vuelta atrás: `cancelled()` resuelve de inmediato.
#[tokio::test]
async fn cancelling_is_irreversible_and_resolves_immediately_afterwards() {
    let token = TokenCancellation::new();
    token.cancel();
    assert!(token.is_cancelled());

    tokio::time::timeout(Duration::from_millis(200), token.cancelled())
        .await
        .expect("tras cancelar, esperar no debe bloquear");

    // Sigue cancelado: no existe forma de "descancelar".
    assert!(token.is_cancelled());
}
