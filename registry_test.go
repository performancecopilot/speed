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
	}

	for _, c := range cases {
		m, id, i, err := parseString(c.val)

		if err != nil {
			t.Errorf("Error %v while parsing %v", err, c.val)
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

	m, err := r.AddMetricByString("cow.how.now", 10, CounterSemantics, Int32Type, OneUnit)
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
}

func TestStringInstanceConstruction(t *testing.T) {
	r := NewPCPRegistry()

	m, err := r.AddMetricByString("sheep[limpy,grumpy,chumpy].legs.available", map[string]interface{}{
		"limpy":  10,
		"grumpy": 20,
		"chumpy": 30,
	}, CounterSemantics, Int32Type, OneUnit)
	if err != nil {
		t.Error("Cannot parse, error", err)
		return
	}

	im, ok := m.(*PCPInstanceMetric)
	if !ok {
		t.Error("Expected a PCPInstanceMetric")
	}

	if im.Name() != "sheep.legs.available" {
		t.Errorf("Expected metric name to be %v, got %v", "cow.how.now", im.Name())
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
}
