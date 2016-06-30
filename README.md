# speed [![GoDoc](https://godoc.org/github.com/performancecopilot/speed?status.svg)](https://godoc.org/github.com/performancecopilot/speed) [![Go Report Card](https://goreportcard.com/badge/github.com/performancecopilot/speed)](https://goreportcard.com/report/github.com/performancecopilot/speed)

> Google Summer of Code 2016

An implementation of the PCP instrumentation API in golang

## Install

### Prerequisites

Set up a go environment on your computer. For more information about these steps, please read [how to write go code](https://golang.org/doc/code.html), or [watch the video](https://www.youtube.com/watch?v=XCsL89YtqCs)

- download and install go from [https://golang.org/dl](https://golang.org/dl)

- set up `$GOPATH` to the root folder where you want to keep your go code

- add `$GOPATH/bin` to your `$PATH` by adding `export PATH=$GOPATH/bin:$PATH` to your shell configuration file, preferably to your `.bashrc`

### Getting the library

First download the package without installing it

```sh
go get -d github.com/performancecopilot/speed
```

then go to the source and run make

```sh
cd $GOPATH/src/github.com/performancecopilot/speed
make
```

to install dependencies and build the package

### Getting the examples

All examples are executable go programs. Simply doing

```sh
go get github.com/performancecopilot/speed/examples/<example name>
```

will get the example and add an executable to `$GOPATH/bin`. If it is on your path, simply doing

```sh
<example name>
```

will run the binary, running the example
