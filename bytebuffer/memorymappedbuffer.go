package bytebuffer

import (
	"fmt"
	"os"
	"syscall"
)

// MemoryMappedBuffer is a ByteBuffer that is also mapped into memory
type MemoryMappedBuffer struct {
	*ByteBuffer
	loc  string // location of the memory mapped file
	size int    // size in bytes
}

// NewMemoryMappedBuffer will create and return a new instance of a MemoryMappedBuffer
func NewMemoryMappedBuffer(loc string, size int) (*MemoryMappedBuffer, error) {
	if _, err := os.Stat(loc); err == nil {
		err = os.Remove(loc)
		if err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(loc, syscall.O_CREAT|syscall.O_RDWR|syscall.O_EXCL, 0666)
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

	return &MemoryMappedBuffer{
		NewByteBufferSlice(b),
		loc,
		size,
	}, nil
}

// Unmap will manually delete the memory mapping of a mapped buffer
func (b *MemoryMappedBuffer) Unmap(removefile bool) error {
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
