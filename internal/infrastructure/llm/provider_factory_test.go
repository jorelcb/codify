package llm

import (
	"testing"
)

func TestIsGeminiModel(t *testing.T) {
	cases := map[string]bool{
		"gemini-3.1-pro-preview": true,
		"gemini-1.5-flash":       true,
		"GEMINI-X":               true,
		"claude-sonnet-4-6":      false,
		"claude-opus-4-6":        false,
		"":                       false,
	}
	for in, want := range cases {
		if got := isGeminiModel(in); got != want {
			t.Errorf("isGeminiModel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestResolveAPIKey_GeminiPath(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "g-key")
	t.Setenv("ANTHROPIC_API_KEY", "")

	key, err := ResolveAPIKey("gemini-3.1-pro-preview")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "g-key" {
		t.Fatalf("got %q", key)
	}
}

func TestResolveAPIKey_GeminiFallsBackToGoogle(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "google-key")

	key, err := ResolveAPIKey("gemini-1.5")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "google-key" {
		t.Fatalf("got %q", key)
	}
}

func TestResolveAPIKey_AnthropicPath(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "a-key")
	key, err := ResolveAPIKey("claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "a-key" {
		t.Fatalf("got %q", key)
	}
}

func TestResolveAPIKey_NoModelTriesBoth(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "g")
	key, err := ResolveAPIKey("")
	if err != nil {
		t.Fatalf("ResolveAPIKey: %v", err)
	}
	if key != "g" {
		t.Fatalf("got %q", key)
	}
}

func TestResolveAPIKey_NoneSet(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	if _, err := ResolveAPIKey("claude-sonnet-4-6"); err == nil {
		t.Fatal("expected error when no API key is set")
	}
	if _, err := ResolveAPIKey(""); err == nil {
		t.Fatal("expected error for empty model with no keys")
	}
}

func TestDefaultModel_PassThrough(t *testing.T) {
	if got := DefaultModel("claude-opus-4-6"); got != "claude-opus-4-6" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultModel_PicksAnthropicWhenAvailable(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "x")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	if got := DefaultModel(""); got != defaultModel {
		t.Fatalf("got %q, want %q", got, defaultModel)
	}
}

func TestDefaultModel_PicksGeminiWhenOnlyKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "x")
	t.Setenv("GOOGLE_API_KEY", "")
	if got := DefaultModel(""); got != defaultGeminiModel {
		t.Fatalf("got %q, want %q", got, defaultGeminiModel)
	}
}

func TestOutputLanguageName_Fallback(t *testing.T) {
	if got := outputLanguageName("en"); got != "English" {
		t.Fatalf("got %q", got)
	}
	if got := outputLanguageName("es"); got != "Spanish" {
		t.Fatalf("got %q", got)
	}
	if got := outputLanguageName(""); got != "English" {
		t.Fatalf("empty should default to English, got %q", got)
	}
	// Unknown locale should still return English (and emit a warning).
	if got := outputLanguageName("fr"); got != "English" {
		t.Fatalf("unknown should default to English, got %q", got)
	}
}

func TestNewProvider_GeminiPath(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "x")
	p, err := NewProvider(t.Context(), "gemini-1.5", "x", nil)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if _, ok := p.(*GeminiProvider); !ok {
		t.Fatalf("expected GeminiProvider, got %T", p)
	}
}

func TestNewProvider_AnthropicPath(t *testing.T) {
	p, err := NewProvider(t.Context(), "claude-sonnet-4-6", "x", nil)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if _, ok := p.(*AnthropicProvider); !ok {
		t.Fatalf("expected AnthropicProvider, got %T", p)
	}
}
