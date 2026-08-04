//! Fakes in-memory de todos los driven ports.
//!
//! Permiten ejercitar el núcleo **sin I/O, sin red y sin proveedor real**. Si un test de
//! caso de uso necesitara Docker, red o disco real, el DIP estaría roto (constitución,
//! Principio I: "señal de test").

#![allow(dead_code)]

use codify_core::application::ports::{
    AuditSink, CompletionOutput, CompletionRequest, DiffEngine, Entry, EntryKind, FileContent,
    LocaleDetector, ModelProvider, Prompter, Question, ReferenceResolver, RepoNavigator, Tier,
};
use codify_core::domain::audit::AuditEvent;
use codify_core::domain::change::{
    ApprovalDecision, ChangeProposal, Diff, ProposalId, RiskLevel, Verdict,
};
use codify_core::domain::error::{CoreError, Result};
use codify_core::domain::ports::{Clock, RiskClassifier};
use codify_core::domain::reference::{Reference, ReferenceOrigin, ReferenceState, Repository};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// ---------------------------------------------------------------------------
// ModelProvider
// ---------------------------------------------------------------------------

pub struct FakeModelProvider {
    name: String,
    local: bool,
    tier: Tier,
    /// Retardo por llamada: permite probar que cancelar **aborta la petición en vuelo**.
    delay: std::time::Duration,
    scripted: Mutex<Vec<CompletionOutput>>,
    pub seen: Mutex<Vec<CompletionRequest>>,
}

impl FakeModelProvider {
    pub fn local(name: &str, scripted: Vec<CompletionOutput>) -> Self {
        Self {
            name: name.into(),
            local: true,
            tier: Tier::Cheap,
            delay: std::time::Duration::ZERO,
            scripted: Mutex::new(scripted),
            seen: Mutex::new(Vec::new()),
        }
    }

    /// Cada llamada tarda `d`. Con esto se puede cancelar "a mitad" de una generación.
    pub fn with_delay(mut self, d: std::time::Duration) -> Self {
        self.delay = d;
        self
    }

    pub fn remote(name: &str) -> Self {
        Self {
            name: name.into(),
            local: false,
            tier: Tier::Heavy,
            delay: std::time::Duration::ZERO,
            scripted: Mutex::new(Vec::new()),
            seen: Mutex::new(Vec::new()),
        }
    }

    pub fn with_tier(mut self, tier: Tier) -> Self {
        self.tier = tier;
        self
    }
}

#[async_trait::async_trait]
impl ModelProvider for FakeModelProvider {
    async fn complete(&self, request: CompletionRequest) -> Result<CompletionOutput> {
        self.seen.lock().unwrap().push(request);
        if !self.delay.is_zero() {
            tokio::time::sleep(self.delay).await;
        }
        let mut q = self.scripted.lock().unwrap();
        if q.is_empty() {
            return Ok(CompletionOutput::Text(String::new()));
        }
        Ok(q.remove(0))
    }
    fn is_local(&self) -> bool {
        self.local
    }
    fn tier_hint(&self) -> Tier {
        self.tier
    }
    fn name(&self) -> &str {
        &self.name
    }
}

// ---------------------------------------------------------------------------
// RepoNavigator
// ---------------------------------------------------------------------------

#[derive(Default)]
pub struct FakeRepoNavigator {
    files: HashMap<String, String>,
    pub reads: Mutex<Vec<String>>,
}

impl FakeRepoNavigator {
    pub fn with_files(files: &[(&str, &str)]) -> Self {
        Self {
            files: files
                .iter()
                .map(|(p, c)| (p.to_string(), c.to_string()))
                .collect(),
            reads: Mutex::new(Vec::new()),
        }
    }
}

#[async_trait::async_trait]
impl RepoNavigator for FakeRepoNavigator {
    async fn list(&self, path: &str) -> Result<Vec<Entry>> {
        let prefix = if path.is_empty() {
            String::new()
        } else {
            format!("{path}/")
        };
        Ok(self
            .files
            .keys()
            .filter(|p| p.starts_with(&prefix))
            .map(|p| Entry {
                path: p.clone(),
                kind: EntryKind::File,
                size: self.files[p].len() as u64,
            })
            .collect())
    }

    async fn read(&self, path: &str) -> Result<FileContent> {
        self.reads.lock().unwrap().push(path.to_string());
        self.files
            .get(path)
            .map(|c| FileContent {
                path: path.into(),
                content: c.clone(),
                truncated: false,
            })
            .ok_or_else(|| CoreError::NotFound(path.into()))
    }

    async fn describe(&self) -> Result<Repository> {
        let mut repo = Repository::new("/fake/repo");
        repo.is_empty = self.files.is_empty();
        repo.structural_signals = self.files.keys().cloned().collect();
        Ok(repo)
    }
}

// ---------------------------------------------------------------------------
// ReferenceResolver
// ---------------------------------------------------------------------------

#[derive(Default)]
pub struct FakeReferenceResolver {
    resolved: HashMap<String, String>,
    unresolved: HashMap<String, ReferenceState>,
}

impl FakeReferenceResolver {
    pub fn new() -> Self {
        Self::default()
    }
    pub fn resolving(mut self, key: &str, content: &str) -> Self {
        self.resolved.insert(key.into(), content.into());
        self
    }
    pub fn failing(mut self, key: &str, state: ReferenceState) -> Self {
        self.unresolved.insert(key.into(), state);
        self
    }
}

#[async_trait::async_trait]
impl ReferenceResolver for FakeReferenceResolver {
    async fn resolve(&self, origin: &ReferenceOrigin) -> Reference {
        let key = origin.as_str().to_string();
        if let Some(content) = self.resolved.get(&key) {
            return Reference::resolved(origin.clone(), content.clone());
        }
        let state = self
            .unresolved
            .get(&key)
            .copied()
            .unwrap_or(ReferenceState::Inaccessible);
        Reference::unresolved(origin.clone(), state)
    }
}

// ---------------------------------------------------------------------------
// DiffEngine / RiskClassifier
// ---------------------------------------------------------------------------

pub struct FakeDiffEngine;

impl DiffEngine for FakeDiffEngine {
    fn make(&self, before: &str, after: &str) -> Diff {
        Diff {
            unified: format!("-{before}\n+{after}"),
            before: before.into(),
            after: after.into(),
        }
    }
    fn apply(&self, _before: &str, diff: &Diff) -> Result<String> {
        Ok(diff.after.clone())
    }
    fn revert(&self, _after: &str, diff: &Diff) -> Result<String> {
        Ok(diff.before.clone())
    }
}

/// Política conservadora de v1: todo lo no trivial es alto impacto.
pub struct ConservativeRisk;

impl RiskClassifier for ConservativeRisk {
    fn classify(&self, proposal: &ChangeProposal) -> RiskLevel {
        if proposal.diff.is_empty() {
            RiskLevel::Low
        } else {
            RiskLevel::HighImpact
        }
    }
}

pub struct AlwaysLowRisk;

impl RiskClassifier for AlwaysLowRisk {
    fn classify(&self, _proposal: &ChangeProposal) -> RiskLevel {
        RiskLevel::Low
    }
}

// ---------------------------------------------------------------------------
// Prompter / AuditSink / LocaleDetector / Clock
// ---------------------------------------------------------------------------

pub struct FakePrompter {
    answers: Mutex<Vec<String>>,
    verdict: Verdict,
    pub asked: Mutex<Vec<Question>>,
    pub presented: Mutex<Vec<ProposalId>>,
}

impl FakePrompter {
    pub fn new(answers: Vec<String>, verdict: Verdict) -> Self {
        Self {
            answers: Mutex::new(answers),
            verdict,
            asked: Mutex::new(Vec::new()),
            presented: Mutex::new(Vec::new()),
        }
    }
    pub fn approving() -> Self {
        Self::new(Vec::new(), Verdict::Approve)
    }
    pub fn rejecting() -> Self {
        Self::new(Vec::new(), Verdict::Reject)
    }
}

#[async_trait::async_trait]
impl Prompter for FakePrompter {
    async fn ask(&self, question: Question) -> Result<String> {
        self.asked.lock().unwrap().push(question);
        let mut a = self.answers.lock().unwrap();
        if a.is_empty() {
            return Ok(String::new());
        }
        Ok(a.remove(0))
    }

    async fn present(&self, proposal: &ChangeProposal) -> Result<ApprovalDecision> {
        self.presented.lock().unwrap().push(proposal.id.clone());
        Ok(ApprovalDecision {
            proposal_id: proposal.id.clone(),
            verdict: self.verdict.clone(),
            actor: "fake-user".into(),
            at: "2026-07-27T00:00:00Z".into(),
        })
    }
}

#[derive(Default)]
pub struct RecordingAudit {
    pub events: Mutex<Vec<AuditEvent>>,
}

impl AuditSink for RecordingAudit {
    fn record(&self, event: AuditEvent) {
        self.events.lock().unwrap().push(event);
    }
}

pub struct FixedLocale(pub &'static str);

#[async_trait::async_trait]
impl LocaleDetector for FixedLocale {
    async fn detect(&self, _repo: &Repository) -> String {
        self.0.to_string()
    }
}

pub struct FixedClock;

impl Clock for FixedClock {
    fn now_iso(&self) -> String {
        "2026-07-27T00:00:00Z".into()
    }
}

// ---------------------------------------------------------------------------
// Cancelación / escritura / sonda de proveedor (spec 002)
// ---------------------------------------------------------------------------

use codify_core::application::ports::{
    ArtifactWriter, Cancellation, ProviderDiscovery, ProviderStatus,
};
use codify_core::domain::write::WriteRecord;
use std::sync::atomic::{AtomicBool, Ordering};

/// Cancelación en memoria: determinista y sin depender del runtime real.
#[derive(Default)]
pub struct FakeCancellation {
    flag: AtomicBool,
    notify: tokio::sync::Notify,
}

impl FakeCancellation {
    pub fn new() -> Self {
        Self::default()
    }
}

#[async_trait::async_trait]
impl Cancellation for FakeCancellation {
    fn is_cancelled(&self) -> bool {
        self.flag.load(Ordering::SeqCst)
    }

    async fn cancelled(&self) {
        if self.is_cancelled() {
            return;
        }
        self.notify.notified().await;
    }

    fn cancel(&self) {
        self.flag.store(true, Ordering::SeqCst);
        self.notify.notify_waiters();
    }
}

/// Escritor en memoria. Registra el orden de escritura para poder asertarlo.
#[derive(Default)]
pub struct FakeArtifactWriter {
    pub files: Mutex<HashMap<String, String>>,
    pub order: Mutex<Vec<String>>,
    /// Rutas que deben fallar, para probar que un fallo aislado no arrastra al resto.
    failing: Mutex<Vec<String>>,
}

impl FakeArtifactWriter {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn failing_on(self, path: &str) -> Self {
        self.failing.lock().unwrap().push(path.to_string());
        self
    }
}

#[async_trait::async_trait]
impl ArtifactWriter for FakeArtifactWriter {
    async fn write(&self, path: &str, content: &str) -> WriteRecord {
        self.order.lock().unwrap().push(path.to_string());
        if self.failing.lock().unwrap().iter().any(|p| p == path) {
            return WriteRecord::failed(path, "t0", "fallo simulado");
        }
        self.files
            .lock()
            .unwrap()
            .insert(path.to_string(), content.to_string());
        WriteRecord::written(path, content.len(), "t0")
    }

    async fn read_existing(&self, path: &str) -> Result<Option<String>> {
        Ok(self.files.lock().unwrap().get(path).cloned())
    }
}

/// Sonda con respuesta guionizada.
pub struct FakeProviderDiscovery(pub ProviderStatus);

#[async_trait::async_trait]
impl ProviderDiscovery for FakeProviderDiscovery {
    async fn probe(&self) -> ProviderStatus {
        self.0.clone()
    }
}

/// Factoría que entrega una señal nueva por sesión, y **recuerda la última** para poder
/// cancelarla desde el test.
#[derive(Default)]
pub struct FakeCancellationFactory {
    pub last: Mutex<Option<Arc<FakeCancellation>>>,
}

impl FakeCancellationFactory {
    pub fn new() -> Self {
        Self::default()
    }

    /// Cancela la sesión más reciente creada por esta factoría.
    pub fn cancel_latest(&self) {
        if let Some(c) = self.last.lock().unwrap().as_ref() {
            c.cancel();
        }
    }
}

impl codify_core::application::ports::CancellationFactory for FakeCancellationFactory {
    fn create(&self) -> Arc<dyn Cancellation> {
        let signal = Arc::new(FakeCancellation::new());
        *self.last.lock().unwrap() = Some(signal.clone());
        signal
    }
}
