package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorelcb/ai-context-generator/internal/domain/service"
)

// FileSystemTemplateLoader loads template guides from the filesystem.
type FileSystemTemplateLoader struct {
	basePath string
}

// NewFileSystemTemplateLoader creates a new template loader.
func NewFileSystemTemplateLoader(basePath string) service.TemplateLoader {
	return &FileSystemTemplateLoader{basePath: basePath}
}

// templateMapping maps template file names to their guide names.
var templateMapping = map[string]string{
	"prompt.template":       "prompt",
	"context.template":      "context",
	"scaffolding.template":  "scaffolding",
	"interactions.template": "interactions",
}

// LoadAll reads all template files and returns them as TemplateGuides.
func (l *FileSystemTemplateLoader) LoadAll() ([]service.TemplateGuide, error) {
	var guides []service.TemplateGuide

	for filename, name := range templateMapping {
		fullPath := filepath.Join(l.basePath, filename)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", filename, err)
		}

		guides = append(guides, service.TemplateGuide{
			Name:    name,
			Content: string(content),
		})
	}

	return guides, nil
}
