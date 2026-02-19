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
	mapping  map[string]string
}

// templateMapping maps template file names to their guide names (generate command).
var templateMapping = map[string]string{
	"agents.template":       "agents",
	"context.template":      "context",
	"interactions.template": "interactions",
}

// NewFileSystemTemplateLoader creates a new template loader with the default mapping.
func NewFileSystemTemplateLoader(basePath string) service.TemplateLoader {
	return &FileSystemTemplateLoader{basePath: basePath, mapping: templateMapping}
}

// NewFileSystemTemplateLoaderWithMapping creates a template loader with a custom mapping.
func NewFileSystemTemplateLoaderWithMapping(basePath string, mapping map[string]string) service.TemplateLoader {
	return &FileSystemTemplateLoader{basePath: basePath, mapping: mapping}
}

// LoadAll reads all template files and returns them as TemplateGuides.
func (l *FileSystemTemplateLoader) LoadAll() ([]service.TemplateGuide, error) {
	var guides []service.TemplateGuide

	for filename, name := range l.mapping {
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
