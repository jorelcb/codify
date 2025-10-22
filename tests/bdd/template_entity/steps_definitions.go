package template_entity

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
	ctx.Step(`^the template system is initialized$`, featureContext.theTemplateSystemIsInitialized)
	ctx.Step(`^I have template data with id "([^"]*)", name "([^"]*)", path "([^"]*)", and content "([^"]*)"$`, featureContext.iHaveTemplateData)
	ctx.Step(`^I have template data with empty id, name "([^"]*)", path "([^"]*)", and content "([^"]*)"$`, featureContext.iHaveTemplateDataWithEmptyID)
	ctx.Step(`^I have template data with id "([^"]*)", empty name, path "([^"]*)", and content "([^"]*)"$`, featureContext.iHaveTemplateDataWithEmptyName)
	ctx.Step(`^I have template data with id "([^"]*)", name "([^"]*)", empty path, and content "([^"]*)"$`, featureContext.iHaveTemplateDataWithEmptyPath)
	ctx.Step(`^I have a valid template$`, featureContext.iHaveAValidTemplate)
	ctx.Step(`^I have a valid template with content "([^"]*)"$`, featureContext.iHaveAValidTemplateWithContent)
	ctx.Step(`^I have a variable with name "([^"]*)", required (true|false), and default "([^"]*)"$`, featureContext.iHaveAVariableWithRequired)
	ctx.Step(`^the variable is already added to the template$`, featureContext.theVariableIsAlreadyAddedToTheTemplate)
	ctx.Step(`^I have a template with corrupted data missing id$`, featureContext.iHaveATemplateWithCorruptedData)

	// ========== When Steps ==========
	ctx.Step(`^I create a new template$`, featureContext.iCreateANewTemplate)
	ctx.Step(`^I add the variable to the template$`, featureContext.iAddTheVariableToTheTemplate)
	ctx.Step(`^I try to add the same variable again$`, featureContext.iTryToAddTheSameVariableAgain)
	ctx.Step(`^I update the template content to "([^"]*)"$`, featureContext.iUpdateTheTemplateContentTo)
	ctx.Step(`^I set metadata with version "([^"]*)", author "([^"]*)", and description "([^"]*)"$`, featureContext.iSetMetadataWith)
	ctx.Step(`^I validate the template$`, featureContext.iValidateTheTemplate)

	// ========== Then Steps ==========
	ctx.Step(`^the template should be created successfully$`, featureContext.theTemplateShouldBeCreatedSuccessfully)
	ctx.Step(`^the template id should be "([^"]*)"$`, featureContext.theTemplateIDShouldBe)
	ctx.Step(`^the template name should be "([^"]*)"$`, featureContext.theTemplateNameShouldBe)
	ctx.Step(`^I should get an error "([^"]*)"$`, featureContext.iShouldGetAnError)
	ctx.Step(`^the variable should be added successfully$`, featureContext.theVariableShouldBeAddedSuccessfully)
	ctx.Step(`^the template should have (\d+) variable$`, featureContext.theTemplateShouldHaveNVariables)
	ctx.Step(`^the template content should be "([^"]*)"$`, featureContext.theTemplateContentShouldBe)
	ctx.Step(`^the template updated timestamp should be recent$`, featureContext.theTemplateUpdatedTimestampShouldBeRecent)
	ctx.Step(`^the template metadata version should be "([^"]*)"$`, featureContext.theTemplateMetadataVersionShouldBe)
	ctx.Step(`^the template metadata author should be "([^"]*)"$`, featureContext.theTemplateMetadataAuthorShouldBe)
	ctx.Step(`^the template metadata description should be "([^"]*)"$`, featureContext.theTemplateMetadataDescriptionShouldBe)
	ctx.Step(`^the validation should pass$`, featureContext.theValidationShouldPass)
	ctx.Step(`^the validation should fail with "([^"]*)"$`, featureContext.theValidationShouldFailWith)
}