package msgplens

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Printer struct {
	vis *Visitor
	out io.Writer
	w   *writer
}

func (p *Printer) printType(ctx *LensContext, prefix byte, size int) {
	p.w.write(color(lightGray, "at:"), color(green, ctx.Pos()), " ")
	p.w.write(color(lightGray, "sz:"), color(green, size), " ")
	p.w.write(color(cyan, prefixName(prefix)), " ")
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
			p.w.write(fmt.Sprintf("%q", str))
			p.w.writeln()
			return nil
		},
		Int: func(ctx *LensContext, bts []byte, data int64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(fmt.Sprintf("%d", data))
			p.w.writeln()
			return nil
		},
		Uint: func(ctx *LensContext, bts []byte, data uint64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(fmt.Sprintf("%d", data))
			p.w.writeln()
			return nil
		},
		Bin: func(ctx *LensContext, bts []byte, data []byte) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.writeln(prefixName(bts[0]))
			return nil
		},
		Float: func(ctx *LensContext, bts []byte, data float64) error {
			p.printType(ctx, bts[0], len(bts))
			p.w.write(fmt.Sprintf("%f", data))
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
		EnterArray: func(ctx *LensContext, prefix byte, len int) error {
			p.printType(ctx, prefix, 1)
			p.w.write(color(lightGray, "len:"), color(green, len))
			p.w.writeln()
			p.w.depth++
			return nil
		},
		EnterArrayElem: func(ctx *LensContext, n, cnt int) error {
			p.w.write(color(lightBlue, n), " ")
			return nil
		},
		LeaveArrayElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveArray: func(ctx *LensContext) error {
			p.w.depth--
			return nil
		},

		EnterMap: func(ctx *LensContext, prefix byte, len int) error {
			p.printType(ctx, prefix, 1)
			p.w.write(color(lightGray, "len:"), color(green, len))
			p.w.writeln()
			p.w.depth++
			return nil
		},
		EnterMapKey: func(ctx *LensContext, n, cnt int) error {
			p.w.write(color(blue, "K"))
			p.w.write(color(lightBlue, n), " ")
			return nil
		},
		LeaveMapKey: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		EnterMapElem: func(ctx *LensContext, n, cnt int) error {
			p.w.write(color(blue, "V"))
			p.w.write(color(lightBlue, n), " ")
			return nil
		},
		LeaveMapElem: func(ctx *LensContext, n, cnt int) error {
			return nil
		},
		LeaveMap: func(ctx *LensContext) error {
			p.w.depth--
			return nil
		},
		End: func(ctx *LensContext) error {
			return p.w.Flush()
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
	w.WriteByte('\n')
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
	in  string
	col int
	len int
	pad byte
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
		if f.Flag('-') {
			f.Write([]byte(colorPadLeft(c.col, c.in, ln, pad)))
		} else {
			f.Write([]byte(colorPadRight(c.col, c.in, ln, pad)))
		}
	} else {
		f.Write([]byte(color(c.col, c.in)))
	}
}

func (c colorOut) String() string {
	return color(c.col, c.in)
}

func colorw(col int, v interface{}) colorOut {
	s := fmt.Sprintf("%v", v)
	return colorOut{
		in:  s,
		col: col,
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
