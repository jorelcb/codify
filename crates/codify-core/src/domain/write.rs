//! Registro de lo que llegó (o no) al repositorio.
//!
//! El Dominio nombra este concepto porque la sesión lo **reporta**: es la respuesta a
//! "¿qué se escribió?" (FR-017) y a "¿qué alcanzó a escribirse antes de cancelar?" (FR-023).
//! Sin él, el usuario tendría que inspeccionar archivos para saber en qué estado quedó su
//! repositorio — exactamente lo que el producto promete evitar.

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum WriteOutcome {
    Written,
    /// No hacía falta escribir (contenido idéntico al existente).
    Skipped(String),
    /// Falló, con su motivo. **No aborta** el resto de la sesión.
    Failed(String),
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct WriteRecord {
    /// Ruta relativa a la raíz del repositorio.
    pub path: String,
    pub bytes: usize,
    pub at: String,
    pub outcome: WriteOutcome,
}

impl WriteRecord {
    pub fn written(path: impl Into<String>, bytes: usize, at: impl Into<String>) -> Self {
        Self {
            path: path.into(),
            bytes,
            at: at.into(),
            outcome: WriteOutcome::Written,
        }
    }

    pub fn skipped(path: impl Into<String>, at: impl Into<String>, why: impl Into<String>) -> Self {
        Self {
            path: path.into(),
            bytes: 0,
            at: at.into(),
            outcome: WriteOutcome::Skipped(why.into()),
        }
    }

    pub fn failed(path: impl Into<String>, at: impl Into<String>, why: impl Into<String>) -> Self {
        Self {
            path: path.into(),
            bytes: 0,
            at: at.into(),
            outcome: WriteOutcome::Failed(why.into()),
        }
    }

    /// `true` solo si el contenido llegó de verdad al disco.
    pub fn reached_disk(&self) -> bool {
        matches!(self.outcome, WriteOutcome::Written)
    }

    /// Descripción legible para la auditoría y para la interfaz.
    pub fn summary(&self) -> String {
        match &self.outcome {
            WriteOutcome::Written => format!("{} ({} bytes)", self.path, self.bytes),
            WriteOutcome::Skipped(why) => format!("{} (omitido: {why})", self.path),
            WriteOutcome::Failed(why) => format!("{} (falló: {why})", self.path),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn only_a_written_record_reached_disk() {
        assert!(WriteRecord::written("AGENTS.md", 120, "t0").reached_disk());
        assert!(!WriteRecord::skipped("AGENTS.md", "t0", "sin cambios").reached_disk());
        assert!(!WriteRecord::failed("AGENTS.md", "t0", "permiso denegado").reached_disk());
    }

    #[test]
    fn summary_states_why_when_something_did_not_reach_disk() {
        assert!(WriteRecord::failed("x.md", "t0", "permiso denegado")
            .summary()
            .contains("permiso denegado"));
        assert!(WriteRecord::skipped("x.md", "t0", "sin cambios")
            .summary()
            .contains("sin cambios"));
    }
}
