// A golang implementation of the acme factory example from mmv.py in PCP core
// (https://github.com/performancecopilot/pcp/blob/master/src/python/pcp/mmv.py#L21-L70)
//
// TODO:
// - modify values section when support for multiple valued metrics are added to speed
// - add update section when memory mapped buffer with pmdammv visibility is done
package main

import "github.com/performancecopilot/speed"

func main() {
	instances := []string{"Anvils", "Rockets", "Giant Rubber Bands"}

	indom, err := speed.NewPCPInstanceDomain(
		"Acme Products",
		"Acme products",
		"Most popular products produced by the Acme Corporation",
	)
	if err != nil {
		panic(err)
	}

	err = indom.AddInstances(instances)
	if err != nil {
		panic(err)
	}

	countmetric, err := speed.NewPCPMetric(
		0,
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

	timemetric, err := speed.NewPCPMetric(
		0,
		"products.time",
		indom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.MicrosecondUnit,
		"Machine time spent producing Acme products",
		"",
	)
	if err != nil {
		panic(err)
	}

	writer, err := speed.NewPCPWriter("acme", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	writer.Register(countmetric)
	writer.Register(timemetric)

	writer.Start()

	// TODO: add update code after finishing memory mapped buffer implementation

	writer.Stop()
}
