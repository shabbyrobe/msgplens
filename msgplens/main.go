package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/shabbyrobe/msgplens"
)

func main() {
	// enc := msgplens.NewPrinter(os.Stdout)
	// b, _ := ioutil.ReadAll(os.Stdin)
	// msgplens.WalkBytes(enc, b)
	// fmt.Println(enc.String())
	repr := msgplens.NewRepresenter()
	b, _ := ioutil.ReadAll(os.Stdin)
	msgplens.WalkBytes(repr, b)

	m, err := json.Marshal(repr.Nodes())
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(m)
}
