package usage_tracking

import (
	"os"

	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
)

// FeatureContext is the per-scenario state for usage_tracking.
type FeatureContext struct {
	originalHome string
	originalCwd  string
	tempHome     string
	tempCwd      string

	log              usagedomain.Log
	computedCost     int
	trackingDisabled bool
	flagDisabled     bool
}

func (f *FeatureContext) SetupTest() {}

// reset re-routes both HOME and cwd to fresh temp dirs so each scenario is
// isolated. Without the chdir, the recorder writes a project-level
// .codify/usage.json into the test source tree (which then gets committed
// by accident).
func (f *FeatureContext) reset() {
	if f.originalHome != "" {
		_ = os.Setenv("HOME", f.originalHome)
	}
	if f.originalCwd != "" {
		_ = os.Chdir(f.originalCwd)
	}
	if f.tempHome != "" {
		_ = os.RemoveAll(f.tempHome)
	}
	if f.tempCwd != "" {
		_ = os.RemoveAll(f.tempCwd)
	}
	// Tests anteriores pueden haber dejado el env var seteado; limpiarlo
	_ = os.Unsetenv("CODIFY_NO_USAGE_TRACKING")
	f.originalHome = os.Getenv("HOME")
	f.originalCwd, _ = os.Getwd()
	f.tempHome, _ = os.MkdirTemp("", "codify-bdd-usage-home-*")
	f.tempCwd, _ = os.MkdirTemp("", "codify-bdd-usage-cwd-*")
	_ = os.Setenv("HOME", f.tempHome)
	_ = os.Chdir(f.tempCwd)
	f.log = usagedomain.Log{}
	f.computedCost = 0
	f.trackingDisabled = false
	f.flagDisabled = false
}
