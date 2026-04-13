package workflow_catalog

import (
	"context"
	"encoding/json"
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

	// ========== Plugin Given Steps ==========
	ctx.Step(`^I parse annotations from the release-cycle template$`, featureContext.iParseAnnotationsFromReleaseCycleTemplate)

	// ========== Plugin When Steps ==========
	ctx.Step(`^I generate a plugin manifest for "([^"]*)"$`, featureContext.iGenerateAPluginManifestFor)
	ctx.Step(`^I generate plugin hooks from the annotations$`, featureContext.iGeneratePluginHooksFromAnnotations)
	ctx.Step(`^I generate a plugin skill for "([^"]*)"$`, featureContext.iGenerateAPluginSkillFor)
	ctx.Step(`^I generate a plugin agent for "([^"]*)" in "([^"]*)"$`, featureContext.iGenerateAPluginAgentForIn)

	// ========== Plugin Then Steps ==========
	ctx.Step(`^the plugin manifest should be valid JSON$`, featureContext.thePluginManifestShouldBeValidJSON)
	ctx.Step(`^the plugin manifest name should be "([^"]*)"$`, featureContext.thePluginManifestNameShouldBe)
	ctx.Step(`^the plugin manifest should have version "([^"]*)"$`, featureContext.thePluginManifestShouldHaveVersion)
	ctx.Step(`^the plugin hooks should be valid JSON$`, featureContext.thePluginHooksShouldBeValidJSON)
	ctx.Step(`^the plugin hooks should contain "([^"]*)"$`, featureContext.thePluginHooksShouldContain)
	ctx.Step(`^the plugin skill should not contain "([^"]*)"$`, featureContext.thePluginSkillShouldNotContain)
	ctx.Step(`^the plugin skill should contain "([^"]*)"$`, featureContext.thePluginSkillShouldContain)
	ctx.Step(`^the plugin agent should contain "([^"]*)"$`, featureContext.thePluginAgentShouldContain)
}

// ========== Given Steps ==========

func (f *FeatureContext) theWorkflowCatalogIsLoaded() error {
	// Catalog is loaded at package init time, nothing to do
	return nil
}

func (f *FeatureContext) iHaveWorkflowCategory(name string) error {
	f.category, f.err = catalog.FindWorkflowCategory(name)
	if f.err != nil {
		return fmt.Errorf("failed to find workflow category %q: %w", name, f.err)
	}
	return nil
}

// ========== When Steps ==========

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

// ========== Then Steps ==========

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

// ========== Plugin Given Steps ==========

func (f *FeatureContext) iParseAnnotationsFromReleaseCycleTemplate() error {
	data, err := root.TemplatesFS.ReadFile(filepath.Join("templates", "en", "workflows", "release_cycle.template"))
	if err != nil {
		return fmt.Errorf("failed to read release_cycle template: %w", err)
	}
	f.annotations = catalog.ParseAnnotations(string(data))
	return nil
}

// ========== Plugin When Steps ==========

func (f *FeatureContext) iGenerateAPluginManifestFor(presetName string) error {
	meta, ok := catalog.WorkflowMetadata[presetName]
	if !ok {
		meta = catalog.WorkflowMeta{Description: "Test workflow"}
	}
	f.pluginManifest = catalog.GeneratePluginManifest(presetName, meta.Description)
	return nil
}

func (f *FeatureContext) iGeneratePluginHooksFromAnnotations() error {
	if f.annotations == nil {
		return fmt.Errorf("no annotations parsed")
	}
	f.pluginHooks = catalog.GeneratePluginHooks(f.annotations)
	return nil
}

func (f *FeatureContext) iGenerateAPluginSkillFor(presetName string) error {
	data, err := root.TemplatesFS.ReadFile(filepath.Join("templates", "en", "workflows", presetName+".template"))
	if err != nil {
		return fmt.Errorf("failed to read template for %s: %w", presetName, err)
	}
	f.pluginSkill = catalog.TransformToPluginSkill(presetName, string(data))
	return nil
}

func (f *FeatureContext) iGenerateAPluginAgentForIn(presetName, locale string) error {
	f.pluginAgent = catalog.GenerateWorkflowAgent(presetName, locale)
	return nil
}

// ========== Plugin Then Steps ==========

func (f *FeatureContext) thePluginManifestShouldBeValidJSON() error {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(f.pluginManifest), &m); err != nil {
		return fmt.Errorf("plugin manifest is not valid JSON: %w", err)
	}
	return nil
}

func (f *FeatureContext) thePluginManifestNameShouldBe(expected string) error {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(f.pluginManifest), &m); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, expected, m["name"], "manifest name mismatch")
}

func (f *FeatureContext) thePluginManifestShouldHaveVersion(expected string) error {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(f.pluginManifest), &m); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, expected, m["version"], "manifest version mismatch")
}

func (f *FeatureContext) thePluginHooksShouldBeValidJSON() error {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(f.pluginHooks), &m); err != nil {
		return fmt.Errorf("plugin hooks is not valid JSON: %w", err)
	}
	return nil
}

func (f *FeatureContext) thePluginHooksShouldContain(substring string) error {
	if !strings.Contains(f.pluginHooks, substring) {
		return fmt.Errorf("expected plugin hooks to contain %q, got:\n%s", substring, f.pluginHooks)
	}
	return nil
}

func (f *FeatureContext) thePluginSkillShouldNotContain(substring string) error {
	if strings.Contains(f.pluginSkill, substring) {
		return fmt.Errorf("expected plugin skill NOT to contain %q", substring)
	}
	return nil
}

func (f *FeatureContext) thePluginSkillShouldContain(substring string) error {
	if !strings.Contains(f.pluginSkill, substring) {
		return fmt.Errorf("expected plugin skill to contain %q", substring)
	}
	return nil
}

func (f *FeatureContext) thePluginAgentShouldContain(substring string) error {
	if !strings.Contains(f.pluginAgent, substring) {
		return fmt.Errorf("expected plugin agent to contain %q, got:\n%s", substring, f.pluginAgent)
	}
	return nil
}
