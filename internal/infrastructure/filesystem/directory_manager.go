package filesystem

import (
	"fmt"
	"os"

	"github.com/jorelcb/codify/internal/domain/service" // Importar la interfaz
)

// DirectoryManager handles directory operations
type DirectoryManager struct{}

// NewDirectoryManager creates a new instance of DirectoryManager and returns it as a service.DirectoryManager interface.
func NewDirectoryManager() service.DirectoryManager {
	return &DirectoryManager{}
}

// CreateDir creates a directory and any necessary parents.
// Uses os.MkdirAll internally.
func (d *DirectoryManager) CreateDir(path string, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// Exists checks if a path exists in the filesystem.
// Returns true if exists, false if not (or error).
func (d *DirectoryManager) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check path existence %s: %w", path, err)
}
