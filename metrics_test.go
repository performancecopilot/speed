package speed

import (
	"math"
	"testing"
)

// only tests that work on 32 bit architectures or both go here
// tests only working on 64 bit architectures go in _amd64_test.go
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

		{Int32Type, math.MinInt32, true},
		{Int64Type, math.MinInt32, true},
		{Uint32Type, math.MinInt32, false},
		{Uint64Type, math.MinInt32, false},

		{Int32Type, int32(math.MinInt32), true},
		{Int64Type, int64(math.MinInt32), true},
		{Uint32Type, int32(math.MinInt32), false},
		{Uint64Type, int64(math.MinInt32), false},

		{Int32Type, math.MaxInt32, true},
		{Int64Type, math.MaxInt32, true},
		{Uint32Type, math.MaxInt32, true},
		{Uint64Type, math.MaxInt32, true},

		{Int32Type, int32(math.MaxInt32), true},
		{Int64Type, int64(math.MaxInt32), true},
		{Uint32Type, int32(math.MaxInt32), false},
		{Uint64Type, int64(math.MaxInt32), false},

		{Int64Type, int64(math.MaxInt64), true},

		{Uint32Type, uint32(math.MaxUint32), true},
		{Uint64Type, uint64(math.MaxUint32), true},

		{Uint32Type, uint(math.MaxUint32), true},

		{Uint32Type, uint32(math.MaxUint32), true},
		{Uint64Type, uint64(math.MaxUint64), true},

		{FloatType, math.MaxFloat32, true},
		{DoubleType, math.MaxFloat32, true},
		{FloatType, -math.MaxFloat32, true},
		{DoubleType, -math.MaxFloat32, true},

		{FloatType, float32(math.MaxFloat32), true},
		{DoubleType, float32(math.MaxFloat32), false},
		{FloatType, float32(-math.MaxFloat32), true},
		{DoubleType, float32(-math.MaxFloat32), false},

		{FloatType, float64(math.MaxFloat32), true},
		{DoubleType, float64(math.MaxFloat32), true},
		{FloatType, float64(-math.MaxFloat32), true},
		{DoubleType, float64(-math.MaxFloat32), true},

		{StringType, 10, false},
		{StringType, 10.10, false},
		{StringType, "10", true},
	}

	for _, c := range cases {
		r := c.t.IsCompatible(c.v)
		if r != c.result {
			t.Errorf("%v.IsCompatible(%v(%T)) should be %v, not %v", c.t, c.v, c.v, c.result, r)
		}
	}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		t           MetricType
		val, resval interface{}
	}{
		{Int32Type, 10, int32(10)},
		{Int64Type, 10, int64(10)},
		{Uint32Type, 10, uint32(10)},
		{Uint64Type, 10, uint64(10)},

		{Int32Type, int32(10), int32(10)},
		{Int64Type, int64(10), int64(10)},
		{Uint32Type, uint32(10), uint32(10)},
		{Uint64Type, uint64(10), uint64(10)},

		{Uint32Type, uint(10), uint32(10)},
		{Uint64Type, uint(10), uint64(10)},

		{Uint32Type, uint32(10), uint32(10)},
		{Uint64Type, uint64(10), uint64(10)},

		{FloatType, 3.14, float32(3.14)},
		{DoubleType, 3.14, float64(3.14)},

		{FloatType, float32(3.14), float32(3.14)},
		{DoubleType, float64(3.14), float64(3.14)},
	}

	for _, c := range cases {
		if c.t.resolve(c.val) != c.resval {
			t.Errorf("expected %T to resolve to %T", c.val, c.resval)
		}
	}
}

func TestComposition(t *testing.T) {
	ms := MegabyteUnit.Time(SecondUnit, -1)

	if ms.String() != "MegabyteUnit^1SecondUnit^-1" {
		t.Errorf("expected ms.String() to be MegabyteUnit^1SecondUnit^-1, got %s", ms.String())
	}

	if ms.PMAPI() != 520237056 {
		t.Errorf("expected ms.PMAPI() to be 520237056, got %v", ms.PMAPI())
	}

	hz := NewMetricUnit().Time(SecondUnit, -1)

	if hz.String() != "SecondUnit^-1" {
		t.Errorf("expected hz.String() to be SecondUnit^-1, got %s", ms.String())
	}

	if hz.PMAPI() != 251670528 {
		t.Errorf("expected hz.PMAPI() to be 251670528, got %v", hz.PMAPI())
	}

	cs1 := OneUnit.Space(MegabyteUnit, 2).Time(SecondUnit, -2)
	cs2 := NewMetricUnit().Time(SecondUnit, -2).Space(MegabyteUnit, 2).Count(OneUnit, 1)

	if cs1.PMAPI() != cs2.PMAPI() {
		t.Errorf("expected %v to be equal to %v", cs1.PMAPI(), cs2.PMAPI())
	}

	if cs1.String() != cs2.String() {
		t.Errorf("expected %v to be equal to %v", cs1.String(), cs2.String())
	}
}
