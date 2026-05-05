package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	domain "github.com/jorelcb/codify/internal/domain/config"
)

// Repository implementa lectura/escritura atómica de archivos config.yml.
// No mantiene estado interno — todas las operaciones son sobre paths
// resueltos en el momento. Esto facilita el testing y evita sincronización.
type Repository struct{}

// NewRepository devuelve una instancia lista para usar.
func NewRepository() *Repository {
	return &Repository{}
}

// Load lee el archivo YAML en path y devuelve la Config. Si el archivo no
// existe, devuelve (Config zero, false, nil) — el caller decide si esa
// ausencia es válida o no. Errores de I/O o parse devuelven (zero, false, err).
func (r *Repository) Load(path string) (domain.Config, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Config{}, false, nil
		}
		return domain.Config{}, false, fmt.Errorf("config: read %s: %w", path, err)
	}
	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return domain.Config{}, false, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return cfg, true, nil
}

// Save escribe la Config a path en formato YAML. Si el archivo ya existe,
// crea un backup (.bak) antes de sobrescribir. La escritura es atómica
// (.tmp + rename). Si el directorio padre no existe, lo crea.
//
// Mutación side-effect: si cfg.Version está vacío, se setea a SchemaVersion.
// Si CreatedAt está vacío y el archivo no existe, se setea a now. UpdatedAt
// se setea siempre a now.
func (r *Repository) Save(path string, cfg domain.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: create dir for %s: %w", path, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if cfg.Version == "" {
		cfg.Version = domain.SchemaVersion
	}
	existed := FileExists(path)
	if !existed && cfg.CreatedAt == "" {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now

	// Backup si ya existe
	if existed {
		bakPath := path + ".bak"
		if err := copyFile(path, bakPath); err != nil {
			return fmt.Errorf("config: backup %s: %w", path, err)
		}
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("config: write tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("config: rename to %s: %w", path, err)
	}
	return nil
}

// copyFile copia src a dst preservando el contenido. Helper interno para el
// backup atómico antes de overwriting.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// LoadEffective resuelve la Config efectiva aplicando la cadena de precedencia:
// builtin defaults < user < project. Los flags se aplican aparte por el
// caller (no requieren acceso al filesystem).
//
// Si user o project no existen, se omiten silenciosamente (es un caso normal).
// Errores de parse se propagan — un YAML inválido es un bug que el usuario
// debería ver.
func (r *Repository) LoadEffective() (domain.Config, error) {
	cfg := domain.BuiltinDefaults()

	userPath, err := UserConfigPath()
	if err == nil {
		userCfg, ok, err := r.Load(userPath)
		if err != nil {
			return domain.Config{}, fmt.Errorf("config: load user config: %w", err)
		}
		if ok {
			cfg.Merge(userCfg)
		}
	}

	projectPath, err := ProjectConfigPath()
	if err == nil {
		projectCfg, ok, err := r.Load(projectPath)
		if err != nil {
			return domain.Config{}, fmt.Errorf("config: load project config: %w", err)
		}
		if ok {
			cfg.Merge(projectCfg)
		}
	}

	return cfg, nil
}
