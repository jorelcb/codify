package config_merge

import (
	"os"

	domain "github.com/jorelcb/codify/internal/domain/config"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
)

// FeatureContext holds the state for config_merge feature scenarios.
//
// Each scenario gets a fresh temp HOME and temp cwd so file-based config
// loading is isolated. The previous HOME is restored in reset().
type FeatureContext struct {
	repo *infraconfig.Repository

	originalHome string
	originalCwd  string
	tempHome     string
	tempCwd      string

	effective domain.Config
	loaded    domain.Config
	inMemory  domain.Config
	gotValue  string
	err       error
}

// SetupTest is called once before all scenarios.
func (f *FeatureContext) SetupTest() {
	f.repo = infraconfig.NewRepository()
}

// reset clears state and reroutes HOME/cwd to fresh temp dirs.
func (f *FeatureContext) reset() {
	if f.originalHome != "" {
		_ = os.Setenv("HOME", f.originalHome)
	}
	if f.originalCwd != "" {
		_ = os.Chdir(f.originalCwd)
	}
	if f.tempHome != "" {
		_ = os.RemoveAll(f.tempHome)
	}
	if f.tempCwd != "" {
		_ = os.RemoveAll(f.tempCwd)
	}

	f.originalHome = os.Getenv("HOME")
	f.originalCwd, _ = os.Getwd()

	f.tempHome, _ = os.MkdirTemp("", "codify-bdd-home-*")
	f.tempCwd, _ = os.MkdirTemp("", "codify-bdd-cwd-*")
	_ = os.Setenv("HOME", f.tempHome)
	_ = os.Chdir(f.tempCwd)

	f.effective = domain.Config{}
	f.loaded = domain.Config{}
	f.inMemory = domain.Config{}
	f.gotValue = ""
	f.err = nil
}
