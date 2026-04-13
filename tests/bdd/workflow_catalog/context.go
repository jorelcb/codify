package workflow_catalog

import (
	"github.com/jorelcb/codify/internal/domain/catalog"
)

// FeatureContext holds the state for workflow catalog test scenarios
type FeatureContext struct {
	category      *catalog.SkillCategory
	selection     *catalog.ResolvedSelection
	frontmatter   string
	target        string
	categoryNames []string
	err           error

	// Plugin generation fields
	pluginManifest string
	pluginHooks    string
	pluginSkill    string
	pluginAgent    string
	annotations    []catalog.AnnotationMeta
}

// SetupTest initializes test data (called once before all scenarios)
func (f *FeatureContext) SetupTest() {}

// reset clears context state before each scenario
func (f *FeatureContext) reset() {
	f.category = nil
	f.selection = nil
	f.frontmatter = ""
	f.target = ""
	f.categoryNames = nil
	f.err = nil
	f.pluginManifest = ""
	f.pluginHooks = ""
	f.pluginSkill = ""
	f.pluginAgent = ""
	f.annotations = nil
}
