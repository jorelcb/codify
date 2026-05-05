package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	domain "github.com/jorelcb/codify/internal/domain/audit"
	infraaudit "github.com/jorelcb/codify/internal/infrastructure/audit"
)

// NewAuditCmd construye `codify audit` — auditoría de commits contra
// convenciones documentadas (Conventional Commits, branches protegidas).
//
// Modo default (--rules-only implícito): determinista, sin LLM. Aplica reglas
// hard-coded sobre el git log local.
//
// Modo --with-llm: opt-in, heurístico. Marcado como tal en el output. Va a
// llegar en v1.24.1 — en v1.24.0 imprime un mensaje informativo y sale.
func NewAuditCmd() *cobra.Command {
	var (
		since     string
		limit     int
		strict    bool
		withLLM   bool
		rulesOnly bool
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit recent commits against Conventional Commits and protected-branch rules",
		Long: `Audit recent commits in the local git repo against project conventions:

  Default (--rules-only): deterministic, no LLM, zero cost
    - Conventional Commits header format (type[scope][!]: subject)
    - Recognized type list (feat, fix, docs, refactor, test, chore, ...)
    - Header length ≤72 characters
    - Trivial messages rejection ("wip", "fix", "update", etc.)
    - Direct commits to protected branches (main, master, develop, production)

  --with-llm (v1.24.1+): heuristic, opt-in, marked clearly. Sends commits
    + AGENTS.md to an LLM to flag alignment issues with documented project
    conventions. NOT YET IMPLEMENTED in v1.24.0; preview message only.

Exit codes:
  - 0: clean (no findings)
  - 1: at least one significant finding (or any finding with --strict)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(since, limit, strict, withLLM, rulesOnly, jsonOut)
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "Git ref to start from (default: last N commits, see --limit)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of commits to analyze when --since is empty")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat any finding (including minor) as a failure")
	cmd.Flags().BoolVar(&withLLM, "with-llm", false, "Enable LLM heuristic mode (v1.24.1+)")
	cmd.Flags().BoolVar(&rulesOnly, "rules-only", false, "Force rules-only mode (default; mutually exclusive with --with-llm)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit findings as JSON instead of human-readable text")
	return cmd
}

func runAudit(since string, limit int, strict, withLLM, rulesOnly, jsonOut bool) error {
	if withLLM && rulesOnly {
		return fmt.Errorf("--with-llm and --rules-only are mutually exclusive")
	}

	if withLLM {
		fmt.Fprintln(os.Stderr, "NOTICE: --with-llm is planned for v1.24.1 and not yet implemented in v1.24.0.")
		fmt.Fprintln(os.Stderr, "        Falling back to rules-only audit (deterministic, no LLM, no cost).")
		fmt.Fprintln(os.Stderr)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	report, err := infraaudit.Run(infraaudit.AuditOptions{
		ProjectPath: cwd,
		Since:       since,
		Limit:       limit,
		Strict:      strict,
	})
	if err != nil {
		return fmt.Errorf("audit run: %w", err)
	}

	if jsonOut {
		emitAuditJSON(report)
	} else {
		emitAuditHuman(report)
	}

	if shouldFailAudit(report, strict) {
		os.Exit(1)
	}
	return nil
}

func shouldFailAudit(report domain.Report, strict bool) bool {
	if report.IsClean() {
		return false
	}
	if strict {
		return true
	}
	return report.HasSignificant()
}

func emitAuditHuman(report domain.Report) {
	fmt.Printf("Audited %d commits\n\n", report.CommitsAnalyzed)
	if report.IsClean() {
		fmt.Println("✓ No findings.")
		return
	}

	groups := map[domain.Kind][]domain.Finding{}
	for _, f := range report.Findings {
		groups[f.Kind] = append(groups[f.Kind], f)
	}
	order := []domain.Kind{
		domain.ProtectedBranchDirectCommit,
		domain.CommitMessageInvalidType,
		domain.CommitMessageTrivial,
		domain.CommitMessageHeaderTooLong,
		domain.AgentsAlignmentIssue,
	}
	for _, kind := range order {
		findings := groups[kind]
		if len(findings) == 0 {
			continue
		}
		sort.Slice(findings, func(i, j int) bool { return findings[i].CommitSHA < findings[j].CommitSHA })
		fmt.Printf("  [%s] (%s)\n", kind, findings[0].Severity)
		for _, f := range findings {
			suffix := ""
			if f.Heuristic {
				suffix = " (heuristic)"
			}
			short := f.CommitSHA
			if len(short) > 8 {
				short = short[:8]
			}
			fmt.Printf("    - %s%s — %s\n", short, suffix, f.Detail)
		}
		fmt.Println()
	}
}

func emitAuditJSON(report domain.Report) {
	type out struct {
		Kind      string `json:"kind"`
		Severity  string `json:"severity"`
		CommitSHA string `json:"commit_sha"`
		Path      string `json:"path,omitempty"`
		Detail    string `json:"detail"`
		Heuristic bool   `json:"heuristic,omitempty"`
	}
	type wrapper struct {
		CommitsAnalyzed int   `json:"commits_analyzed"`
		Findings        []out `json:"findings"`
	}
	items := make([]out, 0, len(report.Findings))
	for _, f := range report.Findings {
		items = append(items, out{
			Kind:      string(f.Kind),
			Severity:  string(f.Severity),
			CommitSHA: f.CommitSHA,
			Path:      f.Path,
			Detail:    f.Detail,
			Heuristic: f.Heuristic,
		})
	}
	encodeJSON(wrapper{CommitsAnalyzed: report.CommitsAnalyzed, Findings: items})
}
