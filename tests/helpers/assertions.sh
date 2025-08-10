#!/bin/bash
# assertions.sh - Utilidades de testing para AI Context Generator
set -euo pipefail

# Colores para output
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

# Contadores globales
TEST_COUNT=0
PASSED_COUNT=0
FAILED_COUNT=0

# Función para inicializar suite de tests
init_test_suite() {
    local suite_name="$1"
    echo -e "${YELLOW}=== Running Test Suite: $suite_name ===${NC}"
    TEST_COUNT=0
    PASSED_COUNT=0
    FAILED_COUNT=0
}

# Función para finalizar suite de tests
finish_test_suite() {
    local suite_name="$1"
    echo -e "${YELLOW}=== Test Suite Results: $suite_name ===${NC}"
    echo -e "Total tests: $TEST_COUNT"
    echo -e "${GREEN}Passed: $PASSED_COUNT${NC}"
    echo -e "${RED}Failed: $FAILED_COUNT${NC}"
    
    if [ $FAILED_COUNT -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}❌ Some tests failed!${NC}"
        return 1
    fi
}

# Assert que un comando ejecuta exitosamente
assert_success() {
    local description="$1"
    local command="$2"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   Command: $command"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un comando falla
assert_failure() {
    local description="$1"
    local command="$2"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if ! eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   Command should have failed: $command"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un archivo existe
assert_file_exists() {
    local description="$1"
    local file_path="$2"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ -f "$file_path" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   File does not exist: $file_path"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un directorio existe
assert_directory_exists() {
    local description="$1"
    local dir_path="$2"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ -d "$dir_path" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   Directory does not exist: $dir_path"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un archivo contiene texto específico
assert_file_contains() {
    local description="$1"
    local file_path="$2"
    local expected_text="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ -f "$file_path" ] && grep -q "$expected_text" "$file_path"; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   File does not contain expected text: $expected_text"
        echo -e "   File: $file_path"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un archivo NO contiene texto específico
assert_file_not_contains() {
    local description="$1"
    local file_path="$2"
    local unexpected_text="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ -f "$file_path" ] && ! grep -q "$unexpected_text" "$file_path"; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   File contains unexpected text: $unexpected_text"
        echo -e "   File: $file_path"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que dos strings son iguales
assert_equals() {
    local description="$1"
    local expected="$2"
    local actual="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   Expected: $expected"
        echo -e "   Actual: $actual"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Assert que un JSON es válido
assert_valid_json() {
    local description="$1"
    local json_file="$2"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ -f "$json_file" ] && jq empty "$json_file" 2>/dev/null; then
        echo -e "${GREEN}✅ PASS${NC}: $description"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $description"
        echo -e "   Invalid JSON file: $json_file"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        return 1
    fi
}

# Función para limpiar archivos temporales de test
cleanup_test_files() {
    local test_output_dir="$1"
    if [ -d "$test_output_dir" ]; then
        rm -rf "$test_output_dir"
    fi
}

# Función para crear directorio temporal de test
create_test_output_dir() {
    local base_name="$1"
    local test_dir="/tmp/ai-context-generator-test-$base_name-$$"
    mkdir -p "$test_dir"
    echo "$test_dir"
}