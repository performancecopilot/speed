package speed

// PCPValue defines a set of properties that are supposed to be present in
// a PCP compatible value, but don't necessarily need to be implemented
// generally, by *all* values defined in speed
type PCPValue interface {
	Offset() int   // returns the memory offset a value is supposed to be stored at
	SetOffset(int) // sets the memory offset
}
