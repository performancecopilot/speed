package speed

import (
	"errors"
	"sync"
)

// Registry defines a valid set of instance domains and metrics
type Registry interface {
	HasInstanceDomain(name string) bool                                                           // checks if an instance domain of the passed name is already present or not
	HasMetric(name string) bool                                                                   // checks if an metric of the passed name is already present or not
	AddInstanceDomain(InstanceDomain) error                                                       // adds a InstanceDomain object to the writer
	AddInstanceDomainByName(name string, instances []string) (InstanceDomain, error)              // adds a InstanceDomain object after constructing it using passed name and instances
	AddMetric(Metric) error                                                                       // adds a Metric object to the writer
	AddMetricByString(name string, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error) // adds a Metric object after parsing the passed string for Instances and InstanceDomains
	UpdateMetricByName(name string, val interface{}) error                                        // updates a Metric object by looking it up by name and updating its value
}

// PCPRegistry implements a registry for PCP as the client
type PCPRegistry struct {
	instanceDomains map[uint32]InstanceDomain // a cache for instanceDomains
	metrics         map[uint32]Metric         // a cache for metrics
	mu              sync.Mutex                // mutex to synchronize access
}

// NewPCPRegistry creates a new PCPRegistry object
func NewPCPRegistry() *PCPRegistry {
	return &PCPRegistry{
		instanceDomains: make(map[uint32]InstanceDomain),
		metrics:         make(map[uint32]Metric),
	}
}

// HasInstanceDomain checks if an instance domain of specified name already exists
// in registry or not
func (r *PCPRegistry) HasInstanceDomain(name string) bool {
	id := getHash(name, PCPInstanceDomainBitLength)
	_, present := r.instanceDomains[id]
	return present
}

// AddInstanceDomain will add a new instance domain to the current registry
func (r *PCPRegistry) AddInstanceDomain(indom InstanceDomain) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.HasInstanceDomain(indom.Name()) {
		return errors.New("InstanceDomain is already defined for the current registry")
	}

	r.instanceDomains[indom.ID()] = indom
	return nil
}

// HasMetric checks if a metric of specified name already exists
// in registry or not
func (r *PCPRegistry) HasMetric(name string) bool {
	id := getHash(name, PCPMetricBitLength)
	_, present := r.metrics[id]
	return present
}

// AddMetric will add a new metric to the current registry
func (r *PCPRegistry) AddMetric(m Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.HasMetric(m.Name()) {
		return errors.New("Metric is already defined for the current registry")
	}

	r.metrics[m.ID()] = m
	return nil
}
