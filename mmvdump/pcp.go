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

// Instance defines the contents in a valid instance
type Instance struct {
	Indom    uint64
	Padding  uint32
	Internal int32
	External [NameMax]byte
}

// InstanceDomain defines the contents in a valid instance domain
type InstanceDomain struct {
	Serial, Count               uint32
	Offset, Shorttext, Longtext uint64
}

// Metric defines the contents in a valid Metric
type Metric struct {
	Name                [NameMax]byte
	Item                uint32
	Typ                 Type
	Sem                 Semantics
	Unit                Unit
	Indom               int32
	Padding             uint32
	Shorttext, Longtext uint64
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
	NoSemantics       Semantics = 0
	CounterSemantics  Semantics = 1
	InstantSemantics  Semantics = 3
	DiscreteSemantics Semantics = 4
)

//go:generate stringer -type=Semantics

// Byte Lengths for Different Components
const (
	HeaderLength         uint64 = 40
	TocLength                   = 16
	MetricLength                = 104
	ValueLength                 = 32
	InstanceLength              = 80
	InstanceDomainLength        = 32
	StringLength                = 256
)
