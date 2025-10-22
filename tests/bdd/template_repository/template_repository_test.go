package template_repository

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/ai-context-generator/tests/bdd/commons"
)

func TestTemplateRepositoryFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "template_repository",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./template_repository.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}