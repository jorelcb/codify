# Specification Quality Checklist: La superficie de conexión y modo

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-08-27
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

- **Tres criterios de éxito los decide una persona, no un test**: SC-001, SC-003 y SC-004
  preguntan si alguien *entiende* lo que ve. Están escritos como escenarios ejecutables —«conecta
  sin preguntar», «señala qué control decide si algo sale»— pero su verificación es de las que
  este proyecto marca `[~]` hasta que alguien se sienta delante. **SC-002 y SC-005 sí son
  automáticos**, y son los que evitan que el arreglo se deshaga sin que nadie se entere.
- **Resuelto en la sesión de clarificación del 2026-08-27**: el formulario se revela al pedirlo,
  no ocupa sitio en el caso por defecto; y el modo se muestra en dos sitios con una sola fuente de
  verdad, con comprobación automática de que no puedan discrepar. La segunda pregunta salió de
  mirar el HTML: el modo ya aparecía duplicado y nadie lo había notado.
- **SC-005 quedó cuantificado** —24 caracteres visibles— porque «ancho legible» no es comprobable.
  El número está elegido, no medido, y así se dice en Assumptions.
- **FR-006 es el que da valor duradero**: sin una comprobación automática de unicidad de nombres,
  el defecto que originó este spec volvería la próxima vez que alguien copie una sección — que es
  exactamente como llegó.
