package template

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/jorelcb/codify/internal/domain/service"
)

// FileSystemTemplateLoader loads template guides from an fs.FS (embedded or OS filesystem).
type FileSystemTemplateLoader struct {
	fsys       fs.FS
	basePath   string
	mapping    map[string]string
	language   string
	localeBase string // e.g., "templates/en" — used to resolve language-specific templates
}

// templateMapping maps template file names to their guide names (generate command).
var templateMapping = map[string]string{
	"agents.template":            "agents",
	"context.template":           "context",
	"interactions.template":      "interactions",
	"development_guide.template": "development_guide",
}

// languageTemplateMapping maps language-specific template files to guide names.
var languageTemplateMapping = map[string]string{
	"idioms.template": "idioms",
}

// NewFileSystemTemplateLoader creates a new template loader with the default mapping.
func NewFileSystemTemplateLoader(fsys fs.FS, basePath string) service.TemplateLoader {
	return &FileSystemTemplateLoader{fsys: fsys, basePath: basePath, mapping: templateMapping}
}

// NewFileSystemTemplateLoaderWithLanguage creates a template loader that also loads
// language-specific templates from {localeBase}/languages/{language}/.
func NewFileSystemTemplateLoaderWithLanguage(fsys fs.FS, basePath string, localeBase string, language string) service.TemplateLoader {
	return &FileSystemTemplateLoader{fsys: fsys, basePath: basePath, mapping: templateMapping, localeBase: localeBase, language: language}
}

// NewFileSystemTemplateLoaderWithMapping creates a template loader with a custom mapping.
func NewFileSystemTemplateLoaderWithMapping(fsys fs.FS, basePath string, mapping map[string]string) service.TemplateLoader {
	return &FileSystemTemplateLoader{fsys: fsys, basePath: basePath, mapping: mapping}
}

// LoadAll reads all template files and returns them as TemplateGuides.
// If a language is configured, it also loads language-specific templates.
func (l *FileSystemTemplateLoader) LoadAll() ([]service.TemplateGuide, error) {
	guides, err := l.loadFromMapping(l.basePath, l.mapping)
	if err != nil {
		return nil, err
	}

	if l.language != "" {
		langGuides, err := l.loadLanguageTemplates()
		if err != nil {
			return nil, err
		}
		guides = append(guides, langGuides...)
	}

	return guides, nil
}

// loadFromMapping reads template files from a directory using the given mapping.
func (l *FileSystemTemplateLoader) loadFromMapping(basePath string, mapping map[string]string) ([]service.TemplateGuide, error) {
	var guides []service.TemplateGuide

	for filename, name := range mapping {
		fullPath := filepath.Join(basePath, filename)

		content, err := fs.ReadFile(l.fsys, fullPath)
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

// loadLanguageTemplates loads templates from {localeBase}/languages/{language}/ if they exist.
func (l *FileSystemTemplateLoader) loadLanguageTemplates() ([]service.TemplateGuide, error) {
	langDir := filepath.Join(l.localeBase, "languages", l.language)

	_, err := fs.Stat(l.fsys, langDir)
	if err != nil {
		return nil, nil // Language directory doesn't exist, skip silently
	}

	return l.loadFromMapping(langDir, languageTemplateMapping)
}
