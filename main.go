package main

import (
	"io"
	"log"
	"os"
)

func main() {
	var input io.Reader
	if len(os.Args) >= 2 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	in := NewInterpreter(input, os.Stdout)
	if err := in.Run(); err != nil {
		log.Fatal(err)
	}
}
