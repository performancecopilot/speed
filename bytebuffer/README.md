# bytebuffer [![GoDoc](https://godoc.org/github.com/performancecopilot/speed/bytebuffer?status.svg)](https://godoc.org/github.com/performancecopilot/speed/bytebuffer)

Package bytebuffer implements a java like bytebuffer for go

initially tried to use bytes.Buffer but the main restriction with that is that
it does not allow the freedom to move around in the buffer. Further, it always
writes at the end of the buffer

another attempt was to maintain a position identifier that was always passed
to any function that wrote anything and the function *had* to return the
next writable location, which resulted in calls like

```go
pos = writeString(buffer, pos, "mmv")
```

which became unmaintainable after a while, and along with all the side
maintainance looked extremely ugly

this (tries) to implement a minimal buffer wrapper that gives the freedom
to move around and write anywhere you want
