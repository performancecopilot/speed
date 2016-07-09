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
