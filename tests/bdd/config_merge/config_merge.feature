Feature: Config merge precedence
  As a developer using Codify
  I want config values to follow flag > project > user > builtin precedence
  So that my command-line overrides always win and project-level shared defaults beat user-level personal defaults

  Scenario: Builtin defaults when no config files exist
    Given no user config exists
    And no project config exists
    When I load the effective config
    Then the effective preset should be "clean-ddd"
    And the effective locale should be "en"
    And the effective target should be "claude"

  Scenario: User config overrides builtin
    Given a user config with preset "neutral" and locale "es"
    And no project config exists
    When I load the effective config
    Then the effective preset should be "neutral"
    And the effective locale should be "es"
    And the effective target should be "claude"

  Scenario: Project config overrides user
    Given a user config with preset "neutral" and locale "es"
    And a project config with preset "hexagonal"
    When I load the effective config
    Then the effective preset should be "hexagonal"
    And the effective locale should be "es"

  Scenario: Empty fields in project do not override user
    Given a user config with preset "neutral" and locale "es"
    And a project config with preset "hexagonal"
    When I load the effective config
    Then the effective locale should be "es"

  Scenario: Roundtrip save and load preserves fields
    Given an empty home directory
    When I save a user config with preset "event-driven", locale "es", and target "codex"
    And I load the user config from disk
    Then the loaded preset should be "event-driven"
    And the loaded locale should be "es"
    And the loaded target should be "codex"
    And the loaded version should equal the schema version

  Scenario: Save creates a backup of the previous file
    Given an empty home directory
    When I save a user config with preset "neutral"
    And I save a user config with preset "hexagonal"
    Then a backup file ".bak" should exist next to the user config

  Scenario: Get unknown key returns error
    Given a config in memory
    When I get the value of key "nonexistent"
    Then I should get a config error containing "unknown config key"

  Scenario: Set unknown key returns error
    Given a config in memory
    When I set the value of key "nonexistent" to "x"
    Then I should get a config error containing "unknown config key"
