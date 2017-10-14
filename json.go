package msgplens

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type JSONEncoder struct {
	buf          *bytes.Buffer
	vis          *Visitor
	floatScratch []byte
}

func (j *JSONEncoder) Visitor() *Visitor {
	return j.vis
}

func NewJSONEncoder() *JSONEncoder {
	je := &JSONEncoder{}
	je.buf = &bytes.Buffer{}
	je.vis = &Visitor{
		Nil: func(ctx *LensContext, prefix byte) error { je.buf.WriteString("null"); return nil },
		Str: func(ctx *LensContext, bts []byte, str string) error { je.writeJSONString(str); return nil },
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			je.buf.WriteString(strconv.FormatInt(data, 10))
			return nil
		},
		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			je.buf.WriteString(strconv.FormatUint(data, 10))
			return nil
		},
		Bin:       func(ctx *LensContext, bts []byte, data []byte) error { je.writeBin(data); return nil },
		Float:     func(ctx *LensContext, bts []byte, data float64) error { return je.writeJSONFloat(data) },
		Extension: func(ctx *LensContext, bts []byte) error { je.writeBin(bts); return nil },
		Bool: func(ctx *LensContext, bts []byte, data bool) error {
			if data {
				je.buf.WriteString("true")
			} else {
				je.buf.WriteString("false")
			}
			return nil
		},

		EnterArray:  func(ctx *LensContext, prefix byte, len int) error { je.buf.WriteByte('['); return nil },
		LeaveArray:  func(ctx *LensContext) error { je.buf.WriteByte(']'); return nil },
		EnterMap:    func(ctx *LensContext, prefix byte, len int) error { je.buf.WriteByte('{'); return nil },
		LeaveMapKey: func(ctx *LensContext, n, cnt int) error { je.buf.WriteByte(':'); return nil },
		LeaveMap:    func(ctx *LensContext) error { je.buf.WriteByte('}'); return nil },

		LeaveArrayElem: func(ctx *LensContext, n, cnt int) error {
			if n < cnt-1 {
				je.buf.WriteByte(',')
			}
			return nil
		},

		LeaveMapElem: func(ctx *LensContext, n, cnt int) error {
			if n < cnt-1 {
				je.buf.WriteByte(',')
			}
			return nil
		},
	}
	return je
}

func (j *JSONEncoder) writeBin(data []byte) {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	str := *(*string)(unsafe.Pointer(&reflect.StringHeader{Data: sh.Data, Len: sh.Len}))
	j.writeJSONString(str)
}

func (j *JSONEncoder) writeJSONString(s string) {
	j.buf.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				j.buf.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				j.buf.WriteByte('\\')
				j.buf.WriteByte(b)
			case '\n':
				j.buf.WriteByte('\\')
				j.buf.WriteByte('n')
			case '\r':
				j.buf.WriteByte('\\')
				j.buf.WriteByte('r')
			case '\t':
				j.buf.WriteByte('\\')
				j.buf.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				j.buf.WriteString(`\u00`)
				j.buf.WriteByte(hex[b>>4])
				j.buf.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				j.buf.WriteString(s[start:i])
			}
			j.buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				j.buf.WriteString(s[start:i])
			}
			j.buf.WriteString(`\u202`)
			j.buf.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		j.buf.WriteString(s[start:])
	}
	j.buf.WriteByte('"')
	return
}

func (j *JSONEncoder) writeJSONFloat(f float64) error {
	bits := 64
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Errorf("unsupported value: %v", f)
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	b := j.floatScratch[:0]
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, fmt, -1, int(bits))
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	j.buf.Write(b)
	return nil
}

func (j *JSONEncoder) String() string {
	return j.buf.String()
}

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

var hex = "0123456789abcdef"
