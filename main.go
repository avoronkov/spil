package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	trace  bool
	bigint bool
)

func init() {
	flag.BoolVar(&trace, "trace", false, "trace function calls")
	flag.BoolVar(&trace, "t", false, "trace function calls (shorthand)")

	flag.BoolVar(&bigint, "big", false, "use big math")
	flag.BoolVar(&bigint, "b", false, "use big math (shorthand)")
}

func doMain() int {
	flag.Parse()
	if !trace {
		log.SetOutput(ioutil.Discard)
	}

	builtinDir, err := getBuiltinDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	log.Printf("builtin: %v\n", builtinDir)

	in := NewInterpreter(os.Stdout)
	in.UseBigInt(bigint)
	if err := in.LoadBuiltin(builtinDir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	if err := in.Run(input); err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(doMain())
}

func getBuiltinDir() (string, error) {
	binPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(binPath), "builtin"), nil
}
