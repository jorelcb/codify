//! Conexiones a proveedores remotos (`003`-T012).
//!
//! Vive en `application/` y no en `domain/`: el Dominio de `001` habla de sesión, referencia y
//! artefacto. «Cuenta conectada» es vocabulario de esta capa.

use crate::application::ports::{ReferenciaDeCredencial, Tier};
use serde::{Deserialize, Serialize};

/// En qué estado está una cuenta conectada.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ConnectionState {
    Connected,
    /// La credencial caducó; hace falta reconectar.
    Expired,
    /// El usuario o el proveedor la revocaron.
    Revoked,
    /// La credencial **ya no está en el almacén**: el usuario limpió su llavero, u otra
    /// aplicación la borró.
    ///
    /// Existe porque sin ella la conexión se mostraba `Connected` y no funcionaba: el sistema la
    /// omitía en silencio al armar el grafo y el usuario no tenía dónde enterarse. Un estado que
    /// miente es peor que uno que falta.
    CredentialMissing,
}

impl ConnectionState {
    /// Código estable para el catálogo de la piel, como `ProviderIssue` o `SessionFailure`.
    pub fn code(&self) -> &'static str {
        match self {
            ConnectionState::Connected => "connected",
            ConnectionState::Expired => "expired",
            ConnectionState::Revoked => "revoked",
            ConnectionState::CredentialMissing => "credential_missing",
        }
    }

    pub fn all() -> [ConnectionState; 4] {
        [
            ConnectionState::Connected,
            ConnectionState::Expired,
            ConnectionState::Revoked,
            ConnectionState::CredentialMissing,
        ]
    }
}

/// Una cuenta remota autorizada.
///
/// **No contiene la credencial**, y no tiene campo donde pudiera ir: lleva la referencia con la
/// que pedírsela al almacén. Un `ProviderConnection` serializado a la interfaz o a un registro no
/// puede filtrar un secreto porque no lo tiene — que es más fiable que acordarse de no ponerlo.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ProviderConnection {
    pub id: String,
    /// Lo que el usuario ve.
    pub label: String,
    /// Solo el **host**: es lo que FR-009 necesita para decir quién podría recibir contenido,
    /// sin exponer una URL con credenciales embebidas.
    pub endpoint_host: String,
    /// **Declarado al conectar**, no inferido: el sistema no puede saber si un endpoint sirve un
    /// modelo caro o barato, y adivinarlo produciría un reparto arbitrario.
    pub tier: Tier,
    pub state: ConnectionState,
    /// Con qué pedir el secreto al almacén. Se persiste porque **no es el secreto**; sin ella
    /// la cuenta no sobreviviría a reiniciar.
    credential: ReferenciaDeCredencial,
}

impl ProviderConnection {
    pub fn new(
        id: impl Into<String>,
        label: impl Into<String>,
        endpoint_host: impl Into<String>,
        tier: Tier,
        credential: ReferenciaDeCredencial,
    ) -> Self {
        Self {
            id: id.into(),
            label: label.into(),
            endpoint_host: endpoint_host.into(),
            tier,
            state: ConnectionState::Connected,
            credential,
        }
    }

    /// Con qué referencia pedir el secreto al almacén. No devuelve el secreto.
    pub fn credential(&self) -> &ReferenciaDeCredencial {
        &self.credential
    }

    pub fn usable(&self) -> bool {
        self.state == ConnectionState::Connected
    }
}
