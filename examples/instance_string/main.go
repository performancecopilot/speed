package main

import (
	"flag"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	c, err := speed.NewPCPClient("strings", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	m, err := c.RegisterString("language[go, javascript, php].users", speed.Instances{
		"go":         1,
		"javascript": 100,
		"php":        10,
	}, speed.CounterSemantics, speed.Uint64Type, speed.OneUnit)
	if err != nil {
		panic(err)
	}

	c.MustStart()
	defer c.MustStop()

	metric := m.(speed.InstanceMetric)
	for i := 0; i < *timelimit; i++ {
		v, _ := metric.ValInstance("go")
		metric.MustSetInstance("go", v.(uint64)*2)

		v, _ = metric.ValInstance("javascript")
		metric.MustSetInstance("javascript", v.(uint64)+10)

		v, _ = metric.ValInstance("php")
		metric.MustSetInstance("php", v.(uint64)+1)

		time.Sleep(time.Second)
	}
}
