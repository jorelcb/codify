package template_entity

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/ai-context-generator/tests/bdd/commons"
)

func TestTemplateEntityFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "template_entity",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./template_entity.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}