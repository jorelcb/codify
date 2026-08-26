//! Proveedor remoto genérico contra API compatible con OpenAI (`003`-T020).
//!
//! **Uno solo, no uno por proveedor** (research.md D2). `001`-FR-016 exige no acoplarse a un
//! proveedor específico, y el adapter local ya demostró que el patrón funciona: sirve Ollama y
//! `llama.cpp` sin saber cuál es cuál.
//!
//! Reutiliza el armado y el parseo de la petición del adapter local: la diferencia entre los dos
//! es **a dónde van** y **si llevan credencial**, no cómo se habla.

use crate::application::ports::{
    CompletionOutput, CompletionRequest, ModelProvider, Secreto, Tier,
};
use crate::domain::error::{CoreError, Result};
use crate::infrastructure::providers::local::LocalOpenAiCompatProvider;
use std::time::Duration;

/// Igual que el del proveedor local, y por la misma razón: un modelo grande tarda minutos.
const TIEMPO_MAXIMO: Duration = Duration::from_secs(900);

pub struct RemoteOpenAiCompatProvider {
    name: String,
    base_url: String,
    model: String,
    tier: Tier,
    credencial: Secreto,
    http: reqwest::Client,
}

impl RemoteOpenAiCompatProvider {
    /// El `tier` se **declara**, no se infiere: el sistema no puede saber si un endpoint sirve un
    /// modelo caro o barato, y adivinarlo produciría un reparto arbitrario (data-model.md).
    pub fn new(
        name: impl Into<String>,
        base_url: impl Into<String>,
        model: impl Into<String>,
        tier: Tier,
        credencial: Secreto,
    ) -> Result<Self> {
        Ok(Self {
            name: name.into(),
            base_url: base_url.into(),
            model: model.into(),
            tier,
            credencial,
            http: reqwest::Client::builder()
                .timeout(TIEMPO_MAXIMO)
                .build()
                .map_err(|e| CoreError::Provider(e.to_string()))?,
        })
    }
}

#[async_trait::async_trait]
impl ModelProvider for RemoteOpenAiCompatProvider {
    async fn complete(&self, request: CompletionRequest) -> Result<CompletionOutput> {
        // El armado de la petición es el mismo que el local: hablan el mismo protocolo.
        let payload = LocalOpenAiCompatProvider::payload_para(&self.model, &request);
        let url = format!(
            "{}/v1/chat/completions",
            self.base_url.trim_end_matches('/')
        );

        let resp = self
            .http
            .post(&url)
            .bearer_auth(self.credencial.exponer_para_la_peticion())
            .json(&payload)
            .send()
            .await
            .map_err(|e| {
                if e.is_timeout() {
                    CoreError::ProviderTimeout(format!(
                        "{} no respondió en {} s",
                        self.name,
                        TIEMPO_MAXIMO.as_secs()
                    ))
                } else if e.is_connect() {
                    CoreError::Unavailable(format!("{} no está disponible: {e}", self.name))
                } else {
                    CoreError::Provider(format!("{} no respondió: {e}", self.name))
                }
            })?;

        // `002`-FR-028: un fallo de autorización pide algo distinto del usuario que un fallo del
        // modelo —reconectar frente a reintentar—, así que se distingue aquí, donde se sabe.
        if resp.status() == reqwest::StatusCode::UNAUTHORIZED
            || resp.status() == reqwest::StatusCode::FORBIDDEN
        {
            return Err(CoreError::Unauthorized(format!(
                "{} rechazó la credencial ({})",
                self.name,
                resp.status()
            )));
        }
        if !resp.status().is_success() {
            return Err(CoreError::Provider(format!(
                "{} devolvió {}",
                self.name,
                resp.status()
            )));
        }

        let body = resp
            .json()
            .await
            .map_err(|e| CoreError::Provider(e.to_string()))?;
        LocalOpenAiCompatProvider::parse_response(&body)
    }

    fn is_local(&self) -> bool {
        // Es lo que `ProviderRegistry::for_mode` usa para rechazarlo en modo local — la defensa
        // en profundidad que acompaña a la del tipo.
        false
    }

    fn tier_hint(&self) -> Tier {
        self.tier
    }

    fn name(&self) -> &str {
        &self.name
    }
}
