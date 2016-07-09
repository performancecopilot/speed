// this example showcases speeds metric inference from strings property
package main

import (
	"flag"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	w, err := speed.NewPCPWriter("strings", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	m, err := w.RegisterString("simple.counter", 10, speed.CounterSemantics, speed.Int32Type, speed.OneUnit)
	if err != nil {
		panic(err)
	}

	err = w.Start()
	if err != nil {
		panic(err)
	}

	for i := 0; i < *timelimit; i++ {
		val := m.(*speed.PCPSingletonMetric).Val().(int32)
		val++
		m.(*speed.PCPSingletonMetric).Set(val)
		time.Sleep(time.Second)
	}

	err = w.Stop()
	if err != nil {
		panic(err)
	}
}
