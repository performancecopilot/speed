package speed

import "testing"

func TestIdentifierRegex(t *testing.T) {
	cases := []struct {
		val, iname, indomname, mname string
	}{
		{"sheep[baabaablack].bagsfull.count", "baabaablack", "sheep", "sheep.bagsfull.count"},
		{"sheep[limpy].legs.available", "limpy", "sheep", "sheep.legs.available"},
		{"cow.how.now", "", "cow", "cow.how.now"},
	}

	for _, c := range cases {
		i, id, m, err := parseString(c.val)

		if err != nil {
			t.Errorf("Failing to parse %s\n", c.val)
			continue
		}

		if i != c.iname {
			t.Errorf("Instance name incorrectly parsed, expected %s, got %s\n", c.iname, i)
		}

		if id != c.indomname {
			t.Errorf("Instance Domain name incorrectly parsed, expected %s, got %s\n", c.indomname, id)
		}

		if m != c.mname {
			t.Errorf("Metric name incorrectly parsed, expected %s, got %s\n", c.mname, m)
		}
	}
}
