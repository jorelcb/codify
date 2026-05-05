package audit_rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	domain "github.com/jorelcb/codify/internal/domain/audit"
	infraaudit "github.com/jorelcb/codify/internal/infrastructure/audit"
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

	ctx.Step(`^I audit a commit with header "([^"]*)"$`, featureContext.iAuditACommitWithHeader)
	ctx.Step(`^I audit a commit with a (\d+)-character header$`, featureContext.iAuditACommitWithNCharHeader)
	ctx.Step(`^I check if a commit with (\d+) parents? is a merge commit$`, featureContext.iCheckIfACommitWithNParentsIsMerge)
	ctx.Step(`^I parse the LLM response '(.+)'$`, featureContext.iParseTheLLMResponse)
	ctx.Step(`^I parse the LLM response "([^"]+)"$`, featureContext.iParseTheLLMResponse)
	ctx.Step(`^I parse the LLM response wrapped in JSON code fences$`, featureContext.iParseTheLLMResponseWrappedInFences)

	ctx.Step(`^the audit findings should be empty$`, featureContext.theAuditFindingsShouldBeEmpty)
	ctx.Step(`^the audit should contain a "([^"]*)" finding$`, featureContext.theAuditShouldContainFinding)
	ctx.Step(`^the finding severity should be "([^"]*)"$`, featureContext.theFindingSeverityShouldBe)
	ctx.Step(`^it should be reported as a merge commit$`, featureContext.itShouldBeReportedAsMergeCommit)
	ctx.Step(`^it should not be reported as a merge commit$`, featureContext.itShouldNotBeReportedAsMergeCommit)
	ctx.Step(`^the parsed findings should contain (\d+) entr(?:y|ies)$`, featureContext.theParsedFindingsShouldContainNEntries)
	ctx.Step(`^the parsed finding should have heuristic flag set$`, featureContext.theParsedFindingShouldHaveHeuristicFlagSet)
	ctx.Step(`^the parsed finding kind should be "([^"]*)"$`, featureContext.theParsedFindingKindShouldBe)
	ctx.Step(`^the parsed finding severity should be "([^"]*)"$`, featureContext.theParsedFindingSeverityShouldBe)
	ctx.Step(`^the parser should report an error$`, featureContext.theParserShouldReportAnError)
}

func (f *FeatureContext) iParseTheLLMResponse(raw string) error {
	f.parsedFindings, f.parseErr = infraaudit.ParseLLMFindings(raw)
	return nil
}

func (f *FeatureContext) iParseTheLLMResponseWrappedInFences() error {
	raw := "```json\n[{\"commit_sha\":\"abc\",\"severity\":\"minor\",\"detail\":\"x\"}]\n```"
	f.parsedFindings, f.parseErr = infraaudit.ParseLLMFindings(raw)
	return nil
}

func (f *FeatureContext) theParsedFindingsShouldContainNEntries(n int) error {
	if len(f.parsedFindings) != n {
		return fmt.Errorf("expected %d parsed findings, got %d", n, len(f.parsedFindings))
	}
	return nil
}

func (f *FeatureContext) theParsedFindingShouldHaveHeuristicFlagSet() error {
	if len(f.parsedFindings) == 0 {
		return fmt.Errorf("no parsed findings to check heuristic flag")
	}
	if !f.parsedFindings[0].Heuristic {
		return fmt.Errorf("expected heuristic=true on first parsed finding")
	}
	return nil
}

func (f *FeatureContext) theParsedFindingKindShouldBe(kind string) error {
	if len(f.parsedFindings) == 0 {
		return fmt.Errorf("no parsed findings to check kind")
	}
	if string(f.parsedFindings[0].Kind) != kind {
		return fmt.Errorf("got kind %q, want %q", f.parsedFindings[0].Kind, kind)
	}
	return nil
}

func (f *FeatureContext) theParsedFindingSeverityShouldBe(severity string) error {
	if len(f.parsedFindings) == 0 {
		return fmt.Errorf("no parsed findings to check severity")
	}
	if string(f.parsedFindings[0].Severity) != severity {
		return fmt.Errorf("got severity %q, want %q", f.parsedFindings[0].Severity, severity)
	}
	return nil
}

func (f *FeatureContext) theParserShouldReportAnError() error {
	if f.parseErr == nil {
		return fmt.Errorf("expected parser error, got nil")
	}
	return nil
}

func (f *FeatureContext) iAuditACommitWithHeader(header string) error {
	f.findings = infraaudit.AuditCommitMessageForTest("test-sha", header)
	return nil
}

func (f *FeatureContext) iAuditACommitWithNCharHeader(n int) error {
	header := "feat(api): " + strings.Repeat("a", n-len("feat(api): "))
	if len(header) < n {
		header += strings.Repeat("a", n-len(header))
	}
	f.findings = infraaudit.AuditCommitMessageForTest("test-sha", header)
	return nil
}

func (f *FeatureContext) iCheckIfACommitWithNParentsIsMerge(n int) error {
	f.isMerge = infraaudit.IsMergeCommitForTest(n)
	f.isMergeSet = true
	return nil
}

func (f *FeatureContext) theAuditFindingsShouldBeEmpty() error {
	if len(f.findings) != 0 {
		return fmt.Errorf("expected no findings, got: %+v", f.findings)
	}
	return nil
}

func (f *FeatureContext) theAuditShouldContainFinding(kind string) error {
	for _, fnd := range f.findings {
		if string(fnd.Kind) == kind {
			return nil
		}
	}
	return fmt.Errorf("expected finding kind=%q, got: %+v", kind, f.findings)
}

func (f *FeatureContext) theFindingSeverityShouldBe(severity string) error {
	if len(f.findings) == 0 {
		return fmt.Errorf("no findings to check severity on")
	}
	want := domain.Severity(severity)
	for _, fnd := range f.findings {
		if fnd.Severity == want {
			return nil
		}
	}
	return fmt.Errorf("expected at least one %q-severity finding; got: %+v", severity, f.findings)
}

func (f *FeatureContext) itShouldBeReportedAsMergeCommit() error {
	return assertions.AssertBool(assert.True, f.isMerge, "expected merge commit")
}

func (f *FeatureContext) itShouldNotBeReportedAsMergeCommit() error {
	return assertions.AssertBool(assert.False, f.isMerge, "expected not merge commit")
}
