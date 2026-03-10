package memory

import (
	"fmt"
	"sync"

	"github.com/jorelcb/ai-context-generator/internal/domain/project"
)

// ProjectRepository is an in-memory implementation of project.Repository
type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]*project.Project
	byName   map[string]*project.Project
}

// NewProjectRepository creates a new in-memory project repository
func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[string]*project.Project),
		byName:   make(map[string]*project.Project),
	}
}

// FindByID retrieves a project by its ID
func (r *ProjectRepository) FindByID(id string) (*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proj, exists := r.projects[id]
	if !exists {
		return nil, fmt.Errorf("project with id %s not found", id)
	}
	return proj, nil
}

// FindByName retrieves a project by its name
func (r *ProjectRepository) FindByName(name string) (*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proj, exists := r.byName[name]
	if !exists {
		return nil, fmt.Errorf("project with name %s not found", name)
	}
	return proj, nil
}

// FindAll retrieves all projects
func (r *ProjectRepository) FindAll() ([]*project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*project.Project, 0, len(r.projects))
	for _, proj := range r.projects {
		projects = append(projects, proj)
	}
	return projects, nil
}

// Save persists a project
func (r *ProjectRepository) Save(proj *project.Project) error {
	if proj == nil {
		return fmt.Errorf("project cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate project before saving
	if err := proj.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	// If a project with the same name but different ID exists, remove the old entry
	name := proj.Name().Value()
	if existing, ok := r.byName[name]; ok && existing.ID() != proj.ID() {
		delete(r.projects, existing.ID())
	}

	r.projects[proj.ID()] = proj
	r.byName[name] = proj
	return nil
}

// Delete removes a project
func (r *ProjectRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	proj, exists := r.projects[id]
	if !exists {
		return fmt.Errorf("project with id %s not found", id)
	}

	delete(r.projects, id)
	delete(r.byName, proj.Name().Value())
	return nil
}

// Exists checks if a project exists
func (r *ProjectRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.projects[id]
	return exists, nil
}

// ExistsByName checks if a project with given name exists
func (r *ProjectRepository) ExistsByName(name string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.byName[name]
	return exists, nil
}

// Clear removes all projects (useful for testing)
func (r *ProjectRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.projects = make(map[string]*project.Project)
	r.byName = make(map[string]*project.Project)
}

// Count returns the number of projects
func (r *ProjectRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.projects)
}
