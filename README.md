# AI Context Generator

> **Genera contextos optimizados para agentes de IA usando modelos de lenguaje**

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/jorelcb/ai-context-generator/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## El Problema

Los agentes de IA son ingenieros capaces pero empiezan desde cero. Sin contexto arquitectonico, sin restricciones de dominio, sin un plano maestro. El resultado: codigo inconsistente y decisiones fragmentadas.

## La Solucion

**AI Context Generator** toma la descripcion de tu proyecto y genera archivos de contexto inteligentes usando Anthropic Claude. Estos archivos le dan a tu agente de desarrollo el contexto que necesita para construir con coherencia desde la primera linea.

## Como Funciona

```
1. Tu describes tu proyecto:
   "Sistema de pagos en Go con microservicios, DDD y PostgreSQL"

2. La herramienta usa templates (guias estructurales) + Claude API:
   Templates definen la estructura -> LLM genera el contenido inteligente

3. Obtienes archivos de contexto listos para usar:
   output/mi-sistema-de-pagos/context/
   ├── PROMPT.md            # Rol y mision para tu agente de IA
   ├── CONTEXT.md           # Arquitectura, patrones, dominio
   ├── SCAFFOLDING.md       # Estructura recomendada
   └── INTERACTIONS_LOG.md  # Bitacora inicial
```

## Quick Start

```bash
# Requisito: API key de Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Generar contextos para tu proyecto
ai-context-generator generate my-api \
  --description "API REST de gestion de inventarios en Go con Clean Architecture y PostgreSQL"

# Con hints opcionales
ai-context-generator generate my-api \
  --description "API REST de gestion de inventarios" \
  --language go --type api --architecture ddd

# Usar un modelo especifico
ai-context-generator generate my-api \
  --description "..." --model claude-sonnet-4-6
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
- **Infrastructure Layer**: LLM adapter (Anthropic), filesystem, template loader
- **Interfaces Layer**: CLI con Cobra

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para detalles.

## Testing

```bash
go test ./...
```

## Documentacion

- [Architecture Guide](ARCHITECTURE.md) - DDD/Clean Architecture
- [Getting Started](GETTING_STARTED.md) - Guia de inicio
- [Roadmap](ROADMAP.md) - Plan de desarrollo
- [Changelog](CHANGELOG.md) - Historial de cambios

## Licencia

MIT License - ver [LICENSE](LICENSE).

---

**Construido para potenciar el desarrollo asistido por IA**
