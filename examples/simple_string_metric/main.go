package main

import (
	"flag"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	c, err := speed.NewPCPClient("stringtest")
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

	m, err := c.RegisterString("bat.names", names[0], speed.StringType, speed.InstantSemantics, speed.OneUnit)
	if err != nil {
		panic(err)
	}

	c.MustStart()
	defer c.MustStop()

	metric := m.(speed.SingletonMetric)
	for i := 0; i < *timelimit; i++ {
		metric.MustSet(names[i%len(names)])
		time.Sleep(time.Second)
	}
}
