package speed

import (
	"strings"
	"testing"
)

func BenchmarkGetHash(b *testing.B) {
	strings := []string{
		"a",
		"abcdefghijklmnopqrstuvwxyz",
		"aaaaaaaaaaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"abcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxy",
	}

	l := len(strings)
	for i := 0; i < b.N; i++ {
		_ = hash(strings[i%l], 0)
	}
}

//lint:ignore U1000 keeping for now
type testWriter struct {
	message string
	t       testing.TB
}

func (w *testWriter) Write(b []byte) (int, error) {
	s := string(b)
	if !strings.Contains(s, w.message) {
		w.t.Error("expected log'", string(b), "' to contain", w.message)
	}

	return len(b), nil
}
