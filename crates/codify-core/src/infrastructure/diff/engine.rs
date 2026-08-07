//! Adapter de `DiffEngine` sobre el crate `similar` (T035).
//!
//! El `Diff` del dominio lleva **ambos lados** (`before` y `after`), así que aplicar y revertir
//! no requieren interpretar el texto unificado: son reemplazos verificados. El texto unificado
//! existe para **la persona que decide** — es lo que la piel renderiza.
//!
//! Esa separación es deliberada. Reconstruir el archivo parseando un diff unificado es una
//! fuente clásica de corrupción silenciosa; aquí el parseo no está en el camino crítico y un
//! fallo de renderizado no puede estropear un archivo.

use crate::application::ports::DiffEngine;
use crate::domain::change::Diff;
use crate::domain::error::{CoreError, Result};
use similar::TextDiff;

pub struct SimilarDiffEngine;

impl SimilarDiffEngine {
    /// Comprueba que el texto sobre el que se va a operar es el que el diff espera.
    ///
    /// Sin esto, aplicar un diff a un archivo que cambió por debajo lo sobrescribiría con una
    /// versión basada en un original que ya no existe — y el usuario habría "aprobado" un
    /// cambio que no es el que se le mostró.
    fn ensure_matches(actual: &str, expected: &str, operacion: &str) -> Result<()> {
        if actual == expected {
            return Ok(());
        }
        Err(CoreError::Invalid(format!(
            "no se puede {operacion}: el contenido actual no es el que este diff esperaba \
             (el archivo cambió desde que se propuso)"
        )))
    }
}

impl DiffEngine for SimilarDiffEngine {
    fn make(&self, before: &str, after: &str) -> Diff {
        let unified = TextDiff::from_lines(before, after)
            .unified_diff()
            .context_radius(3)
            .header("antes", "después")
            .to_string();

        Diff {
            unified,
            before: before.to_string(),
            after: after.to_string(),
        }
    }

    fn apply(&self, before: &str, diff: &Diff) -> Result<String> {
        Self::ensure_matches(before, &diff.before, "aplicar el cambio")?;
        Ok(diff.after.clone())
    }

    fn revert(&self, after: &str, diff: &Diff) -> Result<String> {
        Self::ensure_matches(after, &diff.after, "revertir el cambio")?;
        Ok(diff.before.clone())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn an_empty_change_produces_an_empty_diff() {
        let d = SimilarDiffEngine.make("igual\n", "igual\n");
        assert!(d.is_empty());
    }

    /// El texto unificado lleva contexto alrededor del cambio: revisar una línea suelta sin
    /// saber dónde cae no permite decidir.
    #[test]
    fn the_unified_text_carries_surrounding_context() {
        let antes = "uno\ndos\ntres\nCUATRO\ncinco\nseis\nsiete\n";
        let despues = "uno\ndos\ntres\ncuatro\ncinco\nseis\nsiete\n";
        let d = SimilarDiffEngine.make(antes, despues);
        assert!(
            d.unified.contains("tres"),
            "falta el contexto previo:\n{}",
            d.unified
        );
        assert!(
            d.unified.contains("cinco"),
            "falta el contexto posterior:\n{}",
            d.unified
        );
    }

    #[test]
    fn applying_to_the_wrong_text_explains_why_it_refused() {
        let d = SimilarDiffEngine.make("a\n", "b\n");
        let err = SimilarDiffEngine.apply("otra cosa\n", &d).unwrap_err();
        assert!(
            err.to_string().contains("cambió"),
            "el motivo tiene que ser accionable, era: {err}"
        );
    }
}
