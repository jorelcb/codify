# Changelog - AI Context Generator

Todos los cambios notables de este proyecto seran documentados en este archivo.

Basado en [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) y [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-02-19 - Restructure output (AGENTS.md standard) + spec command

### Added
- Output reestructurado: AGENTS.md (root) + context/ (CONTEXT.md, INTERACTIONS_LOG.md)
- System prompts con XML tags (`<role>`, `<task>`, `<workflow>`, `<output_quality>`)
- Templates optimizados para output conciso y accionable
- Comando `spec`: genera CONSTITUTION.md, SPEC.md, PLAN.md, TASKS.md
- Spec context-dependent: lee contexto existente via `--from-context`
- Template loader configurable (mapping custom para spec/)
- Context reader para leer AGENTS.md + CONTEXT.md existentes

### Removed
- PROMPT.md y SCAFFOLDING.md (absorbidos en AGENTS.md)

## [1.0.0] - 2026-02-19 - Primera version funcional

### Added
- Generacion de archivos de contexto usando Anthropic Claude API
- Streaming con progreso por archivo (`[1/4] Generating PROMPT.md... done`)
- Generacion per-file (4 llamadas API independientes, una por archivo)
- CLI con flags: `--description`, `--language`, `--type`, `--architecture`, `--model`
- PromptBuilder para construir system/user prompts por archivo
- FileSystemTemplateLoader para cargar templates como guias estructurales
- GenerateContextCommand (orquesta flujo completo)
- AnthropicProvider con SDK oficial (`anthropic-sdk-go`)
- DDD/Clean Architecture: Domain, Application, Infrastructure, Interfaces
- Filesystem infrastructure (FileWriter, DirectoryManager)
- Value objects con validacion (ProjectDescription, Language, Architecture, etc.)
- Tests unitarios para prompt builder, template loader, generate command

### Output
Genera 4 archivos de contexto optimizados para agentes de IA:
- `PROMPT.md` - Rol y mision para el agente de desarrollo
- `CONTEXT.md` - Arquitectura, patrones, dominio
- `SCAFFOLDING.md` - Estructura recomendada del proyecto
- `INTERACTIONS_LOG.md` - Bitacora inicial de desarrollo
