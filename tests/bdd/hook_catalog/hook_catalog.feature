Feature: Hook catalog management
  As a developer using Codify
  I want to browse and resolve Claude Code hook presets from the catalog
  So that I can generate hook bundles for my project

  Scenario: Find hook category by name
    Given the hook catalog is loaded
    When I look up hook category "hooks"
    Then I should find a hook category with name "hooks"
    And the hook category should have 3 options

  Scenario: Find unknown hook category returns error
    Given the hook catalog is loaded
    When I look up hook category "nonexistent"
    Then I should get a hook catalog error containing "unknown hook category"

  Scenario: Resolve linting preset
    Given the hook catalog is loaded
    And I have hook category "hooks"
    When I resolve hook preset "linting"
    Then the resolved template directory should be "hooks/linting"
    And the resolved template mapping should be nil

  Scenario: Resolve security-guardrails preset
    Given the hook catalog is loaded
    And I have hook category "hooks"
    When I resolve hook preset "security-guardrails"
    Then the resolved template directory should be "hooks/security-guardrails"
    And the resolved template mapping should be nil

  Scenario: Resolve convention-enforcement preset
    Given the hook catalog is loaded
    And I have hook category "hooks"
    When I resolve hook preset "convention-enforcement"
    Then the resolved template directory should be "hooks/convention-enforcement"
    And the resolved template mapping should be nil

  Scenario: Resolve unknown preset returns error
    Given the hook catalog is loaded
    And I have hook category "hooks"
    When I resolve hook preset "nonexistent"
    Then I should get a hook catalog error containing "unknown preset"

  Scenario: All hook descriptions respect 250 char limit
    Given the hook catalog is loaded
    Then all hook descriptions should be at most 250 characters

  Scenario: Hook category names returns registered categories
    Given the hook catalog is loaded
    When I retrieve hook category names
    Then the hook category names should contain "hooks"

  Scenario: Hook preset names enumerates registered presets
    Given the hook catalog is loaded
    When I retrieve hook preset names
    Then the hook preset names should contain "linting"
    And the hook preset names should contain "security-guardrails"
    And the hook preset names should contain "convention-enforcement"
