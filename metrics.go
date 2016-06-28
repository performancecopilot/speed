package speed

import (
	"errors"
	"fmt"
	"sync"

	"github.com/performancecopilot/speed/bytebuffer"
)

// MetricType is an enumerated type representing all valid types for a metric
type MetricType int32

// Possible values for a MetricType
const (
	NoSupportType       MetricType = -1
	Int32Type           MetricType = 0
	Uint32Type          MetricType = 1
	Int64Type           MetricType = 2
	Uint64Type          MetricType = 3
	FloatType           MetricType = 4
	DoubleType          MetricType = 5
	StringType          MetricType = 6
	AggregateType       MetricType = 7
	AggregateStaticType MetricType = 8
	EventType           MetricType = 9
	HighresEventType    MetricType = 10
	UnknownType         MetricType = 255
)

//go:generate stringer -type=MetricType

// IsCompatible checks if the passed value is compatible with the current MetricType
func (m MetricType) IsCompatible(val interface{}) bool {
	switch val.(type) {
	case int32:
		return m == Int32Type
	case uint32:
		return m == Uint32Type
	case int64:
		return m == Int64Type
	case uint64:
		return m == Uint64Type
	default:
		return false
	}
}

// WriteVal implements value writer for the current MetricType to a buffer
func (m MetricType) WriteVal(val interface{}, b bytebuffer.Buffer) {
	switch m {
	case Int32Type:
		b.WriteInt32(val.(int32))
	case Int64Type:
		b.WriteInt64(val.(int64))
	case Uint32Type:
		b.WriteUint32(val.(uint32))
	case Uint64Type:
		b.WriteUint64(val.(uint64))
	}
}

// MetricUnit is an enumerated type representing all possible values for a valid PCP unit
type MetricUnit int32

// SpaceUnit is an enumerated type representing all units for space
type SpaceUnit MetricUnit

// Possible values for SpaceUnit
const (
	ByteUnit SpaceUnit = iota
	KilobyteUnit
	MegabyteUnit
	GigabyteUnit
	TerabyteUnit
	PetabyteUnit
	ExabyteUnit
)

//go:generate stringer -type=SpaceUnit

// TimeUnit is an enumerated type representing all possible units for representing time
type TimeUnit MetricUnit

// Possible Values for TimeUnit
const (
	NanosecondUnit TimeUnit = iota
	MicrosecondUnit
	MillisecondUnit
	SecondUnit
	MinuteUnit
	HourUnit
)

//go:generate stringer -type=TimeUnit

// CountUnit is a type representing a counted quantity
type CountUnit MetricUnit

// OneUnit represents the only CountUnit
const OneUnit CountUnit = iota

//go:generate stringer -type=CountUnit

// MetricSemantics represents an enumerated type representing the possible
// values for the semantics of a metric
type MetricSemantics int32

// Possible values for MetricSemantics
const (
	NoSemantics MetricSemantics = iota
	CounterSemantics
	InstantSemantics
	DiscreteSemantics
)

//go:generate stringer -type=MetricSemantics

// Metric defines the general interface a type needs to implement to qualify
// as a valid PCP metric
type Metric interface {
	// gets the value of the metric
	Val() interface{}

	// Sets the value of the metric to a value, optionally returns an error on failure
	Set(interface{}) error

	// gets the unique id generated for this metric
	ID() uint32

	// gets the name for the metric
	Name() string

	// gets the type of a metric
	Type() MetricType

	// gets the unit of a metric
	Unit() MetricUnit

	// gets the semantics for a metric
	Semantics() MetricSemantics

	// gets the description of a metric
	Description() string
}

// PCPMetricItemBitLength is the maximum bit size of a PCP Metric id
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPMetricItemBitLength = 10

// pcpMetricDesc is a metric metadata wrapper
// each metric type can wrap its metadata by containing a pcpMetricDesc type and only define its own
// specific properties assuming pcpMetricDesc will handle the rest
//
// when writing, this type is supposed to map directly to the pmDesc struct as defined in PCP core
type pcpMetricDesc struct {
	id                                uint32          // unique metric id
	name                              string          // the name
	indom                             InstanceDomain  // the instance domain
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	offset                            int             // memory storage offset for the metric description
	shortDescription, longDescription *PCPString
}

// newpcpMetricDesc creates a new Metric Description wrapper type
func newpcpMetricDesc(n string, i InstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, short, long string) *pcpMetricDesc {
	return &pcpMetricDesc{
		getHash(n, PCPMetricItemBitLength),
		n, i, t, s, u, 0,
		NewPCPString(short), NewPCPString(long),
	}
}

// Offset returns the memory offset the metric description will be written at
func (md *pcpMetricDesc) Offset() int { return md.offset }

// setOffset Sets the memory offset the metric description will be written at
func (md *pcpMetricDesc) setOffset(offset int) { md.offset = offset }

// PCPMetric defines a PCP compatible metric type that can be constructed by specifying values
// for type, semantics and unit
type PCPMetric struct {
	sync.RWMutex
	val    interface{}    // all bets are off, store whatever you want
	desc   *pcpMetricDesc // the metadata associated with this metric
	offset int            // memory storage offset for the metric value
}

// NewPCPMetric creates a new instance of PCPMetric
func NewPCPMetric(val interface{}, name string, indom InstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, short, long string) (*PCPMetric, error) {
	if !t.IsCompatible(val) {
		return nil, errors.New("the passed MetricType and values are incompatible")
	}

	return &PCPMetric{
		val:  val,
		desc: newpcpMetricDesc(name, indom, t, s, u, short, long),
	}, nil
}

// Val returns the current Set value of PCPMetric
func (m *PCPMetric) Val() interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.val
}

// Set Sets the current value of PCPMetric
func (m *PCPMetric) Set(val interface{}) error {
	if !m.desc.t.IsCompatible(val) {
		return errors.New("the value is incompatible with this metrics MetricType")
	}

	if val != m.val {
		m.Lock()
		defer m.Unlock()
		m.val = val
	}
	return nil
}

// ID returns the generated id for PCPMetric
func (m *PCPMetric) ID() uint32 { return m.desc.id }

// Name returns the generated id for PCPMetric
func (m *PCPMetric) Name() string { return m.desc.name }

// Semantics returns the current stored value for PCPMetric
func (m *PCPMetric) Semantics() MetricSemantics { return m.desc.sem }

// Unit returns the unit for PCPMetric
func (m *PCPMetric) Unit() MetricUnit { return m.desc.u }

// Type returns the type for PCPMetric
func (m *PCPMetric) Type() MetricType { return m.desc.t }

// Description returns the description for PCPMetric
func (m *PCPMetric) Description() string {
	sd := m.desc.shortDescription
	ld := m.desc.longDescription
	if len(ld.val) > 0 {
		return sd.val + "\n\n" + ld.val
	}
	return sd.val
}

// Offset returns the memory offset the metric value will be written at
func (m *PCPMetric) Offset() int { return m.offset }

// setOffset Sets the memory offset the metric value will be written at
func (m *PCPMetric) setOffset(offset int) { m.offset = offset }

func (m *PCPMetric) String() string {
	return fmt.Sprintf("Val: %v\n%v", m.val, m.Description())
}

// TODO: implement PCPCounterMetric, PCPGaugeMetric ...
