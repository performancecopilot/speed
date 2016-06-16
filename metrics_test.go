package speed

import "testing"

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
		_ = getHash(strings[i%l], 0)
	}
}
