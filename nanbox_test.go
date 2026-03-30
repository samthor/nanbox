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

func TestPackNil(t *testing.T) {
	packed := PackBytes(nil)
	if len(packed) != 1 {
		t.Errorf("must pack one empty float64")
	}

	consumed, unpacked := UnpackBytes(packed)
	if consumed != 1 || unpacked != nil {
		t.Errorf("bad unpacked nil: consumed=%v unpacked=%v", consumed, unpacked)
	}
}

func TestPackBytesRun(t *testing.T) {
	var src [][]byte
	var packed []float64

	for i := 0; i < 4; i++ {
		length := rand.Int31n(128)
		data := make([]byte, length)
		for j := 0; j < int(length); j++ {
			data[j] = byte(rand.Intn(256))
		}

		src = append(src, data)

		p := PackBytes(data)
		packed = append(packed, p...)
	}

	j := 0
	var out [][]byte
	for j < len(packed) {
		consumed, unpacked := UnpackBytes(packed[j:])
		if consumed == 0 {
			t.Fatalf("could not consume more bytes")
		}
		out = append(out, unpacked)
		j += consumed
	}

	if !reflect.DeepEqual(src, out) {
		t.Errorf("could not enc/dec many bytes: src=%+v out=%+v", src, out)
	}
}

func TestNaNMap(t *testing.T) {
	// this test is entirely unrelated but I wanted to verify Go behavior
	m := map[float64]int{}

	for i := 0; i < 1000; i++ {
		m[math.NaN()] = i
	}
	if len(m) != 1000 {
		t.Errorf("every same NaN must create new entry")
	}

	packedNan := PackInt32(3)
	m[packedNan] = 9999
	if len(m) != 1001 {
		t.Errorf("should create anew")
	}

	_, ok := m[packedNan]
	if ok {
		t.Errorf("packed NaN must not be special")
	}

	if UnpackInt32(packedNan) != 3 {
		t.Errorf("wasn't packed properly")
	}
}
