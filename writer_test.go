package speed

import (
	"fmt"
	"os"
	"testing"
)

func TestMmvFileLocation(t *testing.T) {
	l, present := config["PCP_TMP_DIR"]

	if present {
		loc, _ := mmvFileLocation("test")
		expected := fmt.Sprintf("%v%cmmv%c%v", l, os.PathSeparator, os.PathSeparator, "test")
		if loc != expected {
			t.Errorf("location not expected value, expected %v, got %v", expected, loc)
		}
	}

	delete(config, "PCP_TMP_DIR")
	loc, _ := mmvFileLocation("test")
	expected := fmt.Sprintf("%v%cmmv%c%v", os.TempDir(), os.PathSeparator, os.PathSeparator, "test")
	if loc != expected {
		t.Errorf("location not expected value, expected %v, got %v", expected, loc)
	}

	if present {
		config["PCP_TMP_DIR"] = l
	}

	loc, err := mmvFileLocation(fmt.Sprintf("%v%c", "test", os.PathSeparator))
	if err == nil {
		t.Errorf("expected error, instead got path %v", loc)
	}
}

func TestTocCountAndLength(t *testing.T) {
	w, err := NewPCPWriter("test", ProcessFlag)
	if err != nil {
		t.Errorf("cannot create writer, error: %v", err)
	}

	if w.tocCount() != 2 {
		t.Errorf("expected tocCount to be 2, got %v", w.tocCount())
	}

	expectedLength := HeaderLength + 2*TocLength
	if w.Length() != expectedLength {
		t.Errorf("expected Length to be %v, got %v", expectedLength, w.Length())
	}

	m, err := NewPCPSingletonMetric(10, "test", Int32Type, CounterSemantics, OneUnit, "test", "")
	if err != nil {
		t.Error("Cannot create a metric")
	}

	w.MustRegister(m)
	if w.tocCount() != 3 {
		t.Errorf("expected tocCount to be 3, got %v", w.tocCount())
	}

	expectedLength = HeaderLength + 3*TocLength + 1*MetricLength + 1*ValueLength + 1*StringLength
	if w.Length() != expectedLength {
		t.Errorf("expected Length to be %v, got %v", expectedLength, w.Length())
	}

	indom, _ := NewPCPInstanceDomain("testindom", []string{"test"}, "", "")
	w.MustRegisterIndom(indom)

	m2, err := NewPCPInstanceMetric(
		Instances{
			"test": 1,
		},
		"test2", indom, Int32Type, CounterSemantics, OneUnit, "", "",
	)
	if err != nil {
		t.Error("Cannot create a metric")
	}

	err = w.Register(m2)
	if err != nil {
		t.Error("Cannot register m2")
	}

	if w.tocCount() != 5 {
		t.Errorf("expected tocCount to be 5, got %v", w.tocCount())
	}
}

func TestMapping(t *testing.T) {
	w, err := NewPCPWriter("test", ProcessFlag)
	_, err = w.RegisterString("test.1", 2, CounterSemantics, Int32Type, OneUnit)
	if err != nil {
		t.Error("Cannot Register")
	}

	w.MustStart()
	loc, _ := mmvFileLocation("test")
	if _, err = os.Stat(loc); err != nil {
		t.Error("expected a MMV file to be created on startup")
	}

	_, err = w.RegisterString("test.2", 2, CounterSemantics, Int32Type, OneUnit)
	if err == nil {
		t.Error("expected registration to fail when a mapping is active")
	}

	EraseFileOnStop = true
	err = w.Stop()
	if err != nil {
		t.Error("Cannot stop a mapping")
	}

	if _, err = os.Stat(loc); err == nil {
		t.Error("expected the MMV file be deleted after stopping")
	}

	EraseFileOnStop = false
}
