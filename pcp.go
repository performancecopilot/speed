package speed

// PCPValue defines a set of properties that are supposed to be present in
// a PCP compatible value, but don't necessarily need to be implemented
// generally, by *all* values defined in speed
type PCPValue interface {
	Offset() int // returns the memory offset a value is supposed to be stored at
}

// PCPString defines a string that also has a memory offset containing
// the location where it will be written
type PCPString struct {
	val    string
	offset int
}

func NewPCPString(s string) *PCPString {
	return &PCPString{s, 0}
}

func (s *PCPString) Offset() int { return s.offset }

func (s *PCPString) setOffset(offset int) { s.offset = offset }

func (s *PCPString) String() string { return s.val }
