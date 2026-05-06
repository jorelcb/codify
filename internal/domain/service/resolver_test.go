package service

import (
	"strings"
	"testing"
)

func TestScanMarkers_DetectsHintedAndBareForms(t *testing.T) {
	content := strings.Join([]string{
		"# Title",
		"",
		"The currency is [DEFINE: ISO 4217 code], and",
		"the timezone is [DEFINE].",
		"",
		"Multiple on one line: [DEFINE: a] and [DEFINE: b].",
	}, "\n")

	hits := ScanMarkers(content)

	if got, want := len(hits), 4; got != want {
		t.Fatalf("expected %d hits, got %d: %+v", want, got, hits)
	}
	wantTexts := []string{"[DEFINE: ISO 4217 code]", "[DEFINE]", "[DEFINE: a]", "[DEFINE: b]"}
	for i, h := range hits {
		if h.Text != wantTexts[i] {
			t.Errorf("hit[%d].Text = %q, want %q", i, h.Text, wantTexts[i])
		}
	}
}

func TestScanMarkers_LineNumbersAre1Based(t *testing.T) {
	content := "line1\nline2 [DEFINE: x]\nline3\nline4 [DEFINE]"
	hits := ScanMarkers(content)

	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].Line != 2 {
		t.Errorf("first hit line = %d, want 2", hits[0].Line)
	}
	if hits[1].Line != 4 {
		t.Errorf("second hit line = %d, want 4", hits[1].Line)
	}
}

func TestScanMarkers_NoMarkersReturnsEmpty(t *testing.T) {
	if hits := ScanMarkers("plain markdown without placeholders"); hits != nil {
		t.Fatalf("expected nil, got %+v", hits)
	}
}

func TestScanMarkers_EmptyContent(t *testing.T) {
	if hits := ScanMarkers(""); hits != nil {
		t.Fatalf("expected nil, got %+v", hits)
	}
}

func TestScanMarkers_IgnoresMarkerSpanningNewline(t *testing.T) {
	// The regex excludes newlines inside the hint, so a marker that crosses a
	// line break should not match. This mirrors the LLM contract: hints are
	// single-line.
	content := "broken [DEFINE: this hint\nspans lines]"
	if hits := ScanMarkers(content); hits != nil {
		t.Fatalf("expected nil, got %+v", hits)
	}
}

func TestLiteralSubstitute_ReplacesAnsweredOnly(t *testing.T) {
	content := "Currency is [DEFINE: code], timezone is [DEFINE: tz]."
	hits := []MarkerHit{
		{Text: "[DEFINE: code]", Line: 1, Answer: "USD"},
		{Text: "[DEFINE: tz]", Line: 1, Answer: ""}, // skipped
	}

	got := LiteralSubstitute(content, hits)
	want := "Currency is USD, timezone is [DEFINE: tz]."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLiteralSubstitute_NoAnswers_IsNoop(t *testing.T) {
	content := "Currency is [DEFINE: code]."
	hits := []MarkerHit{{Text: "[DEFINE: code]", Line: 1, Answer: ""}}

	if got := LiteralSubstitute(content, hits); got != content {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestLiteralSubstitute_DuplicateMarkerOnlyReplacesOnce(t *testing.T) {
	// Two hits with the same Text but on different lines: each iteration must
	// consume exactly one occurrence so they map to their respective spots.
	content := "[DEFINE: x]\n[DEFINE: x]"
	hits := []MarkerHit{
		{Text: "[DEFINE: x]", Line: 1, Answer: "first"},
		{Text: "[DEFINE: x]", Line: 2, Answer: "second"},
	}

	got := LiteralSubstitute(content, hits)
	want := "first\nsecond"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
