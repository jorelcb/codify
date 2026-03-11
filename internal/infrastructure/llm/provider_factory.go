package llm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jorelcb/ai-context-generator/internal/domain/service"
)

func envOrEmpty(key string) string {
	return os.Getenv(key)
}

// geminiModelPrefixes identifies Gemini models by prefix.
var geminiModelPrefixes = []string{
	"gemini-",
}

// isGeminiModel returns true if the model name matches a Gemini model.
func isGeminiModel(model string) bool {
	lower := strings.ToLower(model)
	for _, prefix := range geminiModelPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// NewProvider creates the appropriate LLMProvider based on the model name.
// If model is empty or starts with "claude-", uses Anthropic.
// If model starts with "gemini-", uses Google Gemini.
func NewProvider(ctx context.Context, model string, apiKey string, progressOut io.Writer) (service.LLMProvider, error) {
	if isGeminiModel(model) {
		return NewGeminiProvider(ctx, apiKey, model, progressOut)
	}

	// Default: Anthropic
	return NewAnthropicProvider(apiKey, model, progressOut), nil
}

// ResolveAPIKey returns the API key for the given model from the appropriate env var.
// For Gemini models, checks GEMINI_API_KEY then GOOGLE_API_KEY.
// For Anthropic models, checks ANTHROPIC_API_KEY.
func ResolveAPIKey(model string) (string, error) {
	if isGeminiModel(model) {
		return resolveGeminiAPIKey()
	}
	return resolveAnthropicAPIKey()
}

func resolveAnthropicAPIKey() (string, error) {
	key := envOrEmpty("ANTHROPIC_API_KEY")
	if key == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}
	return key, nil
}

func resolveGeminiAPIKey() (string, error) {
	if key := envOrEmpty("GEMINI_API_KEY"); key != "" {
		return key, nil
	}
	if key := envOrEmpty("GOOGLE_API_KEY"); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required")
}

// DefaultModel returns the default model name for the given provider.
func DefaultModel(model string) string {
	if isGeminiModel(model) {
		return defaultGeminiModel
	}
	return defaultModel
}
