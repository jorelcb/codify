package catalog

import (
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/infrastructure/sdd"
)

func TestFindWorkflowCategory(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"workflows", false},
		{"unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := FindWorkflowCategory(tt.name)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cat.Name != tt.name {
				t.Errorf("got name %q, want %q", cat.Name, tt.name)
			}
		})
	}
}

func TestWorkflowCategoryNames(t *testing.T) {
	names := WorkflowCategoryNames()
	if len(names) != 1 {
		t.Fatalf("got %d categories, want 1", len(names))
	}
	if names[0] != "workflows" {
		t.Errorf("unexpected name: %s", names[0])
	}
}

func TestWorkflowResolve_Presets(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")

	presets := []struct {
		name             string
		dir              string
		expectedMappings int
	}{
		{"bug-fix", "workflows", 1},
		{"release-cycle", "workflows", 1},
		// spec-driven-change templates moved to sdd/openspec/workflows/
		// in C.2 (ADR-0011: SDD pluggable). Other workflow presets stay
		// at the top-level workflows/ directory.
		{"spec-driven-change", "sdd/openspec/workflows", 3},
	}

	for _, tt := range presets {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := cat.Resolve(tt.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.TemplateDir != tt.dir {
				t.Errorf("got dir %q, want %q", sel.TemplateDir, tt.dir)
			}
			if len(sel.TemplateMapping) != tt.expectedMappings {
				t.Errorf("got %d mappings, want %d", len(sel.TemplateMapping), tt.expectedMappings)
			}
		})
	}
}

func TestWorkflowResolve_All(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")

	sel, err := cat.Resolve("all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 single-file presets (bug-fix, release-cycle) +
	// 3 mappings from spec-driven-change (propose, apply, archive) = 5
	if len(sel.TemplateMapping) != 5 {
		t.Errorf("got %d mappings, want 5", len(sel.TemplateMapping))
	}
}

func TestWorkflowResolve_UnknownPreset(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")
	_, err := cat.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for unknown preset, got nil")
	}
}

func TestGenerateWorkflowFrontmatter_Antigravity(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("bug_fix", "antigravity")
	if !strings.HasPrefix(fm, "---\n") {
		t.Error("frontmatter should start with ---")
	}
	if !strings.Contains(fm, "description:") {
		t.Error("frontmatter should contain description field")
	}
	if !strings.HasSuffix(fm, "---\n") {
		t.Error("frontmatter should end with ---")
	}
	if strings.Contains(fm, "disable-model-invocation") {
		t.Error("antigravity frontmatter should not contain disable-model-invocation")
	}
	if strings.Contains(fm, "allowed-tools") {
		t.Error("antigravity frontmatter should not contain allowed-tools")
	}
	if strings.Contains(fm, "name:") {
		t.Error("antigravity frontmatter should not contain name field")
	}
}

func TestGenerateWorkflowFrontmatter_Claude(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("spec_propose", "claude")
	if !strings.HasPrefix(fm, "---\n") {
		t.Error("frontmatter should start with ---")
	}
	if !strings.Contains(fm, "name: spec-propose") {
		t.Errorf("claude frontmatter should contain name field, got: %s", fm)
	}
	if !strings.Contains(fm, "description:") {
		t.Error("claude frontmatter should contain description field")
	}
	if !strings.Contains(fm, "disable-model-invocation: true") {
		t.Error("claude frontmatter should contain disable-model-invocation: true")
	}
	if !strings.Contains(fm, "allowed-tools: Bash(*)") {
		t.Error("claude frontmatter should contain allowed-tools: Bash(*)")
	}
	if !strings.HasSuffix(fm, "---\n") {
		t.Error("frontmatter should end with ---")
	}
}

func TestGenerateWorkflowFrontmatter_Unknown(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("unknown_workflow", "antigravity")
	if !strings.Contains(fm, "Workflow for unknown-workflow") {
		t.Errorf("expected fallback description, got: %s", fm)
	}
}

func TestGenerateWorkflowFrontmatter_UnknownClaude(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("unknown_workflow", "claude")
	if !strings.Contains(fm, "Workflow for unknown-workflow") {
		t.Errorf("expected fallback description, got: %s", fm)
	}
	if !strings.Contains(fm, "disable-model-invocation: true") {
		t.Error("claude frontmatter should contain disable-model-invocation: true")
	}
}

func TestStripAnnotationLines(t *testing.T) {
	content := `### 1. Create Branch
// capture: BRANCH_NAME
Create a new branch.

### 2. Run Tests
// turbo
Run test suite.

### 3. Plan
// if the change is large
Break it down.
`
	result := StripAnnotationLines(content)

	if strings.Contains(result, "// capture:") {
		t.Error("should strip capture annotations")
	}
	if strings.Contains(result, "// turbo") {
		t.Error("should strip turbo annotations")
	}
	if strings.Contains(result, "// if ") {
		t.Error("should strip if annotations")
	}
	if !strings.Contains(result, "Create a new branch.") {
		t.Error("should preserve non-annotation content")
	}
	if !strings.Contains(result, "Run test suite.") {
		t.Error("should preserve non-annotation content")
	}
	if !strings.Contains(result, "### 1. Create Branch") {
		t.Error("should preserve step headers")
	}
}

func TestWorkflowMetadata_DescriptionLength(t *testing.T) {
	for name, meta := range WorkflowMetadata {
		if len(meta.Description) > 250 {
			t.Errorf("workflow %q description exceeds 250 chars: %d", name, len(meta.Description))
		}
	}
}

func TestWorkflowPresetNames_IncludesAllAlias(t *testing.T) {
	names := WorkflowPresetNames()
	if len(names) == 0 {
		t.Fatal("WorkflowPresetNames returned empty")
	}
	hasAll := false
	for _, n := range names {
		if n == "all" {
			hasAll = true
		}
	}
	if !hasAll {
		t.Error("'all' alias missing from workflow preset names")
	}
}

// TestResolveWithSpecStandard_OpenSpec valida que el preset spec-driven-change
// produce los workflows OpenSpec cuando el adapter OpenSpec está activo.
// Espejo del comportamiento estático histórico — debe ser idéntico al fallback.
func TestResolveWithSpecStandard_OpenSpec(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")
	std := sdd.NewOpenSpecAdapter()

	sel, err := cat.ResolveWithSpecStandard("spec-driven-change", std)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.TemplateDir != "sdd/openspec/workflows" {
		t.Errorf("got dir %q, want sdd/openspec/workflows", sel.TemplateDir)
	}
	if len(sel.TemplateMapping) != 3 {
		t.Errorf("got %d mappings, want 3", len(sel.TemplateMapping))
	}
	for _, want := range []string{"spec_propose.template", "spec_apply.template", "spec_archive.template"} {
		if _, ok := sel.TemplateMapping[want]; !ok {
			t.Errorf("expected template %q in mapping, got %v", want, sel.TemplateMapping)
		}
	}
}

// TestResolveWithSpecStandard_SpecKit valida que el mismo preset produce los
// workflows Spec-Kit (specify/plan/tasks) cuando el adapter Spec-Kit está
// activo. Es la verificación principal del wiring SDD-aware.
func TestResolveWithSpecStandard_SpecKit(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")
	std := sdd.NewSpecKitAdapter()

	sel, err := cat.ResolveWithSpecStandard("spec-driven-change", std)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.TemplateDir != "sdd/spec-kit/workflows" {
		t.Errorf("got dir %q, want sdd/spec-kit/workflows", sel.TemplateDir)
	}
	if len(sel.TemplateMapping) != 3 {
		t.Errorf("got %d mappings, want 3", len(sel.TemplateMapping))
	}
	for _, want := range []string{"speckit_specify.template", "speckit_plan.template", "speckit_tasks.template"} {
		if _, ok := sel.TemplateMapping[want]; !ok {
			t.Errorf("expected template %q in mapping, got %v", want, sel.TemplateMapping)
		}
	}
}

// TestResolveWithSpecStandard_NilFallsBackToStatic confirma que pasar
// std=nil reproduce el comportamiento del Resolve legacy (campos estáticos
// del preset, que apuntan a OpenSpec). Es lo que mantiene los BDD scenarios
// de v2.2.0 verdes sin tocarlos.
func TestResolveWithSpecStandard_NilFallsBackToStatic(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")

	sel, err := cat.ResolveWithSpecStandard("spec-driven-change", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.TemplateDir != "sdd/openspec/workflows" {
		t.Errorf("nil std should use OpenSpec fallback, got %q", sel.TemplateDir)
	}
}

// TestResolveWithSpecStandard_NonSDDPresetIgnoresStandard confirma que
// pasar un SpecStandard a un preset no-SDD (bug-fix, release-cycle) no
// afecta el resultado. Garantiza que el adapter no se filtra a presets
// que no lo necesitan.
func TestResolveWithSpecStandard_NonSDDPresetIgnoresStandard(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")
	std := sdd.NewSpecKitAdapter() // un adapter no-default a propósito

	for _, preset := range []string{"bug-fix", "release-cycle"} {
		sel, err := cat.ResolveWithSpecStandard(preset, std)
		if err != nil {
			t.Fatalf("preset %q: unexpected error: %v", preset, err)
		}
		if sel.TemplateDir != "workflows" {
			t.Errorf("preset %q: SDD adapter must not affect non-SDD presets, got dir %q", preset, sel.TemplateDir)
		}
	}
}
