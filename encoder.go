package msgplens

import (
	"fmt"
	"io"
)

// HexDecoder is a pretty janky attempt at making a fairly liberal hex-decoding
// Reader without resorting to crazy regular expressions. I'm not sure this is
// better. Don't use it in your next project.
//
// On the plus side, it tolerates input in lots of formats. You can pass in a
// delimited list of 1 to 2 digit hex numbers, optionally prefixed with "0x".
// It'll eat this for breakfast, brackets included: [0x01, 0x1, 1, ff, FF, 0xfF]
//
// It will also accept a continuous stream of hex characters, which it will
// process 2 characters at a time, i.e. "ABCD" is equivalent to "0xab 0xcd".
//
type HexDecoder struct {
	in   io.Reader
	part []byte
	eof  bool
	buf  []byte
	pos  int
}

func NewHexDecoder(rdr io.Reader, buf []byte) *HexDecoder {
	if buf == nil {
		buf = make([]byte, 8192)
	}
	return &HexDecoder{
		in:  rdr,
		buf: buf,
	}
}

const (
	hexDecNone = iota
	hexDecZero
	hexDecNum
	hexDecNext
)

func (h *HexDecoder) Read(b []byte) (n int, err error) {
	partLen := len(h.part)
	if h.eof && partLen == 0 {
		return 0, io.EOF
	}
	inLen := len(b)

	var cur []byte
	var state = hexDecNone

	var curByte byte
	var idx = 0
	var curLastIdx int

	for idx < inLen {
		if partLen > 0 {
			cur = h.part
			partLen = 0
		} else {
			rn, err := h.in.Read(h.buf)
			if err == io.EOF {
				h.eof = true
			} else if err != nil {
				return 0, err
			}
			cur = h.buf[0:rn]
		}

		if len(cur) == 0 {
			goto done
		}

		for i, c := range cur {
			h.pos++
		again:
			if state == hexDecNone {
				if c == '0' {
					state = hexDecZero
					continue
				} else if c >= '1' && c <= '9' {
					state = hexDecNum
					goto num
				} else if (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
					state = hexDecNum
					goto num
				} else if c == ',' || c == ' ' || c == '\n' || c == '\r' ||
					c == '[' || c == ']' || c == '(' || c == ')' || c == '\t' {
					continue
				} else {
					err = fmt.Errorf("unexpected char '%c'", c)
					return
				}
			}

			if state == hexDecZero {
				if c == 'x' {
					state = hexDecNum
					continue
				} else if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
					state = hexDecNext
					// fall through
				} else {
					err = fmt.Errorf("unexpected zero char '%c' at byte pos %d", c, h.pos-1)
					return
				}
			}

		num:
			if state == hexDecNum {
				state = hexDecNext
				if c >= '0' && c <= '9' {
					curByte = c - '0'
				} else if c >= 'a' && c <= 'f' {
					curByte = c - 'a' + 10
				} else if c >= 'A' && c <= 'F' {
					curByte = c - 'A' + 10
				} else {
					err = fmt.Errorf("unexpected num char %c", c)
					return
				}
			} else if state == hexDecNext {
				state = hexDecNone
				if c >= '0' && c <= '9' {
					curByte = (curByte << 4) + (c - '0')
				} else if c >= 'a' && c <= 'f' {
					curByte = (curByte << 4) + (c - 'a' + 10)
				} else if c >= 'A' && c <= 'F' {
					curByte = (curByte << 4) + (c - 'A' + 10)
				}

				b[idx] = curByte
				curByte = 0
				curLastIdx = i + 1
				idx++

				// this character is not part of the committed byte, i.e. 0xa
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					goto again
				}
				if idx == inLen {
					goto done
				}
			}
		}
	}

done:
	if len(cur) > curLastIdx {
		h.part = cur[curLastIdx:]
	} else {
		h.part = nil
	}

	n = idx

	return
}
