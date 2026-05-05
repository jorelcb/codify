package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	domain "github.com/jorelcb/codify/internal/domain/usage"
)

// Repository implementa lectura/escritura atómica de usage.json.
type Repository struct{}

// NewRepository devuelve una instancia lista para usar.
func NewRepository() *Repository {
	return &Repository{}
}

// Load lee usage.json desde path. Si no existe, devuelve un Log vacío
// (no error) — caller decide si esa ausencia es válida.
func (r *Repository) Load(path string) (domain.Log, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log := domain.NewLog()
			log.StartedAt = time.Now().UTC().Format(time.RFC3339)
			return log, nil
		}
		return domain.Log{}, fmt.Errorf("usage: read %s: %w", path, err)
	}
	var log domain.Log
	if err := json.Unmarshal(data, &log); err != nil {
		return domain.Log{}, fmt.Errorf("usage: parse %s: %w", path, err)
	}
	if log.SchemaVersion == "" {
		log.SchemaVersion = domain.SchemaVersion
	}
	// Re-compute totals defensivamente — protege contra archivos editados a mano.
	log.RecomputeTotals()
	return log, nil
}

// Save escribe el log en path con escritura atómica (tmp + rename).
// Crea el directorio padre si no existe. NO hace backup — usage.json es
// append-only y reproducible (regenerable via codify usage --reset si
// está corrupto).
func (r *Repository) Save(path string, log domain.Log) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("usage: create dir for %s: %w", path, err)
	}
	if log.SchemaVersion == "" {
		log.SchemaVersion = domain.SchemaVersion
	}
	if log.StartedAt == "" {
		log.StartedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(&log, "", "  ")
	if err != nil {
		return fmt.Errorf("usage: marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("usage: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("usage: rename to %s: %w", path, err)
	}
	return nil
}

// Append carga el log en path, agrega entry, y persiste. Atómico vs
// concurrent writes desde el mismo proceso (Save usa rename); NO atómico
// vs procesos paralelos — si dos procesos Codify escriben simultáneamente,
// uno puede sobreescribir al otro. Aceptable para el use case (single-user
// CLI), documentado para callers que conozcan el caveat.
func (r *Repository) Append(path string, entry domain.Entry) error {
	log, err := r.Load(path)
	if err != nil {
		return err
	}
	log.Append(entry)
	return r.Save(path, log)
}

// Reset trunca el log persistido en path, archivando el contenido previo
// como `.bak.<timestamp>` para que el usuario pueda recuperarlo. Útil
// cuando el log creció demasiado o el usuario quiere empezar fresh.
func (r *Repository) Reset(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // nada que resetear
		}
		return err
	}
	bak := fmt.Sprintf("%s.bak.%d", path, time.Now().Unix())
	if err := os.Rename(path, bak); err != nil {
		return fmt.Errorf("usage: backup before reset: %w", err)
	}
	fresh := domain.NewLog()
	fresh.StartedAt = time.Now().UTC().Format(time.RFC3339)
	return r.Save(path, fresh)
}
