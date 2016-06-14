// this file implements a MMV file writer for speed

package speed

// Writer defines the interface of a MMV file writer's properties
type Writer interface {
	AddInstanceDomain(*InstanceDomain) error                                                     // adds a InstanceDomain object to the writer
	AddInstanceDomainByName(name string, instances []string) (*InstanceDomain, error)            // adds a InstanceDomain object after constructing it using passed name and instances
	AddMetric(*Metric) error                                                                     // adds a Metric object to the writer
	AddMetricByName(name string, s MetricSemantics, t MetricType, u MetricUnit) (*Metric, error) // adds a Metric object after parsing the name for Instances and InstanceDomains
	UpdateMetricByName(name string, val interface{}) error                                       // updates a Metric object by looking it up by name and updating its value
	Write()                                                                                      // writes an mmv file
}
