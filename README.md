# AI Context Generator

> **Genera contextos optimizados para agentes de IA usando modelos de lenguaje**

[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/jorelcb/ai-context-generator/releases)
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
   output/mi-sistema-de-pagos/
   ├── AGENTS.md              # Root file: tech stack, comandos, convenciones
   └── context/
       ├── CONTEXT.md         # Arquitectura y diseno tecnico
       └── INTERACTIONS_LOG.md # Bitacora de sesiones y ADRs
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

# Generar specs a partir de un contexto existente
ai-context-generator spec my-api \
  --from-context ./output/my-api/
```

## Que genera

### Comando `generate`

Archivos de contexto siguiendo el estandar [AGENTS.md](https://github.com/anthropics/AGENTS.md) — el "sustrato cognitivo" para tu agente de IA:

| Archivo | Ubicacion | Contenido |
|---------|-----------|-----------|
| `AGENTS.md` | Raiz del proyecto | Root file: tech stack, comandos, convenciones, estructura |
| `CONTEXT.md` | `context/` | Arquitectura, componentes, flujo de datos, decisiones de diseno |
| `INTERACTIONS_LOG.md` | `context/` | Bitacora de sesiones y ADRs |

### Comando `spec`

Archivos de especificacion tecnica (SDD) generados a partir de contexto existente:

| Archivo | Contenido |
|---------|-----------|
| `CONSTITUTION.md` | DNA del proyecto: stack, principios, restricciones |
| `SPEC.md` | Especificaciones de features con criterios de aceptacion |
| `PLAN.md` | Diseno tecnico y decisiones de arquitectura |
| `TASKS.md` | Desglose de tareas con dependencias y prioridad |

## Arquitectura

Construido en **Go** con **DDD/Clean Architecture**:

- **Domain Layer**: Entidades, value objects, interfaces de servicio
- **Application Layer**: Commands y Queries (CQRS)
- **Infrastructure Layer**: LLM adapter (Anthropic), filesystem, template loader
- **Interfaces Layer**: CLI con Cobra (`generate`, `spec`, `list`)

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para detalles.

## Testing

```bash
go test ./...
```

## Documentacion

- [Architecture Guide](ARCHITECTURE.md) - DDD/Clean Architecture
- [Getting Started](GETTING_STARTED.md) - Guia de inicio
- [Roadmap](ROADMAP.md) - Plan de desarrollo
- [Changelog](context/CHANGELOG.md) - Historial de cambios

## Licencia

MIT License - ver [LICENSE](LICENSE).

---

**Construido para potenciar el desarrollo asistido por IA**
