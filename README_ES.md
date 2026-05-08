# 🧠 Codify

<div align="center">

[![Version](https://img.shields.io/badge/version-2.1.1-blue?style=for-the-badge)](https://github.com/jorelcb/codify/releases)
[![MCP](https://img.shields.io/badge/MCP-Server-ff6b35?style=for-the-badge)](https://modelcontextprotocol.io)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.23)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=for-the-badge)](LICENSE)
[![Claude](https://img.shields.io/badge/Claude-cc785c?style=for-the-badge)](https://www.anthropic.com)
[![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev)
[![AGENTS.md](https://img.shields.io/badge/Standard-AGENTS.md-purple?style=for-the-badge)](https://github.com/anthropics/AGENTS.md)

**Genera, audita y evoluciona el contexto de tu agente de IA a lo largo del lifecycle del proyecto.** 🏗️

*Porque un agente sin contexto es un pasante con acceso root — y un contexto desactualizado es un pasante leyendo docs de hace tres semanas.*

[English](README.md) | **[Español]**

**Lifecycle:** [🚀 Bootstrap](#-bootstrap-setup-unico) · [🧰 Equip](#-equip-instala-contexto-skills-workflows-hooks-specs) · [🔧 Maintain](#-maintain-lifecycle-continuo)

**Atajos:** [Quick Start](#-quick-start) · [Fases de un vistazo](#%EF%B8%8F-fases-del-lifecycle-de-un-vistazo) · [MCP Server](#-mcp-server) · [Guias por Lenguaje](#-guias-por-lenguaje) · [Arquitectura](#%EF%B8%8F-arquitectura) · [Migrando desde v1.x](#-migrando-desde-v1x) · [FAQ](#-faq) · [Solucion de Problemas](#-solucion-de-problemas)

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

**Codify** equipa a tu agente de IA con seis capas que necesita para dejar de improvisar:

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

┌─────────────────────────────────┐  ┌─────────────────────────────────────┐
│           Hooks                 │  │           Lifecycle                 │
│                                 │  │                                     │
│   Guardrails deterministicos    │  │   Mantener artefactos en el tiempo  │
│   en eventos de tool calls      │  │                                     │
│                                 │  │   config / init                     │
│   hooks                         │  │   check / update / audit / usage    │
└─────────────────────────────────┘  └─────────────────────────────────────┘
       Determinismo                              Custodia
```

- **Contexto** le da al agente memoria arquitectonica — stack, patrones, convenciones, conocimiento de dominio
- **Specs** le dan al agente un plan de implementacion — features, criterios de aceptacion, desglose de tareas
- **Skills** le dan al agente habilidades reutilizables — como hacer commits, versionar, disenar entidades, hacer code review
- **Workflows** le dan al agente recetas de orquestacion — procesos multi-paso como desarrollo de features, correccion de bugs, releases
- **Hooks** agregan guardrails deterministicos — shell scripts en eventos de Claude Code, sin LLM en el loop
- **Lifecycle** mantiene todo en sincronia — `config`, `init`, `check`, `update`, `audit`, `usage`, `watch` — drift detection, regen selectivo, audit de commits, transparencia de costos, watcher foreground

Sigue el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md) — una especificacion abierta respaldada por la Linux Foundation para proveer contexto de proyecto a agentes de IA. Los archivos funcionan directamente con Claude Code, Cursor, Codex y cualquier agente que lea el estandar.

## 🗺️ Fases del lifecycle de un vistazo

Codify agrupa sus comandos en tres fases que reflejan como un developer adopta y usa la herramienta:

```
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Bootstrap  │ ──▶ │    Equip    │ ──▶ │  Maintain   │
    └─────────────┘     └─────────────┘     └─────────────┘
       config              generate            check
       init                analyze             update
                           spec                audit
                           skills              watch
                           workflows           usage
                           hooks               resolve
```

- **Bootstrap (setup unico)** — configura la workstation (`codify config`) o un proyecto (`codify init`).
- **Equip (segun necesites)** — genera contexto, instala skills/workflows/hooks, escribe specs.
- **Maintain (continuo)** — detecta drift, regenera, audita commits, registra uso.

El mismo diagrama esta en `codify --help`. Para la matriz completa de **workstation vs proyecto** y **greenfield vs brownfield** ver [`docs/lifecycle-matrix.md`](docs/lifecycle-matrix.md) (en ingles).

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

Secuencia recomendada para el primer uso. Cada paso mapea a una fase del lifecycle ([ver diagrama arriba](#%EF%B8%8F-fases-del-lifecycle-de-un-vistazo)). Para el tour completo con outputs esperados ver [`docs/getting-started.md`](docs/getting-started.md) (en ingles).

```bash
# 1. Instalar (Homebrew / go install / GitHub Releases)
brew tap jorelcb/tap && brew install codify

# 2. Bootstrap — defaults de workstation (una vez)
codify config

# 3. Bootstrap — tu proyecto (greenfield o brownfield)
cd my-project && codify init

# 4. Equip — instala solo lo que necesites (cada uno es opcional)
codify spec <name> --from-context ./output/<name>/    # SDD specs
codify skills                                         # Skills reutilizables
codify workflows                                      # Recetas multi-paso
codify hooks                                          # Guardrails de Claude Code

# 5. Maintain — mantiene los artefactos honestos a medida que el codigo evoluciona
codify check    # Drift detection — sin LLM, cero costo
codify update   # Regen selectivo cuando los signals cambian
codify audit    # Revisa commits contra convenciones
codify watch    # Watcher foreground
codify usage    # Resumen de tokens y costo LLM
codify resolve  # Llena marcadores [DEFINE: ...]
```

API keys (`ANTHROPIC_API_KEY` o `GEMINI_API_KEY`) solo se requieren para comandos LLM-backed — ver tabla en [`docs/getting-started.md`](docs/getting-started.md#api-keys).

El auto-launch de `codify config` la primera vez es soft (solo TTY, nunca bloquea CI). Opt-out via `--no-auto-config`, `CODIFY_NO_AUTO_CONFIG=1`, o `~/.codify/.no-auto-config`.

---

## 🚀 Bootstrap (setup unico)

> **Setup de una vez** para la workstation y por proyecto. Despues paseas a la [fase Equip](#-equip-instala-contexto-skills-workflows-hooks-specs).

### ⚙️ Configuracion y Bootstrap

Dos comandos complementarios definen como Codify se comporta: **`codify config`** a nivel usuario y **`codify init`** a nivel proyecto. Ambos componen sobre los comandos standalone existentes; son smart entry points, no reemplazos.

#### `codify config` — defaults a nivel usuario

`codify config` gestiona tus preferencias globales en `~/.codify/config.yml`. La primera vez que corres cualquier comando interactivo de Codify en TTY sin que ese archivo exista, se te ofrece lanzar el wizard. Tres respuestas: Yes (correr wizard), No (usar defaults esta vez), Skip permanently (crea `~/.codify/.no-auto-config` para que el prompt no aparezca mas).

| Subcomando | Accion |
|---|---|
| `codify config` | Wizard si no existe config; imprime config actual si existe |
| `codify config get <key>` | Lee un valor |
| `codify config set <key> <value>` | Actualiza un valor |
| `codify config unset <key>` | Limpia un valor |
| `codify config edit` | Abre `~/.codify/config.yml` en `$EDITOR` |
| `codify config list` | Imprime el config efectivo (con merge aplicado) |

Keys validos: `preset`, `locale`, `language`, `model`, `target`, `provider`, `project_name`.

#### `codify init` — bootstrap a nivel proyecto

`codify init` pregunta primero: ¿proyecto nuevo o existente? Segun la respuesta enruta al flujo correcto:

| Respuesta | Flujo interno | Que provees |
|---|---|---|
| **new** | invoca `generate` | nombre + descripcion (inline o ruta a archivo) |
| **existing** | invoca `analyze` | nombre (auto-detectado del cwd, override si querés) |

Despues, ambas ramas recolectan: preset arquitectonico (override del default global), lenguaje, locale, output dir, modelo. Resultado:

- `.codify/config.yml` — defaults a nivel proyecto que persisten para todo el equipo via repo
- `.codify/state.json` — snapshot del estado de generacion (consumido por lifecycle commands)
- `AGENTS.md` y `context/*.md` generados a `output/`

Skills, workflows y hooks NO se incluyen — `init` imprime los comandos recomendados para mantener responsabilidades enfocadas. Corre `codify skills`, `codify workflows`, `codify hooks` por separado cuando los necesites.

#### Precedencia de merge

Cuando cualquier comando resuelve un valor (preset, locale, model, etc.):

```
flags > .codify/config.yml > ~/.codify/config.yml > built-in defaults
```

Setear `--preset hexagonal` en linea de comandos gana sin importar que digan los archivos config. Project-level gana sobre user-level. Built-ins llenan los gaps.

---

## 🧰 Equip (instala contexto, skills, workflows, hooks, specs)

> **Instala solo lo que necesites.** Cada comando abajo es independiente y opcional. Despues de equipar, la [fase Maintain](#-maintain-lifecycle-continuo) mantiene todo honesto a medida que el proyecto evoluciona.

### 📋 Generacion de Contexto

La base. Genera archivos siguiendo el [estandar AGENTS.md](https://github.com/anthropics/AGENTS.md) que le dan a tu agente memoria profunda del proyecto.

#### Cuando usar `generate` vs `analyze`

| Situacion | Usar | Por que |
|---|---|---|
| Proyecto greenfield (sin codigo aun) | `codify generate` | Tu provees la descripcion; el LLM genera contexto contra ella |
| Repo existente con codigo dentro | `codify analyze` | El scanner extrae senales factuales (deps, build targets, CI, frameworks) y los alimenta como ground truth — mucho mas preciso que una descripcion manual |
| Repo existente + quieres sobreescribir lo que el scanner detecta | `codify analyze` primero, luego editar, luego `codify reset-state` | Scan-first, hand-tune segundo |
| Tienes un design doc detallado | `codify generate --from-file ./docs/design.md` | Trata el contenido del archivo como la descripcion |
| En duda | `codify init` | Pregunta "¿nuevo o existente?" y te enruta al flow correcto internamente |

#### Comando `generate` — Contexto desde una descripcion

```bash
codify generate payment-service \
  --description "Microservicio de pagos en Go con gRPC y PostgreSQL" \
  --language go
```

#### Comando `analyze` — Contexto desde un proyecto existente

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

#### Archivos generados

| Archivo | Que hace |
|---------|----------|
| `AGENTS.md` | Root file: tech stack, comandos, convenciones, estructura |
| `CONTEXT.md` | Arquitectura, componentes, flujo de datos, decisiones |
| `INTERACTIONS_LOG.md` | Bitacora de sesiones y ADRs |
| `DEVELOPMENT_GUIDE.md` | Metodologia de trabajo, testing, seguridad, expectativas de entrega |
| `IDIOMS.md` | Concurrencia, error handling, convenciones del lenguaje *(requiere `--language`)* |

Coloca estos archivos en la raiz de tu proyecto. Agentes compatibles (Claude Code, Cursor, Codex, etc.) los leen automaticamente.

#### Opciones

```bash
codify generate [nombre-proyecto] [flags]
```

Todos los flags son opcionales en una terminal — los menus interactivos preguntan por los valores faltantes.

| Flag | Corto | Descripcion | Default |
|------|-------|-------------|---------|
| `--description` | `-d` | Descripcion del proyecto *(requerido sin `--from-file`)* | *(interactivo)* |
| `--from-file` | `-f` | Leer descripcion desde archivo *(alternativa a `-d`)* | — |
| `--preset` | `-p` | Preset de templates (`neutral`, `clean-ddd`, `hexagonal`, `event-driven`) | *(interactivo)* |
| `--model` | `-m` | Modelo LLM (`claude-*` o `gemini-*`) | auto-detectado |
| `--language` | `-l` | Lenguaje (activa guias idiomaticas) | — |
| `--locale` | | Idioma de salida (`en`, `es`) | `en` |
| `--with-specs` | | Tambien genera specs SDD despues del contexto | `false` |
| `--type` | `-t` | Hint de tipo (api, cli, lib...) | — |
| `--architecture` | `-a` | Hint de arquitectura | — |

---

### 📐 Desarrollo Guiado por Specs

A partir de contexto existente, genera especificaciones listas para implementar. Esto habilita **AI Spec-Driven Development (AI SDD)**: tu agente implementa una spec, no improvisa.

```
Tu idea → generate (contexto) → spec (especificaciones) → El agente escribe codigo con contexto completo
```

#### Comando `spec`

```bash
codify spec payment-service \
  --from-context ./output/payment-service/
```

#### `--with-specs` — Pipeline completo en un comando

Disponible en `generate` y `analyze`. Encadena generacion de contexto + specs + actualizacion de AGENTS.md en una sola ejecucion:

```bash
codify generate my-api \
  --description "API REST en Go con PostgreSQL" \
  --language go \
  --with-specs
```

#### Archivos de spec generados

| Archivo | Que hace |
|---------|----------|
| `CONSTITUTION.md` | DNA del proyecto: stack, principios, restricciones |
| `SPEC.md` | Features con criterios de aceptacion |
| `PLAN.md` | Diseno tecnico y decisiones de arquitectura |
| `TASKS.md` | Desglose de tareas con dependencias y prioridad |

---

### 🧩 Agent Skills

Las skills son [Agent Skills](https://agentskills.io) reutilizables (archivos SKILL.md) que le ensenan a tu agente *como* ejecutar tareas especificas — seguir Conventional Commits, aplicar patrones DDD, hacer code reviews, versionar releases. Complementan los archivos de contexto: el contexto le dice al agente *que* es tu proyecto, las skills le dicen *como* hacer las cosas bien.

#### Dos modos

| Modo | Que hace | API key | Costo | Velocidad |
|------|----------|---------|-------|-----------|
| **Static** | Entrega skills pre-construidas desde el catalogo embebido. Listas para produccion, con frontmatter por ecosistema. | No necesaria | Gratis | Instantaneo |
| **Personalized** | El LLM adapta las skills a tu proyecto — los ejemplos usan tu dominio, lenguaje y stack. | Requerida | ~centavos | ~10s |

#### Modo interactivo

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

#### Modo CLI

```bash
# Static: entrega instantanea, sin API key
codify skills --category conventions --preset all --mode static

# Instalar globalmente — skills accesibles desde cualquier proyecto
codify skills --category conventions --preset all --mode static --install global

# Instalar en el proyecto actual — compartible via git
codify skills --category architecture --preset clean-ddd --mode static --install project

# Personalized: adaptado a tu proyecto via LLM
codify skills --category architecture --preset clean-ddd --mode personalized \
  --context "Microservicio Go con DDD, Godog BDD, PostgreSQL"

# Skills de arquitectura para ecosistema Codex
codify skills --category architecture --preset neutral --target codex
```

#### Scopes de instalacion

| Scope | Path (Claude) | Path (Codex) | Uso |
|-------|---------------|--------------|-----|
| `global` | `~/.claude/skills/` | `~/.codex/skills/` | Accesible desde cualquier proyecto |
| `project` | `./.claude/skills/` | `./.agents/skills/` | Committed a git, compartido con el equipo |

#### Catalogo de skills

| Categoria | Preset | Skills |
|-----------|--------|--------|
| `architecture` | `neutral` | Code review, test strategy, safe refactoring, API design |
| `architecture` | `clean-ddd` | DDD entity, Clean Architecture layer, BDD scenario, CQRS command, Hexagonal port |
| `architecture` | `hexagonal` | Port definition, Adapter pattern, Dependency inversion, Hexagonal integration test |
| `architecture` | `event-driven` | Command handler, Domain event, Event projection, Saga orchestrator, Event idempotency |
| `testing` | `foundational` | Test Desiderata — Las 12 propiedades de Kent Beck para buenos tests |
| `testing` | `tdd` | Test-Driven Development — Red-Green-Refactor *(incluye foundational)* |
| `testing` | `bdd` | Behavior-Driven Development — Given/When/Then *(incluye foundational)* |
| `conventions` | `conventional-commit` | Conventional Commits |
| `conventions` | `semantic-versioning` | Semantic Versioning |
| `conventions` | `all` | Todas las skills de convenciones combinadas |

Los cuatro presets de `architecture` son espejo de los cuatro `--preset` de generacion de contexto, asi que las skills instaladas con `hexagonal` se alinean con AGENTS.md/CONTEXT.md generados con `--preset hexagonal`.

#### Ecosistemas target

Cada ecosistema recibe frontmatter YAML especifico y rutas de salida:

| Target | Frontmatter | Ruta de salida |
|--------|-------------|----------------|
| `claude` *(default)* | `name`, `description`, `user-invocable: true` | `.claude/skills/` |
| `codex` | `name`, `description` | `.agents/skills/` |
| `antigravity` | `name`, `description`, `triggers` | `.agents/skills/` |

#### Opciones

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

### 🔄 Workflows

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

#### Dos modos

| Modo | Que hace | API key | Costo | Velocidad |
|------|----------|---------|-------|-----------|
| **Static** | Entrega workflows pre-construidos del catalogo embebido. Frontmatter por ecosistema. | No necesaria | Gratis | Instantaneo |
| **Personalized** | LLM adapta workflows a tu proyecto — los pasos referencian tus herramientas, CI/CD y targets de despliegue. | Requerida | ~centavos | ~10s |

#### Modo interactivo

```bash
codify workflows
# → Selecciona preset (spec-driven-change, bug-fix, release-cycle, all)
# → Selecciona ecosistema target (claude, antigravity)
# → Selecciona modo (static o personalized)
# → Selecciona locale
# → Selecciona ubicacion de instalacion (global, project, o custom)
# → Si personalized: describe tu proyecto, elige modelo
```

#### Modo CLI

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

#### Ecosistemas target

| Target | Salida | Estructura | Diferencia clave |
|--------|--------|------------|------------------|
| `claude` | Native skill | `{workflow}/SKILL.md` con frontmatter YAML | Anotaciones eliminadas, instrucciones en prosa |
| `antigravity` *(default)* | Archivo `.md` plano | `{workflow}.md` con frontmatter YAML | Anotaciones nativas: `// turbo`, `// capture`, `// if`, `// parallel` |

#### Scopes de instalacion

| Scope | Path Claude | Path Antigravity |
|-------|-------------|------------------|
| `global` | `~/.claude/skills/` | `~/.gemini/antigravity/global_workflows/` |
| `project` | `.claude/skills/` | `.agent/workflows/` |

#### Catalogo de workflows

| Preset | Workflow | Descripcion |
|--------|----------|-------------|
| `spec-driven-change` | Cambio Spec-driven | Proponer → aplicar → archivar — ciclo SDD completo con deltas formales, branch creation y cleanup de merge |
| `bug-fix` | Bug Fix | Reproducir → diagnosticar → corregir → testear → PR |
| `release-cycle` | Release Cycle | Bump de version → changelog → tag → deploy |
| `all` | Todos los workflows | Todos los presets de workflow combinados |

#### Spec-driven Change: la filosofia

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

#### Skills vs Workflows

| | Skills | Workflows |
|-|--------|-----------|
| **Proposito** | Ensenan *como* hacer una tarea especifica | Orquestan una *secuencia* de tareas |
| **Alcance** | Responsabilidad unica (ej. "escribir un commit") | Proceso end-to-end (ej. "evolucionar un spec desde propuesta hasta cambio mergeado") |
| **Invocacion** | El agente lee cuando es relevante | El usuario invoca via `/command` |
| **Ejemplos** | Conventional Commits, DDD entity, code review | Ciclo de cambio spec-driven, bug fix, release cycle |

#### Opciones

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

### 🪝 Hooks

Los hooks son **guardrails deterministicos** para Claude Code. Donde los skills (prompts) y los workflows (orquestacion) dependen de que el LLM haga lo correcto, los hooks son scripts shell que **siempre** se ejecutan en eventos del lifecycle (`PreToolUse`, `PostToolUse`, etc.) — hacen cumplir reglas en cada llamada, por exit code.

Las tres capas de artefactos se complementan:

| Capa | Mecanismo | Cuando corre? | Determinismo |
|---|---|---|---|
| **Skills** | Prompt cargado en contexto | Cuando agente o usuario lo invoca | Depende del LLM |
| **Workflows** | Lifecycle multi-skill | Usuario lo invoca via slash command | Depende del LLM |
| **Hooks** | Scripts shell en eventos | Cada llamada a tool que coincida | 100% (exit codes) |

#### Catalogo de presets

| Preset | Evento | Proposito |
|---|---|---|
| `linting` | `PostToolUse` (Edit\|Write) | Auto-formatea y lintea archivos usando la herramienta correcta por lenguaje (Prettier/ESLint, ruff/black, gofmt/gofumpt, rustfmt, rubocop, shfmt). Detecta tools instalados via `command -v` — silencioso si falta uno. |
| `security-guardrails` | `PreToolUse` (Bash, Edit\|Write) | Bloquea comandos Bash peligrosos (`rm -rf /`, `git push --force` a main, `curl \| bash`, fork bombs, formateo de fs) y protege archivos sensibles (`.env*`, `secrets/`, `.git/`, lockfiles, claves privadas, configs CI). |
| `convention-enforcement` | `PreToolUse` (Bash con `if`) | Valida mensajes de commit contra Conventional Commits 1.0.0 (titulo ≤72 chars, tipo valido, sin placeholders triviales) y bloquea push directo/force-push a branches protegidos (`main`, `master`, `develop`, `production`, `release/*`). Requiere Claude Code v2.1.85+. |
| `all` | (combinado) | Los tres presets mergeados en un solo `hooks.json` |

#### Modos de activacion

| Flag | Comportamiento |
|---|---|
| `--install project` (default interactivo) | Mergea en `.claude/settings.json` y copia scripts a `.claude/hooks/`. Crea backup antes de modificar. Idempotente: ejecutarlo dos veces no agrega handlers duplicados. |
| `--install global` | Igual que project pero en `~/.claude/settings.json` y `~/.claude/hooks/` (todos los proyectos) |
| `--output PATH` | **Modo preview** — escribe `{PATH}/hooks.json` + `{PATH}/hooks/*.sh` standalone para inspeccion o merge manual. NO toca `settings.json` |
| `--dry-run` | Imprime el `settings.json` resultante del merge propuesto, sale 0, no escribe nada |

#### Estructura de salida

```
~/.claude/                      O    ./.claude/
├── settings.json   (mergeado)        ├── settings.json   (mergeado)
├── settings.json.codify-backup-…     ├── settings.json.codify-backup-…
└── hooks/                            └── hooks/
    ├── lint.sh                            ├── lint.sh
    ├── block-dangerous-commands.sh        ├── block-dangerous-commands.sh
    ├── protect-sensitive-files.sh         ├── protect-sensitive-files.sh
    ├── validate-commit-message.sh         ├── validate-commit-message.sh
    └── check-protected-branches.sh        └── check-protected-branches.sh
```

#### Modo interactivo

```bash
codify hooks
# → Selecciona preset (linting, security-guardrails, convention-enforcement, all)
# → Selecciona locale (en, es)
# → Selecciona modo de activacion (project / global / preview)
```

#### Modo CLI

```bash
# Activar todo para el proyecto actual (flujo default)
codify hooks --preset all --install project

# Globalmente para todos tus proyectos
codify hooks --preset all --install global

# Solo preview (escribe bundle, no toca settings.json)
codify hooks --preset linting --output ./tmp/preview

# Ver el merge propuesto sin escribir nada
codify hooks --preset all --install project --dry-run

# Mensajes stderr en espanol
codify hooks --preset linting --install project --locale es
```

#### Verificar activacion

```bash
claude
> /hooks
```

#### Rollback

Cada install hace backup del `settings.json` previo a `settings.json.codify-backup-<timestamp>`. Para revertir:

```bash
mv .claude/settings.json.codify-backup-<timestamp> .claude/settings.json
```

#### Requisitos

- **Bash** + **jq** (Linux/macOS nativo; Windows requiere Git Bash o WSL)
- **Claude Code v2.1.85+** (solo para el preset `convention-enforcement`, que usa el campo `if` en handlers)

#### Limitaciones honestas

Los scripts bash usan patrones regex, no AST parsing. Detienen comandos **descuidados** del agente, no adversarios motivados — ofuscacion sofisticada (e.g. `eval $(echo b3JtIC1yZiAv | base64 -d)`) puede burlar la deteccion. Para garantias mas fuertes usa una herramienta dedicada como [bash-guardian](https://github.com/RoaringFerrum/claude-code-bash-guardian). Los scripts son cortos y deliberadamente editables: extiende los arrays de patrones para tu modelo de riesgo especifico.

#### Opciones

```bash
codify hooks [flags]
```

| Flag | Descripcion | Default |
|---|---|---|
| `--preset` `-p` | `linting`, `security-guardrails`, `convention-enforcement`, o `all` | *(interactivo)* |
| `--locale` | Idioma de salida para stderr (`en` o `es`) | `en` |
| `--install` | Scope de instalacion: `global` o `project` (auto-activa) | *(interactivo — default `project`)* |
| `--output` `-o` | Directorio preview: escribe bundle standalone, no toca settings | — |
| `--dry-run` | Imprime el merge propuesto sin escribir nada | `false` |

---

## 🔧 Maintain (lifecycle continuo)

> **Continuo.** Estos comandos operan sobre un proyecto ya equipado. Detectan drift, regeneran artefactos desactualizados, auditan commits y mantienen el costo transparente. Aplican igual a proyectos greenfield y brownfield.

### 🔍 Lifecycle: Drift Detection

Una vez que Codify genera artefactos, el mundo sigue moviendose. Las dependencias cambian, el README evoluciona, alguien edita `AGENTS.md` a mano. Sin chequeo activo, los artefactos se desfasan silenciosamente del proyecto.

`codify check` y su comando complementario `codify reset-state` resuelven esto sin LLM: hashes SHA256 de artefactos y senales de input, capturados al momento de generacion y comparados al momento de check. **Cero costo LLM. Cero red. Totalmente deterministico.**

#### `codify check` — detectar drift en CI o localmente

```bash
codify check                    # reporte legible; exit 1 si hay drift significativo
codify check --strict           # cualquier drift (incluso minor) dispara exit 1
codify check --json             # JSON machine-readable para pipelines CI
codify check -o ./output/my-project   # si los artefactos viven fuera del cwd
```

**Qué detecta:**

| Tipo de drift | Severidad | Que significa |
|---|---|---|
| `artifact_modified` | significant | Un archivo generado (e.g. AGENTS.md) fue editado despues de generacion |
| `artifact_missing` | significant | Un archivo presente en el snapshot ya no esta en disco |
| `signal_changed` | significant | Un input signal (`go.mod`, `Makefile`, `README.md`, etc.) cambio — tu contexto puede haber quedado desfasado |
| `signal_removed` | significant | Un signal trackeado ya no esta en disco |
| `artifact_new` | minor | Un nuevo artefacto aparecio desde el snapshot |
| `signal_added` | minor | Un nuevo signal aparecio (informativo) |

**Exit codes:**

- `0` — sin drift significativo (o sin drift en general)
- `1` — drift significativo (default) o cualquier drift (con `--strict`)
- `2` — no existe `.codify/state.json` (proyecto sin bootstrap)

**Ejemplo de uso en CI (GitHub Actions):**

```yaml
- name: Verify Codify artifacts are in sync
  run: codify check --strict
```

Un exit no-cero falla el job, asi PRs que cambian dependencias sin regenerar contexto se detectan automaticamente.

#### `codify reset-state` — aceptar el FS actual como nuevo baseline

Cuando editaste intencionalmente `AGENTS.md` (e.g. ajustaste una restriccion a mano) y querés que Codify considere eso como la nueva verdad:

```bash
codify reset-state              # recomputa state.json desde el FS actual, escritura atomica
codify reset-state --dry-run    # solo preview, sin cambios
```

El comando es read-only sobre tus artefactos — nunca modifica AGENTS.md ni archivos de context. Solo actualiza `state.json` (con backup en `.bak`). Los `check` siguientes comparan contra el nuevo baseline.

#### Como funciona drift detection por debajo

Cada `codify generate` / `codify analyze` / `codify init` exitoso escribe `.codify/state.json` que contiene:

- Metadata del proyecto (nombre, preset, lenguaje, locale, target)
- Contexto git (commit, branch, remote, dirty status)
- Artefactos: SHA256 + tamano + timestamp de generacion para cada archivo generado
- Input signals: SHA256 de archivos bien-conocidos (`go.mod`, `Makefile`, `README.md`, etc.)

`codify check` recomputa este snapshot desde el FS actual y diffea los dos. La operacion es local, rapida (<100ms tipico), y totalmente reproducible.

---

### 🔄 Lifecycle: Update, Audit y Tracking de Uso

Tres comandos construyen sobre la fundacion de drift detection para cerrar el gap entre "Codify genero artefactos una vez" y "Codify los mantiene a medida que el proyecto evoluciona": `update` regenera selectivamente, `audit` revisa commits contra convenciones documentadas, `usage` expone el costo LLM.

#### `codify update` — regeneracion selectiva

```bash
codify update                    # detecta drift, regenera via analyze si hace falta
codify update --dry-run          # muestra que cambiaria sin costo LLM
codify update --force            # regenera incluso con drift menor
codify update --accept-current   # mantiene FS actual como nuevo baseline (alias de reset-state)
```

Si solo hay hand-edits a artefactos (sin signals cambiando), `update` se rehusa con exit 1 y sugiere `--accept-current` o `reset-state` — diseñado para no perder ediciones intencionales del usuario.

#### `codify audit` — revisar commits contra convenciones

```bash
codify audit                     # ultimos 20 commits, rules-only (cero costo LLM)
codify audit --since main~50     # todos los commits desde main~50
codify audit --strict            # cualquier finding (incl. minor) falla el run
codify audit --json              # machine-readable para CI
codify audit --with-llm          # heuristico — envia commits + AGENTS.md al LLM (registra usage)
```

Findings rules-only: `commit_invalid_type`, `commit_trivial`, `commit_header_too_long`, `protected_branch_direct`. Types reconocidos: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, etc.

#### `codify usage` — transparencia de costos LLM

Cada call LLM se registra automaticamente en `.codify/usage.json` (proyecto) y `~/.codify/usage.json` (global).

```bash
codify usage                       # gasto del proyecto actual
codify usage --global              # agregado de todos los proyectos
codify usage --since 7d            # ultimos 7 dias
codify usage --by command          # breakdown por comando
codify usage --json                # JSON para scripting
codify usage --reset               # archiva log y empieza fresh
```

Costo computado con tabla de precios publica embebida (version `2026-05`). Refleja list prices de Anthropic y Google — **no** descuentos negociados.

**Tres formas de opt-out:**

```bash
codify update --no-tracking                    # por invocacion
export CODIFY_NO_USAGE_TRACKING=1              # por shell
touch ~/.codify/.no-usage-tracking             # permanente
```

#### CI con GitHub Actions

```yaml
# .github/workflows/codify.yml
name: Codify drift + audit
on: [pull_request]

jobs:
  codify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 50
      - name: Install Codify
        run: |
          go install github.com/jorelcb/codify/cmd/codify@latest
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH
      - name: Verify generated artifacts in sync
        run: codify check --strict
      - name: Audit recent commits
        run: codify audit --since origin/main --strict
```

`check` y `audit --rules-only` no requieren API key. `update` y `audit --with-llm` si.

---

### ✏️ Lifecycle: Resolucion de Markers (`codify resolve`)

`codify resolve` es la superficie standalone para llenar markers `[DEFINE: ...]` — los placeholders que el LLM emite en archivos de contexto generados cuando la descripcion del proyecto no cubrio algo. v2.1.0 lo promovio de un hook inline post-`generate` a un comando first-class que puedes correr en cualquier momento sobre archivos existentes.

#### Cuando usarlo

- Ya shippeaste `AGENTS.md` / `CONTEXT.md` con markers intactos (declinaste el prompt inline durante `generate`, o generaste antes de v2.0.5).
- Agregaste a mano una nueva seccion a un archivo de contexto y quieres re-resolver markers que se introdujeron.
- Tu CI flageo markers `[DEFINE]` sin resolver y quieres arreglarlos en un solo paso sin re-generar.

#### Seleccion de archivos

```bash
codify resolve AGENTS.md CONTEXT.md   # lista explicita
codify resolve --all                  # walk del cwd, todo archivo con marker
codify resolve --since=HEAD~5         # archivos cambiados en git desde <ref>
```

El walk de `--all` salta `.git`, `node_modules`, `vendor`, `.codify`, y archivos binarios (byte NUL en los primeros 4KB).

#### Flujo interactivo (LLM-driven por default)

Para cada archivo con markers, Codify hace una llamada LLM de enrichment para traducir el `[DEFINE: hint]` crudo en un prompt amigable con sugerencias grounded:

```
── AGENTS.md (2 markers) ──

     40  ## Currency Configuration
     41
  ▸  42  The supported currency is [DEFINE: ISO 4217 currency code], using two

    ¿Qué moneda usa la aplicación?
    (contexto fintech inferido de linea 12)
    Suggestions:
      1) USD [default]
      2) EUR
      3) MXN
    Your answer (1-N, text, Enter for default, s to skip)
```

Parser de input:
- **entero 1-N** → elige la sugerencia
- **texto libre** → usa como respuesta
- **Enter** → usa default si existe, si no skip
- **`s`** o **`skip`** → skip explicito (case-insensitive)

Cuando el enrichment falla (sin API key, error del provider, respuesta malformada, sanitizer rechazo todo), el prompt cae a la forma legacy (`Your input for L42 (Enter to skip)`) — nunca un fallo duro.

#### Modo de skip

Por default, skipear un marker lo reemplaza con un comentario TODO con fecha en la sintaxis nativa del archivo — la gap queda visible en panels de TODO del IDE y en grep:

| Extension | Reemplazo |
|---|---|
| `.md`, `.html`, `.htm`, `.xml` | `<!-- TODO 2026-05-06: ISO 4217 code -->` |
| `.go`, `.js`, `.ts`, `.java`, `.rs`, `.c`, `.cpp`, `.swift`, `.cs`, ... | `// TODO 2026-05-06: ISO 4217 code` |
| `.py`, `.rb`, `.sh`, `.yml`, `.yaml`, `.toml`, `.ini`, ... | `# TODO 2026-05-06: ISO 4217 code` |
| Otra / desconocida | Marker preservado verbatim (default safe) |

Pasa `--skip-mode=verbatim` para mantener markers `[DEFINE: ...]` crudos en el archivo.

#### Diff preview

Despues del rewrite, antes de tocar el archivo en disco, ves un diff unified pequeno y eliges:

```
About to rewrite AGENTS.md:
    line 41
  - The supported currency is [DEFINE: ISO 4217 currency code], using two
  + The supported currency is USD, using two
    line 43

Apply changes?  Apply / Discard (keep file as-is) / Edit before applying
```

- **Apply** — escribe el contenido propuesto
- **Discard** — archivo intacto, contribuye al contador `FilesDiscarded` del summary
- **Edit** — abre el contenido propuesto en `$EDITOR` (con fallback a `vim` / `vi` / `nano`); se escriben los bytes guardados

Salta el preview con `--no-preview`.

#### Guardrails anti-alucinacion

Dos capas protegen contra el LLM haciendo mas de lo que se le pidio:

1. **Sanitizer de sugerencias.** Antes que el usuario las vea, las sugerencias del enricher se filtran: URLs, paths de archivos, strings multi-linea, texto con markdown fences y valores de mas de 50 chars se descartan. Sugerencias se deduplican case-insensitive y se cap a 3 entradas; el default propuesto por el LLM debe matchear una de las sobrevivientes o se descarta.
2. **Validator post-rewrite.** Despues que el LLM reescribe el archivo con las respuestas del usuario, Codify re-escanea el output y clasifica los markers por frecuencia (los numeros de linea cambian, los conteos de texto no):
   - `Lost` — el usuario skipeo este marker pero el LLM lo borro igual
   - `NotApplied` — el usuario respondio pero el marker sigue en el archivo
   - `Spurious` — markers que no existian en el input pero aparecen en el output
   
   Cualquiera de esos dispara un fallback transparente a substitucion literal deterministica, preservando todas las respuestas del usuario. Va una WARNING a stderr explicando el downgrade.

#### Flags

```bash
codify resolve [files...] [flags]
```

| Flag | Descripcion | Default |
|------|-------------|---------|
| `--all`, `-a` | Walk del cwd recursivo por archivos con markers `[DEFINE]` | `false` |
| `--since` | Solo resuelve archivos cambiados en git desde este ref (e.g. `HEAD~5`) | — |
| `--no-enrich` | Salta el step LLM de question/suggestions (mas barato, menos amigable) | `false` |
| `--no-preview` | Salta el diff preview antes de escribir archivos | `false` |
| `--skip-mode` | `todo` (default, comentario TODO en sintaxis del archivo) o `verbatim` | `todo` |
| `--dry-run` | Walk de markers y reporta que cambiaria sin escribir archivos | `false` |
| `--locale` | Locale de output para los prompts de rewrite/enrichment del LLM | `en` |
| `--model`, `-m` | Modelo LLM | auto-detect |

#### Notas de costo

La llamada de enrichment usa prompt caching de Anthropic (TTL ephemeral de 5 min). Para una generacion tipica de 3 archivos, el mismo system prompt se reusa entre archivos — la segunda y tercera llamadas pegan el cache. Callers de Gemini pagan costo full de input-token por llamada (su API de caching tiene un minimo de 4096 tokens que los prompts del resolver no alcanzan). Sin provider configurado, el resolver cae a la UI legacy + substitucion literal — sin costo LLM, UX menos polished.

El mismo flujo corre automaticamente al final de `codify generate` / `analyze` / `init`, asi que la mayoria de usuarios nunca invocan `codify resolve` a mano. Usalo cuando quieras revisar archivos existentes o cuando quieras alguno de los flags opt-out (`--no-enrich`, `--skip-mode=verbatim`, `--dry-run`).

---

### 👁️ Lifecycle: Watcher Foreground (`codify watch`)

`codify watch` mantiene drift detection corriendo en background de tu sesion de editor. Re-ejecuta `check` automaticamente cuando cualquier archivo registrado en `.codify/state.json` cambia — input signals (e.g. `go.mod`, `Makefile`, `README.md`) y artefactos generados (`AGENTS.md`, `context/*.md`).

```bash
codify watch                         # default 2s debounce, solo reporta
codify watch --debounce 500ms        # debounce mas ajustado para feedback rapido
codify watch --auto-update --strict  # mantiene artefactos sincronizados agresivamente
```

**Comportamiento:**
- Carga `.codify/state.json` una vez al startup; exit 2 si falta
- Se suscribe via `fsnotify` a los dirs padres de los paths registrados (sin walk recursivo)
- Debouncea eventos (default 2s) — cinco saves rapidos disparan UN check, no cinco
- Imprime reportes de drift a stdout y sigue mirando
- `--auto-update` corre `codify update` cuando detecta drift significativo (registra usage LLM)
- `Ctrl+C` sale limpio

#### Por que foreground (no daemon)

`codify watch` es intencionalmente un **proceso foreground**, NO un daemon de sistema. No tiene `--detach`, no hay PID file, no hay reload por señales. Decision documentada en [ADR-008](docs/adr/0008-watch-model-decision.md). Resumen:

- **Manejo de PID files, signal handling, rotacion de logs, integracion con OS services** son problemas dificiles y fuera de scope para un proyecto de un solo mantenedor. Usuarios que necesitan persistencia envuelven con `tmux` / `nohup` / `systemd` / su supervisor preferido.
- **El use case realista es de corta duracion** — arrancas `watch` cuando empezas a codear, lo paras cuando terminas. Horas, no semanas.
- **El scope esta naturalmente acotado** — solo los ~20 paths de `state.json` se observan.

#### Envolver en un process supervisor

Si si queres watch long-running:

```bash
# Sesion tmux que sobrevive cierre del terminal
tmux new-session -d -s codify-watch "cd $(pwd) && codify watch"
tmux attach -t codify-watch         # para inspeccionar; Ctrl+B luego D para detach

# Unit de systemd (~/.config/systemd/user/codify-watch.service)
[Unit]
Description=Codify watch for %i
[Service]
WorkingDirectory=%h/projects/%i
ExecStart=/usr/local/bin/codify watch --debounce 5s
Restart=on-failure
[Install]
WantedBy=default.target

# nohup para sobrevivir la sesion
nohup codify watch > codify-watch.log 2>&1 &
```

#### Alternativa — integracion con git-hooks via `codify check`

Para usuarios cuyo modelo mental es "validar al hacer commit" en vez de "validar mientras edito", `codify check` es la herramienta correcta — es un one-shot deterministico disenado para CI y git hooks. Integrar via tu hook manager preferido:

**lefthook (`lefthook.yml`):**
```yaml
pre-commit:
  commands:
    codify-check:
      run: codify check --strict
```

**pre-commit (`.pre-commit-config.yaml`):**
```yaml
repos:
  - repo: local
    hooks:
      - id: codify-check
        name: Codify drift detection
        entry: codify check --strict
        language: system
        pass_filenames: false
```

**watchexec (alternativa foreground sobre la misma base de FS-events):**
```bash
watchexec -w go.mod -w Makefile -w README.md -- codify check
```

Codify mismo no genera estos configs — la integracion es lo suficientemente corta y especifica del proyecto que copy-paste es el primitive correcto (per [ADR-008](docs/adr/0008-watch-model-decision.md)).

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

#### Herramientas read-only (sin API key)

| Herramienta | Descripcion |
|-------------|-------------|
| `commit_guidance` | Spec de Conventional Commits y contexto comportamental para generar mensajes de commit |
| `version_guidance` | Spec de Semantic Versioning y contexto comportamental para determinar bumps de version |
| `get_usage` | Lee tracking de costos LLM desde `.codify/usage.json` (proyecto) o `~/.codify/usage.json` (global). Pure file read, sin LLM call. Parametros: `scope` (`project`/`global`), `since` (e.g. `7d`/`24h`), `by` (`command`/`model`/`provider`) |

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

Elige la filosofia arquitectonica para tu contexto. Codify trae **4 presets**:

| Preset | Foco | Cuando usar |
|---|---|---|
| `neutral` *(default)* | Sin opinion arquitectonica — la estructura se adapta al proyecto | Greenfield exploratorio, scripts, herramientas, donde querés minima opinion baked in |
| `clean-ddd` | DDD + Clean Architecture + BDD + Domain layered | Sistemas de negocio long-lived, logica rica de dominio, equipos comodos con layered architecture |
| `hexagonal` | Ports & Adapters — mas liviano que clean-ddd | Apps con preocupaciones fuertes de integracion externa, infra swappable, mas simple que DDD completo |
| `event-driven` | CQRS + Event Sourcing + Sagas | Sistemas async, coordinacion multi-servicio, dominios event-first, audit trails |

```bash
# Default — sin opinion arquitectonica
codify generate my-api -d "API REST de inventario en Go"

# Clean + DDD
codify generate my-api -d "API REST de inventario en Go" --preset clean-ddd

# Hexagonal — ports & adapters
codify generate my-payments -d "Servicio de pagos" --preset hexagonal

# Event-driven — CQRS + ES + sagas
codify generate my-orders -d "Procesamiento de ordenes" --preset event-driven
```

### `--from-file` — Descripciones ricas desde archivos

Para descripciones detalladas (documentos de diseno, RFCs, 6-pagers), usa `--from-file` en lugar de `--description`:

```bash
codify generate my-api \
  --from-file ./docs/descripcion-proyecto.md \
  --language go
```

El contenido del archivo se convierte en la descripcion del proyecto. Soporta cualquier formato de texto — markdown, texto plano, etc. Mutuamente excluyente con `--description`.

## 🚀 Migrando desde v1.x

Codify v2.0 tiene **un solo cambio breaking**. Todo lo demas (multi-target Claude/Codex/Antigravity, todos los comandos, todos los flags, todas las claves de config) sigue funcionando identico.

### Que cambio

| v1.x | v2.0 |
|---|---|
| `--preset default` (alias deprecado que resolvia a `clean-ddd` con warning) | **Removido** — devuelve error claro con instrucciones de migracion |
| Valor default del flag `--preset`: `clean-ddd` | **`neutral`** (sin opinion arquitectonica baked in) |
| `default` aceptado en `~/.codify/config.yml` | Mismo error al cargar el config |

El cambio de default refleja una decision documentada en [ADR-001](docs/adr/0001-default-preset-transition.md): el "default" de Codify era DDD/Clean — opinado. v2.0 hace que el default sea arquitectonicamente neutro, asi el agente parte de una base limpia salvo que elijas explicitamente una postura.

### Pasos de migracion

**Si usabas `--preset default` explicito:**

```bash
# Antes (v1.x):
codify generate my-api -d "..." --preset default

# Despues (v2.0): usa clean-ddd (mismo comportamiento que el default v1.x)
codify generate my-api -d "..." --preset clean-ddd

# O adopta el nuevo default explicitamente:
codify generate my-api -d "..." --preset neutral
```

**Si corrias `codify generate` sin `--preset` y queres mantener el comportamiento v1.x:**

Dos opciones:

```bash
# Opcion A — pasar --preset clean-ddd en cada invocacion
codify generate my-api -d "..." --preset clean-ddd

# Opcion B — setearlo como default global (recomendado para CI/scripts)
codify config set preset clean-ddd
```

**Si tu `~/.codify/config.yml` tiene `preset: default`:**

```bash
codify config set preset clean-ddd   # mantener comportamiento v1.x
codify config set preset neutral     # adoptar el default v2.0
```

### Lo que NO cambio

- Todos los targets siguen soportados: `claude`, `codex`, `antigravity` (per [ADR-009](docs/adr/0009-antigravity-deprecation-reversal.md), revierte el plan de deprecacion v1.26)
- Todos los comandos funcionan identico — `generate`, `analyze`, `spec`, `skills`, `workflows`, `hooks`, `config`, `init`, `check`, `update`, `audit`, `usage`, `watch`, `reset-state`
- Todos los demas flags, formatos de output, MCP tools (10 totales)
- Schemas de config, state.json, usage.json — sin cambios
- Tabla de pricing, locales, lenguajes — sin cambios

Si no pasas `--preset` explicito en ningun lado, la unica diferencia observable es que los nuevos AGENTS.md/CONTEXT.md generados van a ser arquitectura-agnosticos en vez de DDD-flavored. Los artefactos existentes no se afectan; `codify check` no flagea drift solo porque cambio la version.

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
    └── mcp/             Servidor MCP (transporte stdio + HTTP, 10 herramientas)
```

### Sistema de templates

```
templates/
├── en/                          Locale ingles
│   ├── neutral/                 Preset default — sin opinion arquitectonica
│   │   ├── agents.template
│   │   ├── context.template
│   │   ├── interactions.template
│   │   └── development_guide.template
│   ├── clean-ddd/               DDD + Clean Architecture + BDD
│   │   └── (mismos archivos)
│   ├── hexagonal/               Ports & Adapters
│   │   └── (mismos archivos)
│   ├── event-driven/            CQRS + Event Sourcing + Sagas
│   │   └── (mismos archivos)
│   ├── spec/                    Templates de especificacion (AI SDD)
│   │   ├── constitution.template
│   │   ├── spec.template
│   │   ├── plan.template
│   │   └── tasks.template
│   ├── skills/                  Templates de Agent Skills (static + guias LLM)
│   │   ├── neutral/             Architecture: review, testing, API design, refactoring
│   │   ├── clean-ddd/           Architecture: DDD entity, layer, BDD, CQRS, Hexagonal port
│   │   ├── hexagonal/           Architecture: port, adapter, dependency inversion, integration test
│   │   ├── event-driven/        Architecture: command handler, domain event, projection, saga, idempotency
│   │   ├── testing/             Testing: Foundational, TDD, BDD
│   │   └── conventions/         Conventions (conventional commits, semver)
│   ├── workflows/              Templates de workflows
│   │   ├── bug_fix.template
│   │   ├── release_cycle.template
│   │   ├── spec_propose.template
│   │   ├── spec_apply.template
│   │   └── spec_archive.template
│   ├── hooks/                  Templates de bundles de hooks
│   │   ├── linting/
│   │   ├── security-guardrails/
│   │   └── convention-enforcement/
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

**v2.1.1**

Snapshot completo de la superficie. Lo que aparece aqui esta shippeado, testeado y se comporta como esta documentado arriba.

**Capa Context**
- ✅ `generate` — contexto desde una descripcion (4 archivos, +1 con `--language`)
- ✅ `analyze` — contexto desde un repo existente via project scanner (18+ patrones de archivos de contexto, parseo de build targets, deteccion CI/CD, frameworks + dependencias en 8 lenguajes)
- ✅ `spec` + flag `--with-specs` — specs SDD (CONSTITUTION, SPEC, PLAN, TASKS)
- ✅ Streaming, reglas de grounding anti-alucinacion, validators de output (markers `[DEFINE]`, frontmatter, balance de code fences)
- ✅ Prompt caching de Anthropic en el loop de generacion por archivo

**Capa Behavior**
- ✅ `skills` — 4 presets de architecture (espejados con los presets de context) + testing + conventions; modos static + personalized; multi-ecosistema (claude, codex, antigravity)
- ✅ `workflows` — spec-driven-change, bug-fix, release-cycle; static + personalized; claude (native skills) + antigravity (anotaciones nativas)
- ✅ `hooks` — linting, security-guardrails, convention-enforcement; auto-install con backup + merge idempotente; `--output` preview y `--dry-run`

**Capa Bootstrap**
- ✅ `config` — wizard de config a nivel usuario con auto-launch SOFT (TTY-gated, triple opt-out); subcomandos `get` / `set` / `unset` / `edit` / `list`
- ✅ `init` — smart router a nivel proyecto (nuevo vs existente) que delega a `generate` o `analyze`

**Capa Lifecycle**
- ✅ `check` — drift detection (artifact_modified, signal_changed, etc.) — deterministico, sin LLM
- ✅ `update` — regeneracion selectiva via `analyze`; rechaza sobreescribir hand-edits sin `--force`
- ✅ `audit` — Conventional Commits + branches protegidas (rules-only, gratis) + modo heuristico `--with-llm` (registra usage)
- ✅ `usage` — tracking local de costos LLM (`.codify/usage.json` + `~/.codify/usage.json`); `--global`, `--since`, `--by`, `--json`, `--reset`
- ✅ `watch` — file watcher foreground con debounce, `--auto-update` opcional
- ✅ `reset-state` — recomputa snapshot sin tocar artefactos
- ✅ `resolve` — resolucion interactiva de markers `[DEFINE]` con prompts LLM-driven (sugerencias grounded + default), modo skip con TODO-anchor, validator post-rewrite (anti-alucinacion), diff preview, seleccion `--all` / `--since` / archivos explicitos, opt-outs `--no-enrich` / `--no-preview` / `--skip-mode=verbatim` / `--dry-run`

**MCP server**
- ✅ 10 tools: 7 generative (context/specs/analyze/skills/workflows/hooks/usage) + 3 read-only (commit_guidance/version_guidance/get_usage)
- ✅ Transportes stdio + HTTP; parametros con enums para validacion mas estricta; sin API key para read-only

**Distribucion**
- ✅ Homebrew tap (`brew install jorelcb/tap/codify`)
- ✅ `go install github.com/jorelcb/codify/cmd/codify@latest`
- ✅ Binarios pre-construidos en GitHub Releases

**Calidad**
- ✅ 9 paquetes BDD con 30+ scenarios; tests unitarios puros en domain + infrastructure
- ✅ Layout interno DDD/Clean Architecture (el proyecto come de su propia comida)

**Limites conocidos (intencionales, no roadmap):**
- Sin modo daemon para `watch` — envolver con tmux/nohup/systemd si se necesita (per [ADR-008](docs/adr/0008-watch-model-decision.md))
- Sin libreria Go `pkg/codify` — embedding via process boundary (CLI/MCP) es el contrato (per [ADR-003](docs/adr/0003-no-public-go-library.md))
- Hooks son Claude Code-only (la primitive subyacente no existe en codex/antigravity)

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

## 🆘 Solucion de Problemas

Los errores comunes y sus fixes rapidos estan consolidados en [`docs/troubleshooting.md`](docs/troubleshooting.md) (en ingles). Los mas frecuentes:

- **`ANTHROPIC_API_KEY or GEMINI_API_KEY environment variable is required`** — configura la key para comandos LLM-backed; los read-only (`check`, `audit --rules-only`, `usage`) no la necesitan.
- **`No snapshot at .codify/state.json...`** — proyecto sin bootstrap: corre `codify init`, `generate`, o `analyze` primero; o `codify reset-state` si el archivo se borro.
- **El prompt `Codify isn't configured globally yet. Run interactive setup now?` bloquea scripts** — pasa `--no-auto-config`, configura `CODIFY_NO_AUTO_CONFIG=1`, o `touch ~/.codify/.no-auto-config`.

Si tu sintoma no aparece en [`docs/troubleshooting.md`](docs/troubleshooting.md), abre un issue con: comando ejecutado, exit code, y stderr. El CHANGELOG y los ADRs documentan la mayoria de las decisiones de diseno.

## 📚 Documentacion

**Empieza aca:**
- [📘 Getting Started](docs/getting-started.md) — Tour end-to-end de 5 minutos con outputs esperados *(en ingles)*
- [📋 Lifecycle Matrix](docs/lifecycle-matrix.md) — Que comando aplica a workstation vs proyecto, greenfield vs brownfield *(en ingles)*
- [📖 Command Reference](docs/command-reference.md) — Cheatsheet de cada comando, agrupado por fase *(en ingles)*
- [🆘 Troubleshooting](docs/troubleshooting.md) — Errores comunes y fixes *(en ingles)*

**Referencia:**
- [📐 ADRs](docs/adr/) — Architectural Decision Records *(en ingles)*
- [📋 AGENTS.md](AGENTS.md) — Contexto del proyecto para agentes de IA
- [🏛️ Arquitectura](context/CONTEXT.md) — Detalle DDD/Clean Architecture
- [📝 Changelog](CHANGELOG.md) — Historial de cambios
- [🧪 Specs](specs/) — Especificaciones tecnicas (SDD)

> Los docs de `docs/` se mantienen en ingles por ahora. Las versiones en espanol pueden agregarse en el futuro segun demanda.

## 📄 Licencia

Apache License 2.0 — ver [LICENSE](LICENSE).

---

<div align="center">

**Contexto. Specs. Skills. Workflows. Hooks. Lifecycle. Tu agente, completamente equipado — y mantenido honesto.** 🧠

*"Un agente sin contexto es un pasante con acceso root — y contexto desactualizado es un pasante leyendo docs de hace tres semanas"*

⭐ Si te sirvio, dale una estrella — nos motiva a seguir construyendo

[🐛 Reportar bug](https://github.com/jorelcb/codify/issues) · [💡 Sugerir feature](https://github.com/jorelcb/codify/issues)

</div>