package service

import (
	"regexp"
	"strings"
)

// defineMarkerRE matches the [DEFINE] / [DEFINE: hint] placeholders the LLM
// emits in generated context files when it lacks information. The dual form
// (with and without colon-hint) is intentional: bare [DEFINE] is legacy from
// pre-v2.0.4, kept matched for backward compatibility on previously-generated
// files. New generations always include a colon-hint per prompt rules.
var defineMarkerRE = regexp.MustCompile(`\[DEFINE(?::[^\]\n]+)?\]`)

// MarkerHit captures a single [DEFINE] occurrence inside a file: its verbatim
// text, 1-based line number, and the user's answer (empty when skipped).
//
// Lives in the domain layer because the marker concept is part of the
// product's vocabulary, not an implementation detail.
type MarkerHit struct {
	Text   string // verbatim "[DEFINE: hint]" or "[DEFINE]"
	Line   int    // 1-based
	Answer string // empty = user chose to skip
}

// EnrichedMarker decorates a MarkerHit with LLM-derived metadata used by the
// interactive prompter to render a friendlier question. Phase 3 populates the
// Question / Suggestions / Default fields. Until then, only the embedded hit
// matters and the rest stay zero-valued so the prompter falls back to the
// legacy "show hint, ask for input" UI.
type EnrichedMarker struct {
	MarkerHit
	Question    string   // natural-language question in the file's locale
	Suggestions []string // grounded suggestions inferred from file context
	Default     string   // optional default — must be one of Suggestions or empty
	Rationale   string   // brief reasoning (shown as helper text)
}

// PromptedAnswer is what the InteractivePrompter returns for each marker.
// Skip is preferred over checking Answer == "" because Phase 3 will accept
// inputs like "s" / "skip" as explicit skip signals — distinguishing those
// from a user who typed a literal "s" as their answer.
type PromptedAnswer struct {
	Answer string
	Skip   bool
}

// InteractivePrompter abstracts the user-facing question loop so the
// orchestrator can be exercised with a scripted prompter in tests. The CLI
// implementation wraps charmbracelet/huh; the test implementation replays a
// canned sequence.
type InteractivePrompter interface {
	// ConfirmTopLevel asks the global "Resolve N markers across M files?"
	// question. Returns true when the user accepts. Errors short-circuit the
	// flow upstream — the caller treats an error the same as a decline.
	ConfirmTopLevel(totalMarkers, totalFiles int) (bool, error)

	// AnnounceFile emits the per-file header. Implementations may render
	// formatting, log to stderr, or no-op (test prompters typically no-op).
	AnnounceFile(path string, markerCount int)

	// AskMarker prompts the user for one marker's answer. Implementations
	// receive the surrounding file content + the EnrichedMarker so they can
	// render context (lines around the marker) and any suggestions.
	AskMarker(fileContent string, marker EnrichedMarker) (PromptedAnswer, error)

	// ReportFileResult prints the per-file outcome. Mirrors AnnounceFile.
	ReportFileResult(path string, resolved int, mode string)
}

// ScanMarkers finds every [DEFINE]/[DEFINE: hint] occurrence in content and
// returns a slice of hits sorted by appearance. Pure function — no IO, no
// state. Used both by the orchestrator and by the post-rewrite validator
// (Phase 1).
func ScanMarkers(content string) []MarkerHit {
	var hits []MarkerHit
	for _, idx := range defineMarkerRE.FindAllStringIndex(content, -1) {
		hits = append(hits, MarkerHit{
			Text: content[idx[0]:idx[1]],
			Line: strings.Count(content[:idx[0]], "\n") + 1,
		})
	}
	return hits
}

// LiteralSubstitute replaces each answered marker with its answer text 1:1,
// preserving skipped markers verbatim. Pure function — used as fallback when
// no LLM provider is available, or as recovery path when the LLM rewrite
// fails or alters skipped markers (Phase 1 validator decision).
//
// Each marker text is replaced once per occurrence; this matters when two
// hits in the slice share the same Text but differ in Line (uncommon but
// possible in templated docs).
func LiteralSubstitute(content string, hits []MarkerHit) string {
	for _, h := range hits {
		if h.Answer == "" {
			continue
		}
		content = strings.Replace(content, h.Text, h.Answer, 1)
	}
	return content
}
