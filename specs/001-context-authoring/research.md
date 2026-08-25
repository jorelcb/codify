# Research — Context Authoring (codify-NG)

Fase 0. Las 6 decisiones raíz fueron cerradas por el operador antes del plan; aquí se registran con racional y alternativas, junto con los unknowns técnicos resueltos. **No quedan NEEDS CLARIFICATION.**

## D1+D2 — Core Rust hexagonal + piel Tauri in-process
- **Decisión**: `codify-core` = crate de biblioteca Rust (dominio hexagonal, sin UI ni protocolo); `codify-app` = Tauri v2 que linkea el core **in-process**. Pieles futuras (MCP, CLI) = crates que linkean el mismo core.
- **Rationale**: un solo lenguaje, integración in-process sin IPC, y aun así reusable por otras pieles (la topología C se preserva porque el core es una biblioteca, no la app). Tauri v2 da app de escritorio ligera (WebView del SO) sin el peso de Electron.
- **Alternativas**: (a) **Core Go headless + Tauri cliente (sidecar)** — reusa el plumbing Go actual, pero mete dos lenguajes e IPC; rechazada por el operador a favor de un rewrite Rust limpio. (b) **Electron** — más pesado, JS-céntrico; rechazado. (c) **egui/nativo Rust puro** — menos flexible para UX rica de diffs/chat que un frontend web; rechazado para fase 1.

## D3 — Conectividad multi-backend con auth pluggable
- **Decisión**: Port `ModelProvider` con adapters: `ollama`, `llamacpp` (servidor con API OpenAI-compatible), `openai_compat` (genérico), y remotos vía **OAuth device-authorization flow** ("conectar cuenta"). Estrategia de auth como valor: `LocalEndpoint | ApiKey | OAuth`. Tokens en el **keychain del SO** (`keyring`).
- **Rationale**: cubre local (privacidad/coste/latencia) y remoto sin obligar a pegar API keys; el patrón "conectar" es generalizable. La mayoría de backends locales exponen API OpenAI-compatible, así que un adapter `openai_compat` cubre Ollama y llama.cpp-server con mínima superficie; adapters dedicados solo donde haga falta.
- **Cero-egress (SC-007)**: se garantiza en el **composition root**: en modo local, el `ProviderRegistry` se construye **solo** con adapters locales — ningún adapter de red existe en el grafo de objetos. Defensa en profundidad: un test que verifica 0 conexiones salientes a hosts no-locales durante una sesión local.
- **Alternativas**: solo API keys (más simple, incumple D3); embeber un modelo en el binario (rechazado en la auditoría: peso + calidad débil; el local va como proveedor externo vía Ollama/llama.cpp).

## D4 — Ingesta dirigida por el agente
- **Decisión**: el loop expone al LLM herramientas (`list_repo`, `read_file`, `fetch_url`) y **el LLM decide qué abrir**, siguiendo referencias (docs + estructura + muestreo selectivo de código). Presupuesto de exploración acotado; se **declara lo que quedó fuera**.
- **Rationale**: implementa "seguir referencias" y "muestreo selectivo" (Q3) de forma natural y transparente; alinea con el ethos agéntico (D5) y con la conectividad D3.
- **Alternativas**: índice de embeddings (mejor para repos enormes pero complejidad/costo prematuros — **diferido post-v1**); heurística fija (ciega a referencias — rechazada).

## D5 — Loop agéntico propio, interfaceado por ports
- **Decisión**: orquestador propio y mínimo en `application/authoring_loop.rs`, dependiente **solo de ports** inyectados (`ModelProvider`, `RepoNavigator`, `ReferenceResolver`, `DiffEngine`, `Prompter`, `AuditSink`). Bucle: percibir → decidir (tool-call) → actuar → integrar, hasta cerrar el contexto.
- **Rationale**: es el diferenciador del producto; controlarlo es clave para cero-egress, modelos locales y el loop **curado**. Interfaceado = testeable sin I/O y reusable por toda piel.
- **Alternativas**: framework de agentes externo (mete supuestos ajenos, dificulta cero-egress y determinismo de test — rechazado).

## D6 — Core dueño del diff/approval
- **Decisión**: `DiffEngine` (generar/aplicar/revertir) + `AuditSink` (log append-only) viven en el core como ports/servicios. La **clasificación de riesgo** (`RiskClassifier`) es un port; en v1 su default es **conservador** (todo cambio no trivial = alto impacto → requiere aprobación), y su afinamiento es el **spec derivado** ya declarado en FR-012. La piel solo renderiza el diff y captura la decisión.
- **Rationale**: mantiene la lógica curada en el dominio, reusable e auditable; la piel es tonta.
- **Alternativas**: que la piel implemente el diff/approval (rompería reuso entre pieles — rechazado).

## D7 — Verificar es comprobar la cita, no confiar en la fuente declarada

- **Decisión**: un segmento solo es `grounded` si su cita textual aparece en el material leído
  (FR-006a/b/c). Lo que no se sostiene se degrada a tentativo declarando el motivo.
- **Alternativas**: (a) **confiar en `sources`** — lo que había, y produjo el hallazgo F-1;
  (b) **cuarto estado para «procedencia falsa»** — rechazado: competiría con SC-002 y multiplica
  lo que la interfaz debe explicar; (c) **comparación byte a byte** — rechazado, un salto de
  línea de más invalidaría una cita legítima.
- **Revisitar si**: los modelos empiezan a citar con fidelidad literal alta y el mínimo de doce
  caracteres pasa a ser el criterio que más degrada.

## D8 — Los artefactos propios se reconocen por ruta canónica

- **Decisión**: lo que vive en una ruta que FR-005 define es salida del sistema; se lee, no
  fundamenta (FR-006d).
- **Alternativas**: (a) **marcador dentro del archivo** — sobrevive a moverlo y distingue lo
  escrito a mano, pero ensucia el artefacto y se pierde si alguien borra la marca;
  (b) **registro de escrituras** — preciso, pero vive en la sesión y el caso es leer artefactos
  de una **anterior**, de la que no queda rastro; (c) **solo si lo editó un humano** — exigiría
  huellas y fallaría en silencio cuando no se pueda saber.
- **Revisitar si**: aparece demanda real de fundamentar sobre un `AGENTS.md` escrito a mano, que
  es el coste que (a) evitaba y hoy se acepta.

## Unknowns técnicos resueltos
- **Diff**: crate `similar` (line/word diff, robusto, sin deps pesadas).
- **HTTP**: `reqwest` (async, TLS) para proveedores remotos y para `fetch_url` (solo URLs públicas; auth/privados **fuera de v1**, se reportan como no resueltos — FR-003/FR-004).
- **OAuth**: `oauth2` (device-authorization flow) + `keyring` para almacenamiento seguro.
- **Async**: `tokio`.
- **Detección de idioma**: heurística sobre el contenido dominante del repo (README/docs) con **override** explícito (FR-019); librería ligera de detección o prompt al modelo — se decide en tasks.
- **Marcado grounded vs tentativo**: representado en el dominio (segmentos del `ContextArtifact`), no como convención de texto frágil; el render lo distingue.
