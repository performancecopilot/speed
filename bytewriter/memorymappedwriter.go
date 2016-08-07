package bytewriter

import (
	"fmt"
	"os"
	"path"
	"syscall"
)

// MemoryMappedWriter is a ByteBuffer that is also mapped into memory
type MemoryMappedWriter struct {
	*ByteWriter
	loc  string // location of the memory mapped file
	size int    // size in bytes
}

// NewMemoryMappedWriter will create and return a new instance of a MemoryMappedWriter
func NewMemoryMappedWriter(loc string, size int) (*MemoryMappedWriter, error) {
	if _, err := os.Stat(loc); err == nil {
		err = os.Remove(loc)
		if err != nil {
			return nil, err
		}
	}

	// ensure destination directory exists
	dir := path.Dir(loc)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(loc, syscall.O_CREAT|syscall.O_RDWR|syscall.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}

	l, err := f.Write(make([]byte, size))
	if err != nil {
		return nil, err
	}
	if l < size {
		return nil, fmt.Errorf("Could not initialize %d bytes", size)
	}

	b, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &MemoryMappedWriter{
		NewByteWriterSlice(b),
		loc,
		size,
	}, nil
}

// Unmap will manually delete the memory mapping of a mapped buffer
func (b *MemoryMappedWriter) Unmap(removefile bool) error {
	if err := syscall.Munmap(b.buffer); err != nil {
		return err
	}

	if removefile {
		if err := os.Remove(b.loc); err != nil {
			return err
		}
	}

	return nil
}
