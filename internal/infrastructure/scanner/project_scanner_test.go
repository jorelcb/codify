package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan_GoProject(t *testing.T) {
	dir := t.TempDir()

	// Create go.mod
	goMod := `module github.com/example/my-api

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/text v0.14.0 // indirect
)
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "Go", result.Language)
	assert.Equal(t, "Gin (HTTP)", result.Framework)
	assert.Contains(t, result.Dependencies, "github.com/gin-gonic/gin")
	assert.Contains(t, result.Dependencies, "github.com/spf13/cobra")
	// indirect deps should be excluded
	assert.NotContains(t, result.Dependencies, "golang.org/x/text")
}

func TestScan_PythonProject(t *testing.T) {
	dir := t.TempDir()

	requirements := `fastapi>=0.100.0
uvicorn>=0.23.0
sqlalchemy>=2.0
pytest>=7.0
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "Python", result.Language)
	assert.Equal(t, "FastAPI", result.Framework)
	assert.Contains(t, result.Dependencies, "fastapi")
	assert.Contains(t, result.Dependencies, "uvicorn")
}

func TestScan_JSProject(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{
  "name": "my-app",
  "dependencies": {
    "next": "^14.0.0",
    "react": "^18.2.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "JavaScript/TypeScript", result.Language)
	assert.Equal(t, "Next.js", result.Framework)
	assert.Contains(t, result.Dependencies, "next")
	assert.Contains(t, result.Dependencies, "react")
	assert.Contains(t, result.Dependencies, "typescript")
}

func TestScan_DirectoryTree(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure
	dirs := []string{
		"cmd/api",
		"internal/domain",
		"internal/application",
		"tests",
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0755))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cmd/api/main.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.DirectoryTree, "cmd/")
	assert.Contains(t, result.DirectoryTree, "internal/")
	assert.Contains(t, result.DirectoryTree, "domain/")
}

func TestScan_README(t *testing.T) {
	dir := t.TempDir()

	readme := "# My Project\n\nThis is a test project.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.README, "# My Project")
	assert.Contains(t, result.README, "test project")
}

func TestScan_ExistingContext(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# My Agent Context"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude Instructions"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.ExistingContext, "AGENTS.md")
	assert.Contains(t, result.ExistingContext, "CLAUDE.md")
	assert.Equal(t, "# My Agent Context", result.ExistingContext["AGENTS.md"])
}

func TestScan_ConfigSignals(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github/workflows"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.ConfigSignals, "GitHub Actions CI/CD")
	assert.Contains(t, result.ConfigSignals, "Docker containerization")
	assert.Contains(t, result.ConfigSignals, "Makefile build system")
}

func TestScan_NonexistentPath(t *testing.T) {
	s := NewProjectScanner()
	_, err := s.Scan("/nonexistent/path")
	assert.Error(t, err)
}

func TestScan_FileNotDirectory(t *testing.T) {
	f, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	s := NewProjectScanner()
	_, err = s.Scan(f.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestScanResult_FormatAsDescription(t *testing.T) {
	result := &ScanResult{
		Language:      "Go",
		Framework:     "Gin (HTTP)",
		Dependencies:  []string{"github.com/gin-gonic/gin", "github.com/spf13/cobra"},
		ConfigSignals: []string{"Docker containerization", "GitHub Actions CI/CD"},
		DirectoryTree: "├── cmd/\n├── internal/\n└── go.mod\n",
		README:        "# My API\n\nA REST API for inventory management.",
		ExistingContext: map[string]string{
			"AGENTS.md": "# Agent context",
		},
	}

	desc := result.FormatAsDescription()

	assert.Contains(t, desc, "**Language:** Go")
	assert.Contains(t, desc, "**Framework:** Gin (HTTP)")
	assert.Contains(t, desc, "github.com/gin-gonic/gin")
	assert.Contains(t, desc, "Docker containerization")
	assert.Contains(t, desc, "cmd/")
	assert.Contains(t, desc, "# My API")
	assert.Contains(t, desc, "--- AGENTS.md ---")
}

func TestScan_ExpandedContextFiles(t *testing.T) {
	dir := t.TempDir()

	// Create expanded context files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CONTRIBUTING.md"), []byte("# Contributing\n\nPlease follow conventional commits."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ARCHITECTURE.md"), []byte("# Architecture\n\nClean Architecture with DDD."), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude/CLAUDE.md"), []byte("# Claude context"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.0.0\ninfo:\n  title: My API"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.ExistingContext, "CONTRIBUTING.md")
	assert.Contains(t, result.ExistingContext, "ARCHITECTURE.md")
	assert.Contains(t, result.ExistingContext, ".claude/CLAUDE.md")
	assert.Contains(t, result.ExistingContext, "openapi.yaml")
}

func TestScan_ContextGlobs(t *testing.T) {
	dir := t.TempDir()

	// Create files matching glob patterns
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor/rules"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".cursor/rules/go-patterns.md"), []byte("# Go patterns"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs/adr"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs/adr/001-use-ddd.md"), []byte("# ADR 001: Use DDD"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs/adr/002-use-bdd.md"), []byte("# ADR 002: Use BDD"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.ExistingContext, ".cursor/rules/go-patterns.md")
	assert.Contains(t, result.ExistingContext, "docs/adr/001-use-ddd.md")
	assert.Contains(t, result.ExistingContext, "docs/adr/002-use-bdd.md")
}

func TestScan_ChangelogTruncation(t *testing.T) {
	dir := t.TempDir()

	// Create a CHANGELOG with 200 lines
	var lines string
	for i := 1; i <= 200; i++ {
		lines += fmt.Sprintf("## v%d.0.0 - 2024-01-%02d\n", i, i%28+1)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(lines), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	changelog := result.ExistingContext["CHANGELOG.md"]
	assert.NotEmpty(t, changelog)
	assert.Contains(t, changelog, "[... truncated ...]")
	// Should be around maxChangelogLines (50) lines
	changelogLines := strings.Split(changelog, "\n")
	assert.LessOrEqual(t, len(changelogLines), maxChangelogLines+3) // +3 for truncation marker and trailing
}

func TestScan_LargeContextFileTruncation(t *testing.T) {
	dir := t.TempDir()

	// Create a CONTRIBUTING.md with 300 lines
	var lines string
	for i := 1; i <= 300; i++ {
		lines += fmt.Sprintf("Line %d of contributing guide\n", i)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CONTRIBUTING.md"), []byte(lines), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	content := result.ExistingContext["CONTRIBUTING.md"]
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "[... truncated ...]")
	contentLines := strings.Split(content, "\n")
	assert.LessOrEqual(t, len(contentLines), maxContextFileLines+3)
}

func TestParseMakefileTargets(t *testing.T) {
	dir := t.TempDir()

	makefile := `.PHONY: build test lint

# Build the application
build:
	go build -o bin/app ./cmd/app

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte(makefile), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.BuildTargets, "Makefile")
	targets := result.BuildTargets["Makefile"]
	assert.Contains(t, targets, "build")
	assert.Contains(t, targets, "test")
	assert.Contains(t, targets, "test-integration")
	assert.Contains(t, targets, "lint")
	assert.Contains(t, targets, "clean")
	// .PHONY should not appear as a target
	for _, t2 := range targets {
		assert.NotEqual(t, ".PHONY", t2)
	}
}

func TestParseTaskfileTargets(t *testing.T) {
	dir := t.TempDir()

	taskfile := `version: '3'

tasks:
  build:
    cmds:
      - go build -o bin/app ./cmd/app

  test:unit:
    cmds:
      - go test ./...

  test:bdd:
    cmds:
      - go test ./tests/bdd/...

  lint:
    cmds:
      - golangci-lint run
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(taskfile), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.BuildTargets, "Taskfile")
	targets := result.BuildTargets["Taskfile"]
	assert.Contains(t, targets, "build")
	assert.Contains(t, targets, "test:unit")
	assert.Contains(t, targets, "test:bdd")
	assert.Contains(t, targets, "lint")
}

func TestFormatAsDescription_WithBuildTargets(t *testing.T) {
	result := &ScanResult{
		Language: "Go",
		BuildTargets: map[string][]string{
			"Makefile": {"build", "lint", "test"},
			"Taskfile": {"build", "test:bdd"},
		},
	}

	desc := result.FormatAsDescription()
	assert.Contains(t, desc, "**Build Targets:**")
	assert.Contains(t, desc, "Makefile:")
	assert.Contains(t, desc, "Taskfile:")
	assert.Contains(t, desc, "build")
}

func TestDetectTestingPatterns_GoWithBDD(t *testing.T) {
	dir := t.TempDir()

	// Create Go test files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "features"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "features/login.feature"), []byte("Feature: Login"), 0644))

	// Create go.mod with godog dependency
	goMod := "module example\n\nrequire (\n\tgithub.com/cucumber/godog v0.14.0\n)\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.TestingSignals, "go test (unit tests)")
	assert.Contains(t, result.TestingSignals, "BDD/Gherkin scenarios")
	assert.Contains(t, result.TestingSignals, "godog/BDD")
}

func TestDetectTestingPatterns_JSWithJest(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.test.ts"), []byte("test('works')"), 0644))

	// Create package.json with jest
	pkgJSON := `{"devDependencies": {"jest": "^29.0.0", "typescript": "^5.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "jest.config.ts"), []byte("module.exports = {}"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.TestingSignals, "test-style tests")
	assert.Contains(t, result.TestingSignals, "Jest")
	assert.Contains(t, result.TestingSignals, "Jest config")
}

func TestDetectTestingPatterns_CoverageConfig(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "codecov.yml"), []byte("coverage:\n  status:\n    project: yes"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]\ntestpaths = tests"), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Contains(t, result.TestingSignals, "Codecov")
	assert.Contains(t, result.TestingSignals, "pytest config")
}

func TestSummarizeCIWorkflows_GitHub(t *testing.T) {
	dir := t.TempDir()

	// Create GitHub Actions workflow
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github/workflows"), 0755))
	workflow := `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: golangci-lint run

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - run: go build ./...
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github/workflows/ci.yml"), []byte(workflow), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	require.Len(t, result.CIWorkflows, 1)
	assert.Equal(t, "ci.yml", result.CIWorkflows[0].File)
	assert.Contains(t, result.CIWorkflows[0].Triggers, "push")
	assert.Contains(t, result.CIWorkflows[0].Triggers, "pull_request")
	assert.Contains(t, result.CIWorkflows[0].Jobs, "lint")
	assert.Contains(t, result.CIWorkflows[0].Jobs, "test")
	assert.Contains(t, result.CIWorkflows[0].Jobs, "build")
}

func TestSummarizeCIWorkflows_InlineTriggers(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".github/workflows"), 0755))
	workflow := `name: Quick CI
on: [push, pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "ok"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".github/workflows/quick.yml"), []byte(workflow), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	require.Len(t, result.CIWorkflows, 1)
	assert.Contains(t, result.CIWorkflows[0].Triggers, "push")
	assert.Contains(t, result.CIWorkflows[0].Triggers, "pull_request")
	assert.Contains(t, result.CIWorkflows[0].Jobs, "check")
}

func TestSummarizeCIWorkflows_NoWorkflows(t *testing.T) {
	dir := t.TempDir()

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Empty(t, result.CIWorkflows)
}

func TestScan_RustProject(t *testing.T) {
	dir := t.TempDir()

	cargoToml := `[package]
name = "my-api"
version = "0.1.0"

[dependencies]
actix-web = "4"
serde = { version = "1", features = ["derive"] }
tokio = { version = "1", features = ["full"] }

[dev-dependencies]
assert_matches = "1"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoToml), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "Rust", result.Language)
	assert.Equal(t, "Actix Web", result.Framework)
	assert.Contains(t, result.Dependencies, "actix-web")
	assert.Contains(t, result.Dependencies, "serde")
	assert.Contains(t, result.Dependencies, "tokio")
	// dev-dependencies should not be included (different section)
	assert.NotContains(t, result.Dependencies, "assert_matches")
}

func TestScan_JavaProject(t *testing.T) {
	dir := t.TempDir()

	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>my-api</artifactId>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot</artifactId>
      <version>3.2.0</version>
    </dependency>
    <dependency>
      <groupId>com.fasterxml.jackson.core</groupId>
      <artifactId>jackson-databind</artifactId>
    </dependency>
  </dependencies>
</project>
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(pomXML), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "Java", result.Language)
	assert.Equal(t, "Spring Boot", result.Framework)
	assert.Contains(t, result.Dependencies, "spring-boot")
	assert.Contains(t, result.Dependencies, "jackson-databind")
}

func TestScan_RubyProject(t *testing.T) {
	dir := t.TempDir()

	gemfile := `source "https://rubygems.org"

gem "rails", "~> 7.1"
gem "pg", "~> 1.1"
gem "puma", ">= 5.0"
gem 'redis', "~> 5.0"

group :development, :test do
  gem "rspec"
end
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Gemfile"), []byte(gemfile), 0644))

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Equal(t, "Ruby", result.Language)
	assert.Equal(t, "Ruby on Rails", result.Framework)
	assert.Contains(t, result.Dependencies, "rails")
	assert.Contains(t, result.Dependencies, "pg")
	assert.Contains(t, result.Dependencies, "puma")
	assert.Contains(t, result.Dependencies, "redis")
	assert.Contains(t, result.Dependencies, "rspec")
}

func TestFilterREADMEContent_Badges(t *testing.T) {
	lines := []string{
		"# My Project",
		"",
		"[![Build Status](https://github.com/example/workflows/badge.svg)](https://github.com/example)",
		"[![Coverage](https://codecov.io/badge.svg)](https://codecov.io)",
		"![License](https://img.shields.io/badge/license-MIT-blue.svg)",
		"",
		"## Description",
		"A great project.",
	}

	filtered := filterREADMEContent(lines)

	assert.Contains(t, strings.Join(filtered, "\n"), "# My Project")
	assert.Contains(t, strings.Join(filtered, "\n"), "## Description")
	assert.Contains(t, strings.Join(filtered, "\n"), "A great project.")
	assert.NotContains(t, strings.Join(filtered, "\n"), "Build Status")
	assert.NotContains(t, strings.Join(filtered, "\n"), "Coverage")
	assert.NotContains(t, strings.Join(filtered, "\n"), "License")
}

func TestFilterREADMEContent_HTMLComments(t *testing.T) {
	lines := []string{
		"# My Project",
		"<!-- This is a comment",
		"that spans multiple lines -->",
		"## Real Content",
		"Important info.",
		"<!-- single line comment -->",
		"More content.",
	}

	filtered := filterREADMEContent(lines)

	result := strings.Join(filtered, "\n")
	assert.Contains(t, result, "# My Project")
	assert.Contains(t, result, "## Real Content")
	assert.Contains(t, result, "Important info.")
	assert.Contains(t, result, "More content.")
	assert.NotContains(t, result, "This is a comment")
	assert.NotContains(t, result, "single line comment")
}

func TestFilterREADMEContent_TableOfContents(t *testing.T) {
	lines := []string{
		"# My Project",
		"",
		"## Table of Contents",
		"- [Installation](#installation)",
		"- [Usage](#usage)",
		"- [Contributing](#contributing)",
		"",
		"## Installation",
		"Run `go install`.",
	}

	filtered := filterREADMEContent(lines)

	result := strings.Join(filtered, "\n")
	assert.Contains(t, result, "# My Project")
	assert.Contains(t, result, "## Installation")
	assert.Contains(t, result, "Run `go install`.")
	assert.NotContains(t, result, "Table of Contents")
	assert.NotContains(t, result, "[Installation](#installation)")
}

func TestFilterREADMEContent_PreservesContent(t *testing.T) {
	lines := []string{
		"# My API",
		"",
		"A REST API for inventory management.",
		"",
		"## Features",
		"- CRUD operations",
		"- Authentication",
		"",
		"## Quick Start",
		"```bash",
		"go run ./cmd/api",
		"```",
	}

	filtered := filterREADMEContent(lines)

	// All meaningful content should be preserved
	assert.Equal(t, len(lines), len(filtered))
	assert.Equal(t, lines, filtered)
}

func TestFilterREADMEContent_CollapsesBlanks(t *testing.T) {
	lines := []string{
		"# Title",
		"",
		"",
		"",
		"",
		"## Section",
	}

	filtered := filterREADMEContent(lines)

	// Should collapse 4 blanks to 2 (max)
	result := strings.Join(filtered, "\n")
	assert.Contains(t, result, "# Title")
	assert.Contains(t, result, "## Section")
	assert.Less(t, len(filtered), len(lines))
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	s := NewProjectScanner()
	result, err := s.Scan(dir)

	require.NoError(t, err)
	assert.Empty(t, result.Language)
	assert.Empty(t, result.Framework)
	assert.Empty(t, result.Dependencies)
	assert.Empty(t, result.README)
	assert.Empty(t, result.ExistingContext)
	assert.Empty(t, result.ConfigSignals)
	assert.Empty(t, result.BuildTargets)
	assert.Empty(t, result.TestingSignals)
}
