package usage

import "testing"

func TestCostCents_KnownModel(t *testing.T) {
	// Sonnet 4.6: input $3/M = 300 cents/M; 1M input = 300 cents
	got := CostCents("claude-sonnet-4-6", 1_000_000, 0, 0, 0)
	if got != 300 {
		t.Errorf("got %d cents for 1M input, want 300", got)
	}
	// Output $15/M = 1500 cents/M; 1M output = 1500 cents
	got = CostCents("claude-sonnet-4-6", 0, 1_000_000, 0, 0)
	if got != 1500 {
		t.Errorf("got %d cents for 1M output, want 1500", got)
	}
}

func TestCostCents_UnknownModel(t *testing.T) {
	got := CostCents("nonexistent-model", 1_000_000, 1_000_000, 0, 0)
	if got != 0 {
		t.Errorf("unknown model should cost 0, got %d", got)
	}
}

func TestCostCents_CacheTiers(t *testing.T) {
	// Sonnet cache read: $0.30/M = 30 cents/M
	got := CostCents("claude-sonnet-4-6", 0, 0, 1_000_000, 0)
	if got != 30 {
		t.Errorf("got %d for 1M cache read, want 30", got)
	}
	// Sonnet cache creation: $3.75/M = 375 cents/M
	got = CostCents("claude-sonnet-4-6", 0, 0, 0, 1_000_000)
	if got != 375 {
		t.Errorf("got %d for 1M cache creation, want 375", got)
	}
}

func TestLog_AppendAndRecomputeTotals(t *testing.T) {
	l := NewLog()
	l.Append(Entry{InputTokens: 100, OutputTokens: 50, CostUSDCents: 5})
	l.Append(Entry{InputTokens: 200, OutputTokens: 100, CostUSDCents: 10})

	if l.Totals.Calls != 2 {
		t.Errorf("calls: got %d", l.Totals.Calls)
	}
	if l.Totals.InputTokens != 300 {
		t.Errorf("input: got %d", l.Totals.InputTokens)
	}
	if l.Totals.CostUSDCents != 15 {
		t.Errorf("cost: got %d", l.Totals.CostUSDCents)
	}

	// Tamper: zero out Totals manually, then recompute
	l.Totals = Totals{}
	l.RecomputeTotals()
	if l.Totals.Calls != 2 || l.Totals.CostUSDCents != 15 {
		t.Errorf("recompute failed: %+v", l.Totals)
	}
}

func TestListModels(t *testing.T) {
	models := ListModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty model list")
	}
	// Verificar orden estable (lexicográfico)
	for i := 1; i < len(models); i++ {
		if models[i] < models[i-1] {
			t.Errorf("models not sorted: %v", models)
			break
		}
	}
}
