package llm

import (
	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
	infrausage "github.com/jorelcb/codify/internal/infrastructure/usage"
	"time"
)

// recordUsage es el shim que llama el provider después de cada GenerateContext.
// Encapsula la creación de la Entry y la propagación a project + global usage.json.
//
// Best-effort — errores se silencian en el recorder, así que llamarla nunca
// puede romper el flujo principal del provider.
func recordUsage(provider, model, command string, inputTokens, outputTokens int, duration time.Duration, success bool) {
	rec := infrausage.NewRecorder(false)
	rec.Record(usagedomain.Entry{
		Command:      command,
		Provider:     provider,
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		DurationMs:   duration.Milliseconds(),
		Success:      success,
	})
}

// commandFromMode mapea el `Mode` interno del provider a un nombre de
// comando útil para usage tracking. Mantiene consistencia con la string
// que ve el usuario en `codify usage`.
func commandFromMode(mode string) string {
	switch mode {
	case "":
		return "generate"
	case "analyze", "spec", "skills", "workflows", "workflow-skills":
		return mode
	case "audit":
		return "audit"
	case "update":
		return "update"
	default:
		return mode
	}
}
