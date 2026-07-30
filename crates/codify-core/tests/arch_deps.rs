//! **Fitness function de arquitectura** — constitución, Principio I [NON-NEGOTIABLE].
//!
//! La Regla de Dependencia no se sostiene con esperanza: se verifica. Estos tests fallan el
//! build si alguien hace que el Dominio mire hacia afuera o que la Aplicación conozca la
//! Infraestructura.

use std::fs;
use std::path::{Path, PathBuf};

fn rust_files(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let entries = match fs::read_dir(dir) {
        Ok(e) => e,
        Err(_) => return out,
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if path.is_dir() {
            out.extend(rust_files(&path));
        } else if path.extension().is_some_and(|e| e == "rs") {
            out.push(path);
        }
    }
    out
}

fn layer_dir(layer: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("src")
        .join(layer)
}

/// Busca patrones prohibidos ignorando comentarios de línea (la documentación sí puede
/// nombrar la capa de la que se independiza).
fn offending_lines(file: &Path, forbidden: &[&str]) -> Vec<String> {
    let content = fs::read_to_string(file).unwrap_or_default();
    content
        .lines()
        .filter(|line| {
            let trimmed = line.trim_start();
            !trimmed.starts_with("//") && !trimmed.starts_with("*")
        })
        .filter(|line| forbidden.iter().any(|f| line.contains(f)))
        .map(|l| format!("{}: {}", file.display(), l.trim()))
        .collect()
}

fn assert_layer_clean(layer: &str, forbidden: &[&str]) {
    let dir = layer_dir(layer);
    assert!(
        dir.exists(),
        "no existe la capa '{layer}' en {}",
        dir.display()
    );

    let violations: Vec<String> = rust_files(&dir)
        .iter()
        .flat_map(|f| offending_lines(f, forbidden))
        .collect();

    assert!(
        violations.is_empty(),
        "Regla de Dependencia violada en la capa '{layer}'.\nPatrones prohibidos: {forbidden:?}\nOcurrencias:\n{}",
        violations.join("\n")
    );
}

#[test]
fn domain_never_imports_outer_layers() {
    assert_layer_clean("domain", &["crate::application", "crate::infrastructure"]);
}

#[test]
fn application_never_imports_infrastructure() {
    assert_layer_clean("application", &["crate::infrastructure"]);
}

/// El Dominio es puro: sin I/O, sin red, sin clientes HTTP. Si necesita tiempo o disco,
/// lo pide por un port.
#[test]
fn domain_is_free_of_io_and_network() {
    assert_layer_clean(
        "domain",
        &[
            "reqwest",
            "std::fs",
            "std::net",
            "tokio::net",
            "tokio::fs",
            "walkdir",
        ],
    );
}

/// La Aplicación orquesta, no ejecuta efectos: el I/O concreto es de Infraestructura.
#[test]
fn application_is_free_of_direct_io_and_network() {
    assert_layer_clean(
        "application",
        &["reqwest", "std::fs", "std::net", "walkdir"],
    );
}

/// Las capas esperadas existen (evita que un refactor las disuelva en silencio).
#[test]
fn expected_layers_exist() {
    for layer in ["domain", "application", "infrastructure"] {
        assert!(layer_dir(layer).exists(), "falta la capa '{layer}'");
    }
}
