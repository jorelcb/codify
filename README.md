# codify

> **Estado: `0.1.0` — en desarrollo temprano.** No es un reemplazo de `codify` v4.x todavía.
> El CLI en producción vive en **[jorelcb/codify-og](https://github.com/jorelcb/codify-og)**
> y se sigue instalando con `brew install jorelcb/tap/codify`.

Capa de **authoring** y de **custodia del ciclo de vida** del contexto de un proyecto para
agentes de código.

## Qué problema resuelve

Un agente de código rinde según el contexto que le des. Escribir ese contexto a mano es
tedioso; generarlo con una herramienta que solo mira un archivo suelto es peor: rellena los
huecos con lo que *suele* hacerse y te entrega una arquitectura plausible y falsa, afirmada
con total confianza.

codify parte de un principio distinto: **lo que no se puede verificar contra una fuente no se
afirma**. El agente lee el repositorio, **sigue las referencias que encuentra** (specs, docs
enlazadas), y todo lo que no logra fundamentar queda marcado como tentativo — nunca
presentado como hecho.

## Cómo está construido

- **`crates/codify-core`** — el núcleo: dominio del authoring y el loop agéntico. Arquitectura
  hexagonal, sin UI ni protocolo adentro. Es una biblioteca reutilizable: hoy la consume la app
  de escritorio; mañana puede exponerse como servidor MCP o CLI sin reescribir el dominio.
- **`crates/codify-app`** — la piel (Tauri). Renderiza diffs y captura decisiones; la lógica
  vive en el core.

Algunas garantías están puestas en el sistema de tipos, no en la buena voluntad: una
referencia no resuelta **no puede** llevar contenido, un fragmento sin procedencia declarada se
degrada a tentativo, y en modo local el grafo de objetos no admite proveedores de modelo
remotos (cero-egress **estructural**, no un flag que un bug pueda saltarse).

## Modelos

Provider-agnóstico. Backends locales (`llama.cpp`, Ollama) vía API OpenAI-compatible, y
remotos como opción explícita. El **modo totalmente local garantiza que nada del repositorio
sale hacia servicios de nube**.

## Desarrollo

```bash
cargo test --workspace     # incluye las fitness functions de arquitectura y cero-egress
cargo clippy --workspace --all-targets -- -D warnings
cargo fmt --all -- --check
```

El proyecto se desarrolla con [Spec-Driven Development](https://github.com/github/spec-kit).
El diseño vigente y su estado de avance viven en
[`specs/001-context-authoring/`](./specs/001-context-authoring/).

## Relación con codify v4.x

Esto es una **reescritura**, no un refactor. El proyecto original (Go) sigue disponible y
funcionando en [jorelcb/codify-og](https://github.com/jorelcb/codify-og); su historia,
releases y documentación permanecen intactas allí.

## Licencia

Apache-2.0 — ver [LICENSE](./LICENSE).
