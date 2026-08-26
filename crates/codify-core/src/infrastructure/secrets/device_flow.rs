//! `AccountConnector` por autorización delegada (`003`-FR-001, T014).
//!
//! La aplicación enseña un código y una dirección; el usuario autoriza fuera y aquí se sondea
//! hasta que el proveedor confirma. **El límite de tiempo es propio**, no el de generación: el
//! de generación son 900 s porque un modelo grande tarda minutos, y esperar a una persona es
//! otra cosa — quien abandona no debería dejar la aplicación colgada un cuarto de hora.

use crate::application::ports::{AccountConnector, Desafio, Secreto};
use crate::domain::error::{CoreError, Result};
use std::time::Duration;

/// Cuánto se espera a que una persona autorice. Ver la nota del módulo.
const ESPERA_HUMANA: Duration = Duration::from_secs(300);

pub struct DeviceFlow {
    device_url: String,
    token_url: String,
    client_id: String,
    espera: Duration,
    http: reqwest::Client,
}

impl DeviceFlow {
    pub fn new(
        device_url: impl Into<String>,
        token_url: impl Into<String>,
        client_id: impl Into<String>,
    ) -> Result<Self> {
        Ok(Self {
            device_url: device_url.into(),
            token_url: token_url.into(),
            client_id: client_id.into(),
            espera: ESPERA_HUMANA,
            http: reqwest::Client::builder()
                .timeout(Duration::from_secs(30))
                .build()
                .map_err(|e| CoreError::Provider(e.to_string()))?,
        })
    }

    /// Reloj a medida, para poder probar el abandono sin esperar cinco minutos.
    pub fn con_espera(mut self, d: Duration) -> Self {
        self.espera = d;
        self
    }
}

#[async_trait::async_trait]
impl AccountConnector for DeviceFlow {
    async fn iniciar(&self) -> Result<Desafio> {
        let resp: serde_json::Value = self
            .http
            .post(&self.device_url)
            .form(&[("client_id", self.client_id.as_str())])
            .send()
            .await
            .map_err(|e| CoreError::Provider(format!("no se pudo iniciar la autorización: {e}")))?
            .json()
            .await
            .map_err(|e| CoreError::Provider(format!("respuesta de autorización ilegible: {e}")))?;

        let codigo = resp["user_code"]
            .as_str()
            .ok_or_else(|| CoreError::Provider("la respuesta no trae 'user_code'".into()))?;
        let url = resp["verification_uri"]
            .as_str()
            .ok_or_else(|| CoreError::Provider("la respuesta no trae 'verification_uri'".into()))?;
        let device_code = resp["device_code"]
            .as_str()
            .ok_or_else(|| CoreError::Provider("la respuesta no trae 'device_code'".into()))?;

        Ok(Desafio::Delegada {
            codigo: codigo.to_string(),
            // El `device_code` viaja pegado a la URL para no añadir un estado que mantener; no es
            // un secreto de larga vida y caduca con el desafío.
            url: format!("{url}#{device_code}"),
        })
    }

    async fn completar(&self, desafio: &Desafio, _respuesta: Option<Secreto>) -> Result<Secreto> {
        let Desafio::Delegada { url, .. } = desafio else {
            return Err(CoreError::Invalid(
                "este conector solo completa desafíos de autorización delegada".into(),
            ));
        };
        let device_code = url.rsplit_once('#').map(|(_, c)| c).unwrap_or_default();

        let limite = tokio::time::Instant::now() + self.espera;
        loop {
            if tokio::time::Instant::now() >= limite {
                // No es un fallo del proveedor: es que nadie autorizó. Decirlo así permite a la
                // piel ofrecer reintentar en vez de sugerir que algo se rompió.
                return Err(CoreError::Unauthorized(
                    "la autorización no se completó a tiempo".into(),
                ));
            }

            let resp: serde_json::Value = self
                .http
                .post(&self.token_url)
                .form(&[
                    ("client_id", self.client_id.as_str()),
                    ("device_code", device_code),
                    ("grant_type", "urn:ietf:params:oauth:grant-type:device_code"),
                ])
                .send()
                .await
                .map_err(|e| CoreError::Provider(format!("sondeo fallido: {e}")))?
                .json()
                .await
                .map_err(|e| CoreError::Provider(format!("respuesta ilegible: {e}")))?;

            if let Some(token) = resp["access_token"].as_str() {
                return Ok(Secreto::new(token));
            }
            match resp["error"].as_str() {
                Some("authorization_pending") | Some("slow_down") => {
                    tokio::time::sleep(Duration::from_secs(2)).await;
                }
                // Denegar es una respuesta, no un error del sistema: se distingue para que el
                // usuario vea qué pasó y pueda reintentar (US1, escenario 4).
                Some("access_denied") => {
                    return Err(CoreError::Unauthorized("la autorización se denegó".into()))
                }
                Some(otro) => {
                    return Err(CoreError::Provider(format!(
                        "el proveedor devolvió: {otro}"
                    )))
                }
                None => return Err(CoreError::Provider("respuesta sin token ni error".into())),
            }
        }
    }
}
