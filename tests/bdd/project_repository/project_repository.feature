Feature: Project Repository (In-Memory)
  As a developer
  I want to persist and retrieve projects
  So that I can manage project storage

  Background:
    Given an empty project repository

  Scenario: Save and retrieve project by ID
    Given I have a project with id "proj-1"
    When I save the project
    Then the project should be saved successfully
    When I retrieve project by id "proj-1"
    Then I should receive the project

  Scenario: Save and retrieve project by name
    Given I have a project with name "my-api"
    When I save the project
    Then the project should be saved successfully
    When I retrieve project by name "my-api"
    Then I should receive the project

  Scenario: Retrieve non-existent project by ID fails
    When I try to retrieve project by id "non-existent"
    Then I should get an error containing "not found"

  Scenario: Retrieve non-existent project by name fails
    When I try to retrieve project by name "non-existent"
    Then I should get an error containing "not found"

  Scenario: Save multiple projects
    Given I have a project with id "proj-1"
    And I have a project with id "proj-2"
    And I have a project with id "proj-3"
    When I save all projects
    Then all 3 projects should be saved
    When I retrieve all projects
    Then I should receive 3 projects

  Scenario: Update existing project
    Given I have a project with id "proj-1"
    When I save the project
    And I add a capability to the project
    And I save the project again
    Then the project should have the new capability

  Scenario: Delete existing project
    Given I have a project with id "proj-1"
    When I save the project
    And I delete project with id "proj-1"
    Then the project should be deleted
    When I try to retrieve project by id "proj-1"
    Then I should get an error containing "not found"

  Scenario: Delete non-existent project fails
    When I try to delete project with id "non-existent"
    Then I should get an error containing "not found"

  Scenario: Check if project exists by ID
    Given I have a project with id "proj-1"
    When I save the project
    Then project with id "proj-1" should exist
    And project with id "non-existent" should not exist

  Scenario: Check if project exists by name
    Given I have a project with name "my-api"
    When I save the project
    Then project with name "my-api" should exist
    And project with name "non-existent" should not exist

  Scenario: Prevent duplicate project names
    Given I have a project with name "my-api" and id "proj-1"
    And I have a project with name "my-api" and id "proj-2"
    When I save the first project
    Then the project should be saved successfully
    When I save the second project
    Then the second project should overwrite by name

  Scenario: Save invalid project fails
    Given I have an invalid project
    When I try to save the project
    Then I should get a validation error

  Scenario: Save nil project fails
    When I try to save a nil project
    Then I should get an error containing "cannot be nil"

  Scenario: Repository is thread-safe
    Given I have 10 projects
    When I save all projects concurrently
    Then all 10 projects should be saved
    And the repository count should be 10

  Scenario: Clear repository
    Given I have 5 projects saved
    When I clear the repository
    Then the repository should be empty
    And the repository count should be 0