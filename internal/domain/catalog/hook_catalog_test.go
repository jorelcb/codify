package catalog

import "testing"

func TestFindHookCategory(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"hooks", false},
		{"unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := FindHookCategory(tt.name)
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

func TestHookCategoryNames(t *testing.T) {
	names := HookCategoryNames()
	if len(names) != 1 {
		t.Fatalf("got %d categories, want 1", len(names))
	}
	if names[0] != "hooks" {
		t.Errorf("unexpected name: %s", names[0])
	}
}

func TestHookResolve_Presets(t *testing.T) {
	cat, _ := FindHookCategory("hooks")

	presets := []string{"linting", "security-guardrails", "convention-enforcement"}

	for _, p := range presets {
		t.Run(p, func(t *testing.T) {
			sel, err := cat.Resolve(p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.TemplateDir != "hooks/"+p {
				t.Errorf("got dir %q, want hooks/%s", sel.TemplateDir, p)
			}
			if sel.TemplateMapping != nil {
				t.Errorf("expected nil mapping (full-directory copy), got %v", sel.TemplateMapping)
			}
		})
	}
}

func TestHookResolve_UnknownPreset(t *testing.T) {
	cat, _ := FindHookCategory("hooks")
	_, err := cat.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for unknown preset, got nil")
	}
}

func TestHookPresetNames(t *testing.T) {
	names := HookPresetNames()
	if len(names) != 3 {
		t.Fatalf("got %d preset names, want 3", len(names))
	}
	want := map[string]bool{"linting": true, "security-guardrails": true, "convention-enforcement": true}
	for _, n := range names {
		if !want[n] {
			t.Errorf("unexpected preset name: %s", n)
		}
	}
}

func TestHookMetadata_DescriptionLength(t *testing.T) {
	for name, meta := range HookMetadata {
		if len(meta.Description) > 250 {
			t.Errorf("hook %q description exceeds 250 chars: %d", name, len(meta.Description))
		}
		if meta.Description == "" {
			t.Errorf("hook %q has empty description", name)
		}
	}
}

func TestHookMetadata_PresetCoverage(t *testing.T) {
	// Every option in HookCategories must have a corresponding HookMetadata entry.
	for _, cat := range HookCategories {
		for _, opt := range cat.Options {
			if _, ok := HookMetadata[opt.Name]; !ok {
				t.Errorf("preset %q has no HookMetadata entry", opt.Name)
			}
		}
	}
}
