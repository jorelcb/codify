//! `AccountConnector` por credencial introducida una sola vez (`003`-FR-001, T015).
//!
//! Es la vía para los proveedores que no ofrecen autorización delegada, que son la mayoría de
//! los frontier. Lo que la clarificación del 2026-08-26 prohibió no es introducir la credencial
//! —sin eso este spec no sirve a los proveedores que lo motivan— sino que quede en un archivo
//! del proyecto, en la configuración o en un registro. De eso se encarga `CredentialStore`.

use crate::application::ports::{AccountConnector, Desafio, InstruccionDeCredencial, Secreto};
use crate::domain::error::{CoreError, Result};

pub struct DirectCredential {
    instruccion: InstruccionDeCredencial,
}

impl DirectCredential {
    /// `instruccion` es **qué** hay que decirle al usuario, no cómo decirlo: la piel lo redacta en
    /// el idioma elegido. Recibía una `String`, y quien lo construía le pasaba una frase en
    /// español que nunca llegó a verse porque la interfaz ignoraba ese campo.
    pub fn new(instruccion: InstruccionDeCredencial) -> Self {
        Self { instruccion }
    }
}

#[async_trait::async_trait]
impl AccountConnector for DirectCredential {
    async fn iniciar(&self) -> Result<Desafio> {
        Ok(Desafio::PideCredencial {
            instruccion: self.instruccion,
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
