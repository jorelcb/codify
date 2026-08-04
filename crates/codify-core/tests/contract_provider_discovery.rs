//! **Contract test del port `ProviderDiscovery`** (T008).
//!
//! La regla que sostiene FR-019: sondear **nunca falla**. Si el backend no responde, el
//! resultado trae un motivo accionable — un `Err` opaco es exactamente la experiencia que
//! este port viene a evitar.

mod fakes;

use codify_core::application::ports::{ProviderDiscovery, ProviderStatus};
use codify_core::infrastructure::providers::probe::LocalProviderProbe;
use fakes::FakeProviderDiscovery;

fn provider_status_contract(status: &ProviderStatus, label: &str) {
    assert!(
        !status.endpoint.is_empty(),
        "[{label}] el endpoint se declara siempre"
    );
    if status.reachable {
        assert!(
            status.detail.is_none(),
            "[{label}] alcanzable no necesita motivo"
        );
    } else {
        let detail = status.detail.as_deref().unwrap_or("");
        assert!(
            !detail.trim().is_empty(),
            "[{label}] si no responde, el motivo NO puede estar vacío"
        );
        assert!(
            status.models.is_empty(),
            "[{label}] sin backend no hay modelos"
        );
    }
}

/// Con un puerto donde no hay nada escuchando: no debe fallar, debe explicar.
#[tokio::test]
async fn probing_an_absent_backend_reports_instead_of_failing() {
    let probe = LocalProviderProbe::new("http://127.0.0.1:1").expect("loopback válido");
    let status = probe.probe().await;

    assert!(!status.reachable, "no hay backend en ese puerto");
    provider_status_contract(&status, "real-ausente");
}

#[tokio::test]
async fn provider_status_contract_holds_for_the_fake() {
    let reachable = FakeProviderDiscovery(ProviderStatus::reachable(
        "http://localhost:11434",
        vec!["qwen2.5-coder".into()],
    ));
    provider_status_contract(&reachable.probe().await, "fake-alcanzable");

    let down = FakeProviderDiscovery(ProviderStatus::unreachable(
        "http://localhost:11434",
        "no hay ningún backend escuchando; arranca Ollama con `ollama serve`",
    ));
    provider_status_contract(&down.probe().await, "fake-caído");
}

/// La sonda no puede convertirse en una vía de salida: solo acepta loopback.
#[test]
fn the_probe_rejects_non_loopback_endpoints() {
    assert!(LocalProviderProbe::new("https://api.remoto.test").is_err());
    assert!(LocalProviderProbe::new("http://localhost:11434").is_ok());
    assert!(LocalProviderProbe::new("http://127.0.0.1:8080").is_ok());
}
