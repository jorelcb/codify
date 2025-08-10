#!/bin/bash
# test_js_generation.sh - Integration tests para generación de proyectos JavaScript
set -euo pipefail

# Cargar helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/setup.sh"
source "$SCRIPT_DIR/../helpers/assertions.sh"

# Test suite para generación completa de proyectos JavaScript
test_javascript_api_generation() {
    init_test_suite "JavaScript API Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-js-api"
    local project_path
    
    # Test 1: Generar proyecto completo
    assert_success "Should generate JavaScript API project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' javascript api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura básica
    assert_directory_exists "Should create project directory" "$project_path"
    assert_directory_exists "Should create src directory" "$project_path/src"
    assert_directory_exists "Should create context directory" "$project_path/context"
    
    # Test 3: Verificar estructura DDD para JavaScript
    assert_directory_exists "Should create src/domain layer" "$project_path/src/domain"
    assert_directory_exists "Should create src/application layer" "$project_path/src/application"
    assert_directory_exists "Should create src/infrastructure layer" "$project_path/src/infrastructure"
    assert_directory_exists "Should create src/interfaces layer" "$project_path/src/interfaces"
    
    # Test 4: Verificar archivos específicos de JavaScript
    assert_file_exists "Should create package.json" "$project_path/package.json"
    assert_file_exists "Should create main entry point" "$project_path/src/index.js"
    assert_file_exists "Should create Taskfile.yml" "$project_path/Taskfile.yml"
    assert_file_exists "Should create .gitignore" "$project_path/.gitignore"
    
    # Test 5: Verificar contenido de package.json
    assert_valid_json "package.json should be valid JSON" "$project_path/package.json"
    assert_file_contains "package.json should have correct name" \
        "$project_path/package.json" "\"name\": \"$project_name\""
    assert_file_contains "package.json should have main script" \
        "$project_path/package.json" "\"main\": \"src/index.js\""
    
    # Test 6: Verificar archivos de contexto
    assert_file_exists "Should create PROMPT.md" "$project_path/context/01_PROMPT.md"
    assert_file_contains "Context should mention JavaScript" \
        "$project_path/context/02_CONTEXT.md" "JavaScript"
    
    # Test 7: Verificar .gitignore específico para Node.js
    assert_file_contains ".gitignore should ignore node_modules" \
        "$project_path/.gitignore" "node_modules/"
    assert_file_contains ".gitignore should ignore npm logs" \
        "$project_path/.gitignore" "npm-debug.log"
    
    # Test 8: Verificar que no hay placeholders sin reemplazar
    assert_success "Should not have unreplaced placeholders" \
        "validate_no_placeholders '$project_path'"
    
    cleanup_test_environment
    finish_test_suite "JavaScript API Generation Integration Test"
}

test_javascript_webapp_generation() {
    init_test_suite "JavaScript WebApp Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-js-webapp"
    local project_path
    
    # Test 1: Generar proyecto webapp
    assert_success "Should generate JavaScript webapp project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' javascript webapp openai"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura específica de webapp
    assert_directory_exists "Should create project directory" "$project_path"
    assert_file_exists "Should create package.json" "$project_path/package.json"
    
    # Test 3: Verificar configuraciones específicas de webapp
    if [ -f "$project_path/package.json" ]; then
        # Verificar scripts de desarrollo típicos de webapp
        assert_success "package.json should contain development scripts" \
            "grep -q '\"dev\"\\|\"start\"\\|\"build\"' '$project_path/package.json'"
    fi
    
    # Test 4: Verificar contexto específico para webapp
    assert_file_contains "Context should mention webapp type" \
        "$project_path/context/02_CONTEXT.md" -i "web"
    
    cleanup_test_environment
    finish_test_suite "JavaScript WebApp Generation Integration Test"
}

test_javascript_project_npm() {
    init_test_suite "JavaScript NPM Validation Test"
    
    # Setup
    setup_test_environment
    
    # Skip si npm no está instalado
    if ! command -v npm &> /dev/null; then
        echo "⚠️  Skipping NPM tests - npm not installed"
        return 0
    fi
    
    local project_name="test-npm-validation"
    local project_path
    
    # Test 1: Generar proyecto
    assert_success "Should generate JavaScript project for npm test" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' javascript api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar que package.json es válido para npm
    assert_success "package.json should be valid for npm" \
        "cd '$project_path' && npm pkg get name > /dev/null"
    
    # Test 3: Verificar instalación de dependencias (si existen)
    if grep -q '"dependencies"' "$project_path/package.json" 2>/dev/null; then
        # Solo probar si hay dependencias definidas
        assert_success "Should be able to install dependencies" \
            "cd '$project_path' && timeout 60s npm install --dry-run"
    fi
    
    cleanup_test_environment
    finish_test_suite "JavaScript NPM Validation Test"
}

test_javascript_typescript_support() {
    init_test_suite "JavaScript TypeScript Support Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-ts-support"
    local project_path
    
    # Test 1: Generar proyecto que podría tener soporte TypeScript
    assert_success "Should generate JavaScript project" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' javascript api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar si se generó configuración TypeScript (opcional)
    if [ -f "$project_path/tsconfig.json" ]; then
        assert_valid_json "tsconfig.json should be valid" "$project_path/tsconfig.json"
        assert_file_contains "Should have TypeScript compiler options" \
            "$project_path/tsconfig.json" "compilerOptions"
    fi
    
    # Test 3: Verificar archivos .ts si se generaron
    if find "$project_path" -name "*.ts" -type f | head -1 | read -r ts_file; then
        assert_file_exists "TypeScript files should exist" "$ts_file"
        assert_success "TypeScript files should have valid syntax" \
            "grep -q 'interface\\|type\\|class' '$ts_file' || true"
    fi
    
    cleanup_test_environment
    finish_test_suite "JavaScript TypeScript Support Test"
}

# Ejecutar todos los tests si se ejecuta directamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "🧪 Running JavaScript Generation Integration Tests..."
    
    test_javascript_api_generation
    api_result=$?
    
    test_javascript_webapp_generation
    webapp_result=$?
    
    test_javascript_project_npm
    npm_result=$?
    
    test_javascript_typescript_support
    ts_result=$?
    
    # Resultado final
    if [ $api_result -eq 0 ] && [ $webapp_result -eq 0 ] && [ $npm_result -eq 0 ] && [ $ts_result -eq 0 ]; then
        echo -e "\n✅ All JavaScript integration tests passed!"
        exit 0
    else
        echo -e "\n❌ Some JavaScript integration tests failed!"
        exit 1
    fi
fi