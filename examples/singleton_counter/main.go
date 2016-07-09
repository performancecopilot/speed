package main

// TODO: update this example with PCPCounterMetric once that is implemented

import (
	"flag"
	"fmt"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	metric, err := speed.NewPCPSingletonMetric(
		0,
		"counter",
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"A Simple Metric",
		"This is a simple counter metric to demonstrate the speed API",
	)
	if err != nil {
		panic(err)
	}

	writer, err := speed.NewPCPWriter("singletoncounter", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	err = writer.Register(metric)
	if err != nil {
		panic(err)
	}

	err = writer.Start()
	if err != nil {
		panic(err)
	}

	fmt.Println("The metric should be visible as mmv.singletoncounter.counter")
	for i := 0; i < *timelimit; i++ {
		v := metric.Val().(int)
		v++
		time.Sleep(time.Second)
		metric.Set(v)
	}

	writer.Stop()
}
