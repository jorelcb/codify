package command

import (
	"context"
	"fmt"
	"time"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/project"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/domain/shared"
)

// GenerateProjectCommand representa el comando para generar un proyecto
type GenerateProjectCommand struct {
	projectRepo project.Repository
	projectGen  *service.ProjectGenerator
}

// NewGenerateProjectCommand crea una nueva instancia del comando
func NewGenerateProjectCommand(
	projectRepo project.Repository,
	projectGen *service.ProjectGenerator,
) *GenerateProjectCommand {
	return &GenerateProjectCommand{
		projectRepo: projectRepo,
		projectGen:  projectGen,
	}
}

// Execute ejecuta el comando de generacion de proyecto
func (c *GenerateProjectCommand) Execute(ctx context.Context, config *dto.ProjectConfig) (*dto.ProjectInfo, error) {
	// 1. Validar configuracion
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 2. Crear value objects del domain
	projectName, err := shared.NewProjectName(config.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid project name: %w", err)
	}

	var language shared.Language
	if config.Language != "" {
		language, err = shared.NewLanguage(config.Language)
		if err != nil {
			return nil, fmt.Errorf("invalid language: %w", err)
		}
	}

	var projectType shared.ProjectType
	if config.Type != "" {
		projectType, err = shared.NewProjectType(config.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid project type: %w", err)
		}
	}

	var architecture shared.Architecture
	if config.Architecture != "" {
		architecture, err = shared.NewArchitecture(config.Architecture)
		if err != nil {
			return nil, fmt.Errorf("invalid architecture: %w", err)
		}
	}

	// 3. Generar ID unico para el proyecto
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

	// 6. Convertir a DTO y retornar
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