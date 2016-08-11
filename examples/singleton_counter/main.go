package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	metric, err := speed.NewPCPCounter(
		0,
		"counter",
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
		metric.Up()
		time.Sleep(time.Second)
	}
}
