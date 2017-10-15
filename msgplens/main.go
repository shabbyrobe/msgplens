package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/shabbyrobe/msgplens"
)

const usage = `
msgplens [options]

Options:
  -inf <fmt>     Input format
  -outf <fmt>    Output format
  -inenc <enc>   Input encoding (optional)
  -outenc <enc>  Input encoding (optional)

Formats:
  msgp   Msgpack (default input, output)
  print  Pretty printed output (default output)
  repr   Full representation of msgpack objects in JSON format (input, output)
  json   Lossy JSON approximation (input, output)

Encodings:
  py3b   Python 3 binary string (input)
  hex    List of bytes as hex numbers. May be comma separated. May be prefixed
         with 0x. Whitespace ignored. Example: "0a120b", "0a 12 0b", "0x0a, 0x12,
         0x0b" (input)
  nums   List of bytes as decimal numbers. May be comma separated. Whitespace 
         ignored (input)
  b64    Base64 (using Golang's encoding/base64.StdEncoding) (input, output)
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	var (
		inFormat    string
		outFormat   string
		inEncoding  string
		outEncoding string
	)

	if len(os.Args) == 1 {
		// Check if we have anything to read from STDIN. If not, print help.
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return usageError{}
		}
	}

	flag.StringVar(&inFormat, "inf", "msgp", "Input format")
	flag.StringVar(&outFormat, "outf", "print", "Output format")
	flag.StringVar(&inEncoding, "inenc", "", "Input encoding")
	flag.StringVar(&outEncoding, "outenc", "", "Output encoding")
	flag.Parse()

	var rdr io.Reader = os.Stdin
	var wrt io.WriteCloser = os.Stdout

	switch inEncoding {
	case "hex":
		rdr = msgplens.NewHexDecoder(rdr, 0)

	case "num":
		bts, err := ioutil.ReadAll(rdr)
		if err != nil {
			return err
		}
		if bts, err = decodeNums(bts, 10); err != nil {
			return err
		}
		rdr = bytes.NewReader(bts)

	case "py3b":
		bts, err := ioutil.ReadAll(rdr)
		if err != nil {
			return err
		}
		if bts, err = decodePy3Bytes(bts); err != nil {
			return err
		}
		rdr = bytes.NewReader(bts)

	case "base64":
		fallthrough
	case "b64":
		rdr = base64.NewDecoder(base64.StdEncoding, rdr)

	case "":
		// all good!

	default:
		return fmt.Errorf("unknown input encoding %s", inEncoding)
	}

	switch outEncoding {
	case "base64":
		fallthrough
	case "b64":
		wrt = base64.NewEncoder(base64.StdEncoding, wrt)
	}

	in, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}

	if inFormat == "msgp" && outFormat == "print" {
		enc := msgplens.NewPrinter(wrt)
		msgplens.WalkBytes(enc, in)

	} else if inFormat == outFormat {
		wrt.Write(in)

	} else {
		var node msgplens.Node
		switch inFormat {
		case "repr":
			node, err = msgplens.ReprUnmarshalNode(in)
			if err != nil {
				return err
			}

		case "json":
			node, err = msgplens.UnmarshalJSON(in)
			if err != nil {
				return err
			}

		case "msgp":
			repr := msgplens.NewRepresenter()
			msgplens.WalkBytes(repr, in)
			node = repr.Nodes()[0]

		default:
			return usageError{fmt.Sprintf("Unknown input format %s", inFormat)}
		}

		switch outFormat {
		case "repr":
			m, err := json.Marshal(node)
			if err != nil {
				panic(err)
			}
			wrt.Write(m)
			fmt.Println()

		case "msgp":
			var buf bytes.Buffer
			if err := node.Msgpack(&buf); err != nil {
				return err
			}
			wrt.Write(buf.Bytes())

		case "json":
			var buf bytes.Buffer
			if err := node.Msgpack(&buf); err != nil {
				return err
			}
			enc := msgplens.NewJSONEncoder()
			msgplens.WalkBytes(enc, buf.Bytes())
			io.WriteString(wrt, enc.String())

		case "print":
			var buf bytes.Buffer
			if err := node.Msgpack(&buf); err != nil {
				return err
			}
			enc := msgplens.NewPrinter(wrt)
			msgplens.WalkBytes(enc, buf.Bytes())
			if err := enc.Flush(); err != nil {
				return err
			}

		default:
			return usageError{fmt.Sprintf("Unknown output format %s", outFormat)}
		}
	}

	wrt.Close()

	return nil
}

type usageError struct {
	msg string
}

func (u usageError) Error() string {
	out := ""
	if u.msg != "" {
		out += u.msg + "\n\n"
	}
	out += strings.TrimLeft(usage, "\n")
	return out
}
