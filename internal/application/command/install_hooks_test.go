package command

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	root "github.com/jorelcb/codify"
	"github.com/jorelcb/codify/internal/application/dto"
	"github.com/jorelcb/codify/internal/infrastructure/filesystem"
	"github.com/jorelcb/codify/internal/infrastructure/settings"
)

// scopeRedirect returns a resolver that maps any scope to the (settingsPath,
// hooksDir) pair under tmp. Lets tests run install commands without
// touching the real ~/.claude or .claude directories.
func scopeRedirect(tmp string) func(string) (string, string, error) {
	return func(scope string) (string, string, error) {
		settingsPath := filepath.Join(tmp, ".claude", "settings.json")
		hooksDir := filepath.Join(tmp, ".claude", "hooks")
		return settingsPath, hooksDir, nil
	}
}

func newTestInstaller(tmp string) *InstallHooksCommand {
	fw := filesystem.NewFileWriter()
	dm := filesystem.NewDirectoryManager()
	deliverer := NewDeliverHooksCommand(fw, dm, root.TemplatesFS)
	return NewInstallHooksCommand(deliverer, fw, dm).WithScopeResolver(scopeRedirect(tmp))
}

func TestInstallHooks_FreshInstallProject(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "linting",
		Locale:   "en",
		Install:  dto.InstallScopeProject,
	}

	result, err := installer.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if result.SettingsPath != settingsPath {
		t.Fatalf("SettingsPath: got %q, want %q", result.SettingsPath, settingsPath)
	}
	if result.BackupPath != "" {
		t.Fatalf("BackupPath: should be empty for fresh install, got %q", result.BackupPath)
	}
	if total(result.HandlersAdded) == 0 {
		t.Fatal("expected at least one handler added")
	}
	if len(result.ScriptsCopied) == 0 {
		t.Fatal("expected at least one script copied")
	}

	// Verify settings.json has the merged hook block.
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}
	hooks, ok := parsed["hooks"].(map[string]any)
	if !ok {
		t.Fatal("settings.json has no hooks key")
	}
	if _, ok := hooks["PostToolUse"]; !ok {
		t.Fatal("PostToolUse missing from settings.json")
	}

	// Verify lint.sh exists and is executable.
	lintPath := filepath.Join(tmp, ".claude", "hooks", "lint.sh")
	info, err := os.Stat(lintPath)
	if err != nil {
		t.Fatalf("lint.sh not written: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("lint.sh not executable: mode %v", info.Mode())
	}
}

func TestInstallHooks_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "all",
		Locale:   "en",
		Install:  dto.InstallScopeProject,
	}

	first, err := installer.Execute(cfg)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	totalAdded := total(first.HandlersAdded)
	if totalAdded == 0 {
		t.Fatal("first install added 0 handlers")
	}

	second, err := installer.Execute(cfg)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if total(second.HandlersAdded) != 0 {
		t.Fatalf("second install added %d handlers, want 0 (idempotent)", total(second.HandlersAdded))
	}
	if total(second.HandlersSkipped) != totalAdded {
		t.Fatalf("second install skipped %d, want %d (every prior handler should be deduped)", total(second.HandlersSkipped), totalAdded)
	}
	if len(second.ScriptsCopied) != 0 {
		t.Fatalf("second install copied %d scripts, want 0 (already on disk)", len(second.ScriptsCopied))
	}
	if len(second.ScriptsSkipped) == 0 {
		t.Fatal("second install reported no skipped scripts (should match originals)")
	}
}

func TestInstallHooks_BackupOnPreExisting(t *testing.T) {
	tmp := t.TempDir()
	settingsDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.json")
	original := []byte(`{"existing":"value"}`)
	if err := os.WriteFile(settingsPath, original, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	installer := newTestInstaller(tmp)
	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "linting",
		Locale:   "en",
		Install:  dto.InstallScopeProject,
	}

	result, err := installer.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup for pre-existing settings.json")
	}
	bak, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(bak) != string(original) {
		t.Fatalf("backup content mismatch")
	}

	// settings.json should now have BOTH existing key AND hooks key.
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read merged: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := parsed["existing"]; got != "value" {
		t.Fatalf("existing key dropped: %v", got)
	}
	if _, ok := parsed["hooks"]; !ok {
		t.Fatal("hooks not added")
	}
}

func TestInstallHooks_DryRun(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "linting",
		Locale:   "en",
		Install:  dto.InstallScopeProject,
		DryRun:   true,
	}

	result, err := installer.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun: result.DryRun must be true")
	}

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote settings.json (err=%v)", err)
	}
	hooksDir := filepath.Join(tmp, ".claude", "hooks")
	if _, err := os.Stat(hooksDir); !os.IsNotExist(err) {
		t.Fatalf("dry run created hooks dir (err=%v)", err)
	}
}

func TestInstallHooks_GlobalRewritesCommandsToHome(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "all",
		Locale:   "en",
		Install:  dto.InstallScopeGlobal,
	}

	if _, err := installer.Execute(cfg); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	got := string(data)

	if strings.Contains(got, "$CLAUDE_PROJECT_DIR") {
		t.Fatalf("global install must rewrite $CLAUDE_PROJECT_DIR; settings.json:\n%s", got)
	}
	if !strings.Contains(got, `$HOME\"/.claude/hooks/`) {
		t.Fatalf("global install must reference $HOME/.claude/hooks/; settings.json:\n%s", got)
	}
}

func TestInstallHooks_ProjectKeepsClaudeProjectDir(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category: "hooks",
		Preset:   "security-guardrails",
		Locale:   "en",
		Install:  dto.InstallScopeProject,
	}

	if _, err := installer.Execute(cfg); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, "$CLAUDE_PROJECT_DIR") {
		t.Fatalf("project install must keep $CLAUDE_PROJECT_DIR; settings.json:\n%s", got)
	}
	if strings.Contains(got, `$HOME\"/.claude/hooks/`) {
		t.Fatalf("project install must NOT rewrite to $HOME; settings.json:\n%s", got)
	}
}

func TestInstallHooks_RejectsCustomScope(t *testing.T) {
	tmp := t.TempDir()
	installer := newTestInstaller(tmp)

	cfg := &dto.HookConfig{
		Category:   "hooks",
		Preset:     "linting",
		Locale:     "en",
		OutputPath: filepath.Join(tmp, "preview"),
		// Install left empty intentionally
	}
	if _, err := installer.Execute(cfg); err == nil {
		t.Fatal("expected error: install scope required")
	}
}

func TestSettingsLoad_EmptyEqualsMissing(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	s, err := settings.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(s.Raw) != 0 {
		t.Fatalf("expected empty Raw, got %v", s.Raw)
	}
}

func total(m map[string]int) int {
	t := 0
	for _, v := range m {
		t += v
	}
	return t
}
