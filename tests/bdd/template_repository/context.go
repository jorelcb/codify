package template_repository

import (
	"fmt"
	"strings"
	"sync"

	"github.com/stretchr/testify/assert"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/persistence/memory"
	"github.com/jorelcb/ai-context-generator/tests/bdd/commons/assertions"
)

// FeatureContext holds the state for template repository test scenarios
type FeatureContext struct {
	repo          *memory.TemplateRepository
	template      *template.Template
	templates     []*template.Template
	retrieved     *template.Template
	retrievedList []*template.Template
	err           error
}

// SetupTest initializes test data (called once before all scenarios)
func (f *FeatureContext) SetupTest() {
	f.repo = memory.NewTemplateRepository()
}

// reset clears context state before each scenario
func (f *FeatureContext) reset() {
	f.repo = memory.NewTemplateRepository()
	f.template = nil
	f.templates = nil
	f.retrieved = nil
	f.retrievedList = nil
	f.err = nil
}

// ========== Given Steps (Setup) ==========

func (f *FeatureContext) anEmptyTemplateRepository() error {
	f.repo = memory.NewTemplateRepository()
	return nil
}

func (f *FeatureContext) iHaveATemplateWithID(id string) error {
	var err error
	f.template, err = template.NewTemplate(id, "Test Template", "/path/test.md", "# Test Content")
	if err != nil {
		return err
	}
	f.templates = append(f.templates, f.template)
	return nil
}

func (f *FeatureContext) iHaveATemplateWithPath(path string) error {
	var err error
	f.template, err = template.NewTemplate("test-1", "Test Template", path, "# Test Content")
	if err != nil {
		return err
	}
	f.templates = append(f.templates, f.template)
	return nil
}

func (f *FeatureContext) iHaveATemplateWithIDAndTag(id, tag string) error {
	tmpl, err := template.NewTemplate(id, "Test Template", "/path/test.md", "# Test Content")
	if err != nil {
		return err
	}
	metadata := template.Metadata{Tags: []string{tag}}
	tmpl.SetMetadata(metadata)
	f.templates = append(f.templates, tmpl)
	return nil
}

func (f *FeatureContext) iHaveAnInvalidTemplate() error {
	// DDD design prevents creating invalid templates via public API
	// Create a valid template for the setup
	f.template, _ = template.NewTemplate("valid-id", "name", "path", "content")
	return nil
}

func (f *FeatureContext) iHaveNTemplates(n int) error {
	f.templates = make([]*template.Template, 0, n)
	for i := 0; i < n; i++ {
		tmpl, err := template.NewTemplate(
			fmt.Sprintf("tmpl-%d", i),
			fmt.Sprintf("Template %d", i),
			fmt.Sprintf("/path/tmpl-%d.md", i),
			"# Content",
		)
		if err != nil {
			return err
		}
		f.templates = append(f.templates, tmpl)
	}
	return nil
}

func (f *FeatureContext) iHaveNTemplatesSaved(n int) error {
	if err := f.iHaveNTemplates(n); err != nil {
		return err
	}
	return f.iSaveAllTemplates()
}

// ========== When Steps (Actions) ==========

func (f *FeatureContext) iSaveTheTemplate() error {
	f.err = f.repo.Save(f.template)
	return nil
}

func (f *FeatureContext) iSaveAllTemplates() error {
	for _, tmpl := range f.templates {
		if err := f.repo.Save(tmpl); err != nil {
			f.err = err
			return nil
		}
	}
	return nil
}

func (f *FeatureContext) iSaveAllTemplatesConcurrently() error {
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(f.templates))

	for _, tmpl := range f.templates {
		wg.Add(1)
		go func(t *template.Template) {
			defer wg.Done()
			if err := f.repo.Save(t); err != nil {
				errorsChan <- err
			}
		}(tmpl)
	}

	wg.Wait()
	close(errorsChan)

	// Check if any errors occurred
	for err := range errorsChan {
		f.err = err
		break
	}
	return nil
}

func (f *FeatureContext) iTryToSaveTheTemplate() error {
	f.err = f.repo.Save(f.template)
	return nil
}

func (f *FeatureContext) iTryToSaveANilTemplate() error {
	f.err = f.repo.Save(nil)
	return nil
}

func (f *FeatureContext) iRetrieveTemplateByID(id string) error {
	f.retrieved, f.err = f.repo.FindByID(id)
	return nil
}

func (f *FeatureContext) iRetrieveTemplateByPath(path string) error {
	f.retrieved, f.err = f.repo.FindByPath(path)
	return nil
}

func (f *FeatureContext) iTryToRetrieveTemplateByID(id string) error {
	return f.iRetrieveTemplateByID(id)
}

func (f *FeatureContext) iTryToRetrieveTemplateByPath(path string) error {
	return f.iRetrieveTemplateByPath(path)
}

func (f *FeatureContext) iRetrieveAllTemplates() error {
	f.retrievedList, f.err = f.repo.FindAll()
	return nil
}

func (f *FeatureContext) iFindTemplatesByTag(tag string) error {
	f.retrievedList, f.err = f.repo.FindByTag(tag)
	return nil
}

func (f *FeatureContext) iUpdateTheTemplateContent() error {
	f.template.SetContent("# Updated Content")
	return nil
}

func (f *FeatureContext) iSaveTheTemplateAgain() error {
	return f.iSaveTheTemplate()
}

func (f *FeatureContext) iDeleteTemplateWithID(id string) error {
	f.err = f.repo.Delete(id)
	return nil
}

func (f *FeatureContext) iTryToDeleteTemplateWithID(id string) error {
	return f.iDeleteTemplateWithID(id)
}

func (f *FeatureContext) iClearTheRepository() error {
	f.repo.Clear()
	return nil
}

// ========== Then Steps (Assertions) ==========

func (f *FeatureContext) theTemplateShouldBeSavedSuccessfully() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected no error")
}

func (f *FeatureContext) iShouldReceiveTheTemplate() error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	return assertions.AssertActual(assert.NotNil, f.retrieved, "expected to receive template")
}

func (f *FeatureContext) iShouldGetAnErrorContaining(expected string) error {
	if f.err == nil {
		return fmt.Errorf("expected error containing %q, but got no error", expected)
	}
	if !strings.Contains(f.err.Error(), expected) {
		return fmt.Errorf("expected error containing %q, got %q", expected, f.err.Error())
	}
	return nil
}

func (f *FeatureContext) allNTemplatesShouldBeSaved(n int) error {
	count := f.repo.Count()
	return assertions.AssertExpectedAndActual(assert.Equal, n, count, "template count mismatch")
}

func (f *FeatureContext) iShouldReceiveNTemplates(n int) error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.retrievedList), "templates count mismatch")
}

func (f *FeatureContext) theTemplateShouldHaveUpdatedContent() error {
	retrieved, err := f.repo.FindByID(f.template.ID())
	if err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, "# Updated Content", retrieved.Content(), "content mismatch")
}

func (f *FeatureContext) theTemplateShouldBeDeleted() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected no error deleting template")
}

func (f *FeatureContext) templateWithIDShouldExist(id string) error {
	exists, err := f.repo.Exists(id)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.True, exists, fmt.Sprintf("expected template %s to exist", id))
}

func (f *FeatureContext) templateWithIDShouldNotExist(id string) error {
	exists, err := f.repo.Exists(id)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.False, exists, fmt.Sprintf("expected template %s to not exist", id))
}

func (f *FeatureContext) iShouldGetAValidationError() error {
	if f.err == nil {
		return fmt.Errorf("expected validation error, but got no error")
	}
	if !strings.Contains(f.err.Error(), "invalid") {
		return fmt.Errorf("expected validation error, got %q", f.err.Error())
	}
	return nil
}

func (f *FeatureContext) theRepositoryCountShouldBe(expected int) error {
	count := f.repo.Count()
	return assertions.AssertExpectedAndActual(assert.Equal, expected, count, "repository count mismatch")
}

func (f *FeatureContext) theRepositoryShouldBeEmpty() error {
	return f.theRepositoryCountShouldBe(0)
}