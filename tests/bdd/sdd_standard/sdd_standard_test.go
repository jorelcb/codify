package sdd_standard

import (
	"testing"

	"github.com/cucumber/godog"
	"github.com/jorelcb/codify/tests/bdd/commons"
)

// TestSDDStandardFeature corre los scenarios de sdd_standard.feature.
// Valida el contrato del port SpecStandard end-to-end para OpenSpec y
// Spec-Kit (ADR-0011), incluyendo resolution precedence y disponibilidad
// de templates en el embedded FS.
func TestSDDStandardFeature(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "sdd_standard",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              commons.Options("./sdd_standard.feature"),
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
