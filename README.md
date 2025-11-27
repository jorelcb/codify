# 🤖 AI Context Generator

> **Plataforma de Bootstrapping para Desarrollo Asistido por IA**

[![Version](https://img.shields.io/badge/version-1.2.0--beta-blue.svg)](https://github.com/jorelcb/ai-context-generator/releases)
[![Build](https://img.shields.io/badge/build-passing-green.svg)](actions)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## 🌀 El Concepto (Incepción)

Este proyecto es una exploración en **Ingeniería de Software Recursiva**:
> Una herramienta construida por IA (Nivel 1), diseñada para generar contextos técnicos (Nivel 4), que permiten a otras IAs (Nivel 5) construir software de alta calidad.

No es solo un generador de plantillas; es un **inyector de arquitectura**.

## 🎯 La Misión

Transformar intenciones humanas vagas en **Entornos de Desarrollo Cognitivamente Optimizados**.

Cuando inicias un proyecto con `ai-context-generator`, no solo obtienes carpetas vacías. Obtienes un "Cerebro Exógeno" pre-configurado con:
1.  **Contexto Semántico:** Archivos `.md` diseñados para ser leídos por LLMs, explicando la arquitectura del proyecto.
2.  **Estructura Sólida:** Scaffolding DDD/Clean Architecture que impone orden desde el segundo cero.
3.  **Reglas de Negocio:** Definiciones claras que evitan alucinaciones en el desarrollo futuro.

## ⚡ Quick Start

### Generar un Nuevo Contexto

```bash
# Crear el "Sustrato Cognitivo" para tu próximo proyecto
ai-context-generator generate my-payment-service \
  --type microservice \
  --lang go \
  --arch ddd
```

### ¿Qué sucede después?
1.  La herramienta genera la carpeta `./my-payment-service`.
2.  Tú abres esa carpeta en tu IDE con tu Agente de IA favorito (Cursor, Windsurf, Gemini Code Assist).
3.  El Agente lee `context/PROMPT.md` y entiende inmediatamente su rol y las restricciones arquitectónicas.
4.  El desarrollo comienza con **contexto perfecto**.

## 🏗️ Estado del Proyecto (Realidad Técnica)

Actualmente en **Fase Beta Técnica (v1.2.0)**.

- ✅ **Core Logic:** Motor de plantillas y definiciones de dominio implementados.
- ✅ **Arquitectura:** El diseño interno sigue estrictamente DDD.
- 🚧 **Infraestructura:** La escritura en disco (Filesystem Adapter) está en desarrollo activo.
- 🚧 **Modo Interactivo:** El CLI aún requiere argumentos explícitos.

## 🤝 Contribución

Este proyecto se construye mediante una dinámica de **Agentes Especializados**.
Si deseas contribuir, consulta `ARCHITECTURE.md` para entender los patrones de diseño que replicamos.

## 📖 Documentación

- [Architecture Guide](ARCHITECTURE.md) - Deep dive into DDD/Clean implementation
- [Contributing Guide](CONTRIBUTING.md) - How to contribute effectively
- [Changelog](CHANGELOG.md) - Version history and updates

## 🧪 Testing

Built with **Behavior-Driven Development (BDD)**:

```bash
# Run all tests
go test ./...
```

Current status: **36/39 scenarios passing (92.3%)**

## 📊 Project Status

**Version 1.2.0 - Beta**

✅ **Working:**
- Core DDD/Clean Architecture implementation
- Template Engine v1 (AST, Tokenizer, Parser)
- CLI structure
- BDD test suite

🚧 **In Progress:**
- File system infrastructure (Writing to disk)
- End-to-end integration
- Advanced AI Context generation logic

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 💝 Acknowledgments

- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) by Eric Evans
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Robert C. Martin
- [Godog](https://github.com/cucumber/godog) - BDD testing framework
- The AI development community for inspiration and feedback

## 🌟 Star History

If this project helps you, please consider giving it a ⭐️!

[![Star History Chart](https://api.star-history.com/svg?repos=jorelcb/ai-context-generator&type=Date)](https://star-history.com/#jorelcb/ai-context-generator&Date)


---

**Built with ❤️ to empower developers working with AI assistants**

[Report Bug](https://github.com/jorelcb/ai-context-generator/issues) · [Request Feature](https://github.com/jorelcb/ai-context-generator/issues) · [Join Discussions](https://github.com/jorelcb/ai-context-generator/discussions)