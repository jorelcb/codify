package usage_tracking

import (
	"os"

	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
)

// FeatureContext is the per-scenario state for usage_tracking.
type FeatureContext struct {
	originalHome string
	tempHome     string

	log              usagedomain.Log
	computedCost     int
	trackingDisabled bool
	flagDisabled     bool
}

func (f *FeatureContext) SetupTest() {}

func (f *FeatureContext) reset() {
	if f.originalHome != "" {
		_ = os.Setenv("HOME", f.originalHome)
	}
	if f.tempHome != "" {
		_ = os.RemoveAll(f.tempHome)
	}
	// Tests anteriores pueden haber dejado el env var seteado; limpiarlo
	_ = os.Unsetenv("CODIFY_NO_USAGE_TRACKING")
	f.originalHome = os.Getenv("HOME")
	f.tempHome, _ = os.MkdirTemp("", "codify-bdd-usage-*")
	_ = os.Setenv("HOME", f.tempHome)
	f.log = usagedomain.Log{}
	f.computedCost = 0
	f.trackingDisabled = false
	f.flagDisabled = false
}
