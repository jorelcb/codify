package llm

import "testing"

func TestValidateOutput_DetectsDefineMarkers(t *testing.T) {
	body := `# Title

The currency is [DEFINE: ISO 4217 code], the timezone is [DEFINE].
Padding text to keep the body above the truncation threshold so we do not
trigger the length-based warning together with the marker assertion below.
`
	r := ValidateOutput(body, "generate", "AGENTS.md")
	if len(r.DefineMarkers) != 2 {
		t.Fatalf("DefineMarkers: got %d, want 2 (got: %v)", len(r.DefineMarkers), r.DefineMarkers)
	}
	if r.Fatal {
		t.Fatal("Fatal should be false; markers are a soft signal")
	}
}

func TestValidateOutput_FlagsUnbalancedFences(t *testing.T) {
	body := "# Title\n\n```go\nfmt.Println(\"hi\")\nbody is long enough to avoid the truncation warning so the only thing flagged is the unclosed code fence in this fixture, kept short on purpose."
	r := ValidateOutput(body, "generate", "DEVELOPMENT_GUIDE.md")
	found := false
	for _, w := range r.Warnings {
		if contains(w, "unbalanced code fences") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unbalanced fence warning, got %v", r.Warnings)
	}
}

func TestValidateOutput_RequiresFrontmatterForSkill(t *testing.T) {
	body := `# Skill body

No frontmatter here, but enough text to exceed the truncation threshold so the
only warning we want surfaced is about the missing YAML frontmatter that every
SKILL.md must declare at the top of the file before any markdown content.
`
	r := ValidateOutput(body, "skills", "SKILL.md")
	found := false
	for _, w := range r.Warnings {
		if contains(w, "expected YAML frontmatter") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected frontmatter warning, got %v", r.Warnings)
	}
}

func TestValidateOutput_WorkflowSkillRequiresAllowedTools(t *testing.T) {
	body := `---
name: ship
description: ship a release
---

# Ship

Workflow body, kept long enough to avoid the truncation warning so the only
issues surfaced relate to the missing required frontmatter fields like the
disable-model-invocation flag and the allowed-tools whitelist.
`
	r := ValidateOutput(body, "workflow-skills", "SKILL.md")
	missingFlag := false
	missingTools := false
	for _, w := range r.Warnings {
		if contains(w, "disable-model-invocation") {
			missingFlag = true
		}
		if contains(w, "allowed-tools") {
			missingTools = true
		}
	}
	if !missingFlag || !missingTools {
		t.Fatalf("expected warnings for both fields, got %v", r.Warnings)
	}
}

func TestValidateOutput_AcceptsValidWorkflowSkill(t *testing.T) {
	body := `---
name: ship
description: ship a release
disable-model-invocation: true
allowed-tools: Bash(git *) Bash(go *)
---

# Ship release

1. Run tests
2. Tag version
3. Push to origin
4. Open a release PR
5. Merge once approved
6. Verify deployment succeeded with smoke checks
7. Announce in the release channel after verification
`
	r := ValidateOutput(body, "workflow-skills", "SKILL.md")
	if len(r.Warnings) != 0 {
		t.Fatalf("valid workflow-skill should produce no warnings, got %v", r.Warnings)
	}
}

func TestValidateOutput_EmptyIsFatal(t *testing.T) {
	r := ValidateOutput("   \n  ", "generate", "AGENTS.md")
	if !r.Fatal {
		t.Fatal("empty output must be Fatal")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
