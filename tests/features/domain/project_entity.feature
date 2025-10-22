Feature: Project Entity
  As a developer using the AI Context Generator
  I want to create and manage project entities
  So that I can generate projects with proper configuration

  Background:
    Given the project system is initialized

  Scenario: Create a valid project
    Given I have project data with:
      | field        | value           |
      | id           | project-1       |
      | name         | my-api          |
      | language     | go              |
      | type         | api             |
      | architecture | ddd             |
      | output_path  | ./output        |
    When I create a new project
    Then the project should be created successfully
    And the project id should be "project-1"
    And the project name should be "my-api"
    And the project language should be "go"
    And the project type should be "api"
    And the project architecture should be "ddd"

  Scenario: Fail to create project without ID
    Given I have project data with empty id
    When I create a new project
    Then I should get an error "project id cannot be empty"

  Scenario: Fail to create project without output path
    Given I have project data with empty output path
    When I create a new project
    Then I should get an error "output path cannot be empty"

  Scenario: Add capability to project
    Given I have a valid project
    When I add capability "database"
    Then the capability should be added successfully
    And the project should have 1 capability
    And the project capabilities should include "database"

  Scenario: Add multiple capabilities
    Given I have a valid project
    When I add capability "database"
    And I add capability "messaging"
    And I add capability "auth"
    Then the project should have 3 capabilities
    And the project capabilities should include "database"
    And the project capabilities should include "messaging"
    And the project capabilities should include "auth"

  Scenario: Prevent duplicate capabilities
    Given I have a valid project
    And I have added capability "database"
    When I try to add capability "database" again
    Then I should get an error "capability database already exists"

  Scenario: Set project metadata
    Given I have a valid project
    When I set metadata "author" to "John Doe"
    And I set metadata "version" to "1.0.0"
    Then the project metadata "author" should be "John Doe"
    And the project metadata "version" should be "1.0.0"

  Scenario: Get project full path
    Given I have a project with name "my-api" and output path "./output"
    When I get the project full path
    Then the full path should be "./output/my-api"

  Scenario: Validate project
    Given I have a valid project
    When I validate the project
    Then the validation should pass