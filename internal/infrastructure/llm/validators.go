package llm

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ValidationResult summarizes structural issues detected in a generated file.
//
// DefineMarkers carry the verbatim "[DEFINE: ...]" snippets that the LLM
// emitted when it could not ground a piece of content in the user's input.
// The CLI surfaces these so the user sees exactly what is missing.
//
// Warnings are non-fatal but worth surfacing (e.g. unbalanced code fences,
// missing frontmatter when one was expected). Fatal indicates the output
// is unusable as-is and the caller should not write it to disk.
type ValidationResult struct {
	DefineMarkers []string
	Warnings      []string
	Fatal         bool
}

var (
	defineMarkerRE = regexp.MustCompile(`\[DEFINE(?::[^\]\n]+)?\]`)
	frontmatterRE  = regexp.MustCompile(`(?s)^---\s*\n.*?\n---\s*\n`)
)

// ValidateOutput inspects the LLM-produced content for structural issues.
// mode is the GenerationRequest.Mode the content came from; fileName is the
// output filename (used to infer per-file expectations, e.g. SKILL.md
// implies frontmatter must be present).
//
// The function never errors — every issue is captured in the returned
// ValidationResult so the caller can show all of them at once.
func ValidateOutput(content, mode, fileName string) ValidationResult {
	result := ValidationResult{}

	if strings.TrimSpace(content) == "" {
		result.Fatal = true
		result.Warnings = append(result.Warnings, "output is empty")
		return result
	}

	// 1. [DEFINE] markers — list verbatim so the user knows exactly what to fill in.
	for _, m := range defineMarkerRE.FindAllString(content, -1) {
		result.DefineMarkers = append(result.DefineMarkers, m)
	}

	// 2. Code fence balance — count triple-backtick line starts. An odd
	//    count indicates an unclosed code block.
	openings := strings.Count(content, "\n```") + boolToInt(strings.HasPrefix(content, "```"))
	if openings%2 != 0 {
		result.Warnings = append(result.Warnings, "unbalanced code fences (odd number of ``` markers)")
	}

	// 3. Frontmatter expected for SKILL.md and workflow .md files.
	expectsFrontmatter := strings.EqualFold(fileName, "SKILL.md") ||
		mode == "skills" || mode == "workflow-skills" || mode == "workflows"
	if expectsFrontmatter && !frontmatterRE.MatchString(content) {
		result.Warnings = append(result.Warnings, "expected YAML frontmatter delimited by --- at the start of the file")
	}

	// 4. workflow-skills must declare disable-model-invocation + allowed-tools
	//    inside the frontmatter for Claude to honor the constraints. Search
	//    only the frontmatter region to avoid false positives in body prose.
	if mode == "workflow-skills" {
		if fm := frontmatterRE.FindString(content); fm != "" {
			if !strings.Contains(fm, "disable-model-invocation:") {
				result.Warnings = append(result.Warnings, "workflow-skill frontmatter missing disable-model-invocation field")
			}
			if !strings.Contains(fm, "allowed-tools:") {
				result.Warnings = append(result.Warnings, "workflow-skill frontmatter missing allowed-tools field")
			}
		}
	}

	// 5. Truncation heuristic: a generated body shorter than 200 chars almost
	//    certainly indicates the model returned an apology or a stub.
	body := strings.TrimSpace(content)
	if len(body) < 200 && !expectsFrontmatter {
		result.Warnings = append(result.Warnings, "output suspiciously short (< 200 chars) — possible truncation or stub")
	}

	return result
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// emitValidationFeedback writes a short summary of validation findings to
// the writer. It is a no-op for nil writers and for empty results, so the
// providers can call it unconditionally.
func emitValidationFeedback(out io.Writer, fileName string, r ValidationResult) {
	if out == nil {
		return
	}
	if len(r.DefineMarkers) == 0 && len(r.Warnings) == 0 {
		return
	}
	if len(r.DefineMarkers) > 0 {
		fmt.Fprintf(out, "    %s has %d [DEFINE] marker(s) the user must resolve\n", fileName, len(r.DefineMarkers))
	}
	for _, w := range r.Warnings {
		fmt.Fprintf(out, "    %s warning: %s\n", fileName, w)
	}
}
