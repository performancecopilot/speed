package speed

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/codahale/hdrhistogram"
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
	c, err := NewPCPClient("test")
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

	m, err := NewPCPSingletonMetric(10, "test", Int32Type, CounterSemantics, OneUnit, "test")
	if err != nil {
		t.Error("Cannot create a metric")
	}

	c.MustRegister(m)
	if c.tocCount() != 3 {
		t.Errorf("expected tocCount to be 3, got %v", c.tocCount())
	}

	var MetricLength = Metric1Length
	if c.r.version2 {
		MetricLength = Metric2Length
	}

	expectedLength = HeaderLength + 3*TocLength + 1*MetricLength + 1*ValueLength + 1*StringLength
	if c.Length() != expectedLength {
		t.Errorf("expected Length to be %v, got %v", expectedLength, c.Length())
	}

	indom, _ := NewPCPInstanceDomain("testindom", []string{"test"})
	c.MustRegisterIndom(indom)

	m2, err := NewPCPInstanceMetric(
		Instances{
			"test": 1,
		},
		"test2", indom, Int32Type, CounterSemantics, OneUnit,
	)
	if err != nil {
		t.Error("Cannot create a metric")
	}

	c.MustRegister(m2)

	if c.tocCount() != 5 {
		t.Errorf("expected tocCount to be 5, got %v", c.tocCount())
	}
}

func TestMapping(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Fatal("Cannot create client")
	}

	_, err = c.RegisterString("test.1", 2, Int32Type, CounterSemantics, OneUnit)
	if err != nil {
		t.Error("Cannot Register")
	}

	c.MustStart()
	loc, _ := mmvFileLocation("test")
	if _, err = os.Stat(loc); err != nil {
		t.Error("expected a MMV file to be created on startup")
	}

	_, err = c.RegisterString("test.2", 2, Int32Type, CounterSemantics, OneUnit)
	if err == nil {
		t.Error("expected registration to fail when a mapping is active")
	}

	EraseFileOnStop = true
	err = c.Stop()
	if err != nil {
		t.Errorf("Cannot stop a mapping, error: %v", err)
	}

	if _, err = os.Stat(loc); err == nil {
		t.Error("expected the MMV file be deleted after stopping")
	}

	EraseFileOnStop = false
}

func findMetric(metric Metric, metrics map[uint64]mmvdump.Metric) (uint64, mmvdump.Metric) {
	for off, m := range metrics {
		if uint32(m.Item()) == metric.ID() {
			return off, m
		}
	}

	return 0, nil
}

func findSingletonValue(off uint64, values map[uint64]*mmvdump.Value) (uint64, *mmvdump.Value) {
	for voff, v := range values {
		if uint64(v.Metric) == off {
			return voff, v
		}
	}

	return 0, nil
}

func findInstanceValue(off, ioff uint64, values map[uint64]*mmvdump.Value) (uint64, *mmvdump.Value) {
	for voff, v := range values {
		if uint64(v.Metric) == off && uint64(v.Instance) == ioff {
			return voff, v
		}
	}

	return 0, nil
}

func findInstanceDomain(indom *PCPInstanceDomain, indoms map[uint64]*mmvdump.InstanceDomain) (uint64, *mmvdump.InstanceDomain) {
	for off, in := range indoms {
		if in.Serial == indom.id {
			return off, in
		}
	}

	return 0, nil
}

func matchString(s string, str *mmvdump.String, t *testing.T) {
	sv := string(str.Payload[:len(s)])
	if sv != s {
		t.Errorf("expected %v, got %v", s, sv)
	}
}

func matchName(n string, name [64]byte, t *testing.T) {
	if s := name[:len(n)]; n != string(s) {
		t.Errorf("expected name to be %v, got %v", n, s)
	}
}

func matchMetricDesc(desc *pcpMetricDesc, metric mmvdump.Metric, strings map[uint64]*mmvdump.String, t *testing.T) {
	switch m := metric.(type) {
	case *mmvdump.Metric1:
		matchName(desc.name, m.Name, t)
	case *mmvdump.Metric2:
		matchString(desc.name, strings[m.Name], t)
	}

	if int32(metric.Sem()) != int32(desc.sem) {
		t.Errorf("expected semantics to be %v, got %v", desc.sem, MetricSemantics(metric.Sem()))
	}

	if int32(metric.Typ()) != int32(desc.t) {
		t.Errorf("expected type to be %v, got %v", desc.t, MetricType(metric.Typ()))
	}

	if int32(metric.Unit()) != int32(desc.u.PMAPI()) {
		t.Errorf("expected unit to be %v, got %v", desc.u, metric.Unit())
	}

	if metric.ShortText() != 0 {
		matchString(desc.shortDescription, strings[metric.ShortText()], t)
	} else if desc.shortDescription != "" {
		t.Error("expected short description to be \"\"")
	}

	if metric.LongText() != 0 {
		matchString(desc.longDescription, strings[metric.LongText()], t)
	} else if desc.longDescription != "" {
		t.Error("expected long description to be \"\"")
	}
}

func matchSingletonMetric(m *pcpSingletonMetric, metric mmvdump.Metric, strings map[uint64]*mmvdump.String, t *testing.T) {
	if metric.Indom() != mmvdump.NoIndom {
		t.Error("expected indom to be null")
	}

	matchMetricDesc(m.pcpMetricDesc, metric, strings, t)
}

func matchSingletonValue(m *pcpSingletonMetric, value *mmvdump.Value, metrics map[uint64]mmvdump.Metric, strings map[uint64]*mmvdump.String, t *testing.T) {
	off, met := findMetric(m, metrics)
	if met == nil {
		t.Errorf("expected to find metric with name %v", m.Name())
		return
	}

	if value.Metric != off {
		t.Errorf("expected value's metric to be at %v", off)
	}

	if m.t == StringType {
		matchString(m.val.(string), strings[uint64(value.Extra)], t)
	} else {
		if av, err := mmvdump.FixedVal(value.Val, mmvdump.Type(m.t)); err != nil || av != m.val {
			t.Errorf("expected the value to be %v, got %v", m.val, av)
		}
	}

	if value.Instance != 0 {
		t.Errorf("expected value instance to be 0")
	}
}

func matchInstanceMetric(m *pcpInstanceMetric, met mmvdump.Metric, strings map[uint64]*mmvdump.String, t *testing.T) {
	if uint32(met.Indom()) != m.indom.id {
		t.Errorf("expected indom id to be %d, got %d", m.indom.id, met.Indom())
	}

	matchMetricDesc(m.pcpMetricDesc, met, strings, t)
}

func matchInstanceValue(v *mmvdump.Value, i *instanceValue, ins string, met *pcpInstanceMetric, metrics map[uint64]mmvdump.Metric, strings map[uint64]*mmvdump.String, t *testing.T) {
	if v.Instance == 0 {
		t.Errorf("expected instance offset to not be 0")
	}

	if met.indom == nil {
		t.Errorf("expected indom to be non nil")
	} else if in := met.indom.instances[ins]; in == nil {
		t.Errorf("expected the instance domain to have an instance %v", ins)
	} else if in.offset != int(v.Instance) {
		t.Errorf("expected the value's instance to be at offset %v, found at %v", in.offset, v.Instance)
	}

	if met.t == StringType {
		matchString(i.val.(string), strings[uint64(v.Extra)], t)
	} else {
		if av, err := mmvdump.FixedVal(v.Val, mmvdump.Type(met.t)); err != nil || av != i.val {
			t.Errorf("expected the value to be %v, got %v", i.val, av)
		}
	}
}

func matchSingletonMetricAndValue(met *pcpSingletonMetric, metrics map[uint64]mmvdump.Metric, values map[uint64]*mmvdump.Value, strings map[uint64]*mmvdump.String, t *testing.T) {
	off, metric := findMetric(met, metrics)

	if metric == nil {
		t.Errorf("expected a metric of name %v", met.name)
		return
	}

	matchSingletonMetric(met, metric, strings, t)

	_, mv := findSingletonValue(off, values)
	if mv == nil {
		t.Errorf("expected a value for metric %v", met.name)
	} else {
		matchSingletonValue(met, mv, metrics, strings, t)
	}
}

func matchInstanceMetricAndValues(met *pcpInstanceMetric, metrics map[uint64]mmvdump.Metric, values map[uint64]*mmvdump.Value, instances map[uint64]mmvdump.Instance, strings map[uint64]*mmvdump.String, t *testing.T) {
	_, metric := findMetric(met, metrics)
	if metric == nil {
		t.Errorf("expected a metric of name %v", met.name)
		return
	}

	matchInstanceMetric(met, metric, strings, t)

	for n, i := range met.indom.instances {
		off, _ := findMetric(met, metrics)
		_, mv := findInstanceValue(off, uint64(i.offset), values)

		if mv == nil {
			t.Errorf("expected a value at offset %v", i.offset)
		} else {
			matchInstanceValue(mv, met.vals[n], n, met, metrics, strings, t)
		}
	}
}

func matchMetricAndValue(m Metric, metrics map[uint64]mmvdump.Metric, values map[uint64]*mmvdump.Value, instances map[uint64]mmvdump.Instance, strings map[uint64]*mmvdump.String, c *PCPClient, t *testing.T) {
	switch met := m.(type) {
	case *PCPSingletonMetric:
		matchSingletonMetricAndValue(met.pcpSingletonMetric, metrics, values, strings, t)
	case *PCPInstanceMetric:
		matchInstanceMetricAndValues(met.pcpInstanceMetric, metrics, values, instances, strings, t)
	case *PCPCounter:
		matchSingletonMetricAndValue(met.pcpSingletonMetric, metrics, values, strings, t)
	case *PCPGauge:
		matchSingletonMetricAndValue(met.pcpSingletonMetric, metrics, values, strings, t)
	case *PCPTimer:
		matchSingletonMetricAndValue(met.pcpSingletonMetric, metrics, values, strings, t)
	case *PCPCounterVector:
		matchInstanceMetricAndValues(met.pcpInstanceMetric, metrics, values, instances, strings, t)
	case *PCPGaugeVector:
		matchInstanceMetricAndValues(met.pcpInstanceMetric, metrics, values, instances, strings, t)
	case *PCPHistogram:
		matchInstanceMetricAndValues(met.pcpInstanceMetric, metrics, values, instances, strings, t)
	}
}

func matchMetricsAndValues(metrics map[uint64]mmvdump.Metric, values map[uint64]*mmvdump.Value, instances map[uint64]mmvdump.Instance, strings map[uint64]*mmvdump.String, c *PCPClient, t *testing.T) {
	if c.Registry().MetricCount() != len(metrics) {
		t.Errorf("expected %v metrics, got %v", c.Registry().MetricCount(), len(metrics))
	}

	if c.Registry().ValuesCount() != len(values) {
		t.Errorf("expected %v values, got %v", c.Registry().ValuesCount(), len(values))
	}

	for _, m := range c.r.metrics {
		matchMetricAndValue(m, metrics, values, instances, strings, c, t)
	}
}

func TestWritingSingletonMetric(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Error(err)
		return
	}

	met, err := NewPCPSingletonMetric(10, "test.1", Int32Type, CounterSemantics, OneUnit, "test")
	if err != nil {
		t.Error(err)
		return
	}
	c.MustRegister(met)

	c.MustStart()
	defer c.MustStop()

	h, toc, m, v, i, ind, s, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	if int(h.Toc) != len(toc) {
		t.Errorf("expected the number of tocs specified in the header and the number of tocs in the toc array to be the same, h.Toc = %d, len(tocs) = %d", h.Toc, len(toc))
	}

	if h.Toc != 3 {
		t.Errorf("expected client to write %d tocs, written %d", 3, h.Toc)
	}

	if h.Flag != int32(ProcessFlag) {
		t.Errorf("expected client to write a ProcessFlag, writing %v", MMVFlag(h.Flag))
	}

	matchMetricsAndValues(m, v, i, s, c, t)

	// strings

	if len(s) != 1 {
		t.Error("expected one string")
	}

	_, me := findMetric(met, m)
	matchString(met.shortDescription, s[me.ShortText()], t)

	// instances

	if len(i) != 0 {
		t.Error("expected no instances when writing a singleton metric")
	}

	// indoms

	if len(ind) != 0 {
		t.Error("expected no indoms when writing a singleton metric")
	}
}

func TestUpdatingSingletonMetric(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Error(err)
		return
	}

	m := c.MustRegisterString("met.1", 10, Int32Type, CounterSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, _, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Fatal("Cannot extract dump from the writer buffer")
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)

	if m.(SingletonMetric).Val().(int32) != 10 {
		t.Errorf("expected metric value to be 10")
	}

	m.(SingletonMetric).MustSet(42)

	_, _, metrics, values, instances, _, strings, err = mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)

	if m.(SingletonMetric).Val().(int32) != 42 {
		t.Errorf("expected metric value to be 42")
	}
}

func matchInstance(i mmvdump.Instance, pi *pcpInstance, id *PCPInstanceDomain, indoms map[uint64]*mmvdump.InstanceDomain, strings map[uint64]*mmvdump.String, t *testing.T) {
	off, _ := findInstanceDomain(id, indoms)
	if i.Indom() != off {
		t.Errorf("expected indom offset to be %d, got %d", i.Indom(), off)
	}

	switch ins := i.(type) {
	case *mmvdump.Instance1:
		matchName(pi.name, ins.External, t)
	case *mmvdump.Instance2:
		matchString(pi.name, strings[ins.External], t)
	}
}

func matchInstanceDomain(id *mmvdump.InstanceDomain, pid *PCPInstanceDomain, strings map[uint64]*mmvdump.String, t *testing.T) {
	if pid.InstanceCount() != int(id.Count) {
		t.Errorf("expected %d instances in instance domain %v, got %d", pid.InstanceCount(), pid.Name(), id.Count)
	}

	if id.Shorttext != 0 {
		matchString(pid.shortDescription, strings[id.Shorttext], t)
	}

	if id.Longtext != 0 {
		matchString(pid.longDescription, strings[id.Longtext], t)
	}
}

func matchInstancesAndInstanceDomains(
	ins map[uint64]mmvdump.Instance,
	ids map[uint64]*mmvdump.InstanceDomain,
	ss map[uint64]*mmvdump.String,
	c *PCPClient,
	t *testing.T,
) {
	if len(ins) != c.r.InstanceCount() {
		t.Errorf("expected %d instances, got %d", c.r.InstanceCount(), len(ins))
	}

	for _, id := range c.r.instanceDomains {
		_, ind := findInstanceDomain(id, ids)

		if ind == nil {
			t.Errorf("expected an instance domain of name %v", id.name)
		} else {
			matchInstanceDomain(ind, id, ss, t)
		}

		for _, i := range id.instances {
			if ioff := uint64(i.offset); ins[ioff] == nil {
				t.Errorf("expected an instance domain at %d", ioff)
			} else {
				matchInstance(ins[ioff], i, id, ids, ss, t)
			}
		}
	}
}

func TestWritingInstanceMetric(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Error(err)
		return
	}

	id, err := NewPCPInstanceDomain("testid", []string{"a", "b", "c"}, "testid")
	if err != nil {
		t.Error(err)
		return
	}
	c.MustRegisterIndom(id)

	m, err := NewPCPInstanceMetric(Instances{
		"a": 1,
		"b": 2,
		"c": 3,
	}, "test.1", id, Uint32Type, CounterSemantics, OneUnit, "", "test long description")
	if err != nil {
		t.Error(err)
		return
	}

	c.MustRegister(m)

	c.MustStart()
	defer c.MustStop()

	h, tocs, mets, vals, ins, ids, ss, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	if int(h.Toc) != len(tocs) {
		t.Errorf("expected the number of tocs specified in the header and the number of tocs in the toc array to be the same, h.Toc = %d, len(tocs) = %d", h.Toc, len(tocs))
	}

	if h.Toc != 5 {
		t.Errorf("expected client to write %d tocs, written %d", 5, h.Toc)
	}

	if h.Flag != int32(ProcessFlag) {
		t.Errorf("expected client to write a ProcessFlag, writing %v", MMVFlag(h.Flag))
	}

	matchMetricsAndValues(mets, vals, ins, ss, c, t)

	matchInstancesAndInstanceDomains(ins, ids, ss, c, t)

	// strings

	_, ind := findInstanceDomain(id, ids)
	matchString(id.shortDescription, ss[ind.Shorttext], t)

	_, me := findMetric(m, mets)
	matchString(m.longDescription, ss[me.LongText()], t)
}

func TestUpdatingInstanceMetric(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Error(err)
		return
	}

	m := c.MustRegisterString("met[a, b].1", Instances{"a": 21, "b": 42}, Int32Type, CounterSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)

	im := m.(InstanceMetric)

	valmatcher := func(v interface{}, val int32, err error, ins string) {
		if err != nil {
			t.Errorf("cannot retrieve instance a value, error: %v", err)
			return
		}

		if v.(int32) != val {
			t.Errorf("expected instance %v's value to be %v", ins, val)
		}
	}

	v, err := im.ValInstance("a")
	valmatcher(v, 21, err, "a")

	v, err = im.ValInstance("b")
	valmatcher(v, 42, err, "b")

	im.MustSetInstance(63, "a")
	im.MustSetInstance(84, "b")

	_, _, metrics, values, instances, indoms, strings, err = mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)

	v, err = im.ValInstance("a")
	valmatcher(v, 63, err, "a")

	v, err = im.ValInstance("b")
	valmatcher(v, 84, err, "b")
}

func TestStringValueWriting(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Error(err)
		return
	}

	metric := c.MustRegisterString("test.str", "kirk", StringType, CounterSemantics, OneUnit)
	c.MustStart()
	defer c.MustStop()

	h, _, m, v, _, _, s, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	if h.Toc != 3 {
		t.Errorf("expected toc to be 3, not %v", h.Toc)
	}

	sm := metric.(*PCPSingletonMetric)
	off, _ := findMetric(sm, m)
	voff, val := findSingletonValue(off, v)

	if val == nil {
		t.Errorf("expected value at %v", voff)
	} else {
		add := uint64(val.Extra)
		if str, ok := s[add]; !ok {
			t.Errorf("expected a string at address %v", add)
		} else {
			if v := string(str.Payload[:4]); v != "kirk" {
				t.Errorf("expected metric value to be kirk, not %v", v)
			}
		}
	}

	sm.MustSet("spock")

	_, _, _, v, _, _, s, err = mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	_, val = findSingletonValue(off, v)
	add := uint64(val.Extra)
	if str, ok := s[add]; !ok {
		t.Errorf("expected a string at address %v", add)
	} else {
		if v := string(str.Payload[:5]); v != "spock" {
			t.Errorf("expected metric value to be spock, not %v", v)
		}
	}
}

func TestWritingDifferentSemantics(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.1", 10, Int32Type, NoSemantics, OneUnit)
	c.MustRegisterString("m.2", 10, Int32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m.3", 10, Int32Type, InstantSemantics, OneUnit)
	c.MustRegisterString("m.4", 10, Int32Type, DiscreteSemantics, OneUnit)

	c.MustRegisterString("m[a, b].5", Instances{"a": 1, "b": 2}, Int32Type, NoSemantics, OneUnit)
	c.MustRegisterString("m[a, b].6", Instances{"a": 3, "b": 4}, Int32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].7", Instances{"a": 5, "b": 6}, Int32Type, InstantSemantics, OneUnit)
	c.MustRegisterString("m[a, b].8", Instances{"a": 7, "b": 8}, Int32Type, DiscreteSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())

	if err != nil {
		t.Errorf("cannot create dump: %v", err)
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)
}

func TestWritingDifferentUnits(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.0", 10, Uint64Type, CounterSemantics, NanosecondUnit)
	c.MustRegisterString("m.1", 10, Uint64Type, CounterSemantics, MicrosecondUnit)
	c.MustRegisterString("m.2", 10, Uint64Type, CounterSemantics, MillisecondUnit)
	c.MustRegisterString("m.3", 10, Uint64Type, CounterSemantics, SecondUnit)
	c.MustRegisterString("m.4", 10, Uint64Type, CounterSemantics, MinuteUnit)
	c.MustRegisterString("m.5", 10, Uint64Type, CounterSemantics, HourUnit)

	c.MustRegisterString("m.6", 10, Uint64Type, CounterSemantics, ByteUnit)
	c.MustRegisterString("m.7", 10, Uint64Type, CounterSemantics, KilobyteUnit)
	c.MustRegisterString("m.8", 10, Uint64Type, CounterSemantics, MegabyteUnit)
	c.MustRegisterString("m.9", 10, Uint64Type, CounterSemantics, GigabyteUnit)
	c.MustRegisterString("m.10", 10, Uint64Type, CounterSemantics, TerabyteUnit)
	c.MustRegisterString("m.11", 10, Uint64Type, CounterSemantics, PetabyteUnit)
	c.MustRegisterString("m.12", 10, Uint64Type, CounterSemantics, ExabyteUnit)

	c.MustRegisterString("m.13", 10, Uint64Type, CounterSemantics, OneUnit)

	c.MustRegisterString("m[a, b].14", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, NanosecondUnit)
	c.MustRegisterString("m[a, b].15", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, MicrosecondUnit)
	c.MustRegisterString("m[a, b].16", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, MillisecondUnit)
	c.MustRegisterString("m[a, b].17", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, SecondUnit)
	c.MustRegisterString("m[a, b].18", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, MinuteUnit)
	c.MustRegisterString("m[a, b].19", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, HourUnit)

	c.MustRegisterString("m[a, b].20", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, ByteUnit)
	c.MustRegisterString("m[a, b].21", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, KilobyteUnit)
	c.MustRegisterString("m[a, b].22", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, MegabyteUnit)
	c.MustRegisterString("m[a, b].23", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, GigabyteUnit)
	c.MustRegisterString("m[a, b].24", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, TerabyteUnit)
	c.MustRegisterString("m[a, b].25", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, PetabyteUnit)
	c.MustRegisterString("m[a, b].26", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, ExabyteUnit)

	c.MustRegisterString("m[a, b].27", Instances{"a": 1, "b": 2}, Uint64Type, CounterSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump: %v", err)
		return
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)
}

func TestWritingDifferentTypes(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.1", 2147483647, Int32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m.2", 2147483647, Int64Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m.3", 4294967295, Uint32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m.4", 4294967295, Uint64Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m.5", 3.14, FloatType, CounterSemantics, OneUnit)
	c.MustRegisterString("m.6", 6.28, DoubleType, CounterSemantics, OneUnit)
	c.MustRegisterString("m.7", "luke", StringType, CounterSemantics, OneUnit)

	c.MustRegisterString("m[a, b].8", Instances{"a": 2147483647, "b": -2147483648}, Int32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].9", Instances{"a": 2147483647, "b": -2147483648}, Int64Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].10", Instances{"a": 4294967295, "b": 0}, Uint32Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].11", Instances{"a": 4294967295, "b": 0}, Uint64Type, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].12", Instances{"a": 3.14, "b": -3.14}, FloatType, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].13", Instances{"a": 6.28, "b": -6.28}, DoubleType, CounterSemantics, OneUnit)
	c.MustRegisterString("m[a, b].14", Instances{"a": "luke", "b": "skywalker"}, StringType, CounterSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump: %v", err)
		return
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)
}

func TestMMV2MetricWriting(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
		return
	}

	m := c.MustRegisterString("it_takes_a_big_man_to_cry_but_it_takes_a_bigger_man_to_laugh_at_that_man",
		21, Int32Type, CounterSemantics, OneUnit)

	c.MustStart()
	defer c.MustStop()

	h, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot create dump, error: %v", err)
	}

	if h.Version != 2 {
		t.Error("expected mmv version to be 2")
	}

	if h.Toc != 3 {
		t.Error("expected tocs to be 3")
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)

	if len(strings) != 1 {
		t.Error("expected one string in the dump")
	}

	_, met := findMetric(m, metrics)
	off := met.(*mmvdump.Metric2).Name
	if str, ok := strings[uint64(off)]; !ok {
		t.Errorf("expected a string at offset %v", off)
	} else if (string(str.Payload[:len(m.Name())])) != m.Name() {
		t.Error("the metric name in strings section doesn't match the registered metric name")
	}
}

func TestMMV2InstanceWriting(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
		return
	}

	c.MustRegisterString(
		"a[it_takes_a_big_man_to_cry_but_it_takes_a_bigger_man_to_laugh_at_that_man].b",
		Instances{
			"it_takes_a_big_man_to_cry_but_it_takes_a_bigger_man_to_laugh_at_that_man": 32,
		}, Int32Type, CounterSemantics, OneUnit,
	)

	c.MustStart()
	defer c.MustStop()

	h, _, metrics, values, instances, indoms, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot create dump, error: %v", err)
	}

	if h.Version != 2 {
		t.Error("expected mmv version to be 2")
	}

	if h.Toc != 5 {
		t.Error("expected tocs to be 3")
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, strings, c, t)

	if len(strings) != 2 {
		t.Errorf("expected two strings in the dump")
	}
}

func toFixed(v float64, p int) float64 {
	return float64(uint64(v*math.Pow(10, float64(p)))) / math.Pow(10, float64(p))
}

func matchSingleDump(expected interface{}, m PCPMetric, c *PCPClient, t *testing.T) {
	_, _, metrics, values, instances, _, strings, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump: %v", err)
		return
	}

	matchMetricsAndValues(metrics, values, instances, strings, c, t)

	off, _ := findMetric(m, metrics)
	_, v := findSingletonValue(off, values)

	if val, err := mmvdump.FixedVal(v.Val, mmvdump.Type(m.Type())); val != expected {
		t.Errorf("expected metric to be %v, got %v", expected, val)
	} else if err != nil {
		t.Errorf("cannot convert stored metric val to float64")
	}
}

func matchSingle(expected, val interface{}, m PCPMetric, c *PCPClient, t *testing.T) {
	if val != expected {
		t.Errorf("expected Val() to return %v", expected)
	}

	matchSingleDump(expected, m, c, t)
}

func TestCounter(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
		return
	}

	m, err := NewPCPCounter(0, "c.1")
	if err != nil {
		t.Errorf("cannot create counter, error: %v", err)
		return
	}

	c.MustRegister(m)

	c.MustStart()
	defer c.MustStop()

	// Up

	m.Up()
	matchSingle(int64(1), m.Val(), m, c, t)

	// Inc

	m.MustInc(9)
	matchSingle(int64(10), m.Val(), m, c, t)

	// Inc decrement

	err = m.Inc(-9)
	if err == nil {
		t.Errorf("expected decrementing a counter to generate an error")
	}
	matchSingle(int64(10), m.Val(), m, c, t)

	// Set less

	err = m.Set(9)
	if err == nil {
		t.Errorf("expected setting a counter to a lesser value to generate an error")
	}
	matchSingle(int64(10), m.Val(), m, c, t)

	// Set more

	err = m.Set(99)
	if err != nil {
		t.Errorf("expected setting a counter to a larger value to not generate an error")
	}
	matchSingle(int64(99), m.Val(), m, c, t)
}

func TestGauge(t *testing.T) {
	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
		return
	}

	m, err := NewPCPGauge(0, "g.1")
	if err != nil {
		t.Errorf("cannot create gauge, error: %v", err)
		return
	}

	c.MustRegister(m)

	c.MustStart()
	defer c.MustStop()

	// Inc

	m.MustInc(10)
	matchSingle(float64(10), m.Val(), m, c, t)

	// Dec

	err = m.Dec(9)
	if err != nil {
		t.Errorf("cannot decrement the gauge")
	}
	matchSingle(float64(1), m.Val(), m, c, t)

	// Set

	err = m.Set(9)
	if err != nil {
		t.Errorf("cannot set the gauge's value")
	}
	matchSingle(float64(9), m.Val(), m, c, t)
}

func TestTimer(t *testing.T) {
	timer, err := NewPCPTimer("t.1", NanosecondUnit)
	if err != nil {
		t.Errorf("cannot create timer, error: %v", err)
		return
	}

	c, err := NewPCPClient("test")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
		return
	}

	c.MustRegister(timer)

	c.MustStart()
	defer c.MustStop()

	err = timer.Start()
	if err != nil {
		t.Errorf("cannot start timer, error: %v", err)
	}

	time.Sleep(time.Second)

	v, err := timer.Stop()
	if err != nil {
		t.Errorf("cannot stop timer, error: %v", err)
	}

	matchSingleDump(v, timer, c, t)
}

func TestCounterVector(t *testing.T) {
	cv, err := NewPCPCounterVector(map[string]int64{
		"m1": 1,
		"m2": 2,
	}, "m.1")

	if err != nil {
		t.Errorf("cannot create CounterVector, error: %v", err)
		return
	}

	c, err := NewPCPClient("c")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
	}

	c.MustRegister(cv)

	c.MustStart()
	defer c.MustStop()

	var val int64

	// Set

	cv.MustSet(10, "m1")

	if val, err = cv.Val("m1"); val != 10 {
		t.Errorf("expected m.1[m1] to be 10, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m1] value, error: %v", err)
	}

	// Inc

	cv.MustInc(10, "m2")

	if val, err = cv.Val("m2"); val != 12 {
		t.Errorf("expected m.1[m2] to be 12, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m2] value, error: %v", err)
	}

	// Up

	cv.Up("m1")

	if val, err = cv.Val("m1"); val != 11 {
		t.Errorf("expected m.1[m1] to be 11, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m1] value, error: %v", err)
	}
}

func TestGaugeVector(t *testing.T) {
	g, err := NewPCPGaugeVector(map[string]float64{
		"m1": 1.2,
		"m2": 2.4,
	}, "m.1")

	if err != nil {
		t.Errorf("cannot create GaugeVector, error: %v", err)
		return
	}

	c, err := NewPCPClient("c")
	if err != nil {
		t.Errorf("cannot create client, error: %v", err)
	}

	c.MustRegister(g)

	c.MustStart()
	defer c.MustStop()

	var val float64

	// Set

	g.MustSet(10, "m1")

	if val, err = g.Val("m1"); val != 10 {
		t.Errorf("expected m.1[m1] to be 10, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m1] value, error: %v", err)
	}

	// Inc

	g.MustInc(10, "m2")

	if val, err = g.Val("m2"); val != 12.4 {
		t.Errorf("expected m.1[m2] to be 12.4, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m2] value, error: %v", err)
	}

	// Dec

	g.MustDec(10, "m2")

	if val, err = g.Val("m2"); toFixed(val, 5) != 2.4 {
		t.Errorf("expected m.1[m2] to be 2.4, got %v", val)
	} else if err != nil {
		t.Errorf("cannot retrieve m.1[m2] value, error: %v", err)
	}
}

func TestHistogram(t *testing.T) {
	hist := hdrhistogram.New(0, 100, 5)

	h, err := NewPCPHistogram("test.hist", 0, 100, 5, OneUnit)
	if err != nil {
		t.Fatalf("cannot create metric, error: %v", err)
	}

	c, err := NewPCPClient("test")
	if err != nil {
		t.Fatalf("cannot create client, error: %v", err)
	}

	c.MustRegister(h)

	c.MustStart()
	defer c.MustStop()

	_, _, m, v, i, id, s, err := mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Fatalf("cannot create dump, error: %v", err)
	}

	matchMetricsAndValues(m, v, i, s, c, t)
	matchInstancesAndInstanceDomains(i, id, s, c, t)

	for i := int64(1); i <= 100; i++ {
		err = hist.RecordValues(i, i)
		if err != nil {
			t.Errorf("hdrhistogram couldn't record, error: %v", err)
		}

		h.MustRecordN(i, i)
	}

	_, _, m, v, _, _, _, err = mmvdump.Dump(c.writer.Bytes())
	if err != nil {
		t.Fatalf("cannot create dump, error: %v", err)
	}

	cases := [...]struct {
		ins string
		val float64
	}{
		{"mean", hist.Mean()},
		{"mean", h.Mean()},

		{"variance", math.Pow(hist.StdDev(), 2)},
		{"variance", h.Variance()},

		{"standard_deviation", hist.StdDev()},
		{"standard_deviation", h.StandardDeviation()},

		{"max", float64(hist.Max())},
		{"max", float64(h.Max())},

		{"min", float64(hist.Min())},
		{"min", float64(h.Min())},
	}

	for _, c := range cases {
		off := h.indom.instances[c.ins].offset
		moff, _ := findMetric(h, m)
		_, dv := findInstanceValue(moff, uint64(off), v)
		val, _ := mmvdump.FixedVal(uint64(dv.Val), mmvdump.DoubleType)
		if c.val != val {
			t.Errorf("expected %v to be %v, got %v", c.ins, c.val, val)
		}
	}
}
