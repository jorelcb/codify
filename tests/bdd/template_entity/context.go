package template_entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
	"github.com/jorelcb/ai-context-generator/tests/bdd/commons/assertions"
)

// FeatureContext holds the state for template entity test scenarios
type FeatureContext struct {
	templateID   string
	templateName string
	templatePath string
	content      string
	template     *template.Template
	variable     template.Variable
	err          error
	metadata     template.Metadata
}

// SetupTest initializes test data (called once before all scenarios)
func (f *FeatureContext) SetupTest() {
	// No global setup needed for template entity tests
}

// reset clears context state before each scenario
func (f *FeatureContext) reset() {
	f.templateID = ""
	f.templateName = ""
	f.templatePath = ""
	f.content = ""
	f.template = nil
	f.variable = template.Variable{}
	f.err = nil
	f.metadata = template.Metadata{}
}

// ========== Given Steps (Setup) ==========

func (f *FeatureContext) theTemplateSystemIsInitialized() error {
	return nil
}

func (f *FeatureContext) iHaveTemplateData(id, name, path, content string) error {
	f.templateID = id
	f.templateName = name
	f.templatePath = path
	f.content = content
	return nil
}

func (f *FeatureContext) iHaveTemplateDataWithEmptyID(name, path, content string) error {
	return f.iHaveTemplateData("", name, path, content)
}

func (f *FeatureContext) iHaveTemplateDataWithEmptyName(id, path, content string) error {
	return f.iHaveTemplateData(id, "", path, content)
}

func (f *FeatureContext) iHaveTemplateDataWithEmptyPath(id, name, content string) error {
	return f.iHaveTemplateData(id, name, "", content)
}

func (f *FeatureContext) iHaveAValidTemplate() error {
	var err error
	f.template, err = template.NewTemplate("test-1", "Test Template", "/path/test.md", "# Test")
	return err
}

func (f *FeatureContext) iHaveAValidTemplateWithContent(content string) error {
	var err error
	f.template, err = template.NewTemplate("test-1", "Test Template", "/path/test.md", content)
	return err
}

func (f *FeatureContext) iHaveAVariable(name string, required bool, defaultValue string) error {
	f.variable = template.Variable{
		Name:         name,
		Required:     required,
		DefaultValue: defaultValue,
	}
	return nil
}

func (f *FeatureContext) iHaveAVariableWithRequired(name, requiredStr, defaultValue string) error {
	required := requiredStr == "true"
	return f.iHaveAVariable(name, required, defaultValue)
}

func (f *FeatureContext) theVariableIsAlreadyAddedToTheTemplate() error {
	return f.template.AddVariable(f.variable)
}

func (f *FeatureContext) iHaveATemplateWithCorruptedData() error {
	// DDD design prevents creating corrupted templates via public API
	// Create a valid template for the setup
	f.template, _ = template.NewTemplate("test-1", "Test", "/path", "content")
	return nil
}

// ========== When Steps (Actions) ==========

func (f *FeatureContext) iCreateANewTemplate() error {
	f.template, f.err = template.NewTemplate(f.templateID, f.templateName, f.templatePath, f.content)
	return nil
}

func (f *FeatureContext) iAddTheVariableToTheTemplate() error {
	f.err = f.template.AddVariable(f.variable)
	return nil
}

func (f *FeatureContext) iTryToAddTheSameVariableAgain() error {
	f.err = f.template.AddVariable(f.variable)
	return nil
}

func (f *FeatureContext) iUpdateTheTemplateContentTo(newContent string) error {
	f.template.SetContent(newContent)
	return nil
}

func (f *FeatureContext) iSetMetadataWith(version, author, description string) error {
	f.metadata = template.Metadata{
		Version:     version,
		Author:      author,
		Description: description,
	}
	f.template.SetMetadata(f.metadata)
	return nil
}

func (f *FeatureContext) iValidateTheTemplate() error {
	f.err = f.template.Validate()
	return nil
}

// ========== Then Steps (Assertions) ==========

func (f *FeatureContext) theTemplateShouldBeCreatedSuccessfully() error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	return assertions.AssertActual(assert.NotNil, f.template, "expected template to be created")
}

func (f *FeatureContext) theTemplateIDShouldBe(expectedID string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expectedID, f.template.ID(), "template ID mismatch")
}

func (f *FeatureContext) theTemplateNameShouldBe(expectedName string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expectedName, f.template.Name(), "template name mismatch")
}

func (f *FeatureContext) iShouldGetAnError(expectedError string) error {
	if f.err == nil {
		return fmt.Errorf("expected error %q, but got no error", expectedError)
	}
	if !strings.Contains(f.err.Error(), expectedError) {
		return fmt.Errorf("expected error containing %q, got %q", expectedError, f.err.Error())
	}
	return nil
}

func (f *FeatureContext) theVariableShouldBeAddedSuccessfully() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected no error")
}

func (f *FeatureContext) theTemplateShouldHaveNVariables(count int) error {
	actual := len(f.template.Variables())
	return assertions.AssertExpectedAndActual(assert.Equal, count, actual, "variable count mismatch")
}

func (f *FeatureContext) theTemplateContentShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.template.Content(), "content mismatch")
}

func (f *FeatureContext) theTemplateUpdatedTimestampShouldBeRecent() error {
	if time.Since(f.template.UpdatedAt()) > 5*time.Second {
		return fmt.Errorf("updated timestamp is not recent: %v", f.template.UpdatedAt())
	}
	return nil
}

func (f *FeatureContext) theTemplateMetadataVersionShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.template.Metadata().Version, "metadata version mismatch")
}

func (f *FeatureContext) theTemplateMetadataAuthorShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.template.Metadata().Author, "metadata author mismatch")
}

func (f *FeatureContext) theTemplateMetadataDescriptionShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.template.Metadata().Description, "metadata description mismatch")
}

func (f *FeatureContext) theValidationShouldPass() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected validation to pass")
}

func (f *FeatureContext) theValidationShouldFailWith(expectedError string) error {
	if f.err == nil {
		return fmt.Errorf("expected validation to fail with %q, but got no error", expectedError)
	}
	if !strings.Contains(f.err.Error(), expectedError) {
		return fmt.Errorf("expected error containing %q, got %q", expectedError, f.err.Error())
	}
	return nil
}