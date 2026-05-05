package usage

import (
	"os"
	"path/filepath"
	"testing"

	domain "github.com/jorelcb/codify/internal/domain/usage"
)

func TestRepository_LoadMissing_ReturnsEmptyLog(t *testing.T) {
	repo := NewRepository()
	log, err := repo.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("missing should not error: %v", err)
	}
	if len(log.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(log.Entries))
	}
	if log.SchemaVersion != domain.SchemaVersion {
		t.Errorf("schema not set: %q", log.SchemaVersion)
	}
}

func TestRepository_AppendAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	repo := NewRepository()

	if err := repo.Append(path, domain.Entry{Command: "test", InputTokens: 100, CostUSDCents: 5}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := repo.Append(path, domain.Entry{Command: "test", InputTokens: 200, CostUSDCents: 10}); err != nil {
		t.Fatalf("append: %v", err)
	}

	log, err := repo.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(log.Entries) != 2 {
		t.Errorf("got %d entries, want 2", len(log.Entries))
	}
	if log.Totals.Calls != 2 || log.Totals.InputTokens != 300 || log.Totals.CostUSDCents != 15 {
		t.Errorf("totals wrong: %+v", log.Totals)
	}
}

func TestRepository_Reset_BackupsAndTruncates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	repo := NewRepository()
	if err := repo.Append(path, domain.Entry{InputTokens: 100, CostUSDCents: 5}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Reset(path); err != nil {
		t.Fatalf("reset: %v", err)
	}
	log, _ := repo.Load(path)
	if len(log.Entries) != 0 {
		t.Errorf("after reset expected 0 entries, got %d", len(log.Entries))
	}
	// Verificar que se creó algún backup
	dir := filepath.Dir(path)
	files, _ := os.ReadDir(dir)
	hasBak := false
	for _, f := range files {
		if filepath.Ext(f.Name()) != "" && f.Name() != "usage.json" {
			hasBak = true
			break
		}
	}
	if !hasBak {
		t.Error("expected backup file after reset")
	}
}

func TestTrackingDisabled_FlagOverride(t *testing.T) {
	if !TrackingDisabled(true) {
		t.Error("flag=true should disable tracking")
	}
}

func TestTrackingDisabled_EnvOverride(t *testing.T) {
	t.Setenv("CODIFY_NO_USAGE_TRACKING", "1")
	if !TrackingDisabled(false) {
		t.Error("env should disable tracking")
	}
}

func TestRecorder_DisabledIsNoop(t *testing.T) {
	t.Setenv("CODIFY_NO_USAGE_TRACKING", "1")
	t.Setenv("HOME", t.TempDir())
	rec := NewRecorder(false)
	rec.Record(domain.Entry{Command: "test", InputTokens: 100})
	// Verify no file was created
	dir := os.Getenv("HOME")
	if _, err := os.Stat(filepath.Join(dir, ".codify", "usage.json")); err == nil {
		t.Error("expected no usage.json with tracking disabled")
	}
}
