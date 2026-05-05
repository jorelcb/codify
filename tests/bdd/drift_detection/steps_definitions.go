package drift_detection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	driftdomain "github.com/jorelcb/codify/internal/domain/drift"
	infradrift "github.com/jorelcb/codify/internal/infrastructure/drift"
	"github.com/jorelcb/codify/internal/infrastructure/snapshot"
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
	ctx.Step(`^a project with AGENTS\.md and go\.mod$`, featureContext.aProjectWithAgentsAndGoMod)
	ctx.Step(`^a snapshot captured from the current state$`, featureContext.aSnapshotCapturedFromTheCurrentState)

	// When
	ctx.Step(`^the AGENTS\.md file is modified$`, featureContext.theAgentsMdFileIsModified)
	ctx.Step(`^the AGENTS\.md file is removed$`, featureContext.theAgentsMdFileIsRemoved)
	ctx.Step(`^a new artifact "([^"]*)" is added$`, featureContext.aNewArtifactIsAdded)
	ctx.Step(`^the go\.mod content changes$`, featureContext.theGoModContentChanges)
	ctx.Step(`^the go\.mod file is removed$`, featureContext.theGoModFileIsRemoved)
	ctx.Step(`^I detect drift against the snapshot$`, featureContext.iDetectDriftAgainstTheSnapshot)
	ctx.Step(`^I detect drift against the same state$`, featureContext.iDetectDriftAgainstTheSnapshot)

	// Then
	ctx.Step(`^the report should be empty$`, featureContext.theReportShouldBeEmpty)
	ctx.Step(`^the report should not have significant drift$`, featureContext.theReportShouldNotHaveSignificantDrift)
	ctx.Step(`^the report should have significant drift$`, featureContext.theReportShouldHaveSignificantDrift)
	ctx.Step(`^the report should contain a "([^"]*)" entry for "([^"]*)"$`, featureContext.theReportShouldContainEntryFor)
	ctx.Step(`^the entry severity should be "([^"]*)"$`, featureContext.theEntrySeverityShouldBe)
	ctx.Step(`^the report should have at least (\d+) entries$`, featureContext.theReportShouldHaveAtLeastNEntries)
	ctx.Step(`^the report should contain at least one "([^"]*)" entry$`, featureContext.theReportShouldContainAtLeastOneEntryOfSeverity)
}

// --- Given ---

func (f *FeatureContext) aProjectWithAgentsAndGoMod() error {
	var err error
	f.projectPath, err = os.MkdirTemp("", "drift-bdd-proj-*")
	if err != nil {
		return err
	}
	f.outputPath, err = os.MkdirTemp("", "drift-bdd-out-*")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(f.outputPath, "AGENTS.md"), []byte("baseline"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(f.projectPath, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		return err
	}
	return nil
}

func (f *FeatureContext) aSnapshotCapturedFromTheCurrentState() error {
	state, err := snapshot.Build(snapshot.BuildOptions{
		ProjectPath: f.projectPath,
		OutputPath:  f.outputPath,
	})
	if err != nil {
		return err
	}
	f.snapshot = state
	return nil
}

// --- When ---

func (f *FeatureContext) theAgentsMdFileIsModified() error {
	return os.WriteFile(filepath.Join(f.outputPath, "AGENTS.md"), []byte("modified content"), 0o644)
}

func (f *FeatureContext) theAgentsMdFileIsRemoved() error {
	return os.Remove(filepath.Join(f.outputPath, "AGENTS.md"))
}

func (f *FeatureContext) aNewArtifactIsAdded(path string) error {
	full := filepath.Join(f.outputPath, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte("new artifact"), 0o644)
}

func (f *FeatureContext) theGoModContentChanges() error {
	return os.WriteFile(filepath.Join(f.projectPath, "go.mod"), []byte("module test-v2\n"), 0o644)
}

func (f *FeatureContext) theGoModFileIsRemoved() error {
	return os.Remove(filepath.Join(f.projectPath, "go.mod"))
}

func (f *FeatureContext) iDetectDriftAgainstTheSnapshot() error {
	report, err := infradrift.NewDetector().Detect(infradrift.DetectOptions{
		Snapshot:    f.snapshot,
		ProjectPath: f.projectPath,
		OutputPath:  f.outputPath,
	})
	f.report = report
	f.err = err
	return nil
}

// --- Then ---

func (f *FeatureContext) theReportShouldBeEmpty() error {
	return assertions.AssertBool(assert.True, f.report.IsEmpty(), "expected empty report")
}

func (f *FeatureContext) theReportShouldNotHaveSignificantDrift() error {
	return assertions.AssertBool(assert.False, f.report.HasSignificant(), "expected no significant drift")
}

func (f *FeatureContext) theReportShouldHaveSignificantDrift() error {
	return assertions.AssertBool(assert.True, f.report.HasSignificant(), "expected significant drift")
}

func (f *FeatureContext) theReportShouldContainEntryFor(kind, path string) error {
	for _, e := range f.report.Entries {
		if string(e.Kind) == kind && e.Path == path {
			return nil
		}
	}
	return fmt.Errorf("expected entry kind=%q path=%q; got %+v", kind, path, f.report.Entries)
}

func (f *FeatureContext) theEntrySeverityShouldBe(severity string) error {
	if len(f.report.Entries) == 0 {
		return fmt.Errorf("no entries to inspect")
	}
	// El step asume que la última entry agregada es la de interés (el escenario
	// la añade justo antes de este then).
	last := f.report.Entries[len(f.report.Entries)-1]
	if string(last.Severity) != severity {
		return fmt.Errorf("expected severity %q, got %q", severity, last.Severity)
	}
	return nil
}

func (f *FeatureContext) theReportShouldHaveAtLeastNEntries(n int) error {
	if len(f.report.Entries) < n {
		return fmt.Errorf("expected at least %d entries, got %d", n, len(f.report.Entries))
	}
	return nil
}

func (f *FeatureContext) theReportShouldContainAtLeastOneEntryOfSeverity(severity string) error {
	want := driftdomain.Severity(severity)
	for _, e := range f.report.Entries {
		if e.Severity == want {
			return nil
		}
	}
	return fmt.Errorf("expected at least one %q-severity entry; got %+v", severity, f.report.Entries)
}
