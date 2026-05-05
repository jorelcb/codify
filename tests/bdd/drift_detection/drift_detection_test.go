package drift_detection

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/codify/tests/bdd/commons"
)

func TestDriftDetectionFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "drift_detection",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./drift_detection.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
