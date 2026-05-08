package sdd_standard

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/sdd"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

// featureContext es el singleton compartido entre steps de un scenario.
// godog hace reset por scenario via el Before hook abajo.
var featureContext = new(FeatureContext)

// InitializeTestSuite corre una sola vez antes del set completo de scenarios.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

// InitializeScenario registra los step bindings y el reset por-scenario.
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// ========== Given Steps ==========
	ctx.Step(`^the SDD registry is loaded with default adapters$`, featureContext.theRegistryIsLoaded)

	// ========== When Steps ==========
	ctx.Step(`^I look up SDD standard "([^"]*)"$`, featureContext.iLookUpStandard)
	ctx.Step(`^I resolve with flag "([^"]*)" project "([^"]*)" user "([^"]*)"$`, featureContext.iResolveWithLayers)

	// ========== Then Steps — adapter contract ==========
	ctx.Step(`^the lookup should succeed$`, featureContext.theLookupShouldSucceed)
	ctx.Step(`^the standard's display name should be "([^"]*)"$`, featureContext.theDisplayNameShouldBe)
	ctx.Step(`^the standard's template directory should be "([^"]*)"$`, featureContext.theTemplateDirShouldBe)
	ctx.Step(`^the standard's output layout should be "([^"]*)"$`, featureContext.theOutputLayoutShouldBe)
	ctx.Step(`^the bootstrap artifacts should include exactly these files:$`, featureContext.theArtifactsShouldIncludeExactly)
	ctx.Step(`^every bootstrap artifact should be marked required$`, featureContext.everyArtifactRequired)
	ctx.Step(`^every required artifact should have a lowercase file name$`, featureContext.everyRequiredArtifactLowercase)
	ctx.Step(`^no bootstrap artifact should have a name containing "([^"]*)"$`, featureContext.noArtifactContaining)
	ctx.Step(`^the required artifact files should be exactly:$`, featureContext.theRequiredFilesShouldBe)
	ctx.Step(`^the optional artifact files should be exactly:$`, featureContext.theOptionalFilesShouldBe)
	ctx.Step(`^the lifecycle workflow IDs should be:$`, featureContext.theLifecycleWorkflowIDsShouldBe)
	ctx.Step(`^the system prompt hints in "([^"]*)" should mention "([^"]*)"$`, featureContext.theHintsInLocaleShouldMention)

	// ========== Then Steps — resolution precedence ==========
	ctx.Step(`^the resolved standard ID should be "([^"]*)"$`, featureContext.theResolvedIDShouldBe)
	ctx.Step(`^resolution should fail with error containing "([^"]*)"$`, featureContext.resolutionShouldFailWith)

	// ========== Then Steps — embedded FS ==========
	ctx.Step(`^the embedded FS should contain template "([^"]*)"$`, featureContext.theFSShouldContainTemplate)
}

// ---------------------------------------------------------------------------
// Given
// ---------------------------------------------------------------------------

func (f *FeatureContext) theRegistryIsLoaded() error {
	f.registry = sdd.NewDefaultRegistry()
	return nil
}

// ---------------------------------------------------------------------------
// When
// ---------------------------------------------------------------------------

func (f *FeatureContext) iLookUpStandard(id string) error {
	if f.registry == nil {
		return fmt.Errorf("registry not initialized — Background step missing?")
	}
	f.standard, f.resolveErr = f.registry.Lookup(id)
	return nil
}

func (f *FeatureContext) iResolveWithLayers(flag, project, user string) error {
	if f.registry == nil {
		return fmt.Errorf("registry not initialized — Background step missing?")
	}
	f.standard, f.resolveErr = f.registry.Resolve(flag, project, user)
	return nil
}

// ---------------------------------------------------------------------------
// Then — adapter contract
// ---------------------------------------------------------------------------

func (f *FeatureContext) theLookupShouldSucceed() error {
	if err := assertions.AssertActual(assert.Nil, f.resolveErr, "expected lookup to succeed"); err != nil {
		return err
	}
	return assertions.AssertActual(assert.NotNil, f.standard, "expected non-nil standard")
}

func (f *FeatureContext) theDisplayNameShouldBe(want string) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, want, f.standard.DisplayName(), "display name mismatch")
}

func (f *FeatureContext) theTemplateDirShouldBe(want string) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, want, f.standard.TemplateDir(), "template dir mismatch")
}

func (f *FeatureContext) theOutputLayoutShouldBe(want string) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	got := layoutLabel(f.standard.OutputLayout())
	return assertions.AssertExpectedAndActual(assert.Equal, want, got, "output layout mismatch")
}

func (f *FeatureContext) theArtifactsShouldIncludeExactly(table *godog.Table) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	want := tableToSet(table)
	got := make(map[string]bool, len(f.standard.BootstrapArtifacts()))
	for _, a := range f.standard.BootstrapArtifacts() {
		got[a.FileName] = true
	}
	if !sameSet(want, got) {
		return fmt.Errorf("artifact set mismatch:\n  want: %v\n  got:  %v", keys(want), keys(got))
	}
	return nil
}

func (f *FeatureContext) everyArtifactRequired() error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	for _, a := range f.standard.BootstrapArtifacts() {
		if !a.Required {
			return fmt.Errorf("artifact %q is not required (expected ALL required for this standard)", a.FileName)
		}
	}
	return nil
}

func (f *FeatureContext) everyRequiredArtifactLowercase() error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	for _, a := range f.standard.BootstrapArtifacts() {
		if !a.Required {
			continue
		}
		if a.FileName != strings.ToLower(a.FileName) {
			return fmt.Errorf("required artifact %q is not lowercase", a.FileName)
		}
	}
	return nil
}

func (f *FeatureContext) noArtifactContaining(needle string) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	needle = strings.ToLower(needle)
	for _, a := range f.standard.BootstrapArtifacts() {
		if strings.Contains(strings.ToLower(a.FileName), needle) {
			return fmt.Errorf("found unexpected artifact %q (must not contain %q)", a.FileName, needle)
		}
	}
	return nil
}

func (f *FeatureContext) theRequiredFilesShouldBe(table *godog.Table) error {
	return f.checkFilteredFiles(table, true)
}

func (f *FeatureContext) theOptionalFilesShouldBe(table *godog.Table) error {
	return f.checkFilteredFiles(table, false)
}

func (f *FeatureContext) checkFilteredFiles(table *godog.Table, requiredFlag bool) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	want := tableToSet(table)
	got := make(map[string]bool)
	for _, a := range f.standard.BootstrapArtifacts() {
		if a.Required == requiredFlag {
			got[a.FileName] = true
		}
	}
	if !sameSet(want, got) {
		label := "optional"
		if requiredFlag {
			label = "required"
		}
		return fmt.Errorf("%s files mismatch:\n  want: %v\n  got:  %v", label, keys(want), keys(got))
	}
	return nil
}

func (f *FeatureContext) theLifecycleWorkflowIDsShouldBe(table *godog.Table) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	want := tableToList(table)
	got := f.standard.LifecycleWorkflowIDs()
	return assertions.AssertExpectedAndActual(assert.Equal, want, got, "lifecycle workflow IDs mismatch")
}

func (f *FeatureContext) theHintsInLocaleShouldMention(locale, needle string) error {
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	hints := f.standard.SystemPromptHints(locale)
	if !strings.Contains(strings.ToLower(hints), strings.ToLower(needle)) {
		return fmt.Errorf("hints in %q should mention %q, got: %s", locale, needle, hints)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Then — resolution
// ---------------------------------------------------------------------------

func (f *FeatureContext) theResolvedIDShouldBe(want string) error {
	if f.resolveErr != nil {
		return fmt.Errorf("expected resolution success, got error: %v", f.resolveErr)
	}
	if f.standard == nil {
		return fmt.Errorf("no standard resolved")
	}
	return assertions.AssertExpectedAndActual(assert.Equal, want, f.standard.ID(), "resolved ID mismatch")
}

func (f *FeatureContext) resolutionShouldFailWith(needle string) error {
	if f.resolveErr == nil {
		return fmt.Errorf("expected resolution to fail with error containing %q, got nil error", needle)
	}
	if !strings.Contains(f.resolveErr.Error(), needle) {
		return fmt.Errorf("expected error containing %q, got %q", needle, f.resolveErr.Error())
	}
	return nil
}

// ---------------------------------------------------------------------------
// Then — embedded FS
// ---------------------------------------------------------------------------

func (f *FeatureContext) theFSShouldContainTemplate(path string) error {
	data, err := root.TemplatesFS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("template %q not found in embedded FS: %w", path, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("template %q is empty", path)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// layoutLabel mapea el enum a un string estable usado en los Gherkin steps.
// Mantener acá (en vez de en el adapter) evita acoplar el dominio a labels
// que solo existen para tests.
func layoutLabel(l service.OutputLayout) string {
	switch l {
	case service.LayoutFlat:
		return "flat"
	case service.LayoutFeatureGrouped:
		return "feature-grouped"
	default:
		return fmt.Sprintf("unknown(%d)", l)
	}
}

func tableToSet(t *godog.Table) map[string]bool {
	out := make(map[string]bool, len(t.Rows))
	for _, row := range t.Rows {
		for _, cell := range row.Cells {
			val := strings.TrimSpace(cell.Value)
			if val != "" {
				out[val] = true
			}
		}
	}
	return out
}

func tableToList(t *godog.Table) []string {
	var out []string
	for _, row := range t.Rows {
		for _, cell := range row.Cells {
			val := strings.TrimSpace(cell.Value)
			if val != "" {
				out = append(out, val)
			}
		}
	}
	return out
}

func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
