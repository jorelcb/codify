package hook_catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

// featureContext is the singleton instance for this feature.
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

	// ========== Given Steps ==========
	ctx.Step(`^the hook catalog is loaded$`, featureContext.theHookCatalogIsLoaded)
	ctx.Step(`^I have hook category "([^"]*)"$`, featureContext.iHaveHookCategory)

	// ========== When Steps ==========
	ctx.Step(`^I look up hook category "([^"]*)"$`, featureContext.iLookUpHookCategory)
	ctx.Step(`^I resolve hook preset "([^"]*)"$`, featureContext.iResolveHookPreset)
	ctx.Step(`^I retrieve hook category names$`, featureContext.iRetrieveHookCategoryNames)
	ctx.Step(`^I retrieve hook preset names$`, featureContext.iRetrieveHookPresetNames)

	// ========== Then Steps ==========
	ctx.Step(`^I should find a hook category with name "([^"]*)"$`, featureContext.iShouldFindAHookCategoryWithName)
	ctx.Step(`^the hook category should have (\d+) options$`, featureContext.theHookCategoryShouldHaveNOptions)
	ctx.Step(`^I should get a hook catalog error containing "([^"]*)"$`, featureContext.iShouldGetAHookCatalogErrorContaining)
	ctx.Step(`^the resolved template directory should be "([^"]*)"$`, featureContext.theResolvedTemplateDirShouldBe)
	ctx.Step(`^the resolved template mapping should be nil$`, featureContext.theResolvedTemplateMappingShouldBeNil)
	ctx.Step(`^all hook descriptions should be at most (\d+) characters$`, featureContext.allHookDescriptionsShouldBeAtMostNChars)
	ctx.Step(`^the hook category names should contain "([^"]*)"$`, featureContext.theHookCategoryNamesShouldContain)
	ctx.Step(`^the hook preset names should contain "([^"]*)"$`, featureContext.theHookPresetNamesShouldContain)
}

// ========== Given / When implementations ==========

func (f *FeatureContext) theHookCatalogIsLoaded() error { return nil }

func (f *FeatureContext) iHaveHookCategory(name string) error {
	f.category, f.err = catalog.FindHookCategory(name)
	if f.err != nil {
		return fmt.Errorf("failed to find hook category %q: %w", name, f.err)
	}
	return nil
}

func (f *FeatureContext) iLookUpHookCategory(name string) error {
	f.category, f.err = catalog.FindHookCategory(name)
	return nil
}

func (f *FeatureContext) iResolveHookPreset(preset string) error {
	if f.category == nil {
		return fmt.Errorf("no hook category set")
	}
	f.selection, f.err = f.category.Resolve(preset)
	return nil
}

func (f *FeatureContext) iRetrieveHookCategoryNames() error {
	f.categoryNames = catalog.HookCategoryNames()
	return nil
}

func (f *FeatureContext) iRetrieveHookPresetNames() error {
	f.presetNames = catalog.HookPresetNames()
	return nil
}

// ========== Then implementations ==========

func (f *FeatureContext) iShouldFindAHookCategoryWithName(name string) error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	if err := assertions.AssertActual(assert.NotNil, f.category, "expected category to be found"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, name, f.category.Name, "category name mismatch")
}

func (f *FeatureContext) theHookCategoryShouldHaveNOptions(n int) error {
	if f.category == nil {
		return fmt.Errorf("no category available")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.category.Options), "option count mismatch")
}

func (f *FeatureContext) iShouldGetAHookCatalogErrorContaining(expected string) error {
	if f.err == nil {
		return fmt.Errorf("expected error containing %q, got nil", expected)
	}
	if !strings.Contains(f.err.Error(), expected) {
		return fmt.Errorf("expected error containing %q, got %q", expected, f.err.Error())
	}
	return nil
}

func (f *FeatureContext) theResolvedTemplateDirShouldBe(dir string) error {
	if f.selection == nil {
		return fmt.Errorf("no selection available")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, dir, f.selection.TemplateDir, "template dir mismatch")
}

func (f *FeatureContext) theResolvedTemplateMappingShouldBeNil() error {
	if f.selection == nil {
		return fmt.Errorf("no selection available")
	}
	if f.selection.TemplateMapping != nil {
		return fmt.Errorf("expected nil mapping (full-directory copy), got %v", f.selection.TemplateMapping)
	}
	return nil
}

func (f *FeatureContext) allHookDescriptionsShouldBeAtMostNChars(maxLen int) error {
	for name, meta := range catalog.HookMetadata {
		if len(meta.Description) > maxLen {
			return fmt.Errorf("hook %q description is %d chars, exceeds %d", name, len(meta.Description), maxLen)
		}
	}
	return nil
}

func (f *FeatureContext) theHookCategoryNamesShouldContain(name string) error {
	for _, n := range f.categoryNames {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("expected category names to contain %q, got %v", name, f.categoryNames)
}

func (f *FeatureContext) theHookPresetNamesShouldContain(name string) error {
	for _, n := range f.presetNames {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("expected preset names to contain %q, got %v", name, f.presetNames)
}
