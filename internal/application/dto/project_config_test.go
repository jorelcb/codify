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
				Name:       "test-project",
				Language:   "go",
				OutputPath: "/tmp/test",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: ProjectConfig{
				Name:       "",
				Language:   "go",
				OutputPath: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing language",
			config: ProjectConfig{
				Name:       "test-project",
				Language:   "",
				OutputPath: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "missing output path",
			config: ProjectConfig{
				Name:       "test-project",
				Language:   "go",
				OutputPath: "",
			},
			wantErr: true,
		},
		{
			name: "all fields empty",
			config: ProjectConfig{
				Name:       "",
				Language:   "",
				OutputPath: "",
			},
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
