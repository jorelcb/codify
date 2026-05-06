package resolver

import (
	"net/url"
	"strings"
)

// SanitizedFinding holds the post-sanitization fields ready for the prompter.
// All strings are trimmed; suggestions are deduplicated case-insensitively
// and pruned of values that look invented (URLs, paths, multi-line, too
// long, markdown).
type SanitizedFinding struct {
	Question    string
	Suggestions []string
	Default     string
	Rationale   string
}

const (
	maxSuggestionLength = 50  // chars; longer values are likely sentences not values
	maxQuestionLength   = 280 // soft cap; keep questions concise
	maxRationaleLength  = 280
	maxSuggestionsKept  = 3
)

// SanitizeFinding prunes hallucination-prone outputs from the LLM enricher.
// Pure function — no IO. The orchestrator uses it before passing the
// EnrichedMarker to the prompter.
//
// Rules:
//   - Trim every field. Empty question / rationale stays empty.
//   - Drop suggestions that are URLs, file paths, multi-line strings,
//     markdown-fenced, or longer than maxSuggestionLength.
//   - Deduplicate suggestions case-insensitively, preserving first-seen order.
//   - Keep at most maxSuggestionsKept suggestions (the LLM was told 2-3,
//     but defensively cap if it overproduces).
//   - Default must match (case-insensitively) one of the kept suggestions;
//     otherwise it's dropped.
//   - Truncate question and rationale to their max lengths if needed.
func SanitizeFinding(question string, suggestions []string, def, rationale string) SanitizedFinding {
	q := truncate(strings.TrimSpace(question), maxQuestionLength)
	r := truncate(strings.TrimSpace(rationale), maxRationaleLength)

	var kept []string
	seen := map[string]bool{}
	for _, s := range suggestions {
		c := cleanSuggestion(s)
		if c == "" {
			continue
		}
		key := strings.ToLower(c)
		if seen[key] {
			continue
		}
		seen[key] = true
		kept = append(kept, c)
		if len(kept) >= maxSuggestionsKept {
			break
		}
	}

	d := strings.TrimSpace(def)
	if d != "" {
		match := ""
		dl := strings.ToLower(d)
		for _, s := range kept {
			if strings.ToLower(s) == dl {
				match = s
				break
			}
		}
		d = match
	}

	return SanitizedFinding{
		Question:    q,
		Suggestions: kept,
		Default:     d,
		Rationale:   r,
	}
}

// cleanSuggestion returns "" when the suggestion looks invented. Otherwise
// returns the trimmed suggestion. Pure function.
func cleanSuggestion(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) > maxSuggestionLength {
		return ""
	}
	if strings.ContainsAny(s, "\n\r") {
		return ""
	}
	if strings.HasPrefix(s, "```") || strings.Contains(s, "```") {
		return ""
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return ""
	}
	if looksLikeURL(s) {
		return ""
	}
	return s
}

// looksLikeURL is a conservative URL detector — net/url accepts almost any
// input as a "URL" so we additionally check for a recognizable scheme prefix.
func looksLikeURL(s string) bool {
	for _, scheme := range []string{"http://", "https://", "ftp://", "file://", "git://"} {
		if strings.HasPrefix(strings.ToLower(s), scheme) {
			return true
		}
	}
	if u, err := url.Parse(s); err == nil && u.Scheme != "" && u.Host != "" {
		return true
	}
	return false
}

// truncate returns s clipped to max bytes. When s is longer, the last 3
// bytes of the result are an ASCII ellipsis. ASCII (not the UTF-8 …
// character) so the result fits the byte budget cleanly across all callers.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	const ellipsis = "..."
	if max <= len(ellipsis) {
		return s[:max]
	}
	return s[:max-len(ellipsis)] + ellipsis
}
