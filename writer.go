package speed

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"sync"
)

// byte lengths of different components in an mmv file
const (
	HeaderLength         = 40
	TocLength            = 16
	MetricLength         = 104
	ValueLength          = 32
	InstanceLength       = 80
	InstanceDomainLength = 32
	StringBlockLength    = 256
)

// Writer defines the interface of a MMV file writer's properties
type Writer interface {
	io.Writer
	Registry() Registry // a writer must contain a registry of metrics and instance domains
	Start() error       // writes an mmv file
}

func mmvFileLocation(name string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return "", errors.New("name cannot have path separator")
	}

	tdir, present := Config["PCP_TMP_DIR"]
	var loc string
	if present {
		loc = path.Join(RootPath, tdir)
	} else {
		loc = os.TempDir()
	}

	return path.Join(loc, "mmv", name), nil
}

// PCPWriter implements a writer that can write PCP compatible MMV files
type PCPWriter struct {
	sync.Mutex
	loc string       // absolute location of the mmv file
	w   io.Writer    // writer to the mmv file
	r   *PCPRegistry // current registry
}

// NewPCPWriter initializes a new PCPWriter object
func NewPCPWriter(name string) (*PCPWriter, error) {
	fileLocation, err := mmvFileLocation(name)
	if err != nil {
		return nil, err
	}

	w, err := os.Create(fileLocation)
	if err != nil {
		return nil, err
	}

	return &PCPWriter{
		loc: fileLocation,
		r:   NewPCPRegistry(),
		w:   w,
	}, nil
}

// Registry returns a writer's registry
func (w *PCPWriter) Registry() Registry {
	w.Lock()
	defer w.Unlock()
	return w.r
}

// Length returns the byte length of data in the mmv file written by the current writer
func (w *PCPWriter) Length() int {
	w.Lock()
	defer w.Unlock()

	tocCount := 2
	if w.Registry().InstanceCount() > 0 {
		tocCount += 2
	}

	return HeaderLength +
		(tocCount * TocLength) +
		(w.Registry().InstanceCount() * InstanceLength) +
		(w.Registry().InstanceDomainCount() * InstanceDomainLength) +
		(w.Registry().MetricCount() * (MetricLength + ValueLength))
}
