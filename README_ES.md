# 🧠 AI Context Generator

<div align="center">

[![Version](https://img.shields.io/badge/version-2.0.0-blue?style=for-the-badge)](https://github.com/jorelcb/ai-context-generator/releases)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.21)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Powered%20by-Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)

**Dale a tu agente de IA el plano maestro que necesita antes de escribir la primera linea de codigo** 🏗️

*Porque un agente sin contexto es un pasante con acceso root.*

[English](README.md) | **[Español]**

[Quick Start](#-quick-start) · [Features](#-features) · [Presets](#-presets) · [Arquitectura](#%EF%B8%8F-arquitectura) · [Docs](#-documentacion)

</div>

---

## 🎯 El Problema

Le dices a tu agente: *"Construye una API de pagos en Go con microservicios"*

Y el agente, con toda su capacidad, improvisa:
- Estructura de carpetas que nadie pidio
- Patrones que contradicen tu arquitectura
- Decisiones que vas a revertir en la siguiente sesion
- Cero continuidad entre sesiones

**No es culpa del agente. Es que empieza desde cero. Cada. Vez.** 🔄

## 💡 La Solucion

**AI Context Generator** toma la descripcion de tu proyecto y genera archivos de contexto inteligentes usando Anthropic Claude. Archivos que le dan a tu agente el plano maestro, las restricciones de dominio y la memoria arquitectonica que necesita.

## ✨ Antes y despues

### 😱 Sin contexto (la realidad actual)

```
Tu: "Crea una API de pagos en Go"

Agente: *crea main.go con todo en un archivo*
Tu: "No, usa Clean Architecture"
Agente: *crea estructura pero mezcla domain con infra*
Tu: "Los repositorios van en infrastructure"
Agente: *refactoriza por tercera vez*
Tu: "¿Y los tests BDD que pedi ayer?"
Agente: "¿Tests BDD? Es la primera vez que me lo mencionas"

Resultado: 45 minutos corrigiendo al agente 😤
```

### 🚀 Con AI Context Generator

```
Tu: "Crea una API de pagos en Go"

Agente: *lee AGENTS.md y CONTEXT.md*
Agente: "Veo que usas DDD con Clean Architecture, PostgreSQL,
         y testing BDD con Godog. Creo el endpoint de pagos
         en internal/domain/payment/ siguiendo tus patrones."

Resultado: Codigo coherente desde la primera linea ✨
```

## ⚡ Quick Start

### Instalacion

```bash
# go install (recomendado)
go install github.com/jorelcb/ai-context-generator/cmd/ai-context-generator@latest

# O build from source
git clone https://github.com/jorelcb/ai-context-generator.git
cd ai-context-generator && go build -o bin/ai-context-generator ./cmd/ai-context-generator
```

### Tu primer contexto en 30 segundos

```bash
# 1. Configura tu API key
export ANTHROPIC_API_KEY="sk-ant-..."

# 2. Describe tu proyecto
ai-context-generator generate payment-service \
  --description "Microservicio de pagos en Go con gRPC, PostgreSQL y Kafka. \
  DDD con Clean Architecture. Stripe como procesador."

# 3. Listo. Archivos generados.
```

### Lo que vas a ver

```
🚀 Generating context for: payment-service
  Model: claude-sonnet-4-6
  Preset: default

  [1/3] Generating AGENTS.md... ✓
  [2/3] Generating CONTEXT.md... ✓
  [3/3] Generating INTERACTIONS_LOG.md... ✓

📁 Output: output/payment-service/
  ├── AGENTS.md              → Root file (tech stack, comandos, convenciones)
  └── context/
      ├── CONTEXT.md         → Arquitectura y diseno tecnico
      └── INTERACTIONS_LOG.md → Bitacora de sesiones y ADRs

✅ Done! 3 files generated
   Total tokens: ~12,450
```

## 🎨 Features

### 📋 Comando `generate` — Contexto para tu agente

Genera archivos siguiendo el estandar [AGENTS.md](https://github.com/anthropics/AGENTS.md):

| Archivo | Que hace |
|---------|----------|
| `AGENTS.md` | Root file: tech stack, comandos, convenciones, estructura |
| `CONTEXT.md` | Arquitectura, componentes, flujo de datos, decisiones |
| `INTERACTIONS_LOG.md` | Bitacora de sesiones y ADRs |

Coloca estos archivos en la raiz de tu proyecto. Agentes compatibles (Claude Code, Cursor, Codex, etc.) los leen automaticamente.

### 📐 Comando `spec` — Especificaciones SDD

A partir de un contexto existente, genera especificaciones tecnicas listas para implementar:

```bash
ai-context-generator spec payment-service \
  --from-context ./output/payment-service/
```

| Archivo | Que hace |
|---------|----------|
| `CONSTITUTION.md` | DNA del proyecto: stack, principios, restricciones |
| `SPEC.md` | Features con criterios de aceptacion |
| `PLAN.md` | Diseno tecnico y decisiones de arquitectura |
| `TASKS.md` | Desglose de tareas con dependencias y prioridad |

### 🔍 Comando `list` — Proyectos generados

```bash
ai-context-generator list
```

## 🎭 Presets

Elige la filosofia de tus contextos:

### `--preset default` *(por defecto)*

Opinionado: **DDD + Clean Architecture + BDD**. Incluye:
- Separacion estricta de capas (Domain → Application → Infrastructure → Interfaces)
- Testing BDD con coverage targets (80% dominio, 70% aplicacion)
- Observabilidad con OpenTelemetry
- Inyeccion de dependencias obligatoria
- Restricciones DEBE/NO DEBE

```bash
ai-context-generator generate my-api \
  --description "API REST de inventarios en Go"
# Usa preset default automaticamente
```

### `--preset neutral`

Sin postura arquitectonica. Deja que el LLM adapte la estructura al proyecto:

```bash
ai-context-generator generate my-api \
  --description "API REST de inventarios en Go" \
  --preset neutral
```

## ⚙️ Opciones

```bash
ai-context-generator generate <nombre> [flags]
```

| Flag | Corto | Descripcion | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Descripcion del proyecto *(requerido)* | — |
| `--preset` | `-p` | Preset de templates | `default` |
| `--model` | `-m` | Modelo de Claude | `claude-sonnet-4-6` |
| `--language` | `-l` | Hint de lenguaje | — |
| `--type` | `-t` | Hint de tipo (api, cli, lib...) | — |
| `--architecture` | `-a` | Hint de arquitectura | — |

## 🏗️ Arquitectura

Construido en Go con lo que predica — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Logica de negocio pura
│   ├── project/         Entidad Project (aggregate root)
│   ├── template/        Entidad Template
│   ├── shared/          Value objects, errores de dominio
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Casos de uso (CQRS)
│   ├── command/         GenerateContext, GenerateSpec
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementaciones
│   ├── llm/             Anthropic Claude adapter + prompt builder
│   ├── template/        Template loader con mapping configurable
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 CLI con Cobra
    └── cli/commands/    generate, spec, list
```

La regla de oro: `Infrastructure → Application → Domain`. Nada en domain depende de nada externo.

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para el detalle completo.

## 🧪 Tests

```bash
# Todos los tests
go test ./...

# BDD con Godog
go test ./tests/...
```

## 📊 Estado del proyecto

**v2.0.0** 🎉

✅ **Funcionando:**
- Generacion de contextos con Claude API (streaming)
- Generacion de specs SDD a partir de contexto existente
- Sistema de presets (default DDD/BDD, neutral)
- Estandar AGENTS.md como root file
- CLI con Cobra (generate, spec, list)
- Templates como guias estructurales para el LLM

🚧 **Proximo:**
- Tests de integracion end-to-end
- Retries y manejo de rate limits
- Modo interactivo (wizard)
- Segundo proveedor LLM
- Binary builds y distribucion

👉 [Roadmap completo](ROADMAP.md)

## 💡 FAQ

**¿Necesito una API key de Anthropic?**
Si. Exportala como `ANTHROPIC_API_KEY`. Obtener en [console.anthropic.com](https://console.anthropic.com).

**¿Cuanto cuesta cada generacion?**
Aproximadamente 3 llamadas API para `generate`, 4 para `spec`. Con claude-sonnet-4-6, cada generacion cuesta centavos.

**¿Funciona con otros LLMs?**
Por ahora solo Anthropic Claude. La interfaz `LLMProvider` esta disenada para agregar mas proveedores sin cambiar el core.

**¿Los templates son fijos?**
Son guias estructurales, no output renderizable. El LLM genera contenido inteligente y especifico a tu proyecto siguiendo la estructura del template.

**¿Puedo personalizar los templates?**
Puedes crear tus propios presets en el directorio `templates/`. Cada preset necesita 3 archivos: `agents.template`, `context.template`, `interactions.template`.

**¿Que agentes soportan los archivos generados?**
Cualquier agente compatible con el estandar [AGENTS.md](https://github.com/anthropics/AGENTS.md): Claude Code, Cursor, GitHub Copilot Workspace, Codex, y mas.

## 📚 Documentacion

- [🏛️ Architecture Guide](ARCHITECTURE.md) — DDD/Clean Architecture
- [🚀 Getting Started](GETTING_STARTED.md) — Guia paso a paso
- [🗺️ Roadmap](ROADMAP.md) — Plan de desarrollo
- [📝 Changelog](context/CHANGELOG.md) — Historial de cambios

## 📄 Licencia

MIT License — ver [LICENSE](LICENSE).

---

<div align="center">

**Construido para potenciar el desarrollo asistido por IA** 🧠

*"Un agente sin contexto es un pasante con acceso root"*

⭐ Si te sirvio, dale una estrella — nos motiva a seguir construyendo

[🐛 Reportar bug](https://github.com/jorelcb/ai-context-generator/issues) · [💡 Sugerir feature](https://github.com/jorelcb/ai-context-generator/issues) · [🗺️ Ver roadmap](ROADMAP.md)

</div>
