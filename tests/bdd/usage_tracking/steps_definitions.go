package usage_tracking

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
	infrausage "github.com/jorelcb/codify/internal/infrastructure/usage"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

var featureContext = new(FeatureContext)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// Given
	ctx.Step(`^an empty home directory$`, featureContext.anEmptyHomeDirectory)
	ctx.Step(`^the env variable "([^"]*)" is set to "([^"]*)"$`, featureContext.theEnvVariableIsSetTo)

	// When
	ctx.Step(`^I record a usage entry with model "([^"]*)", (\d+) input tokens, and (\d+) output tokens$`, featureContext.iRecordAUsageEntry)
	ctx.Step(`^I read the global usage log$`, featureContext.iReadTheGlobalUsageLog)
	ctx.Step(`^I compute cost for model "([^"]*)" with (\d+) input tokens and (\d+) output tokens$`, featureContext.iComputeCostFor)
	ctx.Step(`^I check if tracking is disabled$`, featureContext.iCheckIfTrackingIsDisabled)
	ctx.Step(`^I check tracking with the no-tracking flag enabled$`, featureContext.iCheckTrackingWithFlagEnabled)

	// Then
	ctx.Step(`^the log should contain (\d+) entr(?:y|ies)$`, featureContext.theLogShouldContainNEntries)
	ctx.Step(`^the totals should report (\d+) total tokens$`, featureContext.theTotalsShouldReportNTotalTokens)
	ctx.Step(`^the totals cost should be greater than 0$`, featureContext.theTotalsCostShouldBeGreaterThan0)
	ctx.Step(`^the cost should be (\d+) cents$`, featureContext.theCostShouldBeNCents)
	ctx.Step(`^tracking should be reported as disabled$`, featureContext.trackingShouldBeReportedAsDisabled)
	ctx.Step(`^the global usage file should not exist$`, featureContext.theGlobalUsageFileShouldNotExist)
	ctx.Step(`^the totals call count should be (\d+)$`, featureContext.theTotalsCallCountShouldBe)
}

// --- Given ---

func (f *FeatureContext) anEmptyHomeDirectory() error { return nil }

func (f *FeatureContext) theEnvVariableIsSetTo(name, value string) error {
	return os.Setenv(name, value)
}

// --- When ---

func (f *FeatureContext) iRecordAUsageEntry(model string, inputTokens, outputTokens int) error {
	rec := infrausage.NewRecorder(f.flagDisabled)
	rec.Record(usagedomain.Entry{
		Command:      "test",
		Provider:     "test-provider",
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	})
	return nil
}

func (f *FeatureContext) iReadTheGlobalUsageLog() error {
	path, err := infrausage.UserUsagePath()
	if err != nil {
		return err
	}
	log, err := infrausage.NewRepository().Load(path)
	if err != nil {
		return err
	}
	f.log = log
	return nil
}

func (f *FeatureContext) iComputeCostFor(model string, inputTokens, outputTokens int) error {
	f.computedCost = usagedomain.CostCents(model, inputTokens, outputTokens, 0, 0)
	return nil
}

func (f *FeatureContext) iCheckIfTrackingIsDisabled() error {
	f.trackingDisabled = infrausage.TrackingDisabled(f.flagDisabled)
	return nil
}

func (f *FeatureContext) iCheckTrackingWithFlagEnabled() error {
	f.flagDisabled = true
	f.trackingDisabled = infrausage.TrackingDisabled(true)
	return nil
}

// --- Then ---

func (f *FeatureContext) theLogShouldContainNEntries(n int) error {
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.log.Entries), "entry count")
}

func (f *FeatureContext) theTotalsShouldReportNTotalTokens(n int) error {
	got := f.log.Totals.InputTokens + f.log.Totals.OutputTokens
	return assertions.AssertExpectedAndActual(assert.Equal, n, got, "total tokens")
}

func (f *FeatureContext) theTotalsCostShouldBeGreaterThan0() error {
	if f.log.Totals.CostUSDCents <= 0 {
		return fmt.Errorf("expected cost > 0, got %d", f.log.Totals.CostUSDCents)
	}
	return nil
}

func (f *FeatureContext) theCostShouldBeNCents(n int) error {
	return assertions.AssertExpectedAndActual(assert.Equal, n, f.computedCost, "computed cost")
}

func (f *FeatureContext) trackingShouldBeReportedAsDisabled() error {
	return assertions.AssertBool(assert.True, f.trackingDisabled, "expected tracking disabled")
}

func (f *FeatureContext) theGlobalUsageFileShouldNotExist() error {
	path, err := infrausage.UserUsagePath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("expected file %s to not exist (tracking should be disabled)", path)
	}
	return nil
}

func (f *FeatureContext) theTotalsCallCountShouldBe(n int) error {
	return assertions.AssertExpectedAndActual(assert.Equal, n, f.log.Totals.Calls, "calls")
}
