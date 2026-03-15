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
// If model starts with "gemini-", uses Google Gemini.
// If model starts with "claude-" or is explicitly set, uses Anthropic.
// If model is empty, auto-detects provider from the API key.
func NewProvider(ctx context.Context, model string, apiKey string, progressOut io.Writer) (service.LLMProvider, error) {
	if isGeminiModel(model) {
		return NewGeminiProvider(ctx, apiKey, model, progressOut)
	}
	if model != "" {
		return NewAnthropicProvider(apiKey, model, progressOut), nil
	}
	// Auto-detect: if ANTHROPIC_API_KEY was resolved, use Anthropic; otherwise Gemini
	if envOrEmpty("ANTHROPIC_API_KEY") != "" {
		return NewAnthropicProvider(apiKey, model, progressOut), nil
	}
	return NewGeminiProvider(ctx, apiKey, "", progressOut)
}

// ResolveAPIKey returns the API key for the given model from the appropriate env var.
// For Gemini models, checks GEMINI_API_KEY then GOOGLE_API_KEY.
// For Anthropic models, checks ANTHROPIC_API_KEY.
// When no model is specified (empty string), tries Anthropic first, then Gemini.
func ResolveAPIKey(model string) (string, error) {
	if isGeminiModel(model) {
		return resolveGeminiAPIKey()
	}
	if model != "" {
		return resolveAnthropicAPIKey()
	}
	// No model specified: try Anthropic first, fall back to Gemini
	if key := envOrEmpty("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}
	if key := envOrEmpty("GEMINI_API_KEY"); key != "" {
		return key, nil
	}
	if key := envOrEmpty("GOOGLE_API_KEY"); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("ANTHROPIC_API_KEY or GEMINI_API_KEY environment variable is required")
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
// When model is empty, auto-detects based on available API keys.
func DefaultModel(model string) string {
	if isGeminiModel(model) {
		return defaultGeminiModel
	}
	if model != "" {
		return defaultModel
	}
	// Auto-detect
	if envOrEmpty("ANTHROPIC_API_KEY") != "" {
		return defaultModel
	}
	if envOrEmpty("GEMINI_API_KEY") != "" || envOrEmpty("GOOGLE_API_KEY") != "" {
		return defaultGeminiModel
	}
	return defaultModel
}
