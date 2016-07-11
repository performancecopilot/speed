package speed

import (
	"hash/fnv"

	"github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func initLogging() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.InfoLevel
}

// init maintains a central location of all things that happen when the package is initialized
// instead of everything being scattered in multiple source files
func init() {
	initLogging()

	err := initConfig()
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": "config",
			"error":  err,
		}).Error("error initializing config. maybe PCP isn't installed properly")
	}
}

// generate a unique hash for a string of the specified bit length
// NOTE: make sure this is as fast as possible
//
// see: http://programmers.stackexchange.com/a/145633
func hash(s string, b uint32) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	val := h.Sum32()
	if b == 0 {
		return val
	}
	return val & ((1 << b) - 1)
}
