package catalog

import (
	"strings"
	"testing"
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
		name string
		dir  string
	}{
		{"feature-development", "workflows"},
		{"bug-fix", "workflows"},
		{"release-cycle", "workflows"},
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
			if len(sel.TemplateMapping) != 1 {
				t.Errorf("got %d mappings, want 1", len(sel.TemplateMapping))
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
	if len(sel.TemplateMapping) != 3 {
		t.Errorf("got %d mappings, want 3", len(sel.TemplateMapping))
	}
}

func TestWorkflowResolve_UnknownPreset(t *testing.T) {
	cat, _ := FindWorkflowCategory("workflows")
	_, err := cat.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for unknown preset, got nil")
	}
}

func TestGenerateWorkflowFrontmatter(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("feature_development")
	if !strings.HasPrefix(fm, "---\n") {
		t.Error("frontmatter should start with ---")
	}
	if !strings.Contains(fm, "description:") {
		t.Error("frontmatter should contain description field")
	}
	if !strings.HasSuffix(fm, "---\n") {
		t.Error("frontmatter should end with ---")
	}
}

func TestGenerateWorkflowFrontmatter_Unknown(t *testing.T) {
	fm := GenerateWorkflowFrontmatter("unknown_workflow")
	if !strings.Contains(fm, "Workflow for unknown-workflow") {
		t.Errorf("expected fallback description, got: %s", fm)
	}
}

func TestWorkflowMetadata_DescriptionLength(t *testing.T) {
	for name, meta := range WorkflowMetadata {
		if len(meta.Description) > 250 {
			t.Errorf("workflow %q description exceeds 250 chars: %d", name, len(meta.Description))
		}
	}
}
