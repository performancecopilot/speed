package speed

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/performancecopilot/speed/bytebuffer"
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

// PCPClusterIDBitLength is the bit length of the cluster id
// for a set of PCP metrics
const PCPClusterIDBitLength = 12

// MMVFlag represents an enumerated type to represent mmv flag values
type MMVFlag int

// values for MMVFlag
const (
	NoPrefixFlag MMVFlag = 1 << iota
	ProcessFlag
	SentinelFlag
)

//go:generate stringer -type=MMVFlag

// PCPWriter implements a writer that can write PCP compatible MMV files
type PCPWriter struct {
	sync.Mutex
	loc       string       // absolute location of the mmv file
	clusterID uint32       // cluster identifier for the writer
	flag      MMVFlag      // write flag
	r         *PCPRegistry // current registry
}

// NewPCPWriter initializes a new PCPWriter object
func NewPCPWriter(name string, flag MMVFlag) (*PCPWriter, error) {
	fileLocation, err := mmvFileLocation(name)
	if err != nil {
		return nil, err
	}

	return &PCPWriter{
		loc:       fileLocation,
		r:         NewPCPRegistry(),
		clusterID: getHash(name, PCPClusterIDBitLength),
		flag:      flag,
	}, nil
}

// Registry returns a writer's registry
func (w *PCPWriter) Registry() Registry {
	w.Lock()
	defer w.Unlock()
	return w.r
}

func (w *PCPWriter) tocCount() int {
	ans := 2

	if w.Registry().InstanceCount() > 0 {
		ans += 2
	}

	return ans
}

// Length returns the byte length of data in the mmv file written by the current writer
func (w *PCPWriter) Length() int {
	w.Lock()
	defer w.Unlock()

	return HeaderLength +
		(w.tocCount() * TocLength) +
		(w.Registry().InstanceCount() * InstanceLength) +
		(w.Registry().InstanceDomainCount() * InstanceDomainLength) +
		(w.Registry().MetricCount() * (MetricLength + ValueLength))
}

func (w *PCPWriter) initializeOffsets() {
	indomoffset := HeaderLength + TocLength*w.tocCount()
	instanceoffset := indomoffset + InstanceDomainLength*w.r.InstanceDomainCount()
	metricsoffset := instanceoffset + InstanceLength*w.r.InstanceCount()
	valuesoffset := metricsoffset + MetricLength*w.r.MetricCount()

	w.r.indomoffset = indomoffset
	w.r.instanceoffset = instanceoffset
	w.r.metricsoffset = metricsoffset
	w.r.valuesoffset = valuesoffset

	for _, indom := range w.r.instanceDomains {
		indom.SetOffset(indomoffset)
		indomoffset += InstanceDomainLength

		for _, i := range indom.instances {
			i.SetOffset(instanceoffset)
			instanceoffset += InstanceLength
		}
	}

	for _, metric := range w.r.metrics {
		metric.desc.SetOffset(metricsoffset)
		metricsoffset += MetricLength
		metric.SetOffset(valuesoffset)
		valuesoffset += ValueLength
	}

	// TODO: string offsets
}

func (w *PCPWriter) writeHeaderBlock(buffer bytebuffer.Buffer) {
	// tag
	buffer.WriteString("MMV")

	// version
	buffer.WriteVal(1)

	// generation
	gen := time.Now().Unix()
	buffer.WriteVal(gen)

	// genOffset := buffer.Pos()

	buffer.WriteVal(0)

	// tocCount
	buffer.WriteVal(w.tocCount())

	// flag mask
	buffer.WriteVal(int(w.flag))

	// process identifier
	buffer.WriteVal(os.Getpid())

	// cluster identifier
	buffer.WriteVal(w.clusterID)
}

func (w *PCPWriter) writeTocBlock(buffer bytebuffer.Buffer) {
	tocpos := HeaderLength

	// instance domains toc
	if w.Registry().InstanceDomainCount() > 0 {
		buffer.SetPos(tocpos)
		buffer.WriteVal(1) // Instance Domain identifier
		buffer.WriteVal(w.Registry().InstanceDomainCount())
		buffer.WriteVal(w.r.indomoffset)
		tocpos += TocLength
	}

	// instances toc
	if w.Registry().InstanceCount() > 0 {
		buffer.SetPos(tocpos)
		buffer.WriteVal(2) // Instance identifier
		buffer.WriteVal(w.Registry().InstanceCount())
		buffer.WriteVal(w.r.instanceoffset)
		tocpos += TocLength
	}

	metricsoffset, valuesoffset := w.r.metricsoffset, w.r.valuesoffset
	if w.Registry().MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	// metrics and values toc
	buffer.SetPos(tocpos)
	buffer.WriteVal(3) // Metrics identifier
	buffer.WriteVal(w.Registry().MetricCount())
	buffer.WriteVal(metricsoffset)
	tocpos += TocLength

	buffer.SetPos(tocpos)
	buffer.WriteVal(4) // Values identifier
	buffer.WriteVal(w.Registry().MetricCount())
	buffer.WriteVal(valuesoffset)
	tocpos += TocLength

	// TODO: strings toc
}

// fillData will fill a byte slice with the mmv file
// data as long as something doesn't go wrong
func (w *PCPWriter) fillData(buffer bytebuffer.Buffer) error {
	w.writeHeaderBlock(buffer)
	w.writeTocBlock(buffer)

	return nil
}
