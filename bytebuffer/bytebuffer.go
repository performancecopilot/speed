package bytebuffer

import "errors"

type ByteBuffer struct {
	pos    int
	buffer []byte
}

func NewByteBuffer(n int) *ByteBuffer {
	return &ByteBuffer{
		pos:    0,
		buffer: make([]byte, n),
	}
}

func NewByteBufferSlice(buffer []byte) *ByteBuffer {
	return &ByteBuffer{
		pos:    0,
		buffer: buffer,
	}
}

func (b *ByteBuffer) Pos() int { return b.pos }

func (b *ByteBuffer) SetPos(position int) {
	if position < 0 || position >= len(b.buffer) {
		// TODO: make a better error message
		panic(errors.New("Out of Range"))
	}

	b.pos = position
}

func (b *ByteBuffer) Len() int { return len(b.buffer) }

func (b *ByteBuffer) Buffer() []byte { return b.buffer }

func (b *ByteBuffer) Write(data []byte) {
	l := len(data)

	if b.Pos()+l > b.Len() {
		// TODO: make a better error message
		panic(errors.New("Overflow"))
	}

	for i := 0; i < l; i++ {
		b.buffer[b.pos+i] = data[i]
	}

	b.pos += l
}

func (b *ByteBuffer) WriteString(s string) {
	b.Write([]byte(s))
}

func (b *ByteBuffer) WriteUint32(val uint32) {
	b.Write([]byte{
		byte(val & 0xFF),
		byte((val >> 8) & 0xFF),
		byte((val >> 16) & 0xFF),
		byte(val >> 24),
	})
}

func (b *ByteBuffer) WriteUint64(val uint64) {
	b.WriteUint32(uint32(val & 0xFFFF))
	b.WriteUint32(uint32(val >> 32))
}

func (b *ByteBuffer) WriteInt32(val int32) {
	b.WriteUint32(uint32(val))
}

func (b *ByteBuffer) WriteInt64(val int64) {
	b.WriteUint64(uint64(val))
}

func (b *ByteBuffer) WriteInt(val int) {
	b.WriteUint32(uint32(val))
}
