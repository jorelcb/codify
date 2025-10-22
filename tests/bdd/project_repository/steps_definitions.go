package project_repository

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
	ctx.Step(`^an empty project repository$`, featureContext.anEmptyProjectRepository)
	ctx.Step(`^I have a project with id "([^"]*)"$`, featureContext.iHaveAProjectWithID)
	ctx.Step(`^I have a project with name "([^"]*)"$`, featureContext.iHaveAProjectWithName)
	ctx.Step(`^I have a project with name "([^"]*)" and id "([^"]*)"$`, featureContext.iHaveAProjectWithNameAndID)
	ctx.Step(`^I have an invalid project$`, featureContext.iHaveAnInvalidProject)
	ctx.Step(`^I have (\d+) projects$`, featureContext.iHaveNProjects)
	ctx.Step(`^I have (\d+) projects saved$`, featureContext.iHaveNProjectsSaved)

	// ========== When Steps ==========
	ctx.Step(`^I save the project$`, featureContext.iSaveTheProject)
	ctx.Step(`^I save the first project$`, featureContext.iSaveTheFirstProject)
	ctx.Step(`^I save the second project$`, featureContext.iSaveTheSecondProject)
	ctx.Step(`^I save all projects$`, featureContext.iSaveAllProjects)
	ctx.Step(`^I save all projects concurrently$`, featureContext.iSaveAllProjectsConcurrently)
	ctx.Step(`^I try to save the project$`, featureContext.iTryToSaveTheProject)
	ctx.Step(`^I try to save a nil project$`, featureContext.iTryToSaveANilProject)
	ctx.Step(`^I retrieve project by id "([^"]*)"$`, featureContext.iRetrieveProjectByID)
	ctx.Step(`^I retrieve project by name "([^"]*)"$`, featureContext.iRetrieveProjectByName)
	ctx.Step(`^I try to retrieve project by id "([^"]*)"$`, featureContext.iTryToRetrieveProjectByID)
	ctx.Step(`^I try to retrieve project by name "([^"]*)"$`, featureContext.iTryToRetrieveProjectByName)
	ctx.Step(`^I retrieve all projects$`, featureContext.iRetrieveAllProjects)
	ctx.Step(`^I add a capability to the project$`, featureContext.iAddACapabilityToTheProject)
	ctx.Step(`^I save the project again$`, featureContext.iSaveTheProjectAgain)
	ctx.Step(`^I delete project with id "([^"]*)"$`, featureContext.iDeleteProjectWithID)
	ctx.Step(`^I try to delete project with id "([^"]*)"$`, featureContext.iTryToDeleteProjectWithID)
	ctx.Step(`^I clear the repository$`, featureContext.iClearTheRepository)

	// ========== Then Steps ==========
	ctx.Step(`^the project should be saved successfully$`, featureContext.theProjectShouldBeSavedSuccessfully)
	ctx.Step(`^I should receive the project$`, featureContext.iShouldReceiveTheProject)
	ctx.Step(`^I should get an error containing "([^"]*)"$`, featureContext.iShouldGetAnErrorContaining)
	ctx.Step(`^all (\d+) projects should be saved$`, featureContext.allNProjectsShouldBeSaved)
	ctx.Step(`^I should receive (\d+) projects$`, featureContext.iShouldReceiveNProjects)
	ctx.Step(`^the project should have the new capability$`, featureContext.theProjectShouldHaveTheNewCapability)
	ctx.Step(`^the project should be deleted$`, featureContext.theProjectShouldBeDeleted)
	ctx.Step(`^project with id "([^"]*)" should exist$`, featureContext.projectWithIDShouldExist)
	ctx.Step(`^project with id "([^"]*)" should not exist$`, featureContext.projectWithIDShouldNotExist)
	ctx.Step(`^project with name "([^"]*)" should exist$`, featureContext.projectWithNameShouldExist)
	ctx.Step(`^project with name "([^"]*)" should not exist$`, featureContext.projectWithNameShouldNotExist)
	ctx.Step(`^the second project should overwrite by name$`, featureContext.theSecondProjectShouldOverwriteByName)
	ctx.Step(`^I should get a validation error$`, featureContext.iShouldGetAValidationError)
	ctx.Step(`^the repository count should be (\d+)$`, featureContext.theRepositoryCountShouldBe)
	ctx.Step(`^the repository should be empty$`, featureContext.theRepositoryShouldBeEmpty)
}
