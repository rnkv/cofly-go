package cofly_test

import (
	"reflect"
	"testing"

	"github.com/rnkv/cofly-go"
)

func TestApply(t *testing.T) {
	t.Run("snapshot-unchanged-returns-false-and-sets-change-undefined", func(t *testing.T) {
		var target any = map[string]any{"a": 1, "b": 2}
		var change any = map[string]any{"a": 1, "b": 2} // snapshot

		applied := cofly.Apply(&target, true, &change, true)
		if applied {
			t.Fatalf("expected applied=false, got true")
		}
		if change != cofly.Undefined {
			t.Fatalf("expected change=Undefined, got %#v", change)
		}
		if !reflect.DeepEqual(target, map[string]any{"a": 1, "b": 2}) {
			t.Fatalf("target unexpectedly changed: %#v", target)
		}
	})

	t.Run("snapshot-changed-returns-true-updates-target-and-sets-change-to-diff", func(t *testing.T) {
		oldTarget := map[string]any{"a": 1, "b": 2}
		newSnapshot := map[string]any{"a": 1, "b": 3, "c": "new"}

		var target any = cofly.Clone(oldTarget)
		var change any = cofly.Clone(newSnapshot) // snapshot
		expectedDiff := cofly.Difference(oldTarget, newSnapshot)

		applied := cofly.Apply(&target, true, &change, true)
		if !applied {
			t.Fatalf("expected applied=true, got false")
		}
		if !reflect.DeepEqual(target, newSnapshot) {
			t.Fatalf("expected target updated to snapshot %#v, got %#v", newSnapshot, target)
		}
		if !reflect.DeepEqual(change, expectedDiff) {
			t.Fatalf("expected change set to diff %#v, got %#v", expectedDiff, change)
		}
	})

	t.Run("patch-mode-merges-target-does-not-mutate-change", func(t *testing.T) {
		oldTarget := map[string]any{"a": 1, "b": 2}
		patch := map[string]any{"b": 3}

		var target any = cofly.Clone(oldTarget)
		var change any = cofly.Clone(patch)
		changeBefore := cofly.Clone(change)

		applied := cofly.Apply(&target, false, &change, true)
		if !applied {
			t.Fatalf("expected applied=true, got false")
		}
		if !reflect.DeepEqual(target, map[string]any{"a": 1, "b": 3}) {
			t.Fatalf("unexpected target: %#v", target)
		}
		if !reflect.DeepEqual(change, changeBefore) {
			t.Fatalf("expected change not mutated; before=%#v after=%#v", changeBefore, change)
		}
	})

	t.Run("patch-mode-doClean-controls-deletion", func(t *testing.T) {
		oldTarget := map[string]any{"a": 1, "b": 2}
		patch := map[string]any{"b": cofly.Undefined}

		{
			var target any = cofly.Clone(oldTarget)
			var change any = cofly.Clone(patch)
			_ = cofly.Apply(&target, false, &change, true)
			got := target.(map[string]any)
			if _, ok := got["b"]; ok {
				t.Fatalf("expected b removed with doClean=true, got %#v", got["b"])
			}
		}

		{
			var target any = cofly.Clone(oldTarget)
			var change any = cofly.Clone(patch)
			_ = cofly.Apply(&target, false, &change, false)
			got := target.(map[string]any)
			if got["b"] != cofly.Undefined {
				t.Fatalf("expected b=Undefined with doClean=false, got %#v", got["b"])
			}
		}
	})

	t.Run("patch-mode-nil-change-sets-target-to-nil", func(t *testing.T) {
		var target any = map[string]any{"a": 1}
		var change any = nil
		applied := cofly.Apply(&target, false, &change, true)
		if !applied {
			t.Fatalf("expected applied=true")
		}
		if target != nil {
			t.Fatalf("expected target=nil, got %#v", target)
		}
	})
}
