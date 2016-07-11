package speed

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
)

// Registry defines a valid set of instance domains and metrics
type Registry interface {
	// checks if an instance domain of the passed name is already present or not
	HasInstanceDomain(name string) bool

	// checks if an metric of the passed name is already present or not
	HasMetric(name string) bool

	// returns the number of Metrics in the current registry
	MetricCount() int

	// returns the number of Values in the current registry
	ValuesCount() int

	// returns the number of Instance Domains in the current registry
	InstanceDomainCount() int

	// returns the number of instances across all instance domains in the current registry
	InstanceCount() int

	// returns the number of non null strings initialized in the current registry
	StringCount() int

	// adds a InstanceDomain object to the writer
	AddInstanceDomain(InstanceDomain) error

	// adds a InstanceDomain object after constructing it using passed name and instances
	AddInstanceDomainByName(name string, instances []string) (InstanceDomain, error)

	// adds a Metric object to the writer
	AddMetric(Metric) error

	// adds a Metric object after parsing the passed string for Instances and InstanceDomains
	AddMetricByString(name string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error)
}

// PCPRegistry implements a registry for PCP as the client
type PCPRegistry struct {
	instanceDomains map[string]*PCPInstanceDomain // a cache for instanceDomains
	metrics         map[string]PCPMetric          // a cache for metrics

	// locks
	indomlock   sync.RWMutex
	metricslock sync.RWMutex

	// offsets
	instanceoffset int
	indomoffset    int
	metricsoffset  int
	valuesoffset   int
	stringsoffset  int

	// counts
	instanceCount int
	valueCount    int
	stringcount   int

	mapped bool
}

// NewPCPRegistry creates a new PCPRegistry object
func NewPCPRegistry() *PCPRegistry {
	return &PCPRegistry{
		instanceDomains: make(map[string]*PCPInstanceDomain),
		metrics:         make(map[string]PCPMetric),
	}
}

// InstanceCount returns the number of instances across all indoms in the registry
func (r *PCPRegistry) InstanceCount() int {
	r.indomlock.RLock()
	defer r.indomlock.RUnlock()

	return r.instanceCount
}

// InstanceDomainCount returns the number of instance domains in the registry
func (r *PCPRegistry) InstanceDomainCount() int {
	r.indomlock.RLock()
	defer r.indomlock.RUnlock()

	return len(r.instanceDomains)
}

// MetricCount returns the number of metrics in the registry
func (r *PCPRegistry) MetricCount() int {
	r.metricslock.RLock()
	defer r.metricslock.RUnlock()

	return len(r.metrics)
}

// ValuesCount returns the number of values in the registry
func (r *PCPRegistry) ValuesCount() int { return r.valueCount }

// StringCount returns the number of strings in the registry
func (r *PCPRegistry) StringCount() int { return r.stringcount }

// HasInstanceDomain returns true if the registry already has an indom of the specified name
func (r *PCPRegistry) HasInstanceDomain(name string) bool {
	r.indomlock.RLock()
	defer r.indomlock.RUnlock()

	_, present := r.instanceDomains[name]
	return present
}

// HasMetric returns true if the registry already has a metric of the specified name
func (r *PCPRegistry) HasMetric(name string) bool {
	r.metricslock.RLock()
	defer r.metricslock.RUnlock()

	_, present := r.metrics[name]
	return present
}

// AddInstanceDomain will add a new instance domain to the current registry
func (r *PCPRegistry) AddInstanceDomain(indom InstanceDomain) error {
	if r.HasInstanceDomain(indom.Name()) {
		return errors.New("InstanceDomain is already defined for the current registry")
	}

	r.indomlock.Lock()
	defer r.indomlock.Unlock()

	if r.mapped {
		return errors.New("Cannot add an indom when a mapping is active")
	}

	r.instanceDomains[indom.Name()] = indom.(*PCPInstanceDomain)
	r.instanceCount += indom.InstanceCount()

	if indom.(*PCPInstanceDomain).shortDescription.val != "" {
		r.stringcount++
	}

	if indom.(*PCPInstanceDomain).longDescription.val != "" {
		r.stringcount++
	}

	return nil
}

// AddMetric will add a new metric to the current registry
func (r *PCPRegistry) AddMetric(m Metric) error {
	if r.HasMetric(m.Name()) {
		return errors.New("Metric is already defined for the current registry")
	}

	pcpm := m.(PCPMetric)

	// if it is an indom metric
	if pcpm.Indom() != nil && !r.HasInstanceDomain(pcpm.Indom().Name()) {
		return errors.New("Instance Domain is not defined for current registry")
	}

	if r.mapped {
		return errors.New("Cannot add a metric when a mapping is active")
	}

	r.metricslock.Lock()
	defer r.metricslock.Unlock()

	r.metrics[m.Name()] = pcpm

	if pcpm.Indom() != nil {
		r.valueCount += pcpm.Indom().InstanceCount()
	} else {
		r.valueCount++
	}

	if pcpm.ShortDescription().String() != "" {
		r.stringcount++
	}

	if pcpm.LongDescription().String() != "" {
		r.stringcount++
	}

	return nil
}

// AddInstanceDomainByName adds an instance domain using passed parameters
func (r *PCPRegistry) AddInstanceDomainByName(name string, instances []string) (InstanceDomain, error) {
	if r.HasInstanceDomain(name) {
		return nil, errors.New("The InstanceDomain already exists for this registry")
	}

	indom, err := NewPCPInstanceDomain(name, instances, "", "")
	if err != nil {
		return nil, err
	}

	err = r.AddInstanceDomain(indom)
	if err != nil {
		return nil, err
	}

	return indom, nil
}

const id = "[\\p{L}\\p{N}]+"

var instancesPattern = fmt.Sprintf("(%v)((,\\s?(%v))*)", id, id)
var pattern = fmt.Sprintf("\\A((%v)(\\.%v)*?)(\\[(%v)\\])?((\\.%v)*)\\z", id, id, instancesPattern, id)

var ireg, _ = regexp.Compile(id)
var reg, _ = regexp.Compile(pattern)

func parseString(s string) (metric string, indom string, instances []string, err error) {
	if !reg.MatchString(s) {
		return "", "", nil, errors.New("Invalid String")
	}

	matches := reg.FindStringSubmatch(s)
	n := len(matches)

	indom = matches[1]
	metric = indom + matches[n-2]

	iarr := matches[5]
	if iarr != "" {
		instances = ireg.FindAllString(iarr, -1)
	} else {
		instances = nil
		indom = ""
	}

	return
}

// AddMetricByString dynamically creates a PCPMetric
func (r *PCPRegistry) AddMetricByString(str string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error) {
	metric, indom, instances, err := parseString(str)
	if err != nil {
		return nil, err
	}

	var m Metric

	if instances == nil {
		// singleton metric
		m, err = NewPCPSingletonMetric(val, metric, t, s, u, "", "")
		if err != nil {
			return nil, err
		}

		err = r.AddMetric(m)
		if err != nil {
			return nil, err
		}

		return m, nil
	}

	// instance metric
	mp, ok := val.(Instances)
	if !ok {
		return nil, errors.New("to define an instance metric, a Instances type is required")
	}

	id, err := r.AddInstanceDomainByName(indom, instances)
	if err != nil {
		return nil, err
	}

	m, err = NewPCPInstanceMetric(mp, metric, id.(*PCPInstanceDomain), t, s, u, "", "")
	if err != nil {
		return nil, err
	}

	err = r.AddMetric(m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
