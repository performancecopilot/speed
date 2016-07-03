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

import "io"

// Buffer defines an abstraction for an object that allows writing of binary
// values anywhere within a fixed range
type Buffer interface {
	io.Writer
	Bytes() []byte
	Pos() int
	SetPos(int) error
	Len() int
	WriteVal(val interface{}) error
	WriteString(string) error
	WriteInt32(int32) error
	WriteInt64(int64) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error
	WriteFloat32(float32) error
	WriteFloat64(float64) error
}
