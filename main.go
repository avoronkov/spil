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
	trace     bool
	bigint    bool
	stat      bool
	check     bool
	noBuiltin bool
)

func init() {
	flag.BoolVar(&trace, "trace", false, "trace function calls")
	flag.BoolVar(&trace, "t", false, "trace function calls (shorthand)")

	flag.BoolVar(&bigint, "big", false, "use big math")
	flag.BoolVar(&bigint, "b", false, "use big math (shorthand)")

	flag.BoolVar(&stat, "stat", false, "dump statistics after program exit")
	flag.BoolVar(&stat, "s", false, "dump statistics after program exit (shorthand)")

	flag.BoolVar(&check, "check", false, "make parsing and typechecking only")
	flag.BoolVar(&check, "c", false, "make parsing and typechecking only (shorthand)")

	flag.BoolVar(&noBuiltin, "no-builtin", false, "do not load std library functions")
	flag.BoolVar(&noBuiltin, "B", false, "do not load std library functions (shorthand)")
}

func doMain() int {
	flag.Parse()
	if !trace {
		log.SetOutput(ioutil.Discard)
	}

	builtinDir := ""
	if !noBuiltin {
		var err error
		builtinDir, err = getBuiltinDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		log.Printf("builtin: %v\n", builtinDir)
	}

	in := NewInterpreter(os.Stdout, builtinDir)
	in.UseBigInt(bigint)

	var input io.Reader
	if len(flag.Args()) >= 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	if err := in.Parse(input); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	if err := in.Check(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	if check {
		return 0
	}

	if err := in.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if stat {
		in.Stat()
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
