package speed

import (
	"errors"
	"fmt"
	"math"
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
	case int:
		v := val.(int)
		switch {
		case v < 0:
			return m == Int32Type || m == Int64Type
		case v <= math.MaxInt32:
			return m == Int32Type || m == Int64Type || m == Uint32Type || m == Uint64Type
		case uint32(v) <= math.MaxUint32:
			return m == Int64Type || m == Uint32Type || m == Uint64Type
		case int64(v) <= math.MaxInt64:
			return m == Int64Type || m == Uint64Type
		default:
			return false
		}
	case int32:
		return m == Int32Type
	case int64:
		return m == Int64Type
	case uint:
		v := val.(uint)
		if v > math.MaxUint32 {
			return m == Uint64Type
		}
		return m == Uint32Type || m == Uint64Type
	case uint32:
		return m == Uint32Type
	case uint64:
		return m == Uint64Type
	default:
		return false
	}
}

// WriteVal implements value writer for the current MetricType to a buffer
func (m MetricType) WriteVal(val interface{}, b bytebuffer.Buffer) error {
	switch val.(type) {
	case int:
		switch m {
		case Int32Type:
			return b.WriteInt32(int32(val.(int)))
		case Int64Type:
			return b.WriteInt64(int64(val.(int)))
		case Uint32Type:
			return b.WriteUint32(uint32(val.(int)))
		case Uint64Type:
			return b.WriteUint64(uint64(val.(int)))
		}
	case int32:
		return b.WriteInt32(val.(int32))
	case int64:
		return b.WriteInt64(val.(int64))
	case uint:
		switch m {
		case Uint32Type:
			return b.WriteUint32(uint32(val.(uint)))
		case Uint64Type:
			return b.WriteUint64(uint64(val.(uint)))
		}
	case uint32:
		return b.WriteUint32(val.(uint32))
	case uint64:
		return b.WriteUint64(val.(uint64))
	}

	return errors.New("Invalid Type")
}

///////////////////////////////////////////////////////////////////////////////

// MetricUnit defines the interface for a unit type for speed
type MetricUnit interface {
	// return 32 bit PMAPI representation for the unit
	// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/pmapi.h#L61-L101
	PMAPI() uint32
}

// SpaceUnit is an enumerated type representing all units for space
type SpaceUnit uint32

// Possible values for SpaceUnit
const (
	ByteUnit SpaceUnit = 1<<28 | iota<<16
	KilobyteUnit
	MegabyteUnit
	GigabyteUnit
	TerabyteUnit
	PetabyteUnit
	ExabyteUnit
)

//go:generate stringer -type=SpaceUnit

// PMAPI returns the PMAPI representation for a SpaceUnit
// for space units bits 0-3 are 1 and bits 13-16 are scale
func (s SpaceUnit) PMAPI() uint32 {
	return uint32(s)
}

// TimeUnit is an enumerated type representing all possible units for representing time
type TimeUnit uint32

// Possible Values for TimeUnit
// for time units bits 4-7 are 1 and bits 17-20 are scale
const (
	NanosecondUnit TimeUnit = 1<<24 | iota<<12
	MicrosecondUnit
	MillisecondUnit
	SecondUnit
	MinuteUnit
	HourUnit
)

//go:generate stringer -type=TimeUnit

// PMAPI returns the PMAPI representation for a TimeUnit
func (t TimeUnit) PMAPI() uint32 {
	return uint32(t)
}

// CountUnit is a type representing a counted quantity
type CountUnit uint32

// OneUnit represents the only CountUnit
// for count units bits 8-11 are 1 and bits 21-24 are scale
const OneUnit CountUnit = 1<<20 | iota<<8

//go:generate stringer -type=CountUnit

// PMAPI returns the PMAPI representation for a CountUnit
func (c CountUnit) PMAPI() uint32 {
	return uint32(c)
}

///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////

// Metric defines the general interface a type needs to implement to qualify
// as a valid PCP metric
type Metric interface {
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

///////////////////////////////////////////////////////////////////////////////

// SingletonMetric defines the interface for a metric that stores only one value
type SingletonMetric interface {
	Metric

	// gets the value of the metric
	Val() interface{}

	// Sets the value of the metric to a value, optionally returns an error on failure
	Set(interface{}) error
}

///////////////////////////////////////////////////////////////////////////////

// InstanceMetric defines the interface for a metric that stores multiple values
// in instances and instance domains
type InstanceMetric interface {
	Metric

	// gets the value of a particular instance
	ValInstance(string) (interface{}, error)

	// sets the value of a particular instance
	SetInstance(string, interface{}) error
}

///////////////////////////////////////////////////////////////////////////////

// PCPMetric defines the interface for a metric that is compatible with PCP
type PCPMetric interface {
	Metric

	// a PCPMetric will always have an instance domain, even if it is nil
	Indom() *PCPInstanceDomain

	ShortDescription() *PCPString

	LongDescription() *PCPString
}

///////////////////////////////////////////////////////////////////////////////

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
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	descoffset                        int             // memory storage offset for the metric description
	shortDescription, longDescription *PCPString
}

// newpcpMetricDesc creates a new Metric Description wrapper type
func newpcpMetricDesc(n string, t MetricType, s MetricSemantics, u MetricUnit, shortdesc, longdesc string) *pcpMetricDesc {
	return &pcpMetricDesc{
		getHash(n, PCPMetricItemBitLength),
		n, t, s, u, 0,
		NewPCPString(shortdesc), NewPCPString(longdesc),
	}
}

// ID returns the generated id for PCPMetric
func (md *pcpMetricDesc) ID() uint32 { return md.id }

// Name returns the generated id for PCPMetric
func (md *pcpMetricDesc) Name() string { return md.name }

// Semantics returns the current stored value for PCPMetric
func (md *pcpMetricDesc) Semantics() MetricSemantics { return md.sem }

// Unit returns the unit for PCPMetric
func (md *pcpMetricDesc) Unit() MetricUnit { return md.u }

// Type returns the type for PCPMetric
func (md *pcpMetricDesc) Type() MetricType { return md.t }

// ShortDescription returns the shortdesc value
func (md *pcpMetricDesc) ShortDescription() *PCPString { return md.shortDescription }

// LongDescription returns the longdesc value
func (md *pcpMetricDesc) LongDescription() *PCPString { return md.longDescription }

// Description returns the description for PCPMetric
func (md *pcpMetricDesc) Description() string {
	sd := md.shortDescription
	ld := md.longDescription
	if len(ld.val) > 0 {
		return sd.val + "\n\n" + ld.val
	}
	return sd.val
}

///////////////////////////////////////////////////////////////////////////////

// updateClosure is a closure that will write the modified value of a metric on disk
type updateClosure func(interface{}) error

// newupdateClosure creates a new update closure for an offset, type and buffer
func newupdateClosure(offset int, buffer bytebuffer.Buffer, t MetricType) updateClosure {
	return func(val interface{}) error {
		buffer.SetPos(offset)
		return t.WriteVal(val, buffer)
	}
}

///////////////////////////////////////////////////////////////////////////////

// PCPSingletonMetric defines a singleton metric with no instance domain
// only a value and a valueoffset
type PCPSingletonMetric struct {
	sync.RWMutex
	*pcpMetricDesc
	val         interface{}
	valueoffset int
	update      updateClosure
}

// NewPCPSingletonMetric creates a new instance of PCPSingletonMetric
func NewPCPSingletonMetric(val interface{}, name string, t MetricType, s MetricSemantics, u MetricUnit, shortdesc, longdesc string) (*PCPSingletonMetric, error) {
	if name == "" {
		return nil, errors.New("Metric name cannot be empty")
	}

	if !t.IsCompatible(val) {
		return nil, fmt.Errorf("type %v is not compatible with value %v", t, val)
	}

	return &PCPSingletonMetric{
		sync.RWMutex{},
		newpcpMetricDesc(name, t, s, u, shortdesc, longdesc),
		val, 0, nil,
	}, nil
}

// Val returns the current Set value of PCPSingletonMetric
func (m *PCPSingletonMetric) Val() interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.val
}

// Set Sets the current value of PCPSingletonMetric
func (m *PCPSingletonMetric) Set(val interface{}) error {
	if !m.t.IsCompatible(val) {
		return errors.New("the value is incompatible with this metrics MetricType")
	}

	if val != m.val {
		m.Lock()
		defer m.Unlock()
		if m.update != nil {
			err := m.update(m.val)
			if err != nil {
				return err
			}
		}
		m.val = val
	}

	return nil
}

// Indom returns the instance domain for a PCPSingletonMetric
func (m *PCPSingletonMetric) Indom() *PCPInstanceDomain { return nil }

func (m *PCPSingletonMetric) String() string {
	return fmt.Sprintf("Val: %v\n%v", m.val, m.Description())
}

// TODO: implement PCPCounterMetric, PCPGaugeMetric ...

///////////////////////////////////////////////////////////////////////////////

type instanceValue struct {
	val    interface{}
	offset int
	update updateClosure
}

func newinstanceValue(val interface{}) *instanceValue {
	return &instanceValue{val, 0, nil}
}

// PCPInstanceMetric represents a PCPMetric that can have multiple values
// over multiple instances in an instance domain
type PCPInstanceMetric struct {
	sync.RWMutex
	*pcpMetricDesc
	indom *PCPInstanceDomain
	vals  map[string]*instanceValue
}

// NewPCPInstanceMetric creates a new instance of PCPSingletonMetric
func NewPCPInstanceMetric(vals map[string]interface{}, name string, indom *PCPInstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, shortdesc, longdesc string) (*PCPInstanceMetric, error) {
	if name == "" {
		return nil, errors.New("Metric name cannot be empty")
	}

	if len(vals) != indom.InstanceCount() {
		return nil, errors.New("values for all instances in the instance domain only should be passed")
	}

	mvals := make(map[string]*instanceValue)

	for name := range indom.instances {
		val, present := vals[name]
		if !present {
			return nil, fmt.Errorf("Instance %v not initialized", name)
		}

		if !t.IsCompatible(val) {
			return nil, fmt.Errorf("value %v is incompatible with type %v for Instance %v", val, t, name)
		}

		mvals[name] = newinstanceValue(val)
	}

	return &PCPInstanceMetric{
		sync.RWMutex{},
		newpcpMetricDesc(name, t, s, u, shortdesc, longdesc),
		indom,
		mvals,
	}, nil
}

// Indom returns the instance domain for the metric
func (m *PCPInstanceMetric) Indom() *PCPInstanceDomain { return m.indom }

// ValInstance returns the value for a particular instance of the metric
func (m *PCPInstanceMetric) ValInstance(instance string) (interface{}, error) {
	if !m.indom.HasInstance(instance) {
		return nil, fmt.Errorf("%v is not an instance of this metric", instance)
	}

	m.RLock()
	defer m.RUnlock()

	return m.vals[instance], nil
}

// SetInstance sets the value for a particular instance of the metric
func (m *PCPInstanceMetric) SetInstance(instance string, val interface{}) error {
	if !m.indom.HasInstance(instance) {
		return fmt.Errorf("%v is not an instance of this metric", instance)
	}

	m.Lock()
	defer m.Unlock()

	if m.vals[instance].update != nil {
		err := m.vals[instance].update(val)
		if err != nil {
			return err
		}
	}

	m.vals[instance].val = val
	return nil
}
