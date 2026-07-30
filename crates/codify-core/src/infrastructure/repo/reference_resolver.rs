//! Adapter de `ReferenceResolver`: rutas locales del repo + URLs **públicas** (FR-003).
//!
//! Dos garantías de producto materializadas aquí:
//! 1. **Nunca se fabrica contenido** (SC-006): lo no resuelto vuelve como `Reference`
//!    sin contenido, con su estado (`Inaccessible` / `RequiresAuth` / `OutOfScope`).
//! 2. **Cero-egress** (SC-007): construido con `local_only()`, el resolver **no tiene**
//!    cliente HTTP; una URL remota se reporta `OutOfScope` en vez de salir a la red.

use crate::application::ports::ReferenceResolver;
use crate::domain::reference::{Reference, ReferenceOrigin, ReferenceState};
use async_trait::async_trait;
use std::path::{Component, Path, PathBuf};
use std::time::Duration;

pub struct FsHttpReferenceResolver {
    root: PathBuf,
    /// `None` ⇒ modo local: no existe cliente HTTP en el grafo.
    http: Option<reqwest::Client>,
}

impl FsHttpReferenceResolver {
    /// Resolver de modo local: solo referencias in-repo. Sin cliente HTTP.
    pub fn local_only(root: impl AsRef<Path>) -> Self {
        Self {
            root: root.as_ref().to_path_buf(),
            http: None,
        }
    }

    /// Resolver híbrido: además de las locales, sigue URLs públicas (sin autenticación).
    pub fn with_public_web(root: impl AsRef<Path>) -> Self {
        let http = reqwest::Client::builder()
            .timeout(Duration::from_secs(15))
            .user_agent("codify-ng/0.1")
            .build()
            .ok();
        Self {
            root: root.as_ref().to_path_buf(),
            http,
        }
    }

    pub fn follows_remote(&self) -> bool {
        self.http.is_some()
    }

    fn safe_join(&self, rel: &str) -> Option<PathBuf> {
        let candidate = Path::new(rel);
        if candidate.is_absolute()
            || candidate
                .components()
                .any(|c| matches!(c, Component::ParentDir))
        {
            return None;
        }
        Some(self.root.join(candidate))
    }

    async fn resolve_local(&self, origin: &ReferenceOrigin, rel: &str) -> Reference {
        let Some(path) = self.safe_join(rel) else {
            return Reference::unresolved(origin.clone(), ReferenceState::OutOfScope);
        };
        match tokio::fs::read(&path).await {
            Ok(bytes) if !bytes.contains(&0) => {
                Reference::resolved(origin.clone(), String::from_utf8_lossy(&bytes).to_string())
            }
            Ok(_) => Reference::unresolved(origin.clone(), ReferenceState::OutOfScope),
            Err(_) => Reference::unresolved(origin.clone(), ReferenceState::Inaccessible),
        }
    }

    async fn resolve_remote(&self, origin: &ReferenceOrigin, url: &str) -> Reference {
        // Modo local: el cliente HTTP no existe. La referencia se declara fuera de alcance.
        let Some(client) = &self.http else {
            return Reference::unresolved(origin.clone(), ReferenceState::OutOfScope);
        };

        match client.get(url).send().await {
            Ok(resp) => {
                let status = resp.status();
                if status == reqwest::StatusCode::UNAUTHORIZED
                    || status == reqwest::StatusCode::FORBIDDEN
                {
                    // Requiere autenticación: fuera de v1. Se reporta, no se inventa.
                    return Reference::unresolved(origin.clone(), ReferenceState::RequiresAuth);
                }
                if !status.is_success() {
                    return Reference::unresolved(origin.clone(), ReferenceState::Inaccessible);
                }
                match resp.text().await {
                    Ok(body) => Reference::resolved(origin.clone(), body),
                    Err(_) => Reference::unresolved(origin.clone(), ReferenceState::Inaccessible),
                }
            }
            Err(_) => Reference::unresolved(origin.clone(), ReferenceState::Inaccessible),
        }
    }
}

#[async_trait]
impl ReferenceResolver for FsHttpReferenceResolver {
    async fn resolve(&self, origin: &ReferenceOrigin) -> Reference {
        match origin {
            ReferenceOrigin::LocalPath(p) => {
                let rel = p.clone();
                self.resolve_local(origin, &rel).await
            }
            ReferenceOrigin::PublicUrl(u) => {
                let url = u.clone();
                self.resolve_remote(origin, &url).await
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn local_only_resolver_has_no_http_client() {
        let r = FsHttpReferenceResolver::local_only("/tmp");
        assert!(
            !r.follows_remote(),
            "en modo local no debe existir cliente HTTP"
        );
    }

    #[tokio::test]
    async fn local_mode_reports_remote_reference_as_out_of_scope_without_network() {
        let r = FsHttpReferenceResolver::local_only("/tmp");
        let reference = r
            .resolve(&ReferenceOrigin::PublicUrl(
                "https://example.test/doc.md".into(),
            ))
            .await;
        assert_eq!(reference.state(), ReferenceState::OutOfScope);
        assert_eq!(reference.content(), None, "jamás se fabrica contenido");
    }

    #[tokio::test]
    async fn path_traversal_is_out_of_scope() {
        let r = FsHttpReferenceResolver::local_only("/tmp");
        let reference = r
            .resolve(&ReferenceOrigin::LocalPath("../../etc/passwd".into()))
            .await;
        assert_eq!(reference.state(), ReferenceState::OutOfScope);
        assert_eq!(reference.content(), None);
    }

    #[tokio::test]
    async fn missing_local_file_is_inaccessible_not_fabricated() {
        let r = FsHttpReferenceResolver::local_only("/tmp");
        let reference = r
            .resolve(&ReferenceOrigin::LocalPath(
                "no-existe-jamas-12345.md".into(),
            ))
            .await;
        assert_eq!(reference.state(), ReferenceState::Inaccessible);
        assert_eq!(reference.content(), None);
    }
}
