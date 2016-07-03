package bytebuffer

import "testing"

func TestWriteInt32(t *testing.T) {
	cases := []int32{0, 10, 100, 200, 1000, 10000, 10000000, 1000000000, 2147483647}

	for _, val := range cases {
		b := NewByteBuffer(4)

		err := b.WriteInt32(val)
		if err != nil {
			t.Error(err)
		}

		if b.Pos() != 4 {
			t.Error("Not Writing 4 bytes for int32")
		}

		e := []byte{
			byte(val & 0xFF),
			byte((val >> 8) & 0xFF),
			byte((val >> 16) & 0xFF),
			byte(val >> 24),
		}

		for i := 0; i < 4; i++ {
			if b.buffer[i] != e[i] {
				t.Errorf("pos: %v, expected: %v, got %v", i, e[i], b.buffer[i])
			}
		}
	}
}

func TestWriteString(t *testing.T) {
	cases := []string{"MMV", "Suyash", "This is a little long string"}
	for _, val := range cases {
		b := NewByteBuffer(len(val))

		err := b.WriteString(val)
		if err != nil {
			t.Error(err)
		}

		if b.Pos() != len(val) {
			t.Errorf("Expected to write %v bytes, writing %v bytes", len(val), b.Pos())
		}

		e := []byte(val)
		for i := 0; i < len(val); i++ {
			if b.buffer[i] != e[i] {
				t.Errorf("pos: %v, expected: %v, got %v", i, e[i], b.buffer[i])
			}
		}
	}
}
