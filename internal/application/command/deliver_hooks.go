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
// It supports two flows:
//
//  1. Preview mode (Execute): writes a standalone bundle to OutputPath that
//     the user can inspect or merge manually. This is the v1.19.0 default
//     behavior, retained as an escape hatch.
//
//  2. In-memory build (Build): returns the merged hooks JSON block and the
//     list of script files without touching the filesystem, so the caller
//     (InstallHooksCommand) can merge into settings.json + copy scripts to
//     the agent's hooks directory.
type DeliverHooksCommand struct {
	fileWriter       service.FileWriter
	directoryManager service.DirectoryManager
	templatesFS      embed.FS
}

// HookScript is a single .sh script referenced by hooks.json, kept in-memory
// for callers that want to control where it lands on disk.
type HookScript struct {
	Name    string // filename without directory, e.g. "lint.sh"
	Content []byte
}

// HookBundle is the in-memory representation of a hook preset (or merged
// "all" preset): the parsed hooks.json document plus every script the
// document references.
type HookBundle struct {
	HooksDoc map[string]any // {"hooks": {<event>: [matchers...]}}
	Scripts  []HookScript
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

// Build assembles the in-memory hook bundle for the requested preset(s).
// It does not write to disk and is the entry point used by
// InstallHooksCommand to obtain the merge block + scripts.
func (c *DeliverHooksCommand) Build(locale, preset string) (*HookBundle, error) {
	presets := []string{preset}
	if preset == "all" {
		presets = []string{"linting", "security-guardrails", "convention-enforcement"}
	}

	mergedBytes, err := c.mergeHooksJSON(locale, presets)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(mergedBytes, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse merged hooks document: %w", err)
	}

	var scripts []HookScript
	for _, p := range presets {
		dir := filepath.Join("templates", locale, "hooks", p)
		entries, err := c.templatesFS.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read template dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sh") {
				continue
			}
			data, err := c.templatesFS.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read script %s: %w", e.Name(), err)
			}
			scripts = append(scripts, HookScript{Name: e.Name(), Content: data})
		}
	}

	return &HookBundle{HooksDoc: doc, Scripts: scripts}, nil
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
//
// This is the preview/escape-hatch flow. The default user flow is
// auto-activation via InstallHooksCommand. Execute requires
// config.OutputPath to be set.
func (c *DeliverHooksCommand) Execute(config *dto.HookConfig) (*dto.GenerationResult, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.OutputPath == "" {
		return nil, fmt.Errorf("DeliverHooksCommand requires OutputPath; use InstallHooksCommand for auto-activation")
	}

	if err := c.directoryManager.CreateDir(config.OutputPath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", config.OutputPath, err)
	}
	scriptsDir := filepath.Join(config.OutputPath, "hooks")
	if err := c.directoryManager.CreateDir(scriptsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create scripts directory %s: %w", scriptsDir, err)
	}

	bundle, err := c.Build(config.Locale, config.Preset)
	if err != nil {
		return nil, err
	}

	hooksJSONPath := filepath.Join(config.OutputPath, "hooks.json")
	mergedBytes, err := json.MarshalIndent(bundle.HooksDoc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hooks.json: %w", err)
	}
	mergedBytes = append(mergedBytes, '\n')
	if err := c.fileWriter.WriteFile(hooksJSONPath, mergedBytes, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", hooksJSONPath, err)
	}

	generated := []string{hooksJSONPath}
	for _, sc := range bundle.Scripts {
		dst := filepath.Join(scriptsDir, sc.Name)
		if err := c.fileWriter.WriteFile(dst, sc.Content, 0o755); err != nil {
			return nil, fmt.Errorf("failed to write script %s: %w", dst, err)
		}
		generated = append(generated, dst)
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
