package speed

import (
	"errors"
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
	AddMetricByString(name string, initialval interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error)

	// updates the passed metric with the passed value
	UpdateMetric(m Metric, val interface{}) error

	// updates a Metric object by looking it up by name and updating its value
	UpdateMetricByName(name string, val interface{}) error
}

// PCPRegistry implements a registry for PCP as the client
type PCPRegistry struct {
	instanceDomains map[uint32]*PCPInstanceDomain // a cache for instanceDomains
	metrics         map[uint32]*PCPMetric         // a cache for metrics
	instanceCount   int
	indomlock       sync.RWMutex
	metricslock     sync.RWMutex
	instanceoffset  int
	indomoffset     int
	metricsoffset   int
	valuesoffset    int
	stringsoffset   int
	stringcount     int // the number of strings to be written
	mapped          bool
}

// NewPCPRegistry creates a new PCPRegistry object
func NewPCPRegistry() *PCPRegistry {
	return &PCPRegistry{
		instanceDomains: make(map[uint32]*PCPInstanceDomain),
		metrics:         make(map[uint32]*PCPMetric),
	}
}

// HasInstanceDomain checks if an instance domain of specified name already exists
// in registry or not
func (r *PCPRegistry) HasInstanceDomain(name string) bool {
	r.indomlock.RLock()
	defer r.indomlock.RUnlock()

	id := getHash(name, PCPInstanceDomainBitLength)
	_, present := r.instanceDomains[id]
	return present
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

func (r *PCPRegistry) StringCount() int { return r.stringcount }

// HasMetric checks if a metric of specified name already exists
// in registry or not
func (r *PCPRegistry) HasMetric(name string) bool {
	r.metricslock.RLock()
	defer r.metricslock.RUnlock()

	id := getHash(name, PCPMetricItemBitLength)
	_, present := r.metrics[id]
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

	r.instanceDomains[indom.ID()] = indom.(*PCPInstanceDomain)
	r.instanceCount += indom.InstanceCount()

	if indom.(*PCPInstanceDomain).shortHelpText.val != "" {
		r.stringcount++
	}

	if indom.(*PCPInstanceDomain).longHelpText.val != "" {
		r.stringcount++
	}

	return nil
}

// AddMetric will add a new metric to the current registry
func (r *PCPRegistry) AddMetric(m Metric) error {
	if r.HasMetric(m.Name()) {
		return errors.New("Metric is already defined for the current registry")
	}

	r.metricslock.Lock()
	defer r.metricslock.Unlock()

	if r.mapped {
		return errors.New("Cannot add a metric when a mapping is active")
	}

	pcpm := m.(*PCPMetric)

	r.metrics[m.ID()] = pcpm

	if pcpm.Indom() != nil && !r.HasInstanceDomain(pcpm.Indom().Name()) {
		r.AddInstanceDomain(pcpm.Indom())
	}

	if pcpm.desc.shortDescription.val != "" {
		r.stringcount++
	}

	if pcpm.desc.longDescription.val != "" {
		r.stringcount++
	}

	return nil
}

// AddInstanceDomainByName adds an instance domain using passed parameters
func (r *PCPRegistry) AddInstanceDomainByName(name string, instances []string) (InstanceDomain, error) {
	if r.HasInstanceDomain(name) {
		return nil, errors.New("The InstanceDomain already exists for this registry")
	}

	indom, err := NewPCPInstanceDomain(name, "", "")
	if err != nil {
		return nil, err
	}

	for _, i := range instances {
		indom.AddInstance(i)
	}

	r.AddInstanceDomain(indom)
	return indom, nil
}

// IdentifierPat contains the pattern for a valid name identifier
const IdentifierPat = "[\\p{L}\\p{N}]+"

const p = "\\A((" + IdentifierPat + ")(\\." + IdentifierPat + ")*?)(\\[(" + IdentifierPat + ")\\])?((\\." + IdentifierPat + ")*)\\z"

// identifierRegex gets the *regexp.Regexp object representing a valid metric identifier
var identifierRegex, _ = regexp.Compile(p)

func parseString(name string) (iname string, indomname string, mname string, err error) {
	if !identifierRegex.MatchString(name) {
		return "", "", "", errors.New("I don't know this")
	}

	matches := identifierRegex.FindStringSubmatch(name)
	iname, indomname, mname, err = matches[5], matches[1], matches[1]+matches[6], nil
	return
}

// AddMetricByString provides parfait style metric creation
func (r *PCPRegistry) AddMetricByString(name string, initialval interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error) {
	iname, indomname, mname, err := parseString(name)
	if err != nil {
		return nil, err
	}

	indom, err := r.AddInstanceDomainByName(indomname, []string{iname})
	if err != nil {
		return nil, err
	}

	if r.HasMetric(mname) {
		return nil, errors.New("The Metric already exists for this registry")
	}

	m, err := NewPCPMetric(initialval, name, indom, t, s, u, "", "")
	if err != nil {
		return nil, err
	}

	r.AddMetric(m)

	return m, nil
}

// UpdateMetric updates the passed metric's value
func (r *PCPRegistry) UpdateMetric(m Metric, val interface{}) error {
	r.metricslock.Lock()
	defer r.metricslock.Unlock()

	if r.mapped {
		return errors.New("Cannot update metric when a mapping is active")
	}

	return m.Set(val)
}

// UpdateMetricByName safely updates the value of a metric
func (r *PCPRegistry) UpdateMetricByName(name string, val interface{}) error {
	if !r.HasMetric(name) {
		return errors.New("The Metric doesn't exist for this registry")
	}

	h := getHash(name, PCPMetricItemBitLength)
	m := r.metrics[h]

	return r.UpdateMetric(m, val)
}
