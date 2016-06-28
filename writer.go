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

	// a writer must contain a registry of metrics and instance domains
	Registry() Registry

	// writes an mmv file
	Start() error

	// stops writing and cleans up
	Stop() error

	// adds a metric to be written
	Register(Metric)

	// adds an instance domain to be written
	RegisterIndom(InstanceDomain)

	// adds metric and instance domain from a string
	RegisterString(s string)

	// update a metric
	Update(Metric, interface{})
}

func mmvFileLocation(name string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return "", errors.New("name cannot have path separator")
	}

	tdir, present := config["PCP_TMP_DIR"]
	var loc string
	if present {
		loc = path.Join(rootPath, tdir)
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
		indom.setOffset(indomoffset)
		indom.instanceOffset = instanceoffset
		indomoffset += InstanceDomainLength

		for _, i := range indom.instances {
			i.setOffset(instanceoffset)
			instanceoffset += InstanceLength
		}
	}

	for _, metric := range w.r.metrics {
		metric.desc.setOffset(metricsoffset)
		metricsoffset += MetricLength
		metric.setOffset(valuesoffset)
		valuesoffset += ValueLength
	}

	// TODO: string offsets
}

func (w *PCPWriter) writeHeaderBlock(buffer bytebuffer.Buffer) (gen2offset int, generation int64) {
	// tag
	buffer.WriteString("MMV")
	buffer.SetPos(buffer.Pos() + 1) // extra null byte is needed and \0 isn't a valid escape character in go

	// version
	buffer.WriteUint32(1)

	// generation
	generation = time.Now().Unix()
	buffer.WriteInt64(generation)

	gen2offset = buffer.Pos()

	buffer.WriteInt64(0)

	// tocCount
	buffer.WriteInt(w.tocCount())

	// flag mask
	buffer.WriteInt(int(w.flag))

	// process identifier
	buffer.WriteInt(os.Getpid())

	// cluster identifier
	buffer.WriteUint32(w.clusterID)

	return
}

func (w *PCPWriter) writeTocBlock(buffer bytebuffer.Buffer) {
	tocpos := HeaderLength

	// instance domains toc
	if w.Registry().InstanceDomainCount() > 0 {
		buffer.SetPos(tocpos)
		buffer.WriteInt(1) // Instance Domain identifier
		buffer.WriteInt(w.Registry().InstanceDomainCount())
		buffer.WriteUint64(uint64(w.r.indomoffset))
		tocpos += TocLength
	}

	// instances toc
	if w.Registry().InstanceCount() > 0 {
		buffer.SetPos(tocpos)
		buffer.WriteInt(2) // Instance identifier
		buffer.WriteInt(w.Registry().InstanceCount())
		buffer.WriteUint64(uint64(w.r.instanceoffset))
		tocpos += TocLength
	}

	metricsoffset, valuesoffset := w.r.metricsoffset, w.r.valuesoffset
	if w.Registry().MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	// metrics and values toc
	buffer.SetPos(tocpos)
	buffer.WriteInt(3) // Metrics identifier
	buffer.WriteInt(w.Registry().MetricCount())
	buffer.WriteUint64(uint64(metricsoffset))
	tocpos += TocLength

	buffer.SetPos(tocpos)
	buffer.WriteInt(4) // Values identifier
	buffer.WriteInt(w.Registry().MetricCount())
	buffer.WriteUint64(uint64(valuesoffset))
	tocpos += TocLength

	// TODO: strings toc
}

func (w *PCPWriter) writeInstanceAndInstanceDomainBlock(buffer bytebuffer.Buffer) {
	for _, indom := range w.r.instanceDomains {
		buffer.SetPos(indom.Offset())
		buffer.WriteUint32(indom.ID())
		buffer.WriteInt(indom.InstanceCount())
		buffer.WriteInt64(int64(indom.instanceOffset))
		// TODO: write indom string descriptions offsets

		for _, i := range indom.instances {
			buffer.SetPos(i.Offset())
			buffer.WriteInt64(int64(indom.Offset()))
			buffer.WriteInt(0)
			buffer.WriteUint32(i.id)
			buffer.WriteString(i.name)
		}
	}
}

const (
	MetricNameLimit = 63
	DataValueLength = 16
)

func (w *PCPWriter) writeMetricDesc(desc *pcpMetricDesc, buffer bytebuffer.Buffer) {
	pos := desc.Offset()
	buffer.SetPos(pos)

	buffer.WriteString(desc.name)
	buffer.Write([]byte{0})
	buffer.SetPos(pos + MetricNameLimit + 1)
	buffer.WriteUint32(desc.id)
	buffer.WriteInt32(int32(desc.t))
	buffer.WriteInt32(int32(desc.sem))
	buffer.WriteInt32(int32(desc.u)) // TODO: fix this
	if desc.indom != nil {
		buffer.WriteUint32(desc.indom.ID())
	} else {
		buffer.WriteInt32(-1)
	}
	buffer.WriteInt(0)
	// TODO: write string descriptions
}

func (w *PCPWriter) writeMetricVal(m *PCPMetric, buffer bytebuffer.Buffer) {
	pos := m.Offset()
	buffer.SetPos(pos)

	switch m.desc.t {
	case Int32Type:
		buffer.WriteInt32(m.val.(int32))
	case Int64Type:
		buffer.WriteInt64(m.val.(int64))
	case Uint32Type:
		buffer.WriteUint32(m.val.(uint32))
	case Uint64Type:
		buffer.WriteUint64(m.val.(uint64))
	}

	buffer.SetPos(pos + DataValueLength)
	buffer.WriteInt64(int64(m.desc.Offset()))
	if m.desc.indom != nil {
		buffer.WriteInt64(int64(m.desc.indom.(*PCPInstanceDomain).instanceOffset))
	} else {
		buffer.WriteInt64(0)
	}
}

func (w *PCPWriter) writeMetricsAndValuesBlock(buffer bytebuffer.Buffer) {
	for _, metric := range w.r.metrics {
		w.writeMetricDesc(metric.desc, buffer)
		w.writeMetricVal(metric, buffer)
	}
}

// fillData will fill a byte slice with the mmv file
// data as long as something doesn't go wrong
func (w *PCPWriter) fillData(buffer bytebuffer.Buffer) error {
	gen2offset, generation := w.writeHeaderBlock(buffer)
	w.writeTocBlock(buffer)
	w.writeInstanceAndInstanceDomainBlock(buffer)
	w.writeMetricsAndValuesBlock(buffer)
	// TODO: write strings block

	buffer.SetPos(gen2offset)
	buffer.WriteUint64(uint64(generation))

	return nil
}

func (w *PCPWriter) Write(data []byte) (int, error) {
	f, err := os.Create(w.loc)
	if err != nil {
		panic(err)
	}

	return f.Write(data)
}

// Start dumps existing registry data
func (w *PCPWriter) Start() {
	w.Lock()
	defer w.Unlock()

	l := w.Length()

	w.initializeOffsets()
	buffer := bytebuffer.NewByteBuffer(l)
	w.fillData(buffer)

	w.Write(buffer.Bytes())
	w.r.mapped = true
}

// Stop removes existing mapping and cleans up
func (w *PCPWriter) Stop() {
	w.Lock()
	defer w.Unlock()

	w.r.mapped = false
}

// Register is simply a shorthand for Registry().AddMetric
func (w *PCPWriter) Register(m Metric) { w.Registry().AddMetric(m) }

// RegisterIndom is simply a shorthand for Registry().AddInstanceDomain
func (w *PCPWriter) RegisterIndom(indom InstanceDomain) { w.Registry().AddInstanceDomain(indom) }

// RegisterString is simply a shorthand for Registry().AddMetricByString
func (w *PCPWriter) RegisterString(str string, initialval interface{}, s MetricSemantics, t MetricType, u MetricUnit) {
	w.Registry().AddMetricByString(str, initialval, s, t, u)
}

// Update is simply a shorthand for Registry().UpdateMetric
func (w *PCPWriter) Update(m Metric, val interface{}) { w.Registry().UpdateMetric(m, val) }
