# Contracts — Ports (core hexagonal)

Contratos a nivel de biblioteca: los **ports** del crate `codify-core`. Firmas ilustrativas (Rust). Los adapters concretos viven en `infrastructure/`; el loop (`application/`) depende **solo** de estos traits (D5).

## Driving port (entrada de casos de uso)
La piel (Tauri) invoca al core a través de este servicio.

```rust
#[async_trait]
pub trait AuthoringService {
    async fn start_session(&self, req: StartSession) -> Result<SessionId>;
    async fn session_state(&self, id: SessionId) -> Result<SessionState>;
    /// Mensaje del usuario en el loop de refinamiento (lenguaje natural).
    async fn submit_message(&self, id: SessionId, msg: String) -> Result<()>;
    async fn pending_proposals(&self, id: SessionId) -> Result<Vec<ChangeProposal>>;
    async fn decide(&self, id: SessionId, decision: ApprovalDecision) -> Result<()>;
    async fn set_locale(&self, id: SessionId, locale: Locale) -> Result<()>;
}
// StartSession { repo_root, mode: Local|Hybrid, locale: Option<Locale> }
```

## Driven ports (capacidades que el loop consume — inyectadas)

```rust
#[async_trait]
pub trait ModelProvider {           // D3 — un adapter por backend
    async fn complete(&self, req: Completion) -> Result<Completion Out>; // texto o tool-calls
    fn is_local(&self) -> bool;     // elegibilidad en modo Local (cero-egress)
    fn tier_hint(&self) -> Tier;    // Cheap | Heavy
}

#[async_trait]
pub trait RepoNavigator {           // D4 — muestreo dirigido por el agente
    async fn list(&self, path: &RepoPath) -> Result<Vec<Entry>>;
    async fn read(&self, path: &RepoPath) -> Result<FileContent>;   // acotado por presupuesto
}

#[async_trait]
pub trait ReferenceResolver {       // FR-002/003/004
    async fn resolve(&self, r: &Reference) -> ReferenceResolution;  // Resolved | Inaccessible | RequiresAuth | OutOfScope
    // v1: LocalPath + PublicUrl; RequiresAuth => reportado, nunca inventado
}

pub trait DiffEngine {              // D6
    fn make(&self, before: &str, after: &str) -> Diff;
    fn apply(&self, target: &mut Artifact, diff: &Diff) -> Result<()>;
    fn revert(&self, target: &mut Artifact, diff: &Diff) -> Result<()>;
}

pub trait RiskClassifier {          // FR-012 (default conservador en v1; afinado = spec derivado)
    fn classify(&self, p: &ChangeProposal) -> RiskLevel;   // Low | HighImpact
}

// Nota de capa: Prompter es un port cuyo ADAPTER es la piel (Interface Adapter). La piel es a la
// vez driving (invoca comandos) e implementa este callback humano; NO rompe la Regla de Dependencia
// (el core sigue dependiendo solo de la abstracción, nunca del adapter).
#[async_trait]
pub trait Prompter {                // borde humano; la piel lo implementa
    async fn ask(&self, q: Question) -> Answer;
    async fn present(&self, p: &ChangeProposal) -> ApprovalDecision; // solo HighImpact bloquea
}

pub trait AuditSink { fn record(&self, e: AuditEvent); }   // append-only

#[async_trait]
pub trait LocaleDetector { async fn detect(&self, repo: &Repository) -> Locale; } // FR-019
```

## Reglas de contrato (verificables con contract tests por port)
- Todo port tiene un **fake in-memory** y (donde aplique) un **adapter real**; ambos pasan la misma suite (patrón hex-integration-test).
- `ModelProvider.is_local() == false` **no puede existir** en el grafo cuando `mode = Local` (verificado en el composition root + test de egress).
- `ReferenceResolver` nunca devuelve `Resolved` con contenido fabricado para un `RequiresAuth`/`Inaccessible`.
- **Un `Segment` solo es `Grounded` si sus `quotes` aparecen en el material leído** (FR-006a). Y solo si esas citas están en material que **puede** respaldar: un artefacto del propio sistema se lee pero no fundamenta (FR-006d), así que un segmento apoyado solo en él se degrada igual. La procedencia que el modelo declara sin cita comprobable no cuenta: se degrada a `Tentative` con el motivo (FR-006c). Igual para `Contradiction`, que exige cita de cada fuente en conflicto.
- `DiffEngine.apply` es reversible con `revert` (property test: apply∘revert = identidad).
- **La generación no sobrescribe a ciegas** (US3): antes de escribir consulta `ArtifactWriter.read_existing`; si ya hay contexto y difiere, produce una `ChangeProposal` de origen `Generation` y espera decisión. Rechazar deja el archivo intacto y lo declara como `Skipped` con su motivo — nunca como un olvido.
