package sdd

import (
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
)

func TestSpecKitAdapter_BasicContract(t *testing.T) {
	a := NewSpecKitAdapter()

	if a.ID() != "spec-kit" {
		t.Errorf("ID: got %q, want spec-kit", a.ID())
	}
	if a.DisplayName() == "" {
		t.Error("DisplayName must be non-empty")
	}
	if a.TemplateDir() != "spec-kit" {
		t.Errorf("TemplateDir: got %q, want spec-kit", a.TemplateDir())
	}
	if a.OutputLayout() != service.LayoutFeatureGrouped {
		t.Errorf("OutputLayout: got %v, want LayoutFeatureGrouped", a.OutputLayout())
	}
}

func TestSpecKitAdapter_BootstrapArtifacts(t *testing.T) {
	a := NewSpecKitAdapter()
	arts := a.BootstrapArtifacts()

	// Spec-Kit ships at minimum spec/plan/tasks (required) plus the optional
	// research/data-model/quickstart trio.
	if len(arts) < 3 {
		t.Fatalf("expected at least 3 artifacts, got %d", len(arts))
	}

	// Required ones: spec.md, plan.md, tasks.md.
	required := map[string]bool{}
	for _, art := range arts {
		if art.Required {
			required[art.FileName] = true
		}
	}
	for _, want := range []string{"spec.md", "plan.md", "tasks.md"} {
		if !required[want] {
			t.Errorf("expected required artifact %q, not found in required set %v", want, required)
		}
	}

	// Naming convention: lowercase with hyphens. ZERO uppercase. Validates
	// Spec-Kit's editorial constraint.
	for _, art := range arts {
		for _, r := range art.FileName {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("Spec-Kit file names must be lowercase, got %q", art.FileName)
				break
			}
		}
	}
}

func TestSpecKitAdapter_DoesNotShipConstitution(t *testing.T) {
	a := NewSpecKitAdapter()
	for _, art := range a.BootstrapArtifacts() {
		// CONSTITUTION.md es OpenSpec-specific. Spec-Kit no lo usa.
		if strings.Contains(strings.ToLower(art.FileName), "constitution") {
			t.Errorf("Spec-Kit should not include constitution-like file, got %q", art.FileName)
		}
	}
}

func TestSpecKitAdapter_LifecycleWorkflowIDs(t *testing.T) {
	a := NewSpecKitAdapter()
	ids := a.LifecycleWorkflowIDs()

	if len(ids) == 0 {
		t.Fatal("Spec-Kit must declare lifecycle workflow IDs")
	}
	// Los workflows IDs son los slash commands (specify/plan/tasks).
	// Sus prefixes existen para no chocar con OpenSpec en el global mapping.
	for _, id := range ids {
		if !strings.HasPrefix(id, "speckit_") {
			t.Errorf("Spec-Kit workflow IDs should be namespaced (speckit_*), got %q", id)
		}
	}
}

func TestSpecKitAdapter_SystemPromptHints_MentionsLayout(t *testing.T) {
	a := NewSpecKitAdapter()

	// Las hints deben recordarle al LLM las dos diferencias críticas
	// frente a OpenSpec: layout per-feature y file names lowercase.
	hintsEN := a.SystemPromptHints("en")
	if !strings.Contains(strings.ToLower(hintsEN), "lowercase") {
		t.Errorf("EN hints should mention lowercase convention, got: %s", hintsEN)
	}
	if !strings.Contains(hintsEN, "specs/<feature-id>/") {
		t.Errorf("EN hints should mention specs/<feature-id>/ layout, got: %s", hintsEN)
	}

	hintsES := a.SystemPromptHints("es")
	if !strings.Contains(strings.ToLower(hintsES), "lowercase") {
		t.Errorf("ES hints should mention lowercase convention, got: %s", hintsES)
	}
	if !strings.Contains(hintsES, "specs/<feature-id>/") {
		t.Errorf("ES hints should mention specs/<feature-id>/ layout, got: %s", hintsES)
	}
}
