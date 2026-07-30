//! **Composition root.** Único lugar donde se ensambla el grafo de objetos concretos.
//!
//! Aquí se materializa la garantía de **cero-egress** (restricción de proyecto,
//! [NON-NEGOTIABLE]): el builder delega en `ProviderRegistry::for_mode`, que en modo `Local`
//! rechaza cualquier proveedor no local. El adapter remoto no existe en el grafo — no es un
//! flag de runtime que un bug pueda saltarse.
//!
//! La *definición* de las dependencias vive en `application::deps` (es el orquestador quien
//! las nombra); aquí solo se cablean los concretos.

use crate::application::deps::{AuthoringDeps, ProviderRegistry};
use crate::application::ports::{
    AuditSink, DiffEngine, LocaleDetector, ModelProvider, Prompter, ReferenceResolver,
    RepoNavigator,
};
use crate::domain::error::{CoreError, Result};
use crate::domain::ports::{Clock, RiskClassifier};
use crate::domain::session::Mode;
use std::sync::Arc;

/// Builder del composition root. Las pieles (Tauri, MCP, CLI) y los tests inyectan aquí sus
/// adapters concretos; el núcleo nunca los construye por su cuenta.
pub struct CoreBuilder {
    mode: Mode,
    providers: Vec<Arc<dyn ModelProvider>>,
    navigator: Option<Arc<dyn RepoNavigator>>,
    resolver: Option<Arc<dyn ReferenceResolver>>,
    diff: Option<Arc<dyn DiffEngine>>,
    risk: Option<Arc<dyn RiskClassifier>>,
    prompter: Option<Arc<dyn Prompter>>,
    audit: Option<Arc<dyn AuditSink>>,
    locale: Option<Arc<dyn LocaleDetector>>,
    clock: Option<Arc<dyn Clock>>,
}

impl CoreBuilder {
    pub fn new(mode: Mode) -> Self {
        Self {
            mode,
            providers: Vec::new(),
            navigator: None,
            resolver: None,
            diff: None,
            risk: None,
            prompter: None,
            audit: None,
            locale: None,
            clock: None,
        }
    }

    pub fn provider(mut self, p: Arc<dyn ModelProvider>) -> Self {
        self.providers.push(p);
        self
    }
    pub fn navigator(mut self, v: Arc<dyn RepoNavigator>) -> Self {
        self.navigator = Some(v);
        self
    }
    pub fn resolver(mut self, v: Arc<dyn ReferenceResolver>) -> Self {
        self.resolver = Some(v);
        self
    }
    pub fn diff(mut self, v: Arc<dyn DiffEngine>) -> Self {
        self.diff = Some(v);
        self
    }
    pub fn risk(mut self, v: Arc<dyn RiskClassifier>) -> Self {
        self.risk = Some(v);
        self
    }
    pub fn prompter(mut self, v: Arc<dyn Prompter>) -> Self {
        self.prompter = Some(v);
        self
    }
    pub fn audit(mut self, v: Arc<dyn AuditSink>) -> Self {
        self.audit = Some(v);
        self
    }
    pub fn locale(mut self, v: Arc<dyn LocaleDetector>) -> Self {
        self.locale = Some(v);
        self
    }
    pub fn clock(mut self, v: Arc<dyn Clock>) -> Self {
        self.clock = Some(v);
        self
    }

    pub fn build(self) -> Result<AuthoringDeps> {
        let missing = |what: &str| CoreError::Invalid(format!("falta cablear el port '{what}'"));
        let providers = Arc::new(ProviderRegistry::for_mode(self.mode, self.providers)?);
        Ok(AuthoringDeps {
            providers,
            navigator: self.navigator.ok_or_else(|| missing("RepoNavigator"))?,
            resolver: self.resolver.ok_or_else(|| missing("ReferenceResolver"))?,
            diff: self.diff.ok_or_else(|| missing("DiffEngine"))?,
            risk: self.risk.ok_or_else(|| missing("RiskClassifier"))?,
            prompter: self.prompter.ok_or_else(|| missing("Prompter"))?,
            audit: self.audit.ok_or_else(|| missing("AuditSink"))?,
            locale: self.locale.ok_or_else(|| missing("LocaleDetector"))?,
            clock: self.clock.ok_or_else(|| missing("Clock"))?,
            mode: self.mode,
        })
    }
}
