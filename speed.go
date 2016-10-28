// Package speed implements a golang client for the Performance Co-Pilot
// instrumentation API.
//
// It is based on the C/Perl/Python API implemented in PCP core as well as the
// Java API implemented by `parfait`, a separate project.
//
// Some examples on using the API are implemented as executable go programs in the
// `examples` subdirectory.
package speed

import (
	"hash/fnv"

	"github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Version is the last tagged version of the package
const Version = "1.0.0"

var log = logrus.New()

var logging bool

func initLogging() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.InfoLevel
	logging = false
}

// EnableLogging logging enables logging for logrus if true is passed
// and disables it if false is passed.
func EnableLogging(enable bool) {
	logging = enable
}

// init maintains a central location of all things that happen when the package is initialized
// instead of everything being scattered in multiple source files
func init() {
	initLogging()

	err := initConfig()
	if err != nil && logging {
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

	_, err := h.Write([]byte(s))
	if err != nil {
		panic(err)
	}

	val := h.Sum32()
	if b == 0 {
		return val
	}

	return val & ((1 << b) - 1)
}
