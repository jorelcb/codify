Feature: SDD standard pluggable selection
  As a developer using codify
  I want to choose between OpenSpec and Spec-Kit (and future standards)
  So that codify spec produces the file set my team has standardized on

  Background:
    Given the SDD registry is loaded with default adapters

  # ===========================================================================
  # OpenSpec adapter — preserva el comportamiento histórico de codify v1.x
  # ===========================================================================

  Scenario: OpenSpec adapter is registered as default
    When I look up SDD standard "openspec"
    Then the lookup should succeed
    And the standard's display name should be "OpenSpec"
    And the standard's template directory should be "openspec"
    And the standard's output layout should be "flat"

  Scenario: OpenSpec produces four required files at the root of specs/
    When I look up SDD standard "openspec"
    Then the bootstrap artifacts should include exactly these files:
      | CONSTITUTION.md |
      | SPEC.md         |
      | PLAN.md         |
      | TASKS.md        |
    And every bootstrap artifact should be marked required

  Scenario: OpenSpec ships the propose/apply/archive lifecycle workflows
    When I look up SDD standard "openspec"
    Then the lifecycle workflow IDs should be:
      | spec_propose |
      | spec_apply   |
      | spec_archive |

  # ===========================================================================
  # Spec-Kit adapter — formato GitHub Spec-Kit
  # ===========================================================================

  Scenario: Spec-Kit adapter is registered alongside OpenSpec
    When I look up SDD standard "spec-kit"
    Then the lookup should succeed
    And the standard's display name should be "GitHub Spec-Kit"
    And the standard's template directory should be "spec-kit"
    And the standard's output layout should be "feature-grouped"

  Scenario: Spec-Kit produces lowercase-named files
    When I look up SDD standard "spec-kit"
    Then the bootstrap artifacts should include exactly these files:
      | spec.md       |
      | plan.md       |
      | tasks.md      |
      | research.md   |
      | data-model.md |
      | quickstart.md |
    And every required artifact should have a lowercase file name

  Scenario: Spec-Kit does not include a constitution file
    When I look up SDD standard "spec-kit"
    Then no bootstrap artifact should have a name containing "constitution"

  Scenario: Spec-Kit's spec, plan, tasks are required and the rest are optional
    When I look up SDD standard "spec-kit"
    Then the required artifact files should be exactly:
      | spec.md  |
      | plan.md  |
      | tasks.md |
    And the optional artifact files should be exactly:
      | research.md   |
      | data-model.md |
      | quickstart.md |

  Scenario: Spec-Kit ships specify/plan/tasks lifecycle workflows
    When I look up SDD standard "spec-kit"
    Then the lifecycle workflow IDs should be:
      | speckit_specify |
      | speckit_plan    |
      | speckit_tasks   |

  Scenario: Spec-Kit hints remind the LLM about lowercase and per-feature layout
    When I look up SDD standard "spec-kit"
    Then the system prompt hints in "en" should mention "lowercase"
    And the system prompt hints in "en" should mention "specs/<feature-id>/"
    And the system prompt hints in "es" should mention "lowercase"
    And the system prompt hints in "es" should mention "specs/<feature-id>/"

  # ===========================================================================
  # Resolución por precedencia (ADR-0011)
  # ===========================================================================

  Scenario: Resolution falls through to the default when nothing is set
    When I resolve with flag "" project "" user ""
    Then the resolved standard ID should be "openspec"

  Scenario: User config alone selects the standard
    When I resolve with flag "" project "" user "spec-kit"
    Then the resolved standard ID should be "spec-kit"

  Scenario: Project config overrides user config
    When I resolve with flag "" project "spec-kit" user "openspec"
    Then the resolved standard ID should be "spec-kit"

  Scenario: CLI flag overrides project and user config
    When I resolve with flag "openspec" project "spec-kit" user "spec-kit"
    Then the resolved standard ID should be "openspec"

  Scenario: Unknown flag fails with explicit error listing available standards
    When I resolve with flag "does-not-exist" project "" user ""
    Then resolution should fail with error containing "does-not-exist"
    And resolution should fail with error containing "openspec"
    And resolution should fail with error containing "spec-kit"

  Scenario: Unknown project config value fails before silently falling back
    When I resolve with flag "" project "phantom" user "openspec"
    Then resolution should fail with error containing "phantom"

  # ===========================================================================
  # Templates físicas en el embedded FS
  # ===========================================================================

  Scenario: OpenSpec spec templates exist in the embedded filesystem (en)
    Then the embedded FS should contain template "templates/en/sdd/openspec/spec/constitution.template"
    And the embedded FS should contain template "templates/en/sdd/openspec/spec/spec.template"
    And the embedded FS should contain template "templates/en/sdd/openspec/spec/plan.template"
    And the embedded FS should contain template "templates/en/sdd/openspec/spec/tasks.template"

  Scenario: Spec-Kit spec templates exist in the embedded filesystem (en)
    Then the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_spec.template"
    And the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_plan.template"
    And the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_tasks.template"
    And the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_research.template"
    And the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_data_model.template"
    And the embedded FS should contain template "templates/en/sdd/spec-kit/spec/speckit_quickstart.template"

  Scenario: Both standards have parallel ES templates
    Then the embedded FS should contain template "templates/es/sdd/openspec/spec/spec.template"
    And the embedded FS should contain template "templates/es/sdd/spec-kit/spec/speckit_spec.template"
