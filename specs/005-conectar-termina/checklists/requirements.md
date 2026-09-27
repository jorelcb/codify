# Specification Quality Checklist: Conectar una cuenta llega a su fin

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-09-25
**Feature**: [spec.md](../spec.md)

## Content Quality

- [X] No implementation details (languages, frameworks, APIs)
- [X] Focused on user value and business needs
- [X] Written for non-technical stakeholders
- [X] All mandatory sections completed

## Requirement Completeness

- [X] No [NEEDS CLARIFICATION] markers remain
- [X] Requirements are testable and unambiguous
- [X] Success criteria are measurable
- [X] Success criteria are technology-agnostic (no implementation details)
- [X] All acceptance scenarios are defined
- [X] Edge cases are identified
- [X] Scope is clearly bounded
- [X] Dependencies and assumptions identified

## Feature Readiness

- [X] All functional requirements have clear acceptance criteria
- [X] User scenarios cover primary flows
- [X] Feature meets measurable outcomes defined in Success Criteria
- [X] No implementation details leak into specification

## Notes

- **Resuelto en la sesión de clarificación del 2026-09-26.** La vía delegada queda fuera **y
  declarada**: no se aplaza ningún requisito, porque ninguno la cubría. Lo que se gana es dejar de
  tener un camino invisible, que es literalmente el defecto que trajo este spec — y que el test de
  comandos huérfanos no habría cazado, porque la delegada es una rama dentro de un comando, no un
  comando. De ahí FR-012 y SC-007.
- **La frontera del idioma se decidió hacia el catálogo**: lo que cruza del núcleo a la piel son
  claves, nunca frases. Una frase redactada en el núcleo solo puede estar en un idioma.
- **Los nombres del backend aparecen en el spec a propósito** —`complete_connection`,
  `connect_provider`— en la sección que explica por qué existe. Sin ellos no se puede decir dónde
  está el hueco, y ese hueco *es* el motivo del spec. Los requisitos no los citan.
- **FR-010 es la clase de defecto que este proyecto ya conoce**: una cadena visible que escapa al
  catálogo. Lo nuevo es **dónde** — nace en el backend, y el test que persigue cadenas sueltas solo
  mira `ui/` y `strings.rs`. Hoy no se ve porque la interfaz ignora ese campo; se verá en cuanto
  deje de ignorarlo.
- **SC-001 es el mismo criterio que `004` dejó declarado en fallo.** No se hereda como aprobado:
  se vuelve a medir con una persona delante cuando el camino exista.
