package main

import "github.com/performancecopilot/speed"

func main() {
	metric, err := speed.NewPCPMetric(
		42,
		"simple.counter",
		nil,
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"",
		"",
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

	writer.Stop()
}
