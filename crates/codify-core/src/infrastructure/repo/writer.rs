//! Adapter de `ArtifactWriter` sobre el sistema de archivos.
//!
//! Es el que cierra la deuda de que el núcleo generaba contexto y nunca lo entregaba.
//!
//! Dos reglas de producto viven aquí:
//! 1. **El agente no se sale del repositorio**: mismas defensas de ruta que el navegador.
//! 2. **Un fallo se registra, no explota**: `write` devuelve siempre un `WriteRecord`, para
//!    que un archivo problemático no arrastre a los otros tres.

use crate::application::ports::ArtifactWriter;
use crate::domain::error::{CoreError, Result};
use crate::domain::write::WriteRecord;
use async_trait::async_trait;
use std::path::{Component, Path, PathBuf};

pub struct FsArtifactWriter {
    root: PathBuf,
}

impl FsArtifactWriter {
    pub fn new(root: impl AsRef<Path>) -> Self {
        Self {
            root: root.as_ref().to_path_buf(),
        }
    }

    /// Resuelve una ruta relativa **dentro** del repositorio. Rechaza absolutas y escapes.
    fn safe_join(&self, rel: &str) -> Result<PathBuf> {
        let candidate = Path::new(rel);
        if candidate.is_absolute() {
            return Err(CoreError::Invalid(format!(
                "ruta absoluta no permitida: {rel}"
            )));
        }
        if candidate
            .components()
            .any(|c| matches!(c, Component::ParentDir))
        {
            return Err(CoreError::Invalid(format!(
                "ruta fuera del repositorio: {rel}"
            )));
        }
        Ok(self.root.join(candidate))
    }

    /// Marca temporal sin dependencias de fecha: suficiente para ordenar y auditar.
    fn now() -> String {
        let secs = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0);
        format!("epoch:{secs}")
    }
}

#[async_trait]
impl ArtifactWriter for FsArtifactWriter {
    async fn write(&self, path: &str, content: &str) -> WriteRecord {
        let at = Self::now();

        let target = match self.safe_join(path) {
            Ok(t) => t,
            Err(e) => return WriteRecord::failed(path, at, e.to_string()),
        };

        // `context/` puede no existir todavía.
        if let Some(parent) = target.parent() {
            if let Err(e) = tokio::fs::create_dir_all(parent).await {
                return WriteRecord::failed(
                    path,
                    at,
                    format!("no se pudo crear el directorio: {e}"),
                );
            }
        }

        match tokio::fs::write(&target, content).await {
            Ok(()) => WriteRecord::written(path, content.len(), at),
            Err(e) => WriteRecord::failed(path, at, e.to_string()),
        }
    }

    async fn read_existing(&self, path: &str) -> Result<Option<String>> {
        let target = self.safe_join(path)?;
        match tokio::fs::read_to_string(&target).await {
            Ok(content) => Ok(Some(content)),
            // No existir no es un error: es la respuesta.
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(e) => Err(CoreError::Storage(e.to_string())),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_paths_that_escape_the_repository() {
        let w = FsArtifactWriter::new("/tmp/repo");
        assert!(w.safe_join("../fuera.md").is_err());
        assert!(w.safe_join("/etc/passwd").is_err());
        assert!(w.safe_join("context/../../fuera.md").is_err());
        assert!(w.safe_join("context/CONTEXT.md").is_ok());
    }

    #[tokio::test]
    async fn reading_a_missing_file_is_none_not_an_error() {
        let w = FsArtifactWriter::new(std::env::temp_dir());
        assert_eq!(
            w.read_existing("no-existe-jamas-98765.md").await.unwrap(),
            None
        );
    }

    #[tokio::test]
    async fn a_rejected_path_returns_a_failed_record_instead_of_panicking() {
        let w = FsArtifactWriter::new(std::env::temp_dir());
        let record = w.write("../fuera.md", "x").await;
        assert!(!record.reached_disk());
        assert!(record.summary().contains("falló"));
    }
}
