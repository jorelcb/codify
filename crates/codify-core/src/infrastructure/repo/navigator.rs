//! Adapter de `RepoNavigator` sobre el sistema de archivos.
//!
//! Aplica el **muestreo acotado** de la ingesta (D4): recorta archivos grandes marcando
//! `truncated`, e ignora directorios de build/VCS. Nunca trunca en silencio: el flag viaja
//! al loop, que lo declara.

use crate::application::ports::{Entry, EntryKind, FileContent, RepoNavigator};
use crate::domain::error::{CoreError, Result};
use crate::domain::reference::Repository;
use async_trait::async_trait;
use std::path::{Component, Path, PathBuf};

const DEFAULT_MAX_FILE_BYTES: usize = 64 * 1024;

const IGNORED_DIRS: &[&str] = &[
    ".git",
    "target",
    "node_modules",
    "vendor",
    "dist",
    "build",
    ".venv",
    "__pycache__",
];

/// Manifiestos y señales estructurales que describen el proyecto sin leerlo entero.
const STRUCTURAL_FILES: &[&str] = &[
    "Cargo.toml",
    "package.json",
    "go.mod",
    "pyproject.toml",
    "requirements.txt",
    "pom.xml",
    "build.gradle",
    "Gemfile",
    "composer.json",
    "Dockerfile",
];

pub struct FsRepoNavigator {
    root: PathBuf,
    max_file_bytes: usize,
}

impl FsRepoNavigator {
    pub fn new(root: impl AsRef<Path>) -> Self {
        Self {
            root: root.as_ref().to_path_buf(),
            max_file_bytes: DEFAULT_MAX_FILE_BYTES,
        }
    }

    pub fn with_max_file_bytes(mut self, bytes: usize) -> Self {
        self.max_file_bytes = bytes;
        self
    }

    /// Resuelve una ruta relativa **dentro** del repo. Rechaza escapes (`..`, absolutas):
    /// el agente no puede salirse del repositorio objetivo.
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

    fn is_ignored(path: &Path) -> bool {
        path.components().any(|c| {
            c.as_os_str()
                .to_str()
                .is_some_and(|name| IGNORED_DIRS.contains(&name))
        })
    }
}

#[async_trait]
impl RepoNavigator for FsRepoNavigator {
    async fn list(&self, path: &str) -> Result<Vec<Entry>> {
        let dir = self.safe_join(path)?;
        let mut entries = Vec::new();
        let mut read_dir = tokio::fs::read_dir(&dir)
            .await
            .map_err(|e| CoreError::NotFound(format!("{}: {e}", dir.display())))?;

        while let Some(entry) = read_dir
            .next_entry()
            .await
            .map_err(|e| CoreError::Storage(e.to_string()))?
        {
            let abs = entry.path();
            if Self::is_ignored(&abs) {
                continue;
            }
            let meta = match entry.metadata().await {
                Ok(m) => m,
                Err(_) => continue,
            };
            let rel = abs
                .strip_prefix(&self.root)
                .unwrap_or(&abs)
                .to_string_lossy()
                .to_string();
            entries.push(Entry {
                path: rel,
                kind: if meta.is_dir() {
                    EntryKind::Dir
                } else {
                    EntryKind::File
                },
                size: meta.len(),
            });
        }
        entries.sort_by(|a, b| a.path.cmp(&b.path));
        Ok(entries)
    }

    async fn read(&self, path: &str) -> Result<FileContent> {
        let file = self.safe_join(path)?;
        let bytes = tokio::fs::read(&file)
            .await
            .map_err(|e| CoreError::NotFound(format!("{path}: {e}")))?;

        // Binario: no forma parte del material de grounding.
        if bytes.contains(&0) {
            return Err(CoreError::Invalid(format!(
                "archivo binario omitido: {path}"
            )));
        }

        let full = String::from_utf8_lossy(&bytes).to_string();
        let truncated = full.len() > self.max_file_bytes;
        let content = if truncated {
            full.chars().take(self.max_file_bytes).collect::<String>()
        } else {
            full
        };
        Ok(FileContent {
            path: path.to_string(),
            content,
            truncated,
        })
    }

    async fn describe(&self) -> Result<Repository> {
        let mut repo = Repository::new(&self.root);
        let top = self.list("").await.unwrap_or_default();
        repo.is_empty = top.is_empty();
        repo.structural_signals = top
            .iter()
            .filter(|e| {
                matches!(e.kind, EntryKind::Dir)
                    || STRUCTURAL_FILES.iter().any(|s| e.path.ends_with(s))
            })
            .map(|e| e.path.clone())
            .collect();
        repo.detected_language = detect_language(&top);
        Ok(repo)
    }
}

fn detect_language(entries: &[Entry]) -> Option<String> {
    for entry in entries {
        let lang = match entry.path.as_str() {
            p if p.ends_with("Cargo.toml") => Some("rust"),
            p if p.ends_with("go.mod") => Some("go"),
            p if p.ends_with("package.json") => Some("typescript"),
            p if p.ends_with("pyproject.toml") || p.ends_with("requirements.txt") => Some("python"),
            p if p.ends_with("pom.xml") || p.ends_with("build.gradle") => Some("java"),
            _ => None,
        };
        if let Some(l) = lang {
            return Some(l.to_string());
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn rejects_path_traversal_and_absolute_paths() {
        let nav = FsRepoNavigator::new("/tmp");
        assert!(nav.read("../etc/passwd").await.is_err());
        assert!(nav.read("/etc/passwd").await.is_err());
    }

    #[test]
    fn detects_language_from_manifest() {
        let entries = vec![Entry {
            path: "Cargo.toml".into(),
            kind: EntryKind::File,
            size: 10,
        }];
        assert_eq!(detect_language(&entries), Some("rust".into()));
    }

    #[test]
    fn ignores_build_and_vcs_directories() {
        assert!(FsRepoNavigator::is_ignored(Path::new(
            "/repo/target/debug/x"
        )));
        assert!(FsRepoNavigator::is_ignored(Path::new("/repo/.git/config")));
        assert!(!FsRepoNavigator::is_ignored(Path::new("/repo/src/lib.rs")));
    }
}
