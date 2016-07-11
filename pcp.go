package speed

// PCPString defines a string that also has a memory offset containing
// the location where it will be written
type PCPString struct {
	val    string
	offset int
}

// NewPCPString creates a new instance of a PCPString from a raw string
func NewPCPString(s string) *PCPString {
	return &PCPString{s, 0}
}

func (s *PCPString) String() string { return s.val }
