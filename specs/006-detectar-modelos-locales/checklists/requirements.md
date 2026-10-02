# Specification Quality Checklist: La aplicación ve lo que ya tienes instalado

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-02
**Feature**: [spec.md](../spec.md)

## Content Quality

- [X] No implementation details (languages, frameworks, APIs)
- [X] Focused on user value and business needs
- [X] Written for non-technical stakeholders
- [X] All mandatory sections completed

## Requirement Completeness

- [ ] No [NEEDS CLARIFICATION] markers remain
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

- **Queda un marcador, y decide el tamaño del ciclo**: si entra hablar con un agente de línea de
  órdenes. Es la pregunta que #63 necesita respondida —«qué cuenta como local» cuando un proveedor
  es un programa y no una dirección— y toca el principio **[NON-NEGOTIABLE]** de cero-egress, que
  hoy se sostiene en que la dirección sea del propio equipo. Primera pregunta de `/speckit-clarify`.
- **El «por qué» de este spec no es una opinión: es lo que hay en la máquina.** Cuatro programas
  usables en el `PATH` y la aplicación enseñando una dirección por defecto hacia un servicio que no
  está instalado. Eso se comprobó, no se supuso.
- **SC-002 está escrito como una ausencia** —cero menciones de una dirección por defecto cuando no
  hay nada escuchando— porque el defecto actual es precisamente afirmar algo que no se verificó.
- **FR-003 nace de un caso real y no de imaginar uno**: `llama-server` compilado desde fuente vive
  fuera de toda ruta estándar. Una detección que solo mire sitios conocidos no lo encuentra.
