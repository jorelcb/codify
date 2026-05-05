package config_merge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"

	domain "github.com/jorelcb/codify/internal/domain/config"
	infraconfig "github.com/jorelcb/codify/internal/infrastructure/config"
	"github.com/jorelcb/codify/tests/bdd/commons/assertions"
)

var featureContext = new(FeatureContext)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(featureContext.SetupTest)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		featureContext.reset()
		return c, nil
	})

	// Given
	ctx.Step(`^no user config exists$`, featureContext.noUserConfigExists)
	ctx.Step(`^no project config exists$`, featureContext.noProjectConfigExists)
	ctx.Step(`^a user config with preset "([^"]*)" and locale "([^"]*)"$`, featureContext.aUserConfigWithPresetAndLocale)
	ctx.Step(`^a project config with preset "([^"]*)"$`, featureContext.aProjectConfigWithPreset)
	ctx.Step(`^an empty home directory$`, featureContext.anEmptyHomeDirectory)
	ctx.Step(`^a config in memory$`, featureContext.aConfigInMemory)

	// When
	ctx.Step(`^I load the effective config$`, featureContext.iLoadTheEffectiveConfig)
	ctx.Step(`^I save a user config with preset "([^"]*)", locale "([^"]*)", and target "([^"]*)"$`, featureContext.iSaveUserConfigFull)
	ctx.Step(`^I save a user config with preset "([^"]*)"$`, featureContext.iSaveUserConfigPreset)
	ctx.Step(`^I load the user config from disk$`, featureContext.iLoadUserConfigFromDisk)
	ctx.Step(`^I get the value of key "([^"]*)"$`, featureContext.iGetValueOfKey)
	ctx.Step(`^I set the value of key "([^"]*)" to "([^"]*)"$`, featureContext.iSetValueOfKey)

	// Then
	ctx.Step(`^the effective preset should be "([^"]*)"$`, featureContext.theEffectivePresetShouldBe)
	ctx.Step(`^the effective locale should be "([^"]*)"$`, featureContext.theEffectiveLocaleShouldBe)
	ctx.Step(`^the effective target should be "([^"]*)"$`, featureContext.theEffectiveTargetShouldBe)
	ctx.Step(`^the loaded preset should be "([^"]*)"$`, featureContext.theLoadedPresetShouldBe)
	ctx.Step(`^the loaded locale should be "([^"]*)"$`, featureContext.theLoadedLocaleShouldBe)
	ctx.Step(`^the loaded target should be "([^"]*)"$`, featureContext.theLoadedTargetShouldBe)
	ctx.Step(`^the loaded version should equal the schema version$`, featureContext.theLoadedVersionShouldEqualSchema)
	ctx.Step(`^a backup file "\.bak" should exist next to the user config$`, featureContext.aBackupShouldExist)
	ctx.Step(`^I should get a config error containing "([^"]*)"$`, featureContext.iShouldGetConfigErrorContaining)
}

// --- Given steps ---

func (f *FeatureContext) noUserConfigExists() error            { return nil }
func (f *FeatureContext) noProjectConfigExists() error         { return nil }
func (f *FeatureContext) anEmptyHomeDirectory() error          { return nil }
func (f *FeatureContext) aConfigInMemory() error               { f.inMemory = domain.Config{}; return nil }

func (f *FeatureContext) aUserConfigWithPresetAndLocale(preset, locale string) error {
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	return f.repo.Save(userPath, domain.Config{Preset: preset, Locale: locale})
}

func (f *FeatureContext) aProjectConfigWithPreset(preset string) error {
	projectPath, err := infraconfig.ProjectConfigPath()
	if err != nil {
		return err
	}
	return f.repo.Save(projectPath, domain.Config{Preset: preset})
}

// --- When steps ---

func (f *FeatureContext) iLoadTheEffectiveConfig() error {
	cfg, err := f.repo.LoadEffective()
	f.effective = cfg
	f.err = err
	return nil
}

func (f *FeatureContext) iSaveUserConfigFull(preset, locale, target string) error {
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	return f.repo.Save(userPath, domain.Config{Preset: preset, Locale: locale, Target: target})
}

func (f *FeatureContext) iSaveUserConfigPreset(preset string) error {
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	return f.repo.Save(userPath, domain.Config{Preset: preset})
}

func (f *FeatureContext) iLoadUserConfigFromDisk() error {
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	cfg, _, err := f.repo.Load(userPath)
	f.loaded = cfg
	f.err = err
	return nil
}

func (f *FeatureContext) iGetValueOfKey(key string) error {
	val, err := f.inMemory.Get(key)
	f.gotValue = val
	f.err = err
	return nil
}

func (f *FeatureContext) iSetValueOfKey(key, val string) error {
	err := f.inMemory.Set(key, val)
	f.err = err
	return nil
}

// --- Then steps ---

func (f *FeatureContext) theEffectivePresetShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.effective.Preset, "effective preset")
}

func (f *FeatureContext) theEffectiveLocaleShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.effective.Locale, "effective locale")
}

func (f *FeatureContext) theEffectiveTargetShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.effective.Target, "effective target")
}

func (f *FeatureContext) theLoadedPresetShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.loaded.Preset, "loaded preset")
}

func (f *FeatureContext) theLoadedLocaleShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.loaded.Locale, "loaded locale")
}

func (f *FeatureContext) theLoadedTargetShouldBe(expected string) error {
	return assertions.AssertExpectedAndActual(assert.Equal, expected, f.loaded.Target, "loaded target")
}

func (f *FeatureContext) theLoadedVersionShouldEqualSchema() error {
	return assertions.AssertExpectedAndActual(assert.Equal, domain.SchemaVersion, f.loaded.Version, "schema version")
}

func (f *FeatureContext) aBackupShouldExist() error {
	userPath, err := infraconfig.UserConfigPath()
	if err != nil {
		return err
	}
	bakPath := userPath + ".bak"
	if _, err := os.Stat(bakPath); err != nil {
		return fmt.Errorf("expected backup at %s: %w", bakPath, err)
	}
	return nil
}

func (f *FeatureContext) iShouldGetConfigErrorContaining(needle string) error {
	if f.err == nil {
		return fmt.Errorf("expected error containing %q, got nil", needle)
	}
	if !strings.Contains(f.err.Error(), needle) {
		return fmt.Errorf("expected error to contain %q, got: %v", needle, f.err)
	}
	return nil
}

// asegura que filepath se importe (helper para evitar warnings si la lib no
// se usa; algunos linters quieren explicitness en imports indirectos).
var _ = filepath.Join
