//! Ingesta **dirigida por el agente** (decisión raíz D4).
//!
//! El loop expone herramientas y el modelo decide qué abrir, siguiendo las referencias que
//! encuentra. Dos reglas de producto viven aquí:
//! - **Presupuesto acotado**: la exploración termina, no se dispara en un monorepo.
//! - **Nada se omite en silencio**: lo que queda fuera del presupuesto se declara.

use crate::application::ports::ToolSpec;
use serde::Deserialize;

pub const TOOL_LIST_REPO: &str = "list_repo";
pub const TOOL_READ_FILE: &str = "read_file";
pub const TOOL_FETCH_URL: &str = "fetch_url";
pub const TOOL_NOTE_UNRESOLVED: &str = "note_unresolved";
pub const TOOL_FINALIZE: &str = "finalize";

/// Acción que el modelo pide ejecutar. Traducción tipada de un `ToolCall`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AgentAction {
    ListRepo {
        path: String,
    },
    ReadFile {
        path: String,
    },
    FetchUrl {
        url: String,
    },
    NoteUnresolved {
        what: String,
        reason: String,
    },
    Finalize {
        summary: String,
    },
    /// Herramienta desconocida: se reporta al modelo en vez de abortar el loop.
    Unknown {
        name: String,
    },
}

#[derive(Deserialize)]
struct PathArg {
    #[serde(default)]
    path: String,
}

#[derive(Deserialize)]
struct UrlArg {
    #[serde(default)]
    url: String,
}

#[derive(Deserialize)]
struct UnresolvedArg {
    #[serde(default)]
    what: String,
    #[serde(default)]
    reason: String,
}

#[derive(Deserialize)]
struct SummaryArg {
    #[serde(default)]
    summary: String,
}

/// Traduce el `ToolCall` crudo a una acción de dominio. Nunca falla: los argumentos
/// inválidos degradan a valores vacíos y el loop responde al modelo con el error.
pub fn parse_action(name: &str, arguments: &str) -> AgentAction {
    match name {
        TOOL_LIST_REPO => AgentAction::ListRepo {
            path: serde_json::from_str::<PathArg>(arguments)
                .map(|a| a.path)
                .unwrap_or_default(),
        },
        TOOL_READ_FILE => AgentAction::ReadFile {
            path: serde_json::from_str::<PathArg>(arguments)
                .map(|a| a.path)
                .unwrap_or_default(),
        },
        TOOL_FETCH_URL => AgentAction::FetchUrl {
            url: serde_json::from_str::<UrlArg>(arguments)
                .map(|a| a.url)
                .unwrap_or_default(),
        },
        TOOL_NOTE_UNRESOLVED => {
            let a = serde_json::from_str::<UnresolvedArg>(arguments).ok();
            AgentAction::NoteUnresolved {
                what: a.as_ref().map(|a| a.what.clone()).unwrap_or_default(),
                reason: a.map(|a| a.reason).unwrap_or_default(),
            }
        }
        TOOL_FINALIZE => AgentAction::Finalize {
            summary: serde_json::from_str::<SummaryArg>(arguments)
                .map(|a| a.summary)
                .unwrap_or_default(),
        },
        other => AgentAction::Unknown {
            name: other.to_string(),
        },
    }
}

/// Presupuesto de exploración. Su agotamiento **no** es un fallo: es información que se
/// declara al usuario (FR-001: "declarar qué quedó fuera").
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct IngestBudget {
    max_reads: usize,
    max_fetches: usize,
    max_steps: usize,
    reads_used: usize,
    fetches_used: usize,
    steps_used: usize,
}

impl Default for IngestBudget {
    fn default() -> Self {
        Self::new(40, 10, 60)
    }
}

impl IngestBudget {
    pub fn new(max_reads: usize, max_fetches: usize, max_steps: usize) -> Self {
        Self {
            max_reads,
            max_fetches,
            max_steps,
            reads_used: 0,
            fetches_used: 0,
            steps_used: 0,
        }
    }

    pub fn small() -> Self {
        Self::new(5, 2, 10)
    }

    pub fn try_read(&mut self) -> bool {
        if self.reads_used >= self.max_reads {
            return false;
        }
        self.reads_used += 1;
        true
    }

    pub fn try_fetch(&mut self) -> bool {
        if self.fetches_used >= self.max_fetches {
            return false;
        }
        self.fetches_used += 1;
        true
    }

    pub fn tick_step(&mut self) -> bool {
        if self.steps_used >= self.max_steps {
            return false;
        }
        self.steps_used += 1;
        true
    }

    pub fn exhausted(&self) -> bool {
        self.reads_used >= self.max_reads
            || self.fetches_used >= self.max_fetches
            || self.steps_used >= self.max_steps
    }

    pub fn reads_used(&self) -> usize {
        self.reads_used
    }
    pub fn fetches_used(&self) -> usize {
        self.fetches_used
    }
    pub fn steps_used(&self) -> usize {
        self.steps_used
    }
}

/// Catálogo de herramientas ofrecidas al modelo durante la ingesta.
pub fn ingest_tools(allow_remote: bool) -> Vec<ToolSpec> {
    let mut tools = vec![
        ToolSpec {
            name: TOOL_LIST_REPO.into(),
            description: "Lista las entradas de un directorio del repositorio. '' = raíz.".into(),
            json_schema: r#"{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}"#.into(),
        },
        ToolSpec {
            name: TOOL_READ_FILE.into(),
            description: "Lee un archivo del repositorio por su ruta relativa.".into(),
            json_schema: r#"{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}"#.into(),
        },
        ToolSpec {
            name: TOOL_NOTE_UNRESOLVED.into(),
            description: "Registra algo que NO se pudo verificar. Úsalo en vez de inventar."
                .into(),
            json_schema: r#"{"type":"object","properties":{"what":{"type":"string"},"reason":{"type":"string"}},"required":["what","reason"]}"#.into(),
        },
        ToolSpec {
            name: TOOL_FINALIZE.into(),
            description: "Declara que ya reuniste material suficiente para generar el contexto."
                .into(),
            json_schema: r#"{"type":"object","properties":{"summary":{"type":"string"}},"required":["summary"]}"#.into(),
        },
    ];
    if allow_remote {
        tools.push(ToolSpec {
            name: TOOL_FETCH_URL.into(),
            description: "Trae el contenido de una URL pública referenciada por el repo.".into(),
            json_schema:
                r#"{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}"#
                    .into(),
        });
    }
    tools
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parses_each_tool_call() {
        assert_eq!(
            parse_action(TOOL_READ_FILE, r#"{"path":"README.md"}"#),
            AgentAction::ReadFile {
                path: "README.md".into()
            }
        );
        assert_eq!(
            parse_action(TOOL_LIST_REPO, r#"{"path":""}"#),
            AgentAction::ListRepo {
                path: String::new()
            }
        );
        assert_eq!(
            parse_action(TOOL_FETCH_URL, r#"{"url":"https://x.test/a"}"#),
            AgentAction::FetchUrl {
                url: "https://x.test/a".into()
            }
        );
        assert_eq!(
            parse_action(TOOL_FINALIZE, r#"{"summary":"listo"}"#),
            AgentAction::Finalize {
                summary: "listo".into()
            }
        );
    }

    #[test]
    fn unknown_tool_degrades_instead_of_breaking_the_loop() {
        assert_eq!(
            parse_action("inventada", "{}"),
            AgentAction::Unknown {
                name: "inventada".into()
            }
        );
    }

    #[test]
    fn malformed_arguments_do_not_panic() {
        assert_eq!(
            parse_action(TOOL_READ_FILE, "no-es-json"),
            AgentAction::ReadFile {
                path: String::new()
            }
        );
    }

    #[test]
    fn budget_stops_reads_and_reports_exhaustion() {
        let mut b = IngestBudget::new(2, 1, 10);
        assert!(b.try_read());
        assert!(b.try_read());
        assert!(!b.try_read(), "el tercer read excede el presupuesto");
        assert!(b.exhausted());
        assert_eq!(b.reads_used(), 2);
    }

    #[test]
    fn budget_stops_the_loop_by_steps() {
        let mut b = IngestBudget::new(100, 100, 2);
        assert!(b.tick_step());
        assert!(b.tick_step());
        assert!(!b.tick_step());
    }

    #[test]
    fn fetch_tool_is_absent_when_remote_is_not_allowed() {
        let local = ingest_tools(false);
        assert!(!local.iter().any(|t| t.name == TOOL_FETCH_URL));
        let hybrid = ingest_tools(true);
        assert!(hybrid.iter().any(|t| t.name == TOOL_FETCH_URL));
    }
}
