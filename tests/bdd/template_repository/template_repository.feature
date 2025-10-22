Feature: Template Repository (In-Memory)
  As a developer
  I want to persist and retrieve templates
  So that I can manage template storage

  Background:
    Given an empty template repository

  Scenario: Save and retrieve template by ID
    Given I have a template with id "tmpl-1"
    When I save the template
    Then the template should be saved successfully
    When I retrieve template by id "tmpl-1"
    Then I should receive the template

  Scenario: Save and retrieve template by path
    Given I have a template with path "/templates/api.md"
    When I save the template
    Then the template should be saved successfully
    When I retrieve template by path "/templates/api.md"
    Then I should receive the template

  Scenario: Retrieve non-existent template by ID fails
    When I try to retrieve template by id "non-existent"
    Then I should get an error containing "not found"

  Scenario: Retrieve non-existent template by path fails
    When I try to retrieve template by path "/non/existent"
    Then I should get an error containing "not found"

  Scenario: Save multiple templates
    Given I have a template with id "tmpl-1"
    And I have a template with id "tmpl-2"
    And I have a template with id "tmpl-3"
    When I save all templates
    Then all 3 templates should be saved
    When I retrieve all templates
    Then I should receive 3 templates

  Scenario: Find templates by tag
    Given I have a template with id "tmpl-1" and tag "api"
    And I have a template with id "tmpl-2" and tag "api"
    And I have a template with id "tmpl-3" and tag "cli"
    When I save all templates
    And I find templates by tag "api"
    Then I should receive 2 templates

  Scenario: Update existing template
    Given I have a template with id "tmpl-1"
    When I save the template
    And I update the template content
    And I save the template again
    Then the template should have updated content

  Scenario: Delete existing template
    Given I have a template with id "tmpl-1"
    When I save the template
    And I delete template with id "tmpl-1"
    Then the template should be deleted
    When I try to retrieve template by id "tmpl-1"
    Then I should get an error containing "not found"

  Scenario: Delete non-existent template fails
    When I try to delete template with id "non-existent"
    Then I should get an error containing "not found"

  Scenario: Check if template exists
    Given I have a template with id "tmpl-1"
    When I save the template
    Then template with id "tmpl-1" should exist
    And template with id "non-existent" should not exist

  Scenario: Save invalid template fails
    Given I have an invalid template
    When I try to save the template
    Then I should get a validation error

  Scenario: Save nil template fails
    When I try to save a nil template
    Then I should get an error containing "cannot be nil"

  Scenario: Repository is thread-safe
    Given I have 10 templates
    When I save all templates concurrently
    Then all 10 templates should be saved
    And the repository count should be 10

  Scenario: Clear repository
    Given I have 5 templates saved
    When I clear the repository
    Then the repository should be empty
    And the repository count should be 0