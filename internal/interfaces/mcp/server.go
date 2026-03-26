package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
	"github.com/jorelcb/codify/internal/infrastructure/scanner"
	infratemplate "github.com/jorelcb/codify/internal/infrastructure/template"
)

const serverVersion = "1.13.0"

// validPresets maps preset names for validation.
var validPresets = map[string]bool{
	"default":  true,
	"neutral":  true,
	"workflow": true,
}

// NewServer creates and configures the MCP server with all tools registered.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"codify",
		serverVersion,
		server.WithToolCapabilities(true),
	)

	s.AddTools(
		generateContextTool(),
		generateSpecsTool(),
		analyzeProjectTool(),
		generateSkillsTool(),
		generateWorkflowsTool(),
		commitGuidanceTool(),
		versionGuidanceTool(),
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
		mcp.WithString("preset", mcp.Description("Template preset: default, neutral, or workflow"), mcp.DefaultString("default")),
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

// generateSkillsTool defines the generate_skills MCP tool.
func generateSkillsTool() server.ServerTool {
	tool := mcp.NewTool("generate_skills",
		mcp.WithDescription("Generate AI agent skills (SKILL.md) by category, preset, and mode. Static mode delivers instant skills from the catalog. Personalized mode uses LLM to adapt skills to a specific project context."),
		mcp.WithString("category", mcp.Required(), mcp.Description("Skill category: architecture, testing, or workflow")),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Preset within category. architecture: clean, neutral. testing: foundational, tdd, bdd. workflow: conventional-commit, semantic-versioning, all")),
		mcp.WithString("mode", mcp.Description("Generation mode: static (instant, no API key) or personalized (LLM-adapted)"), mcp.DefaultString("static")),
		mcp.WithString("project_context", mcp.Description("Project description for personalized mode (language, architecture, domain, stack)")),
		mcp.WithString("locale", mcp.Description("Output language: en or es"), mcp.DefaultString("en")),
		mcp.WithString("target", mcp.Description("Target ecosystem: claude, codex, or antigravity"), mcp.DefaultString("claude")),
		mcp.WithString("model", mcp.Description("LLM model (only for personalized mode)"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithString("output", mcp.Description("Output directory (default: ecosystem-specific, e.g. .claude/skills/)")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateSkills}
}

// generateWorkflowsTool defines the generate_workflows MCP tool.
func generateWorkflowsTool() server.ServerTool {
	tool := mcp.NewTool("generate_workflows",
		mcp.WithDescription("Generate Antigravity workflow files (.md) — multi-step recipes with execution annotations. Static mode delivers instant workflows. Personalized mode uses LLM to adapt workflows to a specific project context."),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Workflow preset: feature-development, bug-fix, release-cycle, or all")),
		mcp.WithString("mode", mcp.Description("Generation mode: static (instant, no API key) or personalized (LLM-adapted)"), mcp.DefaultString("static")),
		mcp.WithString("project_context", mcp.Description("Project description for personalized mode (language, tools, CI/CD, deployment)")),
		mcp.WithString("locale", mcp.Description("Output language: en or es"), mcp.DefaultString("en")),
		mcp.WithString("model", mcp.Description("LLM model (only for personalized mode)"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithString("output", mcp.Description("Output directory (default: .agent/workflows/)")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateWorkflows}
}

// commitGuidanceTool defines the commit_guidance MCP knowledge tool.
func commitGuidanceTool() server.ServerTool {
	tool := mcp.NewTool("commit_guidance",
		mcp.WithDescription("Conventional Commits behavioral context. Returns the spec and instructions for generating proper commit messages. No API key needed."),
		mcp.WithString("locale", mcp.Description("Language for the guidance: en or es"), mcp.DefaultString("en")),
	)

	return server.ServerTool{Tool: tool, Handler: handleCommitGuidance}
}

// versionGuidanceTool defines the version_guidance MCP knowledge tool.
func versionGuidanceTool() server.ServerTool {
	tool := mcp.NewTool("version_guidance",
		mcp.WithDescription("Semantic Versioning behavioral context. Returns the spec and instructions for determining version bumps from conventional commits. No API key needed."),
		mcp.WithString("locale", mcp.Description("Language for the guidance: en or es"), mcp.DefaultString("en")),
	)

	return server.ServerTool{Tool: tool, Handler: handleVersionGuidance}
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

func handleGenerateSkills(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categoryName := stringArg(request, "category")
	preset := stringArg(request, "preset")
	mode := stringArgDefault(request, "mode", dto.SkillModeStatic)
	projectContext := stringArg(request, "project_context")
	locale := stringArgDefault(request, "locale", "en")
	target := stringArgDefault(request, "target", "claude")
	model := stringArgDefault(request, "model", "")
	output := stringArg(request, "output")
	if output == "" {
		output = defaultSkillsPath(target)
	}

	// Resolver categoría y preset desde el catálogo
	cat, err := catalog.FindCategory(categoryName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid category: %v", err)), nil
	}

	selection, err := cat.Resolve(preset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid preset: %v", err)), nil
	}

	// Cargar templates
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(
		root.TemplatesFS, filepath.Join("templates", locale, "skills", selection.TemplateDir), selection.TemplateMapping,
	)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load templates: %v", err)), nil
	}

	config := &dto.SkillsConfig{
		Category:       cat.Name,
		Preset:         preset,
		Mode:           mode,
		Locale:         locale,
		Target:         target,
		Model:          model,
		OutputPath:     output,
		ProjectContext: projectContext,
	}

	var result *dto.GenerationResult

	if mode == dto.SkillModePersonalized {
		if projectContext == "" {
			return mcp.NewToolResultError("personalized mode requires project_context parameter"), nil
		}
		result, err = executePersonalizedSkillsMCP(ctx, config, guides)
	} else {
		result, err = executeStaticSkillsMCP(config, guides)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Skills generation failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Agent skills delivered (category: %s, preset: %s, mode: %s, target: %s)\n", categoryName, preset, mode, target))
	sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
	if result.Model != "" && result.Model != "static" {
		sb.WriteString(fmt.Sprintf("Model: %s\n", result.Model))
		sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut))
	}
	sb.WriteString("\nGenerated skills:\n")
	for _, f := range result.GeneratedFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGenerateWorkflows(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	preset := stringArg(request, "preset")
	mode := stringArgDefault(request, "mode", dto.SkillModeStatic)
	projectContext := stringArg(request, "project_context")
	locale := stringArgDefault(request, "locale", "en")
	model := stringArgDefault(request, "model", "")
	output := stringArg(request, "output")
	if output == "" {
		output = filepath.Join(".agent", "workflows")
	}

	// Resolver categoría y preset desde el catálogo de workflows
	cat, err := catalog.FindWorkflowCategory("workflows")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid workflow category: %v", err)), nil
	}

	selection, err := cat.Resolve(preset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid preset: %v", err)), nil
	}

	// Cargar templates
	templatePath := filepath.Join("templates", locale, selection.TemplateDir)
	templateLoader := infratemplate.NewFileSystemTemplateLoaderWithMapping(
		root.TemplatesFS, templatePath, selection.TemplateMapping,
	)
	guides, err := templateLoader.LoadAll()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load workflow templates: %v", err)), nil
	}

	config := &dto.WorkflowConfig{
		Category:       "workflows",
		Preset:         preset,
		Mode:           mode,
		Locale:         locale,
		Model:          model,
		OutputPath:     output,
		ProjectContext: projectContext,
	}

	var result *dto.GenerationResult

	if mode == dto.SkillModePersonalized {
		if projectContext == "" {
			return mcp.NewToolResultError("personalized mode requires project_context parameter"), nil
		}
		result, err = executePersonalizedWorkflowsMCP(ctx, config, guides)
	} else {
		result, err = executeStaticWorkflowsMCP(config, guides)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workflow generation failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Antigravity workflows delivered (preset: %s, mode: %s)\n", preset, mode))
	sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
	if result.Model != "" && result.Model != "static" {
		sb.WriteString(fmt.Sprintf("Model: %s\n", result.Model))
		sb.WriteString(fmt.Sprintf("Tokens: %d in / %d out\n", result.TokensIn, result.TokensOut))
	}
	sb.WriteString("\nGenerated workflows:\n")
	for _, f := range result.GeneratedFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleCommitGuidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := stringArgDefault(request, "locale", "en")
	content, err := loadKnowledgeTemplate(locale, "workflow", "conventional_commit.template")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load commit guidance: %v", err)), nil
	}
	return mcp.NewToolResultText(content), nil
}

func handleVersionGuidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := stringArgDefault(request, "locale", "en")
	content, err := loadKnowledgeTemplate(locale, "workflow", "semantic_versioning.template")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load version guidance: %v", err)), nil
	}
	return mcp.NewToolResultText(content), nil
}

// loadKnowledgeTemplate reads an embedded template and returns its content as behavioral context.
func loadKnowledgeTemplate(locale, preset, filename string) (string, error) {
	path := filepath.Join("templates", locale, "skills", preset, filename)
	data, err := root.TemplatesFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("template not found: %s", path)
	}
	return string(data), nil
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
		templateLoader = infratemplate.NewFileSystemTemplateLoaderWithLanguage(root.TemplatesFS, templatePath, localeBase, language)
	} else {
		templateLoader = infratemplate.NewFileSystemTemplateLoader(root.TemplatesFS, templatePath)
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

	config := &dto.ProjectConfig{
		Name:        name,
		Description: description,
		Language:    language,
		Model:       model,
		OutputPath:  ".",
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
		root.TemplatesFS, filepath.Join("templates", locale, "spec"), specTemplateMapping,
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

	config := &dto.SpecConfig{
		ProjectName:     name,
		FromContextPath: fromContextPath,
		OutputPath:      fromContextPath,
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

func executeStaticSkillsMCP(config *dto.SkillsConfig, guides []service.TemplateGuide) (*dto.GenerationResult, error) {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := command.NewDeliverStaticSkillsCommand(fileWriter, dirManager)
	return cmd.Execute(config, guides)
}

func executePersonalizedSkillsMCP(ctx context.Context, config *dto.SkillsConfig, guides []service.TemplateGuide) (*dto.GenerationResult, error) {
	apiKey, err := llm.ResolveAPIKey(config.Model)
	if err != nil {
		return nil, err
	}

	provider, err := llm.NewProvider(ctx, config.Model, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	skillsCmd := command.NewGenerateSkillsCommand(provider, fileWriter, dirManager)

	return skillsCmd.Execute(ctx, config, guides)
}

func executeStaticWorkflowsMCP(config *dto.WorkflowConfig, guides []service.TemplateGuide) (*dto.GenerationResult, error) {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := command.NewDeliverStaticWorkflowsCommand(fileWriter, dirManager)
	return cmd.Execute(config, guides)
}

func executePersonalizedWorkflowsMCP(ctx context.Context, config *dto.WorkflowConfig, guides []service.TemplateGuide) (*dto.GenerationResult, error) {
	apiKey, err := llm.ResolveAPIKey(config.Model)
	if err != nil {
		return nil, err
	}

	provider, err := llm.NewProvider(ctx, config.Model, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	workflowsCmd := command.NewGenerateWorkflowsCommand(provider, fileWriter, dirManager)

	return workflowsCmd.Execute(ctx, config, guides)
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

// defaultSkillsPath returns the ecosystem-specific default skills directory.
func defaultSkillsPath(target string) string {
	switch target {
	case "codex", "antigravity":
		return filepath.Join(".agents", "skills")
	default: // claude
		return filepath.Join(".claude", "skills")
	}
}
