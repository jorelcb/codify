//! `CredentialStore` contra el almacén del sistema operativo (`003`-FR-002, T013).
//!
//! **Sin respaldo en archivo, a propósito** (research.md D3). Un archivo cifrado solo movería el
//! problema —¿dónde vive la llave?— y daría una promesa de custodia que este producto no puede
//! sostener. Si el almacén del sistema no está, se dice y se ofrece seguir en local.

use crate::application::ports::{CredentialStore, ReferenciaDeCredencial, Secreto};
use crate::domain::error::{CoreError, Result};

const SERVICIO: &str = "dev.jorelcb.codify";

pub struct SystemKeyring;

impl SystemKeyring {
    pub fn new() -> Self {
        Self
    }

    fn entrada(r: &ReferenciaDeCredencial) -> Result<keyring::Entry> {
        keyring::Entry::new(SERVICIO, r.as_str())
            .map_err(|e| CoreError::Storage(format!("almacén de credenciales: {e}")))
    }
}

impl Default for SystemKeyring {
    fn default() -> Self {
        Self::new()
    }
}

#[async_trait::async_trait]
impl CredentialStore for SystemKeyring {
    fn disponible(&self) -> bool {
        // Se comprueba **sin guardar nada**: FR-004 exige poder avisar antes de que el usuario
        // intente conectar. Abrir una entrada basta para saber si hay almacén detrás.
        keyring::Entry::new(SERVICIO, "__sonda__").is_ok()
    }

    async fn guardar(&self, r: &ReferenciaDeCredencial, s: Secreto) -> Result<()> {
        Self::entrada(r)?
            .set_password(s.exponer_para_la_peticion())
            .map_err(|e| CoreError::Storage(format!("no se pudo guardar la credencial: {e}")))
    }

    async fn obtener(&self, r: &ReferenciaDeCredencial) -> Result<Option<Secreto>> {
        match Self::entrada(r)?.get_password() {
            Ok(v) => Ok(Some(Secreto::new(v))),
            Err(keyring::Error::NoEntry) => Ok(None),
            Err(e) => Err(CoreError::Storage(format!(
                "no se pudo leer la credencial: {e}"
            ))),
        }
    }

    async fn borrar(&self, r: &ReferenciaDeCredencial) -> Result<()> {
        // Idempotente: desconectar dos veces no es un error (contracts/ports.md).
        match Self::entrada(r)?.delete_credential() {
            Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
            Err(e) => Err(CoreError::Storage(format!(
                "no se pudo borrar la credencial: {e}"
            ))),
        }
    }
}
