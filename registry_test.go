package speed

import "testing"

func TestIdentifierRegex(t *testing.T) {
	cases := []struct {
		val, indom, metric string
		instances          []string
	}{
		{"sheep[baabaablack].bagsfull.count", "sheep", "sheep.bagsfull.count", []string{"baabaablack"}},
		{"sheep[limpy].legs.available", "sheep", "sheep.legs.available", []string{"limpy"}},
		{"cow.how.now", "", "cow.how.now", nil},
		{"sheep[limpy,grumpy,chumpy].legs.available", "sheep", "sheep.legs.available", []string{"limpy", "grumpy", "chumpy"}},
		{"a", "", "a", []string{}},
		{"a_b", "", "a_b", []string{}},
		{"a_b._i", "", "a_b._i", []string{}},
		{"a_b[c_d, e_f, g_h]._i", "a_b", "a_b._i", []string{"c_d", "e_f", "g_h"}},
	}

	for _, c := range cases {
		m, id, i, err := parseString(c.val)

		if err != nil {
			t.Errorf("Error: '%v' while parsing %v", err, c.val)
			continue
		}

		if id != c.indom {
			t.Errorf("Wrong InstanceDomain for %v, expected %v, got %v", c.val, c.indom, id)
		}

		if m != c.metric {
			t.Errorf("Wrong Metric for %v, expected %v, got %v", c.val, c.metric, m)
		}

		if len(i) != len(c.instances) {
			t.Errorf("Wrong number of Instances for %v, expected %v, got %v", c.val, len(c.instances), i)
		} else {
			m := make(map[string]bool)
			for x := 0; x < len(i); x++ {
				m[i[x]] = true
			}

			for x := 0; x < len(i); x++ {
				_, present := m[c.instances[x]]
				if !present {
					t.Errorf("Instance %v not found in input", c.instances[x])
				}
			}
		}
	}
}

func TestStringSingletonConstruction(t *testing.T) {
	r := NewPCPRegistry()

	m, err := r.AddMetricByString("cow.how.now", 10, Int32Type, CounterSemantics, OneUnit)
	if err != nil {
		t.Error("Cannot parse, error", err)
		return
	}

	sm, ok := m.(*PCPSingletonMetric)
	if !ok {
		t.Error("Expected a PCPSingletonMetric")
	}

	if sm.Name() != "cow.how.now" {
		t.Errorf("Expected metric name to be %v, got %v", "cow.how.now", sm.Name())
	}

	if sm.Val() != int32(10) {
		t.Errorf("Expected metric value to be %v, got %v", 10, sm.Val())
	}

	if r.InstanceCount() != 0 {
		t.Error("Expected Instance Count to be 0")
	}

	if r.InstanceDomainCount() != 0 {
		t.Error("Expected Instance Domain Count to be 0")
	}

	if r.MetricCount() != 1 {
		t.Error("Expected Metric Count to be 1")
	}
}

func TestStringInstanceConstruction(t *testing.T) {
	r := NewPCPRegistry()

	m, err := r.AddMetricByString("sheep[limpy,grumpy,chumpy].legs.available", Instances{
		"limpy":  10,
		"grumpy": 20,
		"chumpy": 30,
	}, Int32Type, CounterSemantics, OneUnit)
	if err != nil {
		t.Error("Cannot parse, error", err)
	}

	im := m.(*PCPInstanceMetric)

	if im.Name() != "sheep.legs.available" {
		t.Errorf("Expected metric name to be %v, got %v", "sheep.legs.available", im.Name())
	}

	for i, v := range map[string]int32{"limpy": 10, "grumpy": 20, "chumpy": 30} {
		val, err := im.ValInstance(i)

		if err != nil {
			t.Errorf("error retrieving instance %v value", i)
		}

		if val != v {
			t.Errorf("wrong value for instance %v, expected %v, got %v", i, v, val)
		}
	}

	if r.InstanceCount() != 3 {
		t.Errorf("Expected Instance Count to be 3, got %v", r.InstanceCount())
	}

	if r.InstanceDomainCount() != 1 {
		t.Errorf("Expected Instance Domain Count to be 1, got %v", r.InstanceDomainCount())
	}

	if r.MetricCount() != 1 {
		t.Errorf("Expected Metric Count to be 1, got %v", r.MetricCount())
	}

	if r.ValuesCount() != 3 {
		t.Errorf("Expected Value Count to be 3, got %v", r.ValuesCount())
	}
}

func TestMMV2MetricRegistration(t *testing.T) {
	r := NewPCPRegistry()

	m, err := NewPCPSingletonMetric(10, "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", Int32Type, CounterSemantics, OneUnit)
	if err != nil {
		t.Errorf("cannot create metric, error: %v", err)
		return
	}

	err = r.AddMetric(m)
	if err != nil {
		t.Errorf("cannot add metric to registry, error: %v", err)
		return
	}

	if r.StringCount() != 1 {
		t.Errorf("expected the metric name to be registered in the strings section")
	}
}
