package pcp

import (
	"fmt"
	"hash/fnv"
)

// MetricType is an enumerated type representing all valid types for a metric
type MetricType int32

// Possible values for a MetricType
const (
	NoSupportType       MetricType = iota
	Int32Type           MetricType = iota
	Uint32Type          MetricType = iota
	Int64Type           MetricType = iota
	Uint64Type          MetricType = iota
	FloatType           MetricType = iota
	DoubleType          MetricType = iota
	StringType          MetricType = iota
	AggregateType       MetricType = iota
	AggregateStaticType MetricType = iota
	EventType           MetricType = iota
	HighresEventType    MetricType = iota
	UnknownType         MetricType = iota
)

func (mt MetricType) String() string {
	switch mt {
	case NoSupportType:
		return "Type: No Support"
	case Int32Type:
		return "Type: Int32"
	case Int64Type:
		return "Type: Int64"
	case Uint32Type:
		return "Type: Uint32"
	case Uint64Type:
		return "Type: Uint64"
	case FloatType:
		return "Type: Float"
	case DoubleType:
		return "Type: Double"
	case StringType:
		return "Type: String"
	case AggregateType:
		return "Type: Aggregate"
	case AggregateStaticType:
		return "Type: Aggregate Static"
	case EventType:
		return "Type: Event"
	case HighresEventType:
		return "Type: Highres Event"
	case UnknownType:
		return "Type: Unknown"
	default:
		return "Type: Invalid"
	}
}

// MetricUnit is an enumerated type representing all possible values for a valid PCP unit
type MetricUnit int32

// SpaceUnit is an enumerated type representing all units for space
type SpaceUnit MetricUnit

// Possible values for SpaceUnit
const (
	ByteUnit     SpaceUnit = iota
	KilobyteUnit SpaceUnit = iota
	MegabyteUnit SpaceUnit = iota
	GigabyteUnit SpaceUnit = iota
	TerabyteUnit SpaceUnit = iota
	PetabyteUnit SpaceUnit = iota
	ExabyteUnit  SpaceUnit = iota
)

func (su SpaceUnit) String() string {
	switch su {
	case ByteUnit:
		return "Unit: Byte"
	case KilobyteUnit:
		return "Unit: Kilobyte"
	case MegabyteUnit:
		return "Unit: Megabyte"
	case GigabyteUnit:
		return "Unit: Gigabyte"
	case TerabyteUnit:
		return "Unit: Terabyte"
	case PetabyteUnit:
		return "Unit: Petabyte"
	case ExabyteUnit:
		return "Unit: Exabyte"
	default:
		return "Unit: Invalid SpaceUnit"
	}
}

// TimeUnit is an enumerated type representing all possible units for representing time
type TimeUnit MetricUnit

// Possible Values for TimeUnit
const (
	NanosecondUnit  TimeUnit = iota
	MicrosecondUnit TimeUnit = iota
	MillisecondUnit TimeUnit = iota
	SecondUnit      TimeUnit = iota
	MinuteUnit      TimeUnit = iota
	HourUnit        TimeUnit = iota
)

func (tu TimeUnit) String() string {
	switch tu {
	case NanosecondUnit:
		return "Unit: Nanosecond"
	case MicrosecondUnit:
		return "Unit: Microsecond"
	case MillisecondUnit:
		return "Unit: Millisecond"
	case SecondUnit:
		return "Unit: Second"
	case MinuteUnit:
		return "Unit: Minute"
	case HourUnit:
		return "Unit: Hour"
	default:
		return "Unit: Invalid TimeUnit"
	}
}

// CountUnit is a type representing a counted quantity
type CountUnit MetricUnit

// OneUnit represents the only CountUnit
const OneUnit CountUnit = iota

func (cu CountUnit) String() string {
	switch cu {
	case OneUnit:
		return "Unit: One"
	default:
		return "Unit: Invalid CounterUnit"
	}
}

// MetricSemantics represents an enumerated type representing the possible
// values for the semantics of a metric
type MetricSemantics int32

// Possible values for MetricSemantics
const (
	NoSemantics       MetricSemantics = iota
	CounterSemantics  MetricSemantics = iota
	InstantSemantics  MetricSemantics = iota
	DiscreteSemantics MetricSemantics = iota
)

func (ms MetricSemantics) String() string {
	switch ms {
	case NoSemantics:
		return "Semantics: None"
	case CounterSemantics:
		return "Semantics: Counter"
	case InstantSemantics:
		return "Semantics: Instant"
	case DiscreteSemantics:
		return "Semantics: Discrete"
	default:
		return "Semantics: Invalid"
	}
}

// Metric defines the general interface a type needs to implement to qualify
// as a valid PCP metric
type Metric interface {
	Val() interface{}           // gets the value of the metric
	Set(interface{}) error      // sets the value of the metric to a value, optionally returns an error on failure
	Type() MetricType           // gets the type of a metric
	Unit() MetricUnit           // gets the unit of a metric
	Semantics() MetricSemantics // gets the semantics for a metric
	Description() string        // gets the description of a metric
}

// generate a unique uint32 hash for a string
// NOTE: make sure this is as fast as possible
func getHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// MetricDesc is a metric metadata wrapper
// each metric type can wrap its metadata by containing a MetricDesc type and only define its own
// specific properties assuming MetricDesc will handle the rest
//
// when writing, this type is supposed to map directly to the pmDesc struct as defined in PCP core
type MetricDesc struct {
	id                                uint32          // unique metric id
	name                              string          // the name
	indom                             InstanceDomain  // the instance domain
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	shortDescription, longDescription string
}

// NewMetricDesc creates a new Metric Description wrapper type
func NewMetricDesc(name string, indom InstanceDomain, t MetricType, semantics MetricSemantics, unit MetricUnit, short, long string) *MetricDesc {
	return &MetricDesc{
		getHash(name), name, indom, t, semantics, unit, short, long,
	}
}

func (md *MetricDesc) String() string {
	return fmt.Sprintf("%s{%v, %v, %v, %v}", md.name, md.indom, md.t, md.sem, md.u)
}
