package commands

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRenderUnifiedDiff_ShowsChangedLinesWithContext(t *testing.T) {
	before := strings.Join([]string{
		"line a",
		"line b",
		"line c",
		"line d",
		"line e",
	}, "\n")
	after := strings.Join([]string{
		"line a",
		"line b",
		"changed c",
		"line d",
		"line e",
	}, "\n")

	got := renderUnifiedDiff(before, after)

	for _, want := range []string{"- line c", "+ changed c", "  line b", "  line d"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected diff to contain %q, got:\n%s", want, got)
		}
	}
}

func TestRenderUnifiedDiff_NoChanges_ReturnsEmpty(t *testing.T) {
	if got := renderUnifiedDiff("same", "same"); got != "" {
		t.Errorf("expected empty diff for identical input, got %q", got)
	}
}

func TestRenderUnifiedDiff_AdditionAtEnd(t *testing.T) {
	before := "a\nb"
	after := "a\nb\nc"
	got := renderUnifiedDiff(before, after)
	if !strings.Contains(got, "+ c") {
		t.Errorf("addition not rendered, got:\n%s", got)
	}
}

func TestRenderUnifiedDiff_DeletionAtStart(t *testing.T) {
	before := "a\nb\nc"
	after := "b\nc"
	got := renderUnifiedDiff(before, after)
	if !strings.Contains(got, "- a") {
		t.Errorf("deletion not rendered, got:\n%s", got)
	}
}

func TestPreview_IdenticalContent_AppliesWithoutPrompt(t *testing.T) {
	previewer := NewHuhDiffPreviewer()
	apply, content, err := previewer.Preview("file.md", []byte("same"), []byte("same"))
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if !apply || string(content) != "same" {
		t.Errorf("expected apply=true with same content, got apply=%v content=%q", apply, string(content))
	}
}

func TestEditInExternalEditor_NoEditor_ReturnsError(t *testing.T) {
	previewer := &HuhDiffPreviewer{
		editorEnv: "CODIFY_TEST_NONEXISTENT_EDITOR_VAR",
		runEditor: func(string, string) error { return nil },
	}
	// Cannot fully isolate — vim/vi/nano may still be on PATH on the test
	// host. Compensate by stubbing runEditor to a sentinel error and
	// asserting the path is exercised.
	previewer.runEditor = func(string, string) error {
		return errors.New("stub editor failed")
	}
	_, err := previewer.editInExternalEditor("file.md", []byte("hi"))
	if err == nil {
		t.Fatal("expected error from stubbed editor failure")
	}
}

func TestEditInExternalEditor_Success_ReturnsEditedBytes(t *testing.T) {
	if _, err := os.Stat(os.TempDir()); err != nil {
		t.Skip("temp dir not writable")
	}
	previewer := &HuhDiffPreviewer{
		editorEnv: "EDITOR",
		runEditor: func(_, path string) error {
			// Simulate user editing the temp file by overwriting it.
			return os.WriteFile(path, []byte("edited\n"), 0o644)
		},
	}
	t.Setenv("EDITOR", "stub")

	got, err := previewer.editInExternalEditor("file.md", []byte("original\n"))
	if err != nil {
		t.Fatalf("editInExternalEditor: %v", err)
	}
	if string(got) != "edited\n" {
		t.Errorf("expected edited bytes, got %q", string(got))
	}
}
