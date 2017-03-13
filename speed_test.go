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

func TestSetLogWriters(t *testing.T) {
	cases := []string{
		"a",
		"abcdefghijklmnopqrstuvwxyz",
		"aaaaaaaaaaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"abcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxyabcdefghijklmnopqrstuvwxy",
	}

	for _, s := range cases {
		w := &testWriter{s, t}
		SetLogWriters(w)

		if len(logWriters) != 1 {
			t.Error("expected the length of logWriters to be 1")
		}

		logger.Info(s)
	}
}

func TestAddLogWriters(t *testing.T) {
	if len(logWriters) != 1 {
		t.Error("expected the length of logWriters to be 1")
	}

	AddLogWriter(&testWriter{"a", t})

	if len(logWriters) != 2 {
		t.Error("expected the length of logWriters to be 2")
	}

	AddLogWriter(&testWriter{"b", t})

	if len(logWriters) != 3 {
		t.Error("expected the length of logWriters to be 3")
	}
}
