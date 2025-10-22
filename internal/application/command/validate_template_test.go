package command

import (
	"context"
	"testing"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
	"github.com/jorelcb/ai-context-generator/internal/infrastructure/persistence/memory"
)

func TestValidateTemplateCommand_Execute(t *testing.T) {
	tests := []struct {
		name         string
		setupTemplate bool
		templateID   string
		wantErr      bool
		wantValid    bool
	}{
		{
			name:          "valid template",
			setupTemplate: true,
			templateID:    "test-template-1",
			wantErr:       false,
			wantValid:     true,
		},
		{
			name:          "non-existent template",
			setupTemplate: false,
			templateID:    "non-existent",
			wantErr:       true,
			wantValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			repo := memory.NewTemplateRepository()

			if tt.setupTemplate {
				tmpl, err := template.NewTemplate(
					tt.templateID,
					"Test Template",
					"/tmp/test.md",
					"Test content with {{variable}}",
				)
				if err != nil {
					t.Fatalf("Failed to create test template: %v", err)
				}
				if err := repo.Save(tmpl); err != nil {
					t.Fatalf("Failed to save test template: %v", err)
				}
			}

			cmd := NewValidateTemplateCommand(repo)

			// Execute
			result, err := cmd.Execute(ctx, tt.templateID)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplateCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("ValidateTemplateCommand.Execute() result is nil")
					return
				}
				if result.Valid != tt.wantValid {
					t.Errorf("ValidateTemplateCommand.Execute() Valid = %v, want %v", result.Valid, tt.wantValid)
				}
			}
		})
	}
}