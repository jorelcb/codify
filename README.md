# 🤖 AI Context Generator

> Generador profesional de contextos para agentes de desarrollo con IA, basado en Domain-Driven Design y Clean Architecture

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/jorelcb/ai-context-generator)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Task](https://img.shields.io/badge/task-v3.0+-orange.svg)](https://taskfile.dev)

## 📋 Tabla de Contenidos

- [Descripción](#descripción)
- [Características](#características)
- [Requisitos](#requisitos)
- [Instalación](#instalación)
- [Uso Rápido](#uso-rápido)
- [Arquitectura](#arquitectura)
- [Templates](#templates)
- [Contribuir](#contribuir)
- [Licencia](#licencia)

## 📖 Descripción

AI Context Generator es una herramienta diseñada para crear contextos completos y profesionales para agentes de desarrollo con IA (Claude, GPT-4, etc.). Genera automáticamente toda la documentación necesaria para que un agente entienda tu proyecto, incluyendo arquitectura, scaffolding, prompts especializados y bitácoras de desarrollo.

### ¿Por qué usar esta herramienta?

- **Consistencia**: Todos tus proyectos siguen la misma estructura profesional
- **Mejores resultados con IA**: Los agentes entienden perfectamente tu arquitectura
- **Ahorro de tiempo**: No más copiar/pegar templates manualmente
- **Mejores prácticas**: DDD y Clean Architecture integrados desde el inicio
- **Multi-lenguaje**: Soporta Go, JavaScript, Python y más

## ✨ Características

### 🏗️ Arquitecturas Soportadas
- **Domain-Driven Design (DDD)** con capas bien definidas
- **Clean Architecture** con dependency rule estricta
- **Hexagonal Architecture** (Ports & Adapters)
- **MVC tradicional** para proyectos simples

### 💻 Lenguajes Soportados
- **Go** - Con estructura idiomática y go.mod
- **JavaScript/Node.js** - ESM modules y estructura moderna
- **Python** - Con type hints y pyproject.toml
- **Java** - Maven/Gradle ready (próximamente)
- **Rust** - Cargo structure (próximamente)

### 📦 Tipos de Proyecto
- **API REST** - Con OpenAPI/Swagger
- **CLI Tool** - Aplicaciones de línea de comandos
- **Library/Package** - Librerías reutilizables
- **Web Application** - SPAs y aplicaciones web
- **Microservice** - Servicios distribuidos

### 🤖 Proveedores de IA
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Google (Gemini)
- Azure OpenAI
- AWS Bedrock
- Local/Self-hosted (Ollama, etc.)

### 🔧 Características Técnicas
- **Task Runner**: Usa `go-task` para automatización
- **Templates modulares**: Combina base + tipo + lenguaje + capacidades
- **Scaffolding automático**: Genera estructura completa del proyecto
- **Documentación lista**: 4 documentos de contexto para agentes
- **Observabilidad**: OpenTelemetry integrado desde el inicio
- **Testing**: Estructura de tests incluida

## 📋 Requisitos

### Obligatorios
- **Task** (go-task) v3.0+
  ```bash
  # Linux/macOS
  sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d
  
  # macOS con Homebrew
  brew install go-task
  
  # Windows con Scoop
  scoop install task
  ```

- **Git** - Para control de versiones
- **jq** - Para procesamiento JSON (usado internamente)
  ```bash
  # Ubuntu/Debian
  sudo apt-get install jq
  
  # macOS
  brew install jq
  
  # Fedora
  sudo dnf install jq
  ```

### Opcionales (según el lenguaje)
- **Go 1.21+** para proyectos Go
- **Node.js 18+ LTS** para proyectos JavaScript
- **Python 3.11+** para proyectos Python

## 🚀 Instalación

### Opción 1: Instalación Nueva

```bash
# Descargar el script de instalación
curl -O https://raw.githubusercontent.com/tu-usuario/ai-context-generator/main/install.sh

# Hacer ejecutable
chmod +x install.sh

# Ejecutar instalador
./install.sh

# Navegar al directorio
cd ai-context-generator

# Inicializar
task init
```

### Opción 2: Actualización desde v1.x

```bash
# Desde el directorio padre del generador
./update.sh

# El script preservará tus proyectos en output/
```

### Opción 3: Instalación Manual

```bash
# Clonar repositorio
git clone https://github.com/jorelcb/ai-context-generator.git
cd ai-context-generator

# Inicializar
task init
```

## 🎯 Uso Rápido

### Crear un proyecto interactivo

```bash
task new
```

El wizard te preguntará:
1. Nombre del proyecto
2. Lenguaje de programación
3. Tipo de proyecto
4. Proveedor de IA
5. Capacidades adicionales

### Crear un proyecto rápido

```bash
task new:quick -- mi-api go api openai
```

### Ver proyectos generados

```bash
task list
```

### Limpiar proyectos

```bash
task clean  # Elimina todos
task clean:project -- mi-api  # Elimina uno específico
```

## 🏛️ Arquitectura

### Estructura del Generador

```
ai-context-generator/
├── Taskfile.yml              # Configuración principal de tareas
├── scripts/                  # Scripts del generador
│   ├── wizard.sh            # Wizard interactivo
│   └── generate_scaffolding.sh
├── templates/               # Templates modulares
│   ├── base/               # Templates fundamentales
│   │   ├── prompt.template
│   │   ├── context.template
│   │   ├── scaffolding.template
│   │   └── interactions.template
│   ├── languages/          # Por lenguaje
│   │   └── go/
│   │       └── scaffolding.md
│   ├── scaffolding/        # Configuración por lenguaje
│   │   └── go/
│   │       └── Taskfile.yml
│   ├── types/             # Por tipo de proyecto
│   │   └── api/
│   └── capabilities/      # Capacidades opcionales
│       └── database/
└── output/               # Proyectos generados
```

### Estructura de un Proyecto Generado

```
mi-proyecto/
├── Taskfile.yml         # Tareas de desarrollo
├── context/            # Documentación para agentes IA
│   ├── 01_PROMPT.md   # Instrucciones principales
│   ├── 02_CONTEXT.md  # Contexto del proyecto
│   ├── 03_SCAFFOLDING.md # Estructura detallada
│   └── 04_INTERACTIONS_LOG.md # Bitácora
├── cmd/               # Entry points (Go)
├── internal/          # Código privado con DDD
│   ├── domain/       # Lógica de negocio
│   ├── application/  # Casos de uso
│   ├── infrastructure/ # Adaptadores
│   └── interfaces/   # API/UI
└── README.md
```

## 📝 Templates

### Templates Base
Los templates base contienen la estructura común para todos los proyectos:
- **prompt.template**: Define el rol y misión del agente
- **context.template**: Arquitectura y requisitos técnicos
- **scaffolding.template**: Estructura de carpetas y archivos
- **interactions.template**: Bitácora de desarrollo

### Personalización
Puedes personalizar los templates editando los archivos en `templates/`:

```bash
# Editar template de prompt
vi templates/base/prompt.template

# Agregar nuevo lenguaje
mkdir templates/languages/rust
vi templates/languages/rust/scaffolding.md
```

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor:

1. Fork el proyecto
2. Crea tu feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la branch (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

### Agregar un Nuevo Lenguaje

1. Crear directorio en `templates/languages/tu-lenguaje/`
2. Agregar `scaffolding.md` con estructura idiomática
3. Crear `templates/scaffolding/tu-lenguaje/Taskfile.yml`
4. Actualizar `scripts/generate_scaffolding.sh`
5. Documentar en este README

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🙏 Agradecimientos

- [Task](https://taskfile.dev) - Por el excelente task runner
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) - Por los principios arquitectónicos
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Por la estructura en capas
- La comunidad de desarrollo con IA - Por el feedback y mejoras

## 📞 Contacto

Tu Nombre - [@JorelSan_](https://twitter.com/tu-twitter)

Link del Proyecto: [https://github.com/jorelcb/ai-context-generator](https://github.com/jorelcb/ai-context-generator)

---

Hecho con ❤️ para mejorar el desarrollo con agentes de IA