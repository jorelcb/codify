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

- **Resuelto en la sesión de clarificación del 2026-10-02.** Lo que clasifica a un proveedor no es
  dónde corre el programa sino **dónde ocurre la inferencia**. Este ciclo detecta todo y cablea solo
  lo que puede sostener esa propiedad; un agente de inferencia remota se encuentra y se nombra por
  lo que es.
- **Y se corrigió un error de lectura del agente, no del spec**: el cero-egress estructural rige el
  camino local, que es lo que la constitución dice. El producto siempre quiso **los dos caminos**.
  Tratar «nada sale del equipo» como condición universal dejaría fuera la mitad del producto. No
  hace falta enmienda constitucional: hacía falta dejar de leerla de más.
- **La variable de entorno deja de ser el modo de integrar un modelo local.** Ninguna herramienta
  comparable le pide eso al usuario. Sobrevive como anulación de desarrollo, declarada.
- **El «por qué» de este spec no es una opinión: es lo que hay en la máquina.** Cuatro programas
  usables en el `PATH` y la aplicación enseñando una dirección por defecto hacia un servicio que no
  está instalado. Eso se comprobó, no se supuso.
- **SC-002 está escrito como una ausencia** —cero menciones de una dirección por defecto cuando no
  hay nada escuchando— porque el defecto actual es precisamente afirmar algo que no se verificó.
- **FR-003 nace de un caso real y no de imaginar uno**: `llama-server` compilado desde fuente vive
  fuera de toda ruta estándar. Una detección que solo mire sitios conocidos no lo encuentra.
