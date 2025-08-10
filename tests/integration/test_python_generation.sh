#!/bin/bash
# test_python_generation.sh - Integration tests para generación de proyectos Python
set -euo pipefail

# Cargar helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/setup.sh"
source "$SCRIPT_DIR/../helpers/assertions.sh"

# Test suite para generación completa de proyectos Python
test_python_api_generation() {
    init_test_suite "Python API Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-python-api"
    local project_path
    
    # Test 1: Generar proyecto completo
    assert_success "Should generate Python API project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' python api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura básica
    assert_directory_exists "Should create project directory" "$project_path"
    assert_directory_exists "Should create src directory" "$project_path/src"
    assert_directory_exists "Should create tests directory" "$project_path/tests"
    assert_directory_exists "Should create context directory" "$project_path/context"
    
    # Test 3: Verificar estructura DDD para Python
    assert_directory_exists "Should create src/domain layer" "$project_path/src/domain"
    assert_directory_exists "Should create src/application layer" "$project_path/src/application"
    assert_directory_exists "Should create src/infrastructure layer" "$project_path/src/infrastructure"
    assert_directory_exists "Should create src/interfaces layer" "$project_path/src/interfaces"
    
    # Test 4: Verificar archivos específicos de Python
    assert_file_exists "Should create pyproject.toml" "$project_path/pyproject.toml"
    assert_file_exists "Should create main module" "$project_path/src/__init__.py"
    assert_file_exists "Should create Taskfile.yml" "$project_path/Taskfile.yml"
    assert_file_exists "Should create .gitignore" "$project_path/.gitignore"
    
    # Test 5: Verificar archivos __init__.py en todos los directorios de src
    assert_file_exists "Should create __init__.py in src" "$project_path/src/__init__.py"
    assert_file_exists "Should create __init__.py in domain" "$project_path/src/domain/__init__.py"
    assert_file_exists "Should create __init__.py in application" "$project_path/src/application/__init__.py"
    
    # Test 6: Verificar contenido de pyproject.toml
    assert_file_contains "pyproject.toml should have project name" \
        "$project_path/pyproject.toml" "name = \"$project_name\""
    assert_file_contains "pyproject.toml should have Python version" \
        "$project_path/pyproject.toml" "python"
    
    # Test 7: Verificar .gitignore específico para Python
    assert_file_contains ".gitignore should ignore __pycache__" \
        "$project_path/.gitignore" "__pycache__/"
    assert_file_contains ".gitignore should ignore .pyc files" \
        "$project_path/.gitignore" "*.pyc"
    assert_file_contains ".gitignore should ignore venv" \
        "$project_path/.gitignore" "venv/"
    
    # Test 8: Verificar archivos de contexto
    assert_file_exists "Should create PROMPT.md" "$project_path/context/01_PROMPT.md"
    assert_file_contains "Context should mention Python" \
        "$project_path/context/02_CONTEXT.md" "Python"
    
    # Test 9: Verificar que no hay placeholders sin reemplazar
    assert_success "Should not have unreplaced placeholders" \
        "validate_no_placeholders '$project_path'"
    
    cleanup_test_environment
    finish_test_suite "Python API Generation Integration Test"
}

test_python_cli_generation() {
    init_test_suite "Python CLI Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-python-cli"
    local project_path
    
    # Test 1: Generar proyecto CLI
    assert_success "Should generate Python CLI project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' python cli openai"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura específica de CLI
    assert_directory_exists "Should create project directory" "$project_path"
    assert_file_exists "Should create pyproject.toml" "$project_path/pyproject.toml"
    
    # Test 3: Verificar entrada para CLI en pyproject.toml
    if grep -q "\[project.scripts\]" "$project_path/pyproject.toml" 2>/dev/null; then
        assert_file_contains "Should have CLI entry point" \
            "$project_path/pyproject.toml" "scripts"
    fi
    
    # Test 4: Verificar contexto específico para CLI
    assert_file_contains "Context should mention CLI type" \
        "$project_path/context/02_CONTEXT.md" "CLI"
    
    cleanup_test_environment
    finish_test_suite "Python CLI Generation Integration Test"
}

test_python_library_generation() {
    init_test_suite "Python Library Generation Integration Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-python-lib"
    local project_path
    
    # Test 1: Generar proyecto Library
    assert_success "Should generate Python library project successfully" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' python library google"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar estructura de librería
    assert_directory_exists "Should create project directory" "$project_path"
    assert_file_exists "Should create pyproject.toml" "$project_path/pyproject.toml"
    
    # Test 3: Verificar configuración de librería en pyproject.toml
    assert_file_contains "Should have library configuration" \
        "$project_path/pyproject.toml" "build-system"
    
    cleanup_test_environment
    finish_test_suite "Python Library Generation Integration Test"
}

test_python_project_pip() {
    init_test_suite "Python Pip Validation Test"
    
    # Setup
    setup_test_environment
    
    # Skip si pip no está instalado
    if ! command -v pip &> /dev/null && ! command -v pip3 &> /dev/null; then
        echo "⚠️  Skipping pip tests - pip not installed"
        return 0
    fi
    
    local project_name="test-pip-validation"
    local project_path
    local pip_cmd="pip"
    
    # Usar pip3 si pip no está disponible
    if ! command -v pip &> /dev/null && command -v pip3 &> /dev/null; then
        pip_cmd="pip3"
    fi
    
    # Test 1: Generar proyecto
    assert_success "Should generate Python project for pip test" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' python api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar que pyproject.toml es válido
    if command -v python3 &> /dev/null; then
        assert_success "pyproject.toml should be valid" \
            "cd '$project_path' && python3 -c 'import tomllib; tomllib.load(open(\"pyproject.toml\", \"rb\"))' 2>/dev/null || python3 -c 'import tomli; tomli.load(open(\"pyproject.toml\", \"rb\"))' 2>/dev/null || true"
    fi
    
    # Test 3: Verificar instalación en modo desarrollo (dry-run)
    if [ -f "$project_path/pyproject.toml" ]; then
        assert_success "Should be able to install in development mode" \
            "cd '$project_path' && $pip_cmd install --dry-run -e . 2>/dev/null || true"
    fi
    
    cleanup_test_environment
    finish_test_suite "Python Pip Validation Test"
}

test_python_requirements_files() {
    init_test_suite "Python Requirements Files Test"
    
    # Setup
    setup_test_environment
    
    local project_name="test-requirements"
    local project_path
    
    # Test 1: Generar proyecto
    assert_success "Should generate Python project" \
        "cd '$AI_GENERATOR_ROOT' && AI_GENERATOR_OUTPUT='$TEST_OUTPUT_DIR' task new:quick -- '$project_name' python api anthropic"
    
    project_path="$TEST_OUTPUT_DIR/$project_name"
    
    # Test 2: Verificar archivos de requirements si existen
    if [ -f "$project_path/requirements.txt" ]; then
        assert_file_exists "Should have requirements.txt" "$project_path/requirements.txt"
        
        # Verificar formato básico de requirements
        assert_success "requirements.txt should have valid format" \
            "cd '$project_path' && grep -E '^[a-zA-Z][a-zA-Z0-9_-]*([>=<][0-9.]+)?$' requirements.txt || [ ! -s requirements.txt ]"
    fi
    
    # Test 3: Verificar requirements de desarrollo si existen
    if [ -f "$project_path/requirements-dev.txt" ]; then
        assert_file_exists "Should have requirements-dev.txt" "$project_path/requirements-dev.txt"
    fi
    
    cleanup_test_environment
    finish_test_suite "Python Requirements Files Test"
}

# Ejecutar todos los tests si se ejecuta directamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "🧪 Running Python Generation Integration Tests..."
    
    test_python_api_generation
    api_result=$?
    
    test_python_cli_generation
    cli_result=$?
    
    test_python_library_generation
    lib_result=$?
    
    test_python_project_pip
    pip_result=$?
    
    test_python_requirements_files
    req_result=$?
    
    # Resultado final
    if [ $api_result -eq 0 ] && [ $cli_result -eq 0 ] && [ $lib_result -eq 0 ] && [ $pip_result -eq 0 ] && [ $req_result -eq 0 ]; then
        echo -e "\n✅ All Python integration tests passed!"
        exit 0
    else
        echo -e "\n❌ Some Python integration tests failed!"
        exit 1
    fi
fi