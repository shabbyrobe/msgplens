package msgplens

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	styleKeyLen         = 4
	styleTypeLen        = 8
	styleAttrNameLen    = 4
	styleAttrValueLen   = 4
	styleAttrNameColor  = darkGray
	styleAttrValueColor = lightGray
	styleKeyIndexColor  = lightMagenta
	styleKeyTypeColor   = magenta
	styleIntColor       = green
	styleFloatColor     = yellow
	styleStringColor    = lightBlue
)

type Printer struct {
	vis *Visitor
	out io.Writer
	w   *writer

	AllowExtra bool
}

func (p *Printer) printType(ctx *LensContext, prefix byte, size int) {
	p.w.writef("%s%[2]*s",
		colorw(styleAttrNameColor, "at:"),
		styleAttrValueLen, colorw(styleAttrValueColor, ctx.Pos()))
	p.w.writef("%s%[2]*s",
		colorw(styleAttrNameColor, "sz:"),
		styleAttrValueLen, colorw(styleAttrValueColor, size))

	p.w.write(color(lightCyan, fmt.Sprintf("0x%02x (%03d) ", prefix, prefix)))
	p.w.writef("%[1]*s", styleTypeLen, colorw(cyan, prefixName(prefix)))
	p.w.write(" ")
}

func (p *Printer) Flush() error {
	return p.w.Flush()
}

func NewPrinter(out io.Writer) *Printer {
	p := &Printer{
		out: out,
		w: &writer{
			Writer: bufio.NewWriter(out),
			indent: "  ",
		},
	}
	p.vis = &Visitor{
		Str: func(ctx *LensContext, bts []byte, str string) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(color(styleStringColor, fmt.Sprintf("%q", str)))
			p.w.writeln()
			return nil
		},

		Int: func(ctx *LensContext, bts []byte, data int64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.writef("%d", colorw(styleIntColor, data))
			p.w.writeln()
			return nil
		},

		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.writef("%d", colorw(styleIntColor, data))
			p.w.writeln()
			return nil
		},

		Bin: func(ctx *LensContext, bts []byte, data []byte) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.writeln(prefixName(bts[0]))
			return nil
		},

		Float64: func(ctx *LensContext, bts []byte, data float64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(color(styleFloatColor, fmt.Sprintf("%g", data)))
			p.w.writeln()
			return nil
		},

		Float32: func(ctx *LensContext, bts []byte, data float32) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(color(styleFloatColor, fmt.Sprintf("%g", data)))
			p.w.writeln()
			return nil
		},

		Bool: func(ctx *LensContext, bts []byte, data bool) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.writeln()
			return nil
		},

		Nil: func(ctx *LensContext, prefix byte) error {
			p.printType(ctx, prefix, 1)
			p.w.writeln()
			return nil
		},

		Extension: func(ctx *LensContext, bts []byte) error {
			p.w.writeln(prefixName(bts[0]))
			return nil
		},

		EnterArray: func(ctx *LensContext, prefix byte, cnt int) error {
			p.printType(ctx, prefix, 1)
			p.w.write(color(styleAttrNameColor, "len:"), color(styleAttrValueColor, cnt))
			p.w.write(" [")
			if cnt > 0 {
				p.w.writeln()
			}
			p.w.depth++
			return nil
		},

		EnterArrayElem: func(ctx *LensContext, n, cnt int) error {
			p.w.writef("%[1]*s", styleKeyLen, colorw(styleKeyIndexColor, n))
			return nil
		},

		LeaveArrayElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},

		LeaveArray: func(ctx *LensContext, prefix byte, cnt int, bts []byte) error {
			p.w.depth--
			p.w.writeln("]")
			return nil
		},

		EnterMap: func(ctx *LensContext, prefix byte, cnt int) error {
			p.printType(ctx, prefix, 1)
			p.w.write(color(styleAttrNameColor, "len:"), color(styleAttrValueColor, cnt))
			p.w.write(" {")
			if cnt > 0 {
				p.w.writeln()
			}
			p.w.depth++
			return nil
		},

		EnterMapKey: func(ctx *LensContext, n, cnt int) error {
			p.w.writef("%[1]*s", styleKeyLen, colorw(styleKeyTypeColor, "K").Append(styleKeyIndexColor, n))
			return nil
		},

		LeaveMapKey: func(ctx *LensContext, n, cnt int) error {
			return nil
		},

		EnterMapElem: func(ctx *LensContext, n, cnt int) error {
			p.w.writef("%[1]*s", styleKeyLen, colorw(styleKeyTypeColor, "V").Append(styleKeyIndexColor, n))
			return nil
		},

		LeaveMapElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},

		LeaveMap: func(ctx *LensContext, prefix byte, cnt int, bts []byte) error {
			p.w.depth--
			p.w.writeln("}")
			return nil
		},

		End: func(ctx *LensContext, left []byte) (err error) {
			lln := len(left)
			if lln > 0 {
				p.w.writeln()
				p.w.writelnf("%d bytes remaining:", lln)
				p.w.depth++
				for i, b := range left {
					if i > 0 && i%26 == 0 {
						p.w.writeln()
					} else if i > 0 && i%2 == 0 {
						p.w.write(" ")
					}
					p.w.writef("%02x", b)
				}
				p.w.depth--
				p.w.writeln()

				if !p.AllowExtra {
					err = fmt.Errorf("%d bytes found at end of input", lln)
				}
			}
			ferr := p.w.Flush()
			if err == nil {
				err = ferr
			}
			return
		},
	}
	return p
}

func (p *Printer) Visitor() *Visitor {
	return p.vis
}

type writer struct {
	*bufio.Writer
	depth       int
	indent      string
	curIndented bool
}

func (w *writer) writeIndent() {
	if !w.curIndented {
		w.WriteString(strings.Repeat(w.indent, w.depth))
		w.curIndented = true
	}
}

func (w *writer) nextLine() {
	w.WriteByte('\n')
	w.curIndented = false
}

func (w *writer) write(strs ...string) {
	for _, str := range strs {
		for {
			idx := strings.IndexByte(str, '\n')
			piece := str
			if idx >= 0 {
				piece = str[0:idx]
			}
			w.writeIndent()
			w.WriteString(piece)
			if idx == -1 {
				break
			} else {
				w.nextLine()
			}
			str = str[idx+1:]
		}
	}
}

func (w *writer) writef(line string, args ...interface{}) {
	w.write(fmt.Sprintf(line, args...))
}

func (w *writer) writeln(strs ...string) {
	w.write(strs...)
	w.curIndented = false
	w.WriteByte('\n')
}

func (w *writer) writelnf(line string, args ...interface{}) {
	w.writef(line, args...)
	w.writeln()
}

var spaceOnly = regexp.MustCompile(`^[ \t]+$`)

func (w *writer) writeIndented(block string) {
	block = strings.TrimLeft(block, "\n")
	indent := strings.Repeat(w.indent, w.depth)
	parts := strings.Split(block, "\n")

	if len(parts) == 0 {
		return
	}

	i := 0
	var c rune
	for i, c = range parts[0] {
		if c != ' ' && c != '\t' {
			break
		}
	}
	strip := parts[0][0:i]

	for i, ln := range parts {
		ln = strings.TrimPrefix(ln, strip)
		if spaceOnly.MatchString(ln) {
			ln = ""
		}
		if ln != "" {
			io.WriteString(w, indent)
			fmt.Fprintf(w, ln)
		}
		if i != len(parts)-1 {
			w.WriteByte('\n')
		}
	}
}

const (
	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func color(col int, v interface{}) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", col, v)
}

type colorOut struct {
	out    string
	len    int
	pad    byte
	padCol int
}

func (c colorOut) Pad(b byte) colorOut {
	c.pad = b
	return c
}

func (c colorOut) Format(f fmt.State, r rune) {
	ln, ok := f.Width()
	if ok {
		pad := c.pad
		if pad == 0 {
			pad = ' '
		}
		diff := ln - c.len
		neg := f.Flag('-')
		if neg && diff > 0 {
			f.Write(bytes.Repeat([]byte{pad}, diff))
		}
		f.Write([]byte(c.out))
		if !neg && diff > 0 {
			f.Write(bytes.Repeat([]byte{pad}, diff))
		}
	} else {
		f.Write([]byte(c.out))
	}
}

func (c colorOut) Append(col int, v interface{}) colorOut {
	s := fmt.Sprintf("%v", v)
	c.out += color(col, s)
	c.len += len(s)
	return c
}

func (c colorOut) String() string {
	return c.out
}

func colorw(col int, v interface{}) colorOut {
	s := fmt.Sprintf("%v", v)
	out := color(col, s)
	return colorOut{
		out: out,
		len: len(s),
	}
}

func colorPadLeft(col int, v interface{}, w int, b byte) string {
	s := fmt.Sprintf("%v", v)
	diff := w - len(s)
	if diff > 0 {
		s = strings.Repeat(string(b), diff) + s
	}
	return color(col, s)
}

func colorPadRight(col int, v interface{}, w int, b byte) string {
	s := fmt.Sprintf("%v", v)
	diff := w - len(s)
	if diff > 0 {
		s = s + strings.Repeat(string(b), diff)
	}
	return color(col, s)
}
