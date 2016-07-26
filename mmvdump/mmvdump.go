// Package mmvdump implements a go port of the C mmvdump utility included in PCP Core
//
// https://github.com/performancecopilot/pcp/blob/master/src/pmdas/mmv/mmvdump.c
//
// It has been written for maximum portability with the C equivalent, without having to use cgo or any other ninja stuff
//
// the main difference is that the reader is separate from the cli with the reading primarily implemented in mmvdump.go while the cli is implemented in cmd/mmvdump
//
// the cli application is completely go gettable and outputs the same things, in mostly the same way as the C cli app, to try it out,
//
// ```
// go get github.com/performancecopilot/speed/mmvdump/cmd/mmvdump
// ```
package mmvdump

import (
	"errors"
	"fmt"
	"math"
	"unsafe"
)

func readHeader(data []byte) (*Header, error) {
	if uint64(len(data)) < HeaderLength {
		return nil, errors.New("file too small to contain a valid Header")
	}

	header := (*Header)(unsafe.Pointer(&data[0]))

	if m := header.Magic[:3]; string(m) != "MMV" {
		return nil, fmt.Errorf("Bad Magic: %v", string(m))
	}

	if header.G1 != header.G2 {
		return nil, fmt.Errorf("Mismatched version numbers, %v and %v", header.G1, header.G2)
	}

	return header, nil
}

func readToc(data []byte, offset uint64) (*Toc, error) {
	if uint64(len(data)) < offset+TocLength {
		return nil, errors.New("Incomplete/Partially Written TOC")
	}

	return (*Toc)(unsafe.Pointer(&data[offset])), nil
}

func readInstance(data []byte, offset uint64) (*Instance, error) {
	if uint64(len(data)) < offset+InstanceLength {
		return nil, errors.New("Incomplete/Partially Written Instance")
	}

	return (*Instance)(unsafe.Pointer(&data[offset])), nil
}

func readInstanceDomain(data []byte, offset uint64) (*InstanceDomain, error) {
	if uint64(len(data)) < offset+InstanceDomainLength {
		return nil, errors.New("Incomplete/Partially Written InstanceDomain")
	}

	return (*InstanceDomain)(unsafe.Pointer(&data[offset])), nil
}

func readMetric(data []byte, offset uint64) (*Metric, error) {
	if uint64(len(data)) < offset+MetricLength {
		return nil, errors.New("Incomplete/Partially Written Metric")
	}

	return (*Metric)(unsafe.Pointer(&data[offset])), nil
}

func readValue(data []byte, offset uint64) (*Value, error) {
	if uint64(len(data)) < offset+ValueLength {
		return nil, errors.New("Incomplete/Partially Written Value")
	}

	return (*Value)(unsafe.Pointer(&data[offset])), nil
}

func readString(data []byte, offset uint64) (*String, error) {
	if uint64(len(data)) < offset+StringLength {
		return nil, errors.New("Incomplete/Partially Written String")
	}

	return (*String)(unsafe.Pointer(&data[offset])), nil
}

func readTocs(data []byte, count int32) ([]*Toc, error) {
	tocs := make([]*Toc, count)

	for i := int32(0); i < count; i++ {
		t, err := readToc(data, HeaderLength+uint64(i)*TocLength)
		if err != nil {
			return nil, err
		}
		tocs[i] = t
	}

	return tocs, nil
}

func readInstances(data []byte, offset uint64, count int32) (map[uint64]*Instance, error) {
	instances := make(map[uint64]*Instance)
	for i := int32(0); i < count; i, offset = i+1, offset+InstanceLength {
		instance, err := readInstance(data, offset)
		if err != nil {
			return nil, err
		}
		instances[offset] = instance
	}

	return instances, nil
}

func readInstanceDomains(data []byte, offset uint64, count int32) (map[uint64]*InstanceDomain, error) {
	indoms := make(map[uint64]*InstanceDomain)
	for i := int32(0); i < count; i, offset = i+1, offset+InstanceDomainLength {
		indom, err := readInstanceDomain(data, offset)
		if err != nil {
			return nil, err
		}
		indoms[offset] = indom
	}

	return indoms, nil
}

func readMetrics(data []byte, offset uint64, count int32) (map[uint64]*Metric, error) {
	metrics := make(map[uint64]*Metric)
	for i := int32(0); i < count; i, offset = i+1, offset+MetricLength {
		metric, err := readMetric(data, offset)
		if err != nil {
			return nil, err
		}
		metrics[offset] = metric
	}

	return metrics, nil
}

func readValues(data []byte, offset uint64, count int32) (map[uint64]*Value, error) {
	values := make(map[uint64]*Value)
	for i := int32(0); i < count; i, offset = i+1, offset+ValueLength {
		value, err := readValue(data, offset)
		if err != nil {
			return nil, err
		}
		values[offset] = value
	}

	return values, nil
}

func readStrings(data []byte, offset uint64, count int32) (map[uint64]*String, error) {
	strings := make(map[uint64]*String)
	for i := int32(0); i < count; i, offset = i+1, offset+StringLength {
		str, err := readString(data, offset)
		if err != nil {
			return nil, err
		}
		strings[offset] = str
	}

	return strings, nil
}

// Dump creates a data dump from the passed data
func Dump(data []byte) (
	h *Header,
	tocs []*Toc,
	metrics map[uint64]*Metric,
	values map[uint64]*Value,
	instances map[uint64]*Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
	err error,
) {
	h, err = readHeader(data)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	tocs, err = readTocs(data, h.Toc)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	for _, toc := range tocs {
		switch toc.Type {
		case TocInstances:
			instances, err = readInstances(data, toc.Offset, toc.Count)
		case TocIndoms:
			indoms, err = readInstanceDomains(data, toc.Offset, toc.Count)
		case TocMetrics:
			metrics, err = readMetrics(data, toc.Offset, toc.Count)
		case TocValues:
			values, err = readValues(data, toc.Offset, toc.Count)
		case TocStrings:
			strings, err = readStrings(data, toc.Offset, toc.Count)
		}
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
	}

	return
}

// FixedVal will infer a fixed size value from the passed data
func FixedVal(data uint64, t Type) (interface{}, error) {
	switch t {
	case Int32Type:
		return int32(data), nil
	case Uint32Type:
		return uint32(data), nil
	case Int64Type:
		return int64(data), nil
	case Uint64Type:
		return data, nil
	case FloatType:
		return math.Float32frombits(uint32(data)), nil
	case DoubleType:
		return math.Float64frombits(data), nil
	}

	return nil, errors.New("invalid type")
}

// StringVal will infer the string corresponding to an address
func StringVal(data uint64, strings map[uint64]*String) (string, error) {
	str, ok := strings[data]
	if !ok {
		return "", errors.New("invalid string address")
	}

	return string(str.Payload[:]), nil
}
