// Package config provee la persistencia de Codify config a nivel usuario
// (~/.codify/config.yml) y proyecto (.codify/config.yml).
package config

import (
	"errors"
	"os"
	"path/filepath"
)

// codifyDirName es el nombre del directorio de Codify, idéntico tanto en home
// como en proyecto. Se mantiene constante para que `mv` o `git` traten
// consistentemente ambos.
const codifyDirName = ".codify"

// configFileName es el nombre del archivo YAML de configuración.
const configFileName = "config.yml"

// noAutoConfigMarker es el marker file que desactiva el auto-launch SOFT del
// wizard `codify config` en first-run interactivo (ver ADR-007).
const noAutoConfigMarker = ".no-auto-config"

// stateFileName es el nombre del archivo de snapshot de estado por proyecto.
// Se persiste cuando `codify init` corre exitosamente; consumido por los
// lifecycle commands a partir de v1.23 (ADR-004).
const stateFileName = "state.json"

// UserConfigDir devuelve ~/.codify/.
func UserConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, codifyDirName), nil
}

// UserConfigPath devuelve ~/.codify/config.yml. El archivo puede no existir.
func UserConfigPath() (string, error) {
	dir, err := UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// UserNoAutoConfigMarker devuelve ~/.codify/.no-auto-config — su existencia
// indica que el usuario eligió "skip-permanently" en el prompt de auto-launch.
func UserNoAutoConfigMarker() (string, error) {
	dir, err := UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, noAutoConfigMarker), nil
}

// ProjectConfigDir devuelve .codify/ resuelto contra el cwd. La función no
// busca recursivamente hacia arriba — opera estrictamente sobre el cwd actual.
func ProjectConfigDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, codifyDirName), nil
}

// ProjectConfigPath devuelve .codify/config.yml resuelto contra cwd.
func ProjectConfigPath() (string, error) {
	dir, err := ProjectConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// ProjectStatePath devuelve .codify/state.json resuelto contra cwd.
func ProjectStatePath() (string, error) {
	dir, err := ProjectConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

// EnsureUserConfigDir crea ~/.codify/ si no existe (mode 0755).
func EnsureUserConfigDir() error {
	dir, err := UserConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// EnsureProjectConfigDir crea .codify/ si no existe.
func EnsureProjectConfigDir() error {
	dir, err := ProjectConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// FileExists reporta si la ruta apunta a un archivo regular existente.
// Útil para chequear presencia de config sin abrirlo.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		// Otros errores (permisos): tratarlos como "no existe" para que los
		// callers no exploten en flujos de detección.
		return false
	}
	return !info.IsDir()
}
