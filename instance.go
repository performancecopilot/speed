package speed

// Instance wraps a PCP compatible Instance
type Instance struct {
	name  string
	id    uint32
	indom *InstanceDomain
}

// newInstance generates a new Instance type based on the passed parameters
// the id is passed explicitly as it is assumed that this will be constructed
// after initializing the InstanceDomain
// this is not a part of the public API as this is not supposed to be used directly,
// but instead added using the AddInstance method of InstanceDomain
func newInstance(id uint32, name string, indom *InstanceDomain) *Instance {
	return &Instance{
		name, id, indom,
	}
}

func (i *Instance) String() string {
	return "Instance: " + i.name
}
