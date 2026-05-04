package command

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
)

// DeliverHooksCommand copies a Claude Code hook bundle (hooks.json + scripts)
// from the embedded catalog to the user's filesystem.
//
// This command is static-only: there is no LLM-personalized variant of hooks.
// The output is a standalone bundle the user merges manually into their
// settings.json — codify does not auto-activate the hooks.
type DeliverHooksCommand struct {
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
	templatesFS      embed.FS
}

// NewDeliverHooksCommand creates a new DeliverHooksCommand.
func NewDeliverHooksCommand(
	fileWriter service.FileWriter,
	directoryManager service.DirectoryManager,
	templatesFS embed.FS,
) *DeliverHooksCommand {
	return &DeliverHooksCommand{
		fileWriter:       fileWriter,
		directoryManager: directoryManager,
		templatesFS:      templatesFS,
	}
}

// Execute writes hooks.json and the auxiliary scripts to config.OutputPath.
//
// Output layout:
//
//	{OutputPath}/
//	├── hooks.json     ← block to merge into settings.json
//	└── hooks/
//	    └── *.sh       ← scripts referenced by hooks.json
//
// For preset == "all", the three preset directories are merged: hooks.json
// objects are unioned per event (PreToolUse, PostToolUse), and every script
// is copied into hooks/.
func (c *DeliverHooksCommand) Execute(config *dto.HookConfig) (*dto.GenerationResult, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if err := c.directoryManager.CreateDir(config.OutputPath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}
	scriptsDir := filepath.Join(config.OutputPath, "hooks")
	if err := c.directoryManager.CreateDir(scriptsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create scripts directory %s: %w", scriptsDir, err)
	}

	presets := []string{config.Preset}
	if config.Preset == "all" {
		presets = []string{"linting", "security-guardrails", "convention-enforcement"}
	}

	// 1. Build merged hooks.json from selected presets.
	mergedHooks, err := c.mergeHooksJSON(config.Locale, presets)
	if err != nil {
		return nil, err
	}
	hooksJSONPath := filepath.Join(config.OutputPath, "hooks.json")
	if err := c.fileWriter.WriteFile(hooksJSONPath, mergedHooks, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", hooksJSONPath, err)
	}

	generated := []string{hooksJSONPath}

	// 2. Copy every .sh script from the selected preset directories.
	for _, p := range presets {
		dir := filepath.Join("templates", config.Locale, "hooks", p)
		entries, err := c.templatesFS.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read template dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sh") {
				continue
			}
			srcPath := filepath.Join(dir, e.Name())
			data, err := c.templatesFS.ReadFile(srcPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read script %s: %w", srcPath, err)
			}
			dstPath := filepath.Join(scriptsDir, e.Name())
			if err := c.fileWriter.WriteFile(dstPath, data, 0o755); err != nil {
				return nil, fmt.Errorf("failed to write script %s: %w", dstPath, err)
			}
			generated = append(generated, dstPath)
		}
	}

	sort.Strings(generated)

	return &dto.GenerationResult{
		OutputPath:     config.OutputPath,
		GeneratedFiles: generated,
		Model:          "static",
	}, nil
}

// mergeHooksJSON reads hooks.json from each preset directory under
// templates/{locale}/hooks/{preset}/ and combines them into a single
// {"hooks": {<event>: [handlers...]}} document.
//
// Merge strategy: per-event arrays are concatenated in preset order
// (linting → security-guardrails → convention-enforcement). No
// deduplication; if two presets define handlers for the same event,
// both run (Claude Code dedupes identical command strings automatically).
func (c *DeliverHooksCommand) mergeHooksJSON(locale string, presets []string) ([]byte, error) {
	type hooksDoc struct {
		Hooks map[string][]any `json:"hooks"`
	}

	merged := hooksDoc{Hooks: map[string][]any{}}

	for _, p := range presets {
		path := filepath.Join("templates", locale, "hooks", p, "hooks.json")
		data, err := c.templatesFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		var doc hooksDoc
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		for event, handlers := range doc.Hooks {
			merged.Hooks[event] = append(merged.Hooks[event], handlers...)
		}
	}

	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged hooks.json: %w", err)
	}
	out = append(out, '\n')
	return out, nil
}
