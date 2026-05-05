package architecture_catalog

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/codify/tests/bdd/commons"
)

func TestArchitectureCatalogFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "architecture_catalog",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./architecture_catalog.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
