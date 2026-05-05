package watch_loop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/jorelcb/codify/internal/infrastructure/watch"
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
	ctx.Step(`^a temp project with a watched file "([^"]*)"$`, featureContext.aTempProjectWithWatchedFile)
	ctx.Step(`^an unwatched sibling file "([^"]*)"$`, featureContext.anUnwatchedSiblingFile)

	// When
	ctx.Step(`^I start the watcher with debounce (\d+)ms$`, featureContext.iStartTheWatcherWithDebounceMs)
	ctx.Step(`^I modify the watched file$`, featureContext.iModifyTheWatchedFile)
	ctx.Step(`^I modify the unwatched file$`, featureContext.iModifyTheUnwatchedFile)
	ctx.Step(`^I modify the watched file (\d+) times in (\d+)ms$`, featureContext.iModifyTheWatchedFileNTimesInMs)
	ctx.Step(`^I wait (\d+)ms$`, featureContext.iWaitMs)
	ctx.Step(`^I cancel the watcher context$`, featureContext.iCancelTheWatcherContext)
	ctx.Step(`^I create a watcher with no paths$`, featureContext.iCreateAWatcherWithNoPaths)
	ctx.Step(`^I create a watcher without an OnEvent callback$`, featureContext.iCreateAWatcherWithoutOnEvent)

	// Then
	ctx.Step(`^exactly (\d+) watch events? should have fired$`, featureContext.exactlyNWatchEventsShouldHaveFired)
	ctx.Step(`^no watch events should have fired$`, featureContext.noWatchEventsShouldHaveFired)
	ctx.Step(`^the event paths should include "([^"]*)"$`, featureContext.theEventPathsShouldInclude)
	ctx.Step(`^the watcher should return without error within (\d+) seconds?$`, featureContext.theWatcherShouldReturnWithoutErrorWithinNSeconds)
	ctx.Step(`^the watcher should return an error containing "([^"]*)"$`, featureContext.theWatcherShouldReturnAnErrorContaining)
}

// --- Given ---

func (f *FeatureContext) aTempProjectWithWatchedFile(name string) error {
	dir, err := os.MkdirTemp("", "codify-bdd-watch-*")
	if err != nil {
		return err
	}
	f.tempDir = dir
	f.watchedPath = filepath.Join(dir, name)
	if err := os.WriteFile(f.watchedPath, []byte("v0"), 0o644); err != nil {
		return err
	}
	return nil
}

func (f *FeatureContext) anUnwatchedSiblingFile(name string) error {
	if f.tempDir == "" {
		return fmt.Errorf("must call 'a temp project' first")
	}
	f.otherPath = filepath.Join(f.tempDir, name)
	return os.WriteFile(f.otherPath, []byte("v0"), 0o644)
}

// --- When ---

func (f *FeatureContext) iStartTheWatcherWithDebounceMs(ms int) error {
	if f.watchedPath == "" {
		return fmt.Errorf("no watched file set")
	}
	w, err := watch.New(watch.Options{
		Paths:    []string{f.watchedPath},
		Debounce: time.Duration(ms) * time.Millisecond,
		OnEvent:  f.recordEvent,
	})
	if err != nil {
		return err
	}
	f.watcher = w
	f.startCtx, f.startCancel = context.WithCancel(context.Background())
	f.startDone = make(chan error, 1)
	go func() {
		f.startDone <- w.Start(f.startCtx)
	}()
	// Pequeña pausa para que fsnotify registre la subscripción antes de que
	// el siguiente step modifique el archivo.
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (f *FeatureContext) iModifyTheWatchedFile() error {
	return os.WriteFile(f.watchedPath, []byte("v1"), 0o644)
}

func (f *FeatureContext) iModifyTheUnwatchedFile() error {
	return os.WriteFile(f.otherPath, []byte("v1"), 0o644)
}

func (f *FeatureContext) iModifyTheWatchedFileNTimesInMs(n, totalMs int) error {
	gap := time.Duration(totalMs/n) * time.Millisecond
	for i := 0; i < n; i++ {
		if err := os.WriteFile(f.watchedPath, []byte{byte('a' + i)}, 0o644); err != nil {
			return err
		}
		time.Sleep(gap)
	}
	return nil
}

func (f *FeatureContext) iWaitMs(ms int) error {
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil
}

func (f *FeatureContext) iCancelTheWatcherContext() error {
	if f.startCancel != nil {
		f.startCancel()
	}
	return nil
}

func (f *FeatureContext) iCreateAWatcherWithNoPaths() error {
	_, f.constructionErr = watch.New(watch.Options{
		OnEvent: f.recordEvent,
	})
	return nil
}

func (f *FeatureContext) iCreateAWatcherWithoutOnEvent() error {
	_, f.constructionErr = watch.New(watch.Options{
		Paths: []string{f.watchedPath},
	})
	return nil
}

// --- Then ---

func (f *FeatureContext) exactlyNWatchEventsShouldHaveFired(n int) error {
	got := f.eventCount()
	if got != n {
		return fmt.Errorf("expected %d events, got %d", n, got)
	}
	return nil
}

func (f *FeatureContext) noWatchEventsShouldHaveFired() error {
	got := f.eventCount()
	if got != 0 {
		return fmt.Errorf("expected 0 events, got %d", got)
	}
	return nil
}

func (f *FeatureContext) theEventPathsShouldInclude(name string) error {
	paths := f.firstEventPaths()
	for _, p := range paths {
		if filepath.Base(p) == name {
			return nil
		}
	}
	return fmt.Errorf("expected event paths to include %q; got: %v", name, paths)
}

func (f *FeatureContext) theWatcherShouldReturnWithoutErrorWithinNSeconds(n int) error {
	select {
	case err := <-f.startDone:
		f.startDone = nil // mark as drained
		if err != nil {
			return fmt.Errorf("watcher returned error on cancel: %v", err)
		}
		return nil
	case <-time.After(time.Duration(n) * time.Second):
		return fmt.Errorf("watcher did not exit within %d seconds", n)
	}
}

func (f *FeatureContext) theWatcherShouldReturnAnErrorContaining(needle string) error {
	if f.constructionErr == nil {
		return fmt.Errorf("expected construction error containing %q, got nil", needle)
	}
	if !strings.Contains(f.constructionErr.Error(), needle) {
		return fmt.Errorf("expected error to contain %q, got: %v", needle, f.constructionErr)
	}
	return nil
}
