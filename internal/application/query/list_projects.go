package query

import (
	"context"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/project"
)

// ListProjectsQuery representa la consulta para listar proyectos
type ListProjectsQuery struct {
	projectRepo project.Repository
}

// NewListProjectsQuery crea una nueva instancia de la query
func NewListProjectsQuery(projectRepo project.Repository) *ListProjectsQuery {
	return &ListProjectsQuery{
		projectRepo: projectRepo,
	}
}

// Execute ejecuta la consulta y retorna la lista de proyectos
func (q *ListProjectsQuery) Execute(ctx context.Context) (*dto.ProjectListResult, error) {
	// Obtener todos los proyectos del repositorio
	projects, err := q.projectRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convertir domain entities a DTOs
	projectInfos := make([]dto.ProjectInfo, 0, len(projects))
	for _, proj := range projects {
		projectInfos = append(projectInfos, dto.ProjectInfo{
			ID:           proj.ID(),
			Name:         proj.Name().String(),
			Language:     proj.Language().String(),
			Type:         proj.ProjectType().String(),
			Architecture: proj.Architecture().String(),
			OutputPath:   proj.OutputPath(),
			Capabilities: proj.Capabilities(),
			CreatedAt:    proj.CreatedAt(),
			UpdatedAt:    proj.UpdatedAt(),
		})
	}

	return &dto.ProjectListResult{
		Projects: projectInfos,
		Total:    len(projectInfos),
	}, nil
}