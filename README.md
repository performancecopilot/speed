![Speed](images/speed.png)

Golang implementation of the Performance Co-Pilot (PCP) instrumentation API

> A **Google Summer of Code 2016** project!

[![Build Status](https://travis-ci.org/performancecopilot/speed.svg?branch=master)](https://travis-ci.org/performancecopilot/speed) [![Coverage Status](https://coveralls.io/repos/github/performancecopilot/speed/badge.svg?branch=master)](https://coveralls.io/github/performancecopilot/speed?branch=master) [![GoDoc](https://godoc.org/github.com/performancecopilot/speed?status.svg)](https://godoc.org/github.com/performancecopilot/speed) [![Go Report Card](https://goreportcard.com/badge/github.com/performancecopilot/speed)](https://goreportcard.com/report/github.com/performancecopilot/speed)

## Install

### Prerequisites

#### PCP

Install Performance Co-Pilot on your local machine, either using prebuilt archives or by getting and building the source code. For detailed instructions, read the [PCP Install.md](https://github.com/performancecopilot/pcp/blob/master/INSTALL.md).

#### Go

Set up a go environment on your computer. For more information about these steps, please read [how to write go code](https://golang.org/doc/code.html), or [watch the video](https://www.youtube.com/watch?v=XCsL89YtqCs)

- download and install go 1.6 or above from [https://golang.org/dl](https://golang.org/dl)

- set up `$GOPATH` to the root folder where you want to keep your go code

- add `$GOPATH/bin` to your `$PATH` by adding `export PATH=$GOPATH/bin:$PATH` to your shell configuration file, such as to your `.bashrc`, if using a Bourne shell variant.

### Getting the library

First download the package

```sh
go get github.com/performancecopilot/speed
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
