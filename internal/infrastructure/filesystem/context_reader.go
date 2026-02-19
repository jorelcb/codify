package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextReader reads existing context files from a previously generated output directory.
type ContextReader struct{}

// NewContextReader creates a new ContextReader.
func NewContextReader() *ContextReader {
	return &ContextReader{}
}

// ReadExistingContext reads CONTEXT.md (required) and AGENTS.md (optional) from the given
// directory structure and returns their concatenated content.
// It expects the directory layout produced by the generate command:
//
//	<basePath>/AGENTS.md (optional, at root)
//	<basePath>/context/CONTEXT.md (required)
func (r *ContextReader) ReadExistingContext(basePath string) (string, error) {
	// CONTEXT.md is required (in context/ subdirectory)
	contextPath := filepath.Join(basePath, "context", "CONTEXT.md")
	contextContent, err := os.ReadFile(contextPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CONTEXT.md from %s: %w", contextPath, err)
	}

	if strings.TrimSpace(string(contextContent)) == "" {
		return "", fmt.Errorf("CONTEXT.md at %s is empty", contextPath)
	}

	var sb strings.Builder

	// AGENTS.md is optional (at root)
	agentsPath := filepath.Join(basePath, "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err == nil && len(strings.TrimSpace(string(agentsContent))) > 0 {
		sb.WriteString("--- AGENTS.md ---\n")
		sb.WriteString(string(agentsContent))
		sb.WriteString("\n\n")
	}

	sb.WriteString("--- CONTEXT.md ---\n")
	sb.WriteString(string(contextContent))

	return sb.String(), nil
}
