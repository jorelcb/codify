//! Artefactos de contexto y su *groundedness*.
//!
//! Invariante central de producto (SC-002): lo que no se pudo verificar contra una fuente
//! **nunca** se presenta como hecho. La distinción grounded / tentative / contradiction es
//! parte del dominio, no una convención frágil de texto.

use serde::{Deserialize, Serialize};

/// Los artefactos de contexto que codify-NG produce.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ArtifactKind {
    Agents,
    Context,
    DevelopmentGuide,
    InteractionsLog,
    Idioms,
}

impl ArtifactKind {
    /// Ruta relativa del artefacto dentro del repositorio objetivo.
    pub fn file_path(&self) -> &'static str {
        match self {
            ArtifactKind::Agents => "AGENTS.md",
            ArtifactKind::Context => "context/CONTEXT.md",
            ArtifactKind::DevelopmentGuide => "context/DEVELOPMENT_GUIDE.md",
            ArtifactKind::InteractionsLog => "context/INTERACTIONS_LOG.md",
            ArtifactKind::Idioms => "context/IDIOMS.md",
        }
    }

    /// Conjunto por defecto. `Idioms` se emite solo si hay lenguaje aplicable detectable,
    /// por eso no forma parte del conjunto base.
    pub fn default_set() -> Vec<ArtifactKind> {
        vec![
            ArtifactKind::Agents,
            ArtifactKind::Context,
            ArtifactKind::DevelopmentGuide,
            ArtifactKind::InteractionsLog,
        ]
    }
}

/// Grado de fundamentación de un fragmento de contexto.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum Groundedness {
    /// Verificado contra una o más fuentes leídas. `quotes` son los fragmentos **textuales**
    /// que lo respaldan: el núcleo comprueba que aparecen en el material antes de admitir
    /// este estado (FR-006a). Sin cita comprobable no hay `Grounded`.
    Grounded {
        sources: Vec<String>,
        quotes: Vec<String>,
    },
    /// Inferido o pendiente. `acknowledged` = el usuario lo difirió explícitamente.
    Tentative { reason: String, acknowledged: bool },
    /// Las fuentes se contradicen entre sí: se señala, no se resuelve en silencio (FR-008).
    /// Exige cita comprobable de **cada** fuente en conflicto (FR-006b).
    Contradiction {
        sources: Vec<String>,
        quotes: Vec<String>,
        note: String,
    },
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Segment {
    pub text: String,
    pub groundedness: Groundedness,
}

impl Segment {
    pub fn grounded(text: impl Into<String>, sources: Vec<String>, quotes: Vec<String>) -> Self {
        Self {
            text: text.into(),
            groundedness: Groundedness::Grounded { sources, quotes },
        }
    }

    pub fn tentative(text: impl Into<String>, reason: impl Into<String>) -> Self {
        Self {
            text: text.into(),
            groundedness: Groundedness::Tentative {
                reason: reason.into(),
                acknowledged: false,
            },
        }
    }

    pub fn contradiction(
        text: impl Into<String>,
        sources: Vec<String>,
        quotes: Vec<String>,
        note: impl Into<String>,
    ) -> Self {
        Self {
            text: text.into(),
            groundedness: Groundedness::Contradiction {
                sources,
                quotes,
                note: note.into(),
            },
        }
    }

    pub fn is_grounded(&self) -> bool {
        matches!(self.groundedness, Groundedness::Grounded { .. })
    }

    pub fn is_contradiction(&self) -> bool {
        matches!(self.groundedness, Groundedness::Contradiction { .. })
    }

    /// Tentativo y todavía **sin atender** (ni resuelto ni diferido explícitamente).
    /// Es lo que bloquea el cierre de la sesión (FR-013).
    pub fn is_unattended_tentative(&self) -> bool {
        matches!(
            self.groundedness,
            Groundedness::Tentative {
                acknowledged: false,
                ..
            }
        )
    }

    /// Marca el segmento como diferido de forma explícita por el usuario.
    pub fn acknowledge(&mut self) {
        if let Groundedness::Tentative { acknowledged, .. } = &mut self.groundedness {
            *acknowledged = true;
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ContextArtifact {
    pub kind: ArtifactKind,
    pub segments: Vec<Segment>,
    pub locale: String,
}

impl ContextArtifact {
    pub fn new(kind: ArtifactKind, locale: impl Into<String>) -> Self {
        Self {
            kind,
            segments: Vec::new(),
            locale: locale.into(),
        }
    }

    pub fn with_segments(mut self, segments: Vec<Segment>) -> Self {
        self.segments = segments;
        self
    }

    pub fn unattended_tentative_count(&self) -> usize {
        self.segments
            .iter()
            .filter(|s| s.is_unattended_tentative())
            .count()
    }

    pub fn contradictions(&self) -> Vec<&Segment> {
        self.segments
            .iter()
            .filter(|s| s.is_contradiction())
            .collect()
    }

    /// Renderiza el artefacto a Markdown. Lo tentativo y las contradicciones se marcan de
    /// forma **distinguible** para que ningún lector (humano o agente) los tome por hecho.
    pub fn render(&self) -> String {
        let mut out = String::new();
        for seg in &self.segments {
            match &seg.groundedness {
                Groundedness::Grounded { .. } => out.push_str(&seg.text),
                Groundedness::Tentative { reason, .. } => {
                    out.push_str(&format!("{} <!-- TENTATIVO: {} -->", seg.text, reason));
                }
                Groundedness::Contradiction { sources, note, .. } => {
                    out.push_str(&format!(
                        "{} <!-- CONTRADICCIÓN entre fuentes [{}]: {} -->",
                        seg.text,
                        sources.join(", "),
                        note
                    ));
                }
            }
            out.push('\n');
        }
        out
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tentative_segment_starts_unattended_and_can_be_acknowledged() {
        let mut seg = Segment::tentative("stack por definir", "ninguna fuente lo cubre");
        assert!(seg.is_unattended_tentative());
        seg.acknowledge();
        assert!(!seg.is_unattended_tentative());
    }

    #[test]
    fn render_marks_non_grounded_content_distinguishably() {
        let art = ContextArtifact::new(ArtifactKind::Context, "es").with_segments(vec![
            Segment::grounded(
                "Motor: Temporal",
                vec!["SPEC-30".into()],
                vec!["el motor es Temporal".into()],
            ),
            Segment::tentative("Observabilidad: por definir", "sin fuente"),
        ]);
        let rendered = art.render();
        assert!(rendered.contains("Motor: Temporal"));
        assert!(rendered.contains("TENTATIVO"));
        // El contenido grounded no se ensucia con marcas.
        assert!(!rendered.starts_with("<!--"));
    }

    #[test]
    fn artifact_counts_unattended_tentative_segments() {
        let art = ContextArtifact::new(ArtifactKind::Agents, "en").with_segments(vec![
            Segment::grounded("a", vec![], vec![]),
            Segment::tentative("b", "r"),
            Segment::tentative("c", "r"),
        ]);
        assert_eq!(art.unattended_tentative_count(), 2);
    }
}
