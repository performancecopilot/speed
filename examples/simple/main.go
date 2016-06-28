package main

import "github.com/performancecopilot/speed"

func main() {
	metric := speed.NewPCPMetric(
		int32(42),
		"simple.counter",
		nil,
		speed.Int32Type,
		speed.CounterSemantics,
		speed.MetricUnit(speed.OneUnit),
		"",
		"",
	)

	writer, err := speed.NewPCPWriter("simple", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	writer.Register(metric)

	writer.Start()

	writer.Stop()
}
