Feature: Template Entity
  As a developer using the AI Context Generator
  I want to create and manage template entities
  So that I can ensure templates have valid structure and metadata

  Background:
    Given the template system is initialized

  Scenario: Create a valid template
    Given I have template data with id "template-1", name "API Template", path "/templates/api.md", and content "# {{PROJECT_NAME}}"
    When I create a new template
    Then the template should be created successfully
    And the template id should be "template-1"
    And the template name should be "API Template"

  Scenario: Fail to create template without ID
    Given I have template data with empty id, name "Test", path "/path", and content "content"
    When I create a new template
    Then I should get an error "template id cannot be empty"

  Scenario: Fail to create template without name
    Given I have template data with id "test-1", empty name, path "/path", and content "content"
    When I create a new template
    Then I should get an error "template name cannot be empty"

  Scenario: Fail to create template without path
    Given I have template data with id "test-1", name "Test", empty path, and content "content"
    When I create a new template
    Then I should get an error "template path cannot be empty"

  Scenario: Add variable to template
    Given I have a valid template
    And I have a variable with name "PROJECT_NAME", required true, and default ""
    When I add the variable to the template
    Then the variable should be added successfully
    And the template should have 1 variable

  Scenario: Prevent duplicate variables
    Given I have a valid template
    And I have a variable with name "PROJECT_NAME", required true, and default ""
    And the variable is already added to the template
    When I try to add the same variable again
    Then I should get an error "variable PROJECT_NAME already exists"

  Scenario: Update template content
    Given I have a valid template with content "Old content"
    When I update the template content to "New content"
    Then the template content should be "New content"
    And the template updated timestamp should be recent

  Scenario: Set template metadata
    Given I have a valid template
    When I set metadata with version "1.0", author "John Doe", and description "API template"
    Then the template metadata version should be "1.0"
    And the template metadata author should be "John Doe"
    And the template metadata description should be "API template"

  Scenario: Validate template
    Given I have a valid template
    When I validate the template
    Then the validation should pass

  Scenario: Validate template is called on valid template
    Given I have a valid template
    When I validate the template
    Then the validation should pass