package settings

import (
	"errors"
	"os"
	"path/filepath"
)

// GlobalSettingsPath returns the absolute path to ~/.claude/settings.json.
func GlobalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// ProjectSettingsPath returns .claude/settings.json resolved against the
// current working directory. The file may not yet exist.
func ProjectSettingsPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".claude", "settings.json"), nil
}

// GlobalHooksDir returns ~/.claude/hooks/.
func GlobalHooksDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "hooks"), nil
}

// ProjectHooksDir returns .claude/hooks/ relative to the current working directory.
func ProjectHooksDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".claude", "hooks"), nil
}

// ResolveScope returns the (settings.json, hooks dir) pair for the given
// install scope. Scope must be "global" or "project"; any other value
// returns an error so callers fail loudly instead of silently picking a
// default.
func ResolveScope(scope string) (settingsPath string, hooksDir string, err error) {
	switch scope {
	case "global":
		settingsPath, err = GlobalSettingsPath()
		if err != nil {
			return "", "", err
		}
		hooksDir, err = GlobalHooksDir()
		if err != nil {
			return "", "", err
		}
	case "project":
		settingsPath, err = ProjectSettingsPath()
		if err != nil {
			return "", "", err
		}
		hooksDir, err = ProjectHooksDir()
		if err != nil {
			return "", "", err
		}
	default:
		return "", "", errors.New("settings: scope must be \"global\" or \"project\"")
	}
	return settingsPath, hooksDir, nil
}
