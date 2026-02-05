package cofly_test

import (
	"math"
	"testing"

	"github.com/rnkv/cofly-go"
)

func TestEqual(t *testing.T) {
	t.Run("nil-handling", func(t *testing.T) {
		if !cofly.Equal(nil, nil) {
			t.Fatalf("Equal(nil,nil): expected true")
		}

		if cofly.Equal(1, nil) {
			t.Fatalf("Equal(1,nil): expected false")
		}

		if cofly.Equal(nil, 1) {
			t.Fatalf("Equal(nil,1): expected false")
		}
	})

	t.Run("primitives-and-type-mismatches", func(t *testing.T) {
		type testCase struct {
			name string
			a    any
			b    any
			want bool
		}

		testCases := []testCase{
			{"bool-true", true, true, true},
			{"bool-false", true, false, false},
			{"int-eq", 1, 1, true},
			{"int-neq", 1, 2, false},
			{"float-eq", 1.5, 1.5, true},
			{"float-neq", 1.5, 2.5, false},
			{"string-eq", "x", "x", true},
			{"string-neq", "x", "y", false},

			{"bool-int", true, 1, false},
			{"string-int", "1", 1, false},
			{"string-float", "1.0", 1.0, false},
			{"map-slice", map[string]any{}, []any{}, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if got := cofly.Equal(tc.a, tc.b); got != tc.want {
					t.Fatalf("Equal(%#v,%#v): want %v, got %v", tc.a, tc.b, tc.want, got)
				}
			})
		}
	})

	t.Run("int-float64-equivalence", func(t *testing.T) {
		if !cofly.Equal(1, 1.0) {
			t.Fatalf("Equal(1,1.0): expected true")
		}
		if !cofly.Equal(1.0, 1) {
			t.Fatalf("Equal(1.0,1): expected true")
		}
		if cofly.Equal(1.0, 2) {
			t.Fatalf("Equal(1.0,2): expected false")
		}
		if cofly.Equal(2, 1.0) {
			t.Fatalf("Equal(2,1.0): expected false")
		}
	})

	t.Run("float64-nan", func(t *testing.T) {
		nan := math.NaN()
		if cofly.Equal(nan, nan) {
			t.Fatalf("Equal(NaN,NaN): expected false")
		}
	})

	t.Run("float64-infinities-and-zero-sign", func(t *testing.T) {
		if !cofly.Equal(math.Inf(1), math.Inf(1)) {
			t.Fatalf("Equal(+Inf,+Inf): expected true")
		}
		if cofly.Equal(math.Inf(1), math.Inf(-1)) {
			t.Fatalf("Equal(+Inf,-Inf): expected false")
		}
		// In Go, 0.0 == -0.0 is true; we keep that semantics.
		if !cofly.Equal(0.0, math.Copysign(0.0, -1.0)) {
			t.Fatalf("Equal(0.0,-0.0): expected true")
		}
	})

	t.Run("maps-deep-equality", func(t *testing.T) {
		a := map[string]any{
			"b": 2,
			"a": 1,
			"nested": map[string]any{
				"x": 1,
				"y": []any{1, "two", map[string]any{"z": 3}},
			},
		}
		b := map[string]any{
			"a": 1,
			"b": 2,
			"nested": map[string]any{
				"x": 1,
				"y": []any{1, "two", map[string]any{"z": 3}},
			},
		}

		if !cofly.Equal(a, b) {
			t.Fatalf("expected maps to be equal")
		}

		b["b"] = 3
		if cofly.Equal(a, b) {
			t.Fatalf("expected maps to be not equal after mutation")
		}
	})

	t.Run("arrays-deep-equality", func(t *testing.T) {
		a := []any{
			"head",
			1,
			1.0, // int-float equivalence should still work
			map[string]any{"x": 1},
			[]any{"a", "b"},
		}
		b := []any{
			"head",
			1,
			1, // compare float64(1.0) to int(1)
			map[string]any{"x": 1},
			[]any{"a", "b"},
		}

		if !cofly.Equal(a, b) {
			t.Fatalf("expected arrays to be equal")
		}

		b[3] = map[string]any{"x": 2}
		if cofly.Equal(a, b) {
			t.Fatalf("expected arrays to be not equal after nested change")
		}
	})

	t.Run("symmetry-on-supported-samples", func(t *testing.T) {
		samples := [][2]any{
			{nil, nil},
			{1, 1.0},
			{true, true},
			{"x", "x"},
			{[]any{1, "a"}, []any{1.0, "a"}},
			{map[string]any{"a": 1}, map[string]any{"a": 1.0}},
		}

		for _, s := range samples {
			a, b := s[0], s[1]
			if cofly.Equal(a, b) != cofly.Equal(b, a) {
				t.Fatalf("expected symmetry for (%#v, %#v)", a, b)
			}
		}
	})

	t.Run("unsupported-types-return-false", func(t *testing.T) {
		type S struct{ A int }

		if cofly.Equal(S{A: 1}, S{A: 1}) {
			t.Fatalf("expected Equal(unsupported,unsupported)=false")
		}
		if cofly.Equal(S{A: 1}, 1) {
			t.Fatalf("expected Equal(unsupported,int)=false")
		}
		if cofly.Equal(1, S{A: 1}) {
			t.Fatalf("expected Equal(int,unsupported)=false")
		}
	})

	t.Run("equal-iff-difference-undefined-on-supported-values", func(t *testing.T) {
		values := []any{
			nil,
			true,
			false,
			0,
			1,
			-1,
			0.0,
			1.0,
			-1.0,
			math.Inf(1),
			math.Inf(-1),
			math.NaN(),
			"",
			"x",
			[]any{},
			[]any{1, "a", 2.0, map[string]any{"k": 1}},
			map[string]any{},
			map[string]any{"a": 1, "nested": map[string]any{"x": 1.0, "arr": []any{true, "z"}}},
		}

		for i := range values {
			for j := range values {
				a, b := values[i], values[j]

				// Difference is not defined for every cross-type pair. In particular, it panics when
				// newValue is map[string]any and oldValue is []any (by design in Difference()).
				if _, ok := a.([]any); ok {
					if _, ok := b.(map[string]any); ok {
						continue
					}
				}

				eq := cofly.Equal(a, b)
				diff := cofly.Difference(a, b)
				isU := diff == cofly.Undefined

				if eq != isU {
					t.Fatalf("Equal/Difference mismatch: a=%#v b=%#v Equal=%v Difference=%#v", a, b, eq, diff)
				}
			}
		}
	})
}
