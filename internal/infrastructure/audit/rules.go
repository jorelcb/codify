// Package audit implementa la auditoría rules-only y LLM-mode sobre commits
// del proyecto. Las reglas determinísticas viven acá; LLM-mode delega a los
// providers existentes vía un prompt construido on-the-fly.
package audit

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	domain "github.com/jorelcb/codify/internal/domain/audit"
)

// validCommitTypes son los types reconocidos por Conventional Commits 1.0.0
// más los comunes por convención (ci, build, perf, revert).
var validCommitTypes = map[string]bool{
	"feat":     true,
	"fix":      true,
	"docs":     true,
	"style":    true,
	"refactor": true,
	"perf":     true,
	"test":     true,
	"build":    true,
	"ci":       true,
	"chore":    true,
	"revert":   true,
}

// trivialMessages son patrones de mensajes que típicamente reflejan trabajo
// sin terminar o commits descuidados. La detección es exact-match (después
// de strip del type) y opcionalmente con número.
var trivialMessages = []*regexp.Regexp{
	regexp.MustCompile(`^(wip|tmp|temp|test|asdf|xx+|\.\.+)\s*$`),
	regexp.MustCompile(`^(fix|update|change|edit)$`),
	regexp.MustCompile(`^(work|stuff|things)\s*$`),
}

// protectedBranches son las branches donde commits directos (no via merge
// commit / squash de PR) son sospechosos.
var protectedBranches = map[string]bool{
	"main":       true,
	"master":     true,
	"develop":    true,
	"production": true,
}

// commitHeaderRegex captura type, scope opcional, breaking marker, subject.
//   feat(api): add endpoint
//   fix!: emergency rollback
//   chore: bump deps
var commitHeaderRegex = regexp.MustCompile(`^(\w+)(\([^)]+\))?(!)?:\s*(.+)$`)

// AuditOptions parametriza la auditoría.
type AuditOptions struct {
	ProjectPath string
	Since       string // ref git, e.g. "HEAD~20" o ""
	Limit       int    // máximo de commits a analizar (default 20 si Since vacío)
	Strict      bool   // todos los findings cuentan como fail (CI usage)
}

// Run ejecuta la auditoría rules-only sobre los commits del proyecto.
// Retorna un Report con findings + count de commits analizados.
func Run(opts AuditOptions) (domain.Report, error) {
	report := domain.Report{}

	commits, err := listCommits(opts.ProjectPath, opts.Since, opts.Limit)
	if err != nil {
		return report, fmt.Errorf("git log failed: %w", err)
	}
	report.CommitsAnalyzed = len(commits)

	currentBranch := currentBranch(opts.ProjectPath)
	branchProtected := protectedBranches[currentBranch]

	for _, c := range commits {
		report.Findings = append(report.Findings, auditCommitMessage(c)...)
		if branchProtected && !isMergeCommit(c) {
			report.Findings = append(report.Findings, domain.Finding{
				Kind:      domain.ProtectedBranchDirectCommit,
				Severity:  domain.SeverityOf(domain.ProtectedBranchDirectCommit),
				CommitSHA: c.SHA,
				Path:      currentBranch,
				Detail:    fmt.Sprintf("direct commit on protected branch %q (no merge commit detected)", currentBranch),
			})
		}
	}

	return report, nil
}

// commit captura los campos del commit que las reglas inspeccionan. Mantenido
// minimal — añadir solo lo que las reglas necesitan.
type commit struct {
	SHA      string
	Header   string
	Parents  []string
}

// auditCommitMessage aplica todas las reglas que inspeccionan el mensaje.
func auditCommitMessage(c commit) []domain.Finding {
	findings := []domain.Finding{}
	header := c.Header

	// Header demasiado largo
	if len(header) > 72 {
		findings = append(findings, domain.Finding{
			Kind:      domain.CommitMessageHeaderTooLong,
			Severity:  domain.SeverityOf(domain.CommitMessageHeaderTooLong),
			CommitSHA: c.SHA,
			Detail:    fmt.Sprintf("header is %d chars (recommend ≤72)", len(header)),
		})
	}

	// Trivial messages — antes de parsear el type, porque "fix" sin colon también es trivial
	for _, re := range trivialMessages {
		if re.MatchString(strings.TrimSpace(header)) {
			findings = append(findings, domain.Finding{
				Kind:      domain.CommitMessageTrivial,
				Severity:  domain.SeverityOf(domain.CommitMessageTrivial),
				CommitSHA: c.SHA,
				Detail:    fmt.Sprintf("message %q looks like a placeholder", strings.TrimSpace(header)),
			})
			return findings // no aporta seguir parseando un mensaje trivial
		}
	}

	// Conventional Commit type
	matches := commitHeaderRegex.FindStringSubmatch(header)
	if matches == nil {
		findings = append(findings, domain.Finding{
			Kind:      domain.CommitMessageInvalidType,
			Severity:  domain.SeverityOf(domain.CommitMessageInvalidType),
			CommitSHA: c.SHA,
			Detail:    "header does not match Conventional Commits format (type[scope][!]: subject)",
		})
		return findings
	}

	commitType := matches[1]
	if !validCommitTypes[commitType] {
		findings = append(findings, domain.Finding{
			Kind:      domain.CommitMessageInvalidType,
			Severity:  domain.SeverityOf(domain.CommitMessageInvalidType),
			CommitSHA: c.SHA,
			Detail:    fmt.Sprintf("type %q is not a recognized Conventional Commits type", commitType),
		})
	}

	return findings
}

func isMergeCommit(c commit) bool {
	return len(c.Parents) > 1
}

// listCommits ejecuta `git log --format=...` y parsea los commits.
func listCommits(projectPath, since string, limit int) ([]commit, error) {
	args := []string{"log", "--format=%H|%P|%s"}
	if since != "" {
		args = append(args, since+"..HEAD")
	} else {
		if limit <= 0 {
			limit = 20
		}
		args = append(args, fmt.Sprintf("-n%d", limit))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commits := []commit{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		c := commit{
			SHA:    parts[0],
			Header: parts[2],
		}
		if parts[1] != "" {
			c.Parents = strings.Fields(parts[1])
		}
		commits = append(commits, c)
	}
	return commits, nil
}

func currentBranch(projectPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
