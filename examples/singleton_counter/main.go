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
	)
	if err != nil {
		panic(err)
	}

	client, err := speed.NewPCPClient("singletoncounter")
	if err != nil {
		panic(err)
	}

	err = client.Register(metric)
	if err != nil {
		panic(err)
	}

	client.MustStart()
	defer client.MustStop()

	fmt.Println("The metric should be visible as mmv.singletoncounter.counter")
	for i := 0; i < *timelimit; i++ {
		v := metric.Val().(int32)
		v++
		metric.MustSet(v)
		time.Sleep(time.Second)
	}
}
