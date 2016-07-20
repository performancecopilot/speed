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
	StringLength         = 256
)

// MaxMetricNameLength is the maximum length for a metric name
const MaxMetricNameLength = 63

// MaxDataValueSize is the maximum byte length for a stored metric value, unless it is a string
const MaxDataValueSize = 16

// EraseFileOnStop if set to true, will also delete the memory mapped file
var EraseFileOnStop = false

var writerlog = log.WithField("prefix", "writer")

// Client defines the interface for a type that can talk to an instrumentation agent
type Client interface {
	// a client must contain a registry of metrics
	Registry() Registry

	// starts monitoring
	Start() error

	// Start that will panic on failure
	MustStart()

	// stop monitoring
	Stop() error

	// Stop that will panic on failure
	MustStop()

	// adds a metric to be monitored
	Register(Metric) error

	// tries to add a metric to be written and panics on error
	MustRegister(Metric)

	// adds metric from a string
	RegisterString(string, interface{}, MetricSemantics, MetricType, MetricUnit) error

	// tries to add a metric from a string and panics on an error
	MustRegisterString(string, interface{}, MetricSemantics, MetricType, MetricUnit) error
}

///////////////////////////////////////////////////////////////////////////////

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

// PCPClient implements a client that can generate instrumentation for PCP
type PCPClient struct {
	sync.Mutex
	loc       string            // absolute location of the mmv file
	clusterID uint32            // cluster identifier for the writer
	flag      MMVFlag           // write flag
	r         *PCPRegistry      // current registry
	buffer    bytebuffer.Buffer // current Buffer
}

// NewPCPClient initializes a new PCPClient object
func NewPCPClient(name string, flag MMVFlag) (*PCPClient, error) {
	fileLocation, err := mmvFileLocation(name)
	if err != nil {
		return nil, err
	}

	writerlog.WithField("location", fileLocation).Info("deduced location to write the MMV file")

	return &PCPClient{
		loc:       fileLocation,
		r:         NewPCPRegistry(),
		clusterID: hash(name, PCPClusterIDBitLength),
		flag:      flag,
		buffer:    nil,
	}, nil
}

// Registry returns a writer's registry
func (w *PCPClient) Registry() Registry {
	return w.r
}

func (w *PCPClient) tocCount() int {
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
func (w *PCPClient) Length() int {
	return HeaderLength +
		(w.tocCount() * TocLength) +
		(w.Registry().InstanceCount() * InstanceLength) +
		(w.Registry().InstanceDomainCount() * InstanceDomainLength) +
		(w.Registry().MetricCount() * MetricLength) +
		(w.Registry().ValuesCount() * ValueLength) +
		(w.Registry().StringCount() * StringLength)
}

func (w *PCPClient) initializeInstanceAndInstanceDomainOffsets(instanceoffset, indomoffset int, stringsoffset *int) {
	for _, indom := range w.r.instanceDomains {
		indom.offset = indomoffset
		indom.instanceOffset = instanceoffset
		indomoffset += InstanceDomainLength

		for _, i := range indom.instances {
			i.offset = instanceoffset
			instanceoffset += InstanceLength
		}

		if indom.shortDescription.val != "" {
			indom.shortDescription.offset = *stringsoffset
			*stringsoffset += StringLength
		}

		if indom.longDescription.val != "" {
			indom.longDescription.offset = *stringsoffset
			*stringsoffset += StringLength
		}
	}
}

func (w *PCPClient) initializeSingletonMetricOffsets(metric *PCPSingletonMetric, metricsoffset, valuesoffset, stringsoffset *int) {
	metric.descoffset = *metricsoffset
	*metricsoffset += MetricLength
	metric.valueoffset = *valuesoffset
	*valuesoffset += ValueLength

	if metric.shortDescription.val != "" {
		metric.shortDescription.offset = *stringsoffset
		*stringsoffset += StringLength
	}

	if metric.longDescription.val != "" {
		metric.longDescription.offset = *stringsoffset
		*stringsoffset += StringLength
	}
}

func (w *PCPClient) initializeInstanceMetricOffsets(metric *PCPInstanceMetric, metricsoffset, valuesoffset, stringsoffset *int) {
	metric.descoffset = *metricsoffset
	*metricsoffset += MetricLength

	for name := range metric.indom.instances {
		metric.vals[name].offset = *valuesoffset
		*valuesoffset += ValueLength
	}

	if metric.shortDescription.val != "" {
		metric.shortDescription.offset = *stringsoffset
		*stringsoffset += StringLength
	}

	if metric.longDescription.val != "" {
		metric.longDescription.offset = *stringsoffset
		*stringsoffset += StringLength
	}
}

func (w *PCPClient) initializeOffsets() {
	indomoffset := HeaderLength + TocLength*w.tocCount()
	instanceoffset := indomoffset + InstanceDomainLength*w.r.InstanceDomainCount()
	metricsoffset := instanceoffset + InstanceLength*w.r.InstanceCount()
	valuesoffset := metricsoffset + MetricLength*w.r.MetricCount()
	stringsoffset := valuesoffset + ValueLength*w.r.ValuesCount()

	w.r.indomoffset = indomoffset
	w.r.instanceoffset = instanceoffset
	w.r.metricsoffset = metricsoffset
	w.r.valuesoffset = valuesoffset
	w.r.stringsoffset = stringsoffset

	w.initializeInstanceAndInstanceDomainOffsets(instanceoffset, indomoffset, &stringsoffset)

	for _, metric := range w.r.metrics {
		switch m := metric.(type) {
		case *PCPSingletonMetric:
			w.initializeSingletonMetricOffsets(m, &metricsoffset, &valuesoffset, &stringsoffset)
		case *PCPInstanceMetric:
			w.initializeInstanceMetricOffsets(m, &metricsoffset, &valuesoffset, &stringsoffset)
		}
	}
}

func (w *PCPClient) writeHeaderBlock() (generation2offset int, generation int64) {
	// tag
	w.buffer.MustWriteString("MMV")
	w.buffer.MustSetPos(w.buffer.Pos() + 1) // extra null byte is needed and \0 isn't a valid escape character in go

	// version
	w.buffer.MustWriteUint32(1)

	// generation
	generation = time.Now().Unix()
	w.buffer.MustWriteInt64(generation)

	generation2offset = w.buffer.Pos()

	w.buffer.MustWriteInt64(0)

	// tocCount
	w.buffer.MustWriteInt32(int32(w.tocCount()))

	// flag mask
	w.buffer.MustWriteInt32(int32(w.flag))

	// process identifier
	w.buffer.MustWriteInt32(int32(os.Getpid()))

	// cluster identifier
	w.buffer.MustWriteUint32(w.clusterID)

	return
}

func (w *PCPClient) writeSingleToc(pos, identifier, count, offset int) {
	w.buffer.MustSetPos(pos)
	w.buffer.MustWriteInt32(int32(identifier))
	w.buffer.MustWriteInt32(int32(count))
	w.buffer.MustWriteUint64(uint64(offset))
}

func (w *PCPClient) writeTocBlock() {
	tocpos := HeaderLength

	// instance domains toc
	if w.Registry().InstanceDomainCount() > 0 {
		// 1 is the identifier for instance domains
		w.writeSingleToc(tocpos, 1, w.r.InstanceDomainCount(), w.r.indomoffset)
		tocpos += TocLength
	}

	// instances toc
	if w.Registry().InstanceCount() > 0 {
		// 2 is the identifier for instances
		w.writeSingleToc(tocpos, 2, w.r.InstanceCount(), w.r.instanceoffset)
		tocpos += TocLength
	}

	// metrics and values toc
	metricsoffset, valuesoffset := w.r.metricsoffset, w.r.valuesoffset
	if w.Registry().MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	// 3 is the identifier for metrics
	w.writeSingleToc(tocpos, 3, w.r.MetricCount(), metricsoffset)
	tocpos += TocLength

	// 4 is the identifier for values
	w.writeSingleToc(tocpos, 4, w.r.ValuesCount(), valuesoffset)
	tocpos += TocLength

	// strings toc
	if w.Registry().StringCount() > 0 {
		// 5 is the identifier for strings
		w.writeSingleToc(tocpos, 5, w.r.StringCount(), w.r.stringsoffset)
	}
}

func (w *PCPClient) writeInstanceAndInstanceDomainBlock() {
	for _, indom := range w.r.instanceDomains {
		w.buffer.MustSetPos(indom.offset)
		w.buffer.MustWriteUint32(indom.ID())
		w.buffer.MustWriteInt32(int32(indom.InstanceCount()))
		w.buffer.MustWriteInt64(int64(indom.instanceOffset))

		so, lo := indom.shortDescription.offset, indom.longDescription.offset
		w.buffer.MustWriteInt64(int64(so))
		w.buffer.MustWriteInt64(int64(lo))

		if so != 0 {
			w.buffer.MustSetPos(so)
			w.buffer.MustWriteString(indom.shortDescription.val)
		}

		if lo != 0 {
			w.buffer.MustSetPos(lo)
			w.buffer.MustWriteString(indom.longDescription.val)
		}

		for _, i := range indom.instances {
			w.buffer.MustSetPos(i.offset)
			w.buffer.MustWriteInt64(int64(indom.offset))
			w.buffer.MustWriteInt32(0)
			w.buffer.MustWriteUint32(i.id)
			w.buffer.MustWriteString(i.name)
		}
	}
}

func (w *PCPClient) writeMetricDesc(m PCPMetric, pos int) {
	w.buffer.MustSetPos(pos)

	w.buffer.MustWriteString(m.Name())
	w.buffer.MustSetPos(pos + MaxMetricNameLength + 1)
	w.buffer.MustWriteUint32(m.ID())
	w.buffer.MustWriteInt32(int32(m.Type()))
	w.buffer.MustWriteInt32(int32(m.Semantics()))
	w.buffer.MustWriteUint32(m.Unit().PMAPI())
	if m.Indom() != nil {
		w.buffer.MustWriteUint32(m.Indom().ID())
	} else {
		w.buffer.MustWriteInt32(-1)
	}
	w.buffer.MustWriteInt32(0)

	so, lo := m.ShortDescription().offset, m.LongDescription().offset
	w.buffer.MustWriteInt64(int64(so))
	w.buffer.MustWriteInt64(int64(lo))

	if so != 0 {
		w.buffer.MustSetPos(so)
		w.buffer.MustWriteString(m.ShortDescription().val)
	}

	if lo != 0 {
		w.buffer.MustSetPos(lo)
		w.buffer.MustWriteString(m.LongDescription().val)
	}
}

func (w *PCPClient) writeInstance(t MetricType, val interface{}, valueoffset int) updateClosure {
	offset := valueoffset

	if t == StringType {
		w.buffer.MustSetPos(offset)
		w.buffer.MustWriteUint64(StringLength - 1)
		offset = val.(*PCPString).offset
		w.buffer.MustWriteUint64(uint64(offset))
	}

	update := newupdateClosure(offset, w.buffer, t)
	_ = update(val)

	w.buffer.MustSetPos(valueoffset + MaxDataValueSize)

	return update
}

func (w *PCPClient) writeSingletonMetric(m *PCPSingletonMetric) {
	w.writeMetricDesc(m, m.descoffset)
	m.update = w.writeInstance(m.t, m.val, m.valueoffset)
	w.buffer.MustWriteInt64(int64(m.descoffset))
	w.buffer.MustWriteInt64(0)
}

func (w *PCPClient) writeInstanceMetric(m *PCPInstanceMetric) {
	w.writeMetricDesc(m, m.descoffset)

	for name, i := range m.indom.instances {
		ival := m.vals[name]
		ival.update = w.writeInstance(m.t, ival.val, ival.offset)
		w.buffer.MustWriteInt64(int64(m.descoffset))
		w.buffer.MustWriteInt64(int64(i.offset))
	}
}

func (w *PCPClient) writeMetricsAndValuesBlock() {
	for _, metric := range w.r.metrics {
		switch metric.(type) {
		case *PCPSingletonMetric:
			w.writeSingletonMetric(metric.(*PCPSingletonMetric))
		case *PCPInstanceMetric:
			w.writeInstanceMetric(metric.(*PCPInstanceMetric))
		}
	}
}

// fillData will fill the Buffer with the mmv file
// data as long as something doesn't go wrong
func (w *PCPClient) fillData() error {
	generation2offset, generation := w.writeHeaderBlock()
	w.writeTocBlock()
	w.writeInstanceAndInstanceDomainBlock()
	w.writeMetricsAndValuesBlock()

	w.buffer.MustSetPos(generation2offset)
	w.buffer.MustWriteUint64(uint64(generation))

	return nil
}

// Start dumps existing registry data
func (w *PCPClient) Start() error {
	w.Lock()
	defer w.Unlock()

	l := w.Length()
	writerlog.WithField("length", l).Info("initializing writing the MMV file")

	w.initializeOffsets()
	writerlog.Info("initialized offsets for all written types")

	buffer, err := bytebuffer.NewMemoryMappedBuffer(w.loc, l)
	if err != nil {
		writerlog.WithField("error", err).Error("cannot create MemoryMappedBuffer")
		return err
	}
	w.buffer = buffer
	writerlog.Info("created MemoryMappedBuffer")

	err = w.fillData()
	if err != nil {
		writerlog.WithField("error", err).Error("cannot fill MMV data")
		return err
	}
	writerlog.Info("written data to MMV file")

	w.r.mapped = true

	return nil
}

// MustStart is a start that panics
func (w *PCPClient) MustStart() {
	if err := w.Start(); err != nil {
		panic(err)
	}
}

// Stop removes existing mapping and cleans up
func (w *PCPClient) Stop() error {
	w.Lock()
	defer w.Unlock()

	if !w.r.mapped {
		return errors.New("trying to stop an already stopped mapping")
	}

	writerlog.Info("stopping the writer")

	w.r.mapped = false

	err := w.buffer.(*bytebuffer.MemoryMappedBuffer).Unmap(EraseFileOnStop)
	w.buffer = nil
	if err != nil {
		writerlog.WithField("error", err).Error("error unmapping MemoryMappedBuffer")
		return err
	}

	writerlog.Info("unmapped the memory mapped file")

	return nil
}

// MustStop is a stop that panics
func (w *PCPClient) MustStop() {
	if err := w.Stop(); err != nil {
		panic(err)
	}
}

// Register is simply a shorthand for Registry().AddMetric
func (w *PCPClient) Register(m Metric) error { return w.Registry().AddMetric(m) }

// MustRegister is simply a Register that can panic
func (w *PCPClient) MustRegister(m Metric) {
	if err := w.Register(m); err != nil {
		panic(err)
	}
}

// RegisterIndom is simply a shorthand for Registry().AddInstanceDomain
func (w *PCPClient) RegisterIndom(indom InstanceDomain) error {
	return w.Registry().AddInstanceDomain(indom)
}

// MustRegisterIndom is simply a RegisterIndom that can panic
func (w *PCPClient) MustRegisterIndom(indom InstanceDomain) {
	if err := w.RegisterIndom(indom); err != nil {
		panic(err)
	}
}

// RegisterString is simply a shorthand for Registry().AddMetricByString
func (w *PCPClient) RegisterString(str string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error) {
	return w.Registry().AddMetricByString(str, val, s, t, u)
}

// MustRegisterString is simply a RegisterString that panics
func (w *PCPClient) MustRegisterString(str string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) Metric {
	if m, err := w.RegisterString(str, val, s, t, u); err != nil {
		panic(err)
	} else {
		return m
	}
}
