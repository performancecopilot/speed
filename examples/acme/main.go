// A golang implementation of the acme factory example from mmv.py in PCP core
// (https://github.com/performancecopilot/pcp/blob/master/src/python/pcp/mmv.py#L21-L70)
package main

import (
	"time"

	"github.com/performancecopilot/speed"
)

func main() {
	instances := []string{"Anvils", "Rockets", "Giant_Rubber_Bands"}

	indom, err := speed.NewPCPInstanceDomain(
		"Acme Products",
		instances,
		"Acme products",
		"Most popular products produced by the Acme Corporation",
	)
	if err != nil {
		panic(err)
	}

	countmetric, err := speed.NewPCPInstanceMetric(
		speed.Instances{
			"Anvils":             0,
			"Rockets":            0,
			"Giant_Rubber_Bands": 0,
		},
		"products.count",
		indom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"Acme factory product throughput",
		`Monotonic increasing counter of products produced in the Acme Corporation
		 factory since starting the Acme production application.  Quality guaranteed.`,
	)
	if err != nil {
		panic(err)
	}

	timemetric, err := speed.NewPCPInstanceMetric(
		speed.Instances{
			"Anvils":             0,
			"Rockets":            0,
			"Giant_Rubber_Bands": 0,
		},
		"products.time",
		indom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.MicrosecondUnit,
		"Machine time spent producing Acme products",
	)
	if err != nil {
		panic(err)
	}

	client, err := speed.NewPCPClient("acme")
	if err != nil {
		panic(err)
	}

	client.MustRegisterIndom(indom)
	client.MustRegister(countmetric)
	client.MustRegister(timemetric)

	client.MustStart()
	defer client.MustStop()

	time.Sleep(time.Second * 5)
	err = countmetric.SetInstance(42, "Anvils")
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 5)
}
