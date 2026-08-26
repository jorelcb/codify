//! `003`-FR-008 — un grafo local **no puede** contener un proveedor capaz de salir a la red.
//!
//! Este archivo existe para **no compilar**. Si algún día compila, la garantía de cero-egress
//! dejó de estar en el sistema de tipos y volvió a depender de que alguien se acuerde.
//!
//! Nótese que el argumento da igual: lo que se comprueba es que **el método no existe** en
//! `CoreBuilder<Local>`. Usar aquí el adapter remoto de verdad ataría este test a una tarea de
//! una fase posterior sin ganar nada — el error sería el mismo.

use codify_core::infrastructure::composition::{CoreBuilder, Local};
use codify_core::infrastructure::providers::local::LocalOpenAiCompatProvider;
use std::sync::Arc;

fn main() {
    let cualquiera = Arc::new(
        LocalOpenAiCompatProvider::new("x", "http://127.0.0.1:1", "m").expect("loopback"),
    );

    // `remote_provider` NO existe en `CoreBuilder<Local>`. No es que lo rechace en tiempo de
    // ejecución: no hay método al que llamar.
    let _ = CoreBuilder::<Local>::new().remote_provider(cualquiera);
}
