//! **Composition root.** Único lugar donde se ensambla el grafo de objetos concretos.
//!
//! Aquí se materializa la garantía de **cero-egress** (restricción de proyecto,
//! [NON-NEGOTIABLE]), y desde `003`-FR-008 lo hace en **dos niveles**:
//!
//! 1. **En el tipo**: `CoreBuilder<Local>` no tiene el método que acepta un proveedor remoto.
//!    No es que lo rechace — no hay método al que llamar, y escribirlo es un error de
//!    compilación. Lo comprueba `tests/compile_fail/`, que existe para no compilar.
//! 2. **En tiempo de ejecución**: `ProviderRegistry::for_mode` sigue rechazando un proveedor no
//!    local en modo `Local`. Se mantiene como defensa en profundidad: cubre un proveedor que
//!    llegara por otra vía, y no sobra por estar el primero.
//!
//! El nivel 1 es el que sostiene la palabra «estructuralmente». Un rechazo en runtime dice «no
//! lo hace»; un método que no existe dice «no puede».
//!
//! La *definición* de las dependencias vive en `application::deps` (es el orquestador quien
//! las nombra); aquí solo se cablean los concretos.

use crate::application::deps::{AuthoringDeps, ProviderRegistry};
use crate::application::ports::{
    ArtifactWriter, AuditSink, CancellationFactory, DiffEngine, LocaleDetector, ModelProvider,
    Prompter, ProviderDiscovery, ReferenceResolver, RepoNavigator,
};
use crate::domain::error::{CoreError, Result};
use crate::domain::ports::{Clock, RiskClassifier};
use crate::domain::session::Mode;
use std::sync::Arc;

/// El modo, como parámetro de tipo (`003`-FR-008a).
///
/// Existe para que la diferencia entre un grafo que puede salir a la red y uno que no **sea de
/// tipos**, no de datos. Un `bool` o un `Mode` en un campo se comprueban; un parámetro de tipo
/// decide qué métodos existen.
pub trait ModoDelGrafo: private::Sellado {
    /// El `Mode` que corresponde a este estado.
    ///
    /// Existe para que el modo tenga **una sola fuente**. Si `new()` recibiera un `Mode` además
    /// del parámetro de tipo, `CoreBuilder::<Local>::new(Mode::Hybrid)` compilaría y el grafo
    /// diría una cosa mientras el tipo dice otra — que es precisamente el desacuerdo que este
    /// diseño viene a hacer imposible.
    const MODE: Mode;
}

/// Grafo sin salida a la red. **No admite proveedores remotos** — el método no existe.
pub struct Local;
/// Grafo que puede usar proveedores remotos, con el consentimiento explícito del usuario.
pub struct Hybrid;

impl ModoDelGrafo for Local {
    const MODE: Mode = Mode::Local;
}
impl ModoDelGrafo for Hybrid {
    const MODE: Mode = Mode::Hybrid;
}

mod private {
    /// Impide que alguien fuera de este módulo declare un tercer modo y se salte la distinción.
    pub trait Sellado {}
    impl Sellado for super::Local {}
    impl Sellado for super::Hybrid {}
}

/// Builder del composition root. Las pieles (Tauri, MCP, CLI) y los tests inyectan aquí sus
/// adapters concretos; el núcleo nunca los construye por su cuenta.
///
/// El parámetro `M` decide **qué se puede cablear**: ver `Local` y `Hybrid`.
pub struct CoreBuilder<M: ModoDelGrafo = Local> {
    _modo: std::marker::PhantomData<M>,
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
    writer: Option<Arc<dyn ArtifactWriter>>,
    discovery: Option<Arc<dyn ProviderDiscovery>>,
    cancellations: Option<Arc<dyn CancellationFactory>>,
}

impl<M: ModoDelGrafo> Default for CoreBuilder<M> {
    fn default() -> Self {
        Self::new()
    }
}

impl<M: ModoDelGrafo> CoreBuilder<M> {
    /// El modo **no se pasa**: lo dice el parámetro de tipo. Ver `ModoDelGrafo::MODE`.
    pub fn new() -> Self {
        Self {
            _modo: std::marker::PhantomData,
            mode: M::MODE,
            providers: Vec::new(),
            navigator: None,
            resolver: None,
            diff: None,
            risk: None,
            prompter: None,
            audit: None,
            locale: None,
            clock: None,
            writer: None,
            discovery: None,
            cancellations: None,
        }
    }

    /// Cablea un proveedor **local**. Disponible en los dos modos.
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
    pub fn writer(mut self, v: Arc<dyn ArtifactWriter>) -> Self {
        self.writer = Some(v);
        self
    }
    pub fn discovery(mut self, v: Arc<dyn ProviderDiscovery>) -> Self {
        self.discovery = Some(v);
        self
    }
    pub fn cancellations(mut self, v: Arc<dyn CancellationFactory>) -> Self {
        self.cancellations = Some(v);
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
            writer: self.writer.ok_or_else(|| missing("ArtifactWriter"))?,
            discovery: self.discovery.ok_or_else(|| missing("ProviderDiscovery"))?,
            cancellations: self
                .cancellations
                .ok_or_else(|| missing("CancellationFactory"))?,
            mode: self.mode,
        })
    }
}

impl CoreBuilder<Hybrid> {
    /// Cablea un proveedor **remoto**, capaz de salir a la red.
    ///
    /// Vive solo en este `impl`, y ahí está toda la garantía de `003`-FR-008: en un
    /// `CoreBuilder<Local>` este método **no existe**, así que el programa que lo llamara no
    /// compila. Lo comprueba `tests/compile_fail/local_no_admite_remoto.rs`.
    pub fn remote_provider(mut self, p: Arc<dyn ModelProvider>) -> Self {
        self.providers.push(p);
        self
    }
}
