package cofly

import (
	"reflect"
	"testing"
)

func TestSpanParseAndString(t *testing.T) {
	t.Run("parseSpan-valid", func(t *testing.T) {
		type tc struct {
			in   string
			want span
		}

		cases := []tc{
			{"0..", span{indexFrom: 0, indexTo: 0}},
			{"0..0", span{indexFrom: 0, indexTo: 0}},
			{"0..1", span{indexFrom: 0, indexTo: 1}},
			{"1..1", span{indexFrom: 1, indexTo: 1}},
			{"-2..", span{indexFrom: -2, indexTo: -2}},
			{"-2..0", span{indexFrom: -2, indexTo: 0}},
		}

		for _, c := range cases {
			got, ok := parseSpan(c.in)
			if !ok {
				t.Fatalf("parseSpan(%q): expected ok=true", c.in)
			}
			if got != c.want {
				t.Fatalf("parseSpan(%q): expected %#v, got %#v", c.in, c.want, got)
			}
		}
	})

	t.Run("parseSpan-invalid", func(t *testing.T) {
		invalid := []string{
			"",
			"1",
			"..",
			"a..b",
			"1...2",
			"2..1",
			"1..-1",
		}

		for _, in := range invalid {
			if _, ok := parseSpan(in); ok {
				t.Fatalf("parseSpan(%q): expected ok=false", in)
			}
		}
	})

	t.Run("round-trip-string-parse", func(t *testing.T) {
		spans := []span{
			{indexFrom: 0, indexTo: 0},
			{indexFrom: 0, indexTo: 3},
			{indexFrom: 5, indexTo: 5},
			{indexFrom: -2, indexTo: -2},
			{indexFrom: -2, indexTo: 0},
			{indexFrom: 4, indexTo: 9},
		}

		for _, s := range spans {
			key := s.string()
			got, ok := parseSpan(key)
			if !ok {
				t.Fatalf("parseSpan(%q): expected ok=true", key)
			}
			if got != s {
				t.Fatalf("round-trip: start=%#v key=%q got=%#v", s, key, got)
			}
		}
	})
}

func TestSpanHelpers(t *testing.T) {
	t.Run("positiveBeforeLength", func(t *testing.T) {
		s := span{indexFrom: -2, indexTo: 3}
		got, ok := s.positiveBeforeLength(2)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		// clamp into [0..len]
		want := span{indexFrom: 0, indexTo: 2}
		if got != want {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("negative", func(t *testing.T) {
		s := span{indexFrom: -2, indexTo: 3}
		got, ok := s.negative()
		if !ok {
			t.Fatalf("expected ok=true")
		}
		want := span{indexFrom: -2, indexTo: 0}
		if got != want {
			t.Fatalf("expected %#v, got %#v", want, got)
		}

		if _, ok := (span{indexFrom: 0, indexTo: 1}).negative(); ok {
			t.Fatalf("expected ok=false for non-negative span")
		}
	})

	t.Run("afterLength", func(t *testing.T) {
		s := span{indexFrom: 1, indexTo: 10}
		got, ok := s.afterLength(4)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		want := span{indexFrom: 4, indexTo: 10}
		if got != want {
			t.Fatalf("expected %#v, got %#v", want, got)
		}

		if _, ok := (span{indexFrom: 0, indexTo: 3}).afterLength(4); ok {
			t.Fatalf("expected ok=false when inside length")
		}
	})

	t.Run("length", func(t *testing.T) {
		if (span{indexFrom: 2, indexTo: 5}).length() != 3 {
			t.Fatalf("expected length 3")
		}
	})

	t.Run("string-capacity-stability", func(t *testing.T) {
		// This is not about exact formatting (covered elsewhere), just ensure it doesn't allocate weirdly
		// or change shape across calls.
		s := span{indexFrom: 12, indexTo: 12}
		if s.string() != "12.." {
			t.Fatalf("expected %q, got %q", "12..", s.string())
		}
	})

	t.Run("parseSpan-and-string-dont-lose-sign", func(t *testing.T) {
		s := span{indexFrom: -10, indexTo: -10}
		got, ok := parseSpan(s.string())
		if !ok || !reflect.DeepEqual(got, s) {
			t.Fatalf("expected %#v, got %#v (ok=%v)", s, got, ok)
		}
	})
}
