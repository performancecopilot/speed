package bytebuffer

import (
	"os"
	"path"
	"testing"
)

func TestMemoryMappedBuffer(t *testing.T) {
	filename := "bytebuffer_memorymappedbuffer_test.tmp"
	loc := path.Join(os.TempDir(), filename)

	if _, err := os.Stat(loc); err == nil {
		err = os.Remove(loc)
		if err != nil {
			t.Error("Cannot proceed with test as cannot remove spec file")
			return
		}
	}

	b, err := NewMemoryMappedBuffer(loc, 10)
	if err != nil {
		t.Error("Cannot proceed with test as create buffer\n", err)
		return
	}

	if _, err = os.Stat(loc); err != nil {
		t.Errorf("No File created at %v despite the Buffer being initialized", loc)
		return
	}

	b.MustSetPos(5)
	err = b.WriteString("x")
	if err != nil {
		t.Error("Cannot Write to MemoryMappedBuffer")
		return
	}

	reader, err := os.Open(loc)
	data := make([]byte, 10)
	_, err = reader.Read(data)
	if err != nil {
		t.Error("Cannot read data from memory mapped file")
		return
	}

	if data[5] != 'x' {
		t.Error("Data Written in buffer not getting reflected in file")
	}

	err = b.Unmap(true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(loc); err == nil {
		t.Error("Memory Mapped File not getting deleted on Unmap")
	}
}
