package cofly_test

import (
	"reflect"
	"testing"

	"github.com/rnkv/cofly-go"
)

func mustPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic, got none")
		}
	}()

	fn()
}

func TestDifference(t *testing.T) {
	apply := func(t *testing.T, old, new any) any {
		t.Helper()

		diff := cofly.Difference(old, new)
		if diff == cofly.Undefined {
			return cofly.Clone(old)
		}

		return cofly.Merge(cofly.Clone(old), diff, true)
	}

	t.Run("nil-handling", func(t *testing.T) {
		if got := cofly.Difference(nil, nil); got != cofly.Undefined {
			t.Fatalf("Difference(nil,nil): expected Undefined, got %#v", got)
		}

		if got := cofly.Difference(123, nil); got != nil {
			t.Fatalf("Difference(123,nil): expected nil, got %#v", got)
		}

		if got := cofly.Difference(nil, 123); got != 123 {
			t.Fatalf("Difference(nil,123): expected 123, got %#v", got)
		}
	})

	t.Run("composite-to-nil", func(t *testing.T) {
		oldM := map[string]any{"a": 1}
		if got := cofly.Difference(oldM, nil); got != nil {
			t.Fatalf("Difference(map,nil): expected nil, got %#v", got)
		}

		oldA := []any{"a"}
		if got := cofly.Difference(oldA, nil); got != nil {
			t.Fatalf("Difference(array,nil): expected nil, got %#v", got)
		}

		// "set to nil" should be applicable via Merge as well.
		if got := apply(t, oldM, nil); got != nil {
			t.Fatalf("expected Merge(Clone(map), Difference(map,nil)) == nil, got %#v", got)
		}
		if got := apply(t, oldA, nil); got != nil {
			t.Fatalf("expected Merge(Clone(array), Difference(array,nil)) == nil, got %#v", got)
		}
	})

	t.Run("primitives-equality", func(t *testing.T) {
		testCases := []struct {
			name string
			old  any
			new  any
			want any
		}{
			{"bool-same", true, true, cofly.Undefined},
			{"bool-diff", false, true, true},
			{"int-same", 1, 1, cofly.Undefined},
			{"int-diff", 1, 2, 2},
			{"float-same", 1.5, 1.5, cofly.Undefined},
			{"float-diff", 1.5, 2.5, 2.5},
			{"string-same", "x", "x", cofly.Undefined},
			{"string-diff", "x", "y", "y"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				got := cofly.Difference(testCase.old, testCase.new)

				if !reflect.DeepEqual(got, testCase.want) {
					t.Fatalf("want %#v, got %#v", testCase.want, got)
				}
			})
		}
	})

	t.Run("int-float64-equivalence", func(t *testing.T) {
		if got := cofly.Difference(1, 1.0); got != cofly.Undefined {
			t.Fatalf("Difference(1,1.0): expected Undefined, got %#v", got)
		}

		if got := cofly.Difference(1.0, 1); got != cofly.Undefined {
			t.Fatalf("Difference(1.0,1): expected Undefined, got %#v", got)
		}

		if got := cofly.Difference(1.0, 2); got != 2 {
			t.Fatalf("Difference(1.0,2): expected 2, got %#v", got)
		}
	})

	t.Run("maps-add-update-delete", func(t *testing.T) {
		oldM := map[string]any{"a": 1, "b": 2}
		newM := map[string]any{"a": 1, "b": 3, "c": "new"}

		got := cofly.Difference(oldM, newM)
		gotM, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map diff, got %#v", got)
		}

		if gotM["b"] != 3 || gotM["c"] != "new" {
			t.Fatalf("unexpected diff: %#v", got)
		}
		if _, exists := gotM["a"]; exists {
			t.Fatalf("expected a absent from diff when unchanged, got %#v", gotM["a"])
		}

		// deletion
		oldM2 := map[string]any{"a": 1, "b": 2}
		newM2 := map[string]any{"a": 1}
		got2 := cofly.Difference(oldM2, newM2).(map[string]any)
		if got2["b"] != cofly.Undefined {
			t.Fatalf("expected b=Undefined marker for deletion, got %#v", got2["b"])
		}
	})

	t.Run("arrays-add-update-delete", func(t *testing.T) {
		oldA := []any{"a", "b", "c"}
		newA := []any{"a", "B", "c", "d"}

		expected := map[string]any{
			"1..2": []any{"B"}, // replace "b" -> "B"
			"3..":  []any{"d"}, // append "d"
		}
		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}

		// deletion
		oldA2 := []any{"a", "b"}
		newA2 := []any{"a"}
		expected2 := map[string]any{
			"1..2": []any{}, // delete element at index 1
		}
		got2 := cofly.Difference(oldA2, newA2).(map[string]any)
		if !reflect.DeepEqual(got2, expected2) {
			t.Fatalf("expected %#v, got %#v", expected2, got2)
		}
	})

	t.Run("arrays-empty-old-fast-path", func(t *testing.T) {
		oldA := []any{}
		newA := []any{"a", "b"}
		expected := map[string]any{
			"0..": []any{"a", "b"},
		}

		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}
	})

	t.Run("arrays-empty-new-fast-path", func(t *testing.T) {
		oldA := []any{"a"}
		newA := []any{}
		expected := map[string]any{
			"0..1": []any{},
		}

		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}
	})

	t.Run("arrays-identical-returns-undefined", func(t *testing.T) {
		oldA := []any{
			1,
			"two",
			map[string]any{"nested": map[string]any{"x": 1}},
			[]any{"a", "b"},
		}
		newA := []any{
			1,
			"two",
			map[string]any{"nested": map[string]any{"x": 1}},
			[]any{"a", "b"},
		}

		if got := cofly.Difference(oldA, newA); got != cofly.Undefined {
			t.Fatalf("expected Undefined, got %#v", got)
		}
	})

	t.Run("array-insert-in-the-middle", func(t *testing.T) {
		oldA := []any{"a", "b", "c"}
		newA := []any{"a", "X", "b", "c"}
		expected := map[string]any{
			"1..": []any{"X"},
		}

		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}
	})

	t.Run("array-replacement-and-delete-in-one-splice", func(t *testing.T) {
		oldA := []any{
			map[string]any{"a": 1},
			map[string]any{"b": 1},
		}
		newA := []any{
			map[string]any{"a": 2},
		}

		expected := map[string]any{
			"0..2": []any{
				map[string]any{"a": 2}, // element-level patch for oldA[0]
			},
		}

		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}
	})

	t.Run("array-replacement-and-insert-in-one-splice", func(t *testing.T) {
		oldA := []any{"a"}
		newA := []any{"A", "extra"}
		expected := map[string]any{
			"0..1": []any{"A", "extra"},
		}

		got := cofly.Difference(oldA, newA).(map[string]any)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %#v, got %#v", expected, got)
		}
	})

	t.Run("array-difference-optimization-1", func(t *testing.T) {
		oldArray := []any{"a", "b", "c", "d"}
		newArray := []any{"b", "c", "d", "e"}
		expectedChange := map[string]any{
			"0..1": []any{},    // delete "a"
			"4..":  []any{"e"}, // append "e"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-prepend-1", func(t *testing.T) {
		oldArray := []any{"b", "c", "d"}
		newArray := []any{"a", "b", "c", "d"}
		expectedChange := map[string]any{
			"0..": []any{"a"}, // insert at start
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-delete-prefix-1", func(t *testing.T) {
		oldArray := []any{"a", "b", "c"}
		newArray := []any{"b", "c"}
		expectedChange := map[string]any{
			"0..1": []any{}, // delete prefix element
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-replace-prefix-1", func(t *testing.T) {
		oldArray := []any{"x", "b", "c"}
		newArray := []any{"y", "b", "c"}
		expectedChange := map[string]any{
			"0..1": []any{"y"}, // replace "x" -> "y"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-prefix-and-suffix", func(t *testing.T) {
		oldArray := []any{"a", "b", "c", "d", "e"}
		newArray := []any{"x", "b", "c", "d"}
		expectedChange := map[string]any{
			"0..1": []any{"x"}, // replace "a" -> "x"
			"4..5": []any{},    // delete "e"
		}

		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)
		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-delete-2-and-append-1", func(t *testing.T) {
		oldArray := []any{"a", "b", "c", "d"}
		newArray := []any{"c", "d", "e"}
		expectedChange := map[string]any{
			"0..2": []any{},    // delete "a","b"
			"4..":  []any{"e"}, // append "e"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-prepend-2-and-delete-tail-1", func(t *testing.T) {
		oldArray := []any{"b", "c", "d", "e"}
		newArray := []any{"x", "y", "b", "c", "d"}
		expectedChange := map[string]any{
			"0..":  []any{"x", "y"}, // insert at start
			"3..4": []any{},         // delete "e"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-overlap-single-element", func(t *testing.T) {
		oldArray := []any{"a", "b", "c"}
		newArray := []any{"x", "b", "y"}
		expectedChange := map[string]any{
			"0..1": []any{"x"}, // replace "a" -> "x"
			"2..3": []any{"y"}, // replace "c" -> "y"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-two-max-overlaps-tie-break-first", func(t *testing.T) {
		// Here are two max overlap of length 3: ["a","b","c"] and ["d","e","f"].
		// Due to the order of the overlap() will be selected the first: old i=1, new j=0.

		oldArray := []any{"x", "a", "b", "c", "y", "d", "e", "f"}
		newArray := []any{"a", "b", "c", "m", "d", "e", "f"}
		expectedChange := map[string]any{
			"0..1": []any{},    // delete "x"
			"4..5": []any{"m"}, // replace "y" -> "m"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-overlap-block-repeated-twice", func(t *testing.T) {
		// Overlap ["a","b","c"] appears twice in old and new.
		// The first match is selected: old i=1, new j=0.

		oldArray := []any{"x", "a", "b", "c", "y", "a", "b", "c", "z"}
		newArray := []any{"a", "b", "c", "m", "a", "b", "c"}
		expectedChange := map[string]any{
			"0..1": []any{},    // delete "x"
			"4..5": []any{"m"}, // replace "y" -> "m"
			"8..9": []any{},    // delete "z"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("array-difference-optimization-overlap-block-repeated-thrice", func(t *testing.T) {
		// Overlap ["a","b","c"] appears three times.

		oldArray := []any{"x", "a", "b", "c", "y", "a", "b", "c", "w", "a", "b", "c", "z"}
		newArray := []any{"a", "b", "c", "m", "a", "b", "c", "n", "a", "b", "c"}
		expectedChange := map[string]any{
			"0..1":   []any{},    // delete "x"
			"4..5":   []any{"m"}, // replace "y" -> "m"
			"8..9":   []any{"n"}, // replace "w" -> "n"
			"12..13": []any{},    // delete "z"
		}
		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})

	t.Run("maps-nested-identical-returns-undefined", func(t *testing.T) {
		oldM := map[string]any{"a": map[string]any{"x": 1}}
		newM := map[string]any{"a": map[string]any{"x": 1}}

		if got := cofly.Difference(oldM, newM); got != cofly.Undefined {
			t.Fatalf("expected Undefined, got %#v", got)
		}
	})

	t.Run("merge-after-difference-reconstructs-new-value", func(t *testing.T) {
		testCases := []struct {
			name string
			old  any
			new  any
		}{
			{
				name: "primitive-int",
				old:  1,
				new:  2,
			},
			{
				name: "map-add-update-delete",
				old:  map[string]any{"a": 1, "b": 2},
				new:  map[string]any{"a": 1, "b": 3, "c": "new"},
			},
			{
				name: "nested-map",
				old:  map[string]any{"nested": map[string]any{"a": 1, "b": 2}},
				new:  map[string]any{"nested": map[string]any{"a": 2}},
			},
			{
				name: "array-splices-and-nested-patch",
				old: []any{
					"head",
					map[string]any{"id": 1, "name": "alice"},
					42,
					"keep",
					map[string]any{"nested": map[string]any{"a": 1}},
					3.14,
					"tail",
				},
				new: []any{
					map[string]any{"id": 1, "name": "alice"},
					42,
					"keep",
					map[string]any{"nested": map[string]any{"a": 2}},
					3.14,
					"tail",
					"extra",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := apply(t, tc.old, tc.new)
				if !reflect.DeepEqual(got, tc.new) {
					t.Fatalf("expected %#v, got %#v", tc.new, got)
				}
			})
		}
	})

	t.Run("unsupported-types", func(t *testing.T) {
		type S struct{ A int }

		mustPanic(t, func() {
			_ = cofly.Difference(S{A: 1}, S{A: 1})
		})
	})

	t.Run("array-difference-optimization-mixed-types-and-objects", func(t *testing.T) {
		oldArray := []any{
			"head",
			map[string]any{"id": 1, "name": "alice"},
			42,
			"keep",
			map[string]any{"nested": map[string]any{"a": 1}},
			3.14,
			"tail",
		}

		newArray := []any{
			map[string]any{"id": 1, "name": "alice"}, // the same object by content
			42,                                       // number
			"keep",                                   // string
			map[string]any{"nested": map[string]any{"a": 2}}, // change inside the nested object
			3.14, // float64
			"tail",
			"extra", // new tail
		}

		expectedChange := map[string]any{
			"0..1": []any{}, // removed "head"
			"4..5": []any{
				map[string]any{"nested": map[string]any{"a": 2}}, // patch for the object at position 4
			},
			"7..": []any{"extra"}, // added to the end
		}

		gotChange := cofly.Difference(oldArray, newArray).(map[string]any)

		if !reflect.DeepEqual(gotChange, expectedChange) {
			t.Fatalf("expected %#v, got %#v", expectedChange, gotChange)
		}
	})
}
