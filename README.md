# AI Context Generator

> **Genera contextos optimizados para agentes de IA usando modelos de lenguaje**

[![Version](https://img.shields.io/badge/version-2.0.0--alpha-blue.svg)](https://github.com/jorelcb/ai-context-generator/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## El Problema

Los agentes de IA son ingenieros capaces pero empiezan desde cero. Sin contexto arquitectonico, sin restricciones de dominio, sin un plano maestro. El resultado: codigo inconsistente y decisiones fragmentadas.

## La Solucion

**AI Context Generator** toma la descripcion de tu proyecto y genera archivos de contexto inteligentes usando modelos de IA. Estos archivos le dan a tu agente de desarrollo el contexto que necesita para construir con coherencia desde la primera linea.

## Como Funciona

```
1. Tu describes tu proyecto:
   "Sistema de pagos en Go con microservicios, DDD y PostgreSQL"

2. La herramienta usa templates (guias estructurales) + un LLM:
   Templates definen la estructura -> LLM genera el contenido inteligente

3. Obtienes archivos de contexto listos para usar:
   output/mi-sistema-de-pagos/
   ├── PROMPT.md            # Rol y mision para tu agente de IA
   ├── CONTEXT.md           # Arquitectura, patrones, dominio
   ├── SCAFFOLDING.md       # Estructura recomendada
   └── INTERACTIONS_LOG.md  # Bitacora inicial
```

## Quick Start

```bash
# Generar contextos para tu proyecto
ai-context-generator generate \
  --description "API REST de gestion de inventarios en Go con Clean Architecture y PostgreSQL"

# Listar proyectos generados
ai-context-generator list
```

## Que genera

Los archivos de contexto son el "sustrato cognitivo" que entregas a tu agente de IA (Claude, GPT, Gemini, Cursor, etc.) para que construya tu proyecto con contexto perfecto.

| Archivo | Contenido |
|---------|-----------|
| `PROMPT.md` | Rol, mision, restricciones y directrices para el agente |
| `CONTEXT.md` | Arquitectura, stack, patrones de diseno, dominio de negocio |
| `SCAFFOLDING.md` | Estructura de directorios y archivos recomendada |
| `INTERACTIONS_LOG.md` | Bitacora inicial con decisiones de diseno |

## Arquitectura

Construido en **Go** con **DDD/Clean Architecture**:

- **Domain Layer**: Entidades, value objects, interfaces de servicio
- **Application Layer**: Commands y Queries (CQRS)
- **Infrastructure Layer**: LLM adapters, filesystem, persistence
- **Interfaces Layer**: CLI con Cobra

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para detalles.

## Testing

BDD con Godog:

```bash
go test ./...
```

## Estado del Proyecto

**v2.0.0-alpha** - En desarrollo activo tras redefinicion del objetivo.

- Infraestructura DDD/Clean Architecture
- Filesystem (FileWriter, DirectoryManager)
- CLI base con Cobra
- **En progreso**: Integracion con LLM

## Documentacion

- [Architecture Guide](ARCHITECTURE.md) - DDD/Clean Architecture
- [Roadmap](ROADMAP.md) - Plan de desarrollo
- [Changelog](CHANGELOG.md) - Historial de cambios

## Licencia

MIT License - ver [LICENSE](LICENSE).

---

**Construido para potenciar el desarrollo asistido por IA**