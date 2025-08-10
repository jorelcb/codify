#!/bin/bash
# test_wizard.sh - Unit tests para wizard.sh
set -euo pipefail

# Cargar helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/setup.sh"
source "$SCRIPT_DIR/../helpers/assertions.sh"

# Test suite para wizard functionality
test_wizard_validation() {
    init_test_suite "Wizard Validation Tests"
    
    # Setup
    setup_test_environment
    
    # Mock del wizard script con funciones específicas para testing
    create_wizard_test_functions
    
    # Test 1: Validar nombre de proyecto válido
    assert_success "Valid project name should pass validation" \
        "validate_project_name 'my-test-project'"
    
    # Test 2: Validar nombre de proyecto inválido
    assert_failure "Invalid project name should fail validation" \
        "validate_project_name '123-invalid'"
    
    # Test 3: Validar nombre de proyecto vacío
    assert_failure "Empty project name should fail validation" \
        "validate_project_name ''"
    
    # Test 4: Validar nombre con caracteres especiales
    assert_failure "Project name with special chars should fail validation" \
        "validate_project_name 'my@project'"
    
    # Test 5: Validar lenguaje soportado
    assert_success "Supported language should pass validation" \
        "validate_language 'go'"
    
    # Test 6: Validar lenguaje no soportado
    assert_failure "Unsupported language should fail validation" \
        "validate_language 'cobol'"
    
    # Test 7: Validar tipo de proyecto soportado
    assert_success "Supported project type should pass validation" \
        "validate_project_type 'api'"
    
    # Test 8: Validar tipo de proyecto no soportado
    assert_failure "Unsupported project type should fail validation" \
        "validate_project_type 'blockchain'"
    
    # Test 9: Validar proveedor de IA soportado
    assert_success "Supported AI provider should pass validation" \
        "validate_ai_provider 'openai'"
    
    # Test 10: Validar proveedor de IA no soportado
    assert_failure "Unsupported AI provider should fail validation" \
        "validate_ai_provider 'myai'"
    
    # Cleanup
    cleanup_test_environment
    
    finish_test_suite "Wizard Validation Tests"
}

test_json_generation() {
    init_test_suite "JSON Generation Tests"
    
    # Setup
    setup_test_environment
    create_wizard_test_functions
    
    local test_json_file="$TEST_OUTPUT_DIR/test_config.json"
    
    # Test 1: Generar JSON con datos válidos
    generate_test_json "test-project" "go" "api" "openai" "Test project description" > "$test_json_file"
    assert_valid_json "Generated JSON should be valid" "$test_json_file"
    
    # Test 2: Verificar campos requeridos en JSON
    assert_file_contains "JSON should contain project_name" "$test_json_file" '"project_name": "test-project"'
    assert_file_contains "JSON should contain language" "$test_json_file" '"language": "go"'
    assert_file_contains "JSON should contain project_type" "$test_json_file" '"project_type": "api"'
    assert_file_contains "JSON should contain ai_provider" "$test_json_file" '"ai_provider": "openai"'
    assert_file_contains "JSON should contain description" "$test_json_file" '"description": "Test project description"'
    
    # Test 3: Verificar que capabilities es un array
    assert_file_contains "JSON should have capabilities array" "$test_json_file" '"capabilities": \['
    
    # Test 4: Verificar timestamp format
    local timestamp_pattern='"timestamp": "[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z"'
    assert_success "JSON should contain valid timestamp" \
        "grep -E '$timestamp_pattern' '$test_json_file'"
    
    cleanup_test_environment
    finish_test_suite "JSON Generation Tests"
}

# Funciones auxiliares para testing
create_wizard_test_functions() {
    # Simular funciones del wizard para testing
    validate_project_name() {
        local name="$1"
        [[ -n "$name" ]] && [[ "$name" =~ ^[a-zA-Z][a-zA-Z0-9_-]*$ ]]
    }
    
    validate_language() {
        local lang="$1"
        echo "go javascript python java rust" | grep -qw "$lang"
    }
    
    validate_project_type() {
        local type="$1"
        echo "api cli library webapp service" | grep -qw "$type"
    }
    
    validate_ai_provider() {
        local provider="$1"
        echo "openai anthropic google azure local none" | grep -qw "$provider"
    }
    
    generate_test_json() {
        local project_name="$1"
        local language="$2"
        local project_type="$3"
        local ai_provider="$4"
        local description="$5"
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        
        cat << EOF
{
    "project_name": "$project_name",
    "language": "$language",
    "project_type": "$project_type",
    "ai_provider": "$ai_provider",
    "description": "$description",
    "capabilities": [],
    "include_tests": true,
    "include_docker": true,
    "include_ci": false,
    "timestamp": "$timestamp"
}
EOF
    }
}

# Ejecutar todos los tests si se ejecuta directamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "🧪 Running Wizard Unit Tests..."
    
    test_wizard_validation
    wizard_validation_result=$?
    
    test_json_generation  
    json_generation_result=$?
    
    # Resultado final
    if [ $wizard_validation_result -eq 0 ] && [ $json_generation_result -eq 0 ]; then
        echo -e "\n✅ All wizard unit tests passed!"
        exit 0
    else
        echo -e "\n❌ Some wizard unit tests failed!"
        exit 1
    fi
fi