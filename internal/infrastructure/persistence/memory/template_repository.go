package memory

import (
	"fmt"
	"sync"

	"github.com/jorelcb/ai-context-generator/internal/domain/template"
)

// TemplateRepository is an in-memory implementation of template.Repository
type TemplateRepository struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	byPath    map[string]*template.Template
}

// NewTemplateRepository creates a new in-memory template repository
func NewTemplateRepository() *TemplateRepository {
	return &TemplateRepository{
		templates: make(map[string]*template.Template),
		byPath:    make(map[string]*template.Template),
	}
}

// FindByID retrieves a template by its ID
func (r *TemplateRepository) FindByID(id string) (*template.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tmpl, exists := r.templates[id]
	if !exists {
		return nil, fmt.Errorf("template with id %s not found", id)
	}
	return tmpl, nil
}

// FindByPath retrieves a template by its file path
func (r *TemplateRepository) FindByPath(path string) (*template.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tmpl, exists := r.byPath[path]
	if !exists {
		return nil, fmt.Errorf("template with path %s not found", path)
	}
	return tmpl, nil
}

// FindAll retrieves all templates
func (r *TemplateRepository) FindAll() ([]*template.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]*template.Template, 0, len(r.templates))
	for _, tmpl := range r.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// FindByTag retrieves templates by tag
func (r *TemplateRepository) FindByTag(tag string) ([]*template.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]*template.Template, 0)
	for _, tmpl := range r.templates {
		metadata := tmpl.Metadata()
		for _, t := range metadata.Tags {
			if t == tag {
				templates = append(templates, tmpl)
				break
			}
		}
	}
	return templates, nil
}

// Save persists a template
func (r *TemplateRepository) Save(tmpl *template.Template) error {
	if tmpl == nil {
		return fmt.Errorf("template cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate template before saving
	if err := tmpl.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	r.templates[tmpl.ID()] = tmpl
	r.byPath[tmpl.Path()] = tmpl
	return nil
}

// Delete removes a template
func (r *TemplateRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tmpl, exists := r.templates[id]
	if !exists {
		return fmt.Errorf("template with id %s not found", id)
	}

	delete(r.templates, id)
	delete(r.byPath, tmpl.Path())
	return nil
}

// Exists checks if a template exists
func (r *TemplateRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.templates[id]
	return exists, nil
}

// Clear removes all templates (useful for testing)
func (r *TemplateRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates = make(map[string]*template.Template)
	r.byPath = make(map[string]*template.Template)
}

// Count returns the number of templates
func (r *TemplateRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.templates)
}