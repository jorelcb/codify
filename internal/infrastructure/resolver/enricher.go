// Package resolver hosts infrastructure adapters for the marker resolution
// flow defined in domain/service: the LLM-driven enricher (turns raw
// [DEFINE: ...] markers into user-friendly questions + suggestions + default)
// and a sanitizer that filters hallucinated suggestions before they reach
// the user.
package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// LLMEnricher implements service.MarkerEnricher by calling the configured
// LLM provider once per file. It expects a JSON array response matching
// llmFinding (commit-time-frozen schema, mirrored in the system prompt).
type LLMEnricher struct {
	provider service.LLMProvider
}

// NewLLMEnricher wires the enricher with the active provider. Pass nil to
// short-circuit enrichment — the orchestrator handles a nil enricher by
// falling back to the legacy UI.
func NewLLMEnricher(provider service.LLMProvider) *LLMEnricher {
	if provider == nil {
		return nil
	}
	return &LLMEnricher{provider: provider}
}

// enrichmentSystemPrompt instructs the LLM to translate technical markers
// into friendly questions WITHOUT inventing content. Anti-hallucination is
// enforced by:
//   - hard rule "Do not invent — empty is correct when context is insufficient"
//   - schema requires explicit suggestions=[] / default="" when uncertain
//   - sanitizer downstream filters out suggestions that look invented
//     (URLs, paths, multi-line strings, markdown, > 200 chars)
const enrichmentSystemPrompt = `You translate technical [DEFINE: ...] placeholders embedded in a generated context file into friendly questions for the end user. For each placeholder, return:

- A natural question in the file's locale (the LOCALE field below).
- 2-3 grounded suggestions inferred ONLY from the file's content or strongly implied context. Each suggestion is a short value (one or a few words) — not a sentence, not a URL, not a path.
- An optional default that must be one of the suggestions, OR empty when no suggestion is defensible.
- A brief rationale (one sentence) explaining why those suggestions, citing what in the file implies them.

Return ONLY a JSON array. No prose, no markdown fences. Schema:

[
  {
    "marker_text": "<exact marker as it appears in the file, e.g. [DEFINE: ISO 4217 code]>",
    "question": "<natural-language question in the target locale>",
    "suggestions": ["<short value 1>", "<short value 2>"],
    "default": "<one of suggestions, or empty string>",
    "rationale": "<one short sentence>"
  }
]

CRITICAL anti-hallucination rules:
- If you cannot infer suggestions safely, return suggestions=[] and default="". An empty list is correct — better than an invented list.
- Never propose URLs, file paths, or values longer than ~50 characters as suggestions.
- Each marker_text must match exactly one input placeholder. Do not paraphrase, do not normalize.
- Do not add markers that were not in the input.`

// llmFinding is the JSON schema for one enriched marker. Mirrors the system
// prompt 1:1; any change here must update the prompt.
type llmFinding struct {
	MarkerText  string   `json:"marker_text"`
	Question    string   `json:"question"`
	Suggestions []string `json:"suggestions"`
	Default     string   `json:"default"`
	Rationale   string   `json:"rationale"`
}

// fenceRE strips markdown fences some models still wrap responses in despite
// the system prompt. Best-effort cleanup before json.Unmarshal.
var fenceRE = regexp.MustCompile("(?s)\\A\\s*```(?:json)?\\s*\\n?|\\n?\\s*```\\s*\\z")

// Enrich calls the LLM and returns one EnrichedMarker per input hit. When
// the LLM omits a marker from its response (or returns invalid JSON), the
// returned slice still contains an entry for that hit with empty
// Question / Suggestions — the prompter degrades to the legacy UI.
func (e *LLMEnricher) Enrich(
	ctx context.Context,
	fileName, fileContent, locale string,
	hits []service.MarkerHit,
) ([]service.EnrichedMarker, error) {
	if len(hits) == 0 {
		return nil, nil
	}

	userPrompt := buildEnrichmentUserPrompt(fileName, fileContent, locale, hits)

	resp, err := e.provider.EvaluatePrompt(ctx, service.EvaluationRequest{
		SystemPrompt:    enrichmentSystemPrompt,
		UserPrompt:      userPrompt,
		Command:         "resolve-enrich",
		MaxTokens:       4096,
		CacheableSystem: true,
	})
	if err != nil {
		return enrichFallback(hits), fmt.Errorf("enrichment provider call failed: %w", err)
	}

	cleaned := strings.TrimSpace(fenceRE.ReplaceAllString(resp.Text, ""))
	var findings []llmFinding
	if err := json.Unmarshal([]byte(cleaned), &findings); err != nil {
		return enrichFallback(hits), fmt.Errorf("enrichment response is not valid JSON: %w (response snippet: %q)", err, snippet(cleaned, 200))
	}

	return mergeFindingsIntoHits(hits, findings), nil
}

// buildEnrichmentUserPrompt assembles the per-call user message with the
// file content and the list of markers to enrich. Locale flows into the
// prompt so the LLM produces questions in the user's language.
func buildEnrichmentUserPrompt(fileName, fileContent, locale string, hits []service.MarkerHit) string {
	var markers strings.Builder
	for _, h := range hits {
		fmt.Fprintf(&markers, "  - %s (line %d)\n", h.Text, h.Line)
	}
	return fmt.Sprintf(
		"FILE: %s\nLOCALE: %s\n\nMARKERS TO ENRICH:\n%s\n--- BEGIN FILE ---\n%s\n--- END FILE ---\n",
		fileName, locale, markers.String(), fileContent,
	)
}

// mergeFindingsIntoHits maps the LLM's findings back onto the input hits by
// exact marker_text match, applying the sanitizer to each finding. Hits the
// LLM omitted (or whose finding failed validation) get a zero-value
// EnrichedMarker — the prompter falls back to the legacy UI for those.
func mergeFindingsIntoHits(hits []service.MarkerHit, findings []llmFinding) []service.EnrichedMarker {
	byText := map[string]llmFinding{}
	for _, f := range findings {
		byText[f.MarkerText] = f
	}
	out := make([]service.EnrichedMarker, len(hits))
	for i, h := range hits {
		out[i] = service.EnrichedMarker{MarkerHit: h}
		if f, ok := byText[h.Text]; ok {
			cleaned := SanitizeFinding(f.Question, f.Suggestions, f.Default, f.Rationale)
			out[i].Question = cleaned.Question
			out[i].Suggestions = cleaned.Suggestions
			out[i].Default = cleaned.Default
			out[i].Rationale = cleaned.Rationale
		}
	}
	return out
}

// enrichFallback returns the input hits as zero-value EnrichedMarker entries.
// Used when enrichment fails (network, JSON parse, empty response) so the
// orchestrator can keep walking the prompter loop without enrichment.
func enrichFallback(hits []service.MarkerHit) []service.EnrichedMarker {
	out := make([]service.EnrichedMarker, len(hits))
	for i, h := range hits {
		out[i] = service.EnrichedMarker{MarkerHit: h}
	}
	return out
}

// snippet returns a leading slice of s for diagnostic logs, truncating with
// an ellipsis when the original exceeded max. Empty input becomes "<empty>"
// so the error surfaces the difference between fence-only and truncated JSON.
func snippet(s string, max int) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
