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

    #[error("operación bloqueada por la política de cero-egress: {0}")]
    EgressBlocked(String),

    #[error("fallo de almacenamiento: {0}")]
    Storage(String),
}

pub type Result<T> = std::result::Result<T, CoreError>;
