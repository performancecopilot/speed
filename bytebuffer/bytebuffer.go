// Package bytebuffer implements a java like bytebuffer for go
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

type Buffer struct {
	pos    int
	buffer []byte
}

func NewBuffer(n int) *Buffer {
	return &Buffer{
		pos:    0,
		buffer: make([]byte, n),
	}
}

func NewBufferSlice(buffer []byte) *Buffer {
	return &Buffer{
		pos:    0,
		buffer: buffer,
	}
}

func (b *Buffer) Pos() int { return b.pos }

func (b *Buffer) SetPos(position int) {
	if position < 0 || position >= len(b.buffer) {
		// TODO: make a better error message
		panic(errors.New("Out of Range"))
	}

	b.pos = position
}

func (b *Buffer) Len() int { return len(b.buffer) }

func (b *Buffer) Buffer() []byte { return b.buffer }

func (b *Buffer) Write(data []byte) {
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

func (b *Buffer) WriteString(s string) {
	b.Write([]byte(s))
}

func (b *Buffer) WriteVal(val interface{}) {
	b.WriteString(fmt.Sprintf("%v", val))
}
