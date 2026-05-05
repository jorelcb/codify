package audit_rules

import (
	domain "github.com/jorelcb/codify/internal/domain/audit"
)

// FeatureContext is the per-scenario state for audit_rules.
type FeatureContext struct {
	findings       []domain.Finding
	isMerge        bool
	isMergeSet     bool
	parsedFindings []domain.Finding
	parseErr       error
}

func (f *FeatureContext) SetupTest() {}

func (f *FeatureContext) reset() {
	f.findings = nil
	f.isMerge = false
	f.isMergeSet = false
	f.parsedFindings = nil
	f.parseErr = nil
}
