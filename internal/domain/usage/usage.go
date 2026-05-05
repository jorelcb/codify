package usage

// SchemaVersion del archivo usage.json. Append-only entries; rotación
// manual via `codify usage --reset`.
const SchemaVersion = "1.0"

// Entry describe una sola invocación a un LLM, capturada por el recorder
// que se hooks en cada provider call (ver internal/infrastructure/usage).
type Entry struct {
	Timestamp           string `json:"timestamp"`
	Command             string `json:"command"` // generate, analyze, update, audit, etc.
	Provider            string `json:"provider"` // anthropic, gemini
	Model               string `json:"model"`
	InputTokens         int    `json:"input_tokens"`
	OutputTokens        int    `json:"output_tokens"`
	CacheReadTokens     int    `json:"cache_read_tokens,omitempty"`
	CacheCreationTokens int    `json:"cache_creation_tokens,omitempty"`
	CostUSDCents        int    `json:"cost_usd_cents"`
	DurationMs          int64  `json:"duration_ms,omitempty"`
	Project             string `json:"project,omitempty"` // basename(cwd) al momento de la call
	Success             bool   `json:"success"`
	PricingTableVersion string `json:"pricing_table_version,omitempty"`
}

// Totals son los agregados que se recomputan al cargar usage.json — no se
// almacenan a mano para evitar inconsistencias entre entries y totals.
type Totals struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheReadTokens     int `json:"cache_read_tokens,omitempty"`
	CacheCreationTokens int `json:"cache_creation_tokens,omitempty"`
	CostUSDCents        int `json:"cost_usd_cents"`
	Calls               int `json:"calls"`
}

// Log es el contenedor persistido. SchemaVersion + entries + totals.
// Totals se recomputa al cargar el archivo o al agregar nuevas entries.
type Log struct {
	SchemaVersion string  `json:"schema_version"`
	StartedAt     string  `json:"started_at"`
	Entries       []Entry `json:"entries"`
	Totals        Totals  `json:"totals"`
}

// NewLog devuelve un Log vacío con SchemaVersion seteado.
func NewLog() Log {
	return Log{SchemaVersion: SchemaVersion, Entries: []Entry{}}
}

// Append agrega una entry al log y recomputa Totals. Mutation in-place.
func (l *Log) Append(e Entry) {
	l.Entries = append(l.Entries, e)
	l.Totals.InputTokens += e.InputTokens
	l.Totals.OutputTokens += e.OutputTokens
	l.Totals.CacheReadTokens += e.CacheReadTokens
	l.Totals.CacheCreationTokens += e.CacheCreationTokens
	l.Totals.CostUSDCents += e.CostUSDCents
	l.Totals.Calls++
}

// RecomputeTotals re-calcula Totals desde Entries. Se llama al cargar el
// archivo desde disco para garantizar consistencia ante ediciones manuales
// o entries de schemas viejos.
func (l *Log) RecomputeTotals() {
	t := Totals{}
	for _, e := range l.Entries {
		t.InputTokens += e.InputTokens
		t.OutputTokens += e.OutputTokens
		t.CacheReadTokens += e.CacheReadTokens
		t.CacheCreationTokens += e.CacheCreationTokens
		t.CostUSDCents += e.CostUSDCents
		t.Calls++
	}
	l.Totals = t
}

// FilteredTotals computa Totals sobre un subset de entries. Útil para
// reportes con --by command, --by model, --since, etc.
func FilteredTotals(entries []Entry) Totals {
	t := Totals{}
	for _, e := range entries {
		t.InputTokens += e.InputTokens
		t.OutputTokens += e.OutputTokens
		t.CacheReadTokens += e.CacheReadTokens
		t.CacheCreationTokens += e.CacheCreationTokens
		t.CostUSDCents += e.CostUSDCents
		t.Calls++
	}
	return t
}
