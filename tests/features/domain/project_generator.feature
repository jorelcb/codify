Feature: Project Generator Service
  As a developer using the AI Context Generator
  I want to generate projects through a domain service
  So that I can ensure business rules are enforced

  Background:
    Given the project generator service is initialized with a mock repository

  Scenario: Create a new project successfully
    Given no project with name "my-api" exists
    When I create a project with:
      | field        | value    |
      | id           | proj-1   |
      | name         | my-api   |
      | language     | go       |
      | type         | api      |
      | architecture | ddd      |
      | output_path  | ./output |
    Then the project should be created successfully
    And the project should be saved to the repository

  Scenario: Add capabilities to project during creation
    Given no project with name "my-service" exists
    When I create a project with:
      | field        | value       |
      | id           | proj-2      |
      | name         | my-service  |
      | language     | go          |
      | type         | microservice|
      | architecture | hexagonal   |
      | output_path  | ./output    |
    And I add capabilities "database,messaging,auth"
    Then the project should be created successfully
    And the project should have 3 capabilities

  Scenario: Prevent duplicate project creation
    Given a project with name "existing-api" already exists
    When I try to create a project with name "existing-api"
    Then I should get an error "project with name existing-api already exists"
    And the project should not be saved

  Scenario: Get project by ID
    Given a project with id "proj-1" exists in the repository
    When I get project by id "proj-1"
    Then I should receive the project successfully

  Scenario: Get project by name
    Given a project with name "my-api" exists in the repository
    When I get project by name "my-api"
    Then I should receive the project successfully

  Scenario: List all projects
    Given 3 projects exist in the repository
    When I list all projects
    Then I should receive 3 projects

  Scenario: Delete existing project
    Given a project with id "proj-1" exists in the repository
    When I delete project with id "proj-1"
    Then the project should be deleted successfully
    And the project should be removed from the repository

  Scenario: Fail to delete non-existent project
    Given no project with id "non-existent" exists
    When I try to delete project with id "non-existent"
    Then I should get an error "project with id non-existent does not exist"

  Scenario: Validate project before saving
    Given I have invalid project data with empty output path
    When I try to create the project
    Then I should get a validation error
    And the project should not be saved