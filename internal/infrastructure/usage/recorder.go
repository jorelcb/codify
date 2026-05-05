package usage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	domain "github.com/jorelcb/codify/internal/domain/usage"
)

// Recorder es la API que los providers LLM usan para registrar tokens y
// costos después de cada call. Best-effort: errores de I/O se silencian
// (no rompen el comando que disparó la call).
type Recorder struct {
	repo            *Repository
	disabledByFlag  bool
}

// NewRecorder construye un Recorder. flagDisabled refleja si el usuario
// pasó --no-tracking en esta invocación específica.
func NewRecorder(flagDisabled bool) *Recorder {
	return &Recorder{
		repo:           NewRepository(),
		disabledByFlag: flagDisabled,
	}
}

// Record persiste una Entry en project + global usage.json. Si tracking
// está desactivado (flag/env/marker), es no-op silencioso.
//
// Llamar Record es seguro desde cualquier código path — no necesita ningún
// setup previo, no requiere que existan los archivos.
func (r *Recorder) Record(e domain.Entry) {
	if TrackingDisabled(r.disabledByFlag) {
		return
	}
	// Auto-popular campos derivados que el caller puede no setear
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if e.Project == "" {
		if cwd, err := os.Getwd(); err == nil {
			e.Project = filepath.Base(cwd)
		}
	}
	if e.PricingTableVersion == "" {
		e.PricingTableVersion = domain.PricingTableVersion
	}
	if e.CostUSDCents == 0 && e.Model != "" {
		e.CostUSDCents = domain.CostCents(e.Model, e.InputTokens, e.OutputTokens, e.CacheReadTokens, e.CacheCreationTokens)
	}

	// Project-level (silencia errores; e.g. cwd fuera de un repo Codify es válido)
	if path, err := ProjectUsagePath(); err == nil {
		_ = r.repo.Append(path, e)
	}

	// Global-level (silencia errores; e.g. permisos en HOME)
	if path, err := UserUsagePath(); err == nil {
		_ = r.repo.Append(path, e)
	}
}

// EnvDescribe devuelve un string descriptivo del estado de tracking,
// útil para debugging desde la CLI ("¿por qué no se está registrando?").
func (r *Recorder) EnvDescribe() string {
	if TrackingDisabled(r.disabledByFlag) {
		reasons := []string{}
		if r.disabledByFlag {
			reasons = append(reasons, "--no-tracking flag")
		}
		if os.Getenv("CODIFY_NO_USAGE_TRACKING") == "1" {
			reasons = append(reasons, "CODIFY_NO_USAGE_TRACKING=1")
		}
		if marker, err := UserNoTrackingMarker(); err == nil {
			if _, err := os.Stat(marker); err == nil {
				reasons = append(reasons, fmt.Sprintf("marker %s", marker))
			}
		}
		return fmt.Sprintf("disabled (%s)", joinReasons(reasons))
	}
	return "enabled"
}

func joinReasons(rs []string) string {
	out := ""
	for i, r := range rs {
		if i > 0 {
			out += ", "
		}
		out += r
	}
	return out
}
