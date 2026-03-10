package scanner

import (
	"os"
	"path/filepath"
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
}
