package mmvdump

import (
	"os"
	"testing"
)

func data(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	s, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}

	data := make([]byte, s.Size())
	n, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	if int64(n) != s.Size() {
		panic("Could not read complete file" + filename + " into memory")
	}

	return data
}

func TestMmvDump1(t *testing.T) {
	d := data("testdata/test1.mmv")

	h, tocs, metrics, values, instances, indoms, strings, err := Dump(d)
	if err != nil {
		t.Error(err)
		return
	}

	if h.G1 != h.G2 {
		t.Error("Invalid Header")
	}

	if len(tocs) != 3 {
		t.Errorf("expected number of tocs %d, got %d", 3, len(tocs))
	}

	if len(indoms) != 0 {
		t.Errorf("expected number of indoms %d, got %d", 0, len(indoms))
	}

	if len(strings) != 2 {
		t.Errorf("expected number of strings %d, got %d", 2, len(strings))
	}

	if len(metrics) != 1 {
		t.Errorf("expected number of strings %d, got %d", 1, len(metrics))
	}

	if len(values) != 1 {
		t.Errorf("expected number of strings %d, got %d", 1, len(values))
	}

	if len(instances) != 0 {
		t.Errorf("expected number of strings %d, got %d", 0, len(instances))
	}
}
