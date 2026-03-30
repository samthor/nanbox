package nanbox

import (
	"log"
	"math"
)

const (
	pack32base = 0b1_11111111111_0111 << 48
	pack51base = 0b1_11111111111_1 << 51
)

var (
	pack51mask uint64
)

func init() {
	pack51mask = pack51base
	pack51mask = ^pack51mask
}

// 64 bits -> 1 sign, 11 for exponent, 52 for significand
//  - NaN/Inf is all 1's in exponent
//  - Inf is all 0's in significand (so encoding "0" is hard)
// :. if we have 48 bytes for data (6 bytes), this gives us 2^4-1 values in remainder - 1-15 (zero invalid as Inf).

// PackInt32 packs any int32 into a float64 NaN.
// It places its bits on the right side, while encoding an extra unused sentinel on the left.
// This will never generate Go's built-in NaN bits.
func PackInt32(v int32) (f float64) {
	var enc uint64 = pack32base + uint64(v)
	return math.Float64frombits(enc)
}

// UnpackInt32 unpacks the right-side bits from a float64 NaN.
// It simply returns the right-most 32 bits (in the significand).
func UnpackInt32(f float64) (v int32) {
	raw := math.Float64bits(f)
	return int32(raw)
}

// PackUint51 packs the uint64 here, stripping the right-most 13 bits.
//
// There is 52 bits of significand in float64, but all zeros is treated as Infinity.
// This will never generate Go's built-in NaN bits.
func PackUint51(v uint64) (f float64) {
	var enc uint64 = pack51base + (v & pack51mask)
	return math.Float64frombits(enc)
}

// UnpackUint51 unpacks a uint64 from this NaN.
func UnpackUint51(f float64) (v uint64) {
	raw := math.Float64bits(f)
	return raw ^ pack51base
}

// PackBytes packs any number of bytes into a float64 array.
// This uses a known format which can only be unpacked via UnpackBytes.
// This stores 6 bytes per-float64, even though there is 2^52-1 possibilities, we just use 2^48.
// It contains continuation and remainder data, i.e., length, but only if the byte array has non-zero length.
// This will never generate Go's built-in NaN bits.
func PackBytes(b []byte) (f []float64) {
	for i := 0; i < len(b); i += 6 {
		control := 0b1111 // 15 "continue"

		here := 6
		if i+6 >= len(b) {
			here = len(b) - i       // 1-5
			control = here + 0b1000 // 1-6 "last"
		}

		// place in reverse (low on left)
		var value uint64
		for j := here - 1; j >= 0; j-- {
			value <<= 8
			value += uint64(b[i+j])
		}

		value += uint64(control) << 48
		value += 0b1_11111111111 << 52

		// enc += value

		log.Printf("got enc value=%+v (nan=%v) for slice=%+v", value, math.IsNaN(math.Float64frombits(value)), b[i:i+here])

		f = append(f, math.Float64frombits(value))
	}

	return f
}

// UnpackBytes unpacks a number of bytes from this float64 array previously packed with PackBytes.
// It looks for the control code and may only consume so many float64's.
func UnpackBytes(f []float64) (consumed int, b []byte) {
	for i, each := range f {
		if !math.IsNaN(each) {
			break
		}

		raw := math.Float64bits(each)
		control := raw & (0xf000000000000) >> 48 // this is the 13-16 bits set to true (control)

		count := 6
		var done bool

		if control != 15 {
			count = int(control & 0b0111)
			if count > 6 || count == 0 || (control&0b1000) == 0 {
				break
			}
			done = true
		}

		for i := 0; i < count; i++ {
			next := byte(raw & 0xff)
			b = append(b, next)
			raw >>= 8
		}
		if done {
			return i + 1, b
		}
	}

	return 0, nil
}
