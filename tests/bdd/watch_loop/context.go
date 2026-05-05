package watch_loop

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/jorelcb/codify/internal/infrastructure/watch"
)

// FeatureContext es el estado per-scenario para el watcher BDD.
type FeatureContext struct {
	tempDir      string
	watchedPath  string
	otherPath    string

	watcher    *watch.Watcher
	startCtx   context.Context
	startCancel context.CancelFunc
	startDone  chan error

	mu     sync.Mutex
	events []watch.Event

	constructionErr error
}

func (f *FeatureContext) SetupTest() {}

func (f *FeatureContext) reset() {
	if f.startCancel != nil {
		f.startCancel()
	}
	if f.startDone != nil {
		<-f.startDone
	}
	if f.tempDir != "" {
		_ = os.RemoveAll(f.tempDir)
	}

	f.tempDir = ""
	f.watchedPath = ""
	f.otherPath = ""
	f.watcher = nil
	f.startCtx = nil
	f.startCancel = nil
	f.startDone = nil
	f.events = nil
	f.constructionErr = nil
}

// recordEvent es el OnEvent callback compartido por todos los scenarios.
func (f *FeatureContext) recordEvent(e watch.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
}

// eventCount es thread-safe getter para assertions.
func (f *FeatureContext) eventCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

// firstEventPaths es thread-safe getter para verificar el contenido del
// primer evento.
func (f *FeatureContext) firstEventPaths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) == 0 {
		return nil
	}
	return f.events[0].Paths
}

// silence unused: imports we may need across test files
var _ = time.Second
