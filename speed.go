package speed

import "hash/fnv"

// init maintains a central location of all things that happen when the package is initialized
// instead of everything being scattered in multiple source files
func init() {
	err := initConfig()
	if err != nil {
		// TODO: do something better here, like silently exit
		panic(err)
	}
}

// generate a unique hash for a string of the specified bit length
// NOTE: make sure this is as fast as possible
//
// see: http://programmers.stackexchange.com/a/145633
func getHash(s string, b uint32) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	val := h.Sum32()
	if b == 0 {
		return val
	}
	return val & ((1 << b) - 1)
}
