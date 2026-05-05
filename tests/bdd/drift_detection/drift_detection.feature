Feature: Drift detection between snapshot and FS
  As a developer using Codify
  I want `codify check` to detect divergence between my snapshot and the current FS
  So that I know when artifacts or input signals have drifted out of sync

  Scenario: No drift when nothing changed
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When I detect drift against the same state
    Then the report should be empty
    And the report should not have significant drift

  Scenario: Modified artifact is detected as significant drift
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the AGENTS.md file is modified
    And I detect drift against the snapshot
    Then the report should contain a "artifact_modified" entry for "AGENTS.md"
    And the report should have significant drift

  Scenario: Missing artifact is detected as significant drift
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the AGENTS.md file is removed
    And I detect drift against the snapshot
    Then the report should contain a "artifact_missing" entry for "AGENTS.md"
    And the report should have significant drift

  Scenario: New artifact appearing is detected as minor drift
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When a new artifact "context/CONTEXT.md" is added
    And I detect drift against the snapshot
    Then the report should contain a "artifact_new" entry for "context/CONTEXT.md"
    And the entry severity should be "minor"

  Scenario: Changed input signal is detected as significant drift
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the go.mod content changes
    And I detect drift against the snapshot
    Then the report should contain a "signal_changed" entry for "go.mod"
    And the report should have significant drift

  Scenario: Removed input signal is detected as significant drift
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the go.mod file is removed
    And I detect drift against the snapshot
    Then the report should contain a "signal_removed" entry for "go.mod"
    And the report should have significant drift

  Scenario: Multiple drifts in a single report
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the AGENTS.md file is modified
    And the go.mod content changes
    And I detect drift against the snapshot
    Then the report should have at least 2 entries
    And the report should have significant drift

  Scenario: Severity classification respects the kind
    Given a project with AGENTS.md and go.mod
    And a snapshot captured from the current state
    When the AGENTS.md file is modified
    And a new artifact "context/CONTEXT.md" is added
    And I detect drift against the snapshot
    Then the report should have significant drift
    And the report should contain at least one "minor" entry
