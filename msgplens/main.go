package main

import (
	"io/ioutil"
	"os"

	"github.com/shabbyrobe/msgplens"
)

func main() {
	enc := msgplens.NewPrinter(os.Stdout)
	b, _ := ioutil.ReadAll(os.Stdin)
	msgplens.WalkBytes(enc, b)
	// fmt.Println(enc.String())
}
