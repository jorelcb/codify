package llm

import (
	"context"
	"fmt"

	"github.com/jorelcb/codify/internal/domain/service"
)

// MockProvider is a deterministic stand-in for AnthropicProvider /
// GeminiProvider. It records every call and returns either preset responses
// or a templated default. Used by tests to exercise the orchestration layer
// (application/command) without hitting a real LLM API.
//
// Field semantics:
//   - Responses[guideName] → exact content to return for that guide. Optional.
//   - Default content is "# Mock {{name}}\n\nGenerated for testing." when no
//     entry is present.
//   - Each call appends to Calls, so tests can assert on what was sent.
type MockProvider struct {
	Responses map[string]string
	Calls     []service.GenerationRequest
	Err       error // when non-nil, GenerateContext returns this error instead.
	Tokens    struct {
		In  int
		Out int
	}
}

// NewMockProvider returns a MockProvider with default token counts.
func NewMockProvider() *MockProvider {
	mp := &MockProvider{
		Responses: map[string]string{},
	}
	mp.Tokens.In = 100
	mp.Tokens.Out = 200
	return mp
}

// GenerateContext implements service.LLMProvider for tests.
func (m *MockProvider) GenerateContext(_ context.Context, req service.GenerationRequest) (*service.GenerationResponse, error) {
	m.Calls = append(m.Calls, req)
	if m.Err != nil {
		return nil, m.Err
	}

	files := make([]service.GeneratedFile, 0, len(req.TemplateGuides))
	for _, g := range req.TemplateGuides {
		content, ok := m.Responses[g.Name]
		if !ok {
			content = fmt.Sprintf("# Mock %s\n\nGenerated for testing. Body padded to keep validators happy and exceed the truncation threshold of 200 characters so we don't trigger length warnings during test runs.", g.Name)
		}
		files = append(files, service.GeneratedFile{
			Name:    FileOutputName(g.Name),
			Content: content,
		})
	}

	return &service.GenerationResponse{
		Files:     files,
		Model:     "mock",
		TokensIn:  m.Tokens.In,
		TokensOut: m.Tokens.Out,
	}, nil
}

// LastCall returns the most recent GenerationRequest, or a zero value if
// none has been made.
func (m *MockProvider) LastCall() service.GenerationRequest {
	if len(m.Calls) == 0 {
		return service.GenerationRequest{}
	}
	return m.Calls[len(m.Calls)-1]
}
