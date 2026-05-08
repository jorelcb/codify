Feature: Workflow catalog management
  As a developer using Codify
  I want to browse and resolve workflow presets from the catalog
  So that I can generate workflow files for my project

  Scenario: Find workflow category by name
    Given the workflow catalog is loaded
    When I look up workflow category "workflows"
    Then I should find a workflow category with name "workflows"
    And the workflow category should have 3 options

  Scenario: Find unknown workflow category returns error
    Given the workflow catalog is loaded
    When I look up workflow category "nonexistent"
    Then I should get a workflow catalog error containing "unknown workflow category"

  Scenario: Resolve bug-fix preset
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "bug-fix"
    Then the resolved template directory should be "workflows"
    And the resolved mapping should have 1 entry

  Scenario: Resolve release-cycle preset
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "release-cycle"
    Then the resolved template directory should be "workflows"
    And the resolved mapping should have 1 entry

  Scenario: Resolve spec-driven-change preset
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "spec-driven-change"
    Then the resolved template directory should be "sdd/openspec/workflows"
    And the resolved mapping should have 3 entries

  Scenario: Resolve all workflows combines all presets
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "all"
    Then the resolved mapping should have 5 entries

  Scenario: Resolve unknown preset returns error
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "nonexistent"
    Then I should get a workflow catalog error containing "unknown preset"

  Scenario: Generate Antigravity workflow frontmatter for known workflow
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "bug_fix"
    Then the frontmatter should start with "---"
    And the frontmatter should contain "description:"
    And the frontmatter should end with "---"

  Scenario: Generate Claude workflow frontmatter with skill fields
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "spec_propose" targeting "claude"
    Then the frontmatter should start with "---"
    And the frontmatter should contain "name: spec-propose"
    And the frontmatter should contain "description:"
    And the frontmatter should contain "disable-model-invocation: true"
    And the frontmatter should contain "allowed-tools: Bash(*)"
    And the frontmatter should end with "---"

  Scenario: Generate Claude workflow frontmatter for bug-fix
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "bug_fix" targeting "claude"
    Then the frontmatter should contain "name: bug-fix"
    And the frontmatter should contain "disable-model-invocation: true"

  Scenario: Antigravity frontmatter does not contain skill fields
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "bug_fix" targeting "antigravity"
    Then the frontmatter should contain "description:"
    And the frontmatter should not contain "disable-model-invocation"
    And the frontmatter should not contain "allowed-tools"
    And the frontmatter should not contain "name:"

  Scenario: Generate workflow frontmatter for unknown workflow uses fallback
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "unknown_workflow"
    Then the frontmatter should contain "Workflow for unknown-workflow"

  Scenario: All workflow descriptions respect 250 char limit
    Given the workflow catalog is loaded
    Then all workflow descriptions should be at most 250 characters

  Scenario: Workflow category names returns registered categories
    Given the workflow catalog is loaded
    When I retrieve workflow category names
    Then the workflow category names should contain "workflows"

  # --- Claude skill annotation stripping scenarios ---

  Scenario: Claude skill content strips annotation lines
    Given the workflow catalog is loaded
    When I strip annotations from the bug-fix template
    Then the stripped content should not contain "// turbo"
    And the stripped content should not contain "// capture:"
    And the stripped content should not contain "// if "

  Scenario: Claude skill content preserves non-annotation content
    Given the workflow catalog is loaded
    When I strip annotations from the bug-fix template
    Then the stripped content should contain "Create a branch"
    And the stripped content should contain "Bug Fix Workflow"
    And the stripped content should contain "Run the full test suite"

  Scenario: Antigravity target still generates flat frontmatter
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "bug_fix" targeting "antigravity"
    Then the frontmatter should not contain "disable-model-invocation"
    And the frontmatter should not contain "name:"
