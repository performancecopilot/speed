package speed

import "errors"

// InstanceDomain defines the interface for an instance domain
type InstanceDomain interface {
	AddInstance(name string) error
	ID() uint32
	Name() string
	Description() string
}

// PCPInstanceDomain wraps a PCP compatible instance domain
type PCPInstanceDomain struct {
	id                          uint32
	name                        string
	instances                   map[uint32]*Instance // the instances for this InstanceDomain stored as a map
	shortHelpText, longHelpText string
}

// NewPCPInstanceDomain creates a new instance domain or returns an already created one for the passed name
// NOTE: this is different from parfait's idea of generating ids for InstanceDomains
// We simply generate a unique 32 bit hash for an instance domain name, and if it has not
// already been created, we create it, otherwise we return the already created version
func NewPCPInstanceDomain(name, shortDescription, longDescription string) *PCPInstanceDomain {
	return &PCPInstanceDomain{
		id:            getHash(name),
		name:          name,
		instances:     make(map[uint32]*Instance),
		shortHelpText: shortDescription,
		longHelpText:  longDescription,
	}
}

// AddInstance adds a new instance to the current PCPInstanceDomain
func (indom *PCPInstanceDomain) AddInstance(name string) error {
	h := getHash(name)

	_, present := indom.instances[h]
	if present {
		return errors.New("Instance with same name already created for the InstanceDomain")
	}

	ins := newInstance(h, name, indom)
	indom.instances[h] = ins

	return nil
}

// ID returns the id for PCPInstanceDomain
func (indom *PCPInstanceDomain) ID() uint32 { return indom.id }

// Name returns the name for PCPInstanceDomain
func (indom *PCPInstanceDomain) Name() string { return indom.name }

// Description returns the description for PCPInstanceDomain
func (indom *PCPInstanceDomain) Description() string {
	s, l := indom.shortHelpText, indom.longHelpText
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
