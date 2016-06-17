package speed

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
