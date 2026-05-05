// Package usage implementa la persistencia + recording de tracking LLM.
//
// Layout:
//   .codify/usage.json   — registro per-project (NO commiteado, ver gitignore)
//   ~/.codify/usage.json — registro global del usuario (acumula todas las
//                          invocaciones de Codify, sin importar el proyecto)
//
// Cada call LLM exitosa O fallida agrega una entry a AMBOS archivos.
package usage

import (
	"os"
	"path/filepath"

	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
)

const usageFileName = "usage.json"

// noTrackingMarker es el marker file que desactiva el tracking de uso global.
// Si existe, los providers NO escriben a ~/.codify/usage.json (project sigue
// si está dentro de un proyecto, salvo CODIFY_NO_USAGE_TRACKING).
const noTrackingMarker = ".no-usage-tracking"

// UserUsagePath devuelve ~/.codify/usage.json. El archivo puede no existir.
func UserUsagePath() (string, error) {
	dir, err := infraconfig.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, usageFileName), nil
}

// ProjectUsagePath devuelve .codify/usage.json relativo al cwd.
func ProjectUsagePath() (string, error) {
	dir, err := infraconfig.ProjectConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, usageFileName), nil
}

// UserNoTrackingMarker devuelve ~/.codify/.no-usage-tracking. Su existencia
// desactiva el tracking persistente.
func UserNoTrackingMarker() (string, error) {
	dir, err := infraconfig.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, noTrackingMarker), nil
}

// TrackingDisabled retorna true si cualquiera de los tres mecanismos de
// opt-out está activo: env var, marker file, o flag pasado al runtime.
//
// El parámetro flag permite que comandos individuales pasen su flag local
// (e.g. --no-tracking) sin tener que reimplementar la cadena de opt-outs.
func TrackingDisabled(flag bool) bool {
	if flag {
		return true
	}
	if os.Getenv("CODIFY_NO_USAGE_TRACKING") == "1" {
		return true
	}
	marker, err := UserNoTrackingMarker()
	if err == nil {
		if _, err := os.Stat(marker); err == nil {
			return true
		}
	}
	return false
}
