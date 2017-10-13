package msgplens

import (
	"fmt"
	"math"
)

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

func isfixint(b byte) bool {
	return b>>7 == 0
}

func isnfixint(b byte) bool {
	return b&first3 == FixintNeg
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
