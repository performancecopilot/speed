package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/performancecopilot/speed/mmvdump"
)

var (
	header    *mmvdump.Header
	tocs      []*mmvdump.Toc
	metrics   map[uint64]*mmvdump.Metric
	values    map[uint64]*mmvdump.Value
	instances map[uint64]*mmvdump.Instance
	indoms    map[uint64]*mmvdump.InstanceDomain
	strings   map[uint64]*mmvdump.String
)

func printMetric(offset uint64) {
	m := metrics[offset]

	fmt.Printf("\t[%v/%v] %v\n", m.Item, offset, string(m.Name[:]))
	fmt.Printf("\t\ttype=%v (0x%x), sem=%v (0x%x), pad=0x%x\n", m.Typ, int(m.Typ), m.Sem, int(m.Sem), m.Padding)
	fmt.Printf("\t\tunits=%v\n", m.Unit)

	if m.Indom == mmvdump.NoIndom {
		fmt.Printf("\t\t(no indom)\n")
	} else {
		fmt.Printf("\t\tindom=%d\n", m.Indom)
	}

	if m.Shorttext == 0 {
		fmt.Printf("\t\t(no shorttext)\n")
	} else {
		fmt.Printf("\t\tshorttext=%v\n", string(strings[m.Shorttext].Payload[:]))
	}

	if m.Longtext == 0 {
		fmt.Printf("\t\t(no longtext)\n")
	} else {
		fmt.Printf("\t\tlongtext=%v\n", string(strings[m.Longtext].Payload[:]))
	}
}

func printValue(offset uint64) {
	v := values[offset]
	m := metrics[v.Metric]

	fmt.Printf("\t[%v/%v] %v", m.Item, offset, string(m.Name[:]))

	var (
		a   interface{}
		err error
	)

	if m.Typ != mmvdump.StringType {
		a, err = mmvdump.FixedVal(v.Val, m.Typ)
	} else {
		a, err = mmvdump.StringVal(v.Val, strings)
	}

	if m.Indom != mmvdump.NoIndom {
		i := instances[v.Instance]
		fmt.Printf("[%d or \"%s\"]", i.Internal, string(i.External[:]))
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf(" = %v\n", a)
}

func printString(offset uint64) {
	fmt.Printf("\t[%v] %v\n", offset, string(strings[offset].Payload[:]))
}

func data(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	len := fi.Size()
	data := make([]byte, len)

	_, err = f.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		panic("Usage: mmvdump <file>")
	}

	file := flag.Arg(0)
	d, err := data(file)

	header, tocs, metrics, values, instances, indoms, strings, err = mmvdump.Dump(d)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`
File      = %v
Version   = %v
Generated = %v
Toc Count = %v
Cluster   = %v
Process   = %v
Flags     = 0x%x

`, file, header.Version, header.G1, header.Toc, header.Cluster, header.Process, int(header.Flag))

	toff := mmvdump.HeaderLength
	var (
		itemtype  string
		itemsize  uint64
		printItem func(uint64)
	)

	for ti, toc := range tocs {
		switch toc.Type {
		case mmvdump.TocMetrics:
			itemtype = "metric"
			itemsize = mmvdump.MetricLength
			printItem = printMetric

		case mmvdump.TocValues:
			itemtype = "values"
			itemsize = mmvdump.ValueLength
			printItem = printValue

		case mmvdump.TocStrings:
			itemtype = "strings"
			itemsize = mmvdump.StringLength
			printItem = printString
		}

		fmt.Printf("TOC[%v], offset: %v, %v offset: %v (%v entries)\n", ti, toff, itemtype, toc.Offset, toc.Count)
		for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+itemsize {
			printItem(offset)
		}
		fmt.Println()

		toff += mmvdump.TocLength
	}
}
