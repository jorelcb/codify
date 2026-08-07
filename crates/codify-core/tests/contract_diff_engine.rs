//! **Contract test del port `DiffEngine`** (T032) — FR-010/FR-014/FR-015.
//!
//! La misma suite corre contra el adapter real y contra el fake in-memory (patrón
//! hex-integration-test). Lo que se asserta es comportamiento de dominio, nunca detalles del
//! adapter: si mañana se cambia `similar` por otra librería, esta suite no se toca.
//!
//! Dos propiedades sostienen la promesa del loop curado:
//!
//! 1. **`apply∘revert = identidad`**. Sin esto, "lo de bajo riesgo se auto-aplica y es
//!    revertible" (FR-010) sería una promesa que no se puede cumplir: revertir dejaría el
//!    archivo en un tercer estado que nadie pidió.
//! 2. **Un diff no se aplica sobre un texto que no es el suyo**. Es la que evita corromper en
//!    silencio un archivo que cambió por debajo — y la que hace que "rechazar ⇒ el archivo no
//!    cambia" (FR-015) sea verificable y no confiada.

mod fakes;

use codify_core::application::ports::DiffEngine;
use codify_core::infrastructure::diff::engine::SimilarDiffEngine;
use fakes::FakeDiffEngine;

const ANTES: &str = "# Proyecto\n\nMotor de orquestación: Kafka.\nPersistencia: PostgreSQL.\n";
const DESPUES: &str = "# Proyecto\n\nMotor de orquestación: Temporal.\nPersistencia: PostgreSQL.\n";

/// La suite que **todo** adapter de `DiffEngine` debe pasar.
fn diff_engine_contract(engine: &dyn DiffEngine, label: &str) {
    // --- Ida y vuelta ------------------------------------------------------
    let diff = engine.make(ANTES, DESPUES);

    let aplicado = engine
        .apply(ANTES, &diff)
        .unwrap_or_else(|e| panic!("[{label}] aplicar un diff recién hecho no puede fallar: {e}"));
    assert_eq!(
        aplicado, DESPUES,
        "[{label}] aplicar el diff tiene que producir exactamente el texto de destino"
    );

    let revertido = engine
        .revert(&aplicado, &diff)
        .unwrap_or_else(|e| panic!("[{label}] revertir lo recién aplicado no puede fallar: {e}"));
    assert_eq!(
        revertido, ANTES,
        "[{label}] apply∘revert debe ser la identidad: si no, 'revertible' es mentira (FR-010)"
    );

    // --- Un diff vacío no es un cambio -------------------------------------
    let sin_cambio = engine.make(ANTES, ANTES);
    assert!(
        sin_cambio.is_empty(),
        "[{label}] un diff entre dos textos iguales tiene que reconocerse como vacío"
    );

    // --- El diff no se aplica sobre un texto ajeno -------------------------
    let otro = "# Otro proyecto\n\nNada que ver.\n";
    assert!(
        engine.apply(otro, &diff).is_err(),
        "[{label}] aplicar un diff sobre un texto que no es su origen corrompería el archivo \
         en silencio: tiene que fallar, no adivinar"
    );
    assert!(
        engine.revert(otro, &diff).is_err(),
        "[{label}] revertir sobre un texto que no es el resultado esperado tiene que fallar"
    );

    // --- El diff es legible: lo va a leer una persona ----------------------
    assert!(
        !diff.unified.trim().is_empty(),
        "[{label}] el diff unificado no puede estar vacío: es lo que el usuario lee para decidir"
    );
    assert!(
        diff.unified.contains("Kafka") && diff.unified.contains("Temporal"),
        "[{label}] el diff tiene que mostrar qué sale y qué entra; era:\n{}",
        diff.unified
    );
}

#[test]
fn contract_holds_for_the_real_adapter() {
    diff_engine_contract(&SimilarDiffEngine, "real");
}

#[test]
fn contract_holds_for_the_in_memory_fake() {
    diff_engine_contract(&FakeDiffEngine, "fake");
}

/// El diff conserva **ambos lados**, que es lo que hace la reversión total y verificable
/// sin tener que reconstruir el original a partir del texto unificado.
#[test]
fn the_diff_carries_both_sides() {
    let diff = SimilarDiffEngine.make(ANTES, DESPUES);
    assert_eq!(diff.before, ANTES);
    assert_eq!(diff.after, DESPUES);
}

/// Un cambio de una sola línea no puede producir un diff que reescriba el archivo entero:
/// lo que el usuario tiene que revisar es el cambio, no el documento.
#[test]
fn the_unified_diff_stays_focused_on_what_changed() {
    let diff = SimilarDiffEngine.make(ANTES, DESPUES);
    let tocadas = diff
        .unified
        .lines()
        .filter(|l| {
            (l.starts_with('+') || l.starts_with('-'))
                && !l.starts_with("+++")
                && !l.starts_with("---")
        })
        .count();
    assert_eq!(
        tocadas, 2,
        "solo cambió una línea: se esperaban una de salida y una de entrada, hubo {tocadas}.\n{}",
        diff.unified
    );
}
