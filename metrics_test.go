package speed

import (
	"math"
	"testing"
)

func TestIsCompatible(t *testing.T) {
	cases := []struct {
		t      MetricType
		v      interface{}
		result bool
	}{
		{Int32Type, -1, true},
		{Int64Type, -1, true},
		{Uint64Type, -1, false},
		{Uint32Type, -1, false},
		{Int32Type, 2147483648, false},
		{Int64Type, 2147483648, true},
		{Int32Type, int32(-2147483648), true},
		{Int64Type, int64(-2147483648), true},
		{Uint32Type, int32(-2147483648), false},
		{Uint64Type, int64(-2147483648), false},
		{Uint32Type, uint32(math.MaxUint32), true},
		{Uint64Type, uint64(math.MaxUint32), true},
		{Uint32Type, uint(math.MaxUint32), true},
		{Uint64Type, uint(math.MaxUint64), true},
		// {Uint32Type, math.MaxUint32, true},
		// {Uint64Type, uint64(math.MaxUint64), true},
	}

	for _, c := range cases {
		r := c.t.IsCompatible(c.v)
		if r != c.result {
			t.Errorf("%v.IsCompatible(%v(%T)) should be %v, not %v", c.t, c.v, c.v, c.result, r)
		}
	}
}
