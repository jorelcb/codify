Feature: File watcher event loop
  As a developer using `codify watch`
  I want the foreground watcher to debounce file changes and only fire on watched paths
  So that I get accurate, low-noise notifications during active development

  Scenario: Watcher fires after debounce on a write to a watched file
    Given a temp project with a watched file "agents.md"
    When I start the watcher with debounce 200ms
    And I modify the watched file
    And I wait 600ms
    Then exactly 1 watch event should have fired
    And the event paths should include "agents.md"

  Scenario: Watcher ignores writes to unwatched files in the same directory
    Given a temp project with a watched file "agents.md"
    And an unwatched sibling file "other.txt"
    When I start the watcher with debounce 200ms
    And I modify the unwatched file
    And I wait 600ms
    Then no watch events should have fired

  Scenario: Watcher coalesces multiple rapid writes into one event
    Given a temp project with a watched file "agents.md"
    When I start the watcher with debounce 300ms
    And I modify the watched file 5 times in 100ms
    And I wait 700ms
    Then exactly 1 watch event should have fired

  Scenario: Watcher exits cleanly on context cancellation
    Given a temp project with a watched file "agents.md"
    When I start the watcher with debounce 200ms
    And I cancel the watcher context
    Then the watcher should return without error within 2 seconds

  Scenario: Watcher rejects construction when no path is provided
    When I create a watcher with no paths
    Then the watcher should return an error containing "at least one path"

  Scenario: Watcher rejects construction when no callback is provided
    Given a temp project with a watched file "agents.md"
    When I create a watcher without an OnEvent callback
    Then the watcher should return an error containing "OnEvent"
