Feature: Value Objects
  As a developer using the AI Context Generator
  I want value objects to enforce business rules
  So that invalid data cannot enter the domain

  Scenario Outline: Create valid Language
    Given I have a language value "<language>"
    When I create a Language value object
    Then the Language should be created successfully
    And the Language value should be "<language>"

    Examples:
      | language   |
      | go         |
      | javascript |
      | typescript |
      | python     |
      | java       |
      | rust       |

  Scenario: Reject invalid language
    Given I have a language value "cobol"
    When I create a Language value object
    Then I should get an error "invalid language: cobol"

  Scenario: Reject empty language
    Given I have a language value ""
    When I create a Language value object
    Then I should get an error "language cannot be empty"

  Scenario Outline: Create valid ProjectType
    Given I have a project type value "<type>"
    When I create a ProjectType value object
    Then the ProjectType should be created successfully
    And the ProjectType value should be "<type>"

    Examples:
      | type        |
      | api         |
      | cli         |
      | library     |
      | microservice|
      | webapp      |

  Scenario: Reject invalid project type
    Given I have a project type value "invalidtype"
    When I create a ProjectType value object
    Then I should get an error "invalid project type: invalidtype"

  Scenario Outline: Create valid Architecture
    Given I have an architecture value "<arch>"
    When I create an Architecture value object
    Then the Architecture should be created successfully
    And the Architecture value should be "<arch>"

    Examples:
      | arch       |
      | ddd        |
      | clean      |
      | hexagonal  |
      | layered    |
      | mvc        |
      | cqrs       |

  Scenario: Reject invalid architecture
    Given I have an architecture value "spaghetti"
    When I create an Architecture value object
    Then I should get an error "invalid architecture: spaghetti"

  Scenario Outline: Create valid ProjectName
    Given I have a project name value "<name>"
    When I create a ProjectName value object
    Then the ProjectName should be created successfully
    And the ProjectName value should be "<name>"

    Examples:
      | name           |
      | my-api         |
      | user_service   |
      | awesome-cli    |
      | project123     |

  Scenario: Reject too short project name
    Given I have a project name value "a"
    When I create a ProjectName value object
    Then I should get an error "project name must be at least 2 characters"

  Scenario: Reject too long project name
    Given I have a project name value with 101 characters
    When I create a ProjectName value object
    Then I should get an error "project name must be less than 100 characters"

  Scenario: Reject empty project name
    Given I have a project name value ""
    When I create a ProjectName value object
    Then I should get an error "project name cannot be empty"