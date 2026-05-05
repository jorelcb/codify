// Package state implementa la persistencia atómica de .codify/state.json.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	domain "github.com/jorelcb/codify/internal/domain/state"
)

// Repository es la API de lectura/escritura de state.json.
type Repository struct{}

// NewRepository devuelve una instancia lista para usar.
func NewRepository() *Repository {
	return &Repository{}
}

// Save escribe state a path en formato JSON con indentación. Atómico vía
// .tmp + rename. Crea el directorio padre si no existe. NO hace backup
// porque state.json es regenerable (a diferencia de config.yml).
func (r *Repository) Save(path string, state domain.State) error {
	if state.SchemaVersion == "" {
		state.SchemaVersion = domain.SchemaVersion
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("state: create dir for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(&state, "", "  ")
	if err != nil {
		return fmt.Errorf("state: marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("state: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("state: rename to %s: %w", path, err)
	}
	return nil
}

// Load lee state desde path. Si no existe, devuelve (zero, false, nil).
func (r *Repository) Load(path string) (domain.State, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.State{}, false, nil
		}
		return domain.State{}, false, fmt.Errorf("state: read %s: %w", path, err)
	}
	var s domain.State
	if err := json.Unmarshal(data, &s); err != nil {
		return domain.State{}, false, fmt.Errorf("state: parse %s: %w", path, err)
	}
	return s, true, nil
}
