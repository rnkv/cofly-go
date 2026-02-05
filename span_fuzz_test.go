package cofly

import "testing"

func FuzzParseSpan(f *testing.F) {
	seeds := []string{
		"0..",
		"0..1",
		"-1..0",
		"2..10",
		"not-a-span",
		"..",
		"1",
		"",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		sp, ok := parseSpan(s)
		if !ok {
			return
		}
		// Round-trip should succeed and preserve the parsed span.
		key := sp.string()
		sp2, ok2 := parseSpan(key)
		if !ok2 || sp2 != sp {
			t.Fatalf("round-trip failed: in=%q parsed=%#v key=%q reparsed=%#v ok=%v", s, sp, key, sp2, ok2)
		}
	})
}
