package template

import (
	"fmt"
	"time"
)

// Template represents a template entity in the domain
type Template struct {
	id          string
	name        string
	path        string
	content     string
	variables   []Variable
	metadata    Metadata
	createdAt   time.Time
	updatedAt   time.Time
}

// Variable represents a template variable
type Variable struct {
	Name         string
	DefaultValue string
	Required     bool
	Description  string
}

// Metadata contains template metadata
type Metadata struct {
	Version     string
	Author      string
	Description string
	Tags        []string
}

// NewTemplate creates a new template entity
func NewTemplate(id, name, path, content string) (*Template, error) {
	if id == "" {
		return nil, fmt.Errorf("template id cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}
	if path == "" {
		return nil, fmt.Errorf("template path cannot be empty")
	}

	now := time.Now()
	return &Template{
		id:        id,
		name:      name,
		path:      path,
		content:   content,
		variables: []Variable{},
		metadata:  Metadata{},
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (t *Template) ID() string             { return t.id }
func (t *Template) Name() string           { return t.name }
func (t *Template) Path() string           { return t.path }
func (t *Template) Content() string        { return t.content }
func (t *Template) Variables() []Variable  { return t.variables }
func (t *Template) Metadata() Metadata     { return t.metadata }
func (t *Template) CreatedAt() time.Time   { return t.createdAt }
func (t *Template) UpdatedAt() time.Time   { return t.updatedAt }

// SetContent updates template content
func (t *Template) SetContent(content string) {
	t.content = content
	t.updatedAt = time.Now()
}

// AddVariable adds a variable to the template
func (t *Template) AddVariable(v Variable) error {
	// Check if variable already exists
	for _, existing := range t.variables {
		if existing.Name == v.Name {
			return fmt.Errorf("variable %s already exists", v.Name)
		}
	}
	t.variables = append(t.variables, v)
	t.updatedAt = time.Now()
	return nil
}

// SetMetadata updates template metadata
func (t *Template) SetMetadata(metadata Metadata) {
	t.metadata = metadata
	t.updatedAt = time.Now()
}

// Validate validates the template
func (t *Template) Validate() error {
	if t.id == "" {
		return fmt.Errorf("template id is required")
	}
	if t.name == "" {
		return fmt.Errorf("template name is required")
	}
	if t.path == "" {
		return fmt.Errorf("template path is required")
	}
	// Additional validation logic can be added here
	return nil
}