package query

import (
	"context"

	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// ListTemplatesQuery representa la consulta para listar templates
type ListTemplatesQuery struct {
	templateRepo template.Repository
}

// NewListTemplatesQuery crea una nueva instancia de la query
func NewListTemplatesQuery(templateRepo template.Repository) *ListTemplatesQuery {
	return &ListTemplatesQuery{
		templateRepo: templateRepo,
	}
}

// Execute ejecuta la consulta y retorna la lista de templates
func (q *ListTemplatesQuery) Execute(ctx context.Context) (*dto.TemplateListResult, error) {
	// Obtener todos los templates del repositorio
	templates, err := q.templateRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convertir domain entities a DTOs
	templateInfos := make([]dto.TemplateInfo, 0, len(templates))
	for _, tmpl := range templates {
		templateInfos = append(templateInfos, dto.TemplateInfo{
			ID:          tmpl.ID(),
			Name:        tmpl.Name(),
			Path:        tmpl.Path(),
			Description: "", // TODO: Add Description() method to Template
			Language:    "", // TODO: Add Language() method to Template
			Tags:        []string{}, // TODO: Add Tags() method to Template
			Variables:   []string{}, // TODO: Add ExtractVariables() method to Template
		})
	}

	return &dto.TemplateListResult{
		Templates: templateInfos,
		Total:     len(templateInfos),
	}, nil
}