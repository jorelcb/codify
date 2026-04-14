package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ScanResult holds the extracted signals from a project directory.
type ScanResult struct {
	Language        string              // Detected primary language
	Framework       string              // Detected framework (if any)
	Dependencies    []string            // Key dependencies from manifest
	DirectoryTree   string              // Directory structure (limited depth)
	README          string              // README content (truncated)
	ExistingContext map[string]string    // Existing context files (AGENTS.md, CLAUDE.md, etc.)
	ConfigSignals   []string            // Detected config signals (CI, Docker, etc.)
	BuildTargets    map[string][]string // Build targets by source ("Makefile", "Taskfile")
	TestingSignals  []string            // Detected testing patterns and frameworks
	CIWorkflows     []CIWorkflowSummary // Summarized CI/CD pipeline definitions
}

// CIWorkflowSummary holds a lightweight summary of a CI/CD workflow file.
type CIWorkflowSummary struct {
	File     string   // Filename (e.g., "ci.yml")
	Triggers []string // Trigger events (e.g., "push", "pull_request")
	Jobs     []string // Job names (e.g., "lint", "test", "build")
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
	result.BuildTargets = s.parseBuildTargets(projectPath)
	result.TestingSignals = s.detectTestingPatterns(projectPath, result.Dependencies)
	result.CIWorkflows = s.summarizeCIWorkflows(projectPath)

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

	if len(r.BuildTargets) > 0 {
		sb.WriteString("\n**Build Targets:**\n")
		for source, targets := range r.BuildTargets {
			sb.WriteString(fmt.Sprintf("%s: %s\n", source, strings.Join(targets, ", ")))
		}
	}

	if len(r.TestingSignals) > 0 {
		sb.WriteString("\n**Testing Patterns:**\n")
		for _, sig := range r.TestingSignals {
			sb.WriteString(fmt.Sprintf("- %s\n", sig))
		}
	}

	if len(r.CIWorkflows) > 0 {
		sb.WriteString("\n**CI/CD Pipelines:**\n")
		for _, wf := range r.CIWorkflows {
			sb.WriteString(fmt.Sprintf("- %s: triggers [%s], jobs: %s\n",
				wf.File,
				strings.Join(wf.Triggers, ", "),
				strings.Join(wf.Jobs, ", ")))
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
	"rocket":    "Rocket (HTTP)",
	// Java/Kotlin
	"spring-boot":    "Spring Boot",
	"spring-webflux": "Spring WebFlux",
	"quarkus":        "Quarkus",
	"micronaut":      "Micronaut",
	// Ruby
	"rails":   "Ruby on Rails",
	"sinatra": "Sinatra",
	"hanami":  "Hanami",
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
	case "Rust":
		deps, detectedFramework = parseRustDeps(content)
	case "Java", "Java/Kotlin":
		deps, detectedFramework = parseJavaDeps(content)
	case "Ruby":
		deps, detectedFramework = parseRubyDeps(content)
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

func parseRustDeps(content string) ([]string, string) {
	var deps []string
	var framework string
	inDeps := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Detect [dependencies] section
		if trimmed == "[dependencies]" {
			inDeps = true
			continue
		}
		// Another section header ends dependencies
		if strings.HasPrefix(trimmed, "[") {
			inDeps = false
			continue
		}

		if inDeps && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			// Parse "name = ..." or "name = { version = ... }"
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) >= 1 {
				dep := strings.TrimSpace(parts[0])
				if dep != "" {
					deps = append(deps, dep)
					if fw, ok := frameworkDetectors[dep]; ok && framework == "" {
						framework = fw
					}
				}
			}
		}
	}

	return deps, framework
}

// javaArtifactRegex matches <artifactId>name</artifactId> in pom.xml.
var javaArtifactRegex = regexp.MustCompile(`<artifactId>([^<]+)</artifactId>`)

func parseJavaDeps(content string) ([]string, string) {
	var deps []string
	var framework string

	matches := javaArtifactRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			dep := strings.TrimSpace(match[1])
			if dep != "" && !seen[dep] {
				seen[dep] = true
				deps = append(deps, dep)
				if fw, ok := frameworkDetectors[dep]; ok && framework == "" {
					framework = fw
				}
			}
		}
	}

	return deps, framework
}

// rubyGemRegex matches gem 'name' or gem "name" lines in Gemfile.
var rubyGemRegex = regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]`)

func parseRubyDeps(content string) ([]string, string) {
	var deps []string
	var framework string

	for _, line := range strings.Split(content, "\n") {
		matches := rubyGemRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			dep := matches[1]
			deps = append(deps, dep)
			if fw, ok := frameworkDetectors[dep]; ok && framework == "" {
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
		lines = filterREADMEContent(lines)
		if len(lines) > s.maxReadmeLines {
			lines = lines[:s.maxReadmeLines]
			lines = append(lines, "\n[... truncated ...]")
		}
		return strings.Join(lines, "\n")
	}
	return ""
}

// badgeLineRegex matches markdown badge lines like [![badge](url)](link) or ![img](url).
var badgeLineRegex = regexp.MustCompile(`^\s*(\[!\[|!\[).*\]\(https?://`)

// tocHeadingRegex matches Table of Contents headings.
var tocHeadingRegex = regexp.MustCompile(`(?i)^##\s+(table\s+of\s+contents|toc|contents)\s*$`)

// filterREADMEContent removes noise from README lines: badges, HTML comments, ToC sections,
// and collapses excessive blank lines.
func filterREADMEContent(lines []string) []string {
	var filtered []string
	inHTMLComment := false
	inToC := false
	blankCount := 0

	for _, line := range lines {
		// Handle HTML comment blocks
		if strings.Contains(line, "<!--") {
			inHTMLComment = true
		}
		if inHTMLComment {
			if strings.Contains(line, "-->") {
				inHTMLComment = false
			}
			continue
		}

		// Skip badge lines
		if badgeLineRegex.MatchString(line) {
			continue
		}

		// Handle Table of Contents section
		if tocHeadingRegex.MatchString(line) {
			inToC = true
			continue
		}
		// A new ## heading ends the ToC section
		if inToC {
			if strings.HasPrefix(line, "## ") {
				inToC = false
				// Fall through to include this heading
			} else {
				continue
			}
		}

		// Collapse excessive blank lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blankCount++
			if blankCount >= 3 {
				continue
			}
		} else {
			blankCount = 0
		}

		filtered = append(filtered, line)
	}

	return filtered
}

// contextFileNames are files that indicate existing AI context.
var contextFileNames = []string{
	// AI agent context
	"AGENTS.md", "CLAUDE.md", ".cursorrules",
	".claude/CLAUDE.md", ".claude/settings.json",
	// Project context directory
	"context/CONTEXT.md", "context/DEVELOPMENT_GUIDE.md",
	"context/INTERACTIONS_LOG.md", "context/IDIOMS.md",
	// Project documentation
	"CONTRIBUTING.md", "ARCHITECTURE.md", ".editorconfig",
	".github/CODEOWNERS",
	// API contracts
	"openapi.yaml", "openapi.json", "swagger.yaml", "swagger.json",
	"schema.graphql",
	// Changelog (truncated to recent entries)
	"CHANGELOG.md",
}

// contextFileGlobs are glob patterns for context files that may vary in name.
var contextFileGlobs = []string{
	".cursor/rules/*.md",
	"docs/adr/*.md",
	"proto/*.proto",
}

// maxContextFileLines is the default truncation limit for large context files.
const maxContextFileLines = 200

// maxChangelogLines is the truncation limit for CHANGELOG files.
const maxChangelogLines = 50

func (s *ProjectScanner) readExistingContext(projectPath string) map[string]string {
	found := make(map[string]string)

	// Exact file matches
	for _, name := range contextFileNames {
		content, err := os.ReadFile(filepath.Join(projectPath, name))
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(content))
		if text == "" {
			continue
		}
		found[name] = truncateContextFile(name, text)
	}

	// Glob pattern matches
	for _, pattern := range contextFileGlobs {
		matches, err := filepath.Glob(filepath.Join(projectPath, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			text := strings.TrimSpace(string(content))
			if text == "" {
				continue
			}
			relPath, _ := filepath.Rel(projectPath, match)
			found[relPath] = truncateContextFile(relPath, text)
		}
	}

	return found
}

// truncateContextFile applies file-specific truncation limits.
func truncateContextFile(name, content string) string {
	lines := strings.Split(content, "\n")

	limit := maxContextFileLines
	if strings.EqualFold(filepath.Base(name), "changelog.md") {
		limit = maxChangelogLines
	}

	if len(lines) > limit {
		lines = lines[:limit]
		return strings.Join(lines, "\n") + "\n\n[... truncated ...]"
	}
	return content
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

// makefileTargetRegex matches Makefile target lines (e.g., "build:", "test-unit:").
var makefileTargetRegex = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

func (s *ProjectScanner) parseBuildTargets(projectPath string) map[string][]string {
	targets := make(map[string][]string)

	if makeTargets := s.parseMakefileTargets(projectPath); len(makeTargets) > 0 {
		targets["Makefile"] = makeTargets
	}
	if taskTargets := s.parseTaskfileTargets(projectPath); len(taskTargets) > 0 {
		targets["Taskfile"] = taskTargets
	}

	return targets
}

func (s *ProjectScanner) parseMakefileTargets(projectPath string) []string {
	content, err := os.ReadFile(filepath.Join(projectPath, "Makefile"))
	if err != nil {
		return nil
	}

	var targets []string
	seen := make(map[string]bool)

	for _, line := range strings.Split(string(content), "\n") {
		// Skip .PHONY and comment lines
		if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "#") {
			continue
		}
		matches := makefileTargetRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			target := matches[1]
			if !seen[target] {
				seen[target] = true
				targets = append(targets, target)
			}
		}
	}

	sort.Strings(targets)
	return targets
}

// taskfileTaskRegex matches Taskfile task keys at exactly 2-space indentation under tasks:.
var taskfileTaskRegex = regexp.MustCompile(`^  ([a-zA-Z_][a-zA-Z0-9_:.-]*)\s*:`)

func (s *ProjectScanner) parseTaskfileTargets(projectPath string) []string {
	var content []byte
	var err error

	for _, name := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		content, err = os.ReadFile(filepath.Join(projectPath, name))
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil
	}

	var targets []string
	inTasks := false

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		// Detect top-level "tasks:" section
		if trimmed == "tasks:" {
			inTasks = true
			continue
		}

		// Another top-level key ends the tasks section
		if inTasks && len(line) > 0 && line[0] != ' ' && line[0] != '#' {
			break
		}

		if inTasks {
			matches := taskfileTaskRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				targets = append(targets, matches[1])
			}
		}
	}

	sort.Strings(targets)
	return targets
}

// testFilePatterns maps glob patterns (relative, depth 1-2) to their signal description.
var testFilePatterns = []struct {
	patterns []string
	signal   string
}{
	{[]string{"*_test.go", "**/*_test.go"}, "go test (unit tests)"},
	{[]string{"*.spec.ts", "*.spec.js", "**/*.spec.ts", "**/*.spec.js"}, "spec-style tests"},
	{[]string{"*.test.ts", "*.test.js", "**/*.test.ts", "**/*.test.js"}, "test-style tests"},
	{[]string{"features/*.feature", "tests/**/*.feature", "test/**/*.feature"}, "BDD/Gherkin scenarios"},
}

// testFrameworkDeps maps dependency substrings to testing framework signals.
var testFrameworkDeps = map[string]string{
	"godog":                     "godog/BDD",
	"cucumber":                  "Cucumber/BDD",
	"github.com/cucumber/godog": "godog/BDD",
	"jest":                      "Jest",
	"vitest":                    "Vitest",
	"mocha":                     "Mocha",
	"rspec":                     "RSpec",
}

// coverageConfigFiles maps filenames to coverage tool signals.
var coverageConfigFiles = map[string]string{
	"jest.config.js":   "Jest config",
	"jest.config.ts":   "Jest config",
	"jest.config.json": "Jest config",
	"codecov.yml":      "Codecov",
	".codecov.yml":     "Codecov",
	".coveragerc":      "coverage.py",
	"pytest.ini":       "pytest config",
	".nycrc":           "nyc coverage",
	".nycrc.json":      "nyc coverage",
}

func (s *ProjectScanner) detectTestingPatterns(projectPath string, deps []string) []string {
	seen := make(map[string]bool)
	var signals []string

	addSignal := func(sig string) {
		if !seen[sig] {
			seen[sig] = true
			signals = append(signals, sig)
		}
	}

	// Detect test files by glob patterns
	for _, tp := range testFilePatterns {
		for _, pattern := range tp.patterns {
			matches, _ := filepath.Glob(filepath.Join(projectPath, pattern))
			if len(matches) > 0 {
				addSignal(tp.signal)
				break
			}
		}
	}

	// Detect frameworks from parsed dependencies
	for _, dep := range deps {
		for depKey, sig := range testFrameworkDeps {
			if strings.Contains(dep, depKey) {
				addSignal(sig)
			}
		}
	}

	// Detect coverage configuration files
	for file, sig := range coverageConfigFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			addSignal(sig)
		}
	}

	sort.Strings(signals)
	return signals
}

// maxCIWorkflowFiles limits how many CI workflow files we summarize to avoid token bloat.
const maxCIWorkflowFiles = 5

// ciTriggerRegex matches GitHub Actions trigger keys (e.g., "  push:", "  pull_request:").
var ciTriggerRegex = regexp.MustCompile(`^  ([a-z_]+)\s*:`)

// ciJobRegex matches GitHub Actions job keys (e.g., "  lint:", "  build:").
var ciJobRegex = regexp.MustCompile(`^  ([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

func (s *ProjectScanner) summarizeCIWorkflows(projectPath string) []CIWorkflowSummary {
	var summaries []CIWorkflowSummary

	// GitHub Actions
	for _, ext := range []string{"*.yml", "*.yaml"} {
		matches, _ := filepath.Glob(filepath.Join(projectPath, ".github", "workflows", ext))
		for _, match := range matches {
			if len(summaries) >= maxCIWorkflowFiles {
				break
			}
			if summary := s.parseGitHubWorkflow(match); summary != nil {
				summaries = append(summaries, *summary)
			}
		}
	}

	// GitLab CI
	gitlabCI := filepath.Join(projectPath, ".gitlab-ci.yml")
	if _, err := os.Stat(gitlabCI); err == nil && len(summaries) < maxCIWorkflowFiles {
		if summary := s.parseGitLabCI(gitlabCI); summary != nil {
			summaries = append(summaries, *summary)
		}
	}

	return summaries
}

func (s *ProjectScanner) parseGitHubWorkflow(filePath string) *CIWorkflowSummary {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	summary := &CIWorkflowSummary{
		File: filepath.Base(filePath),
	}

	lines := strings.Split(string(content), "\n")
	section := "" // "on" or "jobs"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect top-level sections
		if strings.HasPrefix(trimmed, "on:") {
			section = "on"
			// Handle inline: "on: [push, pull_request]"
			rest := strings.TrimPrefix(trimmed, "on:")
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "[") {
				rest = strings.Trim(rest, "[] ")
				for _, t := range strings.Split(rest, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						summary.Triggers = append(summary.Triggers, t)
					}
				}
				section = ""
			}
			continue
		}
		if trimmed == "jobs:" {
			section = "jobs"
			continue
		}
		// Another top-level key ends the current section
		if len(line) > 0 && line[0] != ' ' && line[0] != '#' && trimmed != "" {
			section = ""
		}

		switch section {
		case "on":
			if matches := ciTriggerRegex.FindStringSubmatch(line); len(matches) >= 2 {
				summary.Triggers = append(summary.Triggers, matches[1])
			}
		case "jobs":
			if matches := ciJobRegex.FindStringSubmatch(line); len(matches) >= 2 {
				summary.Jobs = append(summary.Jobs, matches[1])
			}
		}
	}

	if len(summary.Triggers) == 0 && len(summary.Jobs) == 0 {
		return nil
	}
	return summary
}

func (s *ProjectScanner) parseGitLabCI(filePath string) *CIWorkflowSummary {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	summary := &CIWorkflowSummary{
		File: ".gitlab-ci.yml",
	}

	// GitLab CI: top-level keys that aren't reserved are job names
	reservedKeys := map[string]bool{
		"image": true, "services": true, "stages": true, "variables": true,
		"before_script": true, "after_script": true, "cache": true,
		"include": true, "default": true, "workflow": true,
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		// Top-level key (no leading whitespace, ends with :)
		if len(line) > 0 && line[0] != ' ' && line[0] != '#' && strings.Contains(trimmed, ":") {
			key := strings.TrimSuffix(strings.Split(trimmed, ":")[0], " ")
			if key != "" && !strings.HasPrefix(key, ".") && !reservedKeys[key] {
				summary.Jobs = append(summary.Jobs, key)
			}
		}
	}

	if len(summary.Jobs) == 0 {
		return nil
	}
	return summary
}
