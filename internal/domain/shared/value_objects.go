package shared

import "fmt"

// Language represents a programming language value object
type Language struct {
	value string
}

// Valid languages
var validLanguages = map[string]bool{
	"go":         true,
	"javascript": true,
	"typescript": true,
	"python":     true,
	"java":       true,
	"rust":       true,
	"csharp":     true,
	"php":        true,
	"ruby":       true,
}

// NewLanguage creates a new Language value object
func NewLanguage(value string) (Language, error) {
	if value == "" {
		return Language{}, fmt.Errorf("language cannot be empty")
	}
	if !validLanguages[value] {
		return Language{}, fmt.Errorf("invalid language: %s", value)
	}
	return Language{value: value}, nil
}

func (l Language) String() string { return l.value }
func (l Language) Value() string  { return l.value }

// ProjectType represents a project type value object
type ProjectType struct {
	value string
}

// Valid project types
var validProjectTypes = map[string]bool{
	"api":         true,
	"cli":         true,
	"library":     true,
	"microservice": true,
	"monolith":    true,
	"webapp":      true,
	"mobile":      true,
	"desktop":     true,
}

// NewProjectType creates a new ProjectType value object
func NewProjectType(value string) (ProjectType, error) {
	if value == "" {
		return ProjectType{}, fmt.Errorf("project type cannot be empty")
	}
	if !validProjectTypes[value] {
		return ProjectType{}, fmt.Errorf("invalid project type: %s", value)
	}
	return ProjectType{value: value}, nil
}

func (p ProjectType) String() string { return p.value }
func (p ProjectType) Value() string  { return p.value }

// Architecture represents an architecture pattern value object
type Architecture struct {
	value string
}

// Valid architectures
var validArchitectures = map[string]bool{
	"ddd":        true,
	"clean":      true,
	"hexagonal":  true,
	"layered":    true,
	"onion":      true,
	"mvc":        true,
	"mvvm":       true,
	"cqrs":       true,
	"eventdriven": true,
}

// NewArchitecture creates a new Architecture value object
func NewArchitecture(value string) (Architecture, error) {
	if value == "" {
		return Architecture{}, fmt.Errorf("architecture cannot be empty")
	}
	if !validArchitectures[value] {
		return Architecture{}, fmt.Errorf("invalid architecture: %s", value)
	}
	return Architecture{value: value}, nil
}

func (a Architecture) String() string { return a.value }
func (a Architecture) Value() string  { return a.value }

// ProjectName represents a project name value object
type ProjectName struct {
	value string
}

// NewProjectName creates a new ProjectName value object
func NewProjectName(value string) (ProjectName, error) {
	if value == "" {
		return ProjectName{}, fmt.Errorf("project name cannot be empty")
	}
	// Add validation rules (e.g., no spaces, alphanumeric + hyphens/underscores)
	if len(value) < 2 {
		return ProjectName{}, fmt.Errorf("project name must be at least 2 characters")
	}
	if len(value) > 100 {
		return ProjectName{}, fmt.Errorf("project name must be less than 100 characters")
	}
	return ProjectName{value: value}, nil
}

func (p ProjectName) String() string { return p.value }
func (p ProjectName) Value() string  { return p.value }

// ProjectDescription represents a project description value object
type ProjectDescription struct {
	value string
}

// NewProjectDescription creates a new ProjectDescription value object
func NewProjectDescription(value string) (ProjectDescription, error) {
	if value == "" {
		return ProjectDescription{}, fmt.Errorf("project description cannot be empty")
	}
	if len(value) < 10 {
		return ProjectDescription{}, fmt.Errorf("project description must be at least 10 characters")
	}
	if len(value) > 10000 {
		return ProjectDescription{}, fmt.Errorf("project description must be less than 10000 characters")
	}
	return ProjectDescription{value: value}, nil
}

func (d ProjectDescription) String() string { return d.value }
func (d ProjectDescription) Value() string  { return d.value }