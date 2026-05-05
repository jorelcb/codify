package architecture_catalog

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

var featureContext = new(FeatureContext)

// InitializeTestSuite is called once before all scenarios.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

// InitializeScenario registers step definitions for each scenario.
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// Given
	ctx.Step(`^the skills catalog is loaded$`, featureContext.theSkillsCatalogIsLoaded)
	ctx.Step(`^I have skill category "([^"]*)"$`, featureContext.iHaveSkillCategory)

	// When
	ctx.Step(`^I resolve architecture preset "([^"]*)"$`, featureContext.iResolveArchitecturePreset)
	ctx.Step(`^I look up legacy preset alias "([^"]*)"$`, featureContext.iLookUpLegacyPresetAlias)
	ctx.Step(`^I retrieve architecture preset names$`, featureContext.iRetrieveArchitecturePresetNames)

	// Then
	ctx.Step(`^the resolved template directory should be "([^"]*)"$`, featureContext.theResolvedTemplateDirShouldBe)
	ctx.Step(`^the resolved mapping should have (\d+) entr(?:y|ies)$`, featureContext.theResolvedMappingShouldHaveNEntries)
	ctx.Step(`^the legacy alias should map to category "([^"]*)" and preset "([^"]*)"$`, featureContext.theLegacyAliasShouldMapTo)
	ctx.Step(`^the architecture preset names should contain "([^"]*)"$`, featureContext.theArchitecturePresetNamesShouldContain)
}

func (f *FeatureContext) theSkillsCatalogIsLoaded() error { return nil }

func (f *FeatureContext) iHaveSkillCategory(name string) error {
	f.category, f.err = catalog.FindCategory(name)
	if f.err != nil {
		return fmt.Errorf("failed to find skill category %q: %w", name, f.err)
	}
	return nil
}

func (f *FeatureContext) iResolveArchitecturePreset(preset string) error {
	if f.category == nil {
		return fmt.Errorf("no skill category set")
	}
	f.selection, f.err = f.category.Resolve(preset)
	if f.err != nil {
		return fmt.Errorf("failed to resolve preset %q: %w", preset, f.err)
	}
	return nil
}

func (f *FeatureContext) iLookUpLegacyPresetAlias(alias string) error {
	mapped, ok := catalog.LegacyPresetMapping[alias]
	f.legacyMap = mapped
	f.legacyOk = ok
	return nil
}

func (f *FeatureContext) iRetrieveArchitecturePresetNames() error {
	if f.category == nil {
		return fmt.Errorf("no skill category set")
	}
	f.presetNames = f.category.OptionNames()
	return nil
}

func (f *FeatureContext) theResolvedTemplateDirShouldBe(expected string) error {
	if err := assertions.AssertActual(assert.NotNil, f.selection, "expected selection to be set"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.selection.TemplateDir, "template dir mismatch")
}

func (f *FeatureContext) theResolvedMappingShouldHaveNEntries(n int) error {
	if err := assertions.AssertActual(assert.NotNil, f.selection, "expected selection to be set"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.selection.TemplateMapping), "mapping count mismatch")
}

func (f *FeatureContext) theLegacyAliasShouldMapTo(category, preset string) error {
	if err := assertions.AssertBool(assert.True, f.legacyOk, "expected legacy alias to be present"); err != nil {
		return err
	}
	if err := assertions.AssertExpectedAndActual(assert.Equal, category, f.legacyMap[0], "legacy category mismatch"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, preset, f.legacyMap[1], "legacy preset mismatch")
}

func (f *FeatureContext) theArchitecturePresetNamesShouldContain(name string) error {
	for _, n := range f.presetNames {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("expected architecture preset names to contain %q, got %v", name, f.presetNames)
}
