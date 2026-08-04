//! Adapter de `Cancellation` sobre `tokio_util::sync::CancellationToken`.
//!
//! El token concreto vive **aquí y solo aquí**: el núcleo depende del trait, no de
//! `tokio-util` (constitución, Principio I — ningún tipo de terceros cruza un port).

use crate::application::ports::{Cancellation, CancellationFactory};
use async_trait::async_trait;
use tokio_util::sync::CancellationToken;

#[derive(Debug, Default, Clone)]
pub struct TokenCancellation {
    token: CancellationToken,
}

impl TokenCancellation {
    pub fn new() -> Self {
        Self::default()
    }

    /// Token hijo, para acotar la cancelación a una sub-tarea sin afectar al resto.
    pub fn child(&self) -> Self {
        Self {
            token: self.token.child_token(),
        }
    }
}

#[async_trait]
impl Cancellation for TokenCancellation {
    fn is_cancelled(&self) -> bool {
        self.token.is_cancelled()
    }

    async fn cancelled(&self) {
        self.token.cancelled().await
    }

    /// Despierta a todos los que esperan `cancelled()`.
    fn cancel(&self) {
        self.token.cancel();
    }
}

/// Factoría del composition root: una señal nueva por sesión.
#[derive(Debug, Default)]
pub struct TokenCancellationFactory;

impl CancellationFactory for TokenCancellationFactory {
    fn create(&self) -> std::sync::Arc<dyn Cancellation> {
        std::sync::Arc::new(TokenCancellation::new())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn a_child_token_is_cancelled_by_its_parent() {
        let parent = TokenCancellation::new();
        let child = parent.child();

        assert!(!child.is_cancelled());
        parent.cancel();
        assert!(child.is_cancelled(), "cancelar el padre alcanza al hijo");
    }

    #[tokio::test]
    async fn cancelling_a_child_leaves_the_parent_alone() {
        let parent = TokenCancellation::new();
        let child = parent.child();

        child.cancel();
        assert!(child.is_cancelled());
        assert!(!parent.is_cancelled(), "el hijo no arrastra al padre");
    }
}
