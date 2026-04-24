package workflow_catalog

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/domain/catalog"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

// featureContext is the singleton instance for this feature
var featureContext = new(FeatureContext)

// InitializeTestSuite is called once before all scenarios
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

// InitializeScenario registers step definitions for each scenario
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// ========== Given Steps ==========
	ctx.Step(`^the workflow catalog is loaded$`, featureContext.theWorkflowCatalogIsLoaded)
	ctx.Step(`^I have workflow category "([^"]*)"$`, featureContext.iHaveWorkflowCategory)

	// ========== When Steps ==========
	ctx.Step(`^I look up workflow category "([^"]*)"$`, featureContext.iLookUpWorkflowCategory)
	ctx.Step(`^I resolve workflow preset "([^"]*)"$`, featureContext.iResolveWorkflowPreset)
	ctx.Step(`^I generate workflow frontmatter for "([^"]*)" targeting "([^"]*)"$`, featureContext.iGenerateWorkflowFrontmatterForTarget)
	ctx.Step(`^I generate workflow frontmatter for "([^"]*)"$`, featureContext.iGenerateWorkflowFrontmatterFor)
	ctx.Step(`^I retrieve workflow category names$`, featureContext.iRetrieveWorkflowCategoryNames)
	ctx.Step(`^I strip annotations from the feature-development template$`, featureContext.iStripAnnotationsFromFeatureDevelopmentTemplate)

	// ========== Then Steps ==========
	ctx.Step(`^I should find a workflow category with name "([^"]*)"$`, featureContext.iShouldFindAWorkflowCategoryWithName)
	ctx.Step(`^the workflow category should have (\d+) options$`, featureContext.theWorkflowCategoryShouldHaveNOptions)
	ctx.Step(`^I should get a workflow catalog error containing "([^"]*)"$`, featureContext.iShouldGetAWorkflowCatalogErrorContaining)
	ctx.Step(`^the resolved template directory should be "([^"]*)"$`, featureContext.theResolvedTemplateDirShouldBe)
	ctx.Step(`^the resolved mapping should have (\d+) entr(?:y|ies)$`, featureContext.theResolvedMappingShouldHaveNEntries)
	ctx.Step(`^the frontmatter should start with "([^"]*)"$`, featureContext.theFrontmatterShouldStartWith)
	ctx.Step(`^the frontmatter should contain "([^"]*)"$`, featureContext.theFrontmatterShouldContain)
	ctx.Step(`^the frontmatter should not contain "([^"]*)"$`, featureContext.theFrontmatterShouldNotContain)
	ctx.Step(`^the frontmatter should end with "([^"]*)"$`, featureContext.theFrontmatterShouldEndWith)
	ctx.Step(`^all workflow descriptions should be at most (\d+) characters$`, featureContext.allWorkflowDescriptionsShouldBeAtMostNChars)
	ctx.Step(`^the workflow category names should contain "([^"]*)"$`, featureContext.theWorkflowCategoryNamesShouldContain)
	ctx.Step(`^the stripped content should not contain "([^"]*)"$`, featureContext.theStrippedContentShouldNotContain)
	ctx.Step(`^the stripped content should contain "([^"]*)"$`, featureContext.theStrippedContentShouldContain)
}

func (f *FeatureContext) theWorkflowCatalogIsLoaded() error { return nil }

func (f *FeatureContext) iHaveWorkflowCategory(name string) error {
	f.category, f.err = catalog.FindWorkflowCategory(name)
	if f.err != nil {
		return fmt.Errorf("failed to find workflow category %q: %w", name, f.err)
	}
	return nil
}

func (f *FeatureContext) iLookUpWorkflowCategory(name string) error {
	f.category, f.err = catalog.FindWorkflowCategory(name)
	return nil
}

func (f *FeatureContext) iResolveWorkflowPreset(preset string) error {
	if f.category == nil {
		return fmt.Errorf("no workflow category set")
	}
	f.selection, f.err = f.category.Resolve(preset)
	return nil
}

func (f *FeatureContext) iGenerateWorkflowFrontmatterFor(guideName string) error {
	target := f.target
	if target == "" {
		target = "antigravity"
	}
	f.frontmatter = catalog.GenerateWorkflowFrontmatter(guideName, target)
	return nil
}

func (f *FeatureContext) iGenerateWorkflowFrontmatterForTarget(guideName, target string) error {
	f.target = target
	f.frontmatter = catalog.GenerateWorkflowFrontmatter(guideName, target)
	return nil
}

func (f *FeatureContext) iRetrieveWorkflowCategoryNames() error {
	f.categoryNames = catalog.WorkflowCategoryNames()
	return nil
}

func (f *FeatureContext) iStripAnnotationsFromFeatureDevelopmentTemplate() error {
	data, err := root.TemplatesFS.ReadFile(filepath.Join("templates", "en", "workflows", "feature_development.template"))
	if err != nil {
		return fmt.Errorf("failed to read feature_development template: %w", err)
	}
	f.strippedContent = catalog.StripAnnotationLines(string(data))
	return nil
}

func (f *FeatureContext) iShouldFindAWorkflowCategoryWithName(name string) error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	if err := assertions.AssertActual(assert.NotNil, f.category, "expected category to be found"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, name, f.category.Name, "category name mismatch")
}

func (f *FeatureContext) theWorkflowCategoryShouldHaveNOptions(n int) error {
	if f.category == nil {
		return fmt.Errorf("no category available")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.category.Options), "option count mismatch")
}

func (f *FeatureContext) iShouldGetAWorkflowCatalogErrorContaining(expected string) error {
	if f.err == nil {
		return fmt.Errorf("expected error containing %q, but got no error", expected)
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

func (f *FeatureContext) theResolvedMappingShouldHaveNEntries(n int) error {
	if f.selection == nil {
		return fmt.Errorf("no selection available")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.selection.TemplateMapping), "mapping count mismatch")
}

func (f *FeatureContext) theFrontmatterShouldStartWith(prefix string) error {
	if !strings.HasPrefix(f.frontmatter, prefix) {
		return fmt.Errorf("expected frontmatter to start with %q, got %q", prefix, f.frontmatter[:min(len(f.frontmatter), 20)])
	}
	return nil
}

func (f *FeatureContext) theFrontmatterShouldContain(substring string) error {
	if !strings.Contains(f.frontmatter, substring) {
		return fmt.Errorf("expected frontmatter to contain %q, got %q", substring, f.frontmatter)
	}
	return nil
}

func (f *FeatureContext) theFrontmatterShouldNotContain(substring string) error {
	if strings.Contains(f.frontmatter, substring) {
		return fmt.Errorf("expected frontmatter NOT to contain %q, but it does: %q", substring, f.frontmatter)
	}
	return nil
}

func (f *FeatureContext) theFrontmatterShouldEndWith(suffix string) error {
	trimmed := strings.TrimRight(f.frontmatter, "\n")
	if !strings.HasSuffix(trimmed, suffix) {
		return fmt.Errorf("expected frontmatter to end with %q, got %q", suffix, trimmed[max(0, len(trimmed)-20):])
	}
	return nil
}

func (f *FeatureContext) allWorkflowDescriptionsShouldBeAtMostNChars(maxLen int) error {
	for name, meta := range catalog.WorkflowMetadata {
		if len(meta.Description) > maxLen {
			return fmt.Errorf("workflow %q description is %d chars, exceeds %d", name, len(meta.Description), maxLen)
		}
	}
	return nil
}

func (f *FeatureContext) theWorkflowCategoryNamesShouldContain(name string) error {
	for _, n := range f.categoryNames {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("expected category names to contain %q, got %v", name, f.categoryNames)
}

func (f *FeatureContext) theStrippedContentShouldNotContain(substring string) error {
	if strings.Contains(f.strippedContent, substring) {
		return fmt.Errorf("expected stripped content NOT to contain %q", substring)
	}
	return nil
}

func (f *FeatureContext) theStrippedContentShouldContain(substring string) error {
	if !strings.Contains(f.strippedContent, substring) {
		return fmt.Errorf("expected stripped content to contain %q", substring)
	}
	return nil
}
