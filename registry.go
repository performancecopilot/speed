package speed

import "sync"

// Registry defines a valid set of instance domains and metrics
type Registry interface {
	AddInstanceDomain(*InstanceDomain) error                                                     // adds a InstanceDomain object to the writer
	AddInstanceDomainByName(name string, instances []string) (*InstanceDomain, error)            // adds a InstanceDomain object after constructing it using passed name and instances
	AddMetric(*Metric) error                                                                     // adds a Metric object to the writer
	AddMetricByName(name string, s MetricSemantics, t MetricType, u MetricUnit) (*Metric, error) // adds a Metric object after parsing the name for Instances and InstanceDomains
	UpdateMetricByName(name string, val interface{}) error                                       // updates a Metric object by looking it up by name and updating its value
}

// PCPRegistry implements a registry to write MMV files to PCP
type PCPRegistry struct {
	instanceDomains map[uint32]*InstanceDomain // a cache for instanceDomains
	metrics         map[uint32]*Metric         // a cache for metrics
	mu              sync.Mutex                 // mutex to synchronize access
}
