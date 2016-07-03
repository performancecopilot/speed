package bytebuffer

import (
	"encoding/binary"
	"errors"
)

// assumes Little Endian, use _arch.go to set it to BigEndian for those archs
var byteOrder = binary.LittleEndian

// ByteBuffer is a simple wrapper over a byte slice that supports writing anywhere
type ByteBuffer struct {
	pos    int
	buffer []byte
}

// NewByteBuffer creates a new ByteBuffer of the specified size
func NewByteBuffer(n int) *ByteBuffer {
	return &ByteBuffer{
		pos:    0,
		buffer: make([]byte, n),
	}
}

// NewByteBufferSlice creates a new ByteBuffer using the passed slice
func NewByteBufferSlice(buffer []byte) *ByteBuffer {
	return &ByteBuffer{
		pos:    0,
		buffer: buffer,
	}
}

// Pos returns the current write position of the ByteBuffer
func (b *ByteBuffer) Pos() int { return b.pos }

// SetPos sets the write position of the ByteBuffer to the specified position
func (b *ByteBuffer) SetPos(position int) error {
	if position < 0 || position >= len(b.buffer) {
		// TODO: make a better error message
		return errors.New("Out of Range")
	}

	b.pos = position
	return nil
}

// Len returns the maximum size of the ByteBuffer
func (b *ByteBuffer) Len() int { return len(b.buffer) }

// Bytes returns the internal byte array of the ByteBuffer
func (b *ByteBuffer) Bytes() []byte { return b.buffer }

func (b *ByteBuffer) Write(data []byte) (int, error) {
	l := len(data)

	if b.Pos()+l > b.Len() {
		// TODO: make a better error message
		return 0, errors.New("Overflow")
	}

	for i := 0; i < l; i++ {
		b.buffer[b.pos+i] = data[i]
	}

	b.pos += l

	return l, nil
}

// WriteVal writes an arbitrary value to the buffer
func (b *ByteBuffer) WriteVal(val interface{}) error {
	return binary.Write(b, byteOrder, val)
}

// WriteString writes a string to the buffer
func (b *ByteBuffer) WriteString(val string) error { return b.WriteVal(val) }

// WriteInt32 writes an int32 to the buffer
func (b *ByteBuffer) WriteInt32(val int32) error { return b.WriteVal(val) }

// WriteInt64 writes an int64 to the buffer
func (b *ByteBuffer) WriteInt64(val int64) error { return b.WriteVal(val) }

// WriteUint32 writes an uint32 to the buffer
func (b *ByteBuffer) WriteUint32(val uint32) error { return b.WriteVal(val) }

// WriteUint64 writes an uint64 to the buffer
func (b *ByteBuffer) WriteUint64(val uint64) error { return b.WriteVal(val) }

// WriteFloat32 writes an float32 to the buffer
func (b *ByteBuffer) WriteFloat32(val float32) error { return b.WriteVal(val) }

// WriteFloat64 writes an float64 to the buffer
func (b *ByteBuffer) WriteFloat64(val float64) error { return b.WriteVal(val) }
