package mmvdump

// MMVVersion is the current mmv format version
const MMVVersion = 1

const (
	// NameMax is the maximum allowed length of a name
	NameMax = 64

	// StringMax is the maximum allowed length of a string
	StringMax = 256

	// NoIndom is a constant used to indicate abscence of an indom from a metric
	NoIndom = -1
)

// Header describes the data in a MMV header
type Header struct {
	Magic            [4]byte
	Version          int32
	G1, G2           uint64
	Toc              int32
	Flag             int32
	Process, Cluster int32
}

// TocType is an enumerated type with different types as values
type TocType int32

// Values for TocType
const (
	TocIndoms TocType = iota + 1
	TocInstances
	TocMetrics
	TocValues
	TocStrings
)

//go:generate stringer --type=TocType

// Toc defines the contents in a valid TOC
type Toc struct {
	Type   TocType
	Count  int32
	Offset uint64
}

// Instance is the base type for all instances
type Instance interface {
	Indom() uint64
	Internal() int32
	Padding() uint32
}

// InstanceBase defines the common contents in a valid instance
type InstanceBase struct {
	indom    uint64
	padding  uint32
	internal int32
}

// Indom returns the indom offset
func (i InstanceBase) Indom() uint64 { return i.indom }

// Internal returns the internal id
func (i InstanceBase) Internal() int32 { return i.internal }

// Padding returns the padding value
func (i InstanceBase) Padding() uint32 { return i.padding }

// Instance1 defines the contents in a valid mmv1 instance
type Instance1 struct {
	InstanceBase
	External [NameMax]byte
}

// Instance2 defines the contents in a valid mmv2 instance
type Instance2 struct {
	InstanceBase
	External uint64
}

// InstanceDomain defines the contents in a valid instance domain
type InstanceDomain struct {
	Serial, Count               uint32
	Offset, Shorttext, Longtext uint64
}

// Metric is the base type for all metrics
type Metric interface {
	Item() uint32
	Typ() Type
	Sem() Semantics
	Unit() Unit
	Indom() int32
	Padding() uint32
	ShortText() uint64
	LongText() uint64
}

// MetricBase defines the common contents in a valid Metric
type MetricBase struct {
	item                uint32
	typ                 Type
	sem                 Semantics
	unit                Unit
	indom               int32
	padding             uint32
	shorttext, longtext uint64
}

// Item returns the item id
func (m MetricBase) Item() uint32 { return m.item }

// Typ returns the type
func (m MetricBase) Typ() Type { return m.typ }

// Sem returns the semantics
func (m MetricBase) Sem() Semantics { return m.sem }

// Unit returns the unit
func (m MetricBase) Unit() Unit { return m.unit }

// Indom returns the indom id
func (m MetricBase) Indom() int32 { return m.indom }

// Padding returns the padding value
func (m MetricBase) Padding() uint32 { return m.padding }

// ShortText returns the shorttext offset
func (m MetricBase) ShortText() uint64 { return m.shorttext }

// LongText returns the longtext offset
func (m MetricBase) LongText() uint64 { return m.longtext }

// Metric1 defines the contents in a valid Metric
type Metric1 struct {
	Name [NameMax]byte
	MetricBase
}

// Metric2 defines the contents in a valid Metric
type Metric2 struct {
	Name uint64
	MetricBase
}

// Value defines the contents in a PCP Value
type Value struct {
	// uint64 is a holder type here, while printing it is expected that
	// the user will infer the value using the Val functions
	Val uint64

	Extra    int64
	Metric   uint64
	Instance uint64
}

// String wraps the payload for a PCP String
type String struct {
	Payload [StringMax]byte
}

// Type is an enumerated type representing all valid types for a metric
type Type int32

// Possible values for a Type
const (
	NoSupportType Type = iota - 1
	Int32Type
	Uint32Type
	Int64Type
	Uint64Type
	FloatType
	DoubleType
	StringType
	UnknownType Type = 255
)

//go:generate stringer --type=Type

// Unit is an enumerated type with all possible units as values
type Unit uint32

// Values for Space Units
const (
	ByteUnit Unit = 1<<28 | iota<<16
	KilobyteUnit
	MegabyteUnit
	GigabyteUnit
	TerabyteUnit
	PetabyteUnit
	ExabyteUnit
)

// Values for Time Units
const (
	NanosecondUnit Unit = 1<<24 | iota<<12
	MicrosecondUnit
	MillisecondUnit
	SecondUnit
	MinuteUnit
	HourUnit
)

// Values for Count Units
const (
	OneUnit Unit = 1<<20 | iota<<8
)

//go:generate stringer --type=Unit

// Semantics represents an enumerated type representing all possible semantics of a metric
type Semantics int32

// Values for Semantics
const (
	NoSemantics Semantics = 0
	CounterSemantics
	_
	InstantSemantics
	DiscreteSemantics
)

//go:generate stringer -type=Semantics

// Byte Lengths for Different Components
const (
	HeaderLength         uint64 = 40
	TocLength            uint64 = 16
	Metric1Length        uint64 = 104
	Metric2Length        uint64 = 48
	ValueLength          uint64 = 32
	Instance1Length      uint64 = 80
	Instance2Length      uint64 = 24
	InstanceDomainLength uint64 = 32
	StringLength         uint64 = 256
)
