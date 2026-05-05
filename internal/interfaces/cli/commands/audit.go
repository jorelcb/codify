package commands

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	domain "github.com/jorelcb/codify/internal/domain/audit"
	servicedomain "github.com/jorelcb/codify/internal/domain/service"
	infraaudit "github.com/jorelcb/codify/internal/infrastructure/audit"
	"github.com/jorelcb/codify/internal/infrastructure/llm"
)

// NewAuditCmd construye `codify audit` — auditoría de commits contra
// convenciones documentadas (Conventional Commits, branches protegidas,
// y opcionalmente alineamiento con AGENTS.md vía LLM).
func NewAuditCmd() *cobra.Command {
	var (
		since      string
		limit      int
		strict     bool
		withLLM    bool
		rulesOnly  bool
		jsonOut    bool
		model      string
		noTracking bool
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

  --with-llm: heuristic, opt-in, marked clearly. Sends commits + AGENTS.md
    to an LLM and asks it to flag alignment issues with documented project
    conventions. Findings are tagged "(heuristic)" and use the dedicated kind
    'agents_alignment_issue'. Records LLM usage in .codify/usage.json unless
    --no-tracking is set. Requires ANTHROPIC_API_KEY or GEMINI_API_KEY.

Exit codes:
  - 0: clean (no findings)
  - 1: at least one significant finding (or any finding with --strict)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(since, limit, strict, withLLM, rulesOnly, jsonOut, model, noTracking)
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "Git ref to start from (default: last N commits, see --limit)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of commits to analyze when --since is empty")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat any finding (including minor) as a failure")
	cmd.Flags().BoolVar(&withLLM, "with-llm", false, "Enable LLM heuristic mode (records usage; requires API key)")
	cmd.Flags().BoolVar(&rulesOnly, "rules-only", false, "Force rules-only mode (default; mutually exclusive with --with-llm)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit findings as JSON instead of human-readable text")
	cmd.Flags().StringVarP(&model, "model", "m", "", "LLM model for --with-llm (default: claude-sonnet-4-6 or gemini-3.1-pro-preview based on env)")
	cmd.Flags().BoolVar(&noTracking, "no-tracking", false, "Skip usage tracking for --with-llm invocation")
	return cmd
}

func runAudit(since string, limit int, strict, withLLM, rulesOnly, jsonOut bool, model string, noTracking bool) error {
	if withLLM && rulesOnly {
		return fmt.Errorf("--with-llm and --rules-only are mutually exclusive")
	}
	if noTracking {
		_ = os.Setenv("CODIFY_NO_USAGE_TRACKING", "1")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 1. Always run rules-only first — deterministic baseline.
	report, err := infraaudit.Run(infraaudit.AuditOptions{
		ProjectPath: cwd,
		Since:       since,
		Limit:       limit,
		Strict:      strict,
	})
	if err != nil {
		return fmt.Errorf("audit run: %w", err)
	}

	// 2. If --with-llm, augment with heuristic findings from the LLM.
	if withLLM {
		llmFindings, llmErr := runLLMAudit(cwd, since, limit, model, report.Findings)
		if llmErr != nil {
			// Honest fallback: report the LLM error but keep rules-only output.
			// Audit shouldn't fail outright if the LLM call hiccups.
			fmt.Fprintf(os.Stderr, "WARNING: --with-llm augmentation failed: %v\n", llmErr)
			fmt.Fprintln(os.Stderr, "         Falling back to rules-only output for this run.")
			fmt.Fprintln(os.Stderr)
		} else {
			report.Findings = append(report.Findings, llmFindings...)
		}
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

// runLLMAudit dispara la auditoría heurística. Construye el prompt,
// instancia el provider, parsea la respuesta. Retorna findings con
// Heuristic=true.
func runLLMAudit(projectPath, since string, limit int, model string, ruleFindings []domain.Finding) ([]domain.Finding, error) {
	apiKey, err := llm.ResolveAPIKey(model)
	if err != nil {
		return nil, fmt.Errorf("resolve API key: %w", err)
	}

	provider, err := llm.NewProvider(context.Background(), model, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	commits, err := infraaudit.CollectCommitsForLLM(projectPath, since, limit)
	if err != nil {
		return nil, fmt.Errorf("collect commits: %w", err)
	}
	if len(commits) == 0 {
		return nil, nil // nothing to audit
	}

	agentsContent := infraaudit.LoadAgentsContent(projectPath)
	userPrompt := infraaudit.BuildAuditUserPrompt(agentsContent, commits, ruleFindings)

	resp, err := provider.EvaluatePrompt(context.Background(), servicedomain.EvaluationRequest{
		SystemPrompt: infraaudit.LLMSystemPrompt,
		UserPrompt:   userPrompt,
		Command:      "audit",
		MaxTokens:    4000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM evaluation: %w", err)
	}

	findings, err := infraaudit.ParseLLMFindings(resp.Text)
	if err != nil {
		return nil, fmt.Errorf("parse LLM output: %w", err)
	}
	return findings, nil
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
