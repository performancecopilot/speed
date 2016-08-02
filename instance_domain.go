package speed

import "errors"

// InstanceDomain defines the interface for an instance domain
type InstanceDomain interface {
	ID() uint32                   // unique identifier for the instance domain
	Name() string                 // name of the instance domain
	Description() string          // description for the instance domain
	HasInstance(name string) bool // checks if an instance is in the indom
	InstanceCount() int           // returns the number of instances in the indom
}

// PCPInstanceDomainBitLength is the maximum bit length of a PCP Instance Domain
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPInstanceDomainBitLength = 22

// PCPInstanceDomain wraps a PCP compatible instance domain
type PCPInstanceDomain struct {
	id                                uint32
	name                              string
	instances                         map[string]*pcpInstance
	offset                            int
	instanceOffset                    int
	shortDescription, longDescription *pcpString
}

// NewPCPInstanceDomain creates a new instance domain or returns an already created one for the passed name
// NOTE: this is different from parfait's idea of generating ids for InstanceDomains
// We simply generate a unique 32 bit hash for an instance domain name, and if it has not
// already been created, we create it, otherwise we return the already created version
func NewPCPInstanceDomain(name string, instances []string, desc ...string) (*PCPInstanceDomain, error) {
	if name == "" {
		return nil, errors.New("Instance Domain name cannot be empty")
	}

	if len(desc) > 2 {
		return nil, errors.New("Only 2 description strings allowed to define an instance domain")
	}

	shortDescription, longDescription := "", ""

	if len(desc) > 0 {
		shortDescription = desc[0]
	}

	if len(desc) > 1 {
		longDescription = desc[1]
	}

	imap := make(map[string]*pcpInstance)

	for _, instance := range instances {
		imap[instance] = newpcpInstance(instance)
	}

	return &PCPInstanceDomain{
		id:               hash(name, PCPInstanceDomainBitLength),
		name:             name,
		instances:        imap,
		shortDescription: newpcpString(shortDescription),
		longDescription:  newpcpString(longDescription),
	}, nil
}

// HasInstance returns true if an instance of the specified name is in the Indom
func (indom *PCPInstanceDomain) HasInstance(name string) bool {
	_, present := indom.instances[name]
	return present
}

// ID returns the id for PCPInstanceDomain
func (indom *PCPInstanceDomain) ID() uint32 { return indom.id }

// Name returns the name for PCPInstanceDomain
func (indom *PCPInstanceDomain) Name() string { return indom.name }

// InstanceCount returns the number of instances in the current instance domain
func (indom *PCPInstanceDomain) InstanceCount() int {
	return len(indom.instances)
}

// MatchInstances returns true if the passed InstanceDomain
// has exactly the same instances as the passed array
func (indom *PCPInstanceDomain) MatchInstances(ins []string) bool {
	if len(ins) != len(indom.instances) {
		return false
	}

	for _, i := range ins {
		if _, ok := indom.instances[i]; !ok {
			return false
		}
	}

	return true
}

// Description returns the description for PCPInstanceDomain
func (indom *PCPInstanceDomain) Description() string {
	s, l := indom.shortDescription.val, indom.longDescription.val
	if l != "" {
		return s + "\n\n" + l
	}
	return s
}

func (indom *PCPInstanceDomain) String() string {
	s := "InstanceDomain: " + indom.name
	if len(indom.instances) > 0 {
		s += "["
		for _, i := range indom.instances {
			s += i.name + ","
		}
		s += "]"
	}
	return s
}
