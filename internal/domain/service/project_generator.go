package service

import (
	"fmt"

	"github.com/jorelcb/ai-context-generator/internal/domain/project"
	"github.com/jorelcb/ai-context-generator/internal/domain/shared"
)

// ProjectGenerator provides domain operations for generating projects
type ProjectGenerator struct {
	projectRepo project.Repository
}

// NewProjectGenerator creates a new project generator service
func NewProjectGenerator(projectRepo project.Repository) *ProjectGenerator {
	return &ProjectGenerator{
		projectRepo: projectRepo,
	}
}

// CreateProject creates a new project with validation
func (s *ProjectGenerator) CreateProject(
	id string,
	name shared.ProjectName,
	language shared.Language,
	projectType shared.ProjectType,
	architecture shared.Architecture,
	outputPath string,
	capabilities []string,
) (*project.Project, error) {

	// Check if project already exists
	exists, err := s.projectRepo.ExistsByName(name.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("project with name %s already exists", name.Value())
	}

	// Create project entity
	proj, err := project.NewProject(id, name, language, projectType, architecture, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Add capabilities
	for _, capability := range capabilities {
		if err := proj.AddCapability(capability); err != nil {
			return nil, fmt.Errorf("failed to add capability %s: %w", capability, err)
		}
	}

	// Validate project
	if err := proj.Validate(); err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}

	// Save project
	if err := s.projectRepo.Save(proj); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	return proj, nil
}

// GetProject retrieves a project by ID
func (s *ProjectGenerator) GetProject(id string) (*project.Project, error) {
	return s.projectRepo.FindByID(id)
}

// GetProjectByName retrieves a project by name
func (s *ProjectGenerator) GetProjectByName(name string) (*project.Project, error) {
	return s.projectRepo.FindByName(name)
}

// ListProjects retrieves all projects
func (s *ProjectGenerator) ListProjects() ([]*project.Project, error) {
	return s.projectRepo.FindAll()
}

// DeleteProject deletes a project by ID
func (s *ProjectGenerator) DeleteProject(id string) error {
	exists, err := s.projectRepo.Exists(id)
	if err != nil {
		return fmt.Errorf("failed to check project existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("project with id %s does not exist", id)
	}

	return s.projectRepo.Delete(id)
}