package command

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/domain/service"
	"github.com/jorelcb/codify/internal/infrastructure/settings"
)

// InstallHooksCommand auto-activates a Claude Code hook bundle by merging it
// into the user's settings.json and copying the auxiliary scripts to the
// agent's hooks directory.
//
// Replaces the v1.19.0 manual-merge ceremony. The user runs
// `codify hooks --install project|global` and the hooks are immediately
// active — no manual settings.json edit required.
//
// The command is idempotent: running it twice with the same preset adds
// zero handlers on the second run (matched by exact command string).
//
// Backups: when settings.json already exists, a timestamped copy is written
// next to it before any modification. The user can rollback with mv.
type InstallHooksCommand struct {
	deliverer    *DeliverHooksCommand
	fileWriter   service.FileWriter
	dirManager   service.DirectoryManager
	loadSettings func(string) (*settings.Settings, error)
	resolveScope func(string) (string, string, error)
}

// InstallResult summarizes everything the install command did. Used by the
// CLI to render a friendly post-install summary, and by the MCP tool to
// build a structured response.
type InstallResult struct {
	SettingsPath    string
	BackupPath      string         // empty when no backup was needed (file did not pre-exist)
	HooksDir        string
	HandlersAdded   map[string]int // by event (PreToolUse, PostToolUse, ...)
	HandlersSkipped map[string]int // by event
	ScriptsCopied   []string       // absolute paths
	ScriptsSkipped  []string       // existed already with identical content
	ScriptsConflict []string       // existed with different content; not overwritten
	DryRun          bool
}

// NewInstallHooksCommand wires the dependencies. settingsLoader and
// scopeResolver are injected so tests can substitute them with tmpdir
// equivalents.
func NewInstallHooksCommand(
	deliverer *DeliverHooksCommand,
	fileWriter service.FileWriter,
	dirManager service.DirectoryManager,
) *InstallHooksCommand {
	return &InstallHooksCommand{
		deliverer:    deliverer,
		fileWriter:   fileWriter,
		dirManager:   dirManager,
		loadSettings: settings.Load,
		resolveScope: settings.ResolveScope,
	}
}

// WithLoader replaces the default settings.Load injection. Used by tests.
func (c *InstallHooksCommand) WithLoader(loader func(string) (*settings.Settings, error)) *InstallHooksCommand {
	c.loadSettings = loader
	return c
}

// WithScopeResolver replaces the default settings.ResolveScope injection.
// Used by tests to redirect "project" / "global" to a temp directory.
func (c *InstallHooksCommand) WithScopeResolver(resolver func(string) (string, string, error)) *InstallHooksCommand {
	c.resolveScope = resolver
	return c
}

// Execute performs the install (or dry run). config.Install must be
// "global" or "project"; otherwise the caller should use DeliverHooksCommand
// directly for preview output.
func (c *InstallHooksCommand) Execute(config *dto.HookConfig) (*InstallResult, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.Install == "" {
		return nil, fmt.Errorf("InstallHooksCommand requires config.Install (global|project)")
	}

	settingsPath, hooksDir, err := c.resolveScope(config.Install)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scope %q: %w", config.Install, err)
	}

	bundle, err := c.deliverer.Build(config.Locale, config.Preset)
	if err != nil {
		return nil, fmt.Errorf("failed to build hook bundle: %w", err)
	}

	// Los templates referencian los scripts vía $CLAUDE_PROJECT_DIR — correcto
	// para scope project. En global, los scripts viven en ~/.claude/hooks/, así
	// que reescribimos las rutas a $HOME para que el handler funcione desde
	// cualquier proyecto activo (incluso uno sin .claude/hooks/ propio).
	if config.Install == dto.InstallScopeGlobal {
		rewriteHookCommandsToHome(bundle.HooksDoc)
	}

	s, err := c.loadSettings(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings %s: %w", settingsPath, err)
	}

	if config.DryRun {
		preview, err := s.PreviewMergedHooks(bundle.HooksDoc)
		if err != nil {
			return nil, fmt.Errorf("dry run preview failed: %w", err)
		}
		// Print the preview directly via the result (printed by the CLI layer).
		return &InstallResult{
			SettingsPath: settingsPath,
			HooksDir:     hooksDir,
			DryRun:       true,
			HandlersAdded: map[string]int{
				"_preview_bytes": len(preview),
			},
		}, nil
	}

	added, skipped, err := s.MergeHooks(bundle.HooksDoc)
	if err != nil {
		return nil, fmt.Errorf("merge into settings.json failed: %w", err)
	}
	backupPath, err := s.Save("")
	if err != nil {
		return nil, fmt.Errorf("save settings.json failed: %w", err)
	}

	if err := c.dirManager.CreateDir(hooksDir, 0o755); err != nil {
		return nil, fmt.Errorf("create hooks dir %s failed: %w", hooksDir, err)
	}

	var copied, alreadyOK, conflicting []string
	for _, sc := range bundle.Scripts {
		dst := filepath.Join(hooksDir, sc.Name)
		existing, statErr := os.ReadFile(dst)
		if statErr == nil {
			if hashEqual(existing, sc.Content) {
				alreadyOK = append(alreadyOK, dst)
				continue
			}
			// Different content already in place — refuse to silently overwrite.
			// The user can delete the existing file or use --output preview to
			// inspect the conflict.
			conflicting = append(conflicting, dst)
			continue
		}
		if !os.IsNotExist(statErr) {
			return nil, fmt.Errorf("stat existing script %s failed: %w", dst, statErr)
		}
		if err := c.fileWriter.WriteFile(dst, sc.Content, 0o755); err != nil {
			return nil, fmt.Errorf("write script %s failed: %w", dst, err)
		}
		copied = append(copied, dst)
	}

	sort.Strings(copied)
	sort.Strings(alreadyOK)
	sort.Strings(conflicting)

	return &InstallResult{
		SettingsPath:    settingsPath,
		BackupPath:      backupPath,
		HooksDir:        hooksDir,
		HandlersAdded:   added,
		HandlersSkipped: skipped,
		ScriptsCopied:   copied,
		ScriptsSkipped:  alreadyOK,
		ScriptsConflict: conflicting,
		DryRun:          false,
	}, nil
}

// rewriteHookCommandsToHome muta in-place el HooksDoc reemplazando
// `"$CLAUDE_PROJECT_DIR"/.claude/hooks/` por `"$HOME"/.claude/hooks/` en cada
// handler. Necesario para scope global: los scripts no están en cada proyecto
// activo, sino en ~/.claude/hooks/.
func rewriteHookCommandsToHome(doc map[string]any) {
	hooks, ok := doc["hooks"].(map[string]any)
	if !ok {
		return
	}
	for _, eventVal := range hooks {
		matchers, ok := eventVal.([]any)
		if !ok {
			continue
		}
		for _, m := range matchers {
			matcher, ok := m.(map[string]any)
			if !ok {
				continue
			}
			handlers, ok := matcher["hooks"].([]any)
			if !ok {
				continue
			}
			for _, h := range handlers {
				handler, ok := h.(map[string]any)
				if !ok {
					continue
				}
				cmd, ok := handler["command"].(string)
				if !ok {
					continue
				}
				handler["command"] = strings.ReplaceAll(
					cmd,
					`"$CLAUDE_PROJECT_DIR"/.claude/hooks/`,
					`"$HOME"/.claude/hooks/`,
				)
			}
		}
	}
}

// hashEqual returns true when both byte slices have the same SHA-256.
// Cheaper to compute than bytes.Equal in the worst case (early exit on
// mismatch is the same), but the explicit hash makes the intent clear:
// "same content".
func hashEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	ha := sha256.Sum256(a)
	hb := sha256.Sum256(b)
	return hex.EncodeToString(ha[:]) == hex.EncodeToString(hb[:])
}
