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

	m, err := NewPCPSingletonMetric(10, "test", Int32Type, CounterSemantics, OneUnit, "test")
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

func matchMetricDesc(desc *PCPMetricDesc, metric *mmvdump.Metric, t *testing.T) {
	if int32(metric.Sem) != int32(desc.sem) {
		t.Errorf("expected semantics to be %v, got %v", desc.sem, MetricSemantics(metric.Sem))
	}

	if int32(metric.Typ) != int32(desc.t) {
		t.Errorf("expected type to be %v, got %v", desc.t, MetricType(metric.Typ))
	}

	if int32(metric.Unit) != int32(desc.u.PMAPI()) {
		t.Errorf("expected unit to be %v, got %v", desc.u, metric.Unit)
	}

	if metric.Shorttext != uint64(desc.shortDescription.offset) {
		t.Errorf("expected shorttext to be %v, got %v", desc.shortDescription.offset, metric.Shorttext)
	}

	if metric.Longtext != uint64(desc.longDescription.offset) {
		t.Errorf("expected longtext to be %v, got %v", desc.longDescription.offset, metric.Longtext)
	}
}

func matchSingletonMetric(m *PCPSingletonMetric, metric *mmvdump.Metric, t *testing.T) {
	if metric.Indom != mmvdump.NoIndom {
		t.Error("expected indom to be null")
	}

	matchMetricDesc(m.PCPMetricDesc, metric, t)
}

func matchSingletonValue(m *PCPSingletonMetric, value *mmvdump.Value, t *testing.T) {
	if value.Metric != uint64(m.descoffset) {
		t.Errorf("expected value's metric to be at %v", m.descoffset)
	}

	if m.t == StringType {
		if int64(m.val.(*pcpString).offset) != value.Extra {
			t.Errorf("expected the string value to be written at %v, got %v", value.Extra, m.val.(*pcpString).offset)
		}
	} else {
		if av, err := mmvdump.FixedVal(value.Val, mmvdump.Type(m.t)); err != nil || av != m.val {
			t.Errorf("expected the value to be %v, got %v", m.val, av)
		}
	}

	if value.Instance != 0 {
		t.Errorf("expected value instance to be 0")
	}
}

func matchString(s *pcpString, str *mmvdump.String, t *testing.T) {
	if s == nil {
		t.Error("expected PCPString to not be nil")
	}

	sv := string(str.Payload[:len(s.val)])
	if sv != s.val {
		t.Errorf("expected %v, got %v", s.val, sv)
	}
}

func matchInstanceMetric(m *PCPInstanceMetric, met *mmvdump.Metric, t *testing.T) {
	if uint32(met.Indom) != m.indom.id {
		t.Errorf("expected indom id to be %d, got %d", m.indom.id, met.Indom)
	}

	matchMetricDesc(m.PCPMetricDesc, met, t)
}

func matchInstanceValue(v *mmvdump.Value, i *instanceValue, ins string, met *PCPInstanceMetric, t *testing.T) {
	if v.Metric != uint64(met.descoffset) {
		t.Errorf("expected value's metric to be at %v", met.descoffset)
	}

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
		if int64(i.val.(*pcpString).offset) != v.Extra {
			t.Errorf("expected the string value to be written at %v, got %v", v.Extra, i.val.(*pcpString).offset)
		}
	} else {
		if av, err := mmvdump.FixedVal(v.Val, mmvdump.Type(met.t)); err != nil || av != i.val {
			t.Errorf("expected the value to be %v, got %v", i.val, av)
		}
	}
}

func matchSingletonMetricAndValue(met *PCPSingletonMetric, metrics map[uint64]*mmvdump.Metric, values map[uint64]*mmvdump.Value, t *testing.T) {
	metric, ok := metrics[uint64(met.descoffset)]
	if !ok {
		t.Errorf("expected a metric at offset %v", met.descoffset)
	} else {
		matchSingletonMetric(met, metric, t)
	}

	mv, ok := values[uint64(met.valueoffset)]
	if !ok {
		t.Errorf("expected a value at offset %v", met.valueoffset)
	} else {
		matchSingletonValue(met, mv, t)
	}
}

func matchInstanceMetricAndValues(met *PCPInstanceMetric, metrics map[uint64]*mmvdump.Metric, values map[uint64]*mmvdump.Value, t *testing.T) {
	metric, ok := metrics[uint64(met.descoffset)]
	if !ok {
		t.Errorf("expected a metric at offset %v", met.descoffset)
	} else {
		matchInstanceMetric(met, metric, t)
	}

	for n, i := range met.vals {
		mv, ok := values[uint64(i.offset)]
		if !ok {
			t.Errorf("expected a value at offset %v", i.offset)
		} else {
			matchInstanceValue(mv, i, n, met, t)
		}
	}
}

func matchMetricsAndValues(metrics map[uint64]*mmvdump.Metric, values map[uint64]*mmvdump.Value, c *PCPClient, t *testing.T) {
	if c.Registry().MetricCount() != len(metrics) {
		t.Errorf("expected %v metrics, got %v", c.Registry().MetricCount(), len(metrics))
	}

	if c.Registry().ValuesCount() != len(values) {
		t.Errorf("expected %v values, got %v", c.Registry().ValuesCount(), len(values))
	}

	for _, m := range c.r.metrics {
		switch met := m.(type) {
		case *PCPSingletonMetric:
			matchSingletonMetricAndValue(met, metrics, values, t)
		case *PCPInstanceMetric:
			matchInstanceMetricAndValues(met, metrics, values, t)
		}
	}
}

func TestWritingSingletonMetric(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
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

	h, toc, m, v, i, ind, s, err := mmvdump.Dump(c.buffer.Bytes())
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

	matchMetricsAndValues(m, v, c, t)

	// strings

	if len(s) != 1 {
		t.Error("expected one string")
	}

	matchString(met.shortDescription, s[uint64(met.shortDescription.offset)], t)

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
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Error(err)
		return
	}

	m := c.MustRegisterString("met.1", 10, CounterSemantics, Int32Type, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, _, _, _, err := mmvdump.Dump(c.buffer.Bytes())
	matchMetricsAndValues(metrics, values, c, t)

	if m.(SingletonMetric).Val().(int32) != 10 {
		t.Errorf("expected metric value to be 10")
	}

	m.(SingletonMetric).MustSet(42)

	_, _, metrics, values, _, _, _, err = mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, c, t)

	if m.(SingletonMetric).Val().(int32) != 42 {
		t.Errorf("expected metric value to be 42")
	}
}

func matchInstance(i *mmvdump.Instance, pi *pcpInstance, id *PCPInstanceDomain, t *testing.T) {
	if i.Indom != uint64(id.offset) {
		t.Errorf("expected indom offset to be %d, got %d", i.Indom, id.offset)
	}

	if in := i.External[:len(pi.name.val)]; pi.name.val != string(in) {
		t.Errorf("expected instance name to be %v, got %v", pi.name.val, in)
	}
}

func matchInstanceDomain(id *mmvdump.InstanceDomain, pid *PCPInstanceDomain, t *testing.T) {
	if pid.InstanceCount() != int(id.Count) {
		t.Errorf("expected %d instances in instance domain %v, got %d", pid.InstanceCount(), pid.Name(), id.Count)
	}

	if id.Offset != uint64(pid.instanceOffset) {
		t.Errorf("expected instance offset to be %d, got %d", pid.instanceOffset, id.Offset)
	}

	if id.Shorttext != uint64(pid.shortDescription.offset) {
		t.Errorf("expected short description to be %d, got %d", pid.shortDescription.offset, id.Shorttext)
	}

	if id.Longtext != uint64(pid.longDescription.offset) {
		t.Errorf("expected long description to be %d, got %d", pid.longDescription.offset, id.Longtext)
	}
}

func matchInstancesAndInstanceDomains(
	ins map[uint64]*mmvdump.Instance,
	ids map[uint64]*mmvdump.InstanceDomain,
	c *PCPClient,
	t *testing.T,
) {
	if len(ins) != c.r.InstanceCount() {
		t.Errorf("expected %d instances, got %d", c.r.InstanceCount(), len(ins))
	}

	for _, id := range c.r.instanceDomains {
		if off := uint64(id.offset); ids[off] == nil {
			t.Errorf("expected an instance domain at %d", id.offset)
		} else {
			matchInstanceDomain(ids[off], id, t)
		}

		for _, i := range id.instances {
			if ioff := uint64(i.offset); ins[ioff] == nil {
				t.Errorf("expected an instance domain at %d", ioff)
			} else {
				matchInstance(ins[ioff], i, id, t)
			}
		}
	}
}

func TestWritingInstanceMetric(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
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

	h, tocs, mets, vals, ins, ids, ss, err := mmvdump.Dump(c.buffer.Bytes())
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

	matchMetricsAndValues(mets, vals, c, t)

	matchInstancesAndInstanceDomains(ins, ids, c, t)

	// strings

	if off := id.shortDescription.offset; ss[uint64(off)] != nil {
		matchString(id.shortDescription, ss[uint64(off)], t)
	} else {
		t.Errorf("expected a string at offset %v", off)
	}

	if off := m.longDescription.offset; ss[uint64(off)] != nil {
		matchString(m.longDescription, ss[uint64(off)], t)
	} else {
		t.Errorf("expected a string at offset %v", off)
	}
}

func TestUpdatingInstanceMetric(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Error(err)
		return
	}

	m := c.MustRegisterString("met[a, b].1", Instances{"a": 21, "b": 42}, CounterSemantics, Int32Type, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, _, err := mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, c, t)

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

	im.MustSetInstance("a", 63)
	im.MustSetInstance("b", 84)

	_, _, metrics, values, instances, indoms, _, err = mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump, error: %v", err)
	}

	matchMetricsAndValues(metrics, values, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, c, t)

	v, err = im.ValInstance("a")
	valmatcher(v, 63, err, "a")

	v, err = im.ValInstance("b")
	valmatcher(v, 84, err, "b")
}

func TestStringValueWriting(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Error(err)
		return
	}

	metric := c.MustRegisterString("test.str", "kirk", CounterSemantics, StringType, OneUnit)
	c.MustStart()
	defer c.MustStop()

	h, _, _, v, _, _, s, err := mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	if h.Toc != 3 {
		t.Errorf("expected toc to be 3, not %v", h.Toc)
	}

	sm := metric.(*PCPSingletonMetric)

	if val, ok := v[uint64(sm.valueoffset)]; !ok {
		t.Errorf("expected value at %v", sm.valueoffset)
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

	_, _, _, v, _, _, s, err = mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Error(err)
		return
	}

	val := v[uint64(sm.valueoffset)]
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
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.1", 10, NoSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m.2", 10, CounterSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m.3", 10, InstantSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m.4", 10, DiscreteSemantics, Int32Type, OneUnit)

	c.MustRegisterString("m[a, b].5", Instances{"a": 1, "b": 2}, NoSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m[a, b].6", Instances{"a": 3, "b": 4}, NoSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m[a, b].7", Instances{"a": 5, "b": 6}, NoSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m[a, b].8", Instances{"a": 7, "b": 8}, NoSemantics, Int32Type, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, _, err := mmvdump.Dump(c.buffer.Bytes())

	if err != nil {
		t.Errorf("cannot create dump: %v", err)
	}

	matchMetricsAndValues(metrics, values, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, c, t)
}

func TestWritingDifferentUnits(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.0", 10, CounterSemantics, Uint64Type, NanosecondUnit)
	c.MustRegisterString("m.1", 10, CounterSemantics, Uint64Type, MicrosecondUnit)
	c.MustRegisterString("m.2", 10, CounterSemantics, Uint64Type, MillisecondUnit)
	c.MustRegisterString("m.3", 10, CounterSemantics, Uint64Type, SecondUnit)
	c.MustRegisterString("m.4", 10, CounterSemantics, Uint64Type, MinuteUnit)
	c.MustRegisterString("m.5", 10, CounterSemantics, Uint64Type, HourUnit)

	c.MustRegisterString("m.6", 10, CounterSemantics, Uint64Type, ByteUnit)
	c.MustRegisterString("m.7", 10, CounterSemantics, Uint64Type, KilobyteUnit)
	c.MustRegisterString("m.8", 10, CounterSemantics, Uint64Type, MegabyteUnit)
	c.MustRegisterString("m.9", 10, CounterSemantics, Uint64Type, GigabyteUnit)
	c.MustRegisterString("m.10", 10, CounterSemantics, Uint64Type, TerabyteUnit)
	c.MustRegisterString("m.11", 10, CounterSemantics, Uint64Type, PetabyteUnit)
	c.MustRegisterString("m.12", 10, CounterSemantics, Uint64Type, ExabyteUnit)

	c.MustRegisterString("m.13", 10, CounterSemantics, Uint64Type, OneUnit)

	c.MustRegisterString("m[a, b].14", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, NanosecondUnit)
	c.MustRegisterString("m[a, b].15", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, MicrosecondUnit)
	c.MustRegisterString("m[a, b].16", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, MillisecondUnit)
	c.MustRegisterString("m[a, b].17", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, SecondUnit)
	c.MustRegisterString("m[a, b].18", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, MinuteUnit)
	c.MustRegisterString("m[a, b].19", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, HourUnit)

	c.MustRegisterString("m[a, b].20", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, ByteUnit)
	c.MustRegisterString("m[a, b].21", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, KilobyteUnit)
	c.MustRegisterString("m[a, b].22", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, MegabyteUnit)
	c.MustRegisterString("m[a, b].23", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, GigabyteUnit)
	c.MustRegisterString("m[a, b].24", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, TerabyteUnit)
	c.MustRegisterString("m[a, b].25", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, PetabyteUnit)
	c.MustRegisterString("m[a, b].26", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, ExabyteUnit)

	c.MustRegisterString("m[a, b].27", Instances{"a": 1, "b": 2}, CounterSemantics, Uint64Type, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, _, err := mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump: %v", err)
		return
	}

	matchMetricsAndValues(metrics, values, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, c, t)
}

func TestWritingDifferentTypes(t *testing.T) {
	c, err := NewPCPClient("test", ProcessFlag)
	if err != nil {
		t.Errorf("cannot create client: %v", err)
		return
	}

	c.MustRegisterString("m.1", 2147483647, CounterSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m.2", 2147483647, CounterSemantics, Int64Type, OneUnit)
	c.MustRegisterString("m.3", 4294967295, CounterSemantics, Uint32Type, OneUnit)
	c.MustRegisterString("m.4", 4294967295, CounterSemantics, Uint64Type, OneUnit)
	c.MustRegisterString("m.5", 3.14, CounterSemantics, FloatType, OneUnit)
	c.MustRegisterString("m.6", 6.28, CounterSemantics, DoubleType, OneUnit)
	c.MustRegisterString("m.7", "luke", CounterSemantics, StringType, OneUnit)

	c.MustRegisterString("m[a, b].8", Instances{"a": 2147483647, "b": -2147483648}, CounterSemantics, Int32Type, OneUnit)
	c.MustRegisterString("m[a, b].9", Instances{"a": 2147483647, "b": -2147483648}, CounterSemantics, Int64Type, OneUnit)
	c.MustRegisterString("m[a, b].10", Instances{"a": 4294967295, "b": 0}, CounterSemantics, Uint32Type, OneUnit)
	c.MustRegisterString("m[a, b].11", Instances{"a": 4294967295, "b": 0}, CounterSemantics, Uint64Type, OneUnit)
	c.MustRegisterString("m[a, b].12", Instances{"a": 3.14, "b": -3.14}, CounterSemantics, FloatType, OneUnit)
	c.MustRegisterString("m[a, b].13", Instances{"a": 6.28, "b": -6.28}, CounterSemantics, DoubleType, OneUnit)
	c.MustRegisterString("m[a, b].14", Instances{"a": "luke", "b": "skywalker"}, CounterSemantics, StringType, OneUnit)

	c.MustStart()
	defer c.MustStop()

	_, _, metrics, values, instances, indoms, _, err := mmvdump.Dump(c.buffer.Bytes())
	if err != nil {
		t.Errorf("cannot get dump: %v", err)
		return
	}

	matchMetricsAndValues(metrics, values, c, t)
	matchInstancesAndInstanceDomains(instances, indoms, c, t)
}
