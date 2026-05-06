package commands

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jorelcb/codify/internal/application/command"
	"github.com/jorelcb/codify/internal/domain/service"
	infraresolver "github.com/jorelcb/codify/internal/infrastructure/resolver"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
)

// NewResolveCmd builds `codify resolve` — interactive marker resolution
// over user-supplied files (or auto-discovered via --all / --since). Reuses
// the same ResolveMarkersCommand that the post-generate hook uses.
func NewResolveCmd() *cobra.Command {
	var (
		allFiles   bool
		since      string
		noEnrich   bool
		noPreview  bool
		skipModeS  string
		dryRun     bool
		locale     string
		modelFlag  string
	)

	cmd := &cobra.Command{
		Use:   "resolve [files...]",
		Short: "Interactively resolve [DEFINE: ...] markers in existing files",
		Long: `Walk one or more files and ask the user to fill in any [DEFINE: ...]
markers the LLM emitted (or that survive from previous generations).

By default the resolver:
  - asks the configured LLM to translate each marker into a natural
    question with grounded suggestions (--no-enrich opts out)
  - shows a diff preview before writing each file (--no-preview opts out)
  - replaces skipped markers with date-stamped TODO comments in the
    file's native syntax (--skip-mode=verbatim opts out)

File selection (mutually exclusive):
  codify resolve <files...>           explicit list
  codify resolve --all                walk cwd, pick every file containing a marker
  codify resolve --since=<ref>        files changed in git since <ref>

Examples:
  codify resolve AGENTS.md CONTEXT.md
  codify resolve --all
  codify resolve --since=HEAD~5
  codify resolve --all --no-enrich --skip-mode=verbatim
  codify resolve --all --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skipMode, err := parseSkipMode(skipModeS)
			if err != nil {
				return err
			}
			files, err := discoverFiles(args, allFiles, since)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Println("No files to resolve.")
				return nil
			}

			provider, _ := buildResolveProvider(modelFlag)

			rcmd := command.NewResolveMarkersCommand(NewHuhPrompter(), provider)
			if !noEnrich && provider != nil {
				rcmd = rcmd.WithEnricher(infraresolver.NewLLMEnricher(provider))
			}
			if !noPreview {
				rcmd = rcmd.WithPreviewer(NewHuhDiffPreviewer())
			}
			if dryRun {
				rcmd = rcmd.WithFileIO(os.ReadFile, dryRunWriteFile)
			}

			result, err := rcmd.Execute(context.Background(), command.ResolveRequest{
				Files:    files,
				Locale:   locale,
				SkipMode: skipMode,
			})
			if err != nil {
				return err
			}
			printResolveSummary(result, dryRun)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFiles, "all", "a", false, "Walk cwd and resolve every file that contains a [DEFINE] marker")
	cmd.Flags().StringVar(&since, "since", "", "Only resolve files changed in git since this ref (e.g. HEAD~5)")
	cmd.Flags().BoolVar(&noEnrich, "no-enrich", false, "Skip the LLM-driven question/suggestions step (cheaper, less friendly)")
	cmd.Flags().BoolVar(&noPreview, "no-preview", false, "Skip the diff preview before writing files")
	cmd.Flags().StringVar(&skipModeS, "skip-mode", "todo", "How to handle skipped markers: 'todo' (TODO comment in file syntax) or 'verbatim'")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Walk markers and report what would change without writing files")
	cmd.Flags().StringVar(&locale, "locale", "en", "Output locale used by the LLM rewrite/enrichment prompts")
	cmd.Flags().StringVarP(&modelFlag, "model", "m", "", "LLM model (default: claude-sonnet-4-6 or gemini-3.1-pro-preview based on env)")

	return cmd
}

func parseSkipMode(s string) (service.SkipMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "todo":
		return service.SkipModeTODO, nil
	case "verbatim":
		return service.SkipModeVerbatim, nil
	default:
		return 0, fmt.Errorf("invalid --skip-mode %q: use 'todo' or 'verbatim'", s)
	}
}

// discoverFiles resolves the input selectors to an absolute list of files.
// Multiple selectors short-circuit: explicit args win, then --all, then
// --since. At most one selector should be provided; the function does not
// enforce mutual exclusion strictly, it just picks deterministically.
func discoverFiles(args []string, allFiles bool, since string) ([]string, error) {
	switch {
	case len(args) > 0:
		return args, nil
	case allFiles:
		return walkForMarkers(".")
	case since != "":
		return gitDiffNames(since)
	default:
		return nil, fmt.Errorf("provide files explicitly, or use --all / --since=<ref>")
	}
}

// walkForMarkers walks root recursively and returns paths whose content
// contains at least one [DEFINE] / [DEFINE: hint] marker. Skips .git, node_modules,
// vendor, and binary files (basic heuristic: contains a NUL byte in first 4KB).
func walkForMarkers(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".codify" {
				return fs.SkipDir
			}
			return nil
		}
		// Cheap binary check.
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		head := data
		if len(head) > 4096 {
			head = head[:4096]
		}
		if bytes.IndexByte(head, 0) != -1 {
			return nil
		}
		if len(service.ScanMarkers(string(data))) > 0 {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	return out, nil
}

// gitDiffNames returns the paths reported by `git diff --name-only <since>...HEAD`,
// filtered to the subset that still exists on disk and contains markers.
func gitDiffNames(since string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", since+"...HEAD")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return nil, fmt.Errorf("git diff: %w", err)
		}
		return nil, fmt.Errorf("git diff failed: %s", firstLine(msg))
	}
	var matched []string
	for _, line := range strings.Split(string(out), "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue // missing/renamed/deleted — skip
		}
		if len(service.ScanMarkers(string(data))) > 0 {
			matched = append(matched, path)
		}
	}
	return matched, nil
}

// firstLine returns the first non-empty line of s, trimmed.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

// dryRunWriteFile is a no-op write that reports the path it would have
// written to. The caller's stdout summary aggregates by file count.
func dryRunWriteFile(path string, _ []byte, _ os.FileMode) error {
	fmt.Printf("  (dry-run) would write %s\n", path)
	return nil
}

// buildResolveProvider returns the active LLM provider, or nil when no API
// key is available. The caller decides whether to require it (no-enrich
// path is fine without; LLM rewrite path silently falls back to literal).
func buildResolveProvider(model string) (service.LLMProvider, error) {
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		// No key available — the resolver gracefully degrades to literal-only
		// substitution for skipped markers and the legacy UI for prompts.
		return nil, nil
	}
	return llm.NewProvider(context.Background(), model, apiKey, nil)
}

// printResolveSummary renders a one-paragraph summary of the resolve pass.
// dry-run callers see the same numbers; the per-file dry-run lines come from
// dryRunWriteFile so the summary stays uniform.
func printResolveSummary(result *command.ResolveResult, dryRun bool) {
	fmt.Println()
	if result.Declined {
		fmt.Println("Resolve cancelled at top-level prompt.")
		return
	}
	if result.TotalMarkers == 0 {
		fmt.Println("No [DEFINE] markers found in selected files.")
		return
	}
	verb := "Resolved"
	if dryRun {
		verb = "Would resolve"
	}
	fmt.Printf(
		"%s %d marker(s) across %d file(s): %d resolved, %d skipped, %d unchanged, %d discarded. (LLM=%d, literal=%d)\n",
		verb,
		result.TotalMarkers,
		result.FilesScanned,
		result.Resolved,
		result.Skipped,
		result.FilesUnchanged,
		result.FilesDiscarded,
		result.UsedLLM,
		result.UsedLiteral,
	)
}
