//! **Harness de cero-egress** — restricción de proyecto [NON-NEGOTIABLE], SC-007.
//!
//! La garantía es *estructural*: en modo `Local` el grafo de objetos no puede contener un
//! proveedor no local. No se comprueba un flag en runtime — se comprueba que el adapter
//! remoto **no existe** en el registro.

mod fakes;

use codify_core::application::deps::ProviderRegistry;
use codify_core::application::ports::Tier;
use codify_core::domain::error::CoreError;
use codify_core::domain::session::Mode;
use codify_core::infrastructure::composition::CoreBuilder;
use fakes::*;
use std::sync::Arc;

#[test]
fn local_mode_rejects_a_remote_provider() {
    let result = ProviderRegistry::for_mode(
        Mode::Local,
        vec![Arc::new(FakeModelProvider::remote("frontier-remoto"))],
    );

    match result {
        Err(CoreError::EgressBlocked(msg)) => {
            assert!(
                msg.contains("frontier-remoto"),
                "el error debe nombrar el proveedor: {msg}"
            );
        }
        Err(other) => panic!("se esperaba EgressBlocked, llegó {other:?}"),
        Ok(_) => panic!("el modo Local NO debe admitir un proveedor remoto"),
    }
}

#[test]
fn local_mode_rejects_a_mixed_registry() {
    let result = ProviderRegistry::for_mode(
        Mode::Local,
        vec![
            Arc::new(FakeModelProvider::local("ollama", vec![])),
            Arc::new(FakeModelProvider::remote("frontier-remoto")),
        ],
    );
    assert!(
        result.is_err(),
        "un solo proveedor remoto contamina todo el registro local"
    );
}

#[test]
fn local_mode_registry_is_fully_local() {
    let registry = ProviderRegistry::for_mode(
        Mode::Local,
        vec![
            Arc::new(FakeModelProvider::local("ollama", vec![])),
            Arc::new(FakeModelProvider::local("llamacpp", vec![]).with_tier(Tier::Heavy)),
        ],
    )
    .expect("proveedores locales deben ser admitidos");

    assert!(registry.is_fully_local());
    assert!(registry.all().iter().all(|p| p.is_local()));
    assert!(registry.pick(Tier::Heavy).is_local());
    assert!(registry.pick(Tier::Cheap).is_local());
}

#[test]
fn hybrid_mode_allows_remote_providers() {
    let registry = ProviderRegistry::for_mode(
        Mode::Hybrid,
        vec![
            Arc::new(FakeModelProvider::local("ollama", vec![])),
            Arc::new(FakeModelProvider::remote("frontier-remoto")),
        ],
    )
    .expect("el modo híbrido sí admite remotos con consentimiento explícito");

    assert!(!registry.is_fully_local());
}

#[test]
fn empty_registry_is_rejected() {
    assert!(ProviderRegistry::for_mode(Mode::Local, vec![]).is_err());
}

/// El composition root completo tampoco puede ensamblarse con un remoto en modo local.
#[test]
fn composition_root_cannot_be_wired_with_remote_provider_in_local_mode() {
    let result = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::remote("frontier-remoto")))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            "x",
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
        .build();

    assert!(
        matches!(result, Err(CoreError::EgressBlocked(_))),
        "el composition root debe bloquear el cableado remoto en modo Local"
    );
}

#[test]
fn composition_root_wires_a_fully_local_graph() {
    let deps = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", vec![])))
        .navigator(Arc::new(FakeRepoNavigator::with_files(&[(
            "README.md",
            "x",
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
        .expect("el grafo totalmente local debe ensamblarse");

    assert!(deps.mode.is_local());
    assert!(deps.providers.is_fully_local());
}

#[test]
fn composition_root_fails_when_a_port_is_missing() {
    let result = CoreBuilder::new(Mode::Local)
        .provider(Arc::new(FakeModelProvider::local("ollama", vec![])))
        .build();
    assert!(
        result.is_err(),
        "faltan ports por cablear: debe fallar explícitamente"
    );
}
