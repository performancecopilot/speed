package main

import (
	"flag"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	w, err := speed.NewPCPClient("stringtest", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	names := []string{
		"Batman",
		"Robin",
		"Nightwing",
		"Batgirl",
		"Red Robin",
		"Red Hood",
	}

	m, err := w.RegisterString("bat.names", names[0], speed.InstantSemantics, speed.StringType, speed.OneUnit)
	if err != nil {
		panic(err)
	}

	w.MustStart()
	defer w.MustStop()

	metric := m.(speed.SingletonMetric)
	for i := 0; i < *timelimit; i++ {
		metric.MustSet(names[i%len(names)])
		time.Sleep(time.Second)
	}
}
