Feature: Template Processor Service
  As a developer using the AI Context Generator
  I want to process templates with variables
  So that I can generate customized content

  Background:
    Given the template processor service is initialized

  Scenario: Process template with all variables provided
    Given I have a template with content "# {{PROJECT_NAME}} - {{LANGUAGE}}"
    And the template has variable "PROJECT_NAME" required true
    And the template has variable "LANGUAGE" required true
    And I have variables:
      | name         | value   |
      | PROJECT_NAME | my-api  |
      | LANGUAGE     | Go      |
    When I process the template
    Then the processed content should be "# my-api - Go"

  Scenario: Process template with default values
    Given I have a template with content "# {{PROJECT_NAME}} - {{AUTHOR}}"
    And the template has variable "PROJECT_NAME" required true
    And the template has variable "AUTHOR" required false with default "Anonymous"
    And I have variables:
      | name         | value  |
      | PROJECT_NAME | my-api |
    When I process the template
    Then the processed content should be "# my-api - Anonymous"

  Scenario: Fail when required variable is missing
    Given I have a template with content "# {{PROJECT_NAME}}"
    And the template has variable "PROJECT_NAME" required true
    And I have no variables
    When I process the template
    Then I should get an error "required variable PROJECT_NAME not provided"

  Scenario: Detect unprocessed variables
    Given I have a template with content "# {{PROJECT_NAME}} - {{UNKNOWN}}"
    And the template has variable "PROJECT_NAME" required true
    And I have variables:
      | name         | value  |
      | PROJECT_NAME | my-api |
    When I process the template
    Then I should get an error containing "unprocessed variables"

  Scenario: Extract variables from template content
    Given I have template content "# {{PROJECT_NAME}}\n{{LANGUAGE}} - {{TYPE}}\n{{PROJECT_NAME}}"
    When I extract variables from the content
    Then I should find 3 unique variables
    And the variables should include "PROJECT_NAME"
    And the variables should include "LANGUAGE"
    And the variables should include "TYPE"

  Scenario: Validate template with all required variables
    Given I have a template with content "# {{PROJECT_NAME}}"
    And the template has variable "PROJECT_NAME" required true
    And I have variables:
      | name         | value  |
      | PROJECT_NAME | my-api |
    When I validate the template
    Then the validation should pass

  Scenario: Validate fails when required variable is missing
    Given I have a template with content "# {{PROJECT_NAME}}"
    And the template has variable "PROJECT_NAME" required true
    And I have no variables
    When I validate the template
    Then the validation should fail with "required variable PROJECT_NAME is missing"