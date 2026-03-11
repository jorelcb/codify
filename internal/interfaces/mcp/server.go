package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/jorelcb/ai-context-generator/internal/application/command"
	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/filesystem"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/llm"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/scanner"
	infratemplate "github.com/jorelcb/ai-context-generator/internal/infrastructure/template"
)

const serverVersion = "2.2.0"

// validPresets maps preset names for validation.
var validPresets = map[string]bool{
	"default": true,
	"neutral": true,
}

// NewServer creates and configures the MCP server with all tools registered.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"ai-context-generator",
		serverVersion,
		server.WithToolCapabilities(true),
	)

	s.AddTools(
		generateContextTool(),
		generateSpecsTool(),
		analyzeProjectTool(),
	)

	return s
}

// generateContextTool defines the generate_context MCP tool.
func generateContextTool() server.ServerTool {
	tool := mcp.NewTool("generate_context",
		mcp.WithDescription("Generate AI-optimized context files for a software project from a description"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Project description")),
		mcp.WithString("language", mcp.Description("Programming language (go, python, javascript, etc.)")),
		mcp.WithString("preset", mcp.Description("Template preset: default (DDD/Clean Architecture) or neutral"), mcp.DefaultString("default")),
		mcp.WithString("locale", mcp.Description("Output language: en (English) or es (Spanish)"), mcp.DefaultString("en")),
		mcp.WithString("model", mcp.Description("Claude model to use"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithBoolean("with_specs", mcp.Description("Also generate SDD spec files after context generation")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateContext}
}

// generateSpecsTool defines the generate_specs MCP tool.
func generateSpecsTool() server.ServerTool {
	tool := mcp.NewTool("generate_specs",
		mcp.WithDescription("Generate SDD specification files from existing context files"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
		mcp.WithString("from_context", mcp.Required(), mcp.Description("Path to existing output directory with context files")),
		mcp.WithString("locale", mcp.Description("Output language: en or es"), mcp.DefaultString("en")),
		mcp.WithString("model", mcp.Description("Claude model to use"), mcp.DefaultString("claude-sonnet-4-6")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateSpecs}
}

// analyzeProjectTool defines the analyze_project MCP tool.
func analyzeProjectTool() server.ServerTool {
	tool := mcp.NewTool("analyze_project",
		mcp.WithDescription("Scan an existing project directory and generate AI context files from its structure, dependencies, and README"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to the project directory to analyze")),
		mcp.WithString("name", mcp.Description("Project name (defaults to directory name)")),
		mcp.WithString("language", mcp.Description("Override detected language")),
		mcp.WithString("preset", mcp.Description("Template preset: default or neutral"), mcp.DefaultString("default")),
		mcp.WithString("locale", mcp.Description("Output language: en or es"), mcp.DefaultString("en")),
		mcp.WithString("model", mcp.Description("Claude model to use"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithBoolean("with_specs", mcp.Description("Also generate SDD spec files after context generation")),
	)

	return server.ServerTool{Tool: tool, Handler: handleAnalyzeProject}
}

// --- Tool Handlers ---

func handleGenerateContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := stringArg(request, "name")
	description := stringArg(request, "description")
	language := stringArg(request, "language")
	preset := stringArgDefault(request, "preset", "default")
	locale := stringArgDefault(request, "locale", "en")
	model := stringArgDefault(request, "model", "")
	withSpecs := boolArg(request, "with_specs")

	result, err := executeGenerate(ctx, name, description, language, preset, locale, model)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Generation failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Context files generated for '%s'\n", name))
	sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
	sb.WriteString(fmt.Sprintf("Model: %s\n", result.Model))
	sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut))
	sb.WriteString("\nGenerated files:\n")
	for _, f := range result.GeneratedFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	if withSpecs {
		specResult, err := executeSpecs(ctx, name, result.OutputPath, locale, model)
		if err != nil {
			sb.WriteString(fmt.Sprintf("\nSpec generation failed: %v\n", err))
		} else {
			sb.WriteString(fmt.Sprintf("\nSpec files generated\n"))
			sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", specResult.TokensIn, specResult.TokensOut))
			sb.WriteString("\nSpec files:\n")
			for _, f := range specResult.GeneratedFiles {
				sb.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGenerateSpecs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := stringArg(request, "name")
	fromContext := stringArg(request, "from_context")
	locale := stringArgDefault(request, "locale", "en")
	model := stringArgDefault(request, "model", "")

	result, err := executeSpecs(ctx, name, fromContext, locale, model)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Spec generation failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Spec files generated for '%s'\n", name))
	sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
	sb.WriteString(fmt.Sprintf("Model: %s\n", result.Model))
	sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut))
	sb.WriteString("\nGenerated files:\n")
	for _, f := range result.GeneratedFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleAnalyzeProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath := stringArg(request, "project_path")
	name := stringArg(request, "name")
	language := stringArg(request, "language")
	preset := stringArgDefault(request, "preset", "default")
	locale := stringArgDefault(request, "locale", "en")
	model := stringArgDefault(request, "model", "")
	withSpecs := boolArg(request, "with_specs")

	// Resolve path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	if name == "" {
		name = filepath.Base(absPath)
	}

	// Scan project
	s := scanner.NewProjectScanner()
	scanResult, err := s.Scan(absPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Scan failed: %v", err)), nil
	}

	// Use detected language if not overridden
	if language == "" && scanResult.Language != "" {
		language = normalizeLanguageFlag(scanResult.Language)
	}

	// Format scan as description and generate
	description := scanResult.FormatAsDescription()
	result, err := executeGenerate(ctx, name, description, language, preset, locale, model)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Generation failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project analyzed and context generated for '%s'\n", name))
	sb.WriteString(fmt.Sprintf("Detected: %s", scanResult.Language))
	if scanResult.Framework != "" {
		sb.WriteString(fmt.Sprintf(" / %s", scanResult.Framework))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
	sb.WriteString(fmt.Sprintf("Model: %s\n", result.Model))
	sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut))
	sb.WriteString("\nGenerated files:\n")
	for _, f := range result.GeneratedFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	if withSpecs {
		specResult, err := executeSpecs(ctx, name, result.OutputPath, locale, model)
		if err != nil {
			sb.WriteString(fmt.Sprintf("\nSpec generation failed: %v\n", err))
		} else {
			sb.WriteString(fmt.Sprintf("\nSpec files generated\n"))
			for _, f := range specResult.GeneratedFiles {
				sb.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// --- Execution helpers (shared by all handlers) ---

func executeGenerate(ctx context.Context, name, description, language, preset, locale, model string) (*dto.GenerationResult, error) {
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return nil, err
	}

	if !validPresets[preset] {
		preset = "default"
	}

	templatePath := filepath.Join("templates", locale, preset)
	localeBase := filepath.Join("templates", locale)

	var templateLoader service.TemplateLoader
	if language != "" {
		templateLoader = infratemplate.NewFileSystemTemplateLoaderWithLanguage(templatePath, localeBase, language)
	} else {
		templateLoader = infratemplate.NewFileSystemTemplateLoader(templatePath)
	}

	guides, err := templateLoader.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	provider, err := llm.NewProvider(ctx, model, apiKey, nil) // no stdout in MCP mode
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	generateCmd := command.NewGenerateContextCommand(provider, fileWriter, dirManager)

	outputPath := filepath.Join("output", name)
	config := &dto.ProjectConfig{
		Name:        name,
		Description: description,
		Language:    language,
		Model:       model,
		OutputPath:  outputPath,
		Locale:      locale,
	}

	return generateCmd.Execute(ctx, config, guides)
}

// specTemplateMapping maps spec template file names to their guide names.
var specTemplateMapping = map[string]string{
	"constitution.template": "constitution",
	"spec.template":         "spec",
	"plan.template":         "plan",
	"tasks.template":        "tasks",
}

func executeSpecs(ctx context.Context, name, fromContextPath, locale, model string) (*dto.GenerationResult, error) {
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return nil, err
	}

	contextReader := filesystem.NewContextReader()
	existingContext, err := contextReader.ReadExistingContext(fromContextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing context: %w", err)
	}

	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(
		filepath.Join("templates", locale, "spec"), specTemplateMapping,
	)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load spec templates: %w", err)
	}

	provider, err := llm.NewProvider(ctx, model, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()

	specCmd := command.NewGenerateSpecCommand(provider, fileWriter, dirManager)

	outputPath := filepath.Join("output", name)
	config := &dto.SpecConfig{
		ProjectName:     name,
		FromContextPath: fromContextPath,
		OutputPath:      outputPath,
		Model:           model,
		Locale:          locale,
	}

	result, err := specCmd.Execute(ctx, config, existingContext, guides)
	if err != nil {
		return nil, err
	}

	// Update AGENTS.md with specs reference
	agentsPath := filepath.Join(fromContextPath, "AGENTS.md")
	content, readErr := os.ReadFile(agentsPath)
	if readErr == nil && !strings.Contains(string(content), "specs/") {
		var specsRef string
		if locale == "es" {
			specsRef = "\n## Especificaciones\n\n" +
				"- Constitucion del proyecto: `specs/CONSTITUTION.md`\n" +
				"- Especificaciones de features: `specs/SPEC.md`\n" +
				"- Diseno tecnico y plan: `specs/PLAN.md`\n" +
				"- Desglose de tareas: `specs/TASKS.md`\n"
		} else {
			specsRef = "\n## Specifications\n\n" +
				"- Project constitution: `specs/CONSTITUTION.md`\n" +
				"- Feature specifications: `specs/SPEC.md`\n" +
				"- Technical design and plan: `specs/PLAN.md`\n" +
				"- Task breakdown: `specs/TASKS.md`\n"
		}
		_ = os.WriteFile(agentsPath, []byte(string(content)+specsRef), 0644)
	}

	return result, nil
}

// --- Argument helpers ---

func stringArg(request mcp.CallToolRequest, name string) string {
	if v, ok := request.GetArguments()[name].(string); ok {
		return v
	}
	return ""
}

func stringArgDefault(request mcp.CallToolRequest, name, defaultVal string) string {
	if v := stringArg(request, name); v != "" {
		return v
	}
	return defaultVal
}

func boolArg(request mcp.CallToolRequest, name string) bool {
	if v, ok := request.GetArguments()[name].(bool); ok {
		return v
	}
	return false
}

// normalizeLanguageFlag maps detected language names to CLI flag values.
func normalizeLanguageFlag(detected string) string {
	mapping := map[string]string{
		"Go":                    "go",
		"JavaScript/TypeScript": "javascript",
		"Python":                "python",
		"Rust":                  "rust",
		"Java":                  "java",
		"Ruby":                  "ruby",
		"Elixir":                "elixir",
		"PHP":                   "php",
		"Swift":                 "swift",
		"C#/.NET":               "csharp",
	}
	if flag, ok := mapping[detected]; ok {
		return flag
	}
	return ""
}
