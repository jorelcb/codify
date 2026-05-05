// Package settings manages Claude Code settings.json files (load, merge, save).
//
// It is intentionally narrow: it does not understand the semantics of every
// settings field, only the shape of the document and how to merge codify's
// hook block idempotently while preserving every other key the user (or
// other tools) may have placed in the file.
//
// Backups: Save() writes a timestamped copy of the previous file before
// overwriting, so a malformed merge can be rolled back manually.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Settings represents a parsed settings.json document plus the path it came
// from. Raw is the full document (every key is preserved on round-trip).
type Settings struct {
	// Path is the absolute path to the settings.json file (may not exist on disk yet).
	Path string

	// Raw holds the full parsed document as a generic map. Empty file or
	// missing file produces an empty map (not nil), so callers can safely
	// merge into it without nil checks.
	Raw map[string]any
}

// Load reads settings.json from path. If the file does not exist, returns
// an empty Settings value (not an error) so callers can merge and Save to
// create it. If the file exists but contains malformed JSON, returns an
// explicit error rather than silently overwriting it.
func Load(path string) (*Settings, error) {
	if path == "" {
		return nil, errors.New("settings: path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Settings{Path: path, Raw: map[string]any{}}, nil
		}
		return nil, fmt.Errorf("settings: read %s: %w", path, err)
	}
	if len(data) == 0 {
		return &Settings{Path: path, Raw: map[string]any{}}, nil
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("settings: parse %s: %w (file exists but is not valid JSON; refusing to overwrite)", path, err)
	}
	if raw == nil {
		raw = map[string]any{}
	}
	return &Settings{Path: path, Raw: raw}, nil
}

// MergeHooks merges a hook block of the form {"hooks": {<event>: [matchers...]}}
// into s.Raw. It returns the number of handlers added per event and the
// number skipped (because an identical command string already exists at
// the same event).
//
// Idempotency is enforced by exact-match comparison of the inner "command"
// string of each handler — running MergeHooks twice with the same input
// yields zero new entries on the second call.
//
// The block argument is the parsed form of a codify-generated hooks.json,
// i.e. the full {"hooks": {...}} document. If block does not contain a
// "hooks" object, MergeHooks returns an error.
func (s *Settings) MergeHooks(block map[string]any) (added, skipped map[string]int, err error) {
	added = map[string]int{}
	skipped = map[string]int{}

	hooksRaw, ok := block["hooks"]
	if !ok {
		return nil, nil, errors.New("settings: input block has no \"hooks\" key")
	}
	incoming, ok := hooksRaw.(map[string]any)
	if !ok {
		return nil, nil, errors.New("settings: input \"hooks\" is not an object")
	}

	if s.Raw == nil {
		s.Raw = map[string]any{}
	}
	target, _ := s.Raw["hooks"].(map[string]any)
	if target == nil {
		target = map[string]any{}
	}

	for event, matchersRaw := range incoming {
		incomingMatchers, ok := matchersRaw.([]any)
		if !ok {
			return nil, nil, fmt.Errorf("settings: hooks.%s is not an array", event)
		}
		existingMatchers, _ := target[event].([]any)

		for _, m := range incomingMatchers {
			matcher, ok := m.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("settings: hooks.%s contains a non-object matcher", event)
			}
			handlersRaw, ok := matcher["hooks"]
			if !ok {
				existingMatchers = append(existingMatchers, matcher)
				added[event]++
				continue
			}
			handlers, ok := handlersRaw.([]any)
			if !ok {
				return nil, nil, fmt.Errorf("settings: hooks.%s[].hooks is not an array", event)
			}

			// Filter handlers by exact command match against existing matchers
			// at the same event. We accept the matcher as-is otherwise.
			fresh := make([]any, 0, len(handlers))
			for _, h := range handlers {
				handler, ok := h.(map[string]any)
				if !ok {
					return nil, nil, fmt.Errorf("settings: hooks.%s[].hooks contains a non-object handler", event)
				}
				cmd, _ := handler["command"].(string)
				if cmd != "" && commandExists(existingMatchers, cmd) {
					skipped[event]++
					continue
				}
				fresh = append(fresh, handler)
				added[event]++
			}
			if len(fresh) == 0 {
				continue
			}
			matcher["hooks"] = fresh
			existingMatchers = append(existingMatchers, matcher)
		}
		target[event] = existingMatchers
	}

	s.Raw["hooks"] = target
	return added, skipped, nil
}

// Save writes s.Raw back to disk at s.Path. If the file already exists, a
// backup is written next to it first (path + "." + backupSuffix). Returns
// the backup path (empty string if no backup was needed).
//
// The write is atomic: the encoder writes to a sibling .tmp file and
// renames it into place. If anything fails before rename, no partial
// settings.json is left behind.
func (s *Settings) Save(backupSuffix string) (string, error) {
	if s.Path == "" {
		return "", errors.New("settings: path is empty")
	}
	if backupSuffix == "" {
		backupSuffix = "codify-backup-" + time.Now().UTC().Format("20060102-150405")
	}

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return "", fmt.Errorf("settings: ensure parent dir: %w", err)
	}

	backupPath := ""
	if existing, err := os.ReadFile(s.Path); err == nil {
		backupPath = s.Path + "." + backupSuffix
		if err := os.WriteFile(backupPath, existing, 0o644); err != nil {
			return "", fmt.Errorf("settings: write backup %s: %w", backupPath, err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("settings: stat existing file: %w", err)
	}

	out, err := json.MarshalIndent(s.Raw, "", "  ")
	if err != nil {
		return "", fmt.Errorf("settings: marshal: %w", err)
	}
	out = append(out, '\n')

	tmpPath := s.Path + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0o644); err != nil {
		return "", fmt.Errorf("settings: write tmp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.Path); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("settings: rename tmp into place: %w", err)
	}
	return backupPath, nil
}

// PreviewMergedHooks returns the JSON serialization of s.Raw with the input
// hook block already merged, without writing anything to disk. Used by
// --dry-run.
func (s *Settings) PreviewMergedHooks(block map[string]any) ([]byte, error) {
	if _, _, err := s.MergeHooks(block); err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(s.Raw, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("settings: marshal preview: %w", err)
	}
	return append(out, '\n'), nil
}

// commandExists returns true if any handler under existingMatchers has a
// "command" field that exactly matches cmd. Used to deduplicate handlers
// across MergeHooks calls.
func commandExists(matchers []any, cmd string) bool {
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
			if existing, _ := handler["command"].(string); existing == cmd {
				return true
			}
		}
	}
	return false
}

// SortedEventNames returns the event names in s.Raw["hooks"] sorted
// alphabetically. Helper for deterministic output.
func (s *Settings) SortedEventNames() []string {
	hooks, _ := s.Raw["hooks"].(map[string]any)
	if hooks == nil {
		return nil
	}
	names := make([]string, 0, len(hooks))
	for k := range hooks {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
