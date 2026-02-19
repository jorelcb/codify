# Changelog - AI Context Generator

Todos los cambios notables de este proyecto seran documentados en este archivo.

Basado en [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) y [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0-alpha] - 2026-02-18 - Pivot: Generador de Contextos con IA

### BREAKING CHANGE

Redefinicion fundamental. El objetivo correcto: generar contextos usando LLM, no renderizar templates con variables.

### Removed
- Template Engine completo (~1300 lineas): tokenizer, AST, parser, engine, processor
- Commands obsoletos: validate_template, list_templates
- CLI validate command
- Tests BDD de template_entity y template_repository

### Changed
- Archivos de contexto reescritos (PROMPT.md, CONTEXT.md, SESSION_STATE.md)
- CLAUDE.md con descripcion inequivoca
- README.md, ARCHITECTURE.md, ROADMAP.md alineados con objetivo correcto

### Kept
- Filesystem infrastructure (FileWriter, DirectoryManager)
- DDD/Clean Architecture structure
- CLI base (Cobra)
- Application patterns (CQRS, DTOs)
- Templates como guias estructurales

## [1.2.0-beta] - 2025-11-26

### Added
- Filesystem Infrastructure: FileWriter y DirectoryManager
- Tests unitarios para filesystem

## [1.1.0] - 2025-10-21

### Added
- Application Layer (Commands, Queries, DTOs)
- CLI funcional conectado a Application Layer

## [1.0.0] - 2024-12-10

### Added
- Suite de testing (bash)
- Release estable inicial

## [0.2.0-alpha] - 2024-06-28

### Added
- DDD/Clean Architecture en templates
- Wizard interactivo

## [0.1.0-alpha] - 2024-02-15

### Added
- Prototipo inicial