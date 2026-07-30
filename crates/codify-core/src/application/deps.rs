//! Dependencias del caso de uso: el conjunto de ports que el loop necesita, más la política
//! de selección de proveedor.
//!
//! Vive en **Application** (no en Infrastructure) porque es el orquestador quien las nombra:
//! si viviera afuera, la Aplicación tendría que mirar hacia la Infraestructura y se rompería
//! la Regla de Dependencia. El *ensamblado* de los concretos sigue siendo trabajo del
//! composition root.

use crate::application::ports::{
    AuditSink, DiffEngine, LocaleDetector, ModelProvider, Prompter, ReferenceResolver,
    RepoNavigator, Tier,
};
use crate::domain::error::{CoreError, Result};
use crate::domain::ports::{Clock, RiskClassifier};
use crate::domain::session::Mode;
use std::sync::Arc;

/// Conjunto de proveedores elegibles para una sesión, validado contra su modo.
///
/// **Cero-egress estructural** (SC-007): en `Mode::Local` el registro no admite un proveedor
/// cuyo `is_local()` sea falso. No es un flag que se consulte al enviar la petición: el
/// adapter remoto sencillamente no llega a existir en el grafo.
pub struct ProviderRegistry {
    providers: Vec<Arc<dyn ModelProvider>>,
    mode: Mode,
}

impl ProviderRegistry {
    pub fn for_mode(mode: Mode, providers: Vec<Arc<dyn ModelProvider>>) -> Result<Self> {
        if providers.is_empty() {
            return Err(CoreError::Invalid(
                "se requiere al menos un proveedor de modelo".into(),
            ));
        }
        if mode.is_local() {
            if let Some(remote) = providers.iter().find(|p| !p.is_local()) {
                return Err(CoreError::EgressBlocked(format!(
                    "proveedor no local '{}' rechazado en modo Local",
                    remote.name()
                )));
            }
        }
        Ok(Self { providers, mode })
    }

    pub fn mode(&self) -> Mode {
        self.mode
    }

    pub fn all(&self) -> &[Arc<dyn ModelProvider>] {
        &self.providers
    }

    /// Enruta por tier. Sin proveedor del tier pedido, degrada al primero disponible; la
    /// degradación se declara al usuario en la capa de aplicación (FR-018).
    pub fn pick(&self, tier: Tier) -> Arc<dyn ModelProvider> {
        self.providers
            .iter()
            .find(|p| p.tier_hint() == tier)
            .cloned()
            .unwrap_or_else(|| self.providers[0].clone())
    }

    /// `true` si **todo** el registro es local. Invariante verificada en CI.
    pub fn is_fully_local(&self) -> bool {
        self.providers.iter().all(|p| p.is_local())
    }
}

/// Ports inyectados en el loop de authoring (constructor injection, D5).
#[derive(Clone)]
pub struct AuthoringDeps {
    pub providers: Arc<ProviderRegistry>,
    pub navigator: Arc<dyn RepoNavigator>,
    pub resolver: Arc<dyn ReferenceResolver>,
    pub diff: Arc<dyn DiffEngine>,
    pub risk: Arc<dyn RiskClassifier>,
    pub prompter: Arc<dyn Prompter>,
    pub audit: Arc<dyn AuditSink>,
    pub locale: Arc<dyn LocaleDetector>,
    pub clock: Arc<dyn Clock>,
    pub mode: Mode,
}
