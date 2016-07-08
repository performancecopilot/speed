package bytebuffer

import "testing"

func TestWriteInt32(t *testing.T) {
	cases := []int32{0, 10, 100, 200, 1000, 10000, 10000000, 1000000000, 2147483647}

	for _, val := range cases {
		b := NewByteBuffer(4)

		err := b.WriteInt32(val)
		if err != nil {
			t.Error(err)
			return
		}

		if b.Pos() != 4 {
			t.Error("Not Writing 4 bytes for int32")
			return
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
		b := NewByteBuffer(8)

		err := b.WriteInt64(val)
		if err != nil {
			t.Error(err)
			return
		}

		if b.Pos() != 8 {
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
			return
		}

		if b.Pos() != len(val) {
			t.Errorf("Expected to write %v bytes, writing %v bytes", len(val), b.Pos())
			return
		}

		e := []byte(val)
		for i := 0; i < len(val); i++ {
			if b.buffer[i] != e[i] {
				t.Errorf("pos: %v, expected: %v, got %v", i, e[i], b.buffer[i])
			}
		}
	}
}

func TestSetPos(t *testing.T) {
	b := NewByteBuffer(4)
	err := b.SetPos(4)
	if err == nil {
		t.Error("Expected error at setting a bytebuffer to a position outside its range")
	}

	b.SetPos(2)
	b.WriteString("a")

	if b.Pos() != 3 {
		t.Error("Position not changing as expected")
		return
	}

	if b.Bytes()[2] != 'a' {
		t.Error("Value was not written at the expected position")
		return
	}

	b.SetPos(2)
	err = b.WriteInt32(10)

	if err == nil {
		t.Error("Expected error in writing a value guaranteed to overflow")
		return
	}

	if b.Pos() != 2 {
		t.Error("Position changing despite a write failure")
		return
	}
}
