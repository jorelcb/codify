package service

import "os"

// FileWriter defines the interface for writing files.
// Implementations should handle creating parent directories if necessary.
type FileWriter interface {
	WriteFile(path string, content []byte, perm os.FileMode) error
}

// DirectoryManager defines the interface for managing directories.
// Implementations should handle creating parent directories if necessary and checking existence.
type DirectoryManager interface {
	CreateDir(path string, perm os.FileMode) error
	Exists(path string) (bool, error)
}