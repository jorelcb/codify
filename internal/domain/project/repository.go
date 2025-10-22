package project

// Repository defines the interface for project persistence
// This is a domain interface - implementations will be in infrastructure layer
type Repository interface {
	// FindByID retrieves a project by its ID
	FindByID(id string) (*Project, error)

	// FindByName retrieves a project by its name
	FindByName(name string) (*Project, error)

	// FindAll retrieves all projects
	FindAll() ([]*Project, error)

	// Save persists a project
	Save(project *Project) error

	// Delete removes a project
	Delete(id string) error

	// Exists checks if a project exists
	Exists(id string) (bool, error)

	// ExistsByName checks if a project with given name exists
	ExistsByName(name string) (bool, error)
}