package bytewriter

import (
	"os"
	"path"
	"testing"
)

func TestMemoryMappedWriter(t *testing.T) {
	filename := "bytebuffer_memorymappedwriter_test.tmp"
	loc := path.Join(os.TempDir(), filename)

	if _, err := os.Stat(loc); err == nil {
		err = os.Remove(loc)
		if err != nil {
			t.Error("Cannot proceed with test as cannot remove spec file")
			return
		}
	}

	w, err := NewMemoryMappedWriter(loc, 10)
	if err != nil {
		t.Error("Cannot proceed with test as create writer failed:", err)
		return
	}

	if _, err = os.Stat(loc); err != nil {
		t.Errorf("No File created at %v despite the Buffer being initialized", loc)
		return
	}

	_, err = w.WriteString("x", 5)
	if err != nil {
		t.Error("Cannot Write to MemoryMappedWriter")
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

	err = w.Unmap(true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(loc); err == nil {
		t.Error("Memory Mapped File not getting deleted on Unmap")
	}
}
