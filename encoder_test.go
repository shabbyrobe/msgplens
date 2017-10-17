package msgplens

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"strconv"
	"testing"
)

func BenchmarkHexDecoder(b *testing.B) {
	rn := make([]byte, 2048)
	rand.Read(rn)
	hexLen := hex.EncodedLen(2048)
	on := make([]byte, hexLen)
	n := hex.Encode(on, rn)
	on = on[:n]

	buf := make([]byte, 8192)

	b.SetBytes(int64(hexLen * b.N))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := NewHexDecoder(bytes.NewReader(on), buf)
		io.Copy(ioutil.Discard, d)
	}
}

func BenchmarkHexDecodeBytes(b *testing.B) {
	rn := make([]byte, 2048)
	rand.Read(rn)
	hexLen := hex.EncodedLen(2048)
	on := make([]byte, hexLen)
	n := hex.Encode(on, rn)
	on = on[:n]

	into := make([]byte, 8192)
	b.SetBytes(int64(hexLen * b.N))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hexDecodeResult, _ = decodeHexStream(into, on)
	}
}

var hexDecodeResult []byte

func decodeHexStream(dst []byte, input []byte) ([]byte, error) {
	idx := 0
	for i := 0; i < len(input); i += 2 {
		c := input[i : i+2]

		i, err := strconv.ParseUint(string(c), 16, 8)
		if err != nil {
			return nil, err
		}
		dst[idx] = byte(i)
		idx++
	}
	return dst[:idx], nil
}
