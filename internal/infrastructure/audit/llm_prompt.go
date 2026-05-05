package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	domain "github.com/jorelcb/codify/internal/domain/audit"
)

// LLMPromptBuilder construye el prompt enviado al LLM en `audit --with-llm`.
// El prompt es self-contained: incluye AGENTS.md, los commits a evaluar, las
// findings rules-only ya identificadas (para que el LLM no las repita), y un
// schema JSON estricto para la respuesta.
type LLMPromptBuilder struct{}

// LLMSystemPrompt es el preamble fijo que setea el rol del LLM y restringe
// el output al schema JSON esperado. Mantenerlo corto reduce tokens y
// minimiza variabilidad de comportamiento entre modelos.
const LLMSystemPrompt = `You are a senior code reviewer auditing recent git commits against the documented project conventions in AGENTS.md.

Your task: identify commits that DO NOT align with documented conventions. Skip commits that the rules-only audit already flagged (those will be reported separately) — focus only on alignment issues that require subjective judgment about the project's stated guidelines.

Return ONLY a valid JSON array. No prose, no markdown fences, no commentary.

Schema (each finding):
{
  "commit_sha": "<full SHA>",
  "severity": "significant" | "minor",
  "detail": "<one-sentence description of the misalignment>"
}

Severity guidelines:
- "significant" — the commit clearly violates a stated MUST or MUST NOT rule in AGENTS.md
- "minor" — the commit is questionable but not a hard violation

If all commits align: return [].`

// BuildAuditUserPrompt arma el user prompt con AGENTS.md + commits + findings
// rules-only (para que el LLM no las re-flagee).
func BuildAuditUserPrompt(agentsContent string, commits []CommitInfo, ruleFindings []domain.Finding) string {
	var sb strings.Builder
	sb.WriteString("# AGENTS.md (project conventions)\n\n")
	if agentsContent == "" {
		sb.WriteString("(AGENTS.md not found in project root — judge based on common best practices)\n")
	} else {
		sb.WriteString(agentsContent)
	}
	sb.WriteString("\n\n# Recent commits\n\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("## %s\n", c.SHA))
		sb.WriteString(fmt.Sprintf("Header: %s\n", c.Header))
		if c.Body != "" {
			sb.WriteString(fmt.Sprintf("Body:\n%s\n", c.Body))
		}
		if c.Files != "" {
			sb.WriteString(fmt.Sprintf("Files changed: %s\n", c.Files))
		}
		sb.WriteString("\n")
	}

	if len(ruleFindings) > 0 {
		sb.WriteString("# Already flagged by rules-only audit (do not repeat)\n\n")
		for _, f := range ruleFindings {
			sb.WriteString(fmt.Sprintf("- %s: %s — %s\n", f.CommitSHA, f.Kind, f.Detail))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Output: JSON array of findings as specified in the system prompt.")
	return sb.String()
}

// CommitInfo es la representación expuesta al LLM. Más rica que el commit
// interno de rules.go porque incluye body + files. La construye CollectCommitsForLLM.
type CommitInfo struct {
	SHA    string
	Header string
	Body   string
	Files  string // resumen "+10 -5 path1, +2 -0 path2"
}

// CollectCommitsForLLM extrae commits del git log con el detalle adicional
// que necesita el LLM (body + stat). Trabaja sobre projectPath; respeta
// since/limit como rules-only.
func CollectCommitsForLLM(projectPath, since string, limit int) ([]CommitInfo, error) {
	args := []string{"log", "--format=%H%x1f%s%x1f%b%x1e"}
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
	commits := []CommitInfo{}
	for _, raw := range strings.Split(string(out), "\x1e") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		parts := strings.SplitN(raw, "\x1f", 3)
		if len(parts) < 2 {
			continue
		}
		ci := CommitInfo{SHA: parts[0], Header: parts[1]}
		if len(parts) >= 3 {
			ci.Body = strings.TrimSpace(parts[2])
		}
		ci.Files = collectFilesStat(projectPath, ci.SHA)
		commits = append(commits, ci)
	}
	return commits, nil
}

// collectFilesStat devuelve un resumen corto de los archivos cambiados por
// el commit. No bloquea audit si falla — devuelve "" silenciosamente.
func collectFilesStat(projectPath, sha string) string {
	cmd := exec.Command("git", "show", "--stat", "--format=", sha)
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// Solo la última línea suele tener el summary "N files changed, M insertions, K deletions"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[len(lines)-1])
}

// LoadAgentsContent lee AGENTS.md desde el cwd o el output dir más probable.
// Si no encuentra, devuelve "" (LLM operará con AGENTS.md "no provided" y
// usará common best practices).
func LoadAgentsContent(projectPath string) string {
	candidates := []string{
		filepath.Join(projectPath, "AGENTS.md"),
		filepath.Join(projectPath, "output", "AGENTS.md"),
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data)
		}
	}
	return ""
}

// llmFindingJSON es el schema que esperamos del LLM. Coincide 1:1 con el
// system prompt; cualquier cambio acá requiere actualizar el prompt.
type llmFindingJSON struct {
	CommitSHA string `json:"commit_sha"`
	Severity  string `json:"severity"`
	Detail    string `json:"detail"`
}

// fenceRegex remueve markdown fences que algunos LLMs agregan a pesar de las
// instrucciones del system prompt. Best-effort cleanup antes del unmarshal.
var fenceRegex = regexp.MustCompile("^\\s*```(?:json)?\\s*\\n?|\\n?\\s*```\\s*$")

// ParseLLMFindings convierte la respuesta cruda del LLM en domain.Finding[].
// Marca todas las findings con Heuristic=true. Si el JSON no parsea, devuelve
// error para que el caller decida si abortar o caer al rules-only.
func ParseLLMFindings(raw string) ([]domain.Finding, error) {
	cleaned := fenceRegex.ReplaceAllString(strings.TrimSpace(raw), "")

	var items []llmFindingJSON
	if err := json.Unmarshal([]byte(cleaned), &items); err != nil {
		return nil, fmt.Errorf("parse LLM JSON: %w (raw: %q)", err, truncate(raw, 200))
	}

	findings := make([]domain.Finding, 0, len(items))
	for _, it := range items {
		if it.CommitSHA == "" {
			continue
		}
		sev := domain.Severity(it.Severity)
		if sev != domain.Significant && sev != domain.Minor {
			sev = domain.Minor // default conservador para output ambiguo
		}
		findings = append(findings, domain.Finding{
			Kind:      domain.AgentsAlignmentIssue,
			Severity:  sev,
			CommitSHA: it.CommitSHA,
			Detail:    it.Detail,
			Heuristic: true,
		})
	}
	return findings, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
