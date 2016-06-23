// Package bytebuffer implements a java like bytebuffer for go
//
// initially tried to use bytes.Buffer but the main restriction with that is that
// it does not allow the freedom to move around in the buffer. Further, it always
// writes at the end of the buffer
//
// another attempt was to maintain a position identifier that was always passed
// to any function that wrote anything and the function *had* to return the
// next writable location, which resulted in calls like
//
//	pos = writeString(buffer, pos, "mmv")
//
// which became unmaintainable after a while, and along with all the side
// maintainance looked extremely ugly
//
// this (tries) to implement a minimal buffer wrapper that gives the freedom
// to move around and write anywhere you want
package bytebuffer

import (
	"errors"
	"fmt"
)

type Buffer interface {
	Pos() int
	SetPos(int)
	Len() int
	Buffer() []byte
	Write([]byte)
	WriteString(string)
	WriteVal(interface{})
}

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

func (b *ByteBuffer) ByteBuffer() []byte { return b.buffer }

func (b *ByteBuffer) Write(data []byte) {
	l := len(data)

	if b.Pos()+l > b.Len() {
		// TODO: make a better error message
		panic(errors.New("Overflow"))
	}

	for i := 0; i < l; i++ {
		data[b.pos+i] = data[i]
	}

	b.pos += l
}

func (b *ByteBuffer) WriteString(s string) {
	b.Write([]byte(s))
}

func (b *ByteBuffer) WriteVal(val interface{}) {
	b.WriteString(fmt.Sprintf("%v", val))
}
