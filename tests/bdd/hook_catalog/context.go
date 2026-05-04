package hook_catalog

import (
	"github.com/jorelcb/codify/internal/domain/catalog"
)

// FeatureContext holds the state for hook catalog test scenarios.
type FeatureContext struct {
	category      *catalog.SkillCategory
	selection     *catalog.ResolvedSelection
	target        string
	categoryNames []string
	presetNames   []string
	err           error
}

// SetupTest is called once before all scenarios.
func (f *FeatureContext) SetupTest() {}

// reset clears context state before each scenario.
func (f *FeatureContext) reset() {
	f.category = nil
	f.selection = nil
	f.target = ""
	f.categoryNames = nil
	f.presetNames = nil
	f.err = nil
}
