Feature: Audit rules-only mode (deterministic)
  As a developer using Codify
  I want `codify audit` to flag commits that violate Conventional Commits
  So that the codebase maintains a clean, standards-aligned history

  Scenario: Valid Conventional Commit produces no findings
    When I audit a commit with header "feat(api): add new endpoint"
    Then the audit findings should be empty

  Scenario: Breaking change marker is valid
    When I audit a commit with header "feat!: drop legacy support"
    Then the audit findings should be empty

  Scenario: Invalid commit type is flagged
    When I audit a commit with header "improvement(api): tweak something"
    Then the audit should contain a "commit_invalid_type" finding
    And the finding severity should be "significant"

  Scenario: Header longer than 72 chars is flagged
    When I audit a commit with a 90-character header
    Then the audit should contain a "commit_header_too_long" finding
    And the finding severity should be "minor"

  Scenario: Trivial messages are flagged
    When I audit a commit with header "wip"
    Then the audit should contain a "commit_trivial" finding

  Scenario: Generic non-Conventional-Commit headers are flagged
    When I audit a commit with header "Fix that bug we discussed"
    Then the audit should contain a "commit_invalid_type" finding

  Scenario: Merge commit detection
    When I check if a commit with 2 parents is a merge commit
    Then it should be reported as a merge commit
    When I check if a commit with 1 parent is a merge commit
    Then it should not be reported as a merge commit

  Scenario: LLM findings parser handles a valid JSON array
    When I parse the LLM response '[{"commit_sha":"abc","severity":"significant","detail":"violates DDD"}]'
    Then the parsed findings should contain 1 entry
    And the parsed finding should have heuristic flag set
    And the parsed finding kind should be "agents_alignment_issue"

  Scenario: LLM findings parser strips markdown fences
    When I parse the LLM response wrapped in JSON code fences
    Then the parsed findings should contain 1 entry

  Scenario: LLM findings parser falls back to minor on invalid severity
    When I parse the LLM response '[{"commit_sha":"abc","severity":"critical","detail":"x"}]'
    Then the parsed finding severity should be "minor"

  Scenario: LLM findings parser rejects non-JSON input
    When I parse the LLM response "this is not JSON"
    Then the parser should report an error
