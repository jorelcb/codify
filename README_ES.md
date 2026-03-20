# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-1.9.0-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.21)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Dale a tu agente de IA el plano maestro que necesita antes de escribir la primera linea de codigo** 🏗️

*Porque un agente sin contexto es un pasante con acceso root.*

[English](README.md) | **[Español]**

[Quick Start](#-quick-start) · [MCP Server](#-mcp-server) · [Features](#-features) · [Skills](#-agent-skills) · [Guias por Lenguaje](#-guias-por-lenguaje) · [Presets](#-presets) · [Arquitectura](#%EF%B8%8F-arquitectura)

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

**Codify** toma la descripcion de tu proyecto y genera archivos de contexto inteligentes usando LLMs (Anthropic Claude o Google Gemini). Archivos que le dan a tu agente el plano maestro, las restricciones de dominio y la memoria arquitectonica que necesita.

Sigue el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md) — una especificacion abierta respaldada por la Linux Foundation para proveer contexto de proyecto a agentes de IA. Esto significa que los archivos funcionan directamente con Claude Code, Cursor, Codex y cualquier agente que lea el estandar.

## 🧭 AI Spec-Driven Development

Esta herramienta habilita una metodologia que llamamos **AI Spec-Driven Development (AI SDD)**: en lugar de ir directo de una idea al codigo, primero generas una capa de especificacion rica que fundamenta el trabajo de tu agente.

```
Tu idea → generate (contexto) → spec (especificaciones) → El agente escribe codigo con contexto completo
```

El comando `generate` crea el **plano arquitectonico** — que es el proyecto, como se construye, que patrones sigue. El comando `spec` toma ese plano y produce **especificaciones listas para implementar** — features, criterios de aceptacion, planes tecnicos y desglose de tareas.

Tu agente no improvisa. Implementa una spec. Esa es la diferencia.

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

### 🚀 Con Codify

```
Tu: "Crea una API de pagos en Go"

Agente: *lee AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md e IDIOMS.md*
Agente: "Veo que usas DDD con Clean Architecture, PostgreSQL,
         testing BDD con Godog, y patrones idiomaticos de Go.
         Creo el endpoint de pagos en internal/domain/payment/
         siguiendo tus patrones y convenciones de concurrencia."

Resultado: Codigo coherente desde la primera linea ✨
```

## ⚡ Quick Start

### Instalacion

```bash
# Homebrew (macOS/Linux — no requiere Go)
brew tap jorelcb/tap
brew install codify

# O via go install
go install github.com/jorelcb/codify/cmd/codify@latest

# O descarga binarios pre-compilados desde GitHub Releases
# https://github.com/jorelcb/codify/releases
```

### Tu primer contexto en 30 segundos

```bash
# 1. Configura tu API key (Claude o Gemini)
export ANTHROPIC_API_KEY="sk-ant-..."   # para Claude (default)
# o
export GEMINI_API_KEY="AI..."           # para Gemini

# 2. Describe tu proyecto (con language para guias idiomaticas)
codify generate payment-service \
  --description "Microservicio de pagos en Go con gRPC, PostgreSQL y Kafka. \
  DDD con Clean Architecture. Stripe como procesador." \
  --language go

# 3. Usa Gemini en vez de Claude
codify generate payment-service \
  --description "Microservicio de pagos en Go" \
  --model gemini-3.1-pro-preview

# 4. Listo. Archivos generados.
```

### Lo que vas a ver

```
🚀 Generating context for: payment-service
  Model: claude-sonnet-4-6
  Preset: default
  Language: go

  [1/5] Generating AGENTS.md... ✓
  [2/5] Generating CONTEXT.md... ✓
  [3/5] Generating INTERACTIONS_LOG.md... ✓
  [4/5] Generating DEVELOPMENT_GUIDE.md... ✓
  [5/5] Generating IDIOMS.md... ✓

📁 Output: output/payment-service/
  ├── AGENTS.md                → Root file (tech stack, comandos, convenciones)
  └── context/
      ├── CONTEXT.md           → Arquitectura y diseno tecnico
      ├── INTERACTIONS_LOG.md  → Bitacora de sesiones y ADRs
      ├── DEVELOPMENT_GUIDE.md → Metodologia, testing, seguridad
      └── IDIOMS.md            → Patrones idiomaticos (Go)

✅ Done! 5 files generated
   Total tokens: ~18,200
```

## 🔌 MCP Server

Usa Codify como **servidor MCP** — tu agente de IA invoca las herramientas directamente, sin necesidad de CLI manual.

### Instalacion

```bash
go install github.com/jorelcb/codify/cmd/codify@latest
```

### Claude Code

Agrega a `.mcp.json` en tu proyecto:

```json
{
  "mcpServers": {
    "codify": {
      "command": "codify",
      "args": ["serve"],
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-...",
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

### Codex CLI

```bash
# Registrar el servidor MCP
codex mcp add codify -- codify serve
```

### Gemini CLI

Agrega a `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "codify": {
      "command": "codify",
      "args": ["serve"],
      "env": {
        "GEMINI_API_KEY": "AI..."
      }
    }
  }
}
```

> Configura la(s) API key(s) del proveedor que quieras usar. El proveedor se auto-detecta segun el parametro `model`. Si el binario no esta en tu PATH, usa la ruta completa (e.g., `/Users/tu-usuario/go/bin/codify`).

### Herramientas MCP disponibles

#### Herramientas generativas (requieren API key de LLM)

| Herramienta | Descripcion |
|-------------|-------------|
| `generate_context` | Genera archivos de contexto a partir de una descripcion |
| `generate_specs` | Genera specs SDD a partir de contexto existente |
| `analyze_project` | Escanea un proyecto existente y genera contexto desde su estructura |
| `generate_skills` | Genera Agent Skills por categoria/preset — soporta modos `static` (instantaneo) y `personalized` (adaptado via LLM con `project_context`) |

Todas las herramientas generativas soportan `locale` (`en`/`es`) y `model`. `generate_context` y `analyze_project` tambien aceptan `with_specs` para encadenar generacion de specs. `generate_skills` acepta `mode` (`static`/`personalized`), `category`, `preset` y `project_context`.

#### Herramientas de conocimiento (sin API key)

| Herramienta | Descripcion |
|-------------|-------------|
| `commit_guidance` | Spec de Conventional Commits y contexto comportamental para generar mensajes de commit |
| `version_guidance` | Spec de Semantic Versioning y contexto comportamental para determinar bumps de version |

Las herramientas de conocimiento inyectan contexto comportamental en el agente que las invoca — de la misma forma que un agente de Claude Code lo haria. El agente recibe la spec e instrucciones, y las aplica a la tarea actual. Soportan `locale` (`en`/`es`).

### Prompts de ejemplo

```
"Genera contexto para un microservicio de pagos en Go con gRPC y PostgreSQL"
→ El agente invoca generate_context

"Analiza mi proyecto en /path/to/my-app y genera specs"
→ El agente invoca analyze_project con with_specs=true

"Genera skills de workflow para mi proyecto"
→ El agente invoca generate_skills con mode=static, category=workflow, preset=all

"Crea skills de DDD adaptadas a mi proyecto Go con Clean Architecture"
→ El agente invoca generate_skills con mode=personalized, project_context="Go con DDD..."

"Ayudame a hacer commit de estos cambios siguiendo conventional commits"
→ El agente invoca commit_guidance, recibe la spec, construye el mensaje

"Que version deberia liberar con los cambios recientes?"
→ El agente invoca version_guidance, recibe las reglas semver, analiza los commits
```

---

## 🎨 Features

### 📋 Comando `generate` — Contexto para tu agente

Genera archivos siguiendo el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md):

| Archivo | Que hace |
|---------|----------|
| `AGENTS.md` | Root file: tech stack, comandos, convenciones, estructura |
| `CONTEXT.md` | Arquitectura, componentes, flujo de datos, decisiones |
| `INTERACTIONS_LOG.md` | Bitacora de sesiones y ADRs |
| `DEVELOPMENT_GUIDE.md` | Metodologia de trabajo, testing, seguridad, expectativas de entrega |
| `IDIOMS.md` | Concurrencia, error handling, convenciones del lenguaje *(requiere `--language`)* |

Coloca estos archivos en la raiz de tu proyecto. Agentes compatibles (Claude Code, Cursor, Codex, etc.) los leen automaticamente.

### 📐 Comando `spec` — Especificaciones AI SDD

A partir de un contexto existente, genera especificaciones tecnicas listas para implementar:

```bash
codify spec payment-service \
  --from-context ./output/payment-service/
```

| Archivo | Que hace |
|---------|----------|
| `CONSTITUTION.md` | DNA del proyecto: stack, principios, restricciones |
| `SPEC.md` | Features con criterios de aceptacion |
| `PLAN.md` | Diseno tecnico y decisiones de arquitectura |
| `TASKS.md` | Desglose de tareas con dependencias y prioridad |

### 🔎 Comando `analyze` — Contexto desde proyectos existentes

Escanea un codebase existente y genera archivos de contexto automaticamente:

```bash
codify analyze /path/to/my-project --with-specs
```

Auto-detecta lenguaje, framework, dependencias, estructura de directorios, README, archivos de contexto existentes y señales de infraestructura (Docker, CI/CD, Makefile, etc.). Todo alimenta al LLM para una generacion mas rica y consciente del proyecto.

### ⚡ `--with-specs` — Pipeline completo en un comando

Disponible en `generate` y `analyze`. Encadena generacion de contexto + specs + actualizacion de AGENTS.md en una sola ejecucion:

```bash
codify generate my-api \
  --description "API REST en Go con PostgreSQL" \
  --language go \
  --with-specs
```

### 🧩 Comando `skills` — Agent Skills

Genera [Agent Skills](https://agentskills.io) reutilizables (SKILL.md) con seleccion guiada interactiva. Dos modos: **static** (instantaneo, sin API key) y **personalized** (adaptado a tu proyecto via LLM).

```bash
# Modo interactivo — seleccion guiada de categoria, preset, modo, target, etc.
codify skills

# Static: entrega instantanea desde el catalogo embebido (sin API key)
codify skills --category workflow --preset all --mode static

# Personalized: el LLM adapta las skills a tu proyecto especifico
codify skills --category architecture --preset clean --mode personalized \
  --context "Microservicio Go con DDD, Godog BDD, PostgreSQL"

# Skills de arquitectura para Codex
codify skills --category architecture --preset neutral --target codex
```

| Categoria | Preset | Skills |
|-----------|--------|--------|
| `architecture` | `clean` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `workflow` | `conventional-commit` | Conventional Commits |
| `workflow` | `semantic-versioning` | Semantic Versioning |
| `workflow` | `all` | Todas las skills de workflow |

**Modos:**
- **Static**: Entrega skills pre-construidas instantaneamente desde el catalogo embebido. Sin LLM, sin API key, sin costo. Incluye frontmatter YAML por ecosistema.
- **Personalized**: Usa LLM para adaptar las skills a tu proyecto — los ejemplos usan tu dominio, lenguaje y stack.

**Ecosistemas target**: `claude` (default), `codex`, `antigravity` — cada uno recibe frontmatter YAML especifico y ruta de salida (`.claude/skills/`, `.agents/skills/`).

### 🔍 Comando `list` — Proyectos generados

```bash
codify list
```

## 🌐 Guias por Lenguaje

Cuando pasas `--language`, la herramienta genera un archivo adicional `IDIOMS.md` con patrones y convenciones especificas de ese lenguaje. Este es uno de los features de mayor impacto — le da a tu agente conocimiento profundo de patrones idiomaticos en lugar de consejos genericos.

| Lenguaje | Que cubre IDIOMS.md |
|----------|---------------------|
| `go` | Goroutines, channels, WaitGroups, `context.Context`, error wrapping con `%w`, table-driven tests |
| `javascript` | async/await, `Promise.all`, `AbortController`, worker threads, TypeScript, ESM, patrones Jest |
| `python` | asyncio, multiprocessing, type hints, pydantic, fixtures pytest, `ruff` |

```bash
# Proyecto Go con guias idiomaticas
codify generate my-api -d "API REST en Go" --language go

# SDK TypeScript con idioms de JS
codify generate my-sdk -d "SDK en TypeScript" --language javascript

# Servicio Python con patrones async
codify generate my-service -d "Servicio con FastAPI" --language python
```

Sin `--language`, la herramienta genera 4 archivos. Con el flag, obtienes 5 — y un output significativamente mas rico.

## 🎭 Presets

Elige la filosofia de tus contextos:

### `--preset default` *(por defecto)*

Recomendado: **DDD + Clean Architecture + BDD**. Incluye:
- Separacion estricta de capas (Domain → Application → Infrastructure → Interfaces)
- Testing BDD con coverage targets (80% dominio, 70% aplicacion)
- Observabilidad con OpenTelemetry
- Inyeccion de dependencias obligatoria
- Restricciones DEBE/NO DEBE
- Metodologia de desarrollo y checklist de auto-validacion

```bash
codify generate my-api \
  --description "API REST de inventarios en Go"
# Usa preset default automaticamente
```

### `--preset neutral`

Sin postura arquitectonica. Deja que el LLM adapte la estructura al proyecto:

```bash
codify generate my-api \
  --description "API REST de inventarios en Go" \
  --preset neutral
```

### `--from-file` — Descripciones ricas desde archivos

Para descripciones detalladas (documentos de diseno, RFCs, 6-pagers), usa `--from-file` en lugar de `--description`:

```bash
codify generate my-api \
  --from-file ./docs/descripcion-proyecto.md \
  --language go
```

El contenido del archivo se convierte en la descripcion del proyecto. Soporta cualquier formato de texto — markdown, texto plano, etc. Mutuamente excluyente con `--description`.

## ⚙️ Opciones

```bash
codify generate <nombre> [flags]
```

| Flag | Corto | Descripcion | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Descripcion del proyecto *(requerido sin `--from-file`)* | — |
| `--from-file` | `-f` | Leer descripcion desde archivo *(alternativa a `-d`)* | — |
| `--preset` | `-p` | Preset de templates | `default` |
| `--model` | `-m` | Modelo LLM (`claude-*` o `gemini-*`) | `claude-sonnet-4-6` |
| `--language` | `-l` | Lenguaje (activa guias idiomaticas) | — |
| `--locale` | | Idioma de salida (`en`, `es`) | `en` |
| `--with-specs` | | Tambien genera specs SDD despues del contexto | `false` |
| `--type` | `-t` | Hint de tipo (api, cli, lib...) | — |
| `--architecture` | `-a` | Hint de arquitectura | — |

## 🏗️ Arquitectura

Construido en Go con lo que predica — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Logica de negocio pura
│   ├── project/         Entidad Project (aggregate root)
│   ├── shared/          Value objects, errores de dominio
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Casos de uso (CQRS)
│   ├── command/         GenerateContext, GenerateSpec, GenerateSkills
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementaciones
│   ├── llm/             Proveedores LLM (Claude, Gemini) + prompt builder
│   ├── template/        Template loader (locale + preset + language-aware)
│   ├── scanner/         Project scanner (deteccion de lenguaje, deps, framework)
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 Puntos de entrada
    ├── cli/commands/    generate, analyze, spec, skills, serve, list
    └── mcp/             Servidor MCP (transporte stdio + HTTP, 4 herramientas)
```

### Sistema de templates

```
templates/
├── en/                          Locale ingles
│   ├── default/                 Preset recomendado (DDD/Clean Architecture)
│   │   ├── agents.template
│   │   ├── context.template
│   │   ├── interactions.template
│   │   └── development_guide.template
│   ├── neutral/                 Preset generico (sin opiniones arquitectonicas)
│   │   └── (mismos archivos)
│   ├── spec/                    Templates de especificacion (AI SDD)
│   │   ├── constitution.template
│   │   ├── spec.template
│   │   ├── plan.template
│   │   └── tasks.template
│   ├── skills/                  Templates de Agent Skills (static + guias LLM)
│   │   ├── default/             Architecture: Clean (DDD, BDD, CQRS, Hexagonal)
│   │   ├── neutral/             Architecture: Neutral (review, testing, API)
│   │   └── workflow/            Workflow (conventional commits, semver)
│   └── languages/               Guias idiomaticas por lenguaje
│       ├── go/idioms.template
│       ├── javascript/idioms.template
│       └── python/idioms.template
└── es/                          Locale español (misma estructura)
```

La regla de oro: `Infrastructure → Application → Domain`. Nada en domain depende de nada externo.

Ver [context/CONTEXT.md](context/CONTEXT.md) para el detalle arquitectonico completo.

## 🧪 Tests

```bash
# Todos los tests
go test ./...

# BDD con Godog
go test ./tests/...
```

## 📊 Estado del proyecto

**v1.9.0** 🎉

✅ **Funcionando:**
- Soporte multi-proveedor LLM (Anthropic Claude + Google Gemini)
- Generacion de contextos con streaming
- Generacion de specs SDD a partir de contexto existente
- Agent Skills con seleccion guiada interactiva y modo dual (static/personalized)
- Categorias de skills (architecture, workflow) con catalogo declarativo
- Skills estaticas: entrega instantanea, sin API key, frontmatter por ecosistema
- Skills personalizadas: adaptadas via LLM al contexto del proyecto (dominio, lenguaje, stack)
- Servidor MCP (transporte stdio + HTTP) con 6 herramientas
- Herramientas de conocimiento MCP (commit_guidance, version_guidance) — sin API key
- Comando `analyze` — escanear proyectos existentes y generar contexto
- Flag `--with-specs` — pipeline completo en un comando
- Sistema de presets (architecture: clean/neutral, workflow: conventional-commit/semantic-versioning)
- Estandar AGENTS.md como root file
- Guias idiomaticas por lenguaje (Go, JavaScript, Python)
- Reglas de grounding anti-alucinacion en prompts
- CLI con Cobra + menus interactivos (charmbracelet/huh)
- Distribucion via Homebrew formula (macOS/Linux)

🚧 **Proximo:**
- Categoria de skills Testing (unit, integration, e2e)
- Tests de integracion end-to-end
- Retries y manejo de rate limits
- Autenticacion MCP server remoto (OAuth/BYOK)

## 💡 FAQ

**¿Que proveedores LLM soporta?**
Anthropic Claude (default) y Google Gemini. Configura `ANTHROPIC_API_KEY` para Claude o `GEMINI_API_KEY` para Gemini. El proveedor se auto-detecta por el flag `--model`: modelos `claude-*` usan Anthropic, modelos `gemini-*` usan Google.

**¿Cuanto cuesta cada generacion?**
4-5 llamadas API para `generate` (depende de `--language`), 4 para `spec`. Cada generacion cuesta centavos con cualquier proveedor.

**¿Los templates son fijos?**
Son guias estructurales, no output renderizable. El LLM genera contenido inteligente y especifico a tu proyecto siguiendo la estructura del template.

**¿Puedo personalizar los templates?**
Puedes crear tus propios presets en `templates/<locale>/`. Cada preset necesita 4 archivos: `agents.template`, `context.template`, `interactions.template` y `development_guide.template`. Templates por lenguaje van en `templates/<locale>/languages/<lang>/idioms.template`.

**¿Que agentes soportan los archivos generados?**
Cualquier agente compatible con el estandar [AGENTS.md](https://github.com/anthropics/AGENTS.md): Claude Code, Cursor, GitHub Copilot Workspace, Codex, y mas.

**¿Que es AI Spec-Driven Development?**
Una metodologia donde generas contexto y especificaciones *antes* de escribir codigo. Tu agente implementa una spec, no improvisa. `generate` crea el plano, `spec` crea el plan de implementacion.

## 📚 Documentacion

- [📋 AGENTS.md](AGENTS.md) — Contexto del proyecto para agentes de IA
- [🏛️ Arquitectura](context/CONTEXT.md) — Detalle DDD/Clean Architecture
- [📝 Changelog](CHANGELOG.md) — Historial de cambios
- [📐 Specs](specs/) — Especificaciones tecnicas (SDD)

## 📄 Licencia

Apache License 2.0 — ver [LICENSE](LICENSE).

---

<div align="center">

**Construido para potenciar el desarrollo asistido por IA** 🧠

*"Un agente sin contexto es un pasante con acceso root"*

⭐ Si te sirvio, dale una estrella — nos motiva a seguir construyendo

[🐛 Reportar bug](https://github.com/jorelcb/codify/issues) · [💡 Sugerir feature](https://github.com/jorelcb/codify/issues)

</div>