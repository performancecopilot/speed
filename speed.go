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
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Version is the last tagged version of the package
const Version = "2.0.0"

var logging bool
var logWriters = []zapcore.WriteSyncer{os.Stdout}
var logger *zap.Logger
var zapEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
}

func initLogging() {
	logging = false
	initializeLogger()
}

// EnableLogging logging enables logging for logrus if true is passed
// and disables it if false is passed.
func EnableLogging(enable bool) {
	logging = enable
}

// AddLogWriter adds a new io.Writer as a target for writing
// logs.
func AddLogWriter(writer io.Writer) {
	logWriters = append(logWriters, zapcore.AddSync(writer))
	initializeLogger()
}

// SetLogWriters will set the passed io.Writer instances as targets for
// writing logs.
func SetLogWriters(writers ...io.Writer) {
	writesyncers := make([]zapcore.WriteSyncer, 0, len(writers))
	for _, w := range writers {
		writesyncers = append(writesyncers, zapcore.AddSync(w))
	}

	logWriters = writesyncers
	initializeLogger()
}

func initializeLogger() {
	ws := zap.CombineWriteSyncers(logWriters...)
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapEncoderConfig),
		ws, zapcore.InfoLevel,
	))
}

// init maintains a central location of all things that happen when the package is initialized
// instead of everything being scattered in multiple source files
func init() {
	initLogging()

	err := initConfig()
	if err != nil && logging {
		logger.Error("error initializing config. maybe PCP isn't installed properly",
			zap.String("module", "config"),
			zap.Error(err),
		)
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
