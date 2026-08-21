# CLAUDE.md

Guía para agentes de código que trabajen en este repositorio.

## Qué es este proyecto

**codify** es la capa de authoring y custodia del ciclo de vida del contexto de un proyecto para
agentes de código. Reescritura en Rust: núcleo hexagonal (`crates/codify-core`) + app de
escritorio Tauri (`crates/codify-app`).

Principio rector: **lo que no se puede verificar contra una fuente no se afirma.** Varias de esas
garantías están puestas en el sistema de tipos, no en la buena voluntad — una referencia no
resuelta no *puede* llevar contenido, y en modo local el grafo de objetos no admite proveedores
remotos.

## Reglas que no se negocian

La constitución del proyecto vive en `.specify/memory/constitution.md` (local, no versionada) y
manda. En resumen:

- **Arquitectura hexagonal**: el código apunta solo hacia adentro (`infrastructure → application →
  domain`). El dominio no importa nada externo. Hay una **fitness function que lo verifica** en
  `crates/codify-core/tests/arch_deps.rs` — si falla, el diseño está mal, no el test.
- **Ports en la capa que los nombra**: ¿el Dominio lo nombra? entonces `domain/interfaces/`; si no,
  es un capability port de `application/`. Firmas con tipos de dominio únicamente.
- **Nombres sin decoración**: prohibido `…Port`, `…Impl`, `I…`.
- **Test-first** (TDD/BDD) con las Test Desiderata como criterio de calidad.
- **Conventional Commits**, y **cero atribución de IA** en commits y notas de release.

## Comandos

```bash
cargo test --workspace                                   # incluye arquitectura y cero-egress
cargo clippy --workspace --all-targets -- -D warnings
cargo fmt --all -- --check
cargo run -p codify-app                                  # levanta la aplicación
./scripts/quickstart-fixture.sh                          # genera el fixture de validación
```

El fixture reproduce el patrón que originó el proyecto: un `README` que **referencia** un `SPEC`
hermano y que, leído solo, invita a suponer un broker de mensajes — el `SPEC` dice explícitamente
que no lo hay. Si el contexto generado menciona uno, el agente se lo inventó en vez de seguir la
referencia. Los escenarios de validación y qué cubre ya el build están en los `quickstart.md`.

## Cómo se trabaja aquí

Spec-Driven Development con [spec-kit](https://github.com/github/spec-kit). Antes de implementar,
lee el spec del feature: qué se decidió, qué se descartó y por qué. Los specs registran también las
**alternativas rechazadas con la señal que justificaría revisitarlas** — si estás a punto de
proponer una, mira si ya se evaluó.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
`specs/002-authoring-experience/plan.md` (codify — la experiencia de authoring).
<!-- SPECKIT END -->
