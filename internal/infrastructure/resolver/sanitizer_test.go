package resolver

import (
	"strings"
	"testing"
)

func TestSanitizeFinding_DropsURLs(t *testing.T) {
	got := SanitizeFinding(
		"What protocol?",
		[]string{"https://example.com", "http", "ftp://foo"},
		"https://example.com",
		"why",
	)
	if len(got.Suggestions) != 1 || got.Suggestions[0] != "http" {
		t.Errorf("URL suggestions should be dropped, kept: %+v", got.Suggestions)
	}
	if got.Default != "" {
		t.Errorf("default should drop because matched suggestion was filtered, got %q", got.Default)
	}
}

func TestSanitizeFinding_DropsPaths(t *testing.T) {
	got := SanitizeFinding(
		"Which file?",
		[]string{"/etc/passwd", "./config.json", "../secret", "config.yaml"},
		"",
		"",
	)
	if len(got.Suggestions) != 1 || got.Suggestions[0] != "config.yaml" {
		t.Errorf("paths should be dropped, kept: %+v", got.Suggestions)
	}
}

func TestSanitizeFinding_DropsTooLong(t *testing.T) {
	long := strings.Repeat("a", maxSuggestionLength+1)
	got := SanitizeFinding("?", []string{"short", long}, "", "")
	if len(got.Suggestions) != 1 || got.Suggestions[0] != "short" {
		t.Errorf("over-length suggestion should be dropped, kept: %+v", got.Suggestions)
	}
}

func TestSanitizeFinding_DropsMultiline(t *testing.T) {
	got := SanitizeFinding("?", []string{"single", "multi\nline"}, "", "")
	if len(got.Suggestions) != 1 || got.Suggestions[0] != "single" {
		t.Errorf("multiline suggestion should be dropped, kept: %+v", got.Suggestions)
	}
}

func TestSanitizeFinding_DropsMarkdownFenced(t *testing.T) {
	got := SanitizeFinding("?", []string{"plain", "```code```"}, "", "")
	if len(got.Suggestions) != 1 || got.Suggestions[0] != "plain" {
		t.Errorf("markdown-fenced suggestion should be dropped, kept: %+v", got.Suggestions)
	}
}

func TestSanitizeFinding_DeduplicatesCaseInsensitive(t *testing.T) {
	got := SanitizeFinding("?", []string{"USD", "usd", "EUR"}, "", "")
	if len(got.Suggestions) != 2 {
		t.Errorf("expected dedup to leave 2 entries, got %+v", got.Suggestions)
	}
	if got.Suggestions[0] != "USD" || got.Suggestions[1] != "EUR" {
		t.Errorf("first-seen order should be preserved, got %+v", got.Suggestions)
	}
}

func TestSanitizeFinding_CapsAtMaxKept(t *testing.T) {
	got := SanitizeFinding("?", []string{"a", "b", "c", "d", "e"}, "", "")
	if len(got.Suggestions) != maxSuggestionsKept {
		t.Errorf("expected cap at %d, got %d", maxSuggestionsKept, len(got.Suggestions))
	}
}

func TestSanitizeFinding_DefaultMustMatchKeptSuggestion(t *testing.T) {
	got := SanitizeFinding("?", []string{"USD", "EUR"}, "MXN", "")
	if got.Default != "" {
		t.Errorf("default that doesn't match any kept suggestion must be dropped, got %q", got.Default)
	}
}

func TestSanitizeFinding_DefaultMatchIsCaseInsensitive(t *testing.T) {
	got := SanitizeFinding("?", []string{"USD"}, "usd", "")
	if got.Default != "USD" {
		t.Errorf("default should normalize to the kept-suggestion casing, got %q", got.Default)
	}
}

func TestSanitizeFinding_TrimsAndTruncates(t *testing.T) {
	longQ := strings.Repeat("q", maxQuestionLength+50)
	got := SanitizeFinding("  "+longQ+"  ", nil, "", "")
	if len(got.Question) != maxQuestionLength {
		t.Errorf("expected truncation to %d, got %d", maxQuestionLength, len(got.Question))
	}
	if !strings.HasSuffix(got.Question, "...") {
		t.Errorf("truncated string should end in ellipsis, got %q", got.Question)
	}
}

func TestSanitizeFinding_EmptyInputsReturnEmpty(t *testing.T) {
	got := SanitizeFinding("", nil, "", "")
	if got.Question != "" || got.Default != "" || got.Rationale != "" || len(got.Suggestions) != 0 {
		t.Errorf("expected all-empty result, got %+v", got)
	}
}
