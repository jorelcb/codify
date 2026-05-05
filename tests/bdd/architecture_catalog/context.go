package architecture_catalog

import (
	"github.com/jorelcb/codify/internal/domain/catalog"
)

// FeatureContext holds the state for architecture catalog test scenarios.
type FeatureContext struct {
	category    *catalog.SkillCategory
	selection   *catalog.ResolvedSelection
	presetNames []string
	legacyMap   [2]string
	legacyOk    bool
	err         error
}

// SetupTest initializes test data (called once before all scenarios).
func (f *FeatureContext) SetupTest() {}

// reset clears context state before each scenario.
func (f *FeatureContext) reset() {
	f.category = nil
	f.selection = nil
	f.presetNames = nil
	f.legacyMap = [2]string{}
	f.legacyOk = false
	f.err = nil
}
