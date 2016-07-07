package speed

// pcpInstance wraps a PCP compatible Instance
type pcpInstance struct {
	name   string
	id     uint32
	offset int
}

// newpcpInstance generates a new Instance type based on the passed parameters
// the id is passed explicitly as it is assumed that this will be constructed
// after initializing the InstanceDomain
// this is not a part of the public API as this is not supposed to be used directly,
// but instead added using the AddInstance method of InstanceDomain
func newpcpInstance(name string) *pcpInstance {
	return &pcpInstance{
		name, getHash(name, 0), 0,
	}
}

func (i *pcpInstance) String() string {
	return "Instance: " + i.name
}
