#!/bin/bash
# run_all_tests.sh - Test runner principal para AI Context Generator
set -euo pipefail

# Colores para output
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Variables globales
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Cargar helpers
source "$SCRIPT_DIR/helpers/setup.sh"

# Función para ejecutar una suite de tests
run_test_suite() {
    local test_file="$1"
    local test_name="$2"
    
    echo -e "${BLUE}🧪 Running $test_name...${NC}"
    
    if [ -f "$test_file" ] && [ -x "$test_file" ]; then
        if "$test_file"; then
            echo -e "${GREEN}✅ $test_name PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}❌ $test_name FAILED${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  $test_name SKIPPED (file not found or not executable)${NC}"
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        return 0
    fi
}

# Función para mostrar resumen final
show_summary() {
    echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}"
    echo -e "Total test suites: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo -e "${YELLOW}Skipped: $SKIPPED_TESTS${NC}"
    
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    fi
    
    echo -e "Success rate: ${success_rate}%"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed!${NC}"
        return 0
    else
        echo -e "\n${RED}💥 Some tests failed!${NC}"
        return 1
    fi
}

# Función principal
main() {
    echo -e "${BLUE}🚀 AI Context Generator - Test Suite Runner${NC}"
    echo -e "Running all tests...\n"
    
    # Verificar entorno de testing
    setup_test_environment > /dev/null 2>&1 || {
        echo -e "${RED}❌ Failed to setup test environment${NC}"
        exit 1
    }
    
    # Hacer ejecutables todos los scripts de test
    find "$SCRIPT_DIR" -name "test_*.sh" -type f -exec chmod +x {} \;
    
    # Ejecutar tests unitarios
    echo -e "${YELLOW}=== UNIT TESTS ===${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    run_test_suite "$SCRIPT_DIR/unit/test_wizard.sh" "Wizard Unit Tests"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    run_test_suite "$SCRIPT_DIR/unit/test_scaffolding.sh" "Scaffolding Unit Tests"
    
    echo ""
    
    # Ejecutar tests de integración
    echo -e "${YELLOW}=== INTEGRATION TESTS ===${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    run_test_suite "$SCRIPT_DIR/integration/test_go_generation.sh" "Go Generation Integration Tests"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    run_test_suite "$SCRIPT_DIR/integration/test_js_generation.sh" "JavaScript Generation Integration Tests"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    run_test_suite "$SCRIPT_DIR/integration/test_python_generation.sh" "Python Generation Integration Tests"
    
    echo ""
    
    # Mostrar resumen
    show_summary
}

# Manejar argumentos de línea de comandos
case "${1:-all}" in
    "unit")
        echo -e "${BLUE}Running only unit tests...${NC}"
        setup_test_environment > /dev/null 2>&1
        find "$SCRIPT_DIR/unit" -name "test_*.sh" -type f -exec chmod +x {} \;
        
        run_test_suite "$SCRIPT_DIR/unit/test_wizard.sh" "Wizard Unit Tests"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        run_test_suite "$SCRIPT_DIR/unit/test_scaffolding.sh" "Scaffolding Unit Tests"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        show_summary
        ;;
    "integration")
        echo -e "${BLUE}Running only integration tests...${NC}"
        setup_test_environment > /dev/null 2>&1
        find "$SCRIPT_DIR/integration" -name "test_*.sh" -type f -exec chmod +x {} \;
        
        run_test_suite "$SCRIPT_DIR/integration/test_go_generation.sh" "Go Generation Tests"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        run_test_suite "$SCRIPT_DIR/integration/test_js_generation.sh" "JavaScript Generation Tests"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        run_test_suite "$SCRIPT_DIR/integration/test_python_generation.sh" "Python Generation Tests"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        show_summary
        ;;
    "help")
        echo "Usage: $0 [unit|integration|all|help]"
        echo ""
        echo "Options:"
        echo "  unit         Run only unit tests"
        echo "  integration  Run only integration tests"
        echo "  all          Run all tests (default)"
        echo "  help         Show this help message"
        exit 0
        ;;
    "all"|*)
        main
        ;;
esac