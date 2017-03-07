![Speed](images/speed.png)

Golang implementation of the Performance Co-Pilot (PCP) instrumentation API

> [A **Google Summer of Code 2016** project!](https://summerofcode.withgoogle.com/archive/2016/projects/5093214348378112/)

[![Build Status](https://travis-ci.org/performancecopilot/speed.svg?branch=master)](https://travis-ci.org/performancecopilot/speed) [![Build status](https://ci.appveyor.com/api/projects/status/gins77i4ej1o5xef?svg=true)](https://ci.appveyor.com/project/suyash/speed) [![Coverage Status](https://coveralls.io/repos/github/performancecopilot/speed/badge.svg?branch=master)](https://coveralls.io/github/performancecopilot/speed?branch=master) [![GoDoc](https://godoc.org/github.com/performancecopilot/speed?status.svg)](https://godoc.org/github.com/performancecopilot/speed) [![Go Report Card](https://goreportcard.com/badge/github.com/performancecopilot/speed)](https://goreportcard.com/report/github.com/performancecopilot/speed)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Install](#install)
  - [Prerequisites](#prerequisites)
    - [PCP](#pcp)
    - [Go](#go)
    - [[Optional] [Vector](http://vectoross.io/)](#optional-vectorhttpvectorossio)
  - [Getting the library](#getting-the-library)
  - [Getting the examples](#getting-the-examples)
- [Walkthrough](#walkthrough)
  - [SingletonMetric](#singletonmetric)
  - [InstanceMetric](#instancemetric)
  - [Counter](#counter)
  - [CounterVector](#countervector)
  - [Gauge](#gauge)
  - [GaugeVector](#gaugevector)
  - [Timer](#timer)
  - [Histogram](#histogram)
- [Visualization through Vector](#visualization-through-vector)
- [Go Kit](#go-kit)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Install

### Prerequisites

#### [PCP](http://pcp.io)

Install Performance Co-Pilot on your local machine, either using prebuilt archives or by getting and building the source code. For detailed instructions, read the [page from vector documentation](http://vectoross.io/docs/installing-performance-co-pilot). For building from source on ubuntu 14.04, a simplified list of steps is [here](https://gist.github.com/suyash/0def9b33890d4a99ca9dd96724e1ac84)

#### [Go](https://golang.org)

Set up a go environment on your computer. For more information about these steps, please read [how to write go code](https://golang.org/doc/code.html), or [watch the video](https://www.youtube.com/watch?v=XCsL89YtqCs)

- download and install go 1.6 or above from [https://golang.org/dl](https://golang.org/dl)

- set up `$GOPATH` to the root folder where you want to keep your go code

- add `$GOPATH/bin` to your `$PATH` by adding `export PATH=$GOPATH/bin:$PATH` to your shell configuration file, such as to your `.bashrc`, if using a Bourne shell variant.

#### [Optional] [Vector](http://vectoross.io/)

Vector is an on premise visualization tool for metrics exposed using Performance Co-Pilot. For more read the [official documentation](http://vectoross.io/docs/) and play around with the [online demo](http://vectoross.io/demo)

### Getting the library

```sh
go get github.com/performancecopilot/speed
```

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

## Walkthrough

There are 3 main components defined in the library, a [__Client__](https://godoc.org/github.com/performancecopilot/speed#Client), a [__Registry__](https://godoc.org/github.com/performancecopilot/speed#Registry) and a [__Metric__](https://godoc.org/github.com/performancecopilot/speed#Metric). A client is created using an application name, and the same name is used to create a memory mapped file in `PCP_TMP_DIR`. Each client contains a registry of metrics that it holds, and will publish on being activated. It also has a `SetFlag` method allowing you to set a mmv flag while a mapping is not active, to one of three values, [`NoPrefixFlag`, `ProcessFlag` and `SentinelFlag`](https://godoc.org/github.com/performancecopilot/speed#MMVFlag). The ProcessFlag is the default and reports metrics prefixed with the application name (i.e. like `mmv.app_name.metric.name`). Setting it to `NoPrefixFlag` will report metrics without being prefixed with the application name (i.e. like `mmv.metric.name`) which can lead to namespace collisions, so be sure of what you're doing.

A client can register metrics to report through 2 interfaces, the first is the `Register` method, that takes a raw metric object. The other is using `RegisterString`, that can take a string with metrics and instances to register similar to the interface in parfait, along with type, semantics and unit, in that order. A client can be activated by calling the `Start` method, deactivated by the `Stop` method. While a client is active, no new metrics can be registered but it is possible to stop existing client for metric registration.

Each client contains an instance of the `Registry` interface, which can give different information like the number of registered metrics and instance domains. It also exports methods to register metrics and instance domains.

Finally, metrics are defined as implementations of different metric interfaces, but they all implement the `Metric` interface, the different metric types defined are

### [SingletonMetric](https://godoc.org/github.com/performancecopilot/speed#SingletonMetric)

This type defines a metric with no instance domain and only one value. It __requires__ type, semantics and unit for construction, and optionally takes a couple of description strings. A simple construction

```go
metric, err := speed.NewPCPSingletonMetric(
	42,                                                             // initial value
	"simple.counter",                                               // name
	speed.Int32Type,                                                // type
	speed.CounterSemantics,                                         // semantics
	speed.OneUnit,                                                  // unit
	"A Simple Metric",                                              // short description
	"This is a simple counter metric to demonstrate the speed API", // long description
)
```

A SingletonMetric supports a `Val` method that returns the metric value and a `Set(interface{})` method that sets the metric value.

### [InstanceMetric](https://godoc.org/github.com/performancecopilot/speed#InstanceMetric)

An `InstanceMetric` is a single metric object containing multiple values of the same type for multiple instances. It also __requires__ an instance domain along with type, semantics and unit for construction, and optionally takes a couple of description strings. A simple construction

```go
indom, err := speed.NewPCPInstanceDomain(
	"Acme Products",                                          // name
	[]string{"Anvils", "Rockets", "Giant_Rubber_Bands"},      // instances
	"Acme products",                                          // short description
	"Most popular products produced by the Acme Corporation", // long description
)

...

countmetric, err := speed.NewPCPInstanceMetric(
	speed.Instances{
		"Anvils":             0,
		"Rockets":            0,
		"Giant_Rubber_Bands": 0,
	},
	"products.count",
	indom,
	speed.Uint64Type,
	speed.CounterSemantics,
	speed.OneUnit,
	"Acme factory product throughput",
	`Monotonic increasing counter of products produced in the Acme Corporation
	factory since starting the Acme production application.  Quality guaranteed.`,
)
```

An instance metric supports a `ValInstance(string)` method that returns the value as well as a `SetInstance(interface{}, string)` that sets the value of a particular instance.

### [Counter](https://godoc.org/github.com/performancecopilot/speed#Counter)

A counter is simply a PCPSingletonMetric with `Int64Type`, `CounterSemantics` and `OneUnit`.
It can optionally take a short and a long description.

A simple example

```go
c, err := speed.NewPCPCounter(0, "a.simple.counter")
```

a counter supports `Set(int64)` to set a value, `Inc(int64)` to increment by a custom delta and `Up()` to increment by 1.

### [CounterVector](https://godoc.org/github.com/performancecopilot/speed#CounterVector)

A CounterVector is a PCPInstanceMetric , with `Int64Type`, `CounterSemantics` and `OneUnit` and an instance domain created and registered on initialization, with the name `metric_name.indom`.

A simple example

```go
c, err := speed.NewPCPCounterVector(
	map[string]uint64{
		"instance1": 0,
		"instance2": 1,
	}, "another.simple.counter"
)
```

It supports `Val(string)`, `Set(uint64, string)`, `Inc(uint64, string)` and `Up(string)` amongst other things.

### [Gauge](https://godoc.org/github.com/performancecopilot/speed#Gauge)

A Gauge is a simple SingletonMetric storing float64 values, i.e. a PCP Singleton Metric with `DoubleType`, `InstantSemantics` and `OneUnit`.

A simple example

```go
g, err := speed.NewPCPGauge(0, "a.sample.gauge")
```

supports `Val()`, `Set(float64)`, `Inc(float64)` and `Dec(float64)`

### [GaugeVector](https://godoc.org/github.com/performancecopilot/speed#GaugeVector)

A Gauge Vector is a PCP instance metric with `DoubleType`, `InstantSemantics` and `OneUnit` and an autogenerated instance domain. A simple example

```go
g, err := NewPCPGaugeVector(map[string]float64{
	"instance1": 1.2,
	"instance2": 2.4,
}, "met")
```

supports `Val(string)`, `Set(float64, string)`, `Inc(float64, string)` and `Dec(float64, string)`

### [Timer](https://godoc.org/github.com/performancecopilot/speed#Timer)

A timer stores the time elapsed for different operations. __It is not compatible with PCP's elapsed type metrics__. It takes a name and a `TimeUnit` for construction.

```go
timer, err := speed.NewPCPTimer("test", speed.NanosecondUnit)
```

calling `timer.Start()` signals the start of an operation

calling `timer.Stop()` signals end of an operation and will return the total elapsed time calculated by the metric so far.

### [Histogram](https://godoc.org/github.com/performancecopilot/speed#Histogram)

A histogram implements a PCP Instance Metric that reports the `mean`, `variance` and `standard_deviation` while using a histogram backed by [codahale's hdrhistogram implementation in golang](https://github.com/codahale/hdrhistogram). Other than these, it also returns a custom percentile and buckets for plotting graphs. It requires a low and a high value and the number of significant figures used at the time of construction.

```
m, err := speed.NewPCPHistogram("hist", 0, 1000, 5)
```

## Visualization through Vector

[Vector supports adding custom widgets for custom metrics](http://vectoross.io/docs/creating-widgets.html). However, that requires you to rebuild vector from scratch after adding the widget configuration. But if it is a one time thing, its worth it. For example here is the configuration I added to display the metric from the basic_histogram example

```
{
	name: 'mmv.histogram_test.hist',
	title: 'speed basic histogram example',
	directive: 'line-time-series',
	dataAttrName: 'data',
	dataModelType: CumulativeMetricDataModel,
	dataModelOptions: {
		name: 'mmv.histogram_test.hist'
	},
	size: {
		width: '50%',
		height: '250px'
	},
	enableVerticalResize: false,
	group: 'speed',
	attrs: {
		forcey: 100,
		percentage: false
	}
}
```

and the visualization I got

![screenshot from 2016-08-27 01 05 56](https://cloud.githubusercontent.com/assets/16324837/18172229/45b0442c-7082-11e6-9edd-ab6f91dc9f2e.png)

## [Go Kit](https://gokit.io)

Go kit provides [a wrapper package](https://godoc.org/github.com/go-kit/kit/metrics/pcp) over speed that can be used for building microservices that expose metrics using PCP.

For modified versions of the examples in go-kit that use pcp to report metrics, see [suyash/kit-pcp-examples](https://github.com/suyash/kit-pcp-examples)

## Projects using Speed

* https://github.com/lzap/pcp-mmvstatsd
