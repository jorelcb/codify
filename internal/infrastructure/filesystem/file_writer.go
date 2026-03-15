package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/codify/internal/domain/service" // Importar la interfaz
)

// FileWriter handles file writing operations securely
type FileWriter struct{}

// NewFileWriter creates a new instance of FileWriter and returns it as a service.FileWriter interface.
func NewFileWriter() service.FileWriter {
	return &FileWriter{}
}

// WriteFile writes content to a file at the specified path.
// It automatically creates parent directories if they don't exist.
// Returns error if operation fails or permissions are denied.
func (w *FileWriter) WriteFile(path string, content []byte, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, content, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}
