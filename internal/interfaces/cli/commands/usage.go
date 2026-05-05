package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
	infrausage "github.com/jorelcb/codify/internal/infrastructure/usage"
)

// NewUsageCmd construye `codify usage` — reporte de tracking de uso LLM.
//
// Default: imprime el report del proyecto actual (.codify/usage.json).
// Con --global: imprime el report acumulado del usuario (~/.codify/usage.json).
//
// El comando NUNCA invoca un LLM ni hace red — es lectura pura del archivo
// JSON local. Por eso es seguro correrlo en cualquier momento sin costo.
func NewUsageCmd() *cobra.Command {
	var (
		global    bool
		since     string
		byField   string
		jsonOut   bool
		reset     bool
	)

	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Report LLM usage and cost tracking from .codify/usage.json or ~/.codify/usage.json",
		Long: `Report LLM token usage and cost from local tracking files. Default scope
is the current project (.codify/usage.json); pass --global for the user-level
aggregate (~/.codify/usage.json).

The report includes total tokens (input/output/cache), call count, and cost
in USD cents. Costs use the embedded pricing table at PricingTableVersion;
they reflect public list prices and may not match your actual invoice if
you have negotiated discounts.

Tracking can be disabled per-invocation with --no-tracking, persistently
via CODIFY_NO_USAGE_TRACKING=1, or globally by creating
~/.codify/.no-usage-tracking. When disabled, no entries are recorded, so
this command will report zero.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsage(global, since, byField, jsonOut, reset)
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Report global user-level usage (~/.codify/usage.json) instead of project-level")
	cmd.Flags().StringVar(&since, "since", "", "Filter entries by minimum age, e.g. 7d, 24h, 30m")
	cmd.Flags().StringVar(&byField, "by", "", "Group totals by field: command, model, provider")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit raw JSON instead of human-readable report")
	cmd.Flags().BoolVar(&reset, "reset", false, "Archive the current usage.json (with timestamp) and start fresh")
	return cmd
}

func runUsage(global bool, since, byField string, jsonOut, reset bool) error {
	repo := infrausage.NewRepository()

	path, err := resolveUsagePath(global)
	if err != nil {
		return err
	}

	if reset {
		if err := repo.Reset(path); err != nil {
			return fmt.Errorf("usage reset: %w", err)
		}
		fmt.Printf("✓ Reset %s (previous log archived as .bak.<timestamp>)\n", path)
		return nil
	}

	log, err := repo.Load(path)
	if err != nil {
		return err
	}

	entries := log.Entries
	if since != "" {
		filtered, err := filterSince(entries, since)
		if err != nil {
			return err
		}
		entries = filtered
	}

	if jsonOut {
		emitUsageJSON(entries, log.SchemaVersion)
		return nil
	}

	emitUsageHuman(path, entries, byField, global)
	return nil
}

func resolveUsagePath(global bool) (string, error) {
	if global {
		return infrausage.UserUsagePath()
	}
	return infrausage.ProjectUsagePath()
}

// filterSince acepta duraciones como "7d", "24h", "30m" y devuelve solo las
// entries cuyo timestamp es posterior a (now - duration).
func filterSince(entries []usagedomain.Entry, since string) ([]usagedomain.Entry, error) {
	dur, err := parseSinceDuration(since)
	if err != nil {
		return nil, fmt.Errorf("--since %q: %w", since, err)
	}
	cutoff := time.Now().Add(-dur)
	out := make([]usagedomain.Entry, 0, len(entries))
	for _, e := range entries {
		ts, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue // entry sin timestamp parseable: skip
		}
		if ts.After(cutoff) {
			out = append(out, e)
		}
	}
	return out, nil
}

// parseSinceDuration acepta "7d", "24h", "30m" además de los unidades nativas
// de time.ParseDuration. Días no son nativos, así que los expandimos a horas.
func parseSinceDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		var days int
		_, err := fmt.Sscanf(s, "%dd", &days)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func emitUsageHuman(path string, entries []usagedomain.Entry, byField string, global bool) {
	scope := "project"
	if global {
		scope = "global"
	}

	total := usagedomain.FilteredTotals(entries)
	fmt.Printf("Codify Usage — %s scope (%s)\n", scope, path)
	fmt.Println(strings.Repeat("═", 60))
	if total.Calls == 0 {
		fmt.Println("No usage recorded yet.")
		return
	}
	fmt.Printf("Total cost:     $%s (%d cents)\n", formatDollars(total.CostUSDCents), total.CostUSDCents)
	fmt.Printf("Total calls:    %d\n", total.Calls)
	fmt.Printf("Total input:    %s tokens\n", humanInt(total.InputTokens))
	fmt.Printf("Total output:   %s tokens\n", humanInt(total.OutputTokens))
	if total.CacheReadTokens > 0 || total.CacheCreationTokens > 0 {
		fmt.Printf("Cache read:     %s tokens\n", humanInt(total.CacheReadTokens))
		fmt.Printf("Cache write:    %s tokens\n", humanInt(total.CacheCreationTokens))
		hitRate := 0
		if denom := total.InputTokens + total.CacheReadTokens; denom > 0 {
			hitRate = 100 * total.CacheReadTokens / denom
		}
		fmt.Printf("Cache hit rate: %d%%\n", hitRate)
	}

	if byField != "" {
		fmt.Println()
		fmt.Printf("By %s:\n", byField)
		groups := groupBy(entries, byField)
		// Sort groups by cost desc
		keys := make([]string, 0, len(groups))
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return groups[keys[i]].CostUSDCents > groups[keys[j]].CostUSDCents
		})
		for _, k := range keys {
			t := groups[k]
			fmt.Printf("  %-25s $%s   %d calls\n", k, formatDollars(t.CostUSDCents), t.Calls)
		}
	}
}

func emitUsageJSON(entries []usagedomain.Entry, schemaVersion string) {
	type out struct {
		SchemaVersion string                  `json:"schema_version"`
		Entries       []usagedomain.Entry     `json:"entries"`
		Totals        usagedomain.Totals      `json:"totals"`
	}
	encodeJSON(out{
		SchemaVersion: schemaVersion,
		Entries:       entries,
		Totals:        usagedomain.FilteredTotals(entries),
	})
}

func groupBy(entries []usagedomain.Entry, field string) map[string]usagedomain.Totals {
	groups := map[string][]usagedomain.Entry{}
	for _, e := range entries {
		var key string
		switch field {
		case "command":
			key = e.Command
		case "model":
			key = e.Model
		case "provider":
			key = e.Provider
		default:
			key = "unknown"
		}
		if key == "" {
			key = "(unset)"
		}
		groups[key] = append(groups[key], e)
	}
	totals := map[string]usagedomain.Totals{}
	for k, es := range groups {
		totals[k] = usagedomain.FilteredTotals(es)
	}
	return totals
}

func formatDollars(cents int) string {
	return fmt.Sprintf("%.2f", float64(cents)/100.0)
}

// humanInt formatea un int como "1.2K", "47M", etc. para tokens grandes.
func humanInt(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000.0)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000.0)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// silence unused warning for os import ya que se usa indirecto via repo
var _ = os.Stderr
