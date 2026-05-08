package sdd

import (
	"strings"
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
)

func TestNewDefaultRegistry_RegistersOpenSpec(t *testing.T) {
	r := NewDefaultRegistry()

	std, err := r.Lookup("openspec")
	if err != nil {
		t.Fatalf("expected openspec to be registered: %v", err)
	}
	if std.ID() != "openspec" {
		t.Errorf("got ID %q, want %q", std.ID(), "openspec")
	}
}

func TestRegistry_Lookup_UnknownReturnsExplicitError(t *testing.T) {
	r := NewDefaultRegistry()

	_, err := r.Lookup("does-not-exist")
	if err == nil {
		t.Fatalf("expected error for unregistered standard")
	}
	msg := err.Error()
	// El error debe listar los IDs disponibles para que el usuario sepa
	// qué tiene a mano.
	if !strings.Contains(msg, "openspec") {
		t.Errorf("error should list available IDs, got: %s", msg)
	}
	if !strings.Contains(msg, "does-not-exist") {
		t.Errorf("error should mention the rejected ID, got: %s", msg)
	}
}

func TestNewDefaultRegistry_RegistersSpecKit(t *testing.T) {
	r := NewDefaultRegistry()

	std, err := r.Lookup("spec-kit")
	if err != nil {
		t.Fatalf("expected spec-kit to be registered: %v", err)
	}
	if std.ID() != "spec-kit" {
		t.Errorf("got ID %q, want spec-kit", std.ID())
	}
}

func TestRegistry_Resolve_PrecedenceFlagWins(t *testing.T) {
	r := NewDefaultRegistry()

	// Al pasar el flag explícito, ese gana sobre project, user y default.
	// El único standard registrado es openspec, así que validamos que el
	// orden no se rompa cuando todos los slots tienen "openspec".
	std, err := r.Resolve("openspec", "openspec", "openspec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if std.ID() != "openspec" {
		t.Errorf("got ID %q, want openspec", std.ID())
	}
}

func TestRegistry_Resolve_FlagOverridesConfig(t *testing.T) {
	r := NewDefaultRegistry()

	// Si project/user piden un standard distinto al del flag, el flag gana.
	// Validamos que el resultado matchea exactamente el flag.
	std, err := r.Resolve("openspec", "spec-kit", "spec-kit")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if std.ID() != "openspec" {
		t.Errorf("flag should win: got %q, want openspec", std.ID())
	}

	// Inverso: project pide openspec, flag pide spec-kit. Flag gana.
	std, err = r.Resolve("spec-kit", "openspec", "openspec")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if std.ID() != "spec-kit" {
		t.Errorf("flag should win: got %q, want spec-kit", std.ID())
	}
}

func TestRegistry_Resolve_FallsThroughToDefault(t *testing.T) {
	r := NewDefaultRegistry()

	// Sin flag, sin config, debe fallback al DefaultStandardID.
	std, err := r.Resolve("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if std.ID() != DefaultStandardID {
		t.Errorf("got ID %q, want default %q", std.ID(), DefaultStandardID)
	}
}

func TestRegistry_Resolve_InvalidNonEmptyFlagFails(t *testing.T) {
	r := NewDefaultRegistry()

	// Si el flag está seteado a un ID desconocido, el resolve debe
	// fallar — preferimos error explícito sobre fallback silencioso.
	_, err := r.Resolve("does-not-exist", "openspec", "openspec")
	if err == nil {
		t.Fatalf("expected error for unknown flag value")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should mention the rejected flag value, got: %s", err.Error())
	}
}

func TestRegistry_Resolve_InvalidProjectIDFailsBeforeUserFallback(t *testing.T) {
	r := NewDefaultRegistry()

	// Si el flag está vacío y project tiene un ID desconocido, debe fallar
	// en lugar de hacer fallback silencioso a user/default. El usuario
	// puso ese valor en .codify/config.yml deliberadamente; ignorarlo
	// silenciosamente generaría confusión.
	_, err := r.Resolve("", "does-not-exist", "openspec")
	if err == nil {
		t.Fatalf("expected error for unknown project standard")
	}
}

func TestRegistry_Register_AllowsOverride(t *testing.T) {
	r := NewDefaultRegistry()

	// Re-registrar un ID existente debe sobreescribir silenciosamente.
	// Pensado para tests, wirings alternativos, y eventual package
	// manager (Track D) que puede shipear adapters propios.
	stub := stubAdapter{id: "openspec", display: "Stub"}
	r.Register(stub)

	std, err := r.Lookup("openspec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if std.DisplayName() != "Stub" {
		t.Errorf("override did not take effect: got DisplayName %q", std.DisplayName())
	}
}

func TestRegistry_AvailableIDs_AlphabeticalOrder(t *testing.T) {
	r := NewDefaultRegistry()
	r.Register(stubAdapter{id: "alpha", display: "Alpha (stub)"})

	ids := r.AvailableIDs()
	// Default registry ya tiene openspec + spec-kit; agregamos alpha.
	if len(ids) != 3 {
		t.Fatalf("expected 3 IDs, got %d (%v)", len(ids), ids)
	}
	// Sorted: alpha, openspec, spec-kit
	expected := []string{"alpha", "openspec", "spec-kit"}
	for i, want := range expected {
		if ids[i] != want {
			t.Errorf("position %d: got %q, want %q", i, ids[i], want)
		}
	}
}

// stubAdapter es un SpecStandard mínimo para tests. Vive aquí en lugar de
// en testdata/ para evitar export accidental al production code.
type stubAdapter struct {
	id      string
	display string
}

func (s stubAdapter) ID() string                             { return s.id }
func (s stubAdapter) DisplayName() string                    { return s.display }
func (s stubAdapter) BootstrapArtifacts() []service.SpecArtifact { return nil }
func (s stubAdapter) OutputLayout() service.OutputLayout     { return service.LayoutFlat }
func (s stubAdapter) TemplateDir() string                    { return s.id }
func (s stubAdapter) SystemPromptHints(locale string) string { return "" }
func (s stubAdapter) LifecycleWorkflowIDs() []string         { return nil }
