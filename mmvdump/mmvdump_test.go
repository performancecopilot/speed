package mmvdump

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestMmvDump1(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/test1.mmv")
	if err != nil {
		t.Fatal(err)
	}

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
		t.Errorf("expected number of metrics %d, got %d", 1, len(metrics))
	}

	if len(values) != 1 {
		t.Errorf("expected number of values %d, got %d", 1, len(values))
	}

	if len(instances) != 0 {
		t.Errorf("expected number of instances %d, got %d", 0, len(instances))
	}
}

func TestInputs(t *testing.T) {
	for _, c := range []struct {
		input, output string
	}{
		{"testdata/test1.mmv", "testdata/output1.golden"},
		{"testdata/test2.mmv", "testdata/output2.golden"},
		{"testdata/test3.mmv", "testdata/output3.golden"},
		{"testdata/test4.mmv", "testdata/output4.golden"},
		{"testdata/test5.mmv", "testdata/output5.golden"},
	} {
		data, err := ioutil.ReadFile(c.input)
		if err != nil {
			t.Fatal(err)
		}

		header, tocs, metrics, values, instances, indoms, strings, err := Dump(data)
		if err != nil {
			t.Fatal(err)
		}

		var b = new(bytes.Buffer)
		err = Write(b, header, tocs, metrics, values, instances, indoms, strings)
		if err != nil {
			t.Fatal(err)
		}

		expected, err := ioutil.ReadFile(c.output)
		if err != nil {
			t.Fatal(err)
		}

		actual := b.Bytes()

		if !bytes.Equal(expected, actual) {
			t.Fatalf(`
Failed for input %s,
expected
-------------------------------------------
%s
-------------------------------------------
got
-------------------------------------------
%s
-------------------------------------------

			`, c.input, string(expected), string(actual))
		}
	}
}
