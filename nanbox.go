// Package nanbox provides helpers to encode/decode data into the unused bits of a NaN.
// It will never generate the bits of Go's built-in NaN.
package nanbox

// 64 bits -> 1 sign, 11 for exponent, 52 for significand
//  - NaN/Inf is all 1's in exponent
//  - Inf is all 0's in significand (so encoding "0" is hard)
// :. if we have 48 bytes for data (6 bytes), this gives us 2^4-1 values in remainder - 1-15 (zero invalid as Inf).
//
// Go's built-in NaN uses 1 on both sides of the significand, with sign set to zero.
// so it will be hard to encode 52 bits while not also colliding here.

import (
	"encoding/binary"
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

// PackUint51 packs the uint64 here, stripping the left-most 13 bits.
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
//
// This wastes (length % 6) space, so try to 6-align your data.
// Also, note that encoding a nil or zero-length buffer generates a single "empty" float64 (the most wasted you can be)!
func PackBytes(b []byte) (f []float64) {
	if len(b) == 0 {
		return []float64{math.Float64frombits(pack51base)}
	}

	for i := 0; i < len(b); i += 6 {
		control := 0b1111 // 15 "continue"

		here := 6
		if i+6 >= len(b) {
			// we're on the last float64 with 1-6 bytes remaining
			here = len(b) - i
			control = here + 0b1000
		}

		// go forwards so packing retains byte order
		var value uint64
		for j := 0; j < here; j++ {
			value <<= 8
			value += uint64(b[i+j])
		}

		value += uint64(control) << 48
		value += 0b1_11111111111 << 52

		f = append(f, math.Float64frombits(value))
	}

	return f
}

// UnpackBytes unpacks a number of bytes from the front of this float64 array, previously packed with PackBytes.
// It looks for the control code and may only consume so many float64's.
//
// This returns zero/nil if no data could be consumed.
// This may return 1/nil if this is unpacking a nil or zero-length buffer (we don't retain nil vs zero).
func UnpackBytes(f []float64) (consumed int, b []byte) {
	for i, each := range f {
		if !math.IsNaN(each) {
			break
		}

		raw := math.Float64bits(each)
		var tmp [8]byte
		binary.BigEndian.PutUint64(tmp[:], raw)
		control := tmp[1] & 0b00001111 // bits 13-16

		// all data case
		if control == 15 {
			b = append(b, tmp[2:]...)
			continue
		}

		// tail case
		count := int(control & 0b0111)
		if count > 6 || (control&0b1000) == 0 {
			break
		} else if count == 0 {
			if i == 0 {
				return 1, nil // allow at start
			}
			break // not valid
		}
		b = append(b, tmp[8-count:]...)
		return i + 1, b
	}

	return 0, nil
}
