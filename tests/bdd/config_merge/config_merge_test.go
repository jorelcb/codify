package config_merge

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/codify/tests/bdd/commons"
)

func TestConfigMergeFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "config_merge",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./config_merge.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
