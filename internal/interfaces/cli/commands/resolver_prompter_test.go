package commands

import (
	"testing"

	"github.com/jorelcb/codify/internal/domain/service"
)

func TestParseEnrichedInput_EmptyWithDefault_UsesDefault(t *testing.T) {
	got := ParseEnrichedInput("", []string{"USD", "EUR"}, "USD")
	if got.Skip || got.Answer != "USD" {
		t.Errorf("empty input + default should pick default, got %+v", got)
	}
}

func TestParseEnrichedInput_EmptyWithoutDefault_Skips(t *testing.T) {
	got := ParseEnrichedInput("", []string{"USD"}, "")
	if !got.Skip {
		t.Errorf("empty input + no default must skip, got %+v", got)
	}
}

func TestParseEnrichedInput_SkipKeyword(t *testing.T) {
	for _, in := range []string{"s", "S", "skip", "SKIP", "  s  "} {
		got := ParseEnrichedInput(in, []string{"x"}, "x")
		if !got.Skip {
			t.Errorf("input %q should skip, got %+v", in, got)
		}
	}
}

func TestParseEnrichedInput_NumericPick(t *testing.T) {
	got := ParseEnrichedInput("2", []string{"USD", "EUR", "MXN"}, "")
	if got.Skip || got.Answer != "EUR" {
		t.Errorf("numeric pick should select 2nd suggestion, got %+v", got)
	}
}

func TestParseEnrichedInput_NumericOutOfRange_TreatedAsFreeText(t *testing.T) {
	got := ParseEnrichedInput("9", []string{"USD"}, "")
	if got.Skip || got.Answer != "9" {
		t.Errorf("out-of-range numeric must be free text, got %+v", got)
	}
}

func TestParseEnrichedInput_FreeText(t *testing.T) {
	got := ParseEnrichedInput("PEN", []string{"USD"}, "USD")
	if got.Skip || got.Answer != "PEN" {
		t.Errorf("free text should pass through, got %+v", got)
	}
}

func TestParseEnrichedInput_TrimsWhitespace(t *testing.T) {
	got := ParseEnrichedInput("  PEN  ", nil, "")
	if got.Answer != "PEN" {
		t.Errorf("expected trimmed PEN, got %+v", got)
	}
}

func TestParseEnrichedInput_NoSuggestionsNumericIsFreeText(t *testing.T) {
	got := ParseEnrichedInput("1", nil, "")
	if got.Skip || got.Answer != "1" {
		t.Errorf("no suggestions => numeric is free text, got %+v", got)
	}
}

// Smoke test of the type wiring: HuhPrompter satisfies the port and the
// constructor returns a non-nil value. AskMarker / ConfirmTopLevel cannot be
// exercised here because they require a TTY.
func TestHuhPrompter_SatisfiesPort(t *testing.T) {
	var _ service.InteractivePrompter = NewHuhPrompter()
}
