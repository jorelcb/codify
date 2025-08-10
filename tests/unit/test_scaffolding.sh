#!/bin/bash
# test_scaffolding.sh - Unit tests para generate_scaffolding.sh
set -euo pipefail

# Cargar helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/setup.sh"
source "$SCRIPT_DIR/../helpers/assertions.sh"

# Test suite para scaffolding functionality
test_directory_creation() {
    init_test_suite "Directory Creation Tests"
    
    # Setup
    setup_test_environment
    
    local test_project_dir="$TEST_OUTPUT_DIR/test-project"
    
    # Test 1: Crear estructura básica de directorios
    create_basic_structure "$test_project_dir" "go"
    assert_directory_exists "Should create project root" "$test_project_dir"
    assert_directory_exists "Should create context directory" "$test_project_dir/context"
    assert_directory_exists "Should create internal directory" "$test_project_dir/internal"
    
    # Test 2: Crear estructura específica de Go
    create_go_structure "$test_project_dir"
    assert_directory_exists "Should create cmd directory for Go" "$test_project_dir/cmd"
    assert_directory_exists "Should create internal/domain" "$test_project_dir/internal/domain"
    assert_directory_exists "Should create internal/application" "$test_project_dir/internal/application"
    assert_directory_exists "Should create internal/infrastructure" "$test_project_dir/internal/infrastructure"
    assert_directory_exists "Should create internal/interfaces" "$test_project_dir/internal/interfaces"
    
    # Test 3: Crear estructura específica de JavaScript
    local js_project_dir="$TEST_OUTPUT_DIR/js-project"
    create_basic_structure "$js_project_dir" "javascript"
    create_javascript_structure "$js_project_dir"
    assert_directory_exists "Should create src directory for JavaScript" "$js_project_dir/src"
    assert_directory_exists "Should create src/domain" "$js_project_dir/src/domain"
    assert_directory_exists "Should create src/application" "$js_project_dir/src/application"
    
    # Test 4: Crear estructura específica de Python
    local py_project_dir="$TEST_OUTPUT_DIR/py-project"
    create_basic_structure "$py_project_dir" "python"
    create_python_structure "$py_project_dir"
    assert_directory_exists "Should create src directory for Python" "$py_project_dir/src"
    assert_directory_exists "Should create tests directory for Python" "$py_project_dir/tests"
    
    cleanup_test_environment
    finish_test_suite "Directory Creation Tests"
}

test_file_generation() {
    init_test_suite "File Generation Tests"
    
    # Setup
    setup_test_environment
    
    local test_project_dir="$TEST_OUTPUT_DIR/file-test-project"
    create_basic_structure "$test_project_dir" "go"
    
    # Test 1: Generar go.mod para proyectos Go
    generate_go_mod "$test_project_dir" "file-test-project"
    assert_file_exists "Should create go.mod file" "$test_project_dir/go.mod"
    assert_file_contains "Should have correct module name in go.mod" \
        "$test_project_dir/go.mod" "module file-test-project"
    assert_file_contains "Should have Go version in go.mod" \
        "$test_project_dir/go.mod" "go 1.21"
    
    # Test 2: Generar main.go
    generate_go_main "$test_project_dir" "file-test-project"
    assert_file_exists "Should create main.go file" "$test_project_dir/cmd/file-test-project/main.go"
    assert_file_contains "Should have package main in main.go" \
        "$test_project_dir/cmd/file-test-project/main.go" "package main"
    
    # Test 3: Generar .gitignore
    generate_gitignore "$test_project_dir" "go"
    assert_file_exists "Should create .gitignore file" "$test_project_dir/.gitignore"
    assert_file_contains "Should ignore Go binaries in .gitignore" \
        "$test_project_dir/.gitignore" "*.exe"
    
    # Test 4: Generar README.md
    generate_readme "$test_project_dir" "file-test-project" "go"
    assert_file_exists "Should create README.md file" "$test_project_dir/README.md"
    assert_file_contains "Should have project name in README" \
        "$test_project_dir/README.md" "file-test-project"
    
    cleanup_test_environment
    finish_test_suite "File Generation Tests"
}

test_template_processing() {
    init_test_suite "Template Processing Tests"
    
    # Setup
    setup_test_environment
    
    # Test 1: Reemplazo de variables simples
    local test_content="Project name: {{PROJECT_NAME}}, Language: {{LANGUAGE}}"
    local expected="Project name: my-project, Language: go"
    local actual=$(process_template_variables "$test_content" "my-project" "go" "api" "openai")
    assert_equals "Should replace template variables correctly" "$expected" "$actual"
    
    # Test 2: Reemplazo de múltiples variables
    local multi_content="{{PROJECT_NAME}} is a {{PROJECT_TYPE}} in {{LANGUAGE}} for {{AI_PROVIDER}}"
    local multi_expected="test-app is a api in go for openai"
    local multi_actual=$(process_template_variables "$multi_content" "test-app" "go" "api" "openai")
    assert_equals "Should replace multiple variables" "$multi_expected" "$multi_actual"
    
    # Test 3: Variables que no existen deben permanecer
    local unchanged_content="{{PROJECT_NAME}} has {{UNKNOWN_VAR}}"
    local unchanged_actual=$(process_template_variables "$unchanged_content" "test" "go" "api" "openai")
    assert_file_contains "Should keep unknown variables unchanged" \
        <(echo "$unchanged_actual") "{{UNKNOWN_VAR}}"
    
    cleanup_test_environment
    finish_test_suite "Template Processing Tests"
}

# Funciones auxiliares para testing
create_basic_structure() {
    local project_dir="$1"
    local language="$2"
    
    mkdir -p "$project_dir"/{context,internal,docs}
}

create_go_structure() {
    local project_dir="$1"
    
    mkdir -p "$project_dir"/cmd
    mkdir -p "$project_dir"/internal/{domain,application,infrastructure,interfaces}
    mkdir -p "$project_dir"/pkg
    mkdir -p "$project_dir"/test
    mkdir -p "$project_dir"/deployments
}

create_javascript_structure() {
    local project_dir="$1"
    
    mkdir -p "$project_dir"/src/{domain,application,infrastructure,interfaces}
    mkdir -p "$project_dir"/{tests,docs}
}

create_python_structure() {
    local project_dir="$1"
    
    mkdir -p "$project_dir"/src
    mkdir -p "$project_dir"/{tests,docs}
}

generate_go_mod() {
    local project_dir="$1"
    local project_name="$2"
    
    cat > "$project_dir/go.mod" << EOF
module $project_name

go 1.21

require (
    github.com/gorilla/mux v1.8.0
)
EOF
}

generate_go_main() {
    local project_dir="$1"
    local project_name="$2"
    
    mkdir -p "$project_dir/cmd/$project_name"
    cat > "$project_dir/cmd/$project_name/main.go" << EOF
package main

import (
    "fmt"
    "log"
    "net/http"
    
    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/health", healthHandler).Methods("GET")
    
    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
EOF
}

generate_gitignore() {
    local project_dir="$1"
    local language="$2"
    
    case "$language" in
        "go")
            cat > "$project_dir/.gitignore" << EOF
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work
EOF
            ;;
        "javascript")
            cat > "$project_dir/.gitignore" << EOF
# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Runtime data
pids
*.pid
*.seed
*.pid.lock

# Coverage directory
coverage/

# Build output
dist/
build/
EOF
            ;;
    esac
}

generate_readme() {
    local project_dir="$1"
    local project_name="$2"
    local language="$3"
    
    cat > "$project_dir/README.md" << EOF
# $project_name

A $language project generated with AI Context Generator.

## Getting Started

\`\`\`bash
task dev
\`\`\`

## Architecture

This project follows Domain-Driven Design (DDD) principles with Clean Architecture.

## Development

\`\`\`bash
# Run tests
task test

# Build
task build

# Run
task run
\`\`\`
EOF
}

process_template_variables() {
    local content="$1"
    local project_name="$2"
    local language="$3"
    local project_type="$4"
    local ai_provider="$5"
    
    echo "$content" | sed \
        -e "s/{{PROJECT_NAME}}/$project_name/g" \
        -e "s/{{LANGUAGE}}/$language/g" \
        -e "s/{{PROJECT_TYPE}}/$project_type/g" \
        -e "s/{{AI_PROVIDER}}/$ai_provider/g"
}

# Ejecutar todos los tests si se ejecuta directamente
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "🧪 Running Scaffolding Unit Tests..."
    
    test_directory_creation
    directory_result=$?
    
    test_file_generation
    file_result=$?
    
    test_template_processing
    template_result=$?
    
    # Resultado final
    if [ $directory_result -eq 0 ] && [ $file_result -eq 0 ] && [ $template_result -eq 0 ]; then
        echo -e "\n✅ All scaffolding unit tests passed!"
        exit 0
    else
        echo -e "\n❌ Some scaffolding unit tests failed!"
        exit 1
    fi
fi