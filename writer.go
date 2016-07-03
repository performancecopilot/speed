package speed

import (
	"errors"
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

// MaxMetricNameLength is the maximum length for a metric name
const MaxMetricNameLength = 63

// MaxDataValueSize is the maximum byte length for a stored metric value, unless it is a string
const MaxDataValueSize = 16

// Writer defines the interface of a MMV file writer's properties
type Writer interface {
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

	if w.Registry().StringCount() > 0 {
		ans++
	}

	return ans
}

// Length returns the byte length of data in the mmv file written by the current writer
func (w *PCPWriter) Length() int {
	return HeaderLength +
		(w.tocCount() * TocLength) +
		(w.Registry().InstanceCount() * InstanceLength) +
		(w.Registry().InstanceDomainCount() * InstanceDomainLength) +
		(w.Registry().MetricCount() * (MetricLength + ValueLength)) +
		(w.Registry().StringCount() * StringBlockLength)
}

func (w *PCPWriter) initializeOffsets() {
	indomoffset := HeaderLength + TocLength*w.tocCount()
	instanceoffset := indomoffset + InstanceDomainLength*w.r.InstanceDomainCount()
	metricsoffset := instanceoffset + InstanceLength*w.r.InstanceCount()
	valuesoffset := metricsoffset + MetricLength*w.r.MetricCount()
	stringsoffset := valuesoffset + ValueLength*w.r.MetricCount()

	w.r.indomoffset = indomoffset
	w.r.instanceoffset = instanceoffset
	w.r.metricsoffset = metricsoffset
	w.r.valuesoffset = valuesoffset
	w.r.stringsoffset = stringsoffset

	for _, indom := range w.r.instanceDomains {
		indom.offset = indomoffset
		indom.instanceOffset = instanceoffset
		indomoffset += InstanceDomainLength

		for _, i := range indom.instances {
			i.offset = instanceoffset
			instanceoffset += InstanceLength
		}

		if indom.shortHelpText.val != "" {
			indom.shortHelpText.offset = stringsoffset
			stringsoffset += StringBlockLength
		}

		if indom.longHelpText.val != "" {
			indom.longHelpText.offset = stringsoffset
			stringsoffset += StringBlockLength
		}
	}

	for _, metric := range w.r.metrics {
		metric.desc.offset = metricsoffset
		metricsoffset += MetricLength
		metric.offset = valuesoffset
		valuesoffset += ValueLength

		if metric.desc.shortDescription.val != "" {
			metric.desc.shortDescription.offset = stringsoffset
			stringsoffset += StringBlockLength
		}

		if metric.desc.longDescription.val != "" {
			metric.desc.longDescription.offset = stringsoffset
			stringsoffset += StringBlockLength
		}
	}
}

func (w *PCPWriter) writeHeaderBlock(buffer bytebuffer.Buffer) (generation2offset int, generation int64) {
	// tag
	buffer.WriteString("MMV")
	buffer.SetPos(buffer.Pos() + 1) // extra null byte is needed and \0 isn't a valid escape character in go

	// version
	buffer.WriteUint32(1)

	// generation
	generation = time.Now().Unix()
	buffer.WriteInt64(generation)

	generation2offset = buffer.Pos()

	buffer.WriteInt64(0)

	// tocCount
	buffer.WriteInt32(int32(w.tocCount()))

	// flag mask
	buffer.WriteInt32(int32(w.flag))

	// process identifier
	buffer.WriteInt32(int32(os.Getpid()))

	// cluster identifier
	buffer.WriteUint32(w.clusterID)

	return
}

func (w *PCPWriter) writeSingleToc(pos, identifier, count, offset int, buffer bytebuffer.Buffer) {
	buffer.SetPos(pos)
	buffer.WriteInt32(int32(identifier))
	buffer.WriteInt32(int32(count))
	buffer.WriteUint64(uint64(offset))
}

func (w *PCPWriter) writeTocBlock(buffer bytebuffer.Buffer) {
	tocpos := HeaderLength

	// instance domains toc
	if w.Registry().InstanceDomainCount() > 0 {
		// 1 is the identifier for instance domains
		w.writeSingleToc(tocpos, 1, w.r.InstanceDomainCount(), w.r.indomoffset, buffer)
		tocpos += TocLength
	}

	// instances toc
	if w.Registry().InstanceCount() > 0 {
		// 2 is the identifier for instances
		w.writeSingleToc(tocpos, 2, w.r.InstanceCount(), w.r.instanceoffset, buffer)
		tocpos += TocLength
	}

	// metrics and values toc
	metricsoffset, valuesoffset := w.r.metricsoffset, w.r.valuesoffset
	if w.Registry().MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	// 3 is the identifier for metrics
	w.writeSingleToc(tocpos, 3, w.r.MetricCount(), metricsoffset, buffer)
	tocpos += TocLength

	// 4 is the identifier for values
	w.writeSingleToc(tocpos, 4, w.r.MetricCount(), valuesoffset, buffer)
	tocpos += TocLength

	// strings toc
	if w.Registry().StringCount() > 0 {
		// 5 is the identifier for strings
		w.writeSingleToc(tocpos, 5, w.r.StringCount(), w.r.stringsoffset, buffer)
	}
}

func (w *PCPWriter) writeInstanceAndInstanceDomainBlock(buffer bytebuffer.Buffer) {
	for _, indom := range w.r.instanceDomains {
		buffer.SetPos(indom.offset)
		buffer.WriteUint32(indom.ID())
		buffer.WriteInt32(int32(indom.InstanceCount()))
		buffer.WriteInt64(int64(indom.instanceOffset))

		so, lo := indom.shortHelpText.offset, indom.longHelpText.offset
		buffer.WriteInt64(int64(so))
		buffer.WriteInt64(int64(lo))

		if so != 0 {
			buffer.SetPos(so)
			buffer.WriteString(indom.shortHelpText.val)
		}

		if lo != 0 {
			buffer.SetPos(lo)
			buffer.WriteString(indom.longHelpText.val)
		}

		for _, i := range indom.instances {
			buffer.SetPos(i.offset)
			buffer.WriteInt64(int64(indom.offset))
			buffer.WriteInt32(0)
			buffer.WriteUint32(i.id)
			buffer.WriteString(i.name)
		}
	}
}

func (w *PCPWriter) writeMetricDesc(desc *pcpMetricDesc, buffer bytebuffer.Buffer) {
	pos := desc.offset
	buffer.SetPos(pos)

	buffer.WriteString(desc.name)
	buffer.Write([]byte{0})
	buffer.SetPos(pos + MaxMetricNameLength + 1)
	buffer.WriteUint32(desc.id)
	buffer.WriteInt32(int32(desc.t))
	buffer.WriteInt32(int32(desc.sem))
	buffer.WriteUint32(desc.u.PMAPI())
	if desc.indom != nil {
		buffer.WriteUint32(desc.indom.ID())
	} else {
		buffer.WriteInt32(-1)
	}
	buffer.WriteInt32(0)

	so, lo := desc.shortDescription.offset, desc.longDescription.offset
	buffer.WriteInt64(int64(so))
	buffer.WriteInt64(int64(lo))

	if so != 0 {
		buffer.SetPos(so)
		buffer.WriteString(desc.shortDescription.val)
	}

	if lo != 0 {
		buffer.SetPos(lo)
		buffer.WriteString(desc.longDescription.val)
	}
}

func (w *PCPWriter) writeMetricVal(m *PCPMetric, buffer bytebuffer.Buffer) {
	pos := m.offset
	buffer.SetPos(pos)

	m.desc.t.WriteVal(m.val, buffer)

	buffer.SetPos(pos + MaxDataValueSize)
	buffer.WriteInt64(int64(m.desc.offset))
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

// fillData will fill the Buffer with the mmv file
// data as long as something doesn't go wrong
func (w *PCPWriter) fillData(buffer bytebuffer.Buffer) error {
	generation2offset, generation := w.writeHeaderBlock(buffer)
	w.writeTocBlock(buffer)
	w.writeInstanceAndInstanceDomainBlock(buffer)
	w.writeMetricsAndValuesBlock(buffer)

	buffer.SetPos(generation2offset)
	buffer.WriteUint64(uint64(generation))

	return nil
}

// Start dumps existing registry data
func (w *PCPWriter) Start() error {
	w.Lock()
	defer w.Unlock()

	l := w.Length()

	w.initializeOffsets()

	buffer, err := bytebuffer.NewMemoryMappedBuffer(w.loc, l)
	if err != nil {
		return err
	}

	w.fillData(buffer)

	w.r.mapped = true

	return nil
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
