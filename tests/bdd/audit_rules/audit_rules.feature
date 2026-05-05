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
