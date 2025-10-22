package project_repository

import (
	"fmt"
	"strings"
	"sync"

	"github.com/stretchr/testify/assert"

	"github.com/jorelcb/ai-context-generator/internal/domain/project"
	"github.com/jorelcb/ai-context-generator/internal/domain/shared"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/persistence/memory"
	"github.com/jorelcb/ai-context-generator/tests/bdd/commons/assertions"
)

// FeatureContext holds the state for project repository test scenarios
type FeatureContext struct {
	repo            *memory.ProjectRepository
	project         *project.Project
	secondProject   *project.Project
	projects        []*project.Project
	retrieved       *project.Project
	retrievedList   []*project.Project
	testCapability  string
	err             error
}

// SetupTest initializes test data (called once before all scenarios)
func (f *FeatureContext) SetupTest() {
	f.repo = memory.NewProjectRepository()
}

// reset clears context state before each scenario
func (f *FeatureContext) reset() {
	f.repo = memory.NewProjectRepository()
	f.project = nil
	f.secondProject = nil
	f.projects = nil
	f.retrieved = nil
	f.retrievedList = nil
	f.err = nil
	f.testCapability = ""
}

// ========== Given Steps (Setup) ==========

func (f *FeatureContext) anEmptyProjectRepository() error {
	f.repo = memory.NewProjectRepository()
	return nil
}

func (f *FeatureContext) iHaveAProjectWithID(id string) error {
	name, err := shared.NewProjectName("test-project")
	if err != nil {
		return err
	}

	lang, err := shared.NewLanguage("go")
	if err != nil {
		return err
	}

	projType, err := shared.NewProjectType("api")
	if err != nil {
		return err
	}

	arch, err := shared.NewArchitecture("clean")
	if err != nil {
		return err
	}

	f.project, err = project.NewProject(id, name, lang, projType, arch, "/tmp/test-project")
	if err != nil {
		return err
	}

	f.projects = append(f.projects, f.project)
	return nil
}

func (f *FeatureContext) iHaveAProjectWithName(name string) error {
	projectName, err := shared.NewProjectName(name)
	if err != nil {
		return err
	}

	lang, err := shared.NewLanguage("go")
	if err != nil {
		return err
	}

	projType, err := shared.NewProjectType("api")
	if err != nil {
		return err
	}

	arch, err := shared.NewArchitecture("clean")
	if err != nil {
		return err
	}

	f.project, err = project.NewProject("proj-1", projectName, lang, projType, arch, "/tmp/"+name)
	if err != nil {
		return err
	}

	f.projects = append(f.projects, f.project)
	return nil
}

func (f *FeatureContext) iHaveAProjectWithNameAndID(name, id string) error {
	projectName, err := shared.NewProjectName(name)
	if err != nil {
		return err
	}

	lang, err := shared.NewLanguage("go")
	if err != nil {
		return err
	}

	projType, err := shared.NewProjectType("api")
	if err != nil {
		return err
	}

	arch, err := shared.NewArchitecture("clean")
	if err != nil {
		return err
	}

	proj, err := project.NewProject(id, projectName, lang, projType, arch, "/tmp/"+name)
	if err != nil {
		return err
	}

	if f.project == nil {
		f.project = proj
	} else {
		f.secondProject = proj
	}

	f.projects = append(f.projects, proj)
	return nil
}

func (f *FeatureContext) iHaveAnInvalidProject() error {
	// DDD design prevents creating invalid projects via public API
	// Create a valid project for the setup
	name, _ := shared.NewProjectName("valid-project")
	lang, _ := shared.NewLanguage("go")
	projType, _ := shared.NewProjectType("api")
	arch, _ := shared.NewArchitecture("clean")
	f.project, _ = project.NewProject("valid-id", name, lang, projType, arch, "/tmp/valid")
	return nil
}

func (f *FeatureContext) iHaveNProjects(n int) error {
	f.projects = make([]*project.Project, 0, n)
	for i := 0; i < n; i++ {
		name, err := shared.NewProjectName(fmt.Sprintf("project-%d", i))
		if err != nil {
			return err
		}

		lang, err := shared.NewLanguage("go")
		if err != nil {
			return err
		}

		projType, err := shared.NewProjectType("api")
		if err != nil {
			return err
		}

		arch, err := shared.NewArchitecture("clean")
		if err != nil {
			return err
		}

		proj, err := project.NewProject(
			fmt.Sprintf("proj-%d", i),
			name,
			lang,
			projType,
			arch,
			fmt.Sprintf("/tmp/project-%d", i),
		)
		if err != nil {
			return err
		}
		f.projects = append(f.projects, proj)
	}
	return nil
}

func (f *FeatureContext) iHaveNProjectsSaved(n int) error {
	if err := f.iHaveNProjects(n); err != nil {
		return err
	}
	return f.iSaveAllProjects()
}

// ========== When Steps (Actions) ==========

func (f *FeatureContext) iSaveTheProject() error {
	f.err = f.repo.Save(f.project)
	return nil
}

func (f *FeatureContext) iSaveTheFirstProject() error {
	if len(f.projects) == 0 {
		return fmt.Errorf("no projects available")
	}
	f.err = f.repo.Save(f.projects[0])
	return nil
}

func (f *FeatureContext) iSaveTheSecondProject() error {
	if len(f.projects) < 2 {
		return fmt.Errorf("second project not available")
	}
	f.err = f.repo.Save(f.projects[1])
	return nil
}

func (f *FeatureContext) iSaveAllProjects() error {
	for _, proj := range f.projects {
		if err := f.repo.Save(proj); err != nil {
			f.err = err
			return nil
		}
	}
	return nil
}

func (f *FeatureContext) iSaveAllProjectsConcurrently() error {
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(f.projects))

	for _, proj := range f.projects {
		wg.Add(1)
		go func(p *project.Project) {
			defer wg.Done()
			if err := f.repo.Save(p); err != nil {
				errorsChan <- err
			}
		}(proj)
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

func (f *FeatureContext) iTryToSaveTheProject() error {
	f.err = f.repo.Save(f.project)
	return nil
}

func (f *FeatureContext) iTryToSaveANilProject() error {
	f.err = f.repo.Save(nil)
	return nil
}

func (f *FeatureContext) iRetrieveProjectByID(id string) error {
	f.retrieved, f.err = f.repo.FindByID(id)
	return nil
}

func (f *FeatureContext) iRetrieveProjectByName(name string) error {
	f.retrieved, f.err = f.repo.FindByName(name)
	return nil
}

func (f *FeatureContext) iTryToRetrieveProjectByID(id string) error {
	return f.iRetrieveProjectByID(id)
}

func (f *FeatureContext) iTryToRetrieveProjectByName(name string) error {
	return f.iRetrieveProjectByName(name)
}

func (f *FeatureContext) iRetrieveAllProjects() error {
	f.retrievedList, f.err = f.repo.FindAll()
	return nil
}

func (f *FeatureContext) iAddACapabilityToTheProject() error {
	f.testCapability = "test-capability"
	f.project.AddCapability(f.testCapability)
	return nil
}

func (f *FeatureContext) iSaveTheProjectAgain() error {
	return f.iSaveTheProject()
}

func (f *FeatureContext) iDeleteProjectWithID(id string) error {
	f.err = f.repo.Delete(id)
	return nil
}

func (f *FeatureContext) iTryToDeleteProjectWithID(id string) error {
	return f.iDeleteProjectWithID(id)
}

func (f *FeatureContext) iClearTheRepository() error {
	f.repo.Clear()
	return nil
}

// ========== Then Steps (Assertions) ==========

func (f *FeatureContext) theProjectShouldBeSavedSuccessfully() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected no error")
}

func (f *FeatureContext) iShouldReceiveTheProject() error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	return assertions.AssertActual(assert.NotNil, f.retrieved, "expected to receive project")
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

func (f *FeatureContext) allNProjectsShouldBeSaved(n int) error {
	count := f.repo.Count()
	return assertions.AssertExpectedAndActual(assert.Equal, n, count, "project count mismatch")
}

func (f *FeatureContext) iShouldReceiveNProjects(n int) error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}
	return assertions.AssertExpectedAndActual(assert.Equal, n, len(f.retrievedList), "projects count mismatch")
}

func (f *FeatureContext) theProjectShouldHaveTheNewCapability() error {
	retrieved, err := f.repo.FindByID(f.project.ID())
	if err != nil {
		return err
	}

	found := false
	for _, cap := range retrieved.Capabilities() {
		if cap == f.testCapability {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("expected project to have capability %q", f.testCapability)
	}
	return nil
}

func (f *FeatureContext) theProjectShouldBeDeleted() error {
	return assertions.AssertActual(assert.Nil, f.err, "expected no error deleting project")
}

func (f *FeatureContext) projectWithIDShouldExist(id string) error {
	exists, err := f.repo.Exists(id)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.True, exists, fmt.Sprintf("expected project %s to exist", id))
}

func (f *FeatureContext) projectWithIDShouldNotExist(id string) error {
	exists, err := f.repo.Exists(id)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.False, exists, fmt.Sprintf("expected project %s to not exist", id))
}

func (f *FeatureContext) projectWithNameShouldExist(name string) error {
	exists, err := f.repo.ExistsByName(name)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.True, exists, fmt.Sprintf("expected project with name %s to exist", name))
}

func (f *FeatureContext) projectWithNameShouldNotExist(name string) error {
	exists, err := f.repo.ExistsByName(name)
	if err != nil {
		return err
	}
	return assertions.AssertBool(assert.False, exists, fmt.Sprintf("expected project with name %s to not exist", name))
}

func (f *FeatureContext) theSecondProjectShouldOverwriteByName() error {
	if err := assertions.AssertActual(assert.Nil, f.err, "expected no error"); err != nil {
		return err
	}

	// Verify that only one project exists with that name
	count := f.repo.Count()
	if err := assertions.AssertExpectedAndActual(assert.Equal, 1, count, "expected 1 project after overwrite"); err != nil {
		return err
	}

	// Verify it's the second project by ID
	if len(f.projects) < 2 {
		return fmt.Errorf("not enough projects to verify overwrite")
	}

	retrieved, err := f.repo.FindByName(f.projects[1].Name().Value())
	if err != nil {
		return err
	}

	return assertions.AssertExpectedAndActual(assert.Equal, f.projects[1].ID(), retrieved.ID(), "expected second project ID")
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