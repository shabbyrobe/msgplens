package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/shabbyrobe/msgplens"
)

type jsonEncoder struct {
	bytes.Buffer
	vis *msgplens.Visitor
}

// NOTE: keep in sync with stringBytes below.
func (j *jsonEncoder) string(s string) {
	j.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				j.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				j.WriteByte('\\')
				j.WriteByte(b)
			case '\n':
				j.WriteByte('\\')
				j.WriteByte('n')
			case '\r':
				j.WriteByte('\\')
				j.WriteByte('r')
			case '\t':
				j.WriteByte('\\')
				j.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				j.WriteString(`\u00`)
				j.WriteByte(hex[b>>4])
				j.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				j.WriteString(s[start:i])
			}
			j.WriteString(`\ufffd`)
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
				j.WriteString(s[start:i])
			}
			j.WriteString(`\u202`)
			j.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		j.WriteString(s[start:])
	}
	j.WriteByte('"')
	return
}

func newJSONEncoder() *jsonEncoder {
	je := &jsonEncoder{}
	je.vis = &msgplens.Visitor{
		Str: func(bts []byte, str string) error { je.string(str); return nil },
		Int: func(bts []byte, data int64) error { je.WriteString(strconv.FormatInt(data, 10)); return nil },
		Uint: func(bts []byte, data uint64) error {
			fmt.Println("UINT", bts, data)
			return nil
		},
		Bin: func(bts []byte, data []byte) error {
			fmt.Println("BIN", bts, data)
			return nil
		},
		Float: func(bts []byte, data float64) error {
			fmt.Println("Float", bts, data)
			return nil
		},
		Bool: func(bts []byte, data bool) error {
			if data {
				je.WriteString("true")
			} else {
				je.WriteString("false")
			}
			return nil
		},
		Nil: func() error { je.WriteString("null"); return nil },

		EnterArray:  func(prefix byte, len int) error { je.WriteByte('['); return nil },
		LeaveArray:  func() error { je.WriteByte(']'); return nil },
		EnterMap:    func(prefix byte, len int) error { je.WriteByte('{'); return nil },
		LeaveMapKey: func(n, cnt int) error { je.WriteByte(':'); return nil },
		LeaveMap:    func() error { je.WriteByte('}'); return nil },

		LeaveArrayElem: func(n, cnt int) error {
			if n < cnt-1 {
				je.WriteByte(',')
			}
			return nil
		},

		LeaveMapElem: func(n, cnt int) error {
			if n < cnt-1 {
				je.WriteByte(',')
			}
			return nil
		},

		Extension: func(bts []byte) error {
			fmt.Println("Extension", bts)
			return nil
		},
	}
	return je
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

func main() {
	enc := newJSONEncoder()
	b, _ := ioutil.ReadAll(os.Stdin)
	msgplens.WalkBytes(enc.vis, b)
	fmt.Println(enc.String())
}
