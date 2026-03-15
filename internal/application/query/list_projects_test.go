package query

import (
	"context"
	"testing"

	"github.com/jorelcb/codify/internal/domain/project"
	"github.com/jorelcb/codify/internal/domain/shared"
	"github.com/jorelcb/codify/internal/infrastructure/persistence/memory"
)

func TestListProjectsQuery_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupProjects int
		wantCount     int
	}{
		{
			name:          "empty repository",
			setupProjects: 0,
			wantCount:     0,
		},
		{
			name:          "one project",
			setupProjects: 1,
			wantCount:     1,
		},
		{
			name:          "multiple projects",
			setupProjects: 5,
			wantCount:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			repo := memory.NewProjectRepository()

			// Add test projects
			for i := 0; i < tt.setupProjects; i++ {
				proj := createTestProject(t, i)
				if err := repo.Save(proj); err != nil {
					t.Fatalf("Failed to save test project: %v", err)
				}
			}

			// Create query
			query := NewListProjectsQuery(repo)

			// Execute
			result, err := query.Execute(ctx)

			// Assert
			if err != nil {
				t.Errorf("ListProjectsQuery.Execute() error = %v", err)
				return
			}

			if result.Total != tt.wantCount {
				t.Errorf("ListProjectsQuery.Execute() Total = %v, want %v", result.Total, tt.wantCount)
			}

			if len(result.Projects) != tt.wantCount {
				t.Errorf("ListProjectsQuery.Execute() len(Projects) = %v, want %v", len(result.Projects), tt.wantCount)
			}
		})
	}
}

func createTestProject(t *testing.T, index int) *project.Project {
	t.Helper()

	name, err := shared.NewProjectName("test-project-" + string(rune('0'+index)))
	if err != nil {
		t.Fatalf("Failed to create project name: %v", err)
	}

	language, err := shared.NewLanguage("go")
	if err != nil {
		t.Fatalf("Failed to create language: %v", err)
	}

	projectType, err := shared.NewProjectType("api")
	if err != nil {
		t.Fatalf("Failed to create project type: %v", err)
	}

	architecture, err := shared.NewArchitecture("clean")
	if err != nil {
		t.Fatalf("Failed to create architecture: %v", err)
	}

	proj, err := project.NewProject(
		"test-id-"+string(rune('0'+index)),
		name,
		language,
		projectType,
		architecture,
		"/tmp/test",
	)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	return proj
}