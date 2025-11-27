package service

import (
	"fmt"
	"os" // os.FileMode is still needed for the interface, but not directly used by domain logic
	"path/filepath"

	"github.com/jorelcb/ai-context-generator/internal/domain/project"
	"github.com/jorelcb/ai-context-generator/internal/domain/shared"
	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// ProjectGenerator provides domain operations for generating projects
type ProjectGenerator struct {
	projectRepo      project.Repository
	fileWriter       FileWriter       // Dependencia de la interfaz
	directoryManager DirectoryManager // Dependencia de la interfaz
	templateEngine   *TemplateEngine  // para renderizar templates
}

// NewProjectGenerator creates a new project generator service
// Ahora acepta las interfaces de sistema de archivos y el template engine
func NewProjectGenerator(
	projectRepo project.Repository,
	fileWriter FileWriter,
	directoryManager DirectoryManager,
	templateEngine *TemplateEngine,
) *ProjectGenerator {
	return &ProjectGenerator{
		projectRepo:      projectRepo,
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
		templateEngine:   templateEngine,
	}
}

// CreateProject creates a new project with validation
// Esta función solo crea la entidad del proyecto y la guarda en el repositorio.
// La generación física de archivos se realizará en GenerateProjectStructure.
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

	// Save project entity (not the files yet)
	if err := s.projectRepo.Save(proj); err != nil {
		return nil, fmt.Errorf("failed to save project entity: %w", err)
	}

	return proj, nil
}

// GenerateProjectStructure creates the physical files and directories for a given project.
// This function assumes the project entity already exists in the repository.
func (s *ProjectGenerator) GenerateProjectStructure(
	proj *project.Project,
	templatesToRender []template.Template, // Lista de templates a renderizar
	templateData map[string]interface{},
) error {
	projectFullPath := proj.FullPath()

	// 1. Create root project directory
	if err := s.directoryManager.CreateDir(projectFullPath, 0755); err != nil {
		return fmt.Errorf("failed to create root project directory at %s: %w", projectFullPath, err)
	}

	// 2. Render and write each template
	for _, tmpl := range templatesToRender {
		// Render content using the template engine
		renderedContent, err := s.templateEngine.Render(tmpl.Content(), templateData)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", tmpl.Name(), err)
		}

		// Determine target path
		// Asume que template.Template tiene un método TargetPath() y FileMode()
		targetPath := filepath.Join(projectFullPath, tmpl.TargetPath())
		fileMode := os.FileMode(tmpl.FileMode())

		// Write the rendered content to file
		if err := s.fileWriter.WriteFile(targetPath, []byte(renderedContent), fileMode); err != nil {
			return fmt.Errorf("failed to write rendered template %s to %s: %w", tmpl.Name(), targetPath, err)
		}
	}

	return nil
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

// Nota: La función RenderAndWrite ha sido eliminada para evitar doble responsabilidad,
// su lógica ahora está integrada en GenerateProjectStructure.