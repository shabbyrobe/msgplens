msgplens: Msgpack inspection tool
=================================

[![GoDoc](https://godoc.org/github.com/shabbyrobe/xmlwriter?status.svg)](https://godoc.org/github.com/shabbyrobe/xmlwriter)

Swiss-army knife for inspecting msgpack objects.

![msgplens screenshot](/doc/example.png?raw=true "msgplens screenshot")

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

