package cofly

import "testing"

func TestParseSplice(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		sp, ok := parseSplice("1..2", []any{"x"})
		if !ok {
			t.Fatalf("expected ok=true")
		}
		if sp.span != (span{indexFrom: 1, indexTo: 2}) {
			t.Fatalf("unexpected span: %#v", sp.span)
		}
		if len(sp.value) != 1 || sp.value[0] != "x" {
			t.Fatalf("unexpected value: %#v", sp.value)
		}
	})

	t.Run("invalid-key", func(t *testing.T) {
		if _, ok := parseSplice("not-a-span", []any{"x"}); ok {
			t.Fatalf("expected ok=false")
		}
	})

	t.Run("invalid-value-type", func(t *testing.T) {
		if _, ok := parseSplice("1..2", "x"); ok {
			t.Fatalf("expected ok=false")
		}
	})
}

func TestParseSplices(t *testing.T) {
	t.Run("all-valid", func(t *testing.T) {
		change := map[string]any{
			"1..2": []any{"B"},
			"3..":  []any{"d"},
		}
		sp := parseSplices(change)
		if len(sp) != 2 {
			t.Fatalf("expected 2 splices, got %d", len(sp))
		}
	})

	t.Run("any-invalid-makes-nil", func(t *testing.T) {
		change := map[string]any{
			"1..2":      []any{"B"},
			"not-span":  []any{"x"},
			"also-bad":  "x",
			"3..":       []any{"d"},
			"2..1":      []any{"bad"},
			"something": []any{},
		}
		if got := parseSplices(change); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})
}

func TestSortSplices(t *testing.T) {
	sp := []splice{
		{span: span{indexFrom: 5, indexTo: 5}},
		{span: span{indexFrom: -1, indexTo: 0}},
		{span: span{indexFrom: 2, indexTo: 3}},
	}
	sortSplices(sp)
	if sp[0].span.indexFrom != -1 || sp[1].span.indexFrom != 2 || sp[2].span.indexFrom != 5 {
		t.Fatalf("unexpected order: %#v", sp)
	}
}
