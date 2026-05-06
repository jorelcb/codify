package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// defineMarkerLineRE matches the same [DEFINE]/[DEFINE: hint] form the LLM
// validator captures. Kept here to avoid a cross-package dep on the llm
// internals — the regex is short and stable.
var defineMarkerLineRE = regexp.MustCompile(`\[DEFINE(?::[^\]\n]+)?\]`)

// markerHit captures a single [DEFINE] occurrence inside a file: its verbatim
// text, 1-based line number, and the user's answer (or empty if skipped).
type markerHit struct {
	Text   string // verbatim "[DEFINE: hint]" or "[DEFINE]"
	Line   int    // 1-based
	Answer string // empty = user chose to skip
}

// resolveDefineMarkers walks the generated files, collects [DEFINE] markers,
// asks the user to fill each in, and rewrites the files with the answers.
//
// When provider is non-nil, the rewrite is delegated to the LLM so the user's
// answer is integrated naturally into the surrounding prose (path B). When
// provider is nil — typically because no API key is configured — we fall
// back to a literal 1:1 substitution of the marker text with the answer
// (path A). Both paths preserve unanswered markers verbatim.
//
// Returns nil if there's nothing to do (no markers anywhere) or if the user
// declines the resolve flow at the top-level prompt.
func resolveDefineMarkers(ctx context.Context, files []string, locale string, provider service.LLMProvider) error {
	if !isInteractive() {
		return nil
	}

	// 1. Scan all files; build per-file marker list.
	type fileMarkers struct {
		path    string
		content string
		hits    []markerHit
	}
	var withMarkers []fileMarkers
	totalMarkers := 0
	for _, p := range files {
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		hits := scanMarkers(string(content))
		if len(hits) == 0 {
			continue
		}
		withMarkers = append(withMarkers, fileMarkers{path: p, content: string(content), hits: hits})
		totalMarkers += len(hits)
	}
	if totalMarkers == 0 {
		return nil
	}

	// 2. Top-level opt-in. Empty input = skip; respects the user who would
	//    rather edit the files in their editor.
	fmt.Println()
	fmt.Printf("Found %d [DEFINE] marker(s) across %d file(s).\n", totalMarkers, len(withMarkers))
	proceed, err := promptConfirm("Resolve them interactively now?", true)
	if err != nil || !proceed {
		fmt.Println("Skipped. Markers remain in the files for manual editing.")
		return nil
	}

	// 3. Per-file: prompt every marker, collect answers, then rewrite.
	for i := range withMarkers {
		fm := &withMarkers[i]
		fmt.Println()
		fmt.Printf("── %s (%d marker%s) ──\n", fm.path, len(fm.hits), pluralS(len(fm.hits)))

		for j := range fm.hits {
			hit := &fm.hits[j]
			showMarkerContext(fm.content, hit)
			ans, err := promptInput(fmt.Sprintf("Your input for L%d (Enter to skip)", hit.Line), "")
			if err != nil {
				return err
			}
			hit.Answer = strings.TrimSpace(ans)
		}

		// Skip files where the user answered nothing.
		answered := 0
		for _, h := range fm.hits {
			if h.Answer != "" {
				answered++
			}
		}
		if answered == 0 {
			fmt.Printf("  (no answers — file unchanged)\n")
			continue
		}

		newContent := fm.content
		usedLLM := false
		if provider != nil {
			rewritten, err := rewriteWithLLM(ctx, provider, fm.path, fm.content, fm.hits, locale)
			if err == nil && rewritten != "" {
				newContent = rewritten
				usedLLM = true
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "  LLM rewrite failed (%v); falling back to literal substitution\n", err)
			}
		}
		if !usedLLM {
			newContent = literalSubstitute(fm.content, fm.hits)
		}

		if err := os.WriteFile(fm.path, []byte(newContent), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "  write failed: %v\n", err)
			continue
		}
		mode := "literal substitution"
		if usedLLM {
			mode = "LLM rewrite"
		}
		fmt.Printf("  ✓ %d marker(s) resolved via %s\n", answered, mode)
	}
	return nil
}

// scanMarkers finds every [DEFINE]/[DEFINE: hint] occurrence in content.
func scanMarkers(content string) []markerHit {
	var hits []markerHit
	for _, idx := range defineMarkerLineRE.FindAllStringIndex(content, -1) {
		hits = append(hits, markerHit{
			Text: content[idx[0]:idx[1]],
			Line: strings.Count(content[:idx[0]], "\n") + 1,
		})
	}
	return hits
}

// showMarkerContext prints the marker line plus a few lines of surrounding
// context so the user can decide what to fill in without alt-tabbing to an
// editor.
func showMarkerContext(content string, hit *markerHit) {
	const radius = 2
	lines := strings.Split(content, "\n")
	from := hit.Line - 1 - radius
	if from < 0 {
		from = 0
	}
	to := hit.Line - 1 + radius
	if to >= len(lines) {
		to = len(lines) - 1
	}
	fmt.Println()
	for i := from; i <= to; i++ {
		marker := "  "
		if i == hit.Line-1 {
			marker = "▸ "
		}
		fmt.Printf("    %s%4d  %s\n", marker, i+1, lines[i])
	}
}

// literalSubstitute replaces each answered marker with the answer text 1:1.
// Skipped markers are preserved verbatim.
func literalSubstitute(content string, hits []markerHit) string {
	for _, h := range hits {
		if h.Answer == "" {
			continue
		}
		content = strings.Replace(content, h.Text, h.Answer, 1)
	}
	return content
}

// rewriteWithLLM asks the LLM to rewrite the affected paragraphs so the
// user's answers integrate naturally with the surrounding prose. Returns
// the rewritten file content (or empty + error on failure).
func rewriteWithLLM(ctx context.Context, provider service.LLMProvider, fileName, content string, hits []markerHit, locale string) (string, error) {
	var answers strings.Builder
	for _, h := range hits {
		if h.Answer == "" {
			continue
		}
		fmt.Fprintf(&answers, "  %s → %s\n", h.Text, h.Answer)
	}
	if answers.Len() == 0 {
		return "", nil
	}

	systemPrompt := strings.TrimSpace(`
You are a precise technical editor. The user has provided answers for
[DEFINE: ...] placeholders embedded in a generated context file. Rewrite the
file by integrating each answer naturally into its surrounding sentence or
paragraph. Preserve ALL other content character-for-character: headings,
lists, code blocks, frontmatter, indentation, blank lines.

Rules:
- Replace each [DEFINE: ...] occurrence so the resulting prose reads as if
  the answer were always there. Adjust grammar minimally to keep the flow.
- Do NOT invent additional content beyond what the user provided.
- If a marker is NOT in the answers map, leave it verbatim — the user chose
  to skip it.
- Output ONLY the full rewritten file content, no preamble, no commentary,
  no markdown fence wrapping.
`)

	userPrompt := fmt.Sprintf("FILE: %s\nLOCALE: %s\n\nANSWERS (placeholder → user input):\n%s\n--- BEGIN FILE ---\n%s\n--- END FILE ---\n",
		fileName, locale, answers.String(), content)

	resp, err := provider.EvaluatePrompt(ctx, service.EvaluationRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Command:      "resolve-defines",
		MaxTokens:    8192,
	})
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(resp.Text)
	// Defensive unwrap: some models still wrap in ```...``` despite the rule.
	if strings.HasPrefix(out, "```") {
		out = strings.TrimPrefix(out, "```markdown")
		out = strings.TrimPrefix(out, "```md")
		out = strings.TrimPrefix(out, "```")
		out = strings.TrimSuffix(out, "```")
		out = strings.TrimSpace(out)
	}
	return out, nil
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
