package command

import (
	"context"
	"testing"

	"github.com/jorelcb/ai-context-generator/internal/application/dto"
	"github.com/jorelcb/ai-context-generator/internal/domain/service"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/persistence/memory"
)

func TestGenerateProjectCommand_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  *dto.ProjectConfig
		wantErr bool
	}{
		{
			name: "successful project generation",
			config: &dto.ProjectConfig{
				Name:         "test-project",
				Language:     "go",
				Type:         "api",
				Architecture: "clean",
				OutputPath:   "/tmp/test-project",
				Capabilities: []string{"rest", "database"},
				Metadata:     map[string]string{"author": "test"},
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing name",
			config: &dto.ProjectConfig{
				Name:         "",
				Language:     "go",
				Type:         "api",
				Architecture: "clean",
				OutputPath:   "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing language",
			config: &dto.ProjectConfig{
				Name:         "test-project",
				Language:     "",
				Type:         "api",
				Architecture: "clean",
				OutputPath:   "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "invalid language",
			config: &dto.ProjectConfig{
				Name:         "test-project",
				Language:     "invalid-language",
				Type:         "api",
				Architecture: "clean",
				OutputPath:   "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "invalid project type",
			config: &dto.ProjectConfig{
				Name:         "test-project",
				Language:     "go",
				Type:         "invalid-type",
				Architecture: "clean",
				OutputPath:   "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "invalid architecture",
			config: &dto.ProjectConfig{
				Name:         "test-project",
				Language:     "go",
				Type:         "api",
				Architecture: "invalid-arch",
				OutputPath:   "/tmp/test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			projectRepo := memory.NewProjectRepository()
			templateRepo := memory.NewTemplateRepository()
			projectGen := service.NewProjectGenerator(projectRepo)
			templateProc := service.NewTemplateProcessor(templateRepo)

			cmd := NewGenerateProjectCommand(
				projectRepo,
				templateRepo,
				projectGen,
				templateProc,
			)

			// Execute
			result, err := cmd.Execute(ctx, tt.config)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateProjectCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify result
				if result == nil {
					t.Error("GenerateProjectCommand.Execute() result is nil")
					return
				}
				if result.ID == "" {
					t.Error("GenerateProjectCommand.Execute() result.ID is empty")
				}
				if result.Name != tt.config.Name {
					t.Errorf("GenerateProjectCommand.Execute() result.Name = %v, want %v", result.Name, tt.config.Name)
				}
				if result.Language != tt.config.Language {
					t.Errorf("GenerateProjectCommand.Execute() result.Language = %v, want %v", result.Language, tt.config.Language)
				}
				if result.Type != tt.config.Type {
					t.Errorf("GenerateProjectCommand.Execute() result.Type = %v, want %v", result.Type, tt.config.Type)
				}
				if result.Architecture != tt.config.Architecture {
					t.Errorf("GenerateProjectCommand.Execute() result.Architecture = %v, want %v", result.Architecture, tt.config.Architecture)
				}
			}
		})
	}
}

func TestGenerateProjectCommand_DuplicateProjectName(t *testing.T) {
	// Setup
	ctx := context.Background()
	projectRepo := memory.NewProjectRepository()
	templateRepo := memory.NewTemplateRepository()
	projectGen := service.NewProjectGenerator(projectRepo)
	templateProc := service.NewTemplateProcessor(templateRepo)

	cmd := NewGenerateProjectCommand(
		projectRepo,
		templateRepo,
		projectGen,
		templateProc,
	)

	config := &dto.ProjectConfig{
		Name:         "duplicate-project",
		Language:     "go",
		Type:         "api",
		Architecture: "clean",
		OutputPath:   "/tmp/duplicate",
	}

	// First execution should succeed
	_, err := cmd.Execute(ctx, config)
	if err != nil {
		t.Fatalf("First Execute() failed: %v", err)
	}

	// Second execution with same name should fail
	_, err = cmd.Execute(ctx, config)
	if err == nil {
		t.Error("Second Execute() with duplicate name should fail")
	}
}