package watch

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// helper: create a file with content
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNew_RequiresOnEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	writeFile(t, path, "x")
	_, err := New(Options{Paths: []string{path}})
	if err == nil {
		t.Error("expected error when OnEvent is nil")
	}
}

func TestNew_RequiresAtLeastOnePath(t *testing.T) {
	_, err := New(Options{OnEvent: func(Event) {}})
	if err == nil {
		t.Error("expected error when Paths is empty")
	}
}

func TestNew_FailsWhenAllPathsHaveNoExistingParent(t *testing.T) {
	// Path under a nonexistent dir
	_, err := New(Options{
		Paths:   []string{"/nonexistent-dir-12345/file.txt"},
		OnEvent: func(Event) {},
	})
	if err == nil {
		t.Error("expected error when no watchable directories exist")
	}
}

func TestStart_DetectsWriteAfterDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	writeFile(t, path, "v1")

	var mu sync.Mutex
	var got []Event
	w, err := New(Options{
		Paths:    []string{path},
		Debounce: 200 * time.Millisecond,
		OnEvent: func(e Event) {
			mu.Lock()
			defer mu.Unlock()
			got = append(got, e)
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Start(ctx)
	}()

	// fsnotify needs a moment to register the subscription
	time.Sleep(50 * time.Millisecond)

	if err := os.WriteFile(path, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond) // wait for debounce + tick

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Errorf("expected at least 1 event, got 0")
	}
	if len(got) > 0 {
		found := false
		absPath, _ := filepath.Abs(path)
		for _, p := range got[0].Paths {
			if p == absPath {
				found = true
			}
		}
		if !found {
			t.Errorf("expected event for %s, got paths: %v", absPath, got[0].Paths)
		}
	}
}

func TestStart_IgnoresFilesOutsideWatchSet(t *testing.T) {
	dir := t.TempDir()
	watched := filepath.Join(dir, "watched.txt")
	other := filepath.Join(dir, "other.txt")
	writeFile(t, watched, "v1")
	writeFile(t, other, "v1")

	var mu sync.Mutex
	var got []Event
	w, err := New(Options{
		Paths:    []string{watched},
		Debounce: 200 * time.Millisecond,
		OnEvent: func(e Event) {
			mu.Lock()
			defer mu.Unlock()
			got = append(got, e)
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Modify the OTHER file (should not trigger)
	if err := os.WriteFile(other, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 0 {
		t.Errorf("expected 0 events for unwatched file, got %d: %+v", len(got), got)
	}
}

func TestStart_DebouncesMultipleQuickWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	writeFile(t, path, "v0")

	var mu sync.Mutex
	var eventCount int
	w, err := New(Options{
		Paths:    []string{path},
		Debounce: 300 * time.Millisecond,
		OnEvent: func(e Event) {
			mu.Lock()
			defer mu.Unlock()
			eventCount++
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- w.Start(ctx)
	}()
	time.Sleep(50 * time.Millisecond)

	// 5 rapid writes — should coalesce into ONE event
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(path, []byte{byte('a' + i)}, 0o644)
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(700 * time.Millisecond) // wait for debounce + buffer

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if eventCount != 1 {
		t.Errorf("expected exactly 1 debounced event, got %d", eventCount)
	}
}

func TestStart_ExitsCleanlyOnContextCancel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	writeFile(t, path, "v1")

	w, err := New(Options{
		Paths:   []string{path},
		OnEvent: func(Event) {},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil on clean shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("watcher did not exit within 2s of context cancel")
	}
}
