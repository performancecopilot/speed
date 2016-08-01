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
	NoSupportType MetricType = iota - 1
	Int32Type
	Uint32Type
	Int64Type
	Uint64Type
	FloatType
	DoubleType
	StringType
	UnknownType MetricType = 255
)

//go:generate stringer -type=MetricType

func (m MetricType) isCompatibleInt(val int) bool {
	v := int64(val)
	switch m {
	case Int32Type:
		return v >= math.MinInt32 && v <= math.MaxInt32
	case Int64Type:
		return v >= math.MinInt64 && v <= math.MaxInt64
	case Uint32Type:
		return v >= 0 && v <= math.MaxUint32
	case Uint64Type:
		return v >= 0 && uint64(v) <= math.MaxUint64
	}
	return false
}

func (m MetricType) isCompatibleUint(val uint) bool {
	switch {
	case val <= math.MaxUint32:
		return m == Uint32Type || m == Uint64Type
	default:
		return m == Uint64Type
	}
}

func (m MetricType) isCompatibleFloat(val float64) bool {
	switch {
	case val >= -math.MaxFloat32 && val <= math.MaxFloat32:
		return m == FloatType || m == DoubleType
	default:
		return m == DoubleType
	}
}

// IsCompatible checks if the passed value is compatible with the current MetricType
func (m MetricType) IsCompatible(val interface{}) bool {
	switch v := val.(type) {
	case int:
		return m.isCompatibleInt(v)
	case int32:
		return m == Int32Type
	case int64:
		return m == Int64Type
	case uint:
		return m.isCompatibleUint(v)
	case uint32:
		return m == Uint32Type
	case uint64:
		return m == Uint64Type
	case float32:
		return m == FloatType
	case float64:
		return m.isCompatibleFloat(v)
	case string:
		return m == StringType
	}
	return false
}

// resolveInt will resolve an int to one of the 4 compatible types
func (m MetricType) resolveInt(val interface{}) interface{} {
	if vi, isInt := val.(int); isInt {
		switch m {
		case Int64Type:
			return int64(vi)
		case Uint32Type:
			return uint32(vi)
		case Uint64Type:
			return uint64(vi)
		}
		return int32(val.(int))
	}

	if vui, isUint := val.(uint); isUint {
		if m == Uint64Type {
			return uint64(vui)
		}
		return uint32(vui)
	}

	return val
}

// resolveFloat will resolve a float64 to one of the 2 compatible types
func (m MetricType) resolveFloat(val interface{}) interface{} {
	_, isFloat64 := val.(float64)
	if isFloat64 && m == FloatType {
		return float32(val.(float64))
	}

	return val
}

func (m MetricType) resolve(val interface{}) interface{} {
	if sval, isString := val.(string); isString {
		if len(sval) > StringLength {
			sval = sval[:StringLength]
		}
		return newpcpString(sval)
	}

	val = m.resolveInt(val)
	val = m.resolveFloat(val)

	return val
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
	_
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

	// sets the value of the metric to a value, optionally returns an error on failure
	Set(interface{}) error

	// tries to set and panics on error
	MustSet(interface{})
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

	// tries to set the value of a particular instance and panics on error
	MustSetInstance(string, interface{})
}

///////////////////////////////////////////////////////////////////////////////

// PCPMetric defines the interface for a metric that is compatible with PCP
type PCPMetric interface {
	Metric

	// a PCPMetric will always have an instance domain, even if it is nil
	Indom() *PCPInstanceDomain

	ShortDescription() string

	LongDescription() string
}

///////////////////////////////////////////////////////////////////////////////

// PCPMetricItemBitLength is the maximum bit size of a PCP Metric id
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPMetricItemBitLength = 10

// PCPMetricDesc is a metric metadata wrapper
// each metric type can wrap its metadata by containing a PCPMetricDesc type and only define its own
// specific properties assuming PCPMetricDesc will handle the rest
//
// when writing, this type is supposed to map directly to the pmDesc struct as defined in PCP core
type PCPMetricDesc struct {
	id                                uint32          // unique metric id
	name                              string          // the name
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	descoffset                        int             // memory storage offset for the metric description
	shortDescription, longDescription *pcpString
}

// newPCPMetricDesc creates a new Metric Description wrapper type
func newPCPMetricDesc(n string, t MetricType, s MetricSemantics, u MetricUnit, shortdesc, longdesc string) *PCPMetricDesc {
	return &PCPMetricDesc{
		hash(n, PCPMetricItemBitLength),
		n, t, s, u, 0,
		newpcpString(shortdesc), newpcpString(longdesc),
	}
}

// ID returns the generated id for PCPMetric
func (md *PCPMetricDesc) ID() uint32 { return md.id }

// Name returns the generated id for PCPMetric
func (md *PCPMetricDesc) Name() string { return md.name }

// Semantics returns the current stored value for PCPMetric
func (md *PCPMetricDesc) Semantics() MetricSemantics { return md.sem }

// Unit returns the unit for PCPMetric
func (md *PCPMetricDesc) Unit() MetricUnit { return md.u }

// Type returns the type for PCPMetric
func (md *PCPMetricDesc) Type() MetricType { return md.t }

// ShortDescription returns the shortdesc value
func (md *PCPMetricDesc) ShortDescription() string { return md.shortDescription.val }

// LongDescription returns the longdesc value
func (md *PCPMetricDesc) LongDescription() string { return md.longDescription.val }

// Description returns the description for PCPMetric
func (md *PCPMetricDesc) Description() string {
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
func newupdateClosure(offset int, buffer bytebuffer.Buffer) updateClosure {
	return func(val interface{}) error {
		if _, isString := val.(string); isString {
			buffer.MustSetPos(offset)
			buffer.MustWrite(make([]byte, StringLength))
		}
		buffer.MustSetPos(offset)
		return buffer.WriteVal(val)
	}
}

///////////////////////////////////////////////////////////////////////////////

// PCPSingletonMetric defines a singleton metric with no instance domain
// only a value and a valueoffset
type PCPSingletonMetric struct {
	sync.RWMutex
	*PCPMetricDesc
	val         interface{}
	valueoffset int
	update      updateClosure
}

// NewPCPSingletonMetric creates a new instance of PCPSingletonMetric
// it takes 2 extra optional strings as short and long description parameters,
// which on not being present are set blank
func NewPCPSingletonMetric(val interface{}, name string, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*PCPSingletonMetric, error) {
	if name == "" {
		return nil, errors.New("Metric name cannot be empty")
	}

	if !t.IsCompatible(val) {
		return nil, fmt.Errorf("type %v is not compatible with value %v", t, val)
	}

	if len(desc) > 2 {
		return nil, errors.New("only 2 optional strings allowed, short and long descriptions")
	}

	shortdesc, longdesc := "", ""

	if len(desc) > 0 {
		shortdesc = desc[0]
	}

	if len(desc) > 1 {
		longdesc = desc[1]
	}

	val = t.resolve(val)

	return &PCPSingletonMetric{
		sync.RWMutex{},
		newPCPMetricDesc(name, t, s, u, shortdesc, longdesc),
		val, 0, nil,
	}, nil
}

// Val returns the current Set value of PCPSingletonMetric
func (m *PCPSingletonMetric) Val() interface{} {
	m.RLock()
	defer m.RUnlock()

	if m.t == StringType {
		return m.val.(*pcpString).val
	}

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
			err := m.update(val)
			if err != nil {
				return err
			}
		}
		m.val = val
	}

	return nil
}

// MustSet is a Set that panics
func (m *PCPSingletonMetric) MustSet(val interface{}) {
	if err := m.Set(val); err != nil {
		panic(err)
	}
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
	*PCPMetricDesc
	indom *PCPInstanceDomain
	vals  map[string]*instanceValue
}

// NewPCPInstanceMetric creates a new instance of PCPSingletonMetric
// it takes 2 extra optional strings as short and long description parameters,
// which on not being present are set blank
func NewPCPInstanceMetric(vals Instances, name string, indom *PCPInstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*PCPInstanceMetric, error) {
	if name == "" {
		return nil, errors.New("Metric name cannot be empty")
	}

	if len(vals) != indom.InstanceCount() {
		return nil, errors.New("values for all instances in the instance domain only should be passed")
	}

	if len(desc) > 2 {
		return nil, errors.New("only 2 optional strings allowed, short and long descriptions")
	}

	shortdesc, longdesc := "", ""

	if len(desc) > 0 {
		shortdesc = desc[0]
	}

	if len(desc) > 1 {
		longdesc = desc[1]
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

		val = t.resolve(val)
		mvals[name] = newinstanceValue(val)
	}

	return &PCPInstanceMetric{
		sync.RWMutex{},
		newPCPMetricDesc(name, t, s, u, shortdesc, longdesc),
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

	ans := m.vals[instance].val

	if m.t == StringType {
		return ans.(*pcpString).val, nil
	}

	return ans, nil
}

// SetInstance sets the value for a particular instance of the metric
func (m *PCPInstanceMetric) SetInstance(instance string, val interface{}) error {
	if !m.t.IsCompatible(val) {
		return errors.New("the value is incompatible with this metrics MetricType")
	}

	if !m.indom.HasInstance(instance) {
		return fmt.Errorf("%v is not an instance of this metric", instance)
	}

	if m.vals[instance].val != val {
		m.Lock()
		defer m.Unlock()

		if m.vals[instance].update != nil {
			err := m.vals[instance].update(val)
			if err != nil {
				return err
			}
		}

		m.vals[instance].val = val
	}

	return nil
}

// MustSetInstance is a SetInstance that panics
func (m *PCPInstanceMetric) MustSetInstance(instance string, val interface{}) {
	if err := m.SetInstance(instance, val); err != nil {
		panic(err)
	}
}
