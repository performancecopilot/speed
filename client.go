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
func (c *PCPClient) Registry() Registry {
	return c.r
}

func (c *PCPClient) tocCount() int {
	ans := 2

	if c.Registry().InstanceCount() > 0 {
		ans += 2
	}

	if c.Registry().StringCount() > 0 {
		ans++
	}

	return ans
}

// Length returns the byte length of data in the mmv file written by the current writer
func (c *PCPClient) Length() int {
	return HeaderLength +
		(c.tocCount() * TocLength) +
		(c.Registry().InstanceCount() * InstanceLength) +
		(c.Registry().InstanceDomainCount() * InstanceDomainLength) +
		(c.Registry().MetricCount() * MetricLength) +
		(c.Registry().ValuesCount() * ValueLength) +
		(c.Registry().StringCount() * StringLength)
}

func (c *PCPClient) initializeInstanceAndInstanceDomainOffsets(instanceoffset, indomoffset int, stringsoffset *int) {
	for _, indom := range c.r.instanceDomains {
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

func (c *PCPClient) initializeSingletonMetricOffsets(metric *PCPSingletonMetric, metricsoffset, valuesoffset, stringsoffset *int) {
	metric.descoffset = *metricsoffset
	*metricsoffset += MetricLength
	metric.valueoffset = *valuesoffset
	*valuesoffset += ValueLength

	if metric.t == StringType {
		metric.val.(*pcpString).offset = *stringsoffset
		*stringsoffset += StringLength
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

func (c *PCPClient) initializeInstanceMetricOffsets(metric *PCPInstanceMetric, metricsoffset, valuesoffset, stringsoffset *int) {
	metric.descoffset = *metricsoffset
	*metricsoffset += MetricLength

	for name := range metric.indom.instances {
		metric.vals[name].offset = *valuesoffset
		*valuesoffset += ValueLength

		if metric.t == StringType {
			metric.vals[name].val.(*pcpString).offset = *stringsoffset
			*stringsoffset += StringLength
		}
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

func (c *PCPClient) initializeOffsets() {
	indomoffset := HeaderLength + TocLength*c.tocCount()
	instanceoffset := indomoffset + InstanceDomainLength*c.r.InstanceDomainCount()
	metricsoffset := instanceoffset + InstanceLength*c.r.InstanceCount()
	valuesoffset := metricsoffset + MetricLength*c.r.MetricCount()
	stringsoffset := valuesoffset + ValueLength*c.r.ValuesCount()

	c.r.indomoffset = indomoffset
	c.r.instanceoffset = instanceoffset
	c.r.metricsoffset = metricsoffset
	c.r.valuesoffset = valuesoffset
	c.r.stringsoffset = stringsoffset

	c.initializeInstanceAndInstanceDomainOffsets(instanceoffset, indomoffset, &stringsoffset)

	for _, metric := range c.r.metrics {
		switch m := metric.(type) {
		case *PCPSingletonMetric:
			c.initializeSingletonMetricOffsets(m, &metricsoffset, &valuesoffset, &stringsoffset)
		case *PCPInstanceMetric:
			c.initializeInstanceMetricOffsets(m, &metricsoffset, &valuesoffset, &stringsoffset)
		}
	}
}

func (c *PCPClient) writeHeaderBlock() (generation2offset int, generation int64) {
	// tag
	c.buffer.MustWriteString("MMV")
	c.buffer.MustSetPos(c.buffer.Pos() + 1) // extra null byte is needed and \0 isn't a valid escape character in go

	// version
	c.buffer.MustWriteUint32(1)

	// generation
	generation = time.Now().Unix()
	c.buffer.MustWriteInt64(generation)

	generation2offset = c.buffer.Pos()

	c.buffer.MustWriteInt64(0)

	// tocCount
	c.buffer.MustWriteInt32(int32(c.tocCount()))

	// flag mask
	c.buffer.MustWriteInt32(int32(c.flag))

	// process identifier
	c.buffer.MustWriteInt32(int32(os.Getpid()))

	// cluster identifier
	c.buffer.MustWriteUint32(c.clusterID)

	return
}

func (c *PCPClient) writeSingleToc(pos, identifier, count, offset int) {
	c.buffer.MustSetPos(pos)
	c.buffer.MustWriteInt32(int32(identifier))
	c.buffer.MustWriteInt32(int32(count))
	c.buffer.MustWriteUint64(uint64(offset))
}

func (c *PCPClient) writeTocBlock() {
	tocpos := HeaderLength

	// instance domains toc
	if c.Registry().InstanceDomainCount() > 0 {
		// 1 is the identifier for instance domains
		c.writeSingleToc(tocpos, 1, c.r.InstanceDomainCount(), c.r.indomoffset)
		tocpos += TocLength
	}

	// instances toc
	if c.Registry().InstanceCount() > 0 {
		// 2 is the identifier for instances
		c.writeSingleToc(tocpos, 2, c.r.InstanceCount(), c.r.instanceoffset)
		tocpos += TocLength
	}

	// metrics and values toc
	metricsoffset, valuesoffset := c.r.metricsoffset, c.r.valuesoffset
	if c.Registry().MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	// 3 is the identifier for metrics
	c.writeSingleToc(tocpos, 3, c.r.MetricCount(), metricsoffset)
	tocpos += TocLength

	// 4 is the identifier for values
	c.writeSingleToc(tocpos, 4, c.r.ValuesCount(), valuesoffset)
	tocpos += TocLength

	// strings toc
	if c.Registry().StringCount() > 0 {
		// 5 is the identifier for strings
		c.writeSingleToc(tocpos, 5, c.r.StringCount(), c.r.stringsoffset)
	}
}

func (c *PCPClient) writeInstanceAndInstanceDomainBlock() {
	for _, indom := range c.r.instanceDomains {
		c.buffer.MustSetPos(indom.offset)
		c.buffer.MustWriteUint32(indom.ID())
		c.buffer.MustWriteInt32(int32(indom.InstanceCount()))
		c.buffer.MustWriteInt64(int64(indom.instanceOffset))

		so, lo := indom.shortDescription.offset, indom.longDescription.offset
		c.buffer.MustWriteInt64(int64(so))
		c.buffer.MustWriteInt64(int64(lo))

		if so != 0 {
			c.buffer.MustSetPos(so)
			c.buffer.MustWriteString(indom.shortDescription.val)
		}

		if lo != 0 {
			c.buffer.MustSetPos(lo)
			c.buffer.MustWriteString(indom.longDescription.val)
		}

		for _, i := range indom.instances {
			c.buffer.MustSetPos(i.offset)
			c.buffer.MustWriteInt64(int64(indom.offset))
			c.buffer.MustWriteInt32(0)
			c.buffer.MustWriteUint32(i.id)
			c.buffer.MustWriteString(i.name)
		}
	}
}

func (c *PCPClient) writeMetricDesc(desc *PCPMetricDesc, indom *PCPInstanceDomain) {
	c.buffer.MustSetPos(desc.descoffset)

	c.buffer.MustWriteString(desc.name)
	c.buffer.MustSetPos(desc.descoffset + MaxMetricNameLength + 1)
	c.buffer.MustWriteUint32(desc.id)
	c.buffer.MustWriteInt32(int32(desc.t))
	c.buffer.MustWriteInt32(int32(desc.sem))
	c.buffer.MustWriteUint32(desc.u.PMAPI())
	if indom != nil {
		c.buffer.MustWriteUint32(indom.ID())
	} else {
		c.buffer.MustWriteInt32(-1)
	}
	c.buffer.MustWriteInt32(0)

	so, lo := desc.shortDescription.offset, desc.longDescription.offset
	c.buffer.MustWriteInt64(int64(so))
	c.buffer.MustWriteInt64(int64(lo))

	if so != 0 {
		c.buffer.MustSetPos(so)
		c.buffer.MustWriteString(desc.shortDescription.val)
	}

	if lo != 0 {
		c.buffer.MustSetPos(lo)
		c.buffer.MustWriteString(desc.longDescription.val)
	}
}

func (c *PCPClient) writeInstance(t MetricType, val interface{}, valueoffset int) updateClosure {
	offset := valueoffset

	if t == StringType {
		c.buffer.MustSetPos(offset)
		c.buffer.MustWriteUint64(StringLength - 1)
		offset = val.(*pcpString).offset
		c.buffer.MustWriteUint64(uint64(offset))
		val = val.(*pcpString).val
	}

	update := newupdateClosure(offset, c.buffer)
	_ = update(val)

	c.buffer.MustSetPos(valueoffset + MaxDataValueSize)

	return update
}

func (c *PCPClient) writeSingletonMetric(m *PCPSingletonMetric) {
	c.writeMetricDesc(m.PCPMetricDesc, m.Indom())
	m.update = c.writeInstance(m.t, m.val, m.valueoffset)
	c.buffer.MustWriteInt64(int64(m.descoffset))
	c.buffer.MustWriteInt64(0)
}

func (c *PCPClient) writeInstanceMetric(m *PCPInstanceMetric) {
	c.writeMetricDesc(m.PCPMetricDesc, m.Indom())

	for name, i := range m.indom.instances {
		ival := m.vals[name]
		ival.update = c.writeInstance(m.t, ival.val, ival.offset)
		c.buffer.MustWriteInt64(int64(m.descoffset))
		c.buffer.MustWriteInt64(int64(i.offset))
	}
}

func (c *PCPClient) writeMetricsAndValuesBlock() {
	for _, metric := range c.r.metrics {
		switch m := metric.(type) {
		case *PCPSingletonMetric:
			c.writeSingletonMetric(m)
		case *PCPInstanceMetric:
			c.writeInstanceMetric(m)
		}
	}
}

// fillData will fill the Buffer with the mmv file
// data as long as something doesn't go wrong
func (c *PCPClient) fillData() error {
	generation2offset, generation := c.writeHeaderBlock()
	c.writeTocBlock()
	c.writeInstanceAndInstanceDomainBlock()
	c.writeMetricsAndValuesBlock()

	c.buffer.MustSetPos(generation2offset)
	c.buffer.MustWriteUint64(uint64(generation))

	return nil
}

// Start dumps existing registry data
func (c *PCPClient) Start() error {
	c.Lock()
	defer c.Unlock()

	l := c.Length()
	writerlog.WithField("length", l).Info("initializing writing the MMV file")

	c.initializeOffsets()
	writerlog.Info("initialized offsets for all written types")

	buffer, err := bytebuffer.NewMemoryMappedBuffer(c.loc, l)
	if err != nil {
		writerlog.WithField("error", err).Error("cannot create MemoryMappedBuffer")
		return err
	}
	c.buffer = buffer
	writerlog.Info("created MemoryMappedBuffer")

	err = c.fillData()
	if err != nil {
		writerlog.WithField("error", err).Error("cannot fill MMV data")
		return err
	}
	writerlog.Info("written data to MMV file")

	c.r.mapped = true

	return nil
}

// MustStart is a start that panics
func (c *PCPClient) MustStart() {
	if err := c.Start(); err != nil {
		panic(err)
	}
}

// Stop removes existing mapping and cleans up
func (c *PCPClient) Stop() error {
	c.Lock()
	defer c.Unlock()

	if !c.r.mapped {
		return errors.New("trying to stop an already stopped mapping")
	}

	writerlog.Info("stopping the writer")

	c.r.mapped = false

	err := c.buffer.(*bytebuffer.MemoryMappedBuffer).Unmap(EraseFileOnStop)
	c.buffer = nil
	if err != nil {
		writerlog.WithField("error", err).Error("error unmapping MemoryMappedBuffer")
		return err
	}

	writerlog.Info("unmapped the memory mapped file")

	return nil
}

// MustStop is a stop that panics
func (c *PCPClient) MustStop() {
	if err := c.Stop(); err != nil {
		panic(err)
	}
}

// Register is simply a shorthand for Registry().AddMetric
func (c *PCPClient) Register(m Metric) error { return c.Registry().AddMetric(m) }

// MustRegister is simply a Register that can panic
func (c *PCPClient) MustRegister(m Metric) {
	if err := c.Register(m); err != nil {
		panic(err)
	}
}

// RegisterIndom is simply a shorthand for Registry().AddInstanceDomain
func (c *PCPClient) RegisterIndom(indom InstanceDomain) error {
	return c.Registry().AddInstanceDomain(indom)
}

// MustRegisterIndom is simply a RegisterIndom that can panic
func (c *PCPClient) MustRegisterIndom(indom InstanceDomain) {
	if err := c.RegisterIndom(indom); err != nil {
		panic(err)
	}
}

// RegisterString is simply a shorthand for Registry().AddMetricByString
func (c *PCPClient) RegisterString(str string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) (Metric, error) {
	return c.Registry().AddMetricByString(str, val, s, t, u)
}

// MustRegisterString is simply a RegisterString that panics
func (c *PCPClient) MustRegisterString(str string, val interface{}, s MetricSemantics, t MetricType, u MetricUnit) Metric {
	if m, err := c.RegisterString(str, val, s, t, u); err != nil {
		panic(err)
	} else {
		return m
	}
}
