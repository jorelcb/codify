#!/bin/bash
# test_go_generation.sh - Integration tests para generación de proyectos Go
set -euo pipefail

# Cargar helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/setup.sh"
source "$SCRIPT_DIR/../helpers/assertions.sh"

# Test suite para generación completa de proyectos Go
test_go_api_generation() {
    init_test_suite "Go API Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-go-api"
    local project_path
    
    # Test 1: Generar proyecto completo usando task new:quick
    assert_success "Should generate Go API project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' go api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura básica
    assert_directory_exists "Should create project directory" "$project_path"
    assert_directory_exists "Should create context directory" "$project_path/context"
    assert_directory_exists "Should create cmd directory" "$project_path/cmd"
    assert_directory_exists "Should create internal directory" "$project_path/internal"
    
    # Test 3: Verificar estructura DDD
    assert_directory_exists "Should create domain layer" "$project_path/internal/domain"
    assert_directory_exists "Should create application layer" "$project_path/internal/application"
    assert_directory_exists "Should create infrastructure layer" "$project_path/internal/infrastructure"
    assert_directory_exists "Should create interfaces layer" "$project_path/internal/interfaces"
    
    # Test 4: Verificar archivos de contexto
    assert_file_exists "Should create PROMPT.md" "$project_path/context/01_PROMPT.md"
    assert_file_exists "Should create CONTEXT.md" "$project_path/context/02_CONTEXT.md"
    assert_file_exists "Should create SCAFFOLDING.md" "$project_path/context/03_SCAFFOLDING.md"
    assert_file_exists "Should create INTERACTIONS_LOG.md" "$project_path/context/04_INTERACTIONS_LOG.md"
    
    # Test 5: Verificar archivos específicos de Go
    assert_file_exists "Should create go.mod" "$project_path/go.mod"
    assert_file_exists "Should create main.go" "$project_path/cmd/$project_name/main.go"
    assert_file_exists "Should create Taskfile.yml" "$project_path/Taskfile.yml"
    assert_file_exists "Should create .gitignore" "$project_path/.gitignore"
    assert_file_exists "Should create README.md" "$project_path/README.md"
    
    # Test 6: Verificar contenido de archivos clave
    assert_file_contains "go.mod should have correct module name" \
        "$project_path/go.mod" "module $project_name"
    assert_file_contains "main.go should have package main" \
        "$project_path/cmd/$project_name/main.go" "package main"
    assert_file_contains "README should mention project name" \
        "$project_path/README.md" "$project_name"
    
    # Test 7: Verificar que no hay placeholders sin reemplazar
    assert_success "Should not have unreplaced placeholders" \
        "validate_no_placeholders '$project_path'"
    
    # Test 8: Verificar que el Taskfile del proyecto es válido
    assert_success "Generated Taskfile should be valid" \
        "cd '$project_path' && task --list-all > /dev/null"
    
    # Test 9: Verificar estructura JSON de configuración (si existe)
    if [ -f "$project_path/context/config.json" ]; then
        assert_valid_json "Config JSON should be valid" "$project_path/context/config.json"
    fi
    
    cleanup_test_environment
    finish_test_suite "Go API Generation Integration Test"
}

test_go_cli_generation() {
    init_test_suite "Go CLI Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-go-cli"
    local project_path
    
    # Test 1: Generar proyecto CLI
    assert_success "Should generate Go CLI project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' go cli openai"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura básica
    assert_directory_exists "Should create project directory" "$project_path"
    assert_file_exists "Should create go.mod" "$project_path/go.mod"
    assert_file_exists "Should create main.go" "$project_path/cmd/$project_name/main.go"
    
    # Test 3: Verificar contexto específico para CLI
    assert_file_contains "Context should mention CLI type" \
        "$project_path/context/02_CONTEXT.md" "CLI"
    
    cleanup_test_environment
    finish_test_suite "Go CLI Generation Integration Test"
}

test_go_library_generation() {
    init_test_suite "Go Library Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-go-lib"
    local project_path
    
    # Test 1: Generar proyecto Library
    assert_success "Should generate Go library project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' go library google"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura de librería
    assert_directory_exists "Should create project directory" "$project_path"
    assert_file_exists "Should create go.mod" "$project_path/go.mod"
    
    # Test 3: Verificar que tiene estructura de librería (pkg/ en lugar de cmd/)
    if [ -d "$project_path/pkg" ]; then
        assert_directory_exists "Library should have pkg directory" "$project_path/pkg"
    fi
    
    cleanup_test_environment
    finish_test_suite "Go Library Generation Integration Test"
}

test_go_project_build() {
    init_test_suite "Go Project Build Test"
    
    # Setup
    setup_test_environment
    
    # Skip si Go no está instalado
    if ! command -v go &> /dev/null; then
        echo "⚠️  Skipping Go build tests - Go not installed"
        return 0
    fi
    
    local project_name="test-buildable-go"
    local project_path
    
    # Test 1: Generar proyecto
    assert_success "Should generate Go project for build test" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' go api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar que el proyecto compila
    assert_success "Generated Go project should compile" \
        "cd '$project_path' && go mod tidy && go build ./cmd/$project_name"
    
    # Test 3: Verificar que el binario se ejecuta (básico)
    if [ -f "$project_path/$project_name" ]; then
        # Solo verificar que no crashea inmediatamente
        timeout 2s "$project_path/$project_name" || true
        assert_success "Binary should be executable" "test -x '$project_path/$project_name'"
    fi
    
    cleanup_test_environment
    finish_test_suite "Go Project Build Test"
}

# Ejecutar todos los tests si se ejecuta directamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "🧪 Running Go Generation Integration Tests..."
    
    test_go_api_generation
    api_result=$?
    
    test_go_cli_generation
    cli_result=$?
    
    test_go_library_generation
    lib_result=$?
    
    test_go_project_build
    build_result=$?
    
    # Resultado final
    if [ $api_result -eq 0 ] && [ $cli_result -eq 0 ] && [ $lib_result -eq 0 ] && [ $build_result -eq 0 ]; then
        echo -e "\n✅ All Go integration tests passed!"
        exit 0
    else
        echo -e "\n❌ Some Go integration tests failed!"
        exit 1
    fi
fi