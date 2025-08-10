#!/bin/bash
# scripts/wizard.sh - Wizard interactivo mejorado con output JSON

# Colores
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Arrays de opciones
LANGUAGES=("go" "javascript" "python" "java" "rust")
PROJECT_TYPES=("api" "cli" "library" "webapp" "service")
AI_PROVIDERS=("openai" "anthropic" "google" "azure" "aws" "local" "none")
CAPABILITIES=("database" "messaging" "caching" "authentication" "monitoring" "testing" "documentation")
ARCHITECTURES=("ddd" "hexagonal" "clean" "mvc" "layered" "microservices")

# Función para imprimir con color
print_color() {
    echo -e "${1}${2}${NC}" >&2
}

# Función para validar nombre de proyecto
validate_project_name() {
    if [[ ! "$1" =~ ^[a-zA-Z][a-zA-Z0-9_-]*$ ]]; then
        return 1
    fi
    return 0
}

# Banner
clear >&2
print_color "$BLUE" "╔══════════════════════════════════════════════════════════╗"
print_color "$BLUE" "║     🤖 GENERADOR DE CONTEXTOS PARA AGENTES DE IA 🤖      ║"
print_color "$BLUE" "╚══════════════════════════════════════════════════════════╝"
echo "" >&2

# === 1. NOMBRE DEL PROYECTO ===
while true; do
    print_color "$CYAN" "📝 Paso 1/8: Nombre del proyecto"
    echo -n "Ingresa el nombre (solo letras, números, - y _): " >&2
    read PROJECT_NAME
    
    if validate_project_name "$PROJECT_NAME"; then
        if [ -d "output/$PROJECT_NAME" ]; then
            print_color "$YELLOW" "⚠️  Ya existe un proyecto con ese nombre"
            echo -n "¿Sobrescribir? (s/n): " >&2
            read OVERWRITE
            if [[ "$OVERWRITE" =~ ^[Ss]$ ]]; then
                break
            fi
        else
            break
        fi
    else
        print_color "$RED" "❌ Nombre inválido. Debe empezar con letra y contener solo letras, números, - y _"
    fi
done
echo "" >&2

# === 2. LENGUAJE ===
print_color "$CYAN" "💻 Paso 2/8: Lenguaje de programación"
echo "" >&2
for i in "${!LANGUAGES[@]}"; do
    echo "  $((i+1))) ${LANGUAGES[$i]}" >&2
done
echo "" >&2

while true; do
    echo -n "Selecciona (1-${#LANGUAGES[@]}): " >&2
    read LANG_CHOICE
    if [[ "$LANG_CHOICE" =~ ^[1-${#LANGUAGES[@]}]$ ]]; then
        LANGUAGE="${LANGUAGES[$((LANG_CHOICE-1))]}"
        break
    else
        print_color "$RED" "❌ Opción inválida"
    fi
done
echo "" >&2

# === 3. TIPO DE PROYECTO ===
print_color "$CYAN" "🏗️  Paso 3/8: Tipo de proyecto"
echo "" >&2
echo "  1) API REST/GraphQL" >&2
echo "  2) CLI Tool" >&2
echo "  3) Library/Package" >&2
echo "  4) Web Application" >&2
echo "  5) Microservice" >&2
echo "" >&2

while true; do
    echo -n "Selecciona (1-5): " >&2
    read TYPE_CHOICE
    if [[ "$TYPE_CHOICE" =~ ^[1-5]$ ]]; then
        PROJECT_TYPE="${PROJECT_TYPES[$((TYPE_CHOICE-1))]}"
        break
    else
        print_color "$RED" "❌ Opción inválida"
    fi
done
echo "" >&2

# === 4. ARQUITECTURA ===
print_color "$CYAN" "🏛️  Paso 4/8: Arquitectura"
echo "" >&2
echo "  1) Domain-Driven Design (DDD)" >&2
echo "  2) Hexagonal (Ports & Adapters)" >&2
echo "  3) Clean Architecture" >&2
echo "  4) MVC Traditional" >&2
echo "  5) Layered Architecture" >&2
echo "  6) Microservices Pattern" >&2
echo "" >&2

while true; do
    echo -n "Selecciona (1-6): " >&2
    read ARCH_CHOICE
    if [[ "$ARCH_CHOICE" =~ ^[1-6]$ ]]; then
        ARCHITECTURE="${ARCHITECTURES[$((ARCH_CHOICE-1))]}"
        break
    else
        print_color "$RED" "❌ Opción inválida"
    fi
done
echo "" >&2

# === 5. PROVEEDOR IA ===
print_color "$CYAN" "🤖 Paso 5/8: Proveedor de IA"
echo "" >&2
for i in "${!AI_PROVIDERS[@]}"; do
    echo "  $((i+1))) ${AI_PROVIDERS[$i]}" >&2
done
echo "" >&2

while true; do
    echo -n "Selecciona (1-${#AI_PROVIDERS[@]}): " >&2
    read AI_CHOICE
    if [[ "$AI_CHOICE" =~ ^[1-${#AI_PROVIDERS[@]}]$ ]]; then
        AI_PROVIDER="${AI_PROVIDERS[$((AI_CHOICE-1))]}"
        break
    else
        print_color "$RED" "❌ Opción inválida"
    fi
done
echo "" >&2

# === 6. DESCRIPCIÓN ===
print_color "$CYAN" "📄 Paso 6/8: Descripción del proyecto"
echo -n "Descripción breve (Enter para omitir): " >&2
read DESCRIPTION
if [ -z "$DESCRIPTION" ]; then
    DESCRIPTION="$PROJECT_TYPE project in $LANGUAGE using $ARCHITECTURE architecture"
fi
echo "" >&2

# === 7. CAPACIDADES ===
print_color "$CYAN" "⚡ Paso 7/8: Capacidades del proyecto"
print_color "$YELLOW" "Selecciona las capacidades (separadas por espacios, ej: 1 3 5)"
echo "" >&2
for i in "${!CAPABILITIES[@]}"; do
    echo "  $((i+1))) ${CAPABILITIES[$i]}" >&2
done
echo "" >&2
echo -n "Capacidades: " >&2
read CAPS_INPUT

SELECTED_CAPS=()
for cap in $CAPS_INPUT; do
    if [[ "$cap" =~ ^[1-${#CAPABILITIES[@]}]$ ]]; then
        SELECTED_CAPS+=("${CAPABILITIES[$((cap-1))]}")
    fi
done
echo "" >&2

# === 8. OPCIONES ADICIONALES ===
print_color "$CYAN" "⚙️  Paso 8/8: Opciones adicionales"
echo "" >&2

echo -n "¿Incluir tests? (s/n) [s]: " >&2
read INCLUDE_TESTS
INCLUDE_TESTS=${INCLUDE_TESTS:-s}

echo -n "¿Incluir Docker? (s/n) [s]: " >&2
read INCLUDE_DOCKER
INCLUDE_DOCKER=${INCLUDE_DOCKER:-s}

echo -n "¿Incluir CI/CD? (s/n) [s]: " >&2
read INCLUDE_CI
INCLUDE_CI=${INCLUDE_CI:-s}

echo -n "¿Incluir documentación API? (s/n) [s]: " >&2
read INCLUDE_API_DOCS
INCLUDE_API_DOCS=${INCLUDE_API_DOCS:-s}

echo -n "¿Usar CQRS pattern? (s/n) [n]: " >&2
read USE_CQRS
USE_CQRS=${USE_CQRS:-n}

echo "" >&2

# === RESUMEN ===
print_color "$GREEN" "═══════════════════════════════════════════════════════"
print_color "$GREEN" "📊 Resumen de configuración:"
print_color "$GREEN" "═══════════════════════════════════════════════════════"
echo "" >&2
echo "  📁 Proyecto:      $PROJECT_NAME" >&2
echo "  💻 Lenguaje:      $LANGUAGE" >&2
echo "  🏗️  Tipo:          $PROJECT_TYPE" >&2
echo "  🏛️  Arquitectura:  $ARCHITECTURE" >&2
echo "  🤖 Proveedor IA:  $AI_PROVIDER" >&2
echo "  📝 Descripción:   $DESCRIPTION" >&2
if [ ${#SELECTED_CAPS[@]} -gt 0 ]; then
    echo "  ⚡ Capacidades:   ${SELECTED_CAPS[*]}" >&2
fi
echo "  🧪 Tests:         $([ "$INCLUDE_TESTS" = "s" ] && echo "Sí" || echo "No")" >&2
echo "  🐳 Docker:        $([ "$INCLUDE_DOCKER" = "s" ] && echo "Sí" || echo "No")" >&2
echo "  🔄 CI/CD:         $([ "$INCLUDE_CI" = "s" ] && echo "Sí" || echo "No")" >&2
echo "  📚 API Docs:      $([ "$INCLUDE_API_DOCS" = "s" ] && echo "Sí" || echo "No")" >&2
echo "  🎯 CQRS:          $([ "$USE_CQRS" = "s" ] && echo "Sí" || echo "No")" >&2
echo "" >&2

# === CONFIRMACIÓN ===
echo -n "¿Confirmar y generar proyecto? (s/n): " >&2
read CONFIRM

if [[ ! "$CONFIRM" =~ ^[Ss]$ ]]; then
    print_color "$RED" "❌ Generación cancelada"
    exit 1
fi

# === GENERAR JSON OUTPUT ===
# Output JSON to stdout (no >&2) para que Task lo capture
cat <<EOF
{
  "project_name": "$PROJECT_NAME",
  "language": "$LANGUAGE",
  "project_type": "$PROJECT_TYPE",
  "architecture": "$ARCHITECTURE",
  "ai_provider": "$AI_PROVIDER",
  "description": "$DESCRIPTION",
  "capabilities": [$(printf '"%s",' "${SELECTED_CAPS[@]}" | sed 's/,$//')]",
  "include_tests": $([ "$INCLUDE_TESTS" = "s" ] && echo "true" || echo "false"),
  "include_docker": $([ "$INCLUDE_DOCKER" = "s" ] && echo "true" || echo "false"),
  "include_ci": $([ "$INCLUDE_CI" = "s" ] && echo "true" || echo "false"),
  "include_api_docs": $([ "$INCLUDE_API_DOCS" = "s" ] && echo "true" || echo "false"),
  "use_cqrs": $([ "$USE_CQRS" = "s" ] && echo "true" || echo "false"),
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

print_color "$GREEN" ""
print_color "$GREEN" "✅ Configuración guardada. Generando proyecto..." >&2
print_color "$GREEN" ""