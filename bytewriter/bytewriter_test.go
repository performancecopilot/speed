package bytewriter

import "testing"

func TestWriteInt32(t *testing.T) {
	cases := []int32{0, 10, 100, 200, 1000, 10000, 10000000, 1000000000, 2147483647}

	for _, val := range cases {
		b := NewByteWriter(4)

		off, err := b.WriteInt32(val, 0)

		if err != nil {
			t.Error(err)
			return
		}

		if off != 4 {
			t.Error("expected offset to be 4")
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

func TestWriteInt64(t *testing.T) {
	cases := []int64{0, 10, 100, 200, 1000, 10000, 10000000, 1000000000, 2147483647,
		4294967295, 10000000000000, 100000000000000000, 9223372036854775807}

	for _, val := range cases {
		w := NewByteWriter(8)

		off, err := w.WriteInt64(val, 0)
		if err != nil {
			t.Error(err)
			return
		}

		if off != 8 {
			t.Error("Not Writing 8 bytes for int32")
			return
		}

		e := []byte{
			byte(val & 0xFF),
			byte((val >> 8) & 0xFF),
			byte((val >> 16) & 0xFF),
			byte((val >> 24) & 0xFF),
			byte((val >> 32) & 0xFF),
			byte((val >> 40) & 0xFF),
			byte((val >> 48) & 0xFF),
			byte(val >> 56),
		}

		for i := 0; i < 8; i++ {
			if w.buffer[i] != e[i] {
				t.Errorf("pos: %v, expected: %v, got %v", i, e[i], w.buffer[i])
			}
		}
	}
}

func TestWriteString(t *testing.T) {
	cases := []string{"MMV", "Suyash", "This is a little long string"}
	for _, val := range cases {
		w := NewByteWriter(len(val))

		off, err := w.WriteString(val, 0)
		if err != nil {
			t.Error(err)
			return
		}

		if off != len(val) {
			t.Errorf("Expected to write %v bytes, writing %v bytes", len(val), off)
			return
		}

		e := []byte(val)
		for i := 0; i < len(val); i++ {
			if w.buffer[i] != e[i] {
				t.Errorf("pos: %v, expected: %v, got %v", i, e[i], w.buffer[i])
			}
		}
	}
}

func TestOffset(t *testing.T) {
	w := NewByteWriter(4)

	off, err := w.WriteString("a", 2)
	if err != nil {
		t.Error("Did not Expect error in writing a value inside the buffer")
		return
	}

	if off != 3 {
		t.Error("Position not changing as expected")
		return
	}

	if w.Bytes()[2] != 'a' {
		t.Error("Value was not written at the expected position")
		return
	}

	off, err = w.WriteInt32(10, 2)
	if err == nil {
		t.Error("Expected error in writing a value guaranteed to overflow")
		return
	}

	if off != -1 {
		t.Error("expected write failure to return -1")
		return
	}
}
