package template

// Repository defines the interface for template persistence
// This is a domain interface - implementations will be in infrastructure layer
type Repository interface {
	// FindByID retrieves a template by its ID
	FindByID(id string) (*Template, error)

	// FindByPath retrieves a template by its file path
	FindByPath(path string) (*Template, error)

	// FindAll retrieves all templates
	FindAll() ([]*Template, error)

	// FindByTag retrieves templates by tag
	FindByTag(tag string) ([]*Template, error)

	// Save persists a template
	Save(template *Template) error

	// Delete removes a template
	Delete(id string) error

	// Exists checks if a template exists
	Exists(id string) (bool, error)
}