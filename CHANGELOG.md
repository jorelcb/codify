# Changelog - AI Context Generator

Todos los cambios notables de este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachanglog.com/en/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0-beta] - 2025-11-26

### Added
- **Core Filesystem Infrastructure:** Implementación de `FileWriter` y `DirectoryManager` para operaciones de I/O en disco.
- **Interfaces de Dominio:** Definición de `service.FileWriter` y `service.DirectoryManager` para abstracciones.
- **Tests Unitarios:** Cobertura inicial para los adaptadores de filesystem.

### Changed
- **Arquitectura Interna:** Refactorización de `ProjectGenerator` para usar inyección de dependencias (interfaces) y adherirse al Principio de Responsabilidad Única (SRP).
- **Visionado del Proyecto:** Actualización de `README.md`, `context/PROMPT.md` y `context/CONTEXT.md` para reflejar la visión de "Incepción" y la naturaleza recursiva del proyecto (IA construyendo para IA).

### Fixed
- **Ambigüedad de Rol:** Clarificación de la responsabilidad del proyecto como "Meta-Arquitecto" que genera contextos para Agentes de Desarrollo.
