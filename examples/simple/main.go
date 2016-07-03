package main

import (
	"fmt"
	"os"

	"github.com/performancecopilot/speed"
)

func main() {
	metric, err := speed.NewPCPMetric(
		42,
		"simple.counter",
		nil,
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"A Simple Metric",
		"This is a simple counter metric to demonstrate the speed API",
	)
	if err != nil {
		panic(err)
	}

	writer, err := speed.NewPCPWriter("simple", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	writer.Register(metric)

	writer.Start()

	fmt.Println("The metric is currently mapped as mmv.simple.simple.counter, to stop the mapping, press enter")
	os.Stdin.Read(make([]byte, 1))

	writer.Stop()
}
