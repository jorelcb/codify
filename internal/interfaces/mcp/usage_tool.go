package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	usagedomain "github.com/jorelcb/codify/internal/domain/usage"
	infrausage "github.com/jorelcb/codify/internal/infrastructure/usage"
)

// getUsageTool define el MCP tool `get_usage`. Permite que un agente externo
// (e.g. Claude Code) consulte el tracking de uso LLM persistido localmente
// sin abrir el archivo manualmente.
//
// El tool es read-only: solo lee usage.json — nunca lo modifica. No invoca
// LLM, no consume API keys.
func getUsageTool() server.ServerTool {
	tool := mcp.NewTool("get_usage",
		mcp.WithDescription("Read LLM usage tracking from local .codify/usage.json (project) or ~/.codify/usage.json (global). Returns totals plus optional grouping/filtering. Read-only; no LLM call, no cost."),
		mcp.WithString("scope", mcp.Description("Which file to read"), mcp.Enum("project", "global"), mcp.DefaultString("project")),
		mcp.WithString("since", mcp.Description("Time filter, e.g. \"7d\", \"24h\", \"30m\"")),
		mcp.WithString("by", mcp.Description("Group totals by field"), mcp.Enum("command", "model", "provider", "")),
	)

	return server.ServerTool{Tool: tool, Handler: handleGetUsage}
}

func handleGetUsage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := stringArgDefault(request, "scope", "project")
	since := stringArg(request, "since")
	byField := stringArg(request, "by")

	repo := infrausage.NewRepository()
	var path string
	var err error
	if scope == "global" {
		path, err = infrausage.UserUsagePath()
	} else {
		path, err = infrausage.ProjectUsagePath()
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("resolve path: %v", err)), nil
	}

	log, err := repo.Load(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("load usage: %v", err)), nil
	}

	entries := log.Entries
	if since != "" {
		filtered, err := filterUsageEntriesSince(entries, since)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("--since %q: %v", since, err)), nil
		}
		entries = filtered
	}

	out := map[string]interface{}{
		"scope":  scope,
		"path":   path,
		"totals": usagedomain.FilteredTotals(entries),
	}

	if byField != "" {
		out["groups"] = groupUsageBy(entries, byField)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// filterUsageEntriesSince — duplica la lógica de cli/commands para evitar
// import-cycle entre mcp y cli.
func filterUsageEntriesSince(entries []usagedomain.Entry, since string) ([]usagedomain.Entry, error) {
	dur, err := parseUsageSinceDur(since)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-dur)
	out := make([]usagedomain.Entry, 0, len(entries))
	for _, e := range entries {
		ts, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}
		if ts.After(cutoff) {
			out = append(out, e)
		}
	}
	return out, nil
}

func parseUsageSinceDur(s string) (time.Duration, error) {
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

func groupUsageBy(entries []usagedomain.Entry, field string) map[string]usagedomain.Totals {
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
