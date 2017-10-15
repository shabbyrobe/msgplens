msgplens: Msgpack inspection tool
=================================

[![GoDoc](https://godoc.org/github.com/shabbyrobe/xmlwriter?status.svg)](https://godoc.org/github.com/shabbyrobe/xmlwriter)

Swiss-army knife for inspecting msgpack objects.

![msgplens screenshot](/doc/example.png?raw=true "msgplens screenshot")

To install:

    go get -u github.com/shabbyrobe/msgplens/msgplens

`msgplens` has the following features:

- Pretty-print msgpack objects
- Lossless JSON representation (`-inf repr`, `-outf repr`)
- Lossy JSON representation (`-inf json`, `-outf json`)
- Accepts several different input encodings (`-inenc b64`, `-inenc hex`, etc)
- Everything useful is exported from the `github.com/shabbyrobe/msgplens` library

And the following (likely temporary) drawbacks:

- Code is not pretty yet
- Not much in the way of tests either
- Code quality is a bit of a mixed bag at the moment so the API isn't super
  stable

This tool cribs heavily from the unexported portions of
http://github.com/tinylib/msgp, though it would be great if it could be exported
and used as a dependency instead!


Examples
--------

You can chop and dice inputs and outputs any way you like.

Read base64 encoded msgpack from STDIN, pretty print:

    $ echo "gA==" | msgplens -inenc b64
    at:0   sz:1   0x80 (128) Fixmap   len:0 {}

Read base64 encoded msgpack from STDIN, convert to lossy JSON:

    $ echo "gaFhoWI=" | ./msgplens -inenc b64 -outf json
    {"a":"b"}

Read hex encoded JSON from STDIN, pretty print:

    $ echo '{"a": "b"}' | xxd -ps | ./msgplens -inf json -inenc hex
    at:0   sz:1   0x81 (129) Fixmap   len:1 {
      K0  at:1   sz:2   0xa1 (161) Fixstr   "a"
      V0  at:3   sz:2   0xa1 (161) Fixstr   "b"
    }

