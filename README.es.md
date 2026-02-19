# AI Context Generator

> **Genera contextos optimizados para agentes de IA usando modelos de lenguaje**

[![Version](https://img.shields.io/badge/version-2.0.0--alpha-blue.svg)](https://github.com/jorelcb/ai-context-generator/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

[English](README.md) | **Espanol**

---

## El Problema

Los agentes de IA son ingenieros capaces pero carecen de contexto inicial. Sin arquitectura definida, sin restricciones de dominio, sin un plano maestro, el resultado es codigo inconsistente.

## La Solucion

Describes tu proyecto, la herramienta usa templates como guias estructurales + un modelo de IA (LLM) para generar archivos de contexto inteligentes. Esos archivos le dan a tu agente de desarrollo el contexto perfecto desde el dia uno.

## Como Funciona

```
1. Describes tu proyecto en texto libre:
   "Sistema de pagos en Go con microservicios, DDD y PostgreSQL"

2. La herramienta combina templates (guias estructurales) + LLM:
   Templates definen la estructura -> LLM genera contenido inteligente

3. Obtienes archivos de contexto listos para usar:
   output/mi-sistema/
   ├── PROMPT.md            # Rol y mision para tu agente de IA
   ├── CONTEXT.md           # Arquitectura, patrones, dominio
   ├── SCAFFOLDING.md       # Estructura recomendada
   └── INTERACTIONS_LOG.md  # Bitacora inicial
```

## Quick Start

```bash
ai-context-generator generate \
  --description "API REST de gestion de inventarios en Go con Clean Architecture"
```

## Estado

**v2.0.0-alpha** - En desarrollo activo.

- DDD/Clean Architecture implementada
- Filesystem infrastructure lista
- CLI base funcional
- **En progreso**: Integracion con LLM

## Documentacion

- [Architecture](ARCHITECTURE.md) - Guia de arquitectura
- [Roadmap](ROADMAP.md) - Plan de desarrollo
- [Changelog](CHANGELOG.md) - Historial de cambios

## Licencia

MIT - ver [LICENSE](LICENSE).