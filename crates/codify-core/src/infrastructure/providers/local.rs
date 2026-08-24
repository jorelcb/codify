//! Proveedor local con API **OpenAI-compatible**: cubre Ollama y `llama.cpp` server con un
//! solo adapter (decisión de `research.md`).
//!
//! Garantía estructural de cero-egress: el constructor **rechaza** cualquier endpoint que no
//! sea loopback. Por construcción `is_local()` no puede mentir.

use crate::application::ports::{
    CompletionOutput, CompletionRequest, ModelProvider, Role, Tier, ToolCall,
};
use crate::domain::error::{CoreError, Result};
use async_trait::async_trait;
use serde_json::{json, Value};
use std::time::Duration;

/// Hosts admitidos como locales.
const LOOPBACK_HOSTS: &[&str] = &["localhost", "127.0.0.1", "::1", "0.0.0.0"];

pub struct LocalOpenAiCompatProvider {
    name: String,
    base_url: String,
    model: String,
    tier: Tier,
    http: reqwest::Client,
}

/// Cuánto se espera a un modelo local antes de darlo por perdido.
///
/// Los 120 s que había aquí antes no daban. Medido el 2026-08-23 con un Qwen2.5-32B en
/// `llama.cpp`: la generación de un solo artefacto iba por 690 tokens a 5,8 t/s —dos minutos
/// largos— cuando el cliente HTTP cortó. La sesión caía a `Failed` sin motivo, y el fallo
/// parecía un misterio del modelo cuando era nuestro reloj.
///
/// Un modelo local grande en hardware de sobremesa va a ese ritmo, y la petición de citas
/// textuales (FR-006a) alarga la salida, así que el margen tiene que dar para eso. Esperar no
/// bloquea a nadie: el trabajo va en segundo plano y la sesión es cancelable (FR-022/FR-023).
const TIEMPO_MAXIMO: Duration = Duration::from_secs(900);

impl LocalOpenAiCompatProvider {
    /// Construye el proveedor **solo** si el endpoint es loopback.
    /// Un endpoint remoto es un error de construcción, no una advertencia en runtime.
    pub fn new(
        name: impl Into<String>,
        base_url: impl Into<String>,
        model: impl Into<String>,
    ) -> Result<Self> {
        let base_url = base_url.into();
        if !Self::is_loopback(&base_url) {
            return Err(CoreError::EgressBlocked(format!(
                "endpoint no local rechazado para un proveedor local: {base_url}"
            )));
        }
        let http = reqwest::Client::builder()
            .timeout(TIEMPO_MAXIMO)
            .build()
            .map_err(|e| CoreError::Provider(e.to_string()))?;
        Ok(Self {
            name: name.into(),
            base_url,
            model: model.into(),
            tier: Tier::Cheap,
            http,
        })
    }

    /// Atajo para Ollama en su puerto por defecto.
    pub fn ollama(model: impl Into<String>) -> Result<Self> {
        Self::new("ollama", "http://localhost:11434", model)
    }

    /// Atajo para `llama.cpp` server en su puerto por defecto.
    pub fn llama_cpp(model: impl Into<String>) -> Result<Self> {
        Self::new("llamacpp", "http://localhost:8080", model)
    }

    pub fn with_tier(mut self, tier: Tier) -> Self {
        self.tier = tier;
        self
    }

    /// Función pura, testeable sin red.
    pub fn is_loopback(base_url: &str) -> bool {
        let without_scheme = base_url
            .trim_start_matches("http://")
            .trim_start_matches("https://");
        let host = without_scheme
            .split('/')
            .next()
            .unwrap_or("")
            .rsplit_once(':')
            .map(|(h, _)| h)
            .unwrap_or_else(|| without_scheme.split('/').next().unwrap_or(""));
        let host = host.trim_start_matches('[').trim_end_matches(']');
        LOOPBACK_HOSTS.contains(&host)
    }

    fn role_str(role: Role) -> &'static str {
        match role {
            Role::User => "user",
            Role::Assistant => "assistant",
            Role::Tool => "tool",
        }
    }

    /// Traduce la petición de dominio al payload OpenAI-compatible.
    /// Pura: permite testear el mapeo sin levantar un servidor.
    pub fn to_payload(&self, request: &CompletionRequest) -> Value {
        let mut messages = vec![json!({"role": "system", "content": request.system})];
        for m in &request.messages {
            messages.push(json!({"role": Self::role_str(m.role), "content": m.content}));
        }
        let tools: Vec<Value> = request
            .tools
            .iter()
            .map(|t| {
                let params: Value =
                    serde_json::from_str(&t.json_schema).unwrap_or_else(|_| json!({}));
                json!({
                    "type": "function",
                    "function": {
                        "name": t.name,
                        "description": t.description,
                        "parameters": params,
                    }
                })
            })
            .collect();

        let mut payload = json!({
            "model": self.model,
            "messages": messages,
            "stream": false,
        });
        if !tools.is_empty() {
            payload["tools"] = Value::Array(tools);
        }
        payload
    }

    /// Traduce la respuesta del backend al tipo de dominio.
    /// Pura: es donde se prueba el parseo de tool-calls.
    pub fn parse_response(body: &Value) -> Result<CompletionOutput> {
        let message = body
            .get("choices")
            .and_then(|c| c.get(0))
            .and_then(|c| c.get("message"))
            .ok_or_else(|| CoreError::Provider("respuesta sin 'choices[0].message'".into()))?;

        if let Some(calls) = message.get("tool_calls").and_then(|c| c.as_array()) {
            if !calls.is_empty() {
                let parsed = calls
                    .iter()
                    .enumerate()
                    .map(|(i, c)| {
                        let f = c.get("function");
                        ToolCall {
                            id: c
                                .get("id")
                                .and_then(|v| v.as_str())
                                .unwrap_or(&format!("call-{i}"))
                                .to_string(),
                            name: f
                                .and_then(|f| f.get("name"))
                                .and_then(|v| v.as_str())
                                .unwrap_or_default()
                                .to_string(),
                            arguments: f
                                .and_then(|f| f.get("arguments"))
                                .map(|v| match v {
                                    Value::String(s) => s.clone(),
                                    other => other.to_string(),
                                })
                                .unwrap_or_else(|| "{}".to_string()),
                        }
                    })
                    .collect();
                return Ok(CompletionOutput::ToolCalls(parsed));
            }
        }

        let text = message
            .get("content")
            .and_then(|v| v.as_str())
            .unwrap_or_default();
        Ok(CompletionOutput::Text(text.to_string()))
    }
}

#[async_trait]
impl ModelProvider for LocalOpenAiCompatProvider {
    async fn complete(&self, request: CompletionRequest) -> Result<CompletionOutput> {
        let url = format!(
            "{}/v1/chat/completions",
            self.base_url.trim_end_matches('/')
        );
        let resp = self
            .http
            .post(&url)
            .json(&self.to_payload(&request))
            .send()
            .await
            .map_err(|e| CoreError::Provider(format!("{} no respondió: {e}", self.name)))?;

        if !resp.status().is_success() {
            return Err(CoreError::Provider(format!(
                "{} devolvió {}",
                self.name,
                resp.status()
            )));
        }
        let body: Value = resp
            .json()
            .await
            .map_err(|e| CoreError::Provider(e.to_string()))?;
        Self::parse_response(&body)
    }

    fn is_local(&self) -> bool {
        // Garantizado por construcción: el constructor rechaza endpoints no loopback.
        true
    }

    fn tier_hint(&self) -> Tier {
        self.tier
    }

    fn name(&self) -> &str {
        &self.name
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::application::ports::{Message, ToolSpec};

    #[test]
    fn rejects_non_loopback_endpoints_at_construction() {
        let err = LocalOpenAiCompatProvider::new("x", "https://api.remoto.test", "m");
        assert!(matches!(err, Err(CoreError::EgressBlocked(_))));
    }

    #[test]
    fn accepts_loopback_endpoints() {
        assert!(LocalOpenAiCompatProvider::is_loopback(
            "http://localhost:11434"
        ));
        assert!(LocalOpenAiCompatProvider::is_loopback(
            "http://127.0.0.1:8080"
        ));
        assert!(!LocalOpenAiCompatProvider::is_loopback(
            "https://api.anthropic.com"
        ));
        assert!(!LocalOpenAiCompatProvider::is_loopback(
            "http://evil.test:11434"
        ));
    }

    #[test]
    fn maps_request_to_openai_payload_with_tools() {
        let p = LocalOpenAiCompatProvider::ollama("qwen").unwrap();
        let req = CompletionRequest {
            system: "sys".into(),
            messages: vec![Message::user("hola")],
            tools: vec![ToolSpec {
                name: "read_file".into(),
                description: "lee".into(),
                json_schema: r#"{"type":"object"}"#.into(),
            }],
        };
        let payload = p.to_payload(&req);
        assert_eq!(payload["messages"][0]["role"], "system");
        assert_eq!(payload["messages"][1]["content"], "hola");
        assert_eq!(payload["tools"][0]["function"]["name"], "read_file");
    }

    #[test]
    fn parses_text_response() {
        let body = json!({"choices":[{"message":{"content":"hola"}}]});
        assert_eq!(
            LocalOpenAiCompatProvider::parse_response(&body).unwrap(),
            CompletionOutput::Text("hola".into())
        );
    }

    #[test]
    fn parses_tool_calls_response() {
        let body = json!({"choices":[{"message":{"tool_calls":[
            {"id":"c1","function":{"name":"read_file","arguments":"{\"path\":\"README.md\"}"}}
        ]}}]});
        match LocalOpenAiCompatProvider::parse_response(&body).unwrap() {
            CompletionOutput::ToolCalls(calls) => {
                assert_eq!(calls.len(), 1);
                assert_eq!(calls[0].name, "read_file");
                assert!(calls[0].arguments.contains("README.md"));
            }
            other => panic!("se esperaban tool calls, llegó {other:?}"),
        }
    }
}
