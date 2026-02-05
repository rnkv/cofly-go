package cofly_test

import (
	"reflect"
	"testing"

	"github.com/rnkv/cofly-go"
)

func TestClone(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := cofly.Clone(nil); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("deep-clone-map-and-array", func(t *testing.T) {
		orig := map[string]any{
			"a": 1,
			"nested": map[string]any{
				"x": 1,
				"arr": []any{
					"v",
					map[string]any{"z": 3},
				},
			},
		}

		cl := cofly.Clone(orig).(map[string]any)

		// mutate original deeply
		orig["a"] = 2
		origNested := orig["nested"].(map[string]any)
		origNested["x"] = 2
		origArr := origNested["arr"].([]any)
		origArr[0] = "changed"
		origArr[1].(map[string]any)["z"] = 4

		wantClone := map[string]any{
			"a": 1,
			"nested": map[string]any{
				"x": 1,
				"arr": []any{
					"v",
					map[string]any{"z": 3},
				},
			},
		}

		if !reflect.DeepEqual(cl, wantClone) {
			t.Fatalf("clone was affected by mutations: want %#v, got %#v", wantClone, cl)
		}
	})

	t.Run("unsupported-types-panic", func(t *testing.T) {
		type S struct{ A int }
		mustPanic(t, func() {
			_ = cofly.Clone(S{A: 1})
		})
	})
}

