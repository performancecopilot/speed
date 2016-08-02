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
		{Int32Type, math.MaxInt32 + 1, false},
		{Int64Type, math.MaxInt32 + 1, true},
		{Uint32Type, math.MaxInt32 + 1, true},
		{Uint64Type, math.MaxInt32 + 1, true},

		{Int32Type, math.MinInt32 - 1, false},
		{Int64Type, math.MinInt32 - 1, true},
		{Uint32Type, math.MinInt32 - 1, false},
		{Uint64Type, math.MinInt32 - 1, false},

		{Int32Type, math.MinInt64, false},
		{Int64Type, math.MinInt64, true},
		{Uint32Type, math.MinInt64, false},
		{Uint64Type, math.MinInt64, false},

		{Int32Type, math.MaxInt64, false},
		{Int64Type, math.MaxInt64, true},
		{Uint32Type, math.MaxInt64, false},
		{Uint64Type, math.MaxInt64, true},

		{Uint32Type, math.MaxUint32 + 1, false},
		{Uint64Type, math.MaxUint32 + 1, true},

		{Uint32Type, uint(math.MaxUint32 + 1), false},
		{Uint64Type, uint(math.MaxUint32 + 1), true},

		{Uint64Type, uint(math.MaxUint64), true},

		{Uint64Type, uint64(math.MaxUint64), true},

		{FloatType, math.MaxFloat32, true},
		{FloatType, -math.MaxFloat32, true},
		{DoubleType, math.MaxFloat32, true},
		{DoubleType, -math.MaxFloat32, true},

		// we cannot test for `math.MaxFloat32 + 1`
		// or even `math.MaxFloat32 + math.MaxUint64`
		// because in floating point those aren't significant
		// enough additions, so the comparison will be equal
		// for more http://stackoverflow.com/q/17588419/3673043

		{FloatType, math.MaxFloat32 * 2, false},
		{DoubleType, math.MaxFloat32 * 2, true},

		{FloatType, math.MaxFloat64, false},
		{FloatType, -math.MaxFloat64, false},

		{DoubleType, math.MaxFloat64, true},
		{DoubleType, -math.MaxFloat64, true},
	}

	for i, c := range cases {
		r := c.t.IsCompatible(c.v)
		if r != c.result {
			f := math.MaxFloat32
			println((f + 1) > f)
			t.Errorf("case %v: %v.IsCompatible(%v(%T)) should be %v, not %v", i+1, c.t, c.v, c.v, c.result, r)
		}
	}
}
