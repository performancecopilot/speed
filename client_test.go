package speed

import (
	"fmt"
	"os"
	"testing"

	"github.com/performancecopilot/speed/mmvdump"
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
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Errorf("cannot create writer, error: %v", err)
	}

	if c.tocCount() != 2 {
		t.Errorf("expected tocCount to be 2, got %v", c.tocCount())
	}

	expectedLength := HeaderLength + 2*TocLength
	if c.Length() != expectedLength {
		t.Errorf("expected Length to be %v, got %v", expectedLength, c.Length())
	}

	m, err := NewPCPSingletonMetric(10, "test", Int32Type, CounterSemantics, OneUnit, "test", "")
	if err != nil {
		t.Error("Cannot create a metric")
	}

	c.MustRegister(m)
	if c.tocCount() != 3 {
		t.Errorf("expected tocCount to be 3, got %v", c.tocCount())
	}

	expectedLength = HeaderLength + 3*TocLength + 1*MetricLength + 1*ValueLength + 1*StringLength
	if c.Length() != expectedLength {
		t.Errorf("expected Length to be %v, got %v", expectedLength, c.Length())
	}

	indom, _ := NewPCPInstanceDomain("testindom", []string{"test"}, "", "")
	c.MustRegisterIndom(indom)

	m2, err := NewPCPInstanceMetric(
		Instances{
			"test": 1,
		},
		"test2", indom, Int32Type, CounterSemantics, OneUnit, "", "",
	)
	if err != nil {
		t.Error("Cannot create a metric")
	}

	err = c.Register(m2)
	if err != nil {
		t.Error("Cannot register m2")
	}

	if c.tocCount() != 5 {
		t.Errorf("expected tocCount to be 5, got %v", c.tocCount())
	}
}

func TestMapping(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	_, err = c.RegisterString("test.1", 2, CounterSemantics, Int32Type, OneUnit)
	if err != nil {
		t.Error("Cannot Register")
	}

	c.MustStart()
	loc, _ := mmvFileLocation("test")
	if _, err = os.Stat(loc); err != nil {
		t.Error("expected a MMV file to be created on startup")
	}

	_, err = c.RegisterString("test.2", 2, CounterSemantics, Int32Type, OneUnit)
	if err == nil {
		t.Error("expected registration to fail when a mapping is active")
	}

	EraseFileOnStop = true
	err = c.Stop()
	if err != nil {
		t.Error("Cannot stop a mapping")
	}

	if _, err = os.Stat(loc); err == nil {
		t.Error("expected the MMV file be deleted after stopping")
	}

	EraseFileOnStop = false
}

func TestWritingSingletonMetric(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Error(err)
		return
	}

	met, err := NewPCPSingletonMetric(10, "test.1", Int32Type, CounterSemantics, OneUnit, "test", "")
	if err != nil {
		t.Error(err)
		return
	}
	c.MustRegister(met)

	c.MustStart()
	defer c.MustStop()

	h, toc, m, v, i, ind, s, err := mmvdump.Dump(c.buffer.Bytes())

	if int(h.Toc) != len(toc) {
		t.Errorf("expected the number of tocs specified in the header and the number of tocs in the toc array to be the same, h.Toc = %d, len(tocs) = %d", h.Toc, len(toc))
	}

	if h.Toc != 3 {
		t.Errorf("expected client to write %d tocs, written %d", 3, h.Toc)
	}

	if h.Flag != int32(ProcessFlag) {
		t.Errorf("expected client to write a ProcessFlag, writing %v", MMVFlag(h.Flag))
	}

	if len(m) != 1 {
		t.Errorf("expected to write %d metrics, writing %d", 1, len(m))
	}

	// metrics

	off := c.r.metricsoffset
	metric, ok := m[uint64(off)]
	if !ok {
		t.Errorf("expected the metric to exist at offset %v", off)
	}

	if metric.Indom != mmvdump.NoIndom {
		t.Error("expected indom to be null")
	}

	if int32(metric.Sem) != int32(CounterSemantics) {
		t.Errorf("expected semantics to be %v, got %v", CounterSemantics, MetricSemantics(metric.Sem))
	}

	if int32(metric.Typ) != int32(Int32Type) {
		t.Errorf("expected type to be %v, got %v", Int32Type, MetricType(metric.Typ))
	}

	if int32(metric.Unit) != int32(OneUnit) {
		t.Errorf("expected unit to be %v, got %v", OneUnit, metric.Unit)
	}

	if metric.Shorttext == 0 {
		t.Error("expected shorttext offset to not be 0")
	}

	if metric.Shorttext != uint64(c.r.stringsoffset) {
		t.Errorf("expected shorttext offset to be %v", c.r.stringsoffset)
	}

	if metric.Longtext != 0 {
		t.Errorf("expected longtext offset to be 0")
	}

	// values

	mv, ok := v[uint64(c.r.valuesoffset)]
	if !ok {
		t.Errorf("expected a value to be written at offset %v", c.r.valuesoffset)
	}

	if mv.Metric != uint64(off) {
		t.Errorf("expected value's metric to be at %v", off)
	}

	if av, err := mmvdump.FixedVal(mv.Val, mmvdump.Int32Type); err != nil || av.(int32) != 10 {
		t.Errorf("expected the value to be %v, got %v", 10, av)
	}

	if mv.Instance != 0 {
		t.Errorf("expected value instance to be 0")
	}

	// strings

	if len(s) != 1 {
		t.Error("expected one string")
	}

	str, ok := s[uint64(c.r.stringsoffset)]
	if !ok {
		t.Errorf("expected a string to be written at offset %v", c.r.stringsoffset)
	}

	sv := string(str.Payload[:4])
	if sv != "test" {
		t.Errorf("expected payload to be %v, got %v", "test", sv)
	}

	// instances

	if len(i) != 0 {
		t.Error("expected no instances when writing a singleton metric")
	}

	// indoms

	if len(ind) != 0 {
		t.Error("expected no indoms when writing a singleton metric")
	}
}
