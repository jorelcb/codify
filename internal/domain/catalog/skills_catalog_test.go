package catalog

import (
	"testing"
)

func TestFindCategory(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"architecture", false},
		{"testing", false},
		{"conventions", false},
		{"unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := FindCategory(tt.name)
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

func TestCategoryResolve_Exclusive(t *testing.T) {
	cat, _ := FindCategory("architecture")

	// Sub-opciones individuales deben funcionar
	sel, err := cat.Resolve("clean")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.TemplateDir != "default" {
		t.Errorf("got dir %q, want %q", sel.TemplateDir, "default")
	}
	if len(sel.TemplateMapping) != 5 {
		t.Errorf("got %d mappings, want 5", len(sel.TemplateMapping))
	}

	// "all" debe fallar en categorías exclusivas
	_, err = cat.Resolve("all")
	if err == nil {
		t.Error("expected error for 'all' on exclusive category, got nil")
	}
}

func TestCategoryResolve_NonExclusive(t *testing.T) {
	cat, _ := FindCategory("conventions")

	// Sub-opción individual
	sel, err := cat.Resolve("conventional-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sel.TemplateMapping) != 1 {
		t.Errorf("got %d mappings, want 1", len(sel.TemplateMapping))
	}

	// "all" debe funcionar y combinar todos los mappings
	sel, err = cat.Resolve("all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sel.TemplateMapping) != 2 {
		t.Errorf("got %d mappings, want 2", len(sel.TemplateMapping))
	}
}

func TestCategoryResolve_Testing(t *testing.T) {
	cat, _ := FindCategory("testing")

	// Exclusive category: "all" should fail
	_, err := cat.Resolve("all")
	if err == nil {
		t.Error("expected error for 'all' on exclusive testing category, got nil")
	}

	// Each preset maps to exactly 1 template
	for _, preset := range []string{"foundational", "tdd", "bdd"} {
		sel, err := cat.Resolve(preset)
		if err != nil {
			t.Fatalf("unexpected error for preset %q: %v", preset, err)
		}
		if sel.TemplateDir != "testing" {
			t.Errorf("preset %q: got dir %q, want %q", preset, sel.TemplateDir, "testing")
		}
		if len(sel.TemplateMapping) != 1 {
			t.Errorf("preset %q: got %d mappings, want 1", preset, len(sel.TemplateMapping))
		}
	}
}

func TestCategoryResolve_UnknownPreset(t *testing.T) {
	cat, _ := FindCategory("architecture")
	_, err := cat.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for unknown preset, got nil")
	}
}

func TestCategoryNames(t *testing.T) {
	names := CategoryNames()
	if len(names) != 3 {
		t.Fatalf("got %d categories, want 3", len(names))
	}
	if names[0] != "architecture" || names[1] != "testing" || names[2] != "conventions" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestLegacyPresetMapping(t *testing.T) {
	tests := []struct {
		legacy   string
		wantCat  string
		wantPre  string
		wantOk   bool
	}{
		{"default", "architecture", "clean", true},
		{"neutral", "architecture", "neutral", true},
		{"workflow", "conventions", "all", true},
		{"conventions", "conventions", "all", true},
		{"unknown", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.legacy, func(t *testing.T) {
			mapped, ok := LegacyPresetMapping[tt.legacy]
			if ok != tt.wantOk {
				t.Fatalf("ok=%v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if mapped[0] != tt.wantCat || mapped[1] != tt.wantPre {
				t.Errorf("got [%s, %s], want [%s, %s]", mapped[0], mapped[1], tt.wantCat, tt.wantPre)
			}
		})
	}
}

func TestOptionNamesAndLabels(t *testing.T) {
	cat, _ := FindCategory("conventions")

	names := cat.OptionNames()
	if len(names) != 2 {
		t.Fatalf("got %d options, want 2", len(names))
	}

	labels := cat.OptionLabels()
	if len(labels) != 2 {
		t.Fatalf("got %d labels, want 2", len(labels))
	}
	if labels[0] != "Conventional Commits" {
		t.Errorf("unexpected label: %s", labels[0])
	}
}
