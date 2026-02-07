package cofly_test

import (
	"reflect"
	"testing"

	"github.com/rnkv/cofly-go"
)

func TestMerge(t *testing.T) {
	t.Run("nil-target-returns-change", func(t *testing.T) {
		if got := cofly.Merge(nil, 123, true); got != 123 {
			t.Fatalf("expected 123, got %#v", got)
		}

		change := map[string]any{"a": 1}
		got := cofly.Merge(nil, change, true)
		if !reflect.DeepEqual(got, change) {
			t.Fatalf("expected %#v, got %#v", change, got)
		}
	})

	t.Run("primitives-replace", func(t *testing.T) {
		if got := cofly.Merge(true, false, true); got != false {
			t.Fatalf("expected false, got %#v", got)
		}
		if got := cofly.Merge(1, 2, true); got != 2 {
			t.Fatalf("expected 2, got %#v", got)
		}
		if got := cofly.Merge(1.5, 2.5, true); got != 2.5 {
			t.Fatalf("expected 2.5, got %#v", got)
		}
		if got := cofly.Merge("x", "y", true); got != "y" {
			t.Fatalf("expected y, got %#v", got)
		}
	})

	t.Run("merge-undefined-string-panics", func(t *testing.T) {
		if got := cofly.Merge("x", cofly.Undefined, true); got != "x" {
			t.Fatalf("expected x, got %#v", got)
		}
	})

	t.Run("map-merge-add-update-delete", func(t *testing.T) {
		target := map[string]any{"a": 1, "b": 2}
		change := map[string]any{"b": 3, "c": "new"}
		want := map[string]any{"a": 1, "b": 3, "c": "new"}

		got := cofly.Merge(cofly.Clone(target), change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("map-merge-doClean-controls-deletion", func(t *testing.T) {
		target := map[string]any{"a": 1, "b": 2}
		change := map[string]any{"b": cofly.Undefined}

		gotClean := cofly.Merge(cofly.Clone(target), change, true).(map[string]any)
		if _, ok := gotClean["b"]; ok {
			t.Fatalf("expected b removed with doClean=true, got %#v", gotClean["b"])
		}

		gotKeep := cofly.Merge(cofly.Clone(target), change, false).(map[string]any)
		if gotKeep["b"] != cofly.Undefined {
			t.Fatalf("expected b=Undefined with doClean=false, got %#v", gotKeep["b"])
		}
	})

	t.Run("map-merge-nested-with-deletion", func(t *testing.T) {
		target := map[string]any{
			"nested": map[string]any{"a": 1, "b": 2},
		}
		change := map[string]any{
			"nested": map[string]any{"a": 2, "b": cofly.Undefined},
		}
		want := map[string]any{
			"nested": map[string]any{"a": 2},
		}

		got := cofly.Merge(cofly.Clone(target), change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("map-or-array-change-nil", func(t *testing.T) {
		if got := cofly.Merge(map[string]any{"a": 1}, (map[string]any)(nil), true); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
		if got := cofly.Merge([]any{"a"}, ([]any)(nil), true); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}

		// untyped nil should also work (regression test for nil interface change)
		if got := cofly.Merge(map[string]any{"a": 1}, nil, true); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
		if got := cofly.Merge([]any{"a"}, nil, true); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("splices-into-array-replace", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"1..2": []any{"B"},
		}
		want := []any{"a", "B", "c"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-append", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"3..": []any{"d"},
		}
		want := []any{"a", "b", "c", "d"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-insert-at-start", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"0..": []any{"x", "y"},
		}
		want := []any{"x", "y", "a", "b", "c"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-delete", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"1..2": []any{},
		}
		want := []any{"a", "c"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-replace-and-insert-in-one-splice", func(t *testing.T) {
		target := []any{"a"}
		change := map[string]any{
			"0..1": []any{"A", "extra"},
		}
		want := []any{"A", "extra"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-element-level-patch", func(t *testing.T) {
		target := []any{
			map[string]any{"nested": map[string]any{"a": 1}},
			"x",
		}
		change := map[string]any{
			"0..1": []any{map[string]any{"nested": map[string]any{"a": 2}}},
		}
		want := []any{
			map[string]any{"nested": map[string]any{"a": 2}},
			"x",
		}

		got := cofly.Merge(cofly.Clone(target), change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-multiple-splices-replace-and-append", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"3..":  []any{"d"},
			"1..2": []any{"B"},
		}
		want := []any{"a", "B", "c", "d"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-multiple-splices-insert-and-replace", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"2..3": []any{"C"},
			"0..":  []any{"x"},
		}
		want := []any{"x", "a", "b", "C"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-multiple-splices-delete-and-append", func(t *testing.T) {
		target := []any{"a", "b", "c", "d"}
		change := map[string]any{
			"1..3": []any{},    // delete "b","c"
			"4..":  []any{"e"}, // append
		}
		want := []any{"a", "d", "e"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-negative-span-before-array", func(t *testing.T) {
		target := []any{"a", "b"}
		change := map[string]any{
			"-1..0": []any{"x", "y"},
		}
		want := []any{"x", "y", "a", "b"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-after-end-with-gap", func(t *testing.T) {
		target := []any{"a", "b", "c"}
		change := map[string]any{
			"5..": []any{"x"},
		}
		want := []any{"a", "b", "c", "x"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-span-past-end-deletes-tail", func(t *testing.T) {
		target := []any{"a", "b", "c", "d"}
		change := map[string]any{
			"2..10": []any{"X"},
		}
		// span inside array is [2..4), so "c","d" are deleted, and "c" is replaced by "X".
		want := []any{"a", "b", "X"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("splices-into-array-overlapping-spans-panics", func(t *testing.T) {
		target := []any{"a", "b", "c", "d"}
		change := map[string]any{
			"1..3": []any{"X"},
			"2..4": []any{"Y"}, // overlaps with 1..3
		}

		mustPanic(t, func() {
			_ = cofly.Merge(target, change, true)
		})
	})

	t.Run("array-merge-non-splices-map-replaces", func(t *testing.T) {
		target := []any{"a", "b"}
		change := map[string]any{
			"0..":      []any{"x"},
			"not-span": []any{"y"}, // makes it not splices
		}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, change) {
			t.Fatalf("expected full replacement with %#v, got %#v", change, got)
		}
	})

	t.Run("array-merge-array-replaces", func(t *testing.T) {
		target := []any{"a", "b"}
		change := []any{"x"}
		want := []any{"x"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("map-merge-array-replaces", func(t *testing.T) {
		target := map[string]any{"a": 1}
		change := []any{"x"}
		want := []any{"x"}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})

	t.Run("array-merge-map-replaces", func(t *testing.T) {
		target := []any{"a"}
		change := map[string]any{"a": 1}
		want := map[string]any{"a": 1}

		got := cofly.Merge(target, change, true)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	})
}
