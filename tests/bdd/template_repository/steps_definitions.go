package template_repository

import (
	"context"

	"github.com/cucumber/godog"
)

// featureContext is the singleton instance for this feature
var featureContext = new(FeatureContext)

// InitializeTestSuite is called once before all scenarios
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

// InitializeScenario registers step definitions for each scenario
func InitializeScenario(ctx *godog.ScenarioContext) {
	// Reset context state before each scenario
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// ========== Given Steps ==========
	ctx.Step(`^an empty template repository$`, featureContext.anEmptyTemplateRepository)
	ctx.Step(`^I have a template with id "([^"]*)"$`, featureContext.iHaveATemplateWithID)
	ctx.Step(`^I have a template with path "([^"]*)"$`, featureContext.iHaveATemplateWithPath)
	ctx.Step(`^I have a template with id "([^"]*)" and tag "([^"]*)"$`, featureContext.iHaveATemplateWithIDAndTag)
	ctx.Step(`^I have an invalid template$`, featureContext.iHaveAnInvalidTemplate)
	ctx.Step(`^I have (\d+) templates$`, featureContext.iHaveNTemplates)
	ctx.Step(`^I have (\d+) templates saved$`, featureContext.iHaveNTemplatesSaved)

	// ========== When Steps ==========
	ctx.Step(`^I save the template$`, featureContext.iSaveTheTemplate)
	ctx.Step(`^I save all templates$`, featureContext.iSaveAllTemplates)
	ctx.Step(`^I save all templates concurrently$`, featureContext.iSaveAllTemplatesConcurrently)
	ctx.Step(`^I try to save the template$`, featureContext.iTryToSaveTheTemplate)
	ctx.Step(`^I try to save a nil template$`, featureContext.iTryToSaveANilTemplate)
	ctx.Step(`^I retrieve template by id "([^"]*)"$`, featureContext.iRetrieveTemplateByID)
	ctx.Step(`^I retrieve template by path "([^"]*)"$`, featureContext.iRetrieveTemplateByPath)
	ctx.Step(`^I try to retrieve template by id "([^"]*)"$`, featureContext.iTryToRetrieveTemplateByID)
	ctx.Step(`^I try to retrieve template by path "([^"]*)"$`, featureContext.iTryToRetrieveTemplateByPath)
	ctx.Step(`^I retrieve all templates$`, featureContext.iRetrieveAllTemplates)
	ctx.Step(`^I find templates by tag "([^"]*)"$`, featureContext.iFindTemplatesByTag)
	ctx.Step(`^I update the template content$`, featureContext.iUpdateTheTemplateContent)
	ctx.Step(`^I save the template again$`, featureContext.iSaveTheTemplateAgain)
	ctx.Step(`^I delete template with id "([^"]*)"$`, featureContext.iDeleteTemplateWithID)
	ctx.Step(`^I try to delete template with id "([^"]*)"$`, featureContext.iTryToDeleteTemplateWithID)
	ctx.Step(`^I clear the repository$`, featureContext.iClearTheRepository)

	// ========== Then Steps ==========
	ctx.Step(`^the template should be saved successfully$`, featureContext.theTemplateShouldBeSavedSuccessfully)
	ctx.Step(`^I should receive the template$`, featureContext.iShouldReceiveTheTemplate)
	ctx.Step(`^I should get an error containing "([^"]*)"$`, featureContext.iShouldGetAnErrorContaining)
	ctx.Step(`^all (\d+) templates should be saved$`, featureContext.allNTemplatesShouldBeSaved)
	ctx.Step(`^I should receive (\d+) templates$`, featureContext.iShouldReceiveNTemplates)
	ctx.Step(`^the template should have updated content$`, featureContext.theTemplateShouldHaveUpdatedContent)
	ctx.Step(`^the template should be deleted$`, featureContext.theTemplateShouldBeDeleted)
	ctx.Step(`^template with id "([^"]*)" should exist$`, featureContext.templateWithIDShouldExist)
	ctx.Step(`^template with id "([^"]*)" should not exist$`, featureContext.templateWithIDShouldNotExist)
	ctx.Step(`^I should get a validation error$`, featureContext.iShouldGetAValidationError)
	ctx.Step(`^the repository count should be (\d+)$`, featureContext.theRepositoryCountShouldBe)
	ctx.Step(`^the repository should be empty$`, featureContext.theRepositoryShouldBeEmpty)
}