package nanbox

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

var (
	known uint64
)

func init() {
	known = math.Float64bits(math.NaN())
}

func TestPackInt32(t *testing.T) {
	for i := -512; i < 512; i++ {

		packed := PackInt32(int32(i))
		unpacked := UnpackInt32(packed)

		if !math.IsNaN(packed) {
			t.Errorf("was not NaN: %+v", packed)
		}

		if unpacked != int32(i) {
			t.Errorf("couldn't pack: %+v", i)
		}

		bits := math.Float64bits(packed)
		if bits == known {
			t.Errorf("should never be known nan: (%d) %+v", i, bits)
		}
	}
}

func toUint64(v int) (u uint64) {
	return uint64(v)
}

func TestPackUint51(t *testing.T) {
	type testcase struct {
		in, out uint64
	}

	cases := []testcase{
		{0, 0},
		{1, 1},
		{toUint64(-1), toUint64(-1) & pack51mask},
	}
	for _, c := range cases {
		packed := PackUint51(c.in)
		unpacked := UnpackUint51(packed)

		if math.Float64bits(packed) == known {
			t.Errorf("should never be known nan")
		}

		if !math.IsNaN(packed) {
			t.Errorf("was not NaN: %+v", packed)
		}

		if unpacked != c.out {
			t.Errorf("bad unpacked=%x wanted=%x (from %x)", unpacked, c.out, c.in)
		}
	}
}

func TestPackBytes(t *testing.T) {
	for i := 0; i < 100; i++ {
		length := rand.Int31n(32)
		data := make([]byte, length)
		for j := 0; j < int(length); j++ {
			data[j] = byte(rand.Intn(256))
		}

		packed := PackBytes(data)
		size := len(packed)
		consumed, unpacked := UnpackBytes(packed)

		for _, p := range packed {
			if !math.IsNaN(p) {
				t.Errorf("was not all NaN: %+v", packed)
			}
			if math.Float64bits(p) == known {
				t.Errorf("should never be known nan")
			}
		}

		if unpacked == nil {
			unpacked = []byte{}
		}

		if consumed != size {
			t.Errorf("wrong length consumed=%d size=%d", consumed, size)
		} else if !reflect.DeepEqual(data, unpacked) {
			t.Errorf("bad data size=%d (src=%+v unpacked=%+v)", size, data, unpacked)
		}
	}
}

func TestUnpackBytes(t *testing.T) {

	errorCases := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		PackInt32(3),
	}

	for _, c := range errorCases {
		_, unpacked := UnpackBytes([]float64{c})
		if unpacked != nil {
			t.Errorf("expected nil out for: %+v", c)
		}
	}

}
