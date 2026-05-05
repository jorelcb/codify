Feature: LLM usage tracking
  As a developer using Codify
  I want every LLM invocation to be recorded with tokens and cost
  So that I can monitor my spending without surprises

  Scenario: Recording an entry persists to disk
    Given an empty home directory
    When I record a usage entry with model "claude-sonnet-4-6", 100000 input tokens, and 50000 output tokens
    And I read the global usage log
    Then the log should contain 1 entry
    And the totals should report 150000 total tokens
    And the totals cost should be greater than 0

  Scenario: Cost calculation matches public list price for known model
    When I compute cost for model "claude-sonnet-4-6" with 1000000 input tokens and 0 output tokens
    Then the cost should be 300 cents

  Scenario: Cost calculation for unknown model returns zero
    When I compute cost for model "nonexistent-model" with 1000000 input tokens and 1000000 output tokens
    Then the cost should be 0 cents

  Scenario: Tracking is disabled by env variable
    Given the env variable "CODIFY_NO_USAGE_TRACKING" is set to "1"
    When I check if tracking is disabled
    Then tracking should be reported as disabled

  Scenario: Tracking is disabled by flag
    When I check tracking with the no-tracking flag enabled
    Then tracking should be reported as disabled

  Scenario: Recording is no-op when tracking is disabled
    Given an empty home directory
    And the env variable "CODIFY_NO_USAGE_TRACKING" is set to "1"
    When I record a usage entry with model "claude-sonnet-4-6", 1000 input tokens, and 500 output tokens
    Then the global usage file should not exist

  Scenario: Append accumulates entries and recomputes totals
    Given an empty home directory
    When I record a usage entry with model "claude-sonnet-4-6", 100000 input tokens, and 50000 output tokens
    And I record a usage entry with model "claude-sonnet-4-6", 200000 input tokens, and 100000 output tokens
    And I read the global usage log
    Then the log should contain 2 entries
    And the totals should report 450000 total tokens
    And the totals call count should be 2
