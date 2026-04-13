Feature: Workflow catalog management
  As a developer using Codify
  I want to browse and resolve workflow presets from the catalog
  So that I can generate Antigravity workflow files for my project

  Scenario: Find workflow category by name
    Given the workflow catalog is loaded
    When I look up workflow category "workflows"
    Then I should find a workflow category with name "workflows"
    And the workflow category should have 3 options

  Scenario: Find unknown workflow category returns error
    Given the workflow catalog is loaded
    When I look up workflow category "nonexistent"
    Then I should get a workflow catalog error containing "unknown workflow category"

  Scenario: Resolve feature-development preset
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "feature-development"
    Then the resolved template directory should be "workflows"
    And the resolved mapping should have 1 entry

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

  Scenario: Resolve all workflows combines all presets
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "all"
    Then the resolved mapping should have 3 entries

  Scenario: Resolve unknown preset returns error
    Given the workflow catalog is loaded
    And I have workflow category "workflows"
    When I resolve workflow preset "nonexistent"
    Then I should get a workflow catalog error containing "unknown preset"

  Scenario: Generate Antigravity workflow frontmatter for known workflow
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "feature_development"
    Then the frontmatter should start with "---"
    And the frontmatter should contain "description:"
    And the frontmatter should end with "---"

  Scenario: Generate Claude workflow frontmatter with user-invocable
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "feature_development" targeting "claude"
    Then the frontmatter should start with "---"
    And the frontmatter should contain "name: feature-development"
    And the frontmatter should contain "description:"
    And the frontmatter should contain "user-invocable: true"
    And the frontmatter should end with "---"

  Scenario: Generate Claude workflow frontmatter for bug-fix
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "bug_fix" targeting "claude"
    Then the frontmatter should contain "name: bug-fix"
    And the frontmatter should contain "user-invocable: true"

  Scenario: Antigravity frontmatter does not contain user-invocable
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "feature_development" targeting "antigravity"
    Then the frontmatter should contain "description:"
    And the frontmatter should not contain "user-invocable"
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

  # --- Plugin generation scenarios ---

  Scenario: Plugin manifest has correct metadata for release-cycle
    Given the workflow catalog is loaded
    When I generate a plugin manifest for "release_cycle"
    Then the plugin manifest should be valid JSON
    And the plugin manifest name should be "codify-wf-release-cycle"
    And the plugin manifest should have version "1.0.0"

  Scenario: Plugin hooks maps turbo annotations to PreToolUse
    Given the workflow catalog is loaded
    And I parse annotations from the release-cycle template
    When I generate plugin hooks from the annotations
    Then the plugin hooks should be valid JSON
    And the plugin hooks should contain "PreToolUse"
    And the plugin hooks should contain "permissionDecision"

  Scenario: Plugin hooks maps capture annotations to PostToolUse
    Given the workflow catalog is loaded
    And I parse annotations from the release-cycle template
    When I generate plugin hooks from the annotations
    Then the plugin hooks should contain "PostToolUse"
    And the plugin hooks should contain "capture-output.sh"

  Scenario: Plugin hooks maps if annotations to prompt hooks
    Given the workflow catalog is loaded
    And I parse annotations from the release-cycle template
    When I generate plugin hooks from the annotations
    Then the plugin hooks should contain "prompt"
    And the plugin hooks should contain "CI/CD deployment"

  Scenario: Plugin SKILL.md has no Antigravity annotations
    Given the workflow catalog is loaded
    When I generate a plugin skill for "feature_development"
    Then the plugin skill should not contain "// turbo"
    And the plugin skill should not contain "// capture:"
    And the plugin skill should not contain "// if "

  Scenario: Plugin SKILL.md preserves workflow content
    Given the workflow catalog is loaded
    When I generate a plugin skill for "feature_development"
    Then the plugin skill should contain "Feature Development"
    And the plugin skill should contain "name: feature-development"

  Scenario: Plugin agent has correct frontmatter in English
    Given the workflow catalog is loaded
    When I generate a plugin agent for "release_cycle" in "en"
    Then the plugin agent should contain "name: workflow-runner"
    And the plugin agent should contain "model: sonnet"
    And the plugin agent should contain "tools: Bash, Read, Edit, Write, Grep, Glob"
    And the plugin agent should contain "workflow execution agent"

  Scenario: Plugin agent has correct frontmatter in Spanish
    Given the workflow catalog is loaded
    When I generate a plugin agent for "feature_development" in "es"
    Then the plugin agent should contain "agente de ejecucion de workflow"
    And the plugin agent should contain "feature-development"

  Scenario: Antigravity target still generates flat frontmatter
    Given the workflow catalog is loaded
    When I generate workflow frontmatter for "feature_development" targeting "antigravity"
    Then the frontmatter should not contain "user-invocable"
    And the frontmatter should not contain "name:"
