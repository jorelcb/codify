package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
)

// fakeProvider is a minimal service.LLMProvider stub for enricher tests.
type fakeProvider struct {
	respText string
	err      error
	captured service.EvaluationRequest
}

func (f *fakeProvider) GenerateContext(_ context.Context, _ service.GenerationRequest) (*service.GenerationResponse, error) {
	return nil, errors.New("not used")
}

func (f *fakeProvider) EvaluatePrompt(_ context.Context, req service.EvaluationRequest) (*service.EvaluationResponse, error) {
	f.captured = req
	if f.err != nil {
		return nil, f.err
	}
	return &service.EvaluationResponse{Text: f.respText}, nil
}

func TestEnrich_HappyPath_MapsFindingsToHits(t *testing.T) {
	provider := &fakeProvider{
		respText: `[
  {
    "marker_text": "[DEFINE: ISO 4217 code]",
    "question": "¿Qué moneda usa la aplicación?",
    "suggestions": ["USD", "EUR"],
    "default": "USD",
    "rationale": "El archivo menciona pagos internacionales."
  }
]`,
	}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: ISO 4217 code]", Line: 42}}

	out, err := enricher.Enrich(context.Background(), "AGENTS.md", "fake content", "es", hits)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 enriched marker, got %d", len(out))
	}
	if out[0].Question != "¿Qué moneda usa la aplicación?" {
		t.Errorf("question: %q", out[0].Question)
	}
	if len(out[0].Suggestions) != 2 || out[0].Default != "USD" {
		t.Errorf("suggestions/default: %+v / %q", out[0].Suggestions, out[0].Default)
	}
}

func TestEnrich_RequestUsesCacheableSystem(t *testing.T) {
	provider := &fakeProvider{respText: "[]"}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: x]", Line: 1}}

	if _, err := enricher.Enrich(context.Background(), "f.md", "content", "en", hits); err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if !provider.captured.CacheableSystem {
		t.Error("enrichment should set CacheableSystem=true so Anthropic can cache the system prompt")
	}
	if provider.captured.Command != "resolve-enrich" {
		t.Errorf("command tag should be resolve-enrich for usage tracking, got %q", provider.captured.Command)
	}
}

func TestEnrich_StripsMarkdownFences(t *testing.T) {
	provider := &fakeProvider{
		respText: "```json\n[{\"marker_text\":\"[DEFINE: x]\",\"question\":\"q\",\"suggestions\":[],\"default\":\"\",\"rationale\":\"\"}]\n```",
	}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: x]", Line: 1}}

	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", hits)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if len(out) != 1 || out[0].Question != "q" {
		t.Errorf("fence-wrapped JSON should still parse: %+v", out)
	}
}

func TestEnrich_InvalidJSON_FallsBackToZeroValueEnrichment(t *testing.T) {
	provider := &fakeProvider{respText: "definitely not JSON"}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: x]", Line: 1}}

	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", hits)
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
	// Even with error, the slice should still be populated with the original
	// hits so the orchestrator can keep walking the prompter loop.
	if len(out) != 1 || out[0].Text != "[DEFINE: x]" {
		t.Errorf("fallback should preserve hits: %+v", out)
	}
	if out[0].Question != "" || len(out[0].Suggestions) != 0 {
		t.Errorf("fallback enrichment must be zero-valued: %+v", out[0])
	}
}

func TestEnrich_ProviderError_FallsBackWithError(t *testing.T) {
	provider := &fakeProvider{err: errors.New("network down")}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: x]", Line: 1}}

	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", hits)
	if err == nil {
		t.Fatal("expected error when provider fails")
	}
	if len(out) != 1 || out[0].Text != "[DEFINE: x]" {
		t.Errorf("fallback should include input hit: %+v", out)
	}
}

func TestEnrich_LLMOmitsMarker_GetsZeroValueEntry(t *testing.T) {
	provider := &fakeProvider{
		respText: `[{"marker_text":"[DEFINE: a]","question":"q","suggestions":["x"],"default":"","rationale":""}]`,
	}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{
		{Text: "[DEFINE: a]", Line: 1},
		{Text: "[DEFINE: b]", Line: 2}, // LLM did NOT enrich this one
	}

	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", hits)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("output length must equal input length, got %d", len(out))
	}
	if out[0].Question == "" {
		t.Errorf("first hit should have enrichment, got %+v", out[0])
	}
	if out[1].Question != "" || len(out[1].Suggestions) != 0 {
		t.Errorf("second hit should be zero-valued, got %+v", out[1])
	}
}

func TestEnrich_FilterAppliesViaSanitizer(t *testing.T) {
	provider := &fakeProvider{
		respText: `[{"marker_text":"[DEFINE: x]","question":"q","suggestions":["valid","https://hallucinated.com"],"default":"","rationale":""}]`,
	}
	enricher := NewLLMEnricher(provider)
	hits := []service.MarkerHit{{Text: "[DEFINE: x]", Line: 1}}

	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", hits)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if len(out[0].Suggestions) != 1 || out[0].Suggestions[0] != "valid" {
		t.Errorf("URL suggestion should be filtered: %+v", out[0].Suggestions)
	}
}

func TestNewLLMEnricher_NilProvider_ReturnsNil(t *testing.T) {
	if enricher := NewLLMEnricher(nil); enricher != nil {
		t.Errorf("expected nil enricher for nil provider, got %+v", enricher)
	}
}

func TestEnrich_EmptyHits_NoCallToProvider(t *testing.T) {
	provider := &fakeProvider{respText: "should-not-be-read"}
	enricher := NewLLMEnricher(provider)
	out, err := enricher.Enrich(context.Background(), "f.md", "c", "en", nil)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil for empty hits, got %+v", out)
	}
	if provider.captured.UserPrompt != "" {
		t.Error("provider should not be called for empty hits")
	}
}
