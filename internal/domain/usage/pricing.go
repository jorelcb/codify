// Package usage define el modelo de tracking de uso LLM persistido en
// .codify/usage.json (proyecto) y ~/.codify/usage.json (global).
//
// Schema y semántica documentados en docs/adr/0005-llm-usage-tracking.md.
package usage

// PricingTableVersion identifica el set de precios embebido. Se incluye en
// cada Entry para que reportes históricos sean reproducibles aún cuando los
// precios del proveedor cambien en versiones futuras.
const PricingTableVersion = "2026-05"

// Pricing describe los costos de un modelo en centavos USD por 1M tokens.
// Los valores reflejan public list prices al momento de PricingTableVersion.
// Descuentos negociados por el usuario NO se reflejan acá — el report es
// "lista de precios", no "factura real". Documentado explícitamente en
// docs/adr/0005-llm-usage-tracking.md.
type Pricing struct {
	InputPerMillionCents        int // por 1M tokens de input (no-cache)
	OutputPerMillionCents       int // por 1M tokens de output
	CacheReadPerMillionCents    int // por 1M tokens leídos desde cache (típicamente 10% de input)
	CacheCreationPerMillionCents int // por 1M tokens escritos a cache (típicamente 125% de input)
}

// pricingTable es el mapa modelo → Pricing. Las claves son los model IDs
// que aparecen en config/Anthropic/Gemini SDK responses.
//
// Fuentes (PricingTableVersion = 2026-05):
//   - Anthropic: https://www.anthropic.com/pricing  (Claude Sonnet 4.6, Opus 4.6)
//   - Google: https://ai.google.dev/pricing         (Gemini 3.1 Pro Preview)
//
// Si el modelo no está en la tabla, costPerCall devuelve 0 — registrar
// tokens sin costo para no generar números fantasma.
var pricingTable = map[string]Pricing{
	// Anthropic Claude Sonnet 4.6 — input $3/M, output $15/M, cache read $0.30/M, cache write $3.75/M
	"claude-sonnet-4-6": {
		InputPerMillionCents:        300,
		OutputPerMillionCents:       1500,
		CacheReadPerMillionCents:    30,
		CacheCreationPerMillionCents: 375,
	},
	// Anthropic Claude Opus 4.6 — input $15/M, output $75/M, cache read $1.50/M, cache write $18.75/M
	"claude-opus-4-6": {
		InputPerMillionCents:        1500,
		OutputPerMillionCents:       7500,
		CacheReadPerMillionCents:    150,
		CacheCreationPerMillionCents: 1875,
	},
	// Anthropic Claude Opus 4.7 — same tier as 4.6 at time of pricing snapshot
	"claude-opus-4-7": {
		InputPerMillionCents:        1500,
		OutputPerMillionCents:       7500,
		CacheReadPerMillionCents:    150,
		CacheCreationPerMillionCents: 1875,
	},
	// Google Gemini 3.1 Pro Preview — input $1.25/M, output $5/M (no published cache tier)
	"gemini-3.1-pro-preview": {
		InputPerMillionCents:  125,
		OutputPerMillionCents: 500,
		// Cache read/creation: not exposed via public pricing for preview tier;
		// treated as no-cache. If pricing details surface, update here.
	},
}

// PriceFor devuelve el Pricing del modelo, o el zero Pricing si no está
// catalogado (en cuyo caso el costo computado será 0).
func PriceFor(model string) Pricing {
	return pricingTable[model]
}

// CostCents calcula el costo en centavos USD de una invocación dada los
// conteos de tokens. Si el modelo no está en la tabla, devuelve 0 — el
// caller puede inspeccionar PricingTableVersion para documentarlo.
func CostCents(model string, inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int) int {
	p := PriceFor(model)
	cents := 0
	cents += inputTokens * p.InputPerMillionCents / 1_000_000
	cents += outputTokens * p.OutputPerMillionCents / 1_000_000
	cents += cacheReadTokens * p.CacheReadPerMillionCents / 1_000_000
	cents += cacheCreationTokens * p.CacheCreationPerMillionCents / 1_000_000
	return cents
}

// ListModels devuelve los model IDs catalogados, en orden estable. Útil
// para tests y documentación.
func ListModels() []string {
	names := make([]string, 0, len(pricingTable))
	for k := range pricingTable {
		names = append(names, k)
	}
	// Orden alfabético para output reproducible
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return names
}
