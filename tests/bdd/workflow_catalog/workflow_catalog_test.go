package workflow_catalog

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/codify/tests/bdd/commons"
)

func TestWorkflowCatalogFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "workflow_catalog",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./workflow_catalog.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
