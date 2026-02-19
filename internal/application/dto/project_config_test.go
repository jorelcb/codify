package dto

import (
	"testing"
)

func TestProjectConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ProjectConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ProjectConfig{
				Name:        "test-project",
				Description: "A REST API for inventory management with Go",
				OutputPath:  "/tmp/test",
			},
			wantErr: false,
		},
		{
			name: "valid config with optional fields",
			config: ProjectConfig{
				Name:         "test-project",
				Description:  "A REST API for inventory management with Go",
				Language:     "go",
				Type:         "api",
				Architecture: "clean",
				OutputPath:   "/tmp/test",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: ProjectConfig{
				Name:        "",
				Description: "A test project description",
				OutputPath:  "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			config: ProjectConfig{
				Name:       "test-project",
				OutputPath: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing output path",
			config: ProjectConfig{
				Name:        "test-project",
				Description: "A test project description",
				OutputPath:  "",
			},
			wantErr: true,
		},
		{
			name: "all fields empty",
			config: ProjectConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
