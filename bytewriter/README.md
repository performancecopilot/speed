# bytewriter [![GoDoc](https://godoc.org/github.com/performancecopilot/speed/bytewriter?status.svg)](https://godoc.org/github.com/performancecopilot/speed/bytewriter)

Package bytewriter implements writers that support concurrent writing within a fixed length block

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

then implemented a minimal buffer wrapper that gives the freedom
to move around and write anywhere you want, which worked if you only need
one position identifier, i.e. only one write operation happens at a time

this implements a writer that supports multiple concurrent writes within a fixed length block
