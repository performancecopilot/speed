package speed

// pcpString defines a string that also has a memory offset containing
// the location where it will be written
type pcpString struct {
	val    string
	offset int
}

// newpcpString creates a new instance of a pcpString from a raw string
func newpcpString(s string) *pcpString {
	return &pcpString{s, 0}
}

func (s *pcpString) String() string { return s.val }
