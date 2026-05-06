Feature: Resolve [DEFINE] markers in generated files
  As a developer using Codify
  I want `codify resolve` to walk me through every [DEFINE: ...] marker
  So that I can integrate my answers without alt-tabbing to an editor

  Scenario: Resolving a single file with one answered marker
    Given a file "AGENTS.md" with content "Currency is [DEFINE: code]."
    And the user accepts the top-level prompt
    And the user answers "USD" for marker on line 1
    When the resolver runs over "AGENTS.md"
    Then the file "AGENTS.md" should equal "Currency is USD."
    And the resolve summary should report 1 marker resolved

  Scenario: Resolving multiple files reports both written
    Given a file "A.md" with content "[DEFINE: a]"
    And a file "B.md" with content "[DEFINE: b]"
    And the user accepts the top-level prompt
    And the user answers "alpha" for marker on line 1
    And the user answers "beta" for marker on line 1
    When the resolver runs over files "A.md" "B.md"
    Then the resolve summary should report 2 files rewritten

  Scenario: Skipping all markers with TODO mode anchors them in markdown comment style
    Given a file "AGENTS.md" with content "currency [DEFINE: code], tz [DEFINE: tz]"
    And the user accepts the top-level prompt
    And the user skips marker on line 1
    And the user answers "America/Mexico_City" for marker on line 1
    When the resolver runs over "AGENTS.md"
    Then the file "AGENTS.md" should contain "<!-- TODO"
    And the file "AGENTS.md" should contain "code -->"
    And the file "AGENTS.md" should contain "America/Mexico_City"

  Scenario: Verbatim skip mode leaves marker text in place
    Given a file "AGENTS.md" with content "currency [DEFINE: code]"
    And the user accepts the top-level prompt
    And the user skips marker on line 1
    When the resolver runs in verbatim skip mode over "AGENTS.md"
    Then the file "AGENTS.md" should equal "currency [DEFINE: code]"

  Scenario: Decline at the top-level prompt leaves files untouched
    Given a file "AGENTS.md" with content "currency [DEFINE: code]"
    And the user declines the top-level prompt
    When the resolver runs over "AGENTS.md"
    Then the file "AGENTS.md" should equal "currency [DEFINE: code]"
    And the resolve summary should report decline

  Scenario: LLM rewrite that hallucinates new markers falls back to literal substitution
    Given a file "AGENTS.md" with content "currency [DEFINE: code], tz [DEFINE: tz]"
    And the LLM provider returns rewritten content "currency USD, tz [DEFINE: tz] [DEFINE: hallucinated]"
    And the user accepts the top-level prompt
    And the user answers "USD" for marker on line 1
    And the user skips marker on line 1
    When the resolver runs over "AGENTS.md"
    Then the file "AGENTS.md" should contain "currency USD"
    And the file "AGENTS.md" should not contain "[DEFINE: hallucinated]"

  Scenario: Diff preview discards the rewrite and leaves the file untouched
    Given a file "AGENTS.md" with content "currency [DEFINE: code]"
    And the user accepts the top-level prompt
    And the user answers "USD" for marker on line 1
    And the diff preview action is "discard"
    When the resolver runs over "AGENTS.md"
    Then the file "AGENTS.md" should equal "currency [DEFINE: code]"
    And the resolve summary should report 1 file discarded
