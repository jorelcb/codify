//! Adapter de `ProviderDiscovery`: sondea el backend local de modelo.
//!
//! Sostiene FR-019 (guiar en vez de callar) y FR-028 (fallos accionables). La regla que
//! gobierna todo el archivo: **sondear nunca falla**. Si el backend no responde, se devuelve
//! un estado con un motivo que el usuario pueda accionar.
//!
//! Cero-egress: el constructor **rechaza** cualquier endpoint que no sea loopback, igual que
//! el proveedor local. Sondear no puede convertirse en una vía de salida.

use crate::application::ports::{ProviderDiscovery, ProviderStatus};
use crate::domain::error::{CoreError, Result};
use crate::infrastructure::providers::local::LocalOpenAiCompatProvider;
use async_trait::async_trait;
use serde_json::Value;
use std::time::Duration;

pub struct LocalProviderProbe {
    base_url: String,
    http: reqwest::Client,
}

impl LocalProviderProbe {
    pub fn new(base_url: impl Into<String>) -> Result<Self> {
        let base_url = base_url.into();
        if !LocalOpenAiCompatProvider::is_loopback(&base_url) {
            return Err(CoreError::EgressBlocked(format!(
                "la sonda solo admite endpoints locales: {base_url}"
            )));
        }
        let http = reqwest::Client::builder()
            .timeout(Duration::from_secs(5))
            .build()
            .map_err(|e| CoreError::Provider(e.to_string()))?;
        Ok(Self { base_url, http })
    }

    /// Extrae los nombres de modelo de una respuesta OpenAI-compatible (`data[].id`) o de la
    /// API nativa de Ollama (`models[].name`). Pura: testeable sin red.
    pub fn parse_models(body: &Value) -> Vec<String> {
        if let Some(items) = body.get("data").and_then(|d| d.as_array()) {
            return items
                .iter()
                .filter_map(|m| m.get("id").and_then(|v| v.as_str()))
                .map(String::from)
                .collect();
        }
        if let Some(items) = body.get("models").and_then(|d| d.as_array()) {
            return items
                .iter()
                .filter_map(|m| m.get("name").and_then(|v| v.as_str()))
                .map(String::from)
                .collect();
        }
        Vec::new()
    }

    async fn try_endpoint(&self, path: &str) -> Option<Vec<String>> {
        let url = format!("{}{path}", self.base_url.trim_end_matches('/'));
        let resp = self.http.get(&url).send().await.ok()?;
        if !resp.status().is_success() {
            return None;
        }
        let body: Value = resp.json().await.ok()?;
        Some(Self::parse_models(&body))
    }
}

#[async_trait]
impl ProviderDiscovery for LocalProviderProbe {
    async fn probe(&self) -> ProviderStatus {
        // Primero el camino estándar; luego la API nativa de Ollama.
        for path in ["/v1/models", "/api/tags"] {
            if let Some(models) = self.try_endpoint(path).await {
                if models.is_empty() {
                    return ProviderStatus::unreachable(
                        &self.base_url,
                        "el backend responde pero no tiene ningún modelo instalado; \
                         descarga uno (por ejemplo `ollama pull qwen2.5-coder`)",
                    );
                }
                return ProviderStatus::reachable(&self.base_url, models);
            }
        }

        ProviderStatus::unreachable(
            &self.base_url,
            format!(
                "no hay ningún backend escuchando en {}; arranca uno (por ejemplo `ollama serve`) \
                 o cambia el endpoint",
                self.base_url
            ),
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn rejects_non_loopback_endpoints() {
        assert!(LocalProviderProbe::new("https://api.remoto.test").is_err());
        assert!(LocalProviderProbe::new("http://localhost:11434").is_ok());
    }

    #[test]
    fn parses_openai_compatible_model_list() {
        let body = json!({"data":[{"id":"qwen2.5-coder"},{"id":"llama3"}]});
        assert_eq!(
            LocalProviderProbe::parse_models(&body),
            vec!["qwen2.5-coder", "llama3"]
        );
    }

    #[test]
    fn parses_ollama_native_model_list() {
        let body = json!({"models":[{"name":"qwen2.5-coder:7b"}]});
        assert_eq!(
            LocalProviderProbe::parse_models(&body),
            vec!["qwen2.5-coder:7b"]
        );
    }

    #[test]
    fn an_unknown_shape_yields_no_models_instead_of_panicking() {
        assert!(LocalProviderProbe::parse_models(&json!({"otra": "forma"})).is_empty());
    }
}
