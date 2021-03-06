package msgplens

import (
	"fmt"
	"math"
)

// Contains portions adapted from http://github.com/tinylib/msgp/msgp

// Copyright (c) 2014 Philip Hofer
// Portions Copyright (c) 2009 The Go Authors (license at http://golang.org) where indicated
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

func getMint64(b []byte) int64 {
	return (int64(b[1]) << 56) | (int64(b[2]) << 48) |
		(int64(b[3]) << 40) | (int64(b[4]) << 32) |
		(int64(b[5]) << 24) | (int64(b[6]) << 16) |
		(int64(b[7]) << 8) | (int64(b[8]))
}

func getMint32(b []byte) int32 {
	return (int32(b[1]) << 24) | (int32(b[2]) << 16) | (int32(b[3]) << 8) | (int32(b[4]))
}

func getMint16(b []byte) (i int16) {
	return (int16(b[1]) << 8) | int16(b[2])
}

func getMint8(b []byte) (i int8) {
	return int8(b[1])
}

func getMuint64(b []byte) uint64 {
	return (uint64(b[1]) << 56) | (uint64(b[2]) << 48) |
		(uint64(b[3]) << 40) | (uint64(b[4]) << 32) |
		(uint64(b[5]) << 24) | (uint64(b[6]) << 16) |
		(uint64(b[7]) << 8) | (uint64(b[8]))
}

func getMuint32(b []byte) uint32 {
	return (uint32(b[1]) << 24) | (uint32(b[2]) << 16) | (uint32(b[3]) << 8) | (uint32(b[4]))
}

func getMuint16(b []byte) uint16 {
	return (uint16(b[1]) << 8) | uint16(b[2])
}

func getMuint8(b []byte) uint8 {
	return uint8(b[1])
}

func putMuint8(b []byte, u uint8) {
	b[0] = Uint8
	b[1] = byte(u)
}

func putMuint16(b []byte, u uint16) {
	b[0] = Uint16
	b[1] = byte(u >> 8)
	b[2] = byte(u)
}

func putMuint32(b []byte, u uint32) {
	b[0] = Uint32
	b[1] = byte(u >> 24)
	b[2] = byte(u >> 16)
	b[3] = byte(u >> 8)
	b[4] = byte(u)
}

func putMuint64(b []byte, u uint64) {
	b[0] = Uint64
	b[1] = byte(u >> 56)
	b[2] = byte(u >> 48)
	b[3] = byte(u >> 40)
	b[4] = byte(u >> 32)
	b[5] = byte(u >> 24)
	b[6] = byte(u >> 16)
	b[7] = byte(u >> 8)
	b[8] = byte(u)
}

func putMint8(b []byte, i int8) {
	b[0] = Int8
	b[1] = byte(i)
}

func putMint16(b []byte, i int16) {
	b[0] = Int16
	b[1] = byte(i >> 8)
	b[2] = byte(i)
}

func putMint32(b []byte, i int32) {
	b[0] = Int32
	b[1] = byte(i >> 24)
	b[2] = byte(i >> 16)
	b[3] = byte(i >> 8)
	b[4] = byte(i)
}

func putMint64(b []byte, i int64) {
	b[0] = Int64
	b[1] = byte(i >> 56)
	b[2] = byte(i >> 48)
	b[3] = byte(i >> 40)
	b[4] = byte(i >> 32)
	b[5] = byte(i >> 24)
	b[6] = byte(i >> 16)
	b[7] = byte(i >> 8)
	b[8] = byte(i)
}

func isfixint(b byte) bool {
	return b>>7 == 0
}

func isnfixint(b byte) bool {
	return b&first3 == FixintNeg
}

func wfixmap(u uint8) byte {
	return Fixmap | (u & last4)
}

func wfixstr(u uint8) byte {
	return (u & last5) | Fixstr
}

func wfixarray(u uint8) byte {
	return (u & last4) | Fixarray
}

func isfixstr(b byte) bool {
	return b&first3 == Fixstr
}

func isfixmap(b byte) bool {
	return b&first4 == Fixmap
}

func wfixint(u uint8) byte {
	return u & last7
}

func wnfixint(i int8) byte {
	return byte(i) | FixintNeg
}

// ReadFloat64Bytes tries to read a float64
// from 'b' and return the value and the remaining bytes.
// Possible errors:
// - ErrShortBytes (too few bytes)
// - TypeError{} (not a float64)
func readFloat64(b []byte) (f float64, err error) {
	if len(b) != 9 {
		err = fmt.Errorf("float64: short bytes")
		return
	}
	if b[0] != Float64 {
		err = fmt.Errorf("float32: invalid type, expected %s", getType(b[0]))
		return
	}
	f = math.Float64frombits(getMuint64(b))
	return
}

// ReadFloat32Bytes tries to read a float64
// from 'b' and return the value and the remaining bytes.
// Possible errors:
// - ErrShortBytes (too few bytes)
// - TypeError{} (not a float32)
func readFloat32(b []byte) (f float32, err error) {
	if len(b) != 5 {
		err = fmt.Errorf("float64: short bytes")
		return
	}
	if b[0] != Float32 {
		err = fmt.Errorf("float32: invalid type, expected %s", getType(b[0]))
		return
	}

	f = math.Float32frombits(getMuint32(b))
	return
}

// ReadInt64Bytes tries to read an int64
// from 'b' and return the value and the remaining bytes.
// Possible errors:
// - ErrShortBytes (too few bytes)
// - TypeError (not a int)
func readInt64(b []byte) (i int64, err error) {
	l := len(b)
	if l < 1 {
		return 0, fmt.Errorf("int64: short bytes")
	}

	lead := b[0]
	if isfixint(lead) {
		i = int64(lead)
		return
	}
	if isnfixint(lead) {
		i = int64(lead)
		return
	}

	switch lead {
	case Int8:
		if l < 2 {
			err = fmt.Errorf("int64: short bytes")
			return
		}
		i = int64(getMint8(b))
		return

	case Int16:
		if l < 3 {
			err = fmt.Errorf("int64: short bytes")
			return
		}
		i = int64(getMint16(b))
		return

	case Int32:
		if l < 5 {
			err = fmt.Errorf("int64: short bytes")
			return
		}
		i = int64(getMint32(b))
		return

	case Int64:
		if l < 9 {
			err = fmt.Errorf("int64: short bytes")
			return
		}
		i = getMint64(b)
		return

	default:
		err = fmt.Errorf("int64: bad prefix %x", lead)
		return
	}
}

// ReadUint64Bytes tries to read a uint64
// from 'b' and return the value and the remaining bytes.
// Possible errors:
// - ErrShortBytes (too few bytes)
// - TypeError{} (not a uint)
func readUint64(b []byte) (u uint64, err error) {
	l := len(b)
	if l < 1 {
		return 0, fmt.Errorf("uint64: short bytes")
	}

	lead := b[0]
	if isfixint(lead) {
		u = uint64(lead)
		return
	}

	switch lead {
	case Uint8:
		if l < 2 {
			err = fmt.Errorf("uint64: short bytes")
			return
		}
		u = uint64(getMuint8(b))
		return

	case Uint16:
		if l < 3 {
			err = fmt.Errorf("uint64: short bytes")
			return
		}
		u = uint64(getMuint16(b))
		return

	case Uint32:
		if l < 5 {
			err = fmt.Errorf("uint64: short bytes")
			return
		}
		u = uint64(getMuint32(b))
		return

	case Uint64:
		if l < 9 {
			err = fmt.Errorf("uint64: short bytes")
			return
		}
		u = getMuint64(b)
		return

	default:
		err = fmt.Errorf("uint64: bad prefix %x", lead)
		return
	}
}

// write prefix and uint8
func prefixu8(b []byte, pre byte, sz uint8) {
	b[0] = pre
	b[1] = byte(sz)
}

// write prefix and big-endian uint16
func prefixu16(b []byte, pre byte, sz uint16) {
	b[0] = pre
	b[1] = byte(sz >> 8)
	b[2] = byte(sz)
}

// write prefix and big-endian uint32
func prefixu32(b []byte, pre byte, sz uint32) {
	b[0] = pre
	b[1] = byte(sz >> 24)
	b[2] = byte(sz >> 16)
	b[3] = byte(sz >> 8)
	b[4] = byte(sz)
}

func prefixu64(b []byte, pre byte, sz uint64) {
	b[0] = pre
	b[1] = byte(sz >> 56)
	b[2] = byte(sz >> 48)
	b[3] = byte(sz >> 40)
	b[4] = byte(sz >> 32)
	b[5] = byte(sz >> 24)
	b[6] = byte(sz >> 16)
	b[7] = byte(sz >> 8)
	b[8] = byte(sz)
}

func writeStringHeader(prefix byte, b []byte, sz uint32) ([]byte, error) {
	if isfixstr(prefix) {
		b[0] = wfixstr(uint8(sz))
		return b[:1], nil

	} else if prefix == Str8 {
		prefixu8(b, Str8, uint8(sz))
		return b[:2], nil

	} else if prefix == Str16 {
		prefixu16(b, Str16, uint16(sz))
		return b[:3], nil

	} else if prefix == Str32 {
		prefixu32(b, Str32, uint32(sz))
		return b[:5], nil

	} else {
		return nil, fmt.Errorf("unsupported prefix %02x", prefix)
	}
}

func writeBinHeader(prefix byte, b []byte, sz uint32) ([]byte, error) {
	switch {
	case prefix == Bin8:
		prefixu8(b, Bin8, uint8(sz))
		return b[0:2], nil
	case prefix == Bin16:
		prefixu16(b, Bin16, uint16(sz))
		return b[0:3], nil
	case prefix == Bin32:
		prefixu32(b, Bin32, uint32(sz))
		return b[0:5], nil
	default:
		return nil, fmt.Errorf("unsupported prefix %02x", prefix)
	}
}

func writeMapHeader(prefix byte, b []byte, sz uint32) ([]byte, error) {
	switch {
	case isfixmap(prefix):
		b[0] = wfixmap(uint8(sz))
		return b[:1], nil

	case prefix == Map16:
		prefixu16(b, Map16, uint16(sz))
		return b[:3], nil

	case prefix == Map32:
		prefixu32(b, Map32, sz)
		return b[:5], nil

	default:
		return nil, fmt.Errorf("unsupported prefix %02x", prefix)
	}
}

func isfixarray(b byte) bool {
	return b&first4 == Fixarray
}

func writeArrayHeader(prefix byte, b []byte, sz uint32) ([]byte, error) {
	if isfixarray(prefix) {
		b[0] = wfixarray(uint8(sz))
		return b[:1], nil

	} else if prefix == Array16 {
		prefixu16(b, Array16, uint16(sz))
		return b[:3], nil

	} else if prefix == Array32 {
		prefixu32(b, Array32, sz)
		return b[:5], nil

	} else {
		return nil, fmt.Errorf("unsupported prefix %02x", prefix)
	}
}
