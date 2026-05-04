# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-1.19.0-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.23)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Contexto. Specs. Skills. Workflows. Todo lo que tu agente de IA necesita antes de escribir la primera linea de codigo.** 🏗️

*Porque un agente sin contexto es un pasante con acceso root.*

[English](README.md) | **[Español]**

[Quick Start](#-quick-start) · [Contexto](#-generacion-de-contexto) · [Specs](#-desarrollo-guiado-por-specs) · [Skills](#-agent-skills) · [Workflows](#-workflows) · [MCP Server](#-mcp-server) · [Guias por Lenguaje](#-guias-por-lenguaje) · [Arquitectura](#%EF%B8%8F-arquitectura)

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

**Codify** equipa a tu agente de IA con cuatro cosas que necesita para dejar de improvisar:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Contexto   │     │    Specs     │     │   Skills     │     │  Workflows   │
│              │     │              │     │              │     │              │
│  Que es el   │     │  Que         │     │  Como hacer  │     │  Recetas     │
│  proyecto    │────▶│  construir   │     │  las cosas   │     │  multi-paso  │
│              │     │              │     │  bien        │     │  bajo demanda│
│  generate    │     │  spec        │     │  skills      │     │  workflows   │
│  analyze     │     │  --with-specs│     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
     Memoria            Plan              Habilidades        Orquestacion
```

- **Contexto** le da al agente memoria arquitectonica — stack, patrones, convenciones, conocimiento de dominio
- **Specs** le dan al agente un plan de implementacion — features, criterios de aceptacion, desglose de tareas
- **Skills** le dan al agente habilidades reutilizables — como hacer commits, versionar, disenar entidades, hacer code review
- **Workflows** le dan al agente recetas de orquestacion — procesos multi-paso como desarrollo de features, correccion de bugs, releases

Sigue el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md) — una especificacion abierta respaldada por la Linux Foundation para proveer contexto de proyecto a agentes de IA. Los archivos funcionan directamente con Claude Code, Cursor, Codex y cualquier agente que lea el estandar.

## ✨ Antes y despues

### 😱 Sin Codify

```
Tu: "Crea una API de pagos en Go"

Agente: *crea main.go con todo en un archivo*
Tu: "No, usa Clean Architecture"
Agente: *crea estructura pero mezcla domain con infra*
Tu: "Los repositorios van en infrastructure"
Agente: *refactoriza por tercera vez*
Tu: "¿Y los tests BDD que pedi ayer?"
Agente: "¿Tests BDD? Es la primera vez que me lo mencionas"
Tu: "Al menos haz commit de esto correctamente"
Agente: *escribe "update code" como mensaje de commit*

Resultado: 45 minutos corrigiendo al agente 😤
```

### 🚀 Con Codify

```
Tu: "Crea una API de pagos en Go"

Agente: *lee AGENTS.md, CONTEXT.md, DEVELOPMENT_GUIDE.md*
Agente: "Veo que usas DDD con Clean Architecture, PostgreSQL,
         testing BDD con Godog, y patrones idiomaticos de Go.
         Creo el endpoint de pagos en internal/domain/payment/
         siguiendo tus patrones y convenciones de concurrencia."

Agente: *lee SKILL.md de conventional commits*
Agente: "Listo. Aqui esta el commit siguiendo Conventional Commits:
         feat(payment): add payment domain entity with Stripe integration"

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

### Cuatro formas de equipar a tu agente

Todos los comandos soportan **modo interactivo** — ejecuta sin flags y los menus te guian por cada opcion. O pasa los flags explicitamente para CI/scripting.

```bash
# 1. Configura tu API key (Claude o Gemini)
export ANTHROPIC_API_KEY="sk-ant-..."   # para Claude (default)
# o
export GEMINI_API_KEY="AI..."           # para Gemini

# ── Contexto: dale a tu agente memoria del proyecto ──
codify generate
# Menus interactivos para: nombre, descripcion, preset, lenguaje, modelo, locale, output, specs

# O pasa todos los flags explicitamente (zero prompts):
codify generate payment-service \
  --description "Microservicio de pagos en Go con gRPC, PostgreSQL y Kafka" \
  --language go

# ── Specs: dale a tu agente un plan de implementacion ──
codify spec payment-service \
  --from-context ./output/payment-service/

# ── Skills: dale a tu agente habilidades reutilizables ──
codify skills
# Menus interactivos para: categoria, preset, modo, target, ubicacion de instalacion
# No necesitas API key para el modo static.

# ── Workflows: dale a tu agente recetas de orquestacion ──
codify workflows
# Menus interactivos para: preset, target, modo, locale, ubicacion de instalacion
# Soporta Claude Code (native skills) y Antigravity (.md nativo).
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

---

## 📋 Generacion de Contexto

La base. Genera archivos siguiendo el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md) que le dan a tu agente memoria profunda del proyecto.

### Comando `generate` — Contexto desde una descripcion

```bash
codify generate payment-service \
  --description "Microservicio de pagos en Go con gRPC y PostgreSQL" \
  --language go
```

### Comando `analyze` — Contexto desde un proyecto existente

Escanea un codebase existente y genera archivos de contexto a partir de lo que encuentra. Usa un **prompt diferenciado** que trata los datos del scan como ground truth factual, produciendo output mas preciso que una descripcion manual.

**Lo que detecta el scanner:**
- Lenguaje, framework y dependencias (Go, JS/TS, Python, Rust, Java, Ruby)
- Estructura de directorios (3 niveles de profundidad)
- Contenido del README (filtrado: badges, comentarios HTML, ToC eliminados)
- Archivos de contexto existentes (18+ patrones: AGENTS.md, .claude/CLAUDE.md, ADRs, specs OpenAPI, etc.)
- Targets de build de Makefile/Taskfile (comandos exactos para AGENTS.md)
- Patrones de testing (frameworks, escenarios BDD, config de cobertura)
- Pipelines CI/CD (triggers y jobs de GitHub Actions, GitLab CI)
- Senales de infraestructura (Docker, Terraform, Kubernetes, Helm)

```bash
codify analyze /path/to/my-project
```

### Archivos generados

| Archivo | Que hace |
|---------|----------|
| `AGENTS.md` | Root file: tech stack, comandos, convenciones, estructura |
| `CONTEXT.md` | Arquitectura, componentes, flujo de datos, decisiones |
| `INTERACTIONS_LOG.md` | Bitacora de sesiones y ADRs |
| `DEVELOPMENT_GUIDE.md` | Metodologia de trabajo, testing, seguridad, expectativas de entrega |
| `IDIOMS.md` | Concurrencia, error handling, convenciones del lenguaje *(requiere `--language`)* |

Coloca estos archivos en la raiz de tu proyecto. Agentes compatibles (Claude Code, Cursor, Codex, etc.) los leen automaticamente.

### Opciones

```bash
codify generate [nombre-proyecto] [flags]
```

Todos los flags son opcionales en una terminal — los menus interactivos preguntan por los valores faltantes.

| Flag | Corto | Descripcion | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Descripcion del proyecto *(requerido sin `--from-file`)* | *(interactivo)* |
| `--from-file` | `-f` | Leer descripcion desde archivo *(alternativa a `-d`)* | — |
| `--preset` | `-p` | Preset de templates (`default`, `neutral`) | *(interactivo)* |
| `--model` | `-m` | Modelo LLM (`claude-*` o `gemini-*`) | auto-detectado |
| `--language` | `-l` | Lenguaje (activa guias idiomaticas) | — |
| `--locale` | | Idioma de salida (`en`, `es`) | `en` |
| `--with-specs` | | Tambien genera specs SDD despues del contexto | `false` |
| `--type` | `-t` | Hint de tipo (api, cli, lib...) | — |
| `--architecture` | `-a` | Hint de arquitectura | — |

---

## 📐 Desarrollo Guiado por Specs

A partir de contexto existente, genera especificaciones listas para implementar. Esto habilita **AI Spec-Driven Development (AI SDD)**: tu agente implementa una spec, no improvisa.

```
Tu idea → generate (contexto) → spec (especificaciones) → El agente escribe codigo con contexto completo
```

### Comando `spec`

```bash
codify spec payment-service \
  --from-context ./output/payment-service/
```

### `--with-specs` — Pipeline completo en un comando

Disponible en `generate` y `analyze`. Encadena generacion de contexto + specs + actualizacion de AGENTS.md en una sola ejecucion:

```bash
codify generate my-api \
  --description "API REST en Go con PostgreSQL" \
  --language go \
  --with-specs
```

### Archivos de spec generados

| Archivo | Que hace |
|---------|----------|
| `CONSTITUTION.md` | DNA del proyecto: stack, principios, restricciones |
| `SPEC.md` | Features con criterios de aceptacion |
| `PLAN.md` | Diseno tecnico y decisiones de arquitectura |
| `TASKS.md` | Desglose de tareas con dependencias y prioridad |

---

## 🧩 Agent Skills

Las skills son [Agent Skills](https://agentskills.io) reutilizables (archivos SKILL.md) que le ensenan a tu agente *como* ejecutar tareas especificas — seguir Conventional Commits, aplicar patrones DDD, hacer code reviews, versionar releases. Complementan los archivos de contexto: el contexto le dice al agente *que* es tu proyecto, las skills le dicen *como* hacer las cosas bien.

### Dos modos

| Modo | Que hace | API key | Costo | Velocidad |
|------|----------|---------|-------|-----------|
| **Static** | Entrega skills pre-construidas desde el catalogo embebido. Listas para produccion, con frontmatter por ecosistema. | No necesaria | Gratis | Instantaneo |
| **Personalized** | El LLM adapta las skills a tu proyecto — los ejemplos usan tu dominio, lenguaje y stack. | Requerida | ~centavos | ~10s |

### Modo interactivo

Solo ejecuta `codify skills` — el menu interactivo te guia por cada decision:

```bash
codify skills
# → Selecciona categoria (architecture, testing, conventions)
# → Selecciona preset (clean, neutral, conventional-commit, ...)
# → Selecciona modo (static o personalized)
# → Selecciona ecosistema target (claude, codex, antigravity)
# → Selecciona ubicacion de instalacion (global, project, o custom)
# → Selecciona locale
# → Si personalized: describe tu proyecto, elige modelo
```

### Modo CLI

```bash
# Static: entrega instantanea, sin API key
codify skills --category conventions --preset all --mode static

# Instalar globalmente — skills accesibles desde cualquier proyecto
codify skills --category conventions --preset all --mode static --install global

# Instalar en el proyecto actual — compartible via git
codify skills --category architecture --preset clean --mode static --install project

# Personalized: adaptado a tu proyecto via LLM
codify skills --category architecture --preset clean --mode personalized \
  --context "Microservicio Go con DDD, Godog BDD, PostgreSQL"

# Skills de arquitectura para ecosistema Codex
codify skills --category architecture --preset neutral --target codex
```

### Scopes de instalacion

| Scope | Path (Claude) | Path (Codex) | Uso |
|-------|---------------|--------------|-----|
| `global` | `~/.claude/skills/` | `~/.codex/skills/` | Accesible desde cualquier proyecto |
| `project` | `./.claude/skills/` | `./.agents/skills/` | Committed a git, compartido con el equipo |

### Catalogo de skills

| Categoria | Preset | Skills |
|-----------|--------|--------|
| `architecture` | `clean` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `testing` | `foundational` | Test Desiderata — Las 12 propiedades de Kent Beck para buenos tests |
| `testing` | `tdd` | Test-Driven Development — Red-Green-Refactor *(incluye foundational)* |
| `testing` | `bdd` | Behavior-Driven Development — Given/When/Then *(incluye foundational)* |
| `conventions` | `conventional-commit` | Conventional Commits |
| `conventions` | `semantic-versioning` | Semantic Versioning |
| `conventions` | `all` | Todas las skills de convenciones combinadas |

### Ecosistemas target

Cada ecosistema recibe frontmatter YAML especifico y rutas de salida:

| Target | Frontmatter | Ruta de salida |
|--------|-------------|----------------|
| `claude` *(default)* | `name`, `description`, `user-invocable: true` | `.claude/skills/` |
| `codex` | `name`, `description` | `.agents/skills/` |
| `antigravity` | `name`, `description`, `triggers` | `.agents/skills/` |

### Opciones

```bash
codify skills [flags]
```

| Flag | Descripcion | Default |
|------|-------------|---------|
| `--category` | Categoria de skill (`architecture`, `testing`, `conventions`) | *(interactivo)* |
| `--preset` | Preset dentro de la categoria | *(interactivo)* |
| `--mode` | Modo de generacion: `static` o `personalized` | *(interactivo)* |
| `--install` | Scope de instalacion: `global` (path del agente) o `project` (dir actual) | *(interactivo)* |
| `--context` | Descripcion del proyecto para modo personalized | — |
| `--target` | Ecosistema target (`claude`, `codex`, `antigravity`) | `claude` |
| `--model` `-m` | Modelo LLM (solo modo personalized) | auto-detectado |
| `--locale` | Idioma de salida (`en`, `es`) | `en` |
| `--output` `-o` | Directorio de salida (sobreescribe `--install`) | especifico del ecosistema |

---

## 🔄 Workflows

Los workflows son recetas de orquestacion multi-paso que los agentes de IA ejecutan bajo demanda. A diferencia de las skills (que ensenan *como* hacer una tarea especifica), los workflows orquestan *secuencias de tareas* — desde la creacion del branch hasta el merge del PR, desde el reporte del bug hasta el deploy del fix.

Codify genera workflows para dos ecosistemas:

| Target | Formato de salida | Ruta de salida | Invocacion |
|--------|-------------------|----------------|------------|
| **Claude Code** | Native skill (SKILL.md con frontmatter) | `.claude/skills/{workflow}/SKILL.md` | `/{skill-name}` |
| **Antigravity** | `.md` nativo con anotaciones de ejecucion (`// turbo`, `// capture`, etc.) | `.agent/workflows/{workflow}.md` | `/workflow-name` |

Cada skill de Claude incluye frontmatter YAML:
- `name` — Nombre del workflow
- `description` — Descripcion breve
- `disable-model-invocation: true` — Solo invocacion explicita del usuario
- `allowed-tools` — Herramientas permitidas para el workflow

### Dos modos

| Modo | Que hace | API key | Costo | Velocidad |
|------|----------|---------|-------|-----------|
| **Static** | Entrega workflows pre-construidos del catalogo embebido. Frontmatter por ecosistema. | No necesaria | Gratis | Instantaneo |
| **Personalized** | LLM adapta workflows a tu proyecto — los pasos referencian tus herramientas, CI/CD y targets de despliegue. | Requerida | ~centavos | ~10s |

### Modo interactivo

```bash
codify workflows
# → Selecciona preset (spec-driven-change, bug-fix, release-cycle, all)
# → Selecciona ecosistema target (claude, antigravity)
# → Selecciona modo (static o personalized)
# → Selecciona locale
# → Selecciona ubicacion de instalacion (global, project, o custom)
# → Si personalized: describe tu proyecto, elige modelo
```

### Modo CLI

```bash
# Claude Code: generar workflow skills nativos
codify workflows --preset all --target claude --mode static

# Claude Code: instalar skills globalmente
codify workflows --preset all --target claude --mode static --install global

# Claude Code: ciclo SDD spec-driven (propose → apply → archive)
codify workflows --preset spec-driven-change --target claude --mode static

# Antigravity: generar archivos de workflow nativos
codify workflows --preset all --target antigravity --mode static

# Antigravity: instalar globalmente
codify workflows --preset all --target antigravity --mode static --install global

# Personalized: skills adaptados a tu proyecto via LLM
codify workflows --preset all --target claude --mode personalized \
  --context "Microservicio Go con CI/CD via GitHub Actions"
```

### Ecosistemas target

| Target | Salida | Estructura | Diferencia clave |
|--------|--------|------------|------------------|
| `claude` | Native skill | `{workflow}/SKILL.md` con frontmatter YAML | Anotaciones eliminadas, instrucciones en prosa |
| `antigravity` *(default)* | Archivo `.md` plano | `{workflow}.md` con frontmatter YAML | Anotaciones nativas: `// turbo`, `// capture`, `// if`, `// parallel` |

### Scopes de instalacion

| Scope | Path Claude | Path Antigravity |
|-------|-------------|------------------|
| `global` | `~/.claude/skills/` | `~/.gemini/antigravity/global_workflows/` |
| `project` | `.claude/skills/` | `.agent/workflows/` |

### Catalogo de workflows

| Preset | Workflow | Descripcion |
|--------|----------|-------------|
| `spec-driven-change` | Cambio Spec-driven | Proponer → aplicar → archivar — ciclo SDD completo con deltas formales, branch creation y cleanup de merge |
| `bug-fix` | Bug Fix | Reproducir → diagnosticar → corregir → testear → PR |
| `release-cycle` | Release Cycle | Bump de version → changelog → tag → deploy |
| `all` | Todos los workflows | Todos los presets de workflow combinados |

### Spec-driven Change: la filosofia

`spec-driven-change` es el workflow recomendado para agregar features y hacer cambios no triviales. Implementa **Spec-Driven Development (SDD)**: una metodologia donde los artefactos formales de planeacion preceden al codigo, y donde cada cambio al sistema es una evolucion trackeable y revisable de las specifications — no solo un diff de codigo.

**El problema con desarrollo IA basado en chat:**
- Los planes desaparecen cuando termina la sesion de chat
- Los code reviews ven *que* cambio pero no *por que* cambio
- Los agentes IA pierden contexto entre sesiones y re-litigan decisiones
- Los specs (cuando existen) se desincronizan del codigo

**La respuesta SDD:**
- **Los specs viven en el repositorio**, organizados por capability bajo `openspec/specs/<capability>/spec.md`
- **Cada cambio es un workspace auto-contenido** bajo `openspec/changes/<change-id>/`
- **Los deltas (ADDED / MODIFIED / REMOVED requirements)** describen como evolucionan los specs, no solo el estado final
- **Los reviewers aprueban intencion primero** (proposal + deltas) antes de aprobar codigo
- **Los cambios archivados preservan audit trail** indefinidamente

#### Las tres fases

Cada fase es un modo cognitivo separado con un hand-off claro:

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  /spec-      │     │  /spec-      │     │  /spec-      │
│  propose     │ ──▶ │  apply       │ ──▶ │  archive     │
│              │     │              │     │              │
│  Planear el  │     │  Ejecutar el │     │  Consolidar  │
│  cambio      │     │  plan        │     │  & cleanup   │
└──────────────┘     └──────────────┘     └──────────────┘
   Intencion          Implementacion         Verdad
```

| Fase | Que produce | Modo cognitivo |
|------|-------------|----------------|
| **Propose** | `proposal.md` (motivacion), `design.md` (decisiones tecnicas), `tasks.md` (checklist atomico), `specs/<capability>/spec.md` (deltas con ADDED/MODIFIED/REMOVED) — ademas un branch de feature con la propuesta committeada | "Que debe cambiar y por que" — sin codigo todavia |
| **Apply** | Ejecucion secuencial de tareas, commits atomicos por tarea, tests, self-review, pull request | "Como hacerlo realidad" — enfocado en implementacion, deltas ya aprobados |
| **Archive** | Deltas mergeados a `openspec/specs/<capability>/spec.md`, cambio movido a `openspec/changes/archive/YYYY-MM-DD-<id>/`, branch mergeado y eliminado | "Hacer durable la verdad" — cerrar el ciclo |

#### Ejemplo concreto

```
$ /spec-propose Agregar autenticacion de dos factores via TOTP

  ✓ Lei openspec/specs/auth-login/spec.md
  ✓ Cree change-id: add-2fa
  ✓ Cree openspec/changes/add-2fa/
      ├── proposal.md       (motivacion, alcance, impacto)
      ├── design.md         (eleccion de libreria TOTP, cambios de schema)
      ├── tasks.md          (8 tareas atomicas en 3 fases)
      └── specs/auth-login/spec.md  (ADDED: requirements 2FA con scenarios G/W/T)
  ✓ Cree branch feature/add-2fa
  ✓ Committee artefactos de propuesta
  → Solicitar review de intencion antes de implementacion

$ /spec-apply add-2fa

  ✓ Implementando tarea 1.1: agregar columnas 2FA a tabla users
  ✓ Test: migracion up/down
  ✓ Commit: "feat: add 2FA schema columns"
  ... (8 tareas, commits atomicos)
  ✓ Test suite completo pasa
  ✓ PR abierto: "add-2fa: Agregar autenticacion 2FA via TOTP"

$ /spec-archive add-2fa

  ✓ Mergee deltas en openspec/specs/auth-login/spec.md
  ✓ Movi a openspec/changes/archive/2026-04-27-add-2fa/
  ✓ Squash-merge del branch feature
  ✓ Elimine local + remoto feature/add-2fa
```

#### Como encaja con el resto de Codify

```
codify generate ─────▶ AGENTS.md, CONTEXT.md       (memoria del proyecto)
codify spec ─────────▶ CONSTITUTION.md, SPEC.md... (specs iniciales)
codify workflows ────▶ /spec-propose, /spec-apply, /spec-archive
  --preset spec-                                   (skills de ciclo SDD)
  driven-change
```

`generate` y `spec` crean el **estado inicial**. El workflow `spec-driven-change` luego gobierna **cada cambio subsecuente**, manteniendo los specs del sistema en sincronia con su codigo.

#### Adopcion de SDD en un codebase existente

Para proyectos brownfield (codebases maduros sin specs formales), el path de adopcion es diferente — los specs deben emerger del comportamiento **real** del codigo, no de aspiraciones. Sigue esta secuencia:

```
1. codify analyze ./mi-proyecto          → AGENTS.md, CONTEXT.md, ... (contexto factual del scan)
2. openspec init                         → workspace openspec/ vacio
3. codify workflows                      → /spec-propose, /spec-apply, /spec-archive
     --preset spec-driven-change
     --target claude --install project
4. Desde tu agente, prompt:
   "Lee AGENTS.md y CONTEXT.md, despues haz ingenieria reversa de specs
    OpenSpec desde el codigo fuente bajo un change llamado 'baseline'.
    Identifica fronteras de capability desde la estructura del codebase.
    Usa requirements ADDED con scenarios GIVEN/WHEN/THEN derivados del
    comportamiento real, no del diseno aspiracional."
5. /spec-archive baseline                → consolida specs baseline en openspec/specs/
```

Este patron (el [retrofitting mode de OpenSpec](https://openspec.dev/)) produce specs **factuales** validados contra codigo existente en lugar de proyecciones desde una descripcion. Despues de archivar el baseline, cada cambio nuevo pasa por el ciclo estandar `/spec-propose → /spec-apply → /spec-archive`. El rol de Codify aqui es proveer el contexto (`analyze`) y los skills de ciclo (`workflows --preset spec-driven-change`); el retrofit del baseline en si es un prompt one-shot contra tu agente, no un comando separado de Codify — manteniendo responsabilidades limpias y evitando solapamiento con el tooling de OpenSpec.

#### Compatibilidad con OpenSpec

La estructura de salida (`openspec/specs/`, `openspec/changes/`, formato delta con ADDED/MODIFIED/REMOVED, scenarios GIVEN/WHEN/THEN) sigue la convencion de [OpenSpec](https://openspec.dev/). Los skills generados por Codify estan disenados para operar sin friccion sobre workspaces OpenSpec.

**Valor que agrega Codify sobre instalar OpenSpec directo:**
- **Personalizacion via LLM**: `--mode personalized --context "..."` adapta los skills a tu stack, herramientas y convenciones
- **Multi-target**: misma metodologia SDD entregada para Claude Code o Antigravity
- **Soporte de locale**: skills en ingles y espanol out of the box
- **Pipeline integrado**: combinado con `codify generate` + `codify spec`, obtienes bootstrap SDD end-to-end

### Skills vs Workflows

| | Skills | Workflows |
|-|--------|-----------|
| **Proposito** | Ensenan *como* hacer una tarea especifica | Orquestan una *secuencia* de tareas |
| **Alcance** | Responsabilidad unica (ej. "escribir un commit") | Proceso end-to-end (ej. "evolucionar un spec desde propuesta hasta cambio mergeado") |
| **Invocacion** | El agente lee cuando es relevante | El usuario invoca via `/command` |
| **Ejemplos** | Conventional Commits, DDD entity, code review | Ciclo de cambio spec-driven, bug fix, release cycle |

### Opciones

```bash
codify workflows [flags]
```

| Flag | Descripcion | Default |
|------|-------------|---------|
| `--preset` `-p` | Preset de workflow | *(interactivo)* |
| `--target` | Ecosistema target: `claude` o `antigravity` | `antigravity` |
| `--mode` | Modo de generacion: `static` o `personalized` | *(interactivo)* |
| `--install` | Scope de instalacion: `global` o `project` | *(interactivo)* |
| `--context` | Descripcion del proyecto para modo personalized | — |
| `--model` `-m` | Modelo LLM (solo modo personalized) | auto-detectado |
| `--locale` | Idioma de salida (`en`, `es`) | `en` |
| `--output` `-o` | Directorio de salida (sobreescribe `--install`) | especifico del target |

---

## 🪝 Hooks

Los hooks son **guardrails deterministicos** para Claude Code. Donde los skills (prompts) y los workflows (orquestacion) dependen de que el LLM haga lo correcto, los hooks son scripts shell que **siempre** se ejecutan en eventos del lifecycle (`PreToolUse`, `PostToolUse`, etc.) — hacen cumplir reglas en cada llamada, por exit code.

Las tres capas de artefactos se complementan:

| Capa | Mecanismo | Cuando corre? | Determinismo |
|---|---|---|---|
| **Skills** | Prompt cargado en contexto | Cuando agente o usuario lo invoca | Depende del LLM |
| **Workflows** | Lifecycle multi-skill | Usuario lo invoca via slash command | Depende del LLM |
| **Hooks** | Scripts shell en eventos | Cada llamada a tool que coincida | 100% (exit codes) |

### Catalogo de presets

| Preset | Evento | Proposito |
|---|---|---|
| `linting` | `PostToolUse` (Edit\|Write) | Auto-formatea y lintea archivos usando la herramienta correcta por lenguaje (Prettier/ESLint, ruff/black, gofmt/gofumpt, rustfmt, rubocop, shfmt). Detecta tools instalados via `command -v` — silencioso si falta uno. |
| `security-guardrails` | `PreToolUse` (Bash, Edit\|Write) | Bloquea comandos Bash peligrosos (`rm -rf /`, `git push --force` a main, `curl \| bash`, fork bombs, formateo de fs) y protege archivos sensibles (`.env*`, `secrets/`, `.git/`, lockfiles, claves privadas, configs CI). |
| `convention-enforcement` | `PreToolUse` (Bash con `if`) | Valida mensajes de commit contra Conventional Commits 1.0.0 (titulo ≤72 chars, tipo valido, sin placeholders triviales) y bloquea push directo/force-push a branches protegidos (`main`, `master`, `develop`, `production`, `release/*`). Requiere Claude Code v2.1.85+. |
| `all` | (combinado) | Los tres presets mergeados en un solo `hooks.json` |

### Estructura de salida

```
{output}/
├── hooks.json         ← bloque para mergear en tu settings.json
└── hooks/
    ├── lint.sh                          (preset linting)
    ├── block-dangerous-commands.sh      (security-guardrails)
    ├── protect-sensitive-files.sh       (security-guardrails)
    ├── validate-commit-message.sh       (convention-enforcement)
    └── check-protected-branches.sh      (convention-enforcement)
```

### Activar el bundle (merge manual)

Codify nunca modifica automaticamente tu `settings.json` — tu decides que mergear. Despues de correr el comando:

```bash
# 1. Mueve los scripts a tu directorio de Claude
cp -r ./codify-hooks/hooks/ ~/.claude/hooks/        # global (todos los proyectos)
# o
cp -r ./codify-hooks/hooks/ .claude/hooks/          # proyecto (commit al repo)

# 2. Abre ./codify-hooks/hooks.json y copia el objeto "hooks" dentro de:
#    ~/.claude/settings.json   (global)  o
#    .claude/settings.json     (proyecto)
# Mergealo como sibling de tus keys existentes (theme, model, mcpServers, etc).

# 3. Verifica activacion
claude
> /hooks
```

### Modo interactivo

```bash
codify hooks
# → Selecciona preset (linting, security-guardrails, convention-enforcement, all)
# → Selecciona locale (en, es)
# → Selecciona ubicacion de salida (proyecto / global / custom)
```

### Modo CLI

```bash
# Bundle de linting en ./codify-hooks/
codify hooks --preset linting

# Todos los hooks combinados, stderr en espanol
codify hooks --preset all --locale es

# Security guardrails en path custom
codify hooks --preset security-guardrails --output ./tmp/sec-hooks
```

### Requisitos

- **Bash** + **jq** (Linux/macOS nativo; Windows requiere Git Bash o WSL)
- **Claude Code v2.1.85+** (solo para el preset `convention-enforcement`, que usa el campo `if` en handlers)

### Limitaciones honestas

Los scripts bash usan patrones regex, no AST parsing. Detienen comandos **descuidados** del agente, no adversarios motivados — ofuscacion sofisticada (e.g. `eval $(echo b3JtIC1yZiAv | base64 -d)`) puede burlar la deteccion. Para garantias mas fuertes usa una herramienta dedicada como [bash-guardian](https://github.com/RoaringFerrum/claude-code-bash-guardian). Los scripts son cortos y deliberadamente editables: extiende los arrays de patrones para tu modelo de riesgo especifico.

### Opciones

```bash
codify hooks [flags]
```

| Flag | Descripcion | Default |
|---|---|---|
| `--preset` `-p` | `linting`, `security-guardrails`, `convention-enforcement`, o `all` | *(interactivo)* |
| `--locale` | Idioma de salida para stderr (`en` o `es`) | `en` |
| `--install` | Scope de instalacion: `global` o `project` | *(interactivo)* |
| `--output` `-o` | Directorio de salida | `./codify-hooks` |

---

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
| `generate_skills` | Genera Agent Skills — soporta modos `static` (instantaneo) y `personalized` (adaptado via LLM) |
| `generate_workflows` | Genera workflows para Claude Code (native skills) o Antigravity (.md nativo) — soporta modos `static` y `personalized` |
| `generate_hooks` | Genera bundles de hooks para Claude Code (guardrails deterministicos). Static-only, Claude-only. Produce `hooks.json` + scripts `.sh` para merge manual al `settings.json` |

Todas las herramientas generativas soportan `locale` (`en`/`es`) y `model`. `generate_context` y `analyze_project` tambien aceptan `with_specs`. `generate_skills` acepta `mode`, `category`, `preset`, `target` y `project_context`. `generate_workflows` acepta `mode`, `preset`, `target` (`claude`/`antigravity`) y `project_context`. `generate_hooks` acepta `preset` (`linting`/`security-guardrails`/`convention-enforcement`/`all`), `locale` y `output` — sin model ni context (static-only).

#### Herramientas de conocimiento (sin API key)

| Herramienta | Descripcion |
|-------------|-------------|
| `commit_guidance` | Spec de Conventional Commits y contexto comportamental para generar mensajes de commit |
| `version_guidance` | Spec de Semantic Versioning y contexto comportamental para determinar bumps de version |

Las herramientas de conocimiento inyectan contexto comportamental en el agente que las invoca — el agente recibe la spec e instrucciones, y las aplica a la tarea actual. Soportan `locale` (`en`/`es`).

### Prompts de ejemplo

```
"Genera contexto para un microservicio de pagos en Go con gRPC y PostgreSQL"
→ El agente invoca generate_context

"Analiza mi proyecto en /path/to/my-app y genera specs"
→ El agente invoca analyze_project con with_specs=true

"Genera skills de convenciones para mi proyecto"
→ El agente invoca generate_skills con mode=static, category=conventions, preset=all

"Crea skills de DDD adaptadas a mi proyecto Go con Clean Architecture"
→ El agente invoca generate_skills con mode=personalized, project_context="Go con DDD..."

"Genera workflow de spec-driven-change para Claude Code"
→ El agente invoca generate_workflows con target=claude, preset=spec-driven-change, mode=static

"Genera todos los workflows adaptados a mi proyecto Go con GitHub Actions"
→ El agente invoca generate_workflows con target=claude, mode=personalized, preset=all, project_context="Go con GitHub Actions"

"Genera hooks para Claude Code que bloqueen comandos peligrosos y validen conventional commits"
→ El agente invoca generate_hooks con preset=all (o security-guardrails + convention-enforcement)

"Ayudame a hacer commit de estos cambios siguiendo conventional commits"
→ El agente invoca commit_guidance, recibe la spec, construye el mensaje

"Que version deberia liberar con los cambios recientes?"
→ El agente invoca version_guidance, recibe las reglas semver, analiza los commits
```

---

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

## 🏗️ Arquitectura

Construido en Go con lo que predica — DDD/Clean Architecture:

```
internal/
├── domain/              💎 Logica de negocio pura
│   ├── project/         Entidad Project (aggregate root)
│   ├── catalog/         Catalogos declarativos de skills + workflows y registros de metadata
│   ├── shared/          Value objects, errores de dominio
│   └── service/         Interfaces: LLMProvider, FileWriter, TemplateLoader
│
├── application/         🔄 Casos de uso (CQRS)
│   ├── command/         GenerateContext, GenerateSpec, GenerateSkills, GenerateWorkflows
│   └── query/           ListProjects
│
├── infrastructure/      🔧 Implementaciones
│   ├── llm/             Proveedores LLM (Claude, Gemini) + prompt builder
│   ├── template/        Template loader (locale + preset + language-aware)
│   ├── scanner/         Project scanner (lenguaje, deps, framework, build targets, testing, CI/CD)
│   └── filesystem/      File writer, directory manager, context reader
│
└── interfaces/          🎯 Puntos de entrada
    ├── cli/commands/    generate, analyze, spec, skills, workflows, serve, list
    └── mcp/             Servidor MCP (transporte stdio + HTTP, 8 herramientas)
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
│   │   ├── testing/             Testing: Foundational, TDD, BDD
│   │   └── conventions/         Conventions (conventional commits, semver)
│   ├── workflows/              Templates de workflows
│   │   ├── bug_fix.template
│   │   ├── release_cycle.template
│   │   ├── spec_propose.template
│   │   ├── spec_apply.template
│   │   └── spec_archive.template
│   └── languages/               Guias idiomaticas por lenguaje
│       ├── go/idioms.template
│       ├── javascript/idioms.template
│       └── python/idioms.template
└── es/                          Locale espanol (misma estructura)
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

**v1.19.0** 🎉

✅ **Funcionando:**
- Soporte multi-proveedor LLM (Anthropic Claude + Google Gemini)
- **Generacion de contextos** con streaming (`generate`, `analyze`)
- **Analyze mejorado** — prompt diferenciado (factual vs aspiracional), scanner enriquecido con 18+ archivos de contexto, parseo de build targets, deteccion de patrones de testing, resumen de pipelines CI/CD, filtrado inteligente de README
- **Generacion de specs SDD** a partir de contexto existente (`spec`, `--with-specs`)
- **Agent Skills** con modo dual (static/personalized), seleccion guiada interactiva y catalogo declarativo
- **Instalacion de skills** — `--install global` o `--install project` para instalacion directa en el path del agente
- Categorias de skills (architecture, testing, conventions) con frontmatter por ecosistema (Claude, Codex, Antigravity)
- **Workflows** — recetas de orquestacion multi-paso para Claude Code (native skills) y Antigravity (anotaciones nativas)
- **Presets de workflows** — spec-driven-change (propose/apply/archive), bug-fix, release-cycle (modos static + personalized, multi-target)
- **UX interactiva unificada** — todos los comandos preguntan por parametros faltantes en terminal
- Servidor MCP (transporte stdio + HTTP) con 8 herramientas
- Herramientas de conocimiento MCP (commit_guidance, version_guidance) — sin API key
- Sistema de presets (default: DDD/Clean, neutral: generico)
- Estandar AGENTS.md como root file
- Guias idiomaticas por lenguaje (Go, JavaScript, Python)
- Parseo de dependencias para 8 lenguajes (Go, JS/TS, Python, Rust, Java, Ruby, PHP, C#)
- Reglas de grounding anti-alucinacion en prompts
- CLI con Cobra + menus interactivos (charmbracelet/huh)
- Distribucion via Homebrew formula (macOS/Linux)

🚧 **Proximo:**
- Tests de integracion end-to-end
- Retries y manejo de rate limits
- Autenticacion MCP server remoto (OAuth/BYOK)

## 💡 FAQ

**¿Que proveedores LLM soporta?**
Anthropic Claude (default) y Google Gemini. Configura `ANTHROPIC_API_KEY` para Claude o `GEMINI_API_KEY` para Gemini. El proveedor se auto-detecta por el flag `--model`: modelos `claude-*` usan Anthropic, modelos `gemini-*` usan Google.

**¿Cuanto cuesta cada generacion?**
4-5 llamadas API para `generate` (depende de `--language`), 4 para `spec`. Skills en modo static son gratis (sin llamadas API). Skills personalizadas usan 1 llamada API por skill. Cada generacion cuesta centavos con cualquier proveedor.

**¿Necesito API key para skills?**
Solo para el modo personalized. El modo static entrega skills pre-construidas instantaneamente desde el catalogo embebido — sin LLM, sin API key, sin costo.

**¿Cual es la diferencia entre skills static y personalized?**
Las skills static son mejores practicas genericas listas para produccion, entregadas al instante. Las skills personalized usan un LLM para adaptar ejemplos, naming y patrones al contexto especifico de tu proyecto (lenguaje, dominio, stack).

**¿Los templates son fijos?**
Son guias estructurales, no output renderizable. El LLM genera contenido inteligente y especifico a tu proyecto siguiendo la estructura del template.

**¿Puedo personalizar los templates?**
Puedes crear tus propios presets en `templates/<locale>/`. Cada preset necesita 4 archivos: `agents.template`, `context.template`, `interactions.template` y `development_guide.template`. Templates por lenguaje van en `templates/<locale>/languages/<lang>/idioms.template`.

**¿Que agentes soportan los archivos generados?**
Cualquier agente compatible con el estandar [AGENTS.md](https://github.com/anthropics/AGENTS.md): Claude Code, Cursor, GitHub Copilot Workspace, Codex, y mas.

**¿Cual es la diferencia entre Skills y Workflows?**
Las skills le ensenan a tu agente *como* hacer una tarea individual (ej. escribir un mensaje de commit, disenar una entidad DDD). Los workflows orquestan una *secuencia* de tareas en un proceso end-to-end (ej. el ciclo completo de desarrollo de una feature, desde el branch hasta el merge del PR). Las skills son pasivas (se leen cuando son relevantes), los workflows son activos (se invocan via `/command`).

**¿Necesito API key para workflows?**
Solo para el modo personalized. El modo static entrega workflows pre-construidos al instante — sin LLM, sin API key, sin costo.

**¿Para que ecosistemas funcionan los workflows?**
Claude Code (`--target claude`) y Antigravity (`--target antigravity`). Los workflows de Claude generan native skills (SKILL.md con frontmatter) que el agente ejecuta via `/skill-name`. Los workflows de Antigravity producen archivos `.md` nativos con anotaciones de ejecucion (`// turbo`, `// capture`, etc.).

**¿Que es AI Spec-Driven Development?**
Una metodologia donde generas contexto y especificaciones *antes* de escribir codigo. Tu agente implementa una spec, no improvisa. `generate` crea el plano, `spec` crea el plan de implementacion, y el workflow `spec-driven-change` gobierna cada cambio subsecuente como una evolucion trackeada del spec (propose → apply → archive) con deltas formales, workspaces de cambio aislados, y audit trails.

**¿Por que tres fases (propose / apply / archive) en lugar de un solo workflow?**
Cada fase es un modo cognitivo distinto. *Propose* responde "¿que debe cambiar y por que?" sin escribir codigo — el LLM se mantiene enfocado en intencion. *Apply* responde "¿como hacerlo realidad?" con los deltas ya aprobados, eliminando ambiguedad de spec del contexto de implementacion. *Archive* cierra el ciclo deterministicamente: mergea deltas a specs fuente-de-verdad, archiva el cambio para auditoria, mergea el branch. Mezclar estas fases diluye atencion y produce planes vagos + codigo descuidado.

**¿Codify reemplaza a OpenSpec?**
No — lo complementa. El preset `spec-driven-change` genera skills que operan sobre workspaces formato OpenSpec (`openspec/specs/`, `openspec/changes/`, deltas ADDED/MODIFIED/REMOVED con scenarios G/W/T). Si ya usas OpenSpec, Codify te da skills de ciclo personalizadas via LLM adaptadas a tu stack. Si no, Codify es tu punto de entrada zero-config a la metodologia — combinado con `codify generate` y `codify spec`, obtienes el pipeline completo desde repo en blanco hasta iteracion gobernada.

## 📚 Documentacion

- [📋 AGENTS.md](AGENTS.md) — Contexto del proyecto para agentes de IA
- [🏛️ Arquitectura](context/CONTEXT.md) — Detalle DDD/Clean Architecture
- [📝 Changelog](CHANGELOG.md) — Historial de cambios
- [📐 Specs](specs/) — Especificaciones tecnicas (SDD)

## 📄 Licencia

Apache License 2.0 — ver [LICENSE](LICENSE).

---

<div align="center">

**Contexto. Specs. Skills. Workflows. Tu agente, completamente equipado.** 🧠

*"Un agente sin contexto es un pasante con acceso root"*

⭐ Si te sirvio, dale una estrella — nos motiva a seguir construyendo

[🐛 Reportar bug](https://github.com/jorelcb/codify/issues) · [💡 Sugerir feature](https://github.com/jorelcb/codify/issues)

</div>