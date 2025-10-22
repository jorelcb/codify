# 🤖 AI Context Generator

> Generador de contextos y documentación para agentes de desarrollo con IA

[![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)](https://github.com/jorelcb/ai-context-generator)
[![Status](https://img.shields.io/badge/status-beta-green.svg)](https://github.com/jorelcb/ai-context-generator/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev)

[English](README.md) | **Español**

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

## 📋 Requisitos

### Obligatorios
- **Git** - Para control de versiones
- **Go 1.21+** - Para compilar y ejecutar la herramienta

### Opcionales (según el lenguaje de proyecto)
- **Node.js 18+ LTS** para proyectos JavaScript
- **Python 3.11+** para proyectos Python

## 🚀 Instalación

```bash
# Clonar repositorio
git clone https://github.com/jorelcb/ai-context-generator.git
cd ai-context-generator

# Compilar
go build -o bin/ai-context-generator ./cmd/ai-context-generator

# Agregar al PATH (opcional)
export PATH=$PATH:$(pwd)/bin
```

## 🎯 Uso

### Generar un nuevo proyecto

```bash
ai-context-generator generate my-awesome-project \
  --language go \
  --type api \
  --architecture clean
```

### Listar proyectos generados

```bash
ai-context-generator list
```

### Validar un template

```bash
ai-context-generator validate ./templates/my-template
```

## 🏛️ Arquitectura

### Estructura del Generador

```
ai-context-generator/
├── cmd/                    # Entry points
│   └── ai-context-generator/
├── internal/               # Código privado (DDD/Clean)
│   ├── domain/            # Lógica de negocio
│   ├── application/       # Casos de uso (CQRS)
│   ├── infrastructure/    # Adaptadores
│   └── interfaces/        # CLI, API
├── templates/             # Templates modulares
└── tests/                 # Tests BDD con Godog
```

### Estructura de un Proyecto Generado

```
mi-proyecto/
├── context/               # Documentación para agentes IA
│   ├── PROMPT.md
│   ├── CONTEXT.md
│   ├── SCAFFOLDING.md
│   └── INTERACTIONS_LOG.md
├── cmd/                   # Entry points
├── internal/              # Código privado con DDD
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   └── interfaces/
└── README.md
```

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor:

1. Fork el proyecto
2. Crea tu feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios siguiendo Conventional Commits
4. Push a la branch (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 🙏 Agradecimientos

- [Task](https://taskfile.dev) - Por el excelente task runner
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) - Por los principios arquitectónicos
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Por la estructura en capas
- La comunidad de desarrollo con IA - Por el feedback y mejoras

---

Hecho con ❤️ para mejorar el desarrollo con agentes de IA