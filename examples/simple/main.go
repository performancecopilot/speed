package main

import (
	"fmt"
	"os"

	"github.com/performancecopilot/speed"
)

func main() {
	metric, err := speed.NewPCPSingletonMetric(
		42,
		"simple.counter",
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"A Simple Metric",
		"This is a simple counter metric to demonstrate the speed API",
	)
	if err != nil {
		panic(err)
	}

	writer, err := speed.NewPCPClient("simple", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	writer.MustRegister(metric)

	writer.MustStart()
	defer writer.MustStop()

	fmt.Println("The metric is currently mapped as mmv.simple.simple.counter, to stop the mapping, press enter")
	_, _ = os.Stdin.Read(make([]byte, 1))
}
