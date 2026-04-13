package catalog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPluginName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"feature_development", "codify-wf-feature-development"},
		{"bug_fix", "codify-wf-bug-fix"},
		{"release_cycle", "codify-wf-release-cycle"},
	}
	for _, tt := range tests {
		got := PluginName(tt.input)
		if got != tt.expected {
			t.Errorf("PluginName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGeneratePluginManifest_ValidJSON(t *testing.T) {
	result := GeneratePluginManifest("release_cycle", "Release process: version bump, changelog, tag, and deploy")

	var manifest map[string]interface{}
	if err := json.Unmarshal([]byte(result), &manifest); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if manifest["name"] != "codify-wf-release-cycle" {
		t.Errorf("expected name 'codify-wf-release-cycle', got %v", manifest["name"])
	}
	if manifest["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", manifest["version"])
	}
	if manifest["description"] != "Release process: version bump, changelog, tag, and deploy" {
		t.Errorf("unexpected description: %v", manifest["description"])
	}
	author, ok := manifest["author"].(map[string]interface{})
	if !ok || author["name"] != "codify" {
		t.Errorf("expected author name 'codify', got %v", manifest["author"])
	}
}

func TestGeneratePluginHooks_TurboAnnotations(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "turbo", Step: 3, StepName: "Update Version"},
		{Type: "turbo", Step: 5, StepName: "Create Commit"},
	}

	result := GeneratePluginHooks(annotations)

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(result), &config); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	hooks, ok := config["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'hooks' key in config")
	}

	preToolUse, ok := hooks["PreToolUse"].([]interface{})
	if !ok {
		t.Fatal("expected 'PreToolUse' key in hooks")
	}
	if len(preToolUse) < 1 {
		t.Fatal("expected at least 1 PreToolUse entry")
	}

	if !strings.Contains(result, "permissionDecision") {
		t.Error("expected hooks to contain 'permissionDecision'")
	}
	if !strings.Contains(result, "allow") {
		t.Error("expected hooks to contain 'allow'")
	}
}

func TestGeneratePluginHooks_CaptureAnnotations(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "capture", Step: 2, Value: "NEW_VERSION"},
	}

	result := GeneratePluginHooks(annotations)

	if !strings.Contains(result, "PostToolUse") {
		t.Error("expected hooks to contain 'PostToolUse'")
	}
	if !strings.Contains(result, "capture-output.sh") {
		t.Error("expected hooks to reference capture-output.sh script")
	}
	if !strings.Contains(result, "${CLAUDE_PLUGIN_ROOT}") {
		t.Error("expected hooks to reference ${CLAUDE_PLUGIN_ROOT}")
	}
}

func TestGeneratePluginHooks_IfAnnotations(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "if", Step: 8, Value: "the project has CI/CD deployment"},
	}

	result := GeneratePluginHooks(annotations)

	if !strings.Contains(result, "prompt") {
		t.Error("expected hooks to contain prompt type")
	}
	if !strings.Contains(result, "CI/CD deployment") {
		t.Error("expected hooks to contain the condition text")
	}
}

func TestGeneratePluginHooks_MixedAnnotations(t *testing.T) {
	annotations := []AnnotationMeta{
		{Type: "turbo", Step: 3},
		{Type: "capture", Step: 2, Value: "NEW_VERSION"},
		{Type: "if", Step: 8, Value: "the project has CI/CD"},
		{Type: "if", Step: 9, Value: "using GitHub releases"},
	}

	result := GeneratePluginHooks(annotations)

	var config hooksConfig
	if err := json.Unmarshal([]byte(result), &config); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// 1 turbo entry + 2 if entries = 3 PreToolUse entries
	if len(config.Hooks["PreToolUse"]) != 3 {
		t.Errorf("expected 3 PreToolUse entries, got %d", len(config.Hooks["PreToolUse"]))
	}
	// 1 capture entry
	if len(config.Hooks["PostToolUse"]) != 1 {
		t.Errorf("expected 1 PostToolUse entry, got %d", len(config.Hooks["PostToolUse"]))
	}
}

func TestGeneratePluginHooks_NoAnnotations(t *testing.T) {
	annotations := []AnnotationMeta{}
	result := GeneratePluginHooks(annotations)

	var config hooksConfig
	if err := json.Unmarshal([]byte(result), &config); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(config.Hooks) != 0 {
		t.Errorf("expected empty hooks, got %d entries", len(config.Hooks))
	}
}

func TestTransformToPluginSkill_StripAnnotations(t *testing.T) {
	content := `# Feature Development Workflow

### 1. Create Feature Branch
// capture: BRANCH_NAME
Create a new branch from the latest main.

### 5. Run Full Test Suite
// turbo
Run the complete test suite.

### 8. Address Review Feedback
// if there is review feedback
Process each review comment.`

	result := TransformToPluginSkill("feature_development", content)

	if strings.Contains(result, "// capture:") {
		t.Error("expected annotations to be stripped: // capture")
	}
	if strings.Contains(result, "// turbo") {
		t.Error("expected annotations to be stripped: // turbo")
	}
	if strings.Contains(result, "// if ") {
		t.Error("expected annotations to be stripped: // if")
	}
}

func TestTransformToPluginSkill_PreserveContent(t *testing.T) {
	content := `# Feature Development Workflow

### 1. Create Feature Branch
// capture: BRANCH_NAME
Create a new branch from the latest main.
- Pull latest changes
- Push to establish tracking`

	result := TransformToPluginSkill("feature_development", content)

	if !strings.Contains(result, "Create a new branch from the latest main.") {
		t.Error("expected non-annotation content to be preserved")
	}
	if !strings.Contains(result, "Pull latest changes") {
		t.Error("expected bullet points to be preserved")
	}
}

func TestTransformToPluginSkill_HasFrontmatter(t *testing.T) {
	content := `# Feature Development Workflow

### 1. Step One
Do something.`

	result := TransformToPluginSkill("feature_development", content)

	if !strings.HasPrefix(result, "---\n") {
		t.Error("expected YAML frontmatter at start")
	}
	if !strings.Contains(result, "name: feature-development") {
		t.Error("expected name in frontmatter")
	}
	if !strings.Contains(result, "description: Full feature lifecycle") {
		t.Error("expected description in frontmatter")
	}
}

func TestTransformToPluginSkill_UnknownPreset(t *testing.T) {
	content := `# Custom Workflow
### 1. Step One
Do something.`

	result := TransformToPluginSkill("custom_workflow", content)

	if !strings.Contains(result, "name: custom-workflow") {
		t.Error("expected fallback name")
	}
	if !strings.Contains(result, "Workflow for custom-workflow") {
		t.Error("expected fallback description")
	}
}

func TestGenerateWorkflowAgent_English(t *testing.T) {
	result := GenerateWorkflowAgent("release_cycle", "en")

	if !strings.Contains(result, "name: workflow-runner") {
		t.Error("expected agent name in frontmatter")
	}
	if !strings.Contains(result, "model: sonnet") {
		t.Error("expected model in frontmatter")
	}
	if !strings.Contains(result, "tools: Bash, Read, Edit, Write, Grep, Glob") {
		t.Error("expected tools in frontmatter")
	}
	if !strings.Contains(result, "maxTurns: 50") {
		t.Error("expected maxTurns in frontmatter")
	}
	if !strings.Contains(result, "release-cycle") {
		t.Error("expected preset name in agent description")
	}
	if !strings.Contains(result, "workflow execution agent") {
		t.Error("expected English instructions")
	}
}

func TestGenerateWorkflowAgent_Spanish(t *testing.T) {
	result := GenerateWorkflowAgent("feature_development", "es")

	if !strings.Contains(result, "agente de ejecucion de workflow") {
		t.Error("expected Spanish instructions")
	}
	if !strings.Contains(result, "feature-development") {
		t.Error("expected preset name in agent")
	}
}