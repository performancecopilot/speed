// Package mmvdump implements the go port of the mmvdump utility inside pcp core
// written in C
package mmvdump

import (
	"errors"
	"fmt"
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

func Dump(data []byte) (
	h *Header,
	ts map[int32]*Toc,
	metrics map[uint32]*Metric,
	values []*Value,
	instances map[int32]*Instance,
	indoms map[uint32]*InstanceDomain,
	strings []*String,
	err error,
) {
	h, err = readHeader(data)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	ts = make(map[int32]*Toc)
	for i := 0; i < int(h.Toc); i++ {
		t, err := readToc(data, HeaderLength+uint64(i)*TocLength)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		ts[int32(t.Type)] = t
	}

	for _, toc := range ts {
		switch toc.Type {
		case TocInstances:
			instances = make(map[int32]*Instance)
			for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+InstanceLength {
				instance, err := readInstance(data, offset)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
				instances[instance.Internal] = instance
			}
		case TocIndoms:
			indoms := make(map[uint32]*InstanceDomain)
			for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+InstanceDomainLength {
				indom, err := readInstanceDomain(data, offset)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
				indoms[indom.Serial] = indom
			}
		case TocMetrics:
			metrics = make(map[uint32]*Metric)
			for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+MetricLength {
				metric, err := readMetric(data, offset)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
				metrics[metric.Item] = metric
			}
		case TocValues:
			values = make([]*Value, toc.Count)
			for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+ValueLength {
				value, err := readValue(data, offset)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
				values[i] = value
			}
		case TocStrings:
			strings = make([]*String, toc.Count)
			for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+StringLength {
				str, err := readString(data, offset)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, nil, err
				}
				strings[i] = str
			}
		}
	}

	return
}
