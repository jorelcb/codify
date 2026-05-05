package drift_detection

import (
	"os"

	driftdomain "github.com/jorelcb/codify/internal/domain/drift"
	statedomain "github.com/jorelcb/codify/internal/domain/state"
)

// FeatureContext is the per-scenario state for drift_detection.
type FeatureContext struct {
	projectPath string
	outputPath  string
	snapshot    statedomain.State
	report      driftdomain.Report
	err         error
}

// SetupTest is called once before all scenarios.
func (f *FeatureContext) SetupTest() {}

// reset releases any temp dirs from the previous scenario and zeroes state.
func (f *FeatureContext) reset() {
	if f.projectPath != "" {
		_ = os.RemoveAll(f.projectPath)
	}
	if f.outputPath != "" {
		_ = os.RemoveAll(f.outputPath)
	}
	f.projectPath = ""
	f.outputPath = ""
	f.snapshot = statedomain.State{}
	f.report = driftdomain.Report{}
	f.err = nil
}
