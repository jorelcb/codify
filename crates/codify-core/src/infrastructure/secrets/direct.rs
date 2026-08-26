//! `AccountConnector` por credencial introducida una sola vez (`003`-FR-001, T015).
//!
//! Es la vía para los proveedores que no ofrecen autorización delegada, que son la mayoría de
//! los frontier. Lo que la clarificación del 2026-08-26 prohibió no es introducir la credencial
//! —sin eso este spec no sirve a los proveedores que lo motivan— sino que quede en un archivo
//! del proyecto, en la configuración o en un registro. De eso se encarga `CredentialStore`.

use crate::application::ports::{AccountConnector, Desafio, Secreto};
use crate::domain::error::{CoreError, Result};

pub struct DirectCredential {
    instrucciones: String,
}

impl DirectCredential {
    /// `instrucciones` es lo que la piel enseña al usuario: dónde encontrar su credencial.
    pub fn new(instrucciones: impl Into<String>) -> Self {
        Self {
            instrucciones: instrucciones.into(),
        }
    }
}

#[async_trait::async_trait]
impl AccountConnector for DirectCredential {
    async fn iniciar(&self) -> Result<Desafio> {
        Ok(Desafio::PideCredencial {
            instrucciones: self.instrucciones.clone(),
        })
    }

    async fn completar(&self, desafio: &Desafio, respuesta: Option<Secreto>) -> Result<Secreto> {
        let Desafio::PideCredencial { .. } = desafio else {
            return Err(CoreError::Invalid(
                "este conector solo completa desafíos de credencial directa".into(),
            ));
        };
        // Abandonar deja el sistema como estaba: no hay conexión a medias que limpiar porque no
        // se creó ninguna (US1, escenario 4).
        respuesta
            .ok_or_else(|| CoreError::Unauthorized("no se introdujo ninguna credencial".into()))
    }
}
