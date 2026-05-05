Feature: Architecture preset catalog
  As a developer using Codify
  I want to resolve architecture presets from the skills catalog
  So that the right templates are loaded for each architectural style

  Scenario: Resolve neutral preset
    Given the skills catalog is loaded
    And I have skill category "architecture"
    When I resolve architecture preset "neutral"
    Then the resolved template directory should be "neutral"
    And the resolved mapping should have 4 entries

  Scenario: Resolve clean-ddd preset
    Given the skills catalog is loaded
    And I have skill category "architecture"
    When I resolve architecture preset "clean-ddd"
    Then the resolved template directory should be "clean-ddd"
    And the resolved mapping should have 5 entries

  Scenario: Resolve hexagonal preset
    Given the skills catalog is loaded
    And I have skill category "architecture"
    When I resolve architecture preset "hexagonal"
    Then the resolved template directory should be "hexagonal"
    And the resolved mapping should have 4 entries

  Scenario: Resolve event-driven preset
    Given the skills catalog is loaded
    And I have skill category "architecture"
    When I resolve architecture preset "event-driven"
    Then the resolved template directory should be "event-driven"
    And the resolved mapping should have 5 entries

  Scenario: Legacy "default" alias resolves to clean-ddd
    Given the skills catalog is loaded
    When I look up legacy preset alias "default"
    Then the legacy alias should map to category "architecture" and preset "clean-ddd"

  Scenario: Legacy "clean" alias still resolves to clean-ddd
    Given the skills catalog is loaded
    When I look up legacy preset alias "clean"
    Then the legacy alias should map to category "architecture" and preset "clean-ddd"

  Scenario: Architecture category exposes 4 presets
    Given the skills catalog is loaded
    And I have skill category "architecture"
    When I retrieve architecture preset names
    Then the architecture preset names should contain "neutral"
    And the architecture preset names should contain "clean-ddd"
    And the architecture preset names should contain "hexagonal"
    And the architecture preset names should contain "event-driven"
