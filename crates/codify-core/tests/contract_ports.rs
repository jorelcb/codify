//! **Contract tests por port** (patrón hex-integration-test): la misma suite corre contra el
//! adapter real y contra el fake in-memory. Lo que se asserta es **comportamiento de
//! dominio**, nunca detalles internos del adapter.
//!
//! Nota de rutas: Cargo solo descubre tests en la raíz de `tests/`, así que los tres
//! contratos del plan viven aquí, seccionados, en vez de en `tests/contract/*.rs`.

mod fakes;

use codify_core::application::ports::{ModelProvider, ReferenceResolver, RepoNavigator, Tier};
use codify_core::domain::reference::{ReferenceOrigin, ReferenceState};
use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
use codify_core::infrastructure::repo::navigator::FsRepoNavigator;
use codify_core::infrastructure::repo::reference_resolver::FsHttpReferenceResolver;
use fakes::*;
use std::path::PathBuf;

/// Crea un repo temporal con contenido conocido. Sin dependencias externas.
fn temp_repo(tag: &str) -> PathBuf {
    let dir = std::env::temp_dir().join(format!("codify-ng-contract-{tag}-{}", std::process::id()));
    let _ = std::fs::remove_dir_all(&dir);
    std::fs::create_dir_all(dir.join("docs")).unwrap();
    std::fs::write(dir.join("README.md"), "# Proyecto\nVer docs/SPEC.md").unwrap();
    std::fs::write(dir.join("docs/SPEC.md"), "Motor: Temporal. Sin broker.").unwrap();
    std::fs::write(dir.join("Cargo.toml"), "[package]\nname=\"x\"").unwrap();
    dir
}

// ===========================================================================
// Contrato: RepoNavigator (T018)
// ===========================================================================

async fn repo_navigator_contract(nav: &dyn RepoNavigator, label: &str) {
    // 1. Lee un archivo existente y devuelve su contenido.
    let file = nav
        .read("README.md")
        .await
        .unwrap_or_else(|e| panic!("[{label}] {e}"));
    assert!(
        file.content.contains("Proyecto"),
        "[{label}] debe devolver el contenido real"
    );
    assert!(
        !file.truncated,
        "[{label}] un archivo pequeño no se recorta"
    );

    // 2. Un archivo inexistente es un error de dominio, no un panic ni contenido vacío.
    assert!(
        nav.read("no-existe.md").await.is_err(),
        "[{label}] debe fallar explícitamente"
    );

    // 3. Lista entradas de la raíz.
    let entries = nav.list("").await.unwrap();
    assert!(
        entries.iter().any(|e| e.path.ends_with("README.md")),
        "[{label}] la raíz debe incluir el README"
    );

    // 4. Describe el repo con señales estructurales.
    let repo = nav.describe().await.unwrap();
    assert!(!repo.is_empty, "[{label}] el repo tiene contenido");
}

#[tokio::test]
async fn repo_navigator_contract_holds_for_the_real_fs_adapter() {
    let dir = temp_repo("nav");
    let nav = FsRepoNavigator::new(&dir);
    repo_navigator_contract(&nav, "fs").await;
    let _ = std::fs::remove_dir_all(&dir);
}

#[tokio::test]
async fn repo_navigator_contract_holds_for_the_in_memory_fake() {
    let nav = FakeRepoNavigator::with_files(&[
        ("README.md", "# Proyecto\nVer docs/SPEC.md"),
        ("docs/SPEC.md", "Motor: Temporal. Sin broker."),
    ]);
    repo_navigator_contract(&nav, "fake").await;
}

/// Regla de dominio propia del adapter real: el agente no puede salirse del repositorio.
#[tokio::test]
async fn real_navigator_refuses_to_escape_the_repository() {
    let dir = temp_repo("escape");
    let nav = FsRepoNavigator::new(&dir);
    assert!(nav.read("../../../etc/passwd").await.is_err());
    assert!(nav.read("/etc/passwd").await.is_err());
    let _ = std::fs::remove_dir_all(&dir);
}

/// El recorte se **declara** con `truncated`; nunca es silencioso.
#[tokio::test]
async fn real_navigator_flags_truncation_instead_of_hiding_it() {
    let dir = temp_repo("trunc");
    std::fs::write(dir.join("grande.md"), "x".repeat(5_000)).unwrap();
    let nav = FsRepoNavigator::new(&dir).with_max_file_bytes(100);
    let file = nav.read("grande.md").await.unwrap();
    assert!(file.truncated, "el recorte debe viajar declarado");
    assert!(file.content.len() <= 100);
    let _ = std::fs::remove_dir_all(&dir);
}

// ===========================================================================
// Contrato: ReferenceResolver (T019)
// ===========================================================================

async fn reference_resolver_contract(
    resolver: &dyn ReferenceResolver,
    resolvable: &str,
    label: &str,
) {
    // 1. Una referencia local existente se resuelve con contenido.
    let ok = resolver
        .resolve(&ReferenceOrigin::LocalPath(resolvable.into()))
        .await;
    assert!(
        ok.is_resolved(),
        "[{label}] debe resolver la referencia local"
    );
    assert!(
        ok.content().is_some(),
        "[{label}] la referencia resuelta expone contenido"
    );

    // 2. Una referencia inexistente NO se resuelve y **jamás** lleva contenido (SC-006).
    let missing = resolver
        .resolve(&ReferenceOrigin::LocalPath("no-existe-xyz.md".into()))
        .await;
    assert!(!missing.is_resolved(), "[{label}] no debe resolverse");
    assert_eq!(
        missing.content(),
        None,
        "[{label}] jamás se fabrica contenido"
    );
    assert_ne!(missing.state(), ReferenceState::Resolved);
}

#[tokio::test]
async fn reference_resolver_contract_holds_for_the_real_adapter() {
    let dir = temp_repo("ref");
    let resolver = FsHttpReferenceResolver::local_only(&dir);
    reference_resolver_contract(&resolver, "docs/SPEC.md", "fs").await;
    let _ = std::fs::remove_dir_all(&dir);
}

#[tokio::test]
async fn reference_resolver_contract_holds_for_the_fake() {
    let resolver = FakeReferenceResolver::new().resolving("docs/SPEC.md", "Motor: Temporal");
    reference_resolver_contract(&resolver, "docs/SPEC.md", "fake").await;
}

/// Lo que requiere autenticación se **reporta** (fuera de v1), nunca se inventa.
#[tokio::test]
async fn requires_auth_is_reported_without_content() {
    let resolver = FakeReferenceResolver::new()
        .failing("https://example.test/privado", ReferenceState::RequiresAuth);
    let r = resolver
        .resolve(&ReferenceOrigin::PublicUrl(
            "https://example.test/privado".into(),
        ))
        .await;
    assert_eq!(r.state(), ReferenceState::RequiresAuth);
    assert_eq!(r.content(), None);
}

/// El resolver de modo local no tiene cliente HTTP: la URL queda fuera de alcance sin red.
#[tokio::test]
async fn local_only_resolver_never_reaches_the_network() {
    let resolver = FsHttpReferenceResolver::local_only(std::env::temp_dir());
    assert!(!resolver.follows_remote());
    let r = resolver
        .resolve(&ReferenceOrigin::PublicUrl("https://example.test/x".into()))
        .await;
    assert_eq!(r.state(), ReferenceState::OutOfScope);
    assert_eq!(r.content(), None);
}

// ===========================================================================
// Contrato: ModelProvider (T020)
// ===========================================================================

fn model_provider_contract(provider: &dyn ModelProvider, expect_local: bool, label: &str) {
    assert_eq!(
        provider.is_local(),
        expect_local,
        "[{label}] localidad declarada"
    );
    assert!(
        !provider.name().is_empty(),
        "[{label}] el proveedor se identifica"
    );
    assert!(matches!(provider.tier_hint(), Tier::Cheap | Tier::Heavy));
}

#[tokio::test]
async fn model_provider_contract_holds_for_the_real_local_adapter() {
    // No requiere que Ollama esté corriendo: el contrato de identidad/localidad es estático.
    let provider = LocalOpenAiCompatProvider::ollama("qwen2.5-coder").unwrap();
    model_provider_contract(&provider, true, "ollama");
}

#[tokio::test]
async fn model_provider_contract_holds_for_the_fake() {
    model_provider_contract(
        &FakeModelProvider::local("fake-local", vec![]),
        true,
        "fake-local",
    );
    model_provider_contract(
        &FakeModelProvider::remote("fake-remoto"),
        false,
        "fake-remoto",
    );
}

/// Invariante de producto: un proveedor "local" no puede apuntar fuera de loopback.
#[test]
fn a_local_provider_cannot_be_built_against_a_remote_endpoint() {
    assert!(LocalOpenAiCompatProvider::new("x", "https://api.remoto.test", "m").is_err());
    assert!(LocalOpenAiCompatProvider::new("x", "http://127.0.0.1:11434", "m").is_ok());
}
