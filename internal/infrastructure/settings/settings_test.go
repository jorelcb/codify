package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Path != path {
		t.Fatalf("Path: got %q, want %q", s.Path, path)
	}
	if s.Raw == nil || len(s.Raw) != 0 {
		t.Fatalf("Raw: got %v, want empty map", s.Raw)
	}
}

func TestLoad_EmptyFileReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Raw == nil || len(s.Raw) != 0 {
		t.Fatalf("Raw: got %v, want empty map", s.Raw)
	}
}

func TestLoad_MalformedReturnsError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("error should mention refusing to overwrite: got %v", err)
	}
}

func TestLoad_PreservesUnknownKeys(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	doc := `{
  "permissions": {"allow": ["Bash(git *)"]},
  "model": "claude-opus-4-6",
  "hooks": {}
}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := s.Raw["permissions"]; !ok {
		t.Fatal("expected permissions key to be preserved")
	}
	if got, _ := s.Raw["model"].(string); got != "claude-opus-4-6" {
		t.Fatalf("model: got %q", got)
	}
}

func TestMergeHooks_AddsToEmptySettings(t *testing.T) {
	s := &Settings{Path: "/tmp/x", Raw: map[string]any{}}

	block := mustParse(t, `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {"type": "command", "command": "/path/script.sh", "timeout": 10}
        ]
      }
    ]
  }
}`)

	added, skipped, err := s.MergeHooks(block)
	if err != nil {
		t.Fatalf("MergeHooks: %v", err)
	}
	if added["PreToolUse"] != 1 {
		t.Fatalf("added PreToolUse: got %d, want 1", added["PreToolUse"])
	}
	if skipped["PreToolUse"] != 0 {
		t.Fatalf("skipped: got %d, want 0", skipped["PreToolUse"])
	}

	hooks := s.Raw["hooks"].(map[string]any)
	pre := hooks["PreToolUse"].([]any)
	if len(pre) != 1 {
		t.Fatalf("PreToolUse handlers: got %d, want 1", len(pre))
	}
}

func TestMergeHooks_Idempotent(t *testing.T) {
	s := &Settings{Path: "/tmp/x", Raw: map[string]any{}}
	block := mustParse(t, `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "/x.sh", "timeout": 10}]
      }
    ]
  }
}`)

	if _, _, err := s.MergeHooks(block); err != nil {
		t.Fatalf("first merge: %v", err)
	}
	added, skipped, err := s.MergeHooks(block)
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}
	if added["PreToolUse"] != 0 {
		t.Fatalf("second merge added: got %d, want 0", added["PreToolUse"])
	}
	if skipped["PreToolUse"] != 1 {
		t.Fatalf("second merge skipped: got %d, want 1", skipped["PreToolUse"])
	}
}

func TestMergeHooks_AddsDifferentCommands(t *testing.T) {
	s := &Settings{Path: "/tmp/x", Raw: map[string]any{}}

	first := mustParse(t, `{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "/a.sh"}]}
    ]
  }
}`)
	second := mustParse(t, `{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "/b.sh"}]}
    ]
  }
}`)

	if _, _, err := s.MergeHooks(first); err != nil {
		t.Fatalf("merge first: %v", err)
	}
	if _, _, err := s.MergeHooks(second); err != nil {
		t.Fatalf("merge second: %v", err)
	}

	hooks := s.Raw["hooks"].(map[string]any)
	pre := hooks["PreToolUse"].([]any)
	if len(pre) != 2 {
		t.Fatalf("matchers: got %d, want 2", len(pre))
	}
}

func TestMergeHooks_RejectsInputWithoutHooksKey(t *testing.T) {
	s := &Settings{Path: "/tmp/x", Raw: map[string]any{}}
	if _, _, err := s.MergeHooks(map[string]any{}); err == nil {
		t.Fatal("expected error when block has no hooks key")
	}
}

func TestSave_CreatesBackupAndAtomic(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	original := `{"existing":"value"}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	s.Raw["new"] = "added"

	backupPath, err := s.Save("test-suffix")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if backupPath == "" {
		t.Fatal("expected non-empty backup path")
	}

	bak, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(bak) != original {
		t.Fatalf("backup content: got %q, want %q", string(bak), original)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(current, &parsed); err != nil {
		t.Fatalf("parse saved: %v", err)
	}
	if got := parsed["existing"]; got != "value" {
		t.Fatalf("existing key dropped: got %v", got)
	}
	if got := parsed["new"]; got != "added" {
		t.Fatalf("new key not saved: got %v", got)
	}
}

func TestSave_NoBackupForNewFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".claude", "settings.json")

	s := &Settings{Path: path, Raw: map[string]any{"x": 1}}
	backupPath, err := s.Save("test")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if backupPath != "" {
		t.Fatalf("expected no backup for new file, got %q", backupPath)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("settings file not written: %v", err)
	}
}

func TestPreviewMergedHooks_DoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	block := mustParse(t, `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"command":"/x.sh"}]}]}}`)

	preview, err := s.PreviewMergedHooks(block)
	if err != nil {
		t.Fatalf("PreviewMergedHooks: %v", err)
	}
	if !strings.Contains(string(preview), "/x.sh") {
		t.Fatalf("preview should contain merged command: %s", preview)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("preview should not write: %v", err)
	}
}

func mustParse(t *testing.T, doc string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(doc), &m); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return m
}
