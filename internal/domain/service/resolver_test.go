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

func TestSkipReplacement_MarkdownUsesHTMLComment(t *testing.T) {
	got := SkipReplacement("[DEFINE: ISO 4217 code]", ".md", SkipModeTODO, "2026-05-06")
	want := "<!-- TODO 2026-05-06: ISO 4217 code -->"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkipReplacement_GoUsesSlashComment(t *testing.T) {
	got := SkipReplacement("[DEFINE: error code]", ".go", SkipModeTODO, "2026-05-06")
	want := "// TODO 2026-05-06: error code"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkipReplacement_PythonUsesHashComment(t *testing.T) {
	got := SkipReplacement("[DEFINE: timezone]", ".py", SkipModeTODO, "2026-05-06")
	want := "# TODO 2026-05-06: timezone"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkipReplacement_UnknownExtension_ReturnsEmpty(t *testing.T) {
	got := SkipReplacement("[DEFINE: x]", ".xyz", SkipModeTODO, "2026-05-06")
	if got != "" {
		t.Errorf("expected empty for unknown ext, got %q", got)
	}
}

func TestSkipReplacement_VerbatimMode_ReturnsEmpty(t *testing.T) {
	got := SkipReplacement("[DEFINE: x]", ".md", SkipModeVerbatim, "2026-05-06")
	if got != "" {
		t.Errorf("verbatim must signal no replacement, got %q", got)
	}
}

func TestSkipReplacement_BareMarker_FallsBackToValueNeeded(t *testing.T) {
	got := SkipReplacement("[DEFINE]", ".md", SkipModeTODO, "2026-05-06")
	want := "<!-- TODO 2026-05-06: value needed -->"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkipReplacement_UnsetMode_ResolvesToTODO(t *testing.T) {
	got := SkipReplacement("[DEFINE: x]", ".md", SkipModeUnset, "2026-05-06")
	want := "<!-- TODO 2026-05-06: x -->"
	if got != want {
		t.Errorf("zero-value mode must default to TODO, got %q want %q", got, want)
	}
}

func TestSkipReplacement_CaseInsensitiveExtension(t *testing.T) {
	got := SkipReplacement("[DEFINE: x]", ".MD", SkipModeTODO, "2026-05-06")
	want := "<!-- TODO 2026-05-06: x -->"
	if got != want {
		t.Errorf("uppercase ext should still match, got %q", got)
	}
}

func TestApplySkipMode_PreservesAnsweredMarkers(t *testing.T) {
	content := "answered [DEFINE: a], skipped [DEFINE: b]"
	hits := []MarkerHit{
		{Text: "[DEFINE: a]", Answer: "alpha"},
		{Text: "[DEFINE: b]", Answer: ""},
	}

	got := ApplySkipMode(content, hits, SkipModeTODO, ".md", "2026-05-06")
	want := "answered [DEFINE: a], skipped <!-- TODO 2026-05-06: b -->"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestApplySkipMode_VerbatimIsNoop(t *testing.T) {
	content := "skipped [DEFINE: b]"
	hits := []MarkerHit{{Text: "[DEFINE: b]", Answer: ""}}

	got := ApplySkipMode(content, hits, SkipModeVerbatim, ".md", "2026-05-06")
	if got != content {
		t.Errorf("verbatim must not modify content, got %q", got)
	}
}

func TestApplySkipMode_Idempotent(t *testing.T) {
	content := "skipped [DEFINE: b]"
	hits := []MarkerHit{{Text: "[DEFINE: b]", Answer: ""}}

	once := ApplySkipMode(content, hits, SkipModeTODO, ".md", "2026-05-06")
	twice := ApplySkipMode(once, hits, SkipModeTODO, ".md", "2026-05-06")
	if once != twice {
		t.Errorf("expected idempotent, once=%q twice=%q", once, twice)
	}
}

func TestValidateRewrite_HappyPath_NoIssues(t *testing.T) {
	hits := []MarkerHit{
		{Text: "[DEFINE: code]", Line: 1, Answer: "USD"},
		{Text: "[DEFINE: tz]", Line: 2, Answer: ""},
	}
	after := "Currency is USD\n[DEFINE: tz]\n"

	d := ValidateRewrite(after, hits)

	if d.HasIssues() {
		t.Fatalf("unexpected issues: %+v", d)
	}
	if len(d.Resolved) != 1 || len(d.Skipped) != 1 {
		t.Errorf("counters: %+v", d)
	}
}

func TestValidateRewrite_LostSkippedMarker(t *testing.T) {
	hits := []MarkerHit{{Text: "[DEFINE: tz]", Line: 1, Answer: ""}}
	after := "Currency is USD\n" // LLM removed the skipped marker

	d := ValidateRewrite(after, hits)

	if !d.HasIssues() || len(d.Lost) != 1 {
		t.Fatalf("expected 1 Lost, got %+v", d)
	}
	if d.Lost[0].Text != "[DEFINE: tz]" {
		t.Errorf("Lost[0]: %+v", d.Lost[0])
	}
}

func TestValidateRewrite_SpuriousMarker(t *testing.T) {
	hits := []MarkerHit{{Text: "[DEFINE: code]", Line: 1, Answer: "USD"}}
	after := "Currency is USD but [DEFINE: invented_field] appeared"

	d := ValidateRewrite(after, hits)

	if !d.HasIssues() || len(d.Spurious) != 1 {
		t.Fatalf("expected 1 Spurious, got %+v", d)
	}
	if d.Spurious[0] != "[DEFINE: invented_field]" {
		t.Errorf("Spurious[0]: %q", d.Spurious[0])
	}
}

func TestValidateRewrite_NotApplied_AnsweredButPresent(t *testing.T) {
	hits := []MarkerHit{{Text: "[DEFINE: code]", Line: 1, Answer: "USD"}}
	after := "Currency is [DEFINE: code]" // LLM left the marker untouched

	d := ValidateRewrite(after, hits)

	if !d.HasIssues() || len(d.NotApplied) != 1 {
		t.Fatalf("expected 1 NotApplied, got %+v", d)
	}
}

func TestValidateRewrite_MixedIssues(t *testing.T) {
	hits := []MarkerHit{
		{Text: "[DEFINE: code]", Line: 1, Answer: "USD"}, // applied
		{Text: "[DEFINE: tz]", Line: 2, Answer: ""},      // should remain
		{Text: "[DEFINE: x]", Line: 3, Answer: "ans"},    // should disappear
	}
	// LLM: applied code, lost tz, did NOT apply x, hallucinated y
	after := "Currency USD, x is [DEFINE: x], extra [DEFINE: y]"

	d := ValidateRewrite(after, hits)

	if !d.HasIssues() {
		t.Fatal("expected issues")
	}
	if len(d.Lost) != 1 || d.Lost[0].Text != "[DEFINE: tz]" {
		t.Errorf("Lost: %+v", d.Lost)
	}
	if len(d.NotApplied) != 1 || d.NotApplied[0].Text != "[DEFINE: x]" {
		t.Errorf("NotApplied: %+v", d.NotApplied)
	}
	if len(d.Spurious) != 1 || d.Spurious[0] != "[DEFINE: y]" {
		t.Errorf("Spurious: %+v", d.Spurious)
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
