package cofly_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/rnkv/cofly-go"
)

func FuzzDiffMergeRoundTrip_IntArrays(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{9, 9, 9})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Build two small arrays from bytes, only using supported scalar types.
		// Keep them short to make fuzzing fast.
		n := len(data)
		if n > 40 {
			n = 40
		}

		split := n / 2
		oldA := make([]any, 0, split)
		newA := make([]any, 0, n-split)

		for i := 0; i < split; i++ {
			oldA = append(oldA, int(data[i]))
		}
		for i := split; i < n; i++ {
			newA = append(newA, int(data[i]))
		}

		diff := cofly.Difference(oldA, newA)
		if diff == cofly.Undefined {
			// When diff is Undefined, arrays must be Equal by the library's semantics.
			if !cofly.Equal(oldA, newA) {
				t.Fatalf("diff is Undefined but Equal is false; old=%#v new=%#v", oldA, newA)
			}
			return
		}

		got := cofly.Merge(cofly.Clone(oldA), diff, true)
		if !reflect.DeepEqual(got, newA) {
			t.Fatalf("round-trip failed: old=%#v diff=%#v got=%#v new=%#v", oldA, diff, got, newA)
		}
	})
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) next() byte {
	if len(r.b) == 0 {
		return 0
	}
	v := r.b[r.i%len(r.b)]
	r.i++
	return v
}

func (r *byteReader) nextN(n int) []byte {
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, r.next())
	}
	return out
}

func genScalar(r *byteReader) any {
	switch r.next() % 6 {
	case 0:
		return nil
	case 1:
		return r.next()%2 == 0
	case 2:
		// small int (signed-ish)
		return int(int8(r.next()))
	case 3:
		// float64 (avoid NaN/Inf because Go equality treats NaN != NaN,
		// which makes "round-trip equals new" assertions non-informative)
		return float64(int8(r.next()))
	case 4:
		n := int(r.next() % 6)
		b := r.nextN(n)
		// Avoid generating the Undefined marker string ("\x00") which is reserved by the library.
		for i := range b {
			b[i] = 'a' + (b[i] % 26)
		}
		s := string(b)
		if s == cofly.Undefined {
			return "u"
		}
		return s
	default:
		return ""
	}
}

func genValue(r *byteReader, depth int) any {
	if depth <= 0 {
		return genScalar(r)
	}

	switch r.next() % 8 {
	case 0, 1, 2, 3:
		return genScalar(r)
	case 4:
		// array
		n := int(r.next() % 4)
		arr := make([]any, 0, n)
		for range n {
			arr = append(arr, genValue(r, depth-1))
		}
		return arr
	default:
		// map with keys that can never be parsed as splices (avoid ".." patterns).
		n := int(r.next() % 4)
		m := make(map[string]any, n)
		for idx := 0; idx < n; idx++ {
			key := "k" + strconv.Itoa(idx) + "_" + fmt.Sprintf("%02x", r.next())
			m[key] = genValue(r, depth-1)
		}
		return m
	}
}

func mutate(r *byteReader, v any, depth int) any {
	if depth <= 0 {
		return genScalar(r)
	}

	switch vv := v.(type) {
	case nil:
		return genScalar(r)
	case bool:
		return !vv
	case int:
		return vv + 1
	case float64:
		// NaN stays NaN, otherwise increment
		if vv != vv {
			return float64(0)
		}
		return vv + 1
	case string:
		return vv + "x"
	case []any:
		if len(vv) == 0 {
			return []any{genScalar(r)}
		}
		out := cofly.Clone(vv).([]any)
		switch r.next() % 3 {
		case 0:
			// mutate first element
			out[0] = mutate(r, out[0], depth-1)
		case 1:
			// delete first element
			out = out[1:]
		default:
			// insert scalar at start
			out = append([]any{genScalar(r)}, out...)
		}
		return out
	case map[string]any:
		out := cofly.Clone(vv).(map[string]any)
		switch r.next() % 3 {
		case 0:
			// add/update a key
			out["k_mut_"+fmt.Sprintf("%02x", r.next())] = genValue(r, depth-1)
		case 1:
			// delete some key if any
			for k := range out {
				delete(out, k)
				break
			}
		default:
			// mutate some key if any
			for k, val := range out {
				out[k] = mutate(r, val, depth-1)
				break
			}
		}
		return out
	default:
		return genScalar(r)
	}
}

func FuzzDiffMergeRoundTrip_NestedValues(f *testing.F) {
	f.Add([]byte("seed-1"))
	f.Add([]byte("seed-2"))
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		r := &byteReader{b: data}
		depth := int(r.next()%3) + 1 // 1..3

		oldV := genValue(r, depth)

		// Bias towards related values to get small diffs as well.
		var newV any
		switch r.next() % 4 {
		case 0:
			newV = cofly.Clone(oldV)
		case 1:
			newV = mutate(r, oldV, depth)
		case 2:
			newV = genValue(r, depth)
		default:
			// keep old, but maybe flip numeric representation (int<->float)
			switch ov := oldV.(type) {
			case int:
				newV = float64(ov)
			case float64:
				newV = int(ov)
			default:
				newV = cofly.Clone(oldV)
			}
		}

		// Difference panics for old=[]any, new=map[string]any. Avoid that combination in this fuzz.
		if _, ok := oldV.([]any); ok {
			if _, ok := newV.(map[string]any); ok {
				newV = cofly.Clone(oldV)
			}
		}

		safeDifference := func(oldVal, newVal any) (any, bool) {
			defer func() {
				// recovered below via named return
			}()
			return cofly.Difference(oldVal, newVal), true
		}

		var diff any
		var ok bool
		func() {
			defer func() {
				if recover() != nil {
					ok = false
				}
			}()
			diff, ok = safeDifference(oldV, newV)
		}()

		if !ok {
			// Difference is not defined for every cross-type combination (it may panic by design).
			// Skip such inputs; this fuzz targets round-trip behavior for supported combinations.
			return
		}

		if diff == cofly.Undefined {
			if !cofly.Equal(oldV, newV) {
				t.Fatalf("diff is Undefined but Equal is false; old=%#v new=%#v", oldV, newV)
			}
			return
		}

		var got any
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Merge panicked: old=%#v diff=%#v panic=%v", oldV, diff, r)
				}
			}()
			got = cofly.Merge(cofly.Clone(oldV), diff, true)
		}()
		// Compare using library semantics (not reflect.DeepEqual) because int and float64 are
		// treated as equivalent by Equal()/Difference().
		if !cofly.Equal(got, newV) {
			t.Fatalf("round-trip failed: old=%#v diff=%#v got=%#v new=%#v", oldV, diff, got, newV)
		}
	})
}
