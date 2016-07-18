package speed

import (
	"math"
	"testing"
)

func TestIsCompatible64(t *testing.T) {
	cases := []struct {
		t      MetricType
		v      interface{}
		result bool
	}{
		{Int32Type, math.MinInt32 - 1, false},
		{Int32Type, math.MinInt64, false},
		{Uint32Type, math.MaxUint32 + 1, false},
		{Uint32Type, math.MaxInt64, false},
	}

	for _, c := range cases {
		r := c.t.IsCompatible(c.v)
		if r != c.result {
			t.Errorf("%v.IsCompatible(%v(%T)) should be %v, not %v", c.t, c.v, c.v, c.result, r)
		}
	}
}
