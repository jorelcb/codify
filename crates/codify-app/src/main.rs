//! Piel de codify-NG (fase 1). Linkea `codify-core` **in-process**.
//!
//! Estado: esqueleto. El cableado de la GUI Tauri y de los comandos llega en T029/T030;
//! esta piel sólo debe renderizar diffs y capturar decisiones — la lógica vive en el core.

fn main() {
    println!(
        "codify-NG · piel (esqueleto). Núcleo: codify-core v{}",
        env!("CARGO_PKG_VERSION")
    );
}
