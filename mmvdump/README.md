# mmvdump [![GoDoc](https://godoc.org/github.com/performancecopilot/speed/mmvdump?status.svg)](https://godoc.org/github.com/performancecopilot/speed/mmvdump)

Package mmvdump implements a go port of the C mmvdump utility included in PCP Core

https://github.com/performancecopilot/pcp/blob/master/src/pmdas/mmv/mmvdump.c

It has been written for maximum portability with the C equivalent, without having to use cgo or any other ninja stuff

the main difference is that the reader is separate from the cli with the reading primarily implemented in mmvdump.go while the cli is implemented in cmd/mmvdump

the cli application is completely go gettable and outputs the same things, in mostly the same way as the C cli app, to try it out,

```
go get github.com/performancecopilot/speed/mmvdump/cmd/mmvdump
```