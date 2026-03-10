package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanResult holds the extracted signals from a project directory.
type ScanResult struct {
	Language       string            // Detected primary language
	Framework      string            // Detected framework (if any)
	Dependencies   []string          // Key dependencies from manifest
	DirectoryTree  string            // Directory structure (limited depth)
	README         string            // README content (truncated)
	ExistingContext map[string]string // Existing context files (AGENTS.md, CLAUDE.md, etc.)
	ConfigSignals  []string          // Detected config signals (CI, Docker, etc.)
}

// ProjectScanner extracts signals from an existing project directory.
type ProjectScanner struct {
	maxTreeDepth   int
	maxReadmeLines int
}

// NewProjectScanner creates a new ProjectScanner with sensible defaults.
func NewProjectScanner() *ProjectScanner {
	return &ProjectScanner{
		maxTreeDepth:   3,
		maxReadmeLines: 100,
	}
}

// Scan analyzes the project at the given path and returns extracted signals.
func (s *ProjectScanner) Scan(projectPath string) (*ScanResult, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access project path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", projectPath)
	}

	result := &ScanResult{
		ExistingContext: make(map[string]string),
	}

	// All scan steps are independent — run them all, collecting what we can
	result.Language, result.Framework, result.Dependencies = s.detectLanguageAndDeps(projectPath)
	result.DirectoryTree = s.buildDirectoryTree(projectPath)
	result.README = s.readREADME(projectPath)
	result.ExistingContext = s.readExistingContext(projectPath)
	result.ConfigSignals = s.detectConfigSignals(projectPath)

	return result, nil
}

// FormatAsDescription formats the scan result as structured text suitable for LLM input.
func (r *ScanResult) FormatAsDescription() string {
	var sb strings.Builder

	sb.WriteString("## Project Analysis (auto-scanned from existing codebase)\n\n")

	if r.Language != "" {
		sb.WriteString(fmt.Sprintf("**Language:** %s\n", r.Language))
	}
	if r.Framework != "" {
		sb.WriteString(fmt.Sprintf("**Framework:** %s\n", r.Framework))
	}

	if len(r.Dependencies) > 0 {
		sb.WriteString("\n**Key Dependencies:**\n")
		for _, dep := range r.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
	}

	if len(r.ConfigSignals) > 0 {
		sb.WriteString("\n**Infrastructure Signals:**\n")
		for _, sig := range r.ConfigSignals {
			sb.WriteString(fmt.Sprintf("- %s\n", sig))
		}
	}

	if r.DirectoryTree != "" {
		sb.WriteString("\n**Directory Structure:**\n```\n")
		sb.WriteString(r.DirectoryTree)
		sb.WriteString("```\n")
	}

	if r.README != "" {
		sb.WriteString("\n**README Content:**\n")
		sb.WriteString(r.README)
		sb.WriteString("\n")
	}

	if len(r.ExistingContext) > 0 {
		sb.WriteString("\n**Existing Context Files Found:**\n")
		for name, content := range r.ExistingContext {
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", name))
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// manifestFiles maps manifest filenames to their language.
var manifestFiles = map[string]string{
	"go.mod":           "Go",
	"package.json":     "JavaScript/TypeScript",
	"pyproject.toml":   "Python",
	"requirements.txt": "Python",
	"Cargo.toml":       "Rust",
	"pom.xml":          "Java",
	"build.gradle":     "Java/Kotlin",
	"Gemfile":          "Ruby",
	"mix.exs":          "Elixir",
	"composer.json":    "PHP",
	"Package.swift":    "Swift",
	"*.csproj":         "C#/.NET",
}

// frameworkDetectors maps dependency patterns to framework names.
var frameworkDetectors = map[string]string{
	// Go
	"github.com/gin-gonic/gin":     "Gin (HTTP)",
	"github.com/labstack/echo":     "Echo (HTTP)",
	"github.com/gofiber/fiber":     "Fiber (HTTP)",
	"github.com/spf13/cobra":       "Cobra (CLI)",
	"github.com/gorilla/mux":       "Gorilla Mux (HTTP)",
	// JavaScript/TypeScript
	"next":    "Next.js",
	"react":   "React",
	"express": "Express",
	"nestjs":  "NestJS",
	"vue":     "Vue.js",
	"svelte":  "Svelte",
	"fastify": "Fastify",
	// Python
	"django":     "Django",
	"flask":      "Flask",
	"fastapi":    "FastAPI",
	"pytest":     "pytest (testing)",
	"sqlalchemy": "SQLAlchemy",
	// Rust
	"actix-web": "Actix Web",
	"axum":      "Axum",
	"tokio":     "Tokio (async)",
}

func (s *ProjectScanner) detectLanguageAndDeps(projectPath string) (language, framework string, deps []string) {
	for manifest, lang := range manifestFiles {
		if manifest == "*.csproj" {
			// Check for any .csproj file
			matches, _ := filepath.Glob(filepath.Join(projectPath, "*.csproj"))
			if len(matches) > 0 {
				language = lang
				break
			}
			continue
		}

		content, err := os.ReadFile(filepath.Join(projectPath, manifest))
		if err != nil {
			continue
		}

		language = lang
		deps, framework = s.parseDependencies(lang, string(content))
		break
	}

	return language, framework, deps
}

func (s *ProjectScanner) parseDependencies(lang, content string) (deps []string, framework string) {
	var detectedFramework string

	switch lang {
	case "Go":
		deps, detectedFramework = parseGoDeps(content)
	case "JavaScript/TypeScript":
		deps, detectedFramework = parseJSDeps(content)
	case "Python":
		deps, detectedFramework = parsePythonDeps(content)
	default:
		return nil, ""
	}

	return deps, detectedFramework
}

func parseGoDeps(content string) ([]string, string) {
	var deps []string
	var framework string
	lines := strings.Split(content, "\n")
	inRequire := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "require (" {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}
		if inRequire && line != "" && !strings.HasPrefix(line, "//") {
			// Skip indirect dependencies
			if strings.Contains(line, "// indirect") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				dep := parts[0]
				deps = append(deps, dep)
				if fw, ok := frameworkDetectors[dep]; ok && framework == "" {
					framework = fw
				}
			}
		}
	}

	return deps, framework
}

func parseJSDeps(content string) ([]string, string) {
	var deps []string
	var framework string

	// Simple extraction: look for "dependencies" and "devDependencies" keys
	for _, section := range []string{`"dependencies"`, `"devDependencies"`} {
		idx := strings.Index(content, section)
		if idx == -1 {
			continue
		}
		// Find the opening brace
		braceStart := strings.Index(content[idx:], "{")
		if braceStart == -1 {
			continue
		}
		braceStart += idx
		// Find matching closing brace
		depth := 0
		for i := braceStart; i < len(content); i++ {
			if content[i] == '{' {
				depth++
			} else if content[i] == '}' {
				depth--
				if depth == 0 {
					block := content[braceStart+1 : i]
					for _, line := range strings.Split(block, "\n") {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, `"`) {
							name := strings.Trim(strings.Split(line, ":")[0], `" `)
							if name != "" {
								deps = append(deps, name)
								if fw, ok := frameworkDetectors[name]; ok && framework == "" {
									framework = fw
								}
							}
						}
					}
					break
				}
			}
		}
	}

	return deps, framework
}

func parsePythonDeps(content string) ([]string, string) {
	var deps []string
	var framework string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		// Handle both requirements.txt (pkg==1.0) and pyproject.toml ("pkg>=1.0")
		dep := line
		for _, sep := range []string{">=", "<=", "==", "!=", "~=", ">", "<", ";"} {
			if idx := strings.Index(dep, sep); idx > 0 {
				dep = dep[:idx]
			}
		}
		dep = strings.Trim(dep, `"' ,`)
		if dep != "" && !strings.Contains(dep, "=") {
			deps = append(deps, dep)
			depLower := strings.ToLower(dep)
			if fw, ok := frameworkDetectors[depLower]; ok && framework == "" {
				framework = fw
			}
		}
	}

	return deps, framework
}

func (s *ProjectScanner) buildDirectoryTree(projectPath string) string {
	var sb strings.Builder
	s.walkTree(&sb, projectPath, "", 0)
	return sb.String()
}

// skipDirs are directories to skip during tree building.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "__pycache__": true,
	".venv": true, "venv": true, ".idea": true, ".vscode": true,
	"dist": true, "build": true, "target": true, ".next": true,
	".DS_Store": true, "coverage": true, ".mypy_cache": true,
	".pytest_cache": true, ".ruff_cache": true,
}

func (s *ProjectScanner) walkTree(sb *strings.Builder, dir, prefix string, depth int) {
	if depth >= s.maxTreeDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Filter and sort: directories first, then files
	var dirs, files []os.DirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != ".github" {
			continue
		}
		if skipDirs[name] {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}

	// Sort each group alphabetically
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	all := append(dirs, files...)
	for i, entry := range all {
		isLast := i == len(all)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("%s%s%s/\n", prefix, connector, entry.Name()))
			s.walkTree(sb, filepath.Join(dir, entry.Name()), childPrefix, depth+1)
		} else {
			sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, entry.Name()))
		}
	}
}

func (s *ProjectScanner) readREADME(projectPath string) string {
	for _, name := range []string{"README.md", "readme.md", "README.rst", "README.txt", "README"} {
		content, err := os.ReadFile(filepath.Join(projectPath, name))
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		if len(lines) > s.maxReadmeLines {
			lines = lines[:s.maxReadmeLines]
			lines = append(lines, "\n[... truncated ...]")
		}
		return strings.Join(lines, "\n")
	}
	return ""
}

// contextFileNames are files that indicate existing AI context.
var contextFileNames = []string{
	"AGENTS.md", "CLAUDE.md", ".cursorrules",
	"context/CONTEXT.md", "context/DEVELOPMENT_GUIDE.md",
	"context/INTERACTIONS_LOG.md", "context/IDIOMS.md",
}

func (s *ProjectScanner) readExistingContext(projectPath string) map[string]string {
	found := make(map[string]string)
	for _, name := range contextFileNames {
		content, err := os.ReadFile(filepath.Join(projectPath, name))
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(content))
		if text != "" {
			found[name] = text
		}
	}
	return found
}

// configSignalFiles maps filenames/dirs to their signal description.
var configSignalFiles = map[string]string{
	".github/workflows":  "GitHub Actions CI/CD",
	".gitlab-ci.yml":     "GitLab CI/CD",
	"Dockerfile":         "Docker containerization",
	"docker-compose.yml": "Docker Compose (multi-service)",
	"docker-compose.yaml": "Docker Compose (multi-service)",
	"Makefile":           "Makefile build system",
	"Taskfile.yml":       "Taskfile build system",
	"Taskfile.yaml":      "Taskfile build system",
	".env.example":       "Environment variable configuration",
	"terraform":          "Terraform infrastructure",
	"k8s":                "Kubernetes manifests",
	"helm":               "Helm charts",
}

func (s *ProjectScanner) detectConfigSignals(projectPath string) []string {
	var signals []string
	for path, signal := range configSignalFiles {
		fullPath := filepath.Join(projectPath, path)
		if _, err := os.Stat(fullPath); err == nil {
			signals = append(signals, signal)
		}
	}
	sort.Strings(signals)
	return signals
}
