package audit

import (
	"strings"
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/audit"
)

func TestParseLLMFindings_HappyPath(t *testing.T) {
	raw := `[
		{"commit_sha": "abc123", "severity": "significant", "detail": "violates DDD layer rule"},
		{"commit_sha": "def456", "severity": "minor", "detail": "naming inconsistency"}
	]`
	findings, err := ParseLLMFindings(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(findings) != 2 {
		t.Errorf("got %d findings, want 2", len(findings))
	}
	if findings[0].Kind != domain.AgentsAlignmentIssue {
		t.Errorf("got kind %q, want AgentsAlignmentIssue", findings[0].Kind)
	}
	if !findings[0].Heuristic {
		t.Errorf("heuristic flag should be true for LLM findings")
	}
	if findings[0].Severity != domain.Significant {
		t.Errorf("got severity %q, want significant", findings[0].Severity)
	}
}

func TestParseLLMFindings_EmptyArray(t *testing.T) {
	findings, err := ParseLLMFindings("[]")
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0", len(findings))
	}
}

func TestParseLLMFindings_StripsMarkdownFences(t *testing.T) {
	raw := "```json\n[{\"commit_sha\":\"a\",\"severity\":\"minor\",\"detail\":\"x\"}]\n```"
	findings, err := ParseLLMFindings(raw)
	if err != nil {
		t.Fatalf("parse fenced: %v", err)
	}
	if len(findings) != 1 || findings[0].CommitSHA != "a" {
		t.Errorf("got %+v", findings)
	}
}

func TestParseLLMFindings_InvalidSeverityFallsBackToMinor(t *testing.T) {
	raw := `[{"commit_sha":"abc","severity":"critical","detail":"nope"}]`
	findings, err := ParseLLMFindings(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if findings[0].Severity != domain.Minor {
		t.Errorf("invalid severity should fall back to minor, got %q", findings[0].Severity)
	}
}

func TestParseLLMFindings_SkipsEntriesWithoutSHA(t *testing.T) {
	raw := `[
		{"commit_sha":"abc","severity":"minor","detail":"ok"},
		{"commit_sha":"","severity":"minor","detail":"missing sha"}
	]`
	findings, err := ParseLLMFindings(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 valid finding, got %d", len(findings))
	}
}

func TestParseLLMFindings_InvalidJSONReturnsError(t *testing.T) {
	_, err := ParseLLMFindings("not json at all")
	if err == nil {
		t.Error("expected error for non-JSON input")
	}
}

func TestBuildAuditUserPrompt_IncludesAgentsAndCommits(t *testing.T) {
	commits := []CommitInfo{
		{SHA: "abc", Header: "feat: x", Body: "details"},
		{SHA: "def", Header: "fix: y"},
	}
	prompt := BuildAuditUserPrompt("# Project rules\nMUST do X", commits, nil)
	if !strings.Contains(prompt, "MUST do X") {
		t.Error("prompt should contain AGENTS.md content")
	}
	if !strings.Contains(prompt, "abc") || !strings.Contains(prompt, "def") {
		t.Error("prompt should list all commits")
	}
}

func TestBuildAuditUserPrompt_NotesAlreadyFlagged(t *testing.T) {
	prior := []domain.Finding{
		{CommitSHA: "abc", Kind: domain.CommitMessageTrivial, Detail: "wip placeholder"},
	}
	prompt := BuildAuditUserPrompt("# rules", []CommitInfo{{SHA: "abc", Header: "wip"}}, prior)
	if !strings.Contains(prompt, "Already flagged") {
		t.Error("prompt should list rule findings to avoid double-counting")
	}
	if !strings.Contains(prompt, "wip placeholder") {
		t.Error("prompt should include the rule finding's detail")
	}
}
