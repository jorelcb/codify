package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// ResolveMarkersCommand walks a set of generated files, asks the user to fill
// in any [DEFINE: ...] markers, and rewrites each file with the answers.
//
// Two paths integrate the answers back into the file:
//   - LLM rewrite (preferred when a provider is wired): the file content +
//     marker→answer map is sent to the LLM, which integrates the answers
//     naturally into surrounding prose.
//   - Literal substitution (fallback): each marker text is replaced verbatim
//     with the user's answer. Used when no provider is configured or when
//     the LLM call fails. Less polished, never loses the user's work.
//
// The command is invoked from three places that share the same flow today:
//   - Post-generation hook in `codify generate` / `analyze` / `init`.
//   - The standalone `codify resolve` command (Phase 5).
//   - Future MCP tool / watch loop integrations.
//
// Hence the command lives in application/command — the CLI layer becomes a
// thin adapter that wires the prompter, provider, and file IO.
type ResolveMarkersCommand struct {
	prompter  service.InteractivePrompter
	provider  service.LLMProvider // optional — nil means literal-only mode
	readFile  func(string) ([]byte, error)
	writeFile func(string, []byte, os.FileMode) error
	stderr    func(format string, args ...any)
}

// ResolveRequest carries everything the command needs to run a resolve pass
// over a set of files. Locale flows into the LLM rewrite prompt; the
// orchestrator itself is locale-agnostic.
type ResolveRequest struct {
	Files  []string
	Locale string
}

// ResolveResult summarizes the outcome. Useful for tests and for the future
// `codify resolve --json` reporting flag.
type ResolveResult struct {
	TotalMarkers   int
	FilesScanned   int
	FilesRewritten int
	FilesUnchanged int // user skipped every marker in these
	Resolved       int // markers actually replaced (sum across files)
	Skipped        int // markers preserved verbatim
	UsedLLM        int // files rewritten via LLM path
	UsedLiteral    int // files rewritten via literal path
	Declined       bool
}

// NewResolveMarkersCommand wires the dependencies. The provider is optional
// — pass nil to force literal-substitution mode (used by tests and by
// pre-existing files where no API key is available).
func NewResolveMarkersCommand(
	prompter service.InteractivePrompter,
	provider service.LLMProvider,
) *ResolveMarkersCommand {
	return &ResolveMarkersCommand{
		prompter: prompter,
		provider: provider,
		readFile: os.ReadFile,
		writeFile: func(path string, data []byte, perm os.FileMode) error {
			return os.WriteFile(path, data, perm)
		},
		stderr: func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format, args...)
		},
	}
}

// WithFileIO replaces the default file IO. Tests use it to redirect reads
// and writes to in-memory maps.
func (c *ResolveMarkersCommand) WithFileIO(
	read func(string) ([]byte, error),
	write func(string, []byte, os.FileMode) error,
) *ResolveMarkersCommand {
	c.readFile = read
	c.writeFile = write
	return c
}

// WithStderr replaces the default stderr writer. Tests capture it.
func (c *ResolveMarkersCommand) WithStderr(stderr func(format string, args ...any)) *ResolveMarkersCommand {
	c.stderr = stderr
	return c
}

// Execute runs the full resolve pass. Returns a result summary on success
// (including when the user declines the top-level prompt — Declined=true,
// no error). Errors are reserved for IO/transport failures that prevent the
// command from completing.
func (c *ResolveMarkersCommand) Execute(ctx context.Context, req ResolveRequest) (*ResolveResult, error) {
	result := &ResolveResult{}

	type fileMarkers struct {
		path    string
		content string
		hits    []service.MarkerHit
	}
	var withMarkers []fileMarkers

	for _, p := range req.Files {
		data, err := c.readFile(p)
		if err != nil {
			continue
		}
		hits := service.ScanMarkers(string(data))
		if len(hits) == 0 {
			continue
		}
		withMarkers = append(withMarkers, fileMarkers{path: p, content: string(data), hits: hits})
		result.TotalMarkers += len(hits)
		result.FilesScanned++
	}

	if result.TotalMarkers == 0 {
		return result, nil
	}

	proceed, err := c.prompter.ConfirmTopLevel(result.TotalMarkers, len(withMarkers))
	if err != nil || !proceed {
		result.Declined = true
		return result, nil
	}

	for i := range withMarkers {
		fm := &withMarkers[i]
		c.prompter.AnnounceFile(fm.path, len(fm.hits))

		for j := range fm.hits {
			hit := &fm.hits[j]
			ans, err := c.prompter.AskMarker(fm.content, service.EnrichedMarker{MarkerHit: *hit})
			if err != nil {
				return result, fmt.Errorf("prompt for %s line %d: %w", fm.path, hit.Line, err)
			}
			if ans.Skip {
				result.Skipped++
				continue
			}
			trimmed := strings.TrimSpace(ans.Answer)
			if trimmed == "" {
				result.Skipped++
				continue
			}
			hit.Answer = trimmed
			result.Resolved++
		}

		answeredInFile := 0
		for _, h := range fm.hits {
			if h.Answer != "" {
				answeredInFile++
			}
		}
		if answeredInFile == 0 {
			result.FilesUnchanged++
			c.prompter.ReportFileResult(fm.path, 0, "unchanged")
			continue
		}

		newContent, mode, err := c.rewriteFile(ctx, fm.path, fm.content, fm.hits, req.Locale)
		if err != nil {
			c.stderr("  write skipped for %s: %v\n", fm.path, err)
			continue
		}

		if err := c.writeFile(fm.path, []byte(newContent), 0o644); err != nil {
			c.stderr("  write failed for %s: %v\n", fm.path, err)
			continue
		}
		result.FilesRewritten++
		switch mode {
		case "llm":
			result.UsedLLM++
		case "literal":
			result.UsedLiteral++
		}
		c.prompter.ReportFileResult(fm.path, answeredInFile, mode)
	}

	return result, nil
}

// rewriteFile chooses the rewrite path (LLM vs literal) and returns the new
// content, the mode tag for reporting, and any error that prevented BOTH
// paths from succeeding. A successful literal fallback is NOT an error here
// — only a complete failure (e.g. provider broken AND literal logic
// crashed, which is unreachable today but kept defensive).
func (c *ResolveMarkersCommand) rewriteFile(
	ctx context.Context,
	path, content string,
	hits []service.MarkerHit,
	locale string,
) (string, string, error) {
	if c.provider != nil {
		rewritten, err := rewriteWithLLM(ctx, c.provider, path, content, hits, locale)
		if err == nil && rewritten != "" {
			return rewritten, "llm", nil
		}
		if err != nil {
			c.stderr("  LLM rewrite failed for %s (%v); falling back to literal substitution\n", path, err)
		}
	}
	return service.LiteralSubstitute(content, hits), "literal", nil
}

// rewriteWithLLM sends the file + answers to the configured provider so the
// answers integrate naturally into surrounding prose. Returns ("", nil) when
// nothing was answered — the caller treats that as a no-op file.
func rewriteWithLLM(
	ctx context.Context,
	provider service.LLMProvider,
	fileName, content string,
	hits []service.MarkerHit,
	locale string,
) (string, error) {
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

	userPrompt := fmt.Sprintf(
		"FILE: %s\nLOCALE: %s\n\nANSWERS (placeholder → user input):\n%s\n--- BEGIN FILE ---\n%s\n--- END FILE ---\n",
		fileName, locale, answers.String(), content,
	)

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
