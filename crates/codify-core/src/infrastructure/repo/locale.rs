//! Adapter de `LocaleDetector` (FR-019).
//!
//! Heurística deliberadamente simple y **sin LLM**: cuenta palabras funcionales del
//! contenido ya leído del repo. El usuario siempre puede sobrescribir el resultado.

use crate::application::ports::LocaleDetector;
use crate::domain::reference::Repository;
use async_trait::async_trait;

const ES_MARKERS: &[&str] = &[
    " el ", " la ", " los ", " las ", " de ", " que ", " para ", " con ", " una ", " del ",
    " por ", " como ", " este ", " esta ",
];
const EN_MARKERS: &[&str] = &[
    " the ", " of ", " and ", " to ", " for ", " with ", " this ", " that ", " is ", " are ",
    " from ", " it ",
];

/// Detecta `es` / `en` a partir de un corpus de texto del repositorio.
pub struct HeuristicLocaleDetector {
    corpus: String,
    fallback: String,
}

impl HeuristicLocaleDetector {
    pub fn new(corpus: impl Into<String>) -> Self {
        Self {
            corpus: corpus.into(),
            fallback: "en".into(),
        }
    }

    pub fn with_fallback(mut self, locale: impl Into<String>) -> Self {
        self.fallback = locale.into();
        self
    }

    /// Función pura: expuesta para poder testear la heurística sin construir el adapter.
    pub fn detect_in(text: &str, fallback: &str) -> String {
        let lower = format!(" {} ", text.to_lowercase().replace(['\n', '\t'], " "));
        let count =
            |markers: &[&str]| -> usize { markers.iter().map(|m| lower.matches(m).count()).sum() };
        let es = count(ES_MARKERS);
        let en = count(EN_MARKERS);
        if es == 0 && en == 0 {
            return fallback.to_string();
        }
        if es > en {
            "es".to_string()
        } else {
            "en".to_string()
        }
    }
}

#[async_trait]
impl LocaleDetector for HeuristicLocaleDetector {
    async fn detect(&self, _repo: &Repository) -> String {
        Self::detect_in(&self.corpus, &self.fallback)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn detects_spanish_dominant_corpus() {
        let text = "Este es el servicio que ejecuta los Runs de larga duración del pipeline, \
                    con workers de Python para la ejecución";
        assert_eq!(HeuristicLocaleDetector::detect_in(text, "en"), "es");
    }

    #[test]
    fn detects_english_dominant_corpus() {
        let text = "This is the service that runs the long running jobs of the pipeline, \
                    with workers for the execution";
        assert_eq!(HeuristicLocaleDetector::detect_in(text, "es"), "en");
    }

    #[test]
    fn falls_back_when_no_signal() {
        assert_eq!(HeuristicLocaleDetector::detect_in("", "en"), "en");
        assert_eq!(HeuristicLocaleDetector::detect_in("x1 y2 z3", "es"), "es");
    }
}
