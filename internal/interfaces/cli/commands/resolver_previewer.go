package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jorelcb/codify/internal/domain/service"
)

// HuhDiffPreviewer implements service.DiffPreviewer for the terminal. It
// shows a unified-style diff of the changed lines, asks the user whether
// to apply, discard, or edit, and returns the chosen content.
//
// The diff renderer is intentionally minimal — line-based, hunks defined by
// runs of equal lines with two lines of context. It is good enough for
// resolver use (changes are localized around marker lines) without pulling
// in a third-party diff dependency.
type HuhDiffPreviewer struct {
	// editorEnv is the environment variable name to read the editor from.
	// Defaults to "EDITOR"; tests override.
	editorEnv string
	// runEditor invokes the editor on the given path. Defaults to a real
	// exec; tests override to simulate user edits without a TTY.
	runEditor func(editorCmd, path string) error
}

// NewHuhDiffPreviewer returns a previewer ready for terminal use.
func NewHuhDiffPreviewer() *HuhDiffPreviewer {
	return &HuhDiffPreviewer{
		editorEnv: "EDITOR",
		runEditor: defaultRunEditor,
	}
}

// Preview shows the diff, asks for confirmation, and returns the user's
// choice. When edit is requested, the proposed content is written to a temp
// file, the editor is invoked synchronously, and the saved content is read
// back. Editor failures (no editor configured, exec failed) degrade to
// applying the original proposed content with a stderr warning.
func (p *HuhDiffPreviewer) Preview(path string, before, after []byte) (bool, []byte, error) {
	if string(before) == string(after) {
		return true, after, nil
	}

	fmt.Println()
	fmt.Printf("About to rewrite %s:\n", path)
	fmt.Println(renderUnifiedDiff(string(before), string(after)))

	choice, err := promptSelect(
		"Apply changes?",
		[]selectOption{
			{"Apply", "apply"},
			{"Discard (keep file as-is)", "discard"},
			{"Edit before applying", "edit"},
		},
		"apply",
	)
	if err != nil {
		return false, nil, err
	}
	switch choice {
	case "discard":
		return false, before, nil
	case "edit":
		edited, err := p.editInExternalEditor(path, after)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  editor failed (%v); applying proposed content as-is\n", err)
			return true, after, nil
		}
		return true, edited, nil
	default:
		return true, after, nil
	}
}

// editInExternalEditor writes proposedContent to a temp file with the same
// extension as path (so the editor picks the right syntax highlighting),
// invokes the configured editor, and returns the edited bytes.
func (p *HuhDiffPreviewer) editInExternalEditor(path string, proposedContent []byte) ([]byte, error) {
	editor := os.Getenv(p.editorEnv)
	if editor == "" {
		for _, candidate := range []string{"vim", "vi", "nano"} {
			if _, err := exec.LookPath(candidate); err == nil {
				editor = candidate
				break
			}
		}
	}
	if editor == "" {
		return nil, fmt.Errorf("no editor configured ($%s unset, no fallback found)", p.editorEnv)
	}

	tmp, err := os.CreateTemp("", "codify-resolve-*"+filepath.Ext(path))
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(proposedContent); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	if err := p.runEditor(editor, tmpPath); err != nil {
		return nil, fmt.Errorf("editor %s exited with error: %w", editor, err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("read edited file: %w", err)
	}
	return edited, nil
}

// defaultRunEditor execs editorCmd path with the current TTY attached.
func defaultRunEditor(editorCmd, path string) error {
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty editor command")
	}
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// renderUnifiedDiff produces a small unified-style diff between before and
// after, scoped to changed line runs with two lines of leading and trailing
// context. Pure function — exported via the test file alongside.
//
// Algorithm: walk both line slices in parallel, building equal/-/+
// segments. When the streams diverge, find the smallest (k,l) pair such
// that beforeLines[i+k] == afterLines[j+l] — that is the next resync
// point. Lookahead is bounded so input shape doesn't blow up the cost;
// for resolver-sized files (hundreds of lines, localized changes) this is
// fine and avoids pulling in an external diff dependency.
func renderUnifiedDiff(before, after string) string {
	const ctx = 2

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	type segment struct {
		kind  byte // ' ' equal, '-' before-only, '+' after-only
		lines []string
	}
	var segs []segment
	appendEqual := func(line string) {
		if n := len(segs); n > 0 && segs[n-1].kind == ' ' {
			segs[n-1].lines = append(segs[n-1].lines, line)
			return
		}
		segs = append(segs, segment{kind: ' ', lines: []string{line}})
	}

	i, j := 0, 0
	for i < len(beforeLines) || j < len(afterLines) {
		switch {
		case i < len(beforeLines) && j < len(afterLines) && beforeLines[i] == afterLines[j]:
			appendEqual(beforeLines[i])
			i++
			j++
		default:
			k, l := findResync(beforeLines, afterLines, i, j)
			if k < 0 {
				if i < len(beforeLines) {
					segs = append(segs, segment{kind: '-', lines: beforeLines[i:]})
				}
				if j < len(afterLines) {
					segs = append(segs, segment{kind: '+', lines: afterLines[j:]})
				}
				i = len(beforeLines)
				j = len(afterLines)
				continue
			}
			if k > 0 {
				segs = append(segs, segment{kind: '-', lines: beforeLines[i : i+k]})
				i += k
			}
			if l > 0 {
				segs = append(segs, segment{kind: '+', lines: afterLines[j : j+l]})
				j += l
			}
		}
	}

	var out strings.Builder
	for k, s := range segs {
		if s.kind != '-' && s.kind != '+' {
			continue
		}
		prefix := s.kind
		// Leading context.
		if k > 0 && segs[k-1].kind == ' ' {
			prev := segs[k-1].lines
			start := len(prev) - ctx
			if start < 0 {
				start = 0
			}
			for _, line := range prev[start:] {
				fmt.Fprintf(&out, "  %s\n", line)
			}
		}
		for _, line := range s.lines {
			fmt.Fprintf(&out, "%c %s\n", prefix, line)
		}
		// Trailing context — only after the last segment of a difference run
		// (next is equal), to avoid duplicating between adjacent -/+ pairs.
		if k+1 < len(segs) && segs[k+1].kind == ' ' && (k+2 >= len(segs) || segs[k+2].kind == ' ') {
			next := segs[k+1].lines
			if len(next) > ctx {
				next = next[:ctx]
			}
			for _, line := range next {
				fmt.Fprintf(&out, "  %s\n", line)
			}
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

// findResync searches for the smallest combined offset (k+l) such that
// beforeLines[i+k] == afterLines[j+l]. Returns (-1, -1) if no resync is
// found within the bounded lookahead window.
func findResync(beforeLines, afterLines []string, i, j int) (int, int) {
	const maxLookahead = 200
	for d := 1; d <= maxLookahead; d++ {
		for k := 0; k <= d; k++ {
			l := d - k
			if i+k >= len(beforeLines) || j+l >= len(afterLines) {
				continue
			}
			if beforeLines[i+k] == afterLines[j+l] {
				return k, l
			}
		}
	}
	return -1, -1
}

// satisfies the port at compile time
var _ service.DiffPreviewer = (*HuhDiffPreviewer)(nil)
