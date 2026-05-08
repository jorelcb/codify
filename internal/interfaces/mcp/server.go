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

const serverVersion = "2.2.0"

// validContextPresets enumerates accepted preset names for context generation
// (generate_context + analyze_project tools). The "default" alias was removed
// in v2.0 (ADR-001 phase 3); the new built-in default is "neutral".
var validContextPresets = map[string]bool{
	"clean-ddd":    true,
	"neutral":      true,
	"hexagonal":    true,
	"event-driven": true,
	"workflow":     true,
}

// normalizeContextPreset validates the preset name. In v2.0 this returns an
// error for "default" (removed) and unknown presets — MCP callers see the
// error in the response. The CLI side has equivalent error handling in
// resolvePreset (cli/commands/generate.go).
func normalizeContextPreset(preset string) (string, error) {
	if preset == "default" {
		return "", fmt.Errorf("preset 'default' was removed in Codify v2.0.0. Use preset='clean-ddd' to keep v1.x behavior, or preset='neutral' (the new default) for no architectural opinion")
	}
	if !validContextPresets[preset] {
		return "", fmt.Errorf("unknown preset %q. Valid presets: neutral, clean-ddd, hexagonal, event-driven", preset)
	}
	return preset, nil
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
		generateHooksTool(),
		commitGuidanceTool(),
		versionGuidanceTool(),
		getUsageTool(),
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
		mcp.WithString("preset", mcp.Description("Template preset for context. Options: neutral (default — no architectural opinion), clean-ddd (DDD + Clean Architecture), hexagonal (Ports & Adapters), event-driven (CQRS + Event Sourcing + Sagas), workflow."), mcp.Enum("neutral", "clean-ddd", "hexagonal", "event-driven", "workflow"), mcp.DefaultString("neutral")),
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
		mcp.WithString("preset", mcp.Description("Template preset for context. Options: neutral (default — no architectural opinion), clean-ddd (DDD + Clean Architecture), hexagonal (Ports & Adapters), event-driven (CQRS + Event Sourcing + Sagas)."), mcp.Enum("neutral", "clean-ddd", "hexagonal", "event-driven"), mcp.DefaultString("neutral")),
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
		mcp.WithString("category", mcp.Required(), mcp.Description("Skill category"), mcp.Enum(catalog.CategoryNames()...)),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Preset within category (or 'all' where supported). architecture: clean, neutral. testing: foundational, tdd, bdd. conventions: conventional-commit, semantic-versioning, all"), mcp.Enum(catalog.AllSkillPresetNames()...)),
		mcp.WithString("mode", mcp.Description("Generation mode"), mcp.Enum("static", "personalized"), mcp.DefaultString("static")),
		mcp.WithString("project_context", mcp.Description("Project description for personalized mode (language, architecture, domain, stack)")),
		mcp.WithString("locale", mcp.Description("Output language"), mcp.Enum("en", "es"), mcp.DefaultString("en")),
		mcp.WithString("target", mcp.Description("Target ecosystem"), mcp.Enum("claude", "codex", "antigravity"), mcp.DefaultString("claude")),
		mcp.WithString("model", mcp.Description("LLM model (only for personalized mode)"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithString("output", mcp.Description("Output directory (default: ecosystem-specific, e.g. .claude/skills/)")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateSkills}
}

// generateWorkflowsTool defines the generate_workflows MCP tool.
func generateWorkflowsTool() server.ServerTool {
	tool := mcp.NewTool("generate_workflows",
		mcp.WithDescription("Generate workflow files for AI agents. Claude target produces native SKILL.md files with frontmatter. Antigravity target produces .md files with execution annotations. Static mode is instant. Personalized mode uses LLM."),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Workflow preset"), mcp.Enum(catalog.WorkflowPresetNames()...)),
		mcp.WithString("target", mcp.Description("Target ecosystem"), mcp.Enum("claude", "antigravity"), mcp.DefaultString("antigravity")),
		mcp.WithString("mode", mcp.Description("Generation mode"), mcp.Enum("static", "personalized"), mcp.DefaultString("static")),
		mcp.WithString("project_context", mcp.Description("Project description for personalized mode (language, tools, CI/CD, deployment)")),
		mcp.WithString("locale", mcp.Description("Output language"), mcp.Enum("en", "es"), mcp.DefaultString("en")),
		mcp.WithString("model", mcp.Description("LLM model (only for personalized mode)"), mcp.DefaultString("claude-sonnet-4-6")),
		mcp.WithString("output", mcp.Description("Output directory (default depends on target)")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateWorkflows}
}

// generateHooksTool defines the generate_hooks MCP tool.
func generateHooksTool() server.ServerTool {
	tool := mcp.NewTool("generate_hooks",
		mcp.WithDescription("Activate Claude Code hook bundles. Default install_scope is 'preview' (writes a standalone bundle to 'output' for the user to merge). Set install_scope to 'global' or 'project' to auto-merge into settings.json + copy scripts into the agent's hooks directory. Static-only, Claude Code-only."),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Hook preset"), mcp.Enum(catalog.HookPresetNames()...)),
		mcp.WithString("install_scope", mcp.Description("Activation mode: global (auto-merge into ~/.claude), project (auto-merge into .claude), or preview (write bundle to output, no settings change)"), mcp.Enum("global", "project", "preview"), mcp.DefaultString("preview")),
		mcp.WithString("locale", mcp.Description("Output language for stderr messages"), mcp.Enum("en", "es"), mcp.DefaultString("en")),
		mcp.WithString("output", mcp.Description("Output directory (required for install_scope=preview; default: ./codify-hooks)")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGenerateHooks}
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
	preset := stringArgDefault(request, "preset", "neutral")
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
	preset := stringArgDefault(request, "preset", "neutral")
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

	// Format scan as description and generate with analyze mode
	description := scanResult.FormatAsDescription()
	result, err := executeGenerateWithMode(ctx, name, description, language, preset, locale, model, "analyze")
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
	target := stringArgDefault(request, "target", "antigravity")
	mode := stringArgDefault(request, "mode", dto.SkillModeStatic)
	projectContext := stringArg(request, "project_context")
	locale := stringArgDefault(request, "locale", "en")
	model := stringArgDefault(request, "model", "")
	output := stringArg(request, "output")
	if output == "" {
		if target == "claude" {
			output = filepath.Join(".claude", "skills")
		} else {
			output = filepath.Join(".agent", "workflows")
		}
	}

	if !dto.ValidWorkflowTargets[target] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid target: %s (available: claude, antigravity)", target)), nil
	}

	cat, err := catalog.FindWorkflowCategory("workflows")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid workflow category: %v", err)), nil
	}

	selection, err := cat.Resolve(preset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid preset: %v", err)), nil
	}

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
		Target:         target,
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

	targetLabel := "Antigravity"
	if target == "claude" {
		targetLabel = "Claude Code"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s workflows delivered (preset: %s, mode: %s)\n", targetLabel, preset, mode))
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

func handleGenerateHooks(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	preset := stringArg(request, "preset")
	locale := stringArgDefault(request, "locale", "en")
	output := stringArg(request, "output")
	scope := stringArgDefault(request, "install_scope", "preview")

	if !dto.ValidHookPresets[preset] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid hook preset: %s (valid: linting, security-guardrails, convention-enforcement, all)", preset)), nil
	}
	if locale != "en" && locale != "es" {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid locale: %s (must be 'en' or 'es')", locale)), nil
	}

	switch scope {
	case "preview":
		if output == "" {
			output = filepath.Join(".", "codify-hooks")
		}
		config := &dto.HookConfig{
			Category:   "hooks",
			Preset:     preset,
			Locale:     locale,
			OutputPath: output,
		}
		result, err := executePreviewHooksMCP(config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Hook bundle generation failed: %v", err)), nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Claude Code hook bundle written (preset: %s, mode: preview)\n", preset))
		sb.WriteString(fmt.Sprintf("Output: %s\n", result.OutputPath))
		sb.WriteString("\nGenerated files:\n")
		for _, f := range result.GeneratedFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", f))
		}
		sb.WriteString("\nThis is preview mode — settings.json was NOT modified.\n")
		sb.WriteString("Re-run with install_scope=global|project to auto-activate.\n")
		return mcp.NewToolResultText(sb.String()), nil

	case dto.InstallScopeGlobal, dto.InstallScopeProject:
		config := &dto.HookConfig{
			Category: "hooks",
			Preset:   preset,
			Locale:   locale,
			Install:  scope,
		}
		result, err := executeInstallHooksMCP(config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Hook activation failed: %v", err)), nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Claude Code hooks activated (preset: %s, scope: %s)\n", preset, scope))
		sb.WriteString(fmt.Sprintf("Settings: %s\n", result.SettingsPath))
		if result.BackupPath != "" {
			sb.WriteString(fmt.Sprintf("Backup:   %s\n", result.BackupPath))
		}
		sb.WriteString(fmt.Sprintf("Hooks dir: %s\n", result.HooksDir))
		if total := sumIntMap(result.HandlersAdded); total > 0 {
			sb.WriteString(fmt.Sprintf("Added:    %d handler(s) across %d event(s)\n", total, len(result.HandlersAdded)))
		}
		if total := sumIntMap(result.HandlersSkipped); total > 0 {
			sb.WriteString(fmt.Sprintf("Skipped:  %d handler(s) already present (idempotent)\n", total))
		}
		if len(result.ScriptsCopied) > 0 {
			sb.WriteString(fmt.Sprintf("Scripts copied: %d\n", len(result.ScriptsCopied)))
		}
		if len(result.ScriptsConflict) > 0 {
			sb.WriteString(fmt.Sprintf("Scripts in conflict: %d (existing differs — not overwritten)\n", len(result.ScriptsConflict)))
		}
		return mcp.NewToolResultText(sb.String()), nil

	default:
		return mcp.NewToolResultError(fmt.Sprintf("Invalid install_scope: %s (must be 'global', 'project', or 'preview')", scope)), nil
	}
}

func sumIntMap(m map[string]int) int {
	t := 0
	for _, v := range m {
		t += v
	}
	return t
}

func handleCommitGuidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := stringArgDefault(request, "locale", "en")
	content, err := loadKnowledgeTemplate(locale, "conventions", "conventional_commit.template")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load commit guidance: %v", err)), nil
	}
	return mcp.NewToolResultText(content), nil
}

func handleVersionGuidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locale := stringArgDefault(request, "locale", "en")
	content, err := loadKnowledgeTemplate(locale, "conventions", "semantic_versioning.template")
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
	return executeGenerateWithMode(ctx, name, description, language, preset, locale, model, "")
}

func executeGenerateWithMode(ctx context.Context, name, description, language, preset, locale, model, mode string) (*dto.GenerationResult, error) {
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return nil, err
	}

	preset, err = normalizeContextPreset(preset)
	if err != nil {
		return nil, err
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
		Mode:        mode,
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

func executePreviewHooksMCP(config *dto.HookConfig) (*dto.GenerationResult, error) {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	cmd := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)
	return cmd.Execute(config)
}

func executeInstallHooksMCP(config *dto.HookConfig) (*command.InstallResult, error) {
	fileWriter := filesystem.NewFileWriter()
	dirManager := filesystem.NewDirectoryManager()
	deliverer := command.NewDeliverHooksCommand(fileWriter, dirManager, root.TemplatesFS)
	installer := command.NewInstallHooksCommand(deliverer, fileWriter, dirManager)
	return installer.Execute(config)
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
