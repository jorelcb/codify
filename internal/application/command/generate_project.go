package command

import (
	"context"
	"fmt"
	"time"

	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/project"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
	"github.com/jorelcb/ai-context-generator/internal/domain/shared"
	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// GenerateProjectCommand representa el comando para generar un proyecto
type GenerateProjectCommand struct {
	projectRepo   project.Repository
	templateRepo  template.Repository
	projectGen    *service.ProjectGenerator
	templateProc  *service.TemplateProcessor
}

// NewGenerateProjectCommand crea una nueva instancia del comando
func NewGenerateProjectCommand(
	projectRepo project.Repository,
	templateRepo template.Repository,
	projectGen *service.ProjectGenerator,
	templateProc *service.TemplateProcessor,
) *GenerateProjectCommand {
	return &GenerateProjectCommand{
		projectRepo:  projectRepo,
		templateRepo: templateRepo,
		projectGen:   projectGen,
		templateProc: templateProc,
	}
}

// Execute ejecuta el comando de generación de proyecto
func (c *GenerateProjectCommand) Execute(ctx context.Context, config *dto.ProjectConfig) (*dto.ProjectInfo, error) {
	// 1. Validar configuración
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 2. Crear value objects del domain
	projectName, err := shared.NewProjectName(config.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid project name: %w", err)
	}

	language, err := shared.NewLanguage(config.Language)
	if err != nil {
		return nil, fmt.Errorf("invalid language: %w", err)
	}

	projectType, err := shared.NewProjectType(config.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid project type: %w", err)
	}

	architecture, err := shared.NewArchitecture(config.Architecture)
	if err != nil {
		return nil, fmt.Errorf("invalid architecture: %w", err)
	}

	// 3. Generar ID único para el proyecto
	id := fmt.Sprintf("proj_%d", time.Now().UnixNano())

	// 4. Crear entidad Project usando el domain service
	proj, err := c.projectGen.CreateProject(
		id,
		projectName,
		language,
		projectType,
		architecture,
		config.OutputPath,
		config.Capabilities,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate project: %w", err)
	}

	// 5. Agregar metadata si existe
	if config.Metadata != nil {
		for key, value := range config.Metadata {
			proj.SetMetadata(key, value)
		}
	}

	// 6. Convertir a DTO y retornar (el proyecto ya fue guardado por el service)
	return &dto.ProjectInfo{
		ID:           proj.ID(),
		Name:         proj.Name().String(),
		Language:     proj.Language().String(),
		Type:         proj.ProjectType().String(),
		Architecture: proj.Architecture().String(),
		OutputPath:   proj.OutputPath(),
		Capabilities: proj.Capabilities(),
		CreatedAt:    proj.CreatedAt(),
		UpdatedAt:    proj.UpdatedAt(),
	}, nil
}