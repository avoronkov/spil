package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var trace = false

func init() {
	flag.BoolVar(&trace, "trace", false, "trace function calls")
}

func doMain() int {
	flag.Parse()
	if !trace {
		log.SetOutput(ioutil.Discard)
	}

	var input io.Reader
	if len(flag.Args()) >= 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return 1
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	in := NewInterpreter(input, os.Stdout)
	if err := in.Run(); err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(doMain())
}
