package audit

import (
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/audit"
)

func TestAuditCommitMessage_ValidConventional(t *testing.T) {
	got := auditCommitMessage(commit{SHA: "abc", Header: "feat(api): add endpoint"})
	if len(got) != 0 {
		t.Errorf("valid CC message should produce no findings, got: %+v", got)
	}
}

func TestAuditCommitMessage_InvalidType(t *testing.T) {
	got := auditCommitMessage(commit{SHA: "abc", Header: "improvement(api): tweak"})
	if len(got) == 0 {
		t.Fatal("expected finding for invalid type")
	}
	found := false
	for _, f := range got {
		if f.Kind == domain.CommitMessageInvalidType {
			found = true
		}
	}
	if !found {
		t.Errorf("expected CommitMessageInvalidType, got: %+v", got)
	}
}

func TestAuditCommitMessage_HeaderTooLong(t *testing.T) {
	long := "feat(api): " + repeat("a", 100)
	got := auditCommitMessage(commit{SHA: "abc", Header: long})
	found := false
	for _, f := range got {
		if f.Kind == domain.CommitMessageHeaderTooLong {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HeaderTooLong, got: %+v", got)
	}
}

func TestAuditCommitMessage_Trivial(t *testing.T) {
	cases := []string{"wip", "fix", "update", "tmp", "asdf"}
	for _, header := range cases {
		got := auditCommitMessage(commit{SHA: "abc", Header: header})
		found := false
		for _, f := range got {
			if f.Kind == domain.CommitMessageTrivial {
				found = true
			}
		}
		if !found {
			t.Errorf("expected Trivial for %q, got: %+v", header, got)
		}
	}
}

func TestAuditCommitMessage_NotConventionalAtAll(t *testing.T) {
	got := auditCommitMessage(commit{SHA: "abc", Header: "Fix that bug we discussed"})
	found := false
	for _, f := range got {
		if f.Kind == domain.CommitMessageInvalidType {
			found = true
		}
	}
	if !found {
		t.Errorf("expected InvalidType for non-CC header, got: %+v", got)
	}
}

func TestAuditCommitMessage_BreakingChange(t *testing.T) {
	got := auditCommitMessage(commit{SHA: "abc", Header: "feat!: drop legacy support"})
	if len(got) != 0 {
		t.Errorf("breaking change marker should be valid, got: %+v", got)
	}
}

func TestIsMergeCommit(t *testing.T) {
	if isMergeCommit(commit{Parents: []string{"a", "b"}}) != true {
		t.Error("two parents = merge commit")
	}
	if isMergeCommit(commit{Parents: []string{"a"}}) != false {
		t.Error("one parent = not merge")
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
