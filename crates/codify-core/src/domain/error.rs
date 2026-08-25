//! Errores de dominio. Los adapters traducen los errores técnicos a estas variantes
//! **antes** de cruzar la frontera del núcleo (constitución, Principio I).

use thiserror::Error;

#[derive(Debug, Error)]
pub enum CoreError {
    #[error("recurso no encontrado: {0}")]
    NotFound(String),

    #[error("entrada inválida: {0}")]
    Invalid(String),

    #[error("acceso denegado o no autorizado: {0}")]
    Unauthorized(String),

    #[error("la capacidad no está disponible: {0}")]
    Unavailable(String),

    #[error("el proveedor de modelo falló: {0}")]
    Provider(String),

    /// El proveedor no contestó a tiempo. Va aparte de `Provider` porque el diagnóstico es el
    /// **opuesto**: aquí se espera más o se usa un modelo más rápido; allí se revisa el prompt.
    /// Confundirlos costó cinco corridas el 2026-08-24 (`002`-FR-028, issue #24).
    #[error("el proveedor de modelo no respondió a tiempo: {0}")]
    ProviderTimeout(String),

    #[error("operación bloqueada por la política de cero-egress: {0}")]
    EgressBlocked(String),

    #[error("fallo de almacenamiento: {0}")]
    Storage(String),

    #[error("la sesión fue cancelada")]
    Cancelled,
}

pub type Result<T> = std::result::Result<T, CoreError>;
