package bytebuffer

import (
	"os"
	"syscall"
)

type MemoryMappedBuffer struct {
	*ByteBuffer
	loc  string // location of the memory mapped file
	size int    // size in bytes
}

func NewMemoryMappedBuffer(loc string, size int) (*MemoryMappedBuffer, error) {
	f, err := os.Create(loc)
	if err != nil {
		return nil, err
	}

	f.Write(make([]byte, size))

	fd := int(f.Fd())

	b, err := syscall.Mmap(fd, 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &MemoryMappedBuffer{
		NewByteBufferSlice(b),
		loc,
		size,
	}, nil
}

func (b *MemoryMappedBuffer) Unmap(removefile bool) error {
	if removefile {
		os.Remove(b.loc)
	}

	return syscall.Munmap(b.buffer)
}
