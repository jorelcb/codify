package project

import (
	"fmt"
	"time"

	"github.com/jorelcb/ai-context-generator/internal/domain/shared"
)

// Project represents a project entity (aggregate root)
type Project struct {
	id           string
	name         shared.ProjectName
	language     shared.Language
	projectType  shared.ProjectType
	architecture shared.Architecture
	outputPath   string
	capabilities []string
	metadata     map[string]string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewProject creates a new project entity
func NewProject(
	id string,
	name shared.ProjectName,
	language shared.Language,
	projectType shared.ProjectType,
	architecture shared.Architecture,
	outputPath string,
) (*Project, error) {
	if id == "" {
		return nil, fmt.Errorf("project id cannot be empty")
	}
	if outputPath == "" {
		return nil, fmt.Errorf("output path cannot be empty")
	}

	now := time.Now()
	return &Project{
		id:           id,
		name:         name,
		language:     language,
		projectType:  projectType,
		architecture: architecture,
		outputPath:   outputPath,
		capabilities: []string{},
		metadata:     make(map[string]string),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Getters
func (p *Project) ID() string                       { return p.id }
func (p *Project) Name() shared.ProjectName         { return p.name }
func (p *Project) Language() shared.Language        { return p.language }
func (p *Project) ProjectType() shared.ProjectType  { return p.projectType }
func (p *Project) Architecture() shared.Architecture { return p.architecture }
func (p *Project) OutputPath() string               { return p.outputPath }
func (p *Project) Capabilities() []string           { return p.capabilities }
func (p *Project) Metadata() map[string]string      { return p.metadata }
func (p *Project) CreatedAt() time.Time             { return p.createdAt }
func (p *Project) UpdatedAt() time.Time             { return p.updatedAt }

// AddCapability adds a capability to the project
func (p *Project) AddCapability(capability string) error {
	if capability == "" {
		return fmt.Errorf("capability cannot be empty")
	}

	// Check if capability already exists
	for _, c := range p.capabilities {
		if c == capability {
			return fmt.Errorf("capability %s already exists", capability)
		}
	}

	p.capabilities = append(p.capabilities, capability)
	p.updatedAt = time.Now()
	return nil
}

// SetMetadata sets a metadata key-value pair
func (p *Project) SetMetadata(key, value string) {
	if p.metadata == nil {
		p.metadata = make(map[string]string)
	}
	p.metadata[key] = value
	p.updatedAt = time.Now()
}

// GetMetadata retrieves a metadata value by key
func (p *Project) GetMetadata(key string) (string, bool) {
	value, exists := p.metadata[key]
	return value, exists
}

// Validate validates the project entity
func (p *Project) Validate() error {
	if p.id == "" {
		return fmt.Errorf("project id is required")
	}
	if p.outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	// Value objects validate themselves
	return nil
}

// FullPath returns the complete output path for the project
func (p *Project) FullPath() string {
	return fmt.Sprintf("%s/%s", p.outputPath, p.name.Value())
}